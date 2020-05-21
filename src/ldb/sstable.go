package ldb

import (
    "fmt"
    "golang.org/x/exp/mmap"
    "../util"
)

////////////////////////////////////////////////////////////////////////////////
// Interface

type ISSTable interface {
    // SSTable Level Attributes
    Version()                           uint32                          // Version
    ConsensusID()                       *util.ConsensusID               // Consensus ID
    Domain()                            []byte                          // Domain Name
    Table()                             []byte                          // Table Name
    StartTime()                         util.IConsensusTime             // Start Time
    EndTime()                           util.IConsensusTime             // End Time
    StartKey()                          util.IData                      // Start Key
    EndKey()                            util.IData                      // End Key
    Level()                             uint32                          // Level
    // Record Operations
    Get(key, group *util.IData)         (util.IRecord, error)           // get record with specified key and attribute group
    Groups(key *util.IData)             ([]util.IData, error)           // retrieve a list of attribute groups for the given key
    Keys(key *util.IData)               ([]util.IData, error)           // retrieve a list of keys with the given key as prefix
    Read(pos, suggest_offset uint32)    ([]util.IRecord, uint32, error) // Batch scan and read operation, return a list of IRecord, bytes read, or error
}

////////////////////////////////////////////////////////////////////////////////
// Interface

type SSTableV1Lookup struct {
    pos         uint32
    len         uint16
}

type SSTableV1 struct {
    // basic attributes
    filepath            string
    consensus_id        *util.ConsensusID
    domain              util.IData
    table               util.IData
    start_time          util.IConsensusTime
    end_time            util.IConsensusTime
    start_key           util.IData
    end_key             util.IData
    level               uint32
    // lookup table
    lookup_scheme       []byte                  // perfect hash scheme represented as byte array
    lookup_table        []SSTableV1Lookup       // lookup table
    // record reader
    reader              *mmap.ReaderAt
}

func NewSSTableV1() (*SSTableV1, error) {
    return nil, fmt.Errorf("TODO")
}

func (t *SSTableV1) Version() uint32 {
    return 1
}

func (t *SSTableV1) ConsensusID() *util.ConsensusID {
    return t.consensus_id
}

func (t *SSTableV1) Domain() []byte {
    return t.domain.Content()
}

func (t *SSTableV1) Table() []byte {
    return t.table.Content()
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
