package ldb

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"

	//"golang.org/x/exp/mmap" - not good, interface requires memory copy
	"../util"
	"github.com/edsrzf/mmap-go"
)

const (
	SSTABLE_MAX_RECORDS = uint32(1024 * 1024)
	SSTABLE_MAX_LENGTH  = uint32(1 * 1024 * 1024 * 1024)
)

////////////////////////////////////////////////////////////////////////////////
// Interface

type ISSTable interface {
	// SSTable Level Attributes
	Version() uint32                // Version
	ConsensusID() util.IConsensusID // Consensus ID
	Domain() []byte                 // Domain Name
	Table() []byte                  // Table Name
	StartTime() util.IConsensusTime // Start Time
	EndTime() util.IConsensusTime   // End Time
	StartKey() util.IData           // Start Key
	EndKey() util.IData             // End Key
	Level() uint32                  // Level
	Count() uint32                  // Record Count
	// Record Operations
	Get(key, group *util.IData) (util.IRecord, error)                // get record with specified key and attribute group
	Groups(key *util.IData) ([]util.IData, error)                    // retrieve a list of attribute groups for the given key
	Keys(key *util.IData) ([]util.IData, error)                      // retrieve a list of keys with the given key as prefix
	Read(pos, suggest_offset uint32) ([]util.IRecord, uint32, error) // Batch scan and read operation, return a list of IRecord, bytes read, or error
	// Close the resource
	Close() error
}

////////////////////////////////////////////////////////////////////////////////
// Implementation V1

type SSTableV1 struct {
	// basic attributes
	filepath     string
	version      uint32
	consensus_id util.IConsensusID
	domain       util.IData
	table        util.IData
	start_time   util.IConsensusTime
	end_time     util.IConsensusTime
	start_key    util.IData
	end_key      util.IData
	level        uint32
	count        uint32 // number of records
	// lookup table
	mph_table     *util.MPHTable
	record_offset []uint32 // record offset table
	// file as mmap
	mmap_data *mmap.MMap
	// record start
	record_start_pos uint32 // start of record position
}

func NewSSTableV1(filepath string) (t *SSTableV1, err error) {

	f, err := os.OpenFile(filepath, os.O_RDONLY, 0644)
	if err != nil {
		return
	}

	mmap_data, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		return
	}

	// creates sstable v1 if mmap is successful
	t = &SSTableV1{filepath: filepath, mmap_data: &mmap_data}
	defer func() {
		if err != nil && t != nil {
			t.Close()
			t = nil
		}
	}()

	// pos tracking
	pos := 0

	////////////////////////////////////////
	// parse header

	// parse version
	if len(mmap_data) < pos+4 {
		err = fmt.Errorf("NewSSTableV1 - no version")
		return
	}
	t.version = binary.BigEndian.Uint32(mmap_data[pos : pos+4])
	if t.version != 1 {
		err = fmt.Errorf("NewSSTableV1 - unsupported version - %d", t.version)
		return
	}
	pos += 4

	// parse consensus_id
	t.consensus_id, err = util.NewMappedConsensusID(mmap_data[pos:])
	if err != nil {
		return
	}
	pos += len(t.consensus_id.Buf())

	// parse domain
	t.domain, err = util.NewStandardMappedData(mmap_data[pos:])
	if err != nil {
		return
	}
	pos += len(t.domain.Buf())

	// parse domain
	t.table, err = util.NewStandardMappedData(mmap_data[pos:])
	if err != nil {
		return
	}
	pos += len(t.table.Buf())

	// parse start time
	t.start_time, err = util.NewConsensusTime(mmap_data[pos:])
	if err != nil {
		return
	}
	pos += len(t.start_time.Buf())

	// parse end time
	t.end_time, err = util.NewConsensusTime(mmap_data[pos:])
	if err != nil {
		return
	}
	pos += len(t.end_time.Buf())

	// parse start key
	t.start_key, err = util.NewStandardMappedData(mmap_data[pos:])
	if err != nil {
		return
	}
	pos += len(t.start_key.Buf())

	// parse end key
	t.end_key, err = util.NewStandardMappedData(mmap_data[pos:])
	if err != nil {
		return
	}
	pos += len(t.end_key.Buf())

	// parse level
	if len(mmap_data) < pos+4 {
		err = fmt.Errorf("NewSSTableV1 - no level")
		return
	}
	t.level = binary.BigEndian.Uint32(mmap_data[pos : pos+4])
	pos += 4

	// parse count
	if len(mmap_data) < pos+4 {
		err = fmt.Errorf("NewSSTableV1 - no count")
		return
	}
	t.count = binary.BigEndian.Uint32(mmap_data[pos : pos+4])
	if t.count > SSTABLE_MAX_RECORDS {
		err = fmt.Errorf("NewSSTableV1 - unsupported count - %d", t.count)
		return
	}
	pos += 4

	// parse header crc32
	computed_header_crc32 := crc32.ChecksumIEEE(mmap_data[:pos])
	if len(mmap_data) < pos+4 {
		err = fmt.Errorf("NewSSTableV1 - no header crc32")
		return
	}
	header_crc32 := binary.BigEndian.Uint32(mmap_data[pos : pos+4])
	if computed_header_crc32 != header_crc32 {
		err = fmt.Errorf("NewSSTableV1 - header crc32 checksum failed - computed %d vs header %d", computed_header_crc32, header_crc32)
		return
	}
	pos += 4

	////////////////////////////////////////
	// mph hash and offset table

	// parse mph
	mph_pos := pos
	mph_length := 0
	t.mph_table, mph_length, err = util.NewMPHTable(mmap_data[pos:])
	if err != nil {
		return
	}
	pos += mph_length

	// parse record offset size
	if len(mmap_data) < pos+4 {
		err = fmt.Errorf("NewSSTableV1 - no record offset size")
		return
	}
	record_offset_size := binary.BigEndian.Uint32(mmap_data[pos : pos+4])
	if record_offset_size > SSTABLE_MAX_RECORDS {
		err = fmt.Errorf("NewSSTableV1 - unsupported record offset size - %d", record_offset_size)
		return
	}
	pos += 4

	// parse each record offset data
	if len(mmap_data) < pos+4*int(record_offset_size) {
		err = fmt.Errorf("NewSSTableV1 - no record offset data")
	}
	t.record_offset = make([]uint32, int(record_offset_size))
	for i := 0; i < int(record_offset_size); i++ {
		t.record_offset[i] = binary.BigEndian.Uint32(mmap_data[pos : pos+4])
		pos += 4
	}

	// parse mph crc32
	computed_mph_crc32 := crc32.ChecksumIEEE(mmap_data[mph_pos:pos])
	if len(mmap_data) < pos+4 {
		err = fmt.Errorf("NewSSTableV1 - no mph crc32")
		return
	}
	mph_crc2 := binary.BigEndian.Uint32(mmap_data[pos : pos+4])
	if computed_mph_crc32 != mph_crc2 {
		err = fmt.Errorf("NewSSTableV1 - mph crc32 checksum failed - computed %d vs mph %d", computed_mph_crc32, mph_crc2)
		return
	}
	pos += 4

	////////////////////////////////////////
	// start of record

	t.record_start_pos = uint32(pos)

	////////////////////////////////////////
	// return parsed SSTableV1

	return t, nil
}

func (t *SSTableV1) Version() uint32 {
	return 1
}

func (t *SSTableV1) ConsensusID() util.IConsensusID {
	return t.consensus_id
}

func (t *SSTableV1) Domain() []byte {
	return t.domain.Data()
}

func (t *SSTableV1) Table() []byte {
	return t.table.Data()
}

func (t *SSTableV1) StartTime() util.IConsensusTime {
	return t.start_time
}

func (t *SSTableV1) EndTime() util.IConsensusTime {
	return t.end_time
}

func (t *SSTableV1) StartKey() util.IData {
	return t.start_key
}

func (t *SSTableV1) EndKey() util.IData {
	return t.end_key
}

func (t *SSTableV1) Level() uint32 {
	return t.level
}

func (t *SSTableV1) Get(key, group *util.IData) (util.IRecord, error) {
	return nil, fmt.Errorf("TODO")
}

func (t *SSTableV1) Groups(key *util.IData) ([]util.IData, error) {
	return nil, fmt.Errorf("TODO")
}

func (t *SSTableV1) Keys(key *util.IData) ([]util.IData, error) {
	return nil, fmt.Errorf("TODO")
}

func (t *SSTableV1) Read(pos, suggest_offset uint32) ([]util.IRecord, uint32, error) {
	return nil, 0, fmt.Errorf("TODO")
}

func (t *SSTableV1) Close() error {
	if t.mmap_data != nil {
		r := t.mmap_data.Unmap()
		if r == nil {
			t.mmap_data = nil
			return r
		} else {
			return r
		}
	}
	return nil
}
