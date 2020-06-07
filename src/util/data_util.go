package util

import (
	"fmt"
	//"time"
	//"math/big"
	"encoding/binary"

	"../collection"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces
////////////////////////////////////////////////////////////////////////////////

type IData interface {

	////////////////////////////////////////
	// embeded interfaces
	//collection.IPrintable
	IEncodable

	////////////////////////////////////////
	// accessor to elements
	IsNil() bool                        // whether Data is nil
	IsPrimitive() bool                  // whether this is Primitive Data
	IsDataArray() bool                  // whether this is Data Array
	IsRecordList() bool                 // whether this is Record List
	Size() uint16                       // size of the Data Array or Record List
	DataAt(i uint16) (IData, error)     // get i-th Data Element - for Data Array only
	RecordAt(i uint16) (IRecord, error) // get i-th Record Element - for Record List only
	LookupEncoder() ILookupEncoder      // get Lookup Encoder
	CompressEncoder() ICompressEncoder  // get Compress Encoder
	Data() []byte                       // get Data Content - for Primitive Data only

	////////////////////////////////////////
	// encoding, decoding, and buf
	DataMagic() byte // 1 byte Data Magic            - return 0xff if not encoded
}

type ILookupEncoder interface {
	EncodeLookup(data []byte) ([]byte, error)
	DecodeLookup(data []byte) ([]byte, error)
}

type ICompressEncoder interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
}

////////////////////////////////////////////////////////////////////////////////
// SimpleMappedData
////////////////////////////////////////////////////////////////////////////////

type SimpleMappedData struct {
	// buf
	decoded bool
	buf     []byte
	// content
	content []byte
}

////////////////////////////////////////
// constructor

func NewSimpleMappedData(buf []byte) (*SimpleMappedData, int, error) {

	result := &SimpleMappedData{decoded: false, buf: buf}
	length, err := result.Decode(nil)
	if err != nil {
		return nil, length, err
	} else {
		return result, length, nil
	}
}

////////////////////////////////////////
// accessor to elements

func (d *SimpleMappedData) IsNil() bool {
	return len(d.content) == 0
}

func (d *SimpleMappedData) IsPrimitive() bool {
	return true
}

func (d *SimpleMappedData) IsDataArray() bool {
	return false
}

func (d *SimpleMappedData) IsRecordList() bool {
	return false
}

func (d *SimpleMappedData) Size() uint16 {
	return 0
}

func (d *SimpleMappedData) DataAt(i uint16) (IData, error) {
	return nil, fmt.Errorf("SimpleMappedData::DataAt - no array element")
}

func (d *SimpleMappedData) RecordAt(i uint16) (IRecord, error) {
	return nil, fmt.Errorf("SimpleMappedData::RecordAt - no record element")
}

func (d *SimpleMappedData) LookupEncoder() ILookupEncoder {
	return nil
}

func (d *SimpleMappedData) CompressEncoder() ICompressEncoder {
	return nil
}

func (d *SimpleMappedData) Data() []byte {
	return d.content
}

////////////////////////////////////////
// encoding, decoding, and buf

func (d *SimpleMappedData) DataMagic() byte {
	return d.buf[0]
}

func (d *SimpleMappedData) Buf() []byte {
	return d.buf
}

func (d *SimpleMappedData) EstBufSize() int {
	return len(d.buf)
}

func (d *SimpleMappedData) IsEncoded() bool {
	return true
}

func (d *SimpleMappedData) Encode(IContext) error {
	return fmt.Errorf("SimpleMappedData::Encode - encode not supported")
}

func (d *SimpleMappedData) IsDecoded() bool {
	return d.decoded
}

func (d *SimpleMappedData) Decode(IContext) (int, error) {

	data, length, err := DecodeVarchar(d.buf)
	if err != nil {
		return length, fmt.Errorf("SimpleMappedData::Decode - error %v", err)
	}

	// fill in content and buf
	d.content = data
	d.buf = d.buf[:length]
	d.decoded = true

	return length, nil
}

////////////////////////////////////////
// deep copy

func (d *SimpleMappedData) Copy() IEncodable {
	// make a deep copy of the buf
	buf := make([]byte, len(d.buf))
	copy(buf, d.buf)
	result, _, err := NewSimpleMappedData(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("SimpleMappedData:Copy - error %v", err))
	}
	return result
}

func (d *SimpleMappedData) CopyConstruct() (IEncodable, error) {
	// make a deep copy of the buf
	buf := make([]byte, len(d.content))
	copy(buf, d.content)
	result := NewPrimitive(buf)
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
// StandardMappedData
////////////////////////////////////////////////////////////////////////////////

type StandardMappedData struct {
	// buf
	decoded bool
	buf     []byte
	// elements
	data_array  []IData
	record_list []IRecord
	size        uint16
	lookup      ILookupEncoder
	compression ICompressEncoder
	content     []byte
}

////////////////////////////////////////
// constructor

func NewStandardMappedData(buf []byte) (*StandardMappedData, int, error) {

	if len(buf) < 1 {
		return nil, 0, fmt.Errorf("NewStandardMappedData - invalid empty buf")
	}

	d := &StandardMappedData{decoded: false, buf: buf}

	// decode
	length, err := d.Decode(nil)
	if err != nil {
		return nil, length, err
	}

	return d, length, nil
}

////////////////////////////////////////
// accessor to elements

func (d *StandardMappedData) IsNil() bool {
	return d.buf[0] == 0x00
}

func (d *StandardMappedData) IsPrimitive() bool {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedData:IsPrimitive - not decoded"))
	}

	return d.data_array == nil && d.record_list == nil
}

func (d *StandardMappedData) IsDataArray() bool {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedData:IsDataArray - not decoded"))
	}

	return d.data_array != nil
}

func (d *StandardMappedData) IsRecordList() bool {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedData:IsRecordList - not decoded"))
	}

	return d.record_list != nil
}

func (d *StandardMappedData) Size() uint16 {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedData:Size - not decoded"))
	}

	return d.size
}

func (d *StandardMappedData) DataAt(idx uint16) (IData, error) {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedData:DataAt - not decoded"))
	}

	if !d.IsDataArray() {
		return nil, fmt.Errorf("StandardMappedData::DataAt - not data array")
	}

	if idx >= d.Size() {
		return nil, fmt.Errorf("StandardMappedData::DataAt - idx [%d] bigger than size [%d]", idx, d.size)
	}

	if d.data_array[idx] != nil {
		return d.data_array[idx], nil
	}

	var err error
	pos := 0
	for i := uint16(0); i <= idx; i++ {
		if d.data_array[i] == nil {
			if len(d.content) < pos {
				return nil, fmt.Errorf("StandardMappedData:DataAt[%d] - invalid content %d - %d, %x", idx, i, len(d.content), d.content)
			}
			d.data_array[i], _, err = NewStandardMappedData(d.content[pos:])
			if err != nil {
				return nil, err
			}
		}
		pos += len(d.data_array[i].Buf())
		//pos += length
	}

	return d.data_array[idx], nil
}

func (d *StandardMappedData) RecordAt(idx uint16) (IRecord, error) {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedData:RecordAt - not decoded"))
	}

	if !d.IsRecordList() {
		return nil, fmt.Errorf("StandardMappedData::RecordAt - not record list")
	}

	if idx >= d.Size() {
		return nil, fmt.Errorf("StandardMappedData::RecordAt - idx [%d] bigger than size [%d]", idx, d.size)
	}

	if d.record_list[idx] != nil {
		return d.record_list[idx], nil
	}

	var err error
	pos := 0
	for i := uint16(0); i <= idx; i++ {
		if d.record_list[i] == nil {
			if len(d.content) < pos {
				return nil, fmt.Errorf("StandardMappedData:RecordAt[%d] - invalid content %d - %d, %x", idx, i, len(d.content), d.content)
			}
			d.record_list[i], _, err = NewMappedRecord(d.content[pos:])
			if err != nil {
				return nil, err
			}
		}
		pos += len(d.record_list[i].Buf())
	}

	return d.record_list[idx], nil
}

func (d *StandardMappedData) LookupEncoder() ILookupEncoder {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedData:LookupEncoder - not decoded"))
	}

	return d.lookup
}

func (d *StandardMappedData) CompressEncoder() ICompressEncoder {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedData:CompressEncoder - not decoded"))
	}

	return d.compression
}

func (d *StandardMappedData) Data() []byte {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedData:Data - not decoded"))
	}

	if d.IsPrimitive() {
		return d.content
	} else {
		return nil
	}
}

////////////////////////////////////////
// encoding, decoding, and buf

func (d *StandardMappedData) DataMagic() byte {
	return d.buf[0]
}

func (d *StandardMappedData) Buf() []byte {
	return d.buf
}

func (d *StandardMappedData) EstBufSize() int {
	return len(d.buf)
}

func (d *StandardMappedData) IsEncoded() bool {
	return true
}

func (d *StandardMappedData) Encode(IContext) error {
	return fmt.Errorf("StandardMappedData:Encode - encode not supported")
}

func (d *StandardMappedData) IsDecoded() bool {
	return d.decoded
}

func (d *StandardMappedData) Decode(IContext) (int, error) {

	format_is_set := false

	pos := 1
	content_length := uint16(0)

	// process data array
	data_array_bits := (d.buf[0] >> 6) & 0x03
	switch data_array_bits {
	case 0x00:
		break
	case 0x01:
		if len(d.buf) < 1+int(pos) {
			return 0, fmt.Errorf("StandardMappedData::Decode - invalid buf 1, data array no size, %d, %x", len(d.buf), d.buf)
		}
		format_is_set = true
		d.size = uint16(d.buf[pos])
		d.data_array = make([]IData, d.size)
		pos += 1
	case 0x02:
		if len(d.buf) < 2+int(pos) {
			return 0, fmt.Errorf("StandardMappedData::Decode - invalid buf 2, data array no size, %d, %x", len(d.buf), d.buf)
		}
		format_is_set = true
		d.size = binary.BigEndian.Uint16(d.buf[pos:])
		d.data_array = make([]IData, d.size)
		pos += 2
	case 0x03:
		return 0, fmt.Errorf("StandardMappedData::Decode - invalid magic - data array: %x", d.buf[0])
	}

	// process record list
	record_list_bits := (d.buf[0] >> 4) & 0x03
	switch record_list_bits {
	case 0x00:
		break
	case 0x01:
		if format_is_set {
			return 0, fmt.Errorf("StandardMappedData::Decode - invalid magic [%x] - format set prior to record list", d.buf[0])
		}
		if len(d.buf) < 1+int(pos) {
			return 0, fmt.Errorf("StandardMappedData::Decode - invalid buf 1, record list no size, %d, %x", len(d.buf), d.buf)
		}
		format_is_set = true
		d.size = uint16(d.buf[pos])
		d.record_list = make([]IRecord, d.size)
		pos += 1
	case 0x02:
		if format_is_set {
			return 0, fmt.Errorf("StandardMappedData::Decode - invalid magic [%x] - format set prior to record list", d.buf[0])
		}
		if len(d.buf) < 2+int(pos) {
			return 0, fmt.Errorf("StandardMappedData::Decode - invalid buf 2, record list no size, %d, %x", len(d.buf), d.buf)
		}
		format_is_set = true
		d.size = binary.BigEndian.Uint16(d.buf[pos:])
		d.record_list = make([]IRecord, d.size)
		pos += 2
	case 0x03:
		return 0, fmt.Errorf("StandardMappedData::Decode - invalid magic [%x] - record list", d.buf[0])
	}

	// process lookup
	lookup_bit := (d.buf[0] >> 3) & 0x01
	switch lookup_bit {
	case 0x00:
		d.lookup = nil
	case 0x01:
		// pos      += 2
		return 0, fmt.Errorf("StandardMappedData::Decode - invalid magic [%x] - lookup not supported", d.buf[0])
	}

	// process compression
	compression_bit := (d.buf[0] >> 2) & 0x01
	switch compression_bit {
	case 0x00:
		d.compression = nil
	case 0x01:
		// pos      += 2
		return 0, fmt.Errorf("StandardMappedData::Decode - invalid magic [%x] - compression not supported", d.buf[0])
	}

	// process content
	length_bits := d.buf[0] & 0x03
	switch length_bits {
	case 0x00:
		break
	case 0x01:
		if len(d.buf) < 1+int(pos) {
			return 0, fmt.Errorf("StandardMappedData::Decode - invalid buf 1, no length, %d, %x", len(d.buf), d.buf)
		}
		content_length = uint16(d.buf[pos])
		if len(d.buf) < 1+int(pos)+int(content_length) {
			return 0, fmt.Errorf("StandardMappedData::Decode - invalid buf 1, missing content, %d, %x", len(d.buf), d.buf)
		}
		pos += 1 + int(content_length)
	case 0x02:
		if len(d.buf) < 2+int(pos) {
			return 0, fmt.Errorf("StandardMappedData::Decode - invalid buf 2, no length, %d, %x", len(d.buf), d.buf)
		}
		content_length = binary.BigEndian.Uint16(d.buf[pos:])
		if len(d.buf) < 2+int(pos)+int(content_length) {
			return 0, fmt.Errorf("StandardMappedData::Decode - invalid buf 2, missing content, %d, %x", len(d.buf), d.buf)
		}
		pos += 2 + int(content_length)
	case 0x03:
		return 0, fmt.Errorf("StandardMappedData::Decode - invalid magic [%x] - length ", d.buf[0])
	}

	d.content = d.buf[pos-int(content_length) : pos]
	d.buf = d.buf[:pos]
	d.decoded = true

	return pos, nil
}

////////////////////////////////////////
// deep copy

func (d *StandardMappedData) Copy() IEncodable {
	// make a deep copy of the buf
	buf := make([]byte, len(d.buf))
	copy(buf, d.buf)
	result, _, err := NewStandardMappedData(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("StandardMappedData:Copy - %s", err))
	}
	return result
}

func (d *StandardMappedData) CopyConstruct() (IEncodable, error) {

	if d.IsPrimitive() {

		buf := make([]byte, len(d.Data()))
		copy(buf, d.Data())

		result := NewPrimitive(buf)

		return result, nil

	} else if d.IsDataArray() {

		result := NewDataArray()

		for i := uint16(0); i < d.Size(); i++ {

			data, err := d.DataAt(i)
			if err != nil {
				return nil, err
			}

			data_copy, err := data.CopyConstruct()
			if err != nil {
				return nil, err
			}

			result.Append(data_copy.(IData))
		}

		return result, nil

	} else if d.IsRecordList() {

		result := NewRecordList()

		for i := uint16(0); i < d.Size(); i++ {

			record, err := d.RecordAt(i)
			if err != nil {
				return nil, err
			}

			record_copy, err := record.CopyConstruct()
			if err != nil {
				return nil, err
			}

			result.Append(record_copy.(IRecord))
		}

		return result, nil

	} else {

		return nil, fmt.Errorf("StandardMappedData::CopyConstruct - unsupported type %x", d.DataMagic())
	}
}

////////////////////////////////////////////////////////////////////////////////
// Primitive
////////////////////////////////////////////////////////////////////////////////

type Primitive struct {
	// buf
	encoded bool
	magic   byte
	buf     []byte
	// data
	data []byte
}

////////////////////////////////////////
// constructor

func NewPrimitive(data []byte) *Primitive {
	return &Primitive{encoded: false, data: data}
}

////////////////////////////////////////
// accessor to elements

func (d *Primitive) IsNil() bool {
	return d.data == nil || len(d.data) == 0
}

func (d *Primitive) IsPrimitive() bool {
	return true
}

func (d *Primitive) IsDataArray() bool {
	return false
}

func (d *Primitive) IsRecordList() bool {
	return false
}

func (d *Primitive) Size() uint16 {
	return uint16(0)
}

func (d *Primitive) DataAt(idx uint16) (IData, error) {

	return nil, fmt.Errorf("Primitive::DataAt - not allowed for primitive data")
}

func (d *Primitive) RecordAt(idx uint16) (IRecord, error) {

	return nil, fmt.Errorf("Primitive::RecordAt - not allowed for primitive data")
}

func (d *Primitive) LookupEncoder() ILookupEncoder {
	return nil
}

func (d *Primitive) CompressEncoder() ICompressEncoder {
	return nil
}

func (d *Primitive) Data() []byte {

	return d.data
}

////////////////////////////////////////
// encoding, decoding, and buf

func (d *Primitive) DataMagic() byte {

	if !d.encoded {
		panic(fmt.Sprintf("Primitive::DataMagic - not encoded"))
	}

	return d.magic
}

func (d *Primitive) Buf() []byte {

	if !d.encoded {
		panic(fmt.Sprintf("Primitive::DataMagic - not encoded"))
	}

	return d.buf
}

func (d *Primitive) EstBufSize() int {
	if len(d.data) < 1<<7 {
		return 1 + len(d.data)
	} else if len(d.data) < 1<<15 {
		return 2 + len(d.data)
	} else if len(d.data) < 1<<23 {
		return 3 + len(d.data)
	} else {
		return 4 + len(d.data)
	}
}

func (d *Primitive) IsEncoded() bool {
	return d.encoded
}

func (d *Primitive) Encode(IContext) error {

	if d.data == nil {
		d.magic = 0x00
		d.buf = nil
		d.encoded = true
		return nil
	}

	// encode content length
	content_len := len(d.data)

	if content_len == 0 {

		buf := []byte{0x00}
		d.magic = 0x00
		//d.magic = 0xff & 0x03
		d.buf = buf
		d.encoded = true
		return nil

	} else if content_len < 256 {

		length_buf := []byte{uint8(content_len)}
		buf := append([]byte{0x01}, length_buf...)
		buf = append(buf, d.data...)
		d.magic = 0x01
		d.buf = buf
		d.encoded = true
		return nil

	} else if content_len < 65536 {

		// length_bits = 0x02
		length_buf := []byte{uint8(content_len >> 8), uint8(content_len)} // BigEndian encoding
		buf := append([]byte{0x02}, length_buf...)
		buf = append(buf, d.data...)
		d.magic = 0x02
		d.buf = buf
		d.encoded = true
		return nil

	} else {

		return fmt.Errorf("Primitive::Encode - content length too big %d", content_len)
	}
}

func (d *Primitive) IsDecoded() bool {
	return true
}

func (d *Primitive) Decode(IContext) (int, error) {
	return 0, fmt.Errorf("Primitive::Decode - decode not supported")
}

////////////////////////////////////////
// deep copy

func (d *Primitive) Copy() IEncodable {

	c := NewPrimitive(d.data)
	if d.data == nil {
		return c
	}

	// make a deep copy of the buf
	c.data = make([]byte, len(d.data))
	copy(c.data, d.data)

	return c
}

func (d *Primitive) CopyConstruct() (IEncodable, error) {

	return d.Copy(), nil
}

////////////////////////////////////////////////////////////////////////////////
// DataArray
////////////////////////////////////////////////////////////////////////////////

type DataArray struct {
	// buf
	encoded    bool
	buf        []byte
	estBufSize int
	// data array
	data_array []IData
}

////////////////////////////////////////
// constructor

func NewDataArray() *DataArray {
	return &DataArray{encoded: false, data_array: []IData{}, estBufSize: 1}
}

////////////////////////////////////////
// accessor to elements

func (d *DataArray) IsNil() bool {
	return d.data_array == nil || len(d.data_array) == 0
}

func (d *DataArray) IsPrimitive() bool {
	return d.IsNil() || false
}

func (d *DataArray) IsDataArray() bool {
	return true
}

func (d *DataArray) IsRecordList() bool {
	return false
}

func (d *DataArray) Size() uint16 {
	return uint16(len(d.data_array))
}

func (d *DataArray) DataAt(idx uint16) (IData, error) {

	if idx >= uint16(len(d.data_array)) {
		return nil, fmt.Errorf("DataArray::DataAt - idx [%d] bigger than size [%d]", idx, len(d.data_array))
	}

	return d.data_array[idx], nil
}

func (d *DataArray) RecordAt(idx uint16) (IRecord, error) {

	return nil, fmt.Errorf("DataArray::RecordAt - not allowed for data array")
}

func (d *DataArray) LookupEncoder() ILookupEncoder {
	return nil
}

func (d *DataArray) CompressEncoder() ICompressEncoder {
	return nil
}

func (d *DataArray) Data() []byte {
	return nil
}

////////////////////////////////////////
// encoding, decoding, and buf

func (d *DataArray) DataMagic() byte {

	if !d.encoded {
		panic(fmt.Sprintf("DataArray::DataMagic - not encoded"))
	}

	return d.buf[0]
}

func (d *DataArray) Buf() []byte {

	if !d.encoded {
		panic(fmt.Sprintf("DataArray::Buf - not encoded"))
	}

	return d.buf
}

func (d *DataArray) EstBufSize() int {
	if d.estBufSize <= 0 {
		return 1
	} else {
		return d.estBufSize
	}
}

func (d *DataArray) IsEncoded() bool {
	return d.encoded
}

func (d *DataArray) Encode(IContext) error {

	if d.data_array == nil {
		d.estBufSize = 1
		return nil
	}

	buf := []byte{0x00}

	// encode size
	size := len(d.data_array)
	if size == 0 {
		buf[0] |= 0x00
	} else if size < 256 {
		buf[0] |= 0x01 << 6
		buf = append(buf, uint8(size))
	} else if size < 65536 {
		buf[0] |= 0x02 << 6
		buf = append(buf, uint8(size>>8), uint8(size)) // BigEndian encoding
	} else {
		return fmt.Errorf("DataArray::Encode - unexpected size %d", size)
	}

	// encode content
	content_buf := []byte{}
	for i := 0; i < len(d.data_array); i++ {
		err := d.data_array[i].Encode(nil)
		if err != nil {
			return fmt.Errorf("DataArray::Encode - err [%v]", err)
		}
		content_buf = append(content_buf, d.data_array[i].Buf()...)
	}

	content_len := len(content_buf)
	if content_len == 0 {

		d.buf = buf
		d.encoded = true
		d.estBufSize = len(d.buf)
		return nil

	} else if content_len < 256 {

		buf[0] |= 0x01
		buf = append(buf, uint8(content_len))
		buf = append(buf, content_buf...)
		d.buf = buf
		d.encoded = true
		d.estBufSize = len(d.buf)
		return nil

	} else if content_len < 65536 {

		buf[0] |= 0x02
		buf = append(buf, uint8(content_len>>8), uint8(content_len))
		buf = append(buf, content_buf...)
		d.buf = buf
		d.encoded = true
		d.estBufSize = len(d.buf)
		return nil

	} else {

		return fmt.Errorf("DataArray::Encode - content length too big %d", content_len)

	}
}

func (d *DataArray) IsDecoded() bool {
	return true
}

func (d *DataArray) Decode(IContext) (int, error) {
	return 0, fmt.Errorf("DataArray::Decode - decode not supported")
}

////////////////////////////////////////
// deep copy

func (d *DataArray) Copy() IEncodable {

	c := NewDataArray()
	if d.data_array == nil {
		return c
	}

	// make a deep copy of the buf
	c.data_array = make([]IData, len(d.data_array))
	for i := 0; i < len(d.data_array); i++ {
		c.data_array[i] = d.data_array[i].Copy().(IData)
	}

	return c
}

func (d *DataArray) CopyConstruct() (IEncodable, error) {

	c := NewDataArray()
	if d.data_array == nil {
		return c, nil
	}

	// make a deep copy of the buf
	c.data_array = make([]IData, len(d.data_array))
	for i := 0; i < len(d.data_array); i++ {
		data_i, err := d.data_array[i].CopyConstruct()
		if err != nil {
			return nil, err
		} else {
			c.data_array[i] = data_i.(IData)
		}
	}

	return c, nil
}

////////////////////////////////////////
// updater

func (d *DataArray) Append(data IData) *DataArray {
	d.data_array = append(d.data_array, data)
	d.encoded = false
	d.estBufSize += data.EstBufSize()
	return d
}

func (d *DataArray) DeleteAt(idx uint16) *DataArray {

	if idx >= uint16(len(d.data_array)) {
		panic(fmt.Sprintf("DataArray::DataAt - idx [%d] bigger than size [%d]", idx, len(d.data_array)))
	}

	d.estBufSize -= d.data_array[idx].EstBufSize()
	d.data_array = append(d.data_array[:idx], d.data_array[idx+1:]...)
	d.encoded = false
	return d
}

////////////////////////////////////////////////////////////////////////////////
// Constructed Record List

type RecordList struct {
	// buf
	encoded    bool
	buf        []byte
	estBufSize int
	// record list
	record_list []IRecord
}

////////////////////////////////////////
// constructor

func NewRecordList() *RecordList {
	return &RecordList{encoded: false, record_list: []IRecord{}, estBufSize: 1}
}

////////////////////////////////////////
// accessor to elements

func (d *RecordList) IsNil() bool {
	return d.record_list == nil || len(d.record_list) == 0
}

func (d *RecordList) IsPrimitive() bool {
	return d.IsNil() || false
}

func (d *RecordList) IsDataArray() bool {
	return false
}

func (d *RecordList) IsRecordList() bool {
	return true
}

func (d *RecordList) Size() uint16 {
	return uint16(len(d.record_list))
}

func (d *RecordList) DataAt(idx uint16) (IData, error) {

	return nil, fmt.Errorf("RecordList::DataAt - not allowed for record list")
}

func (d *RecordList) RecordAt(idx uint16) (IRecord, error) {

	if idx >= uint16(len(d.record_list)) {
		return nil, fmt.Errorf("RecordList::RecordAt - idx [%d] bigger than size [%d]", idx, len(d.record_list))
	}

	return d.record_list[idx], nil
}

func (d *RecordList) LookupEncoder() ILookupEncoder {
	return nil
}

func (d *RecordList) CompressEncoder() ICompressEncoder {
	return nil
}

func (d *RecordList) Data() []byte {
	return nil
}

////////////////////////////////////////
// encoding, decoding, and buf

func (d *RecordList) DataMagic() byte {

	if !d.encoded {
		panic(fmt.Sprintf("RecordList::DataMagic - not encoded"))
	}

	return d.buf[0]
}

func (d *RecordList) Buf() []byte {

	if !d.encoded {
		panic(fmt.Sprintf("RecordList::DataMagic - not encoded"))
	}

	return d.buf
}

func (d *RecordList) EstBufSize() int {
	if d.estBufSize <= 0 {
		return 1
	} else {
		return d.estBufSize
	}
}

func (d *RecordList) IsEncoded() bool {
	return d.encoded
}

func (d *RecordList) Encode(IContext) error {

	if d.record_list == nil {
		return nil
	}

	buf := []byte{0x00}

	// encode size
	size := len(d.record_list)
	if size == 0 {
		buf[0] |= 0x00
	} else if size < 256 {
		buf[0] |= 0x01 << 4
		buf = append(buf, uint8(size))
	} else if size < 65536 {
		buf[0] |= 0x02 << 4
		buf = append(buf, uint8(size>>8), uint8(size)) // BigEndian encoding
	} else {
		return fmt.Errorf("RecordList::Encode - unexpected size %d", size)
	}

	// encode content
	content_buf := []byte{}
	for i := 0; i < len(d.record_list); i++ {
		err := d.record_list[i].Encode(nil)

		if err != nil {
			return fmt.Errorf("RecordList::Encode - error [%v]", err)
		}
		content_buf = append(content_buf, d.record_list[i].Buf()...)
	}

	content_len := len(content_buf)
	if content_len == 0 {

		d.buf = buf
		d.encoded = true
		return nil

	} else if content_len < 256 {

		buf[0] |= 0x01
		buf = append(buf, uint8(content_len))
		buf = append(buf, content_buf...)
		d.buf = buf
		d.encoded = true
		return nil

	} else if content_len < 65536 {

		buf[0] |= 0x02
		buf = append(buf, uint8(content_len>>8), uint8(content_len))
		buf = append(buf, content_buf...)
		d.buf = buf
		d.encoded = true
		return nil

	} else {

		return fmt.Errorf("RecordList::Encode - content length too big %d", content_len)

	}
}

func (d *RecordList) IsDecoded() bool {
	return true
}

func (d *RecordList) Decode(IContext) (int, error) {
	return 0, fmt.Errorf("RecordList::Decode - decode not supported")
}

////////////////////////////////////////
// deep copy

func (d *RecordList) Copy() IEncodable {

	c := NewRecordList()
	if d.record_list == nil {
		return c
	}

	// make a deep copy of the buf
	c.record_list = make([]IRecord, len(d.record_list))
	for i := 0; i < len(d.record_list); i++ {
		c.record_list[i] = d.record_list[i].Copy().(IRecord)
	}

	return c
}

func (d *RecordList) CopyConstruct() (IEncodable, error) {

	c := NewRecordList()
	if d.record_list == nil {
		return c, nil
	}

	// make a deep copy of the buf
	c.record_list = make([]IRecord, len(d.record_list))
	for i := 0; i < len(d.record_list); i++ {
		record_i, err := d.record_list[i].CopyConstruct()
		if err != nil {
			return nil, err
		} else {
			c.record_list[i] = record_i.(IRecord)
		}
	}

	return c, nil
}

////////////////////////////////////////
// updater

func (d *RecordList) Append(record IRecord) *RecordList {
	d.record_list = append(d.record_list, record)
	d.encoded = false
	d.estBufSize += record.EstBufSize()
	return d
}

func (d *RecordList) DeleteAt(idx uint16) *RecordList {

	if idx >= uint16(len(d.record_list)) {
		panic(fmt.Sprintf("DataArray::DeleteAt - idx [%d] bigger than size [%d]", idx, len(d.record_list)))
	}

	d.estBufSize -= d.record_list[idx].EstBufSize()
	d.record_list = append(d.record_list[:idx], d.record_list[idx+1:]...)
	d.encoded = false
	return d
}

////////////////////////////////////////////////////////////////////////////////
// utilities
////////////////////////////////////////////////////////////////////////////////

func estimateDataBufSize(d IData) int {
	if collection.IsNil(d) {
		return 1
	} else {
		return d.EstBufSize()
	}
}
