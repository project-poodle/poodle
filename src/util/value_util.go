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

type IValue interface {

	////////////////////////////////////////
	// embeded interfaces
	//collection.IPrintable
	IEncodable

	////////////////////////////////////////
	// accessor to elements
	IsNil() bool                        // whether Value is nil
	IsPrimitive() bool                  // whether this is Primitive Value
	IsValueArray() bool                 // whether this is Value Array
	IsRecordList() bool                 // whether this is Record List
	Size() uint16                       // size of the Value Array or Record List
	ValueAt(i uint16) (IValue, error)   // get i-th Value Element - for Value Array only
	RecordAt(i uint16) (IRecord, error) // get i-th Record Element - for Record List only
	LookupEncoder() ILookupEncoder      // get Lookup Encoder
	CompressEncoder() ICompressEncoder  // get Compress Encoder
	Value() []byte                      // get Value Content - for Primitive Value only

	////////////////////////////////////////
	// encoding, decoding, and buf
	ValueMagic() byte // 1 byte Value Magic            - return 0xff if not encoded
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
// SimpleMappedValue
////////////////////////////////////////////////////////////////////////////////

type SimpleMappedValue struct {
	// buf
	decoded bool
	buf     []byte
	// content
	content []byte
}

////////////////////////////////////////
// constructor

func NewSimpleMappedValue(buf []byte) (*SimpleMappedValue, int, error) {

	result := &SimpleMappedValue{decoded: false, buf: buf}
	length, err := result.Decode(nil)
	if err != nil {
		return nil, length, err
	} else {
		return result, length, nil
	}
}

////////////////////////////////////////
// accessor to elements

func (d *SimpleMappedValue) IsNil() bool {
	return len(d.content) == 0
}

func (d *SimpleMappedValue) IsPrimitive() bool {
	return true
}

func (d *SimpleMappedValue) IsValueArray() bool {
	return false
}

func (d *SimpleMappedValue) IsRecordList() bool {
	return false
}

func (d *SimpleMappedValue) Size() uint16 {
	return 0
}

func (d *SimpleMappedValue) ValueAt(i uint16) (IValue, error) {
	return nil, fmt.Errorf("SimpleMappedValue::ValueAt - no array element")
}

func (d *SimpleMappedValue) RecordAt(i uint16) (IRecord, error) {
	return nil, fmt.Errorf("SimpleMappedValue::RecordAt - no record element")
}

func (d *SimpleMappedValue) LookupEncoder() ILookupEncoder {
	return nil
}

func (d *SimpleMappedValue) CompressEncoder() ICompressEncoder {
	return nil
}

func (d *SimpleMappedValue) Value() []byte {
	return d.content
}

////////////////////////////////////////
// encoding, decoding, and buf

func (d *SimpleMappedValue) ValueMagic() byte {
	return d.buf[0]
}

func (d *SimpleMappedValue) Buf() []byte {
	return d.buf
}

func (d *SimpleMappedValue) EstBufSize() int {
	return len(d.buf)
}

func (d *SimpleMappedValue) IsEncoded() bool {
	return true
}

func (d *SimpleMappedValue) Encode(IContext) error {
	return fmt.Errorf("SimpleMappedValue::Encode - encode not supported")
}

func (d *SimpleMappedValue) IsDecoded() bool {
	return d.decoded
}

func (d *SimpleMappedValue) Decode(IContext) (int, error) {

	data, length, err := DecodeVarchar(d.buf)
	if err != nil {
		return length, fmt.Errorf("SimpleMappedValue::Decode - error %v", err)
	}

	// fill in content and buf
	d.content = data
	d.buf = d.buf[:length]
	d.decoded = true

	return length, nil
}

////////////////////////////////////////
// deep copy

func (d *SimpleMappedValue) Copy() IEncodable {
	// make a deep copy of the buf
	buf := make([]byte, len(d.buf))
	copy(buf, d.buf)
	result, _, err := NewSimpleMappedValue(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("SimpleMappedValue:Copy - error %v", err))
	}
	return result
}

func (d *SimpleMappedValue) CopyConstruct() (IEncodable, error) {
	// make a deep copy of the buf
	buf := make([]byte, len(d.content))
	copy(buf, d.content)
	result := NewPrimitive(buf)
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
// StandardMappedValue
////////////////////////////////////////////////////////////////////////////////

type StandardMappedValue struct {
	// buf
	decoded bool
	buf     []byte
	// elements
	data_array  []IValue
	record_list []IRecord
	size        uint16
	lookup      ILookupEncoder
	compression ICompressEncoder
	content     []byte
}

////////////////////////////////////////
// constructor

func NewStandardMappedValue(buf []byte) (*StandardMappedValue, int, error) {

	if len(buf) < 1 {
		return nil, 0, fmt.Errorf("NewStandardMappedValue - invalid empty buf")
	}

	d := &StandardMappedValue{decoded: false, buf: buf}

	// decode
	length, err := d.Decode(nil)
	if err != nil {
		return nil, length, err
	}

	return d, length, nil
}

////////////////////////////////////////
// accessor to elements

func (d *StandardMappedValue) IsNil() bool {
	return d.buf[0] == 0x00
}

func (d *StandardMappedValue) IsPrimitive() bool {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedValue:IsPrimitive - not decoded"))
	}

	return d.data_array == nil && d.record_list == nil
}

func (d *StandardMappedValue) IsValueArray() bool {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedValue:IsValueArray - not decoded"))
	}

	return d.data_array != nil
}

func (d *StandardMappedValue) IsRecordList() bool {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedValue:IsRecordList - not decoded"))
	}

	return d.record_list != nil
}

func (d *StandardMappedValue) Size() uint16 {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedValue:Size - not decoded"))
	}

	return d.size
}

func (d *StandardMappedValue) ValueAt(idx uint16) (IValue, error) {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedValue:ValueAt - not decoded"))
	}

	if !d.IsValueArray() {
		return nil, fmt.Errorf("StandardMappedValue::ValueAt - not data array")
	}

	if idx >= d.Size() {
		return nil, fmt.Errorf("StandardMappedValue::ValueAt - idx [%d] bigger than size [%d]", idx, d.size)
	}

	if d.data_array[idx] != nil {
		return d.data_array[idx], nil
	}

	var err error
	pos := 0
	for i := uint16(0); i <= idx; i++ {
		if d.data_array[i] == nil {
			if len(d.content) < pos {
				return nil, fmt.Errorf("StandardMappedValue:ValueAt[%d] - invalid content %d - %d, %x", idx, i, len(d.content), d.content)
			}
			d.data_array[i], _, err = NewStandardMappedValue(d.content[pos:])
			if err != nil {
				return nil, err
			}
		}
		pos += len(d.data_array[i].Buf())
		//pos += length
	}

	return d.data_array[idx], nil
}

func (d *StandardMappedValue) RecordAt(idx uint16) (IRecord, error) {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedValue:RecordAt - not decoded"))
	}

	if !d.IsRecordList() {
		return nil, fmt.Errorf("StandardMappedValue::RecordAt - not record list")
	}

	if idx >= d.Size() {
		return nil, fmt.Errorf("StandardMappedValue::RecordAt - idx [%d] bigger than size [%d]", idx, d.size)
	}

	if d.record_list[idx] != nil {
		return d.record_list[idx], nil
	}

	var err error
	pos := 0
	for i := uint16(0); i <= idx; i++ {
		if d.record_list[i] == nil {
			if len(d.content) < pos {
				return nil, fmt.Errorf("StandardMappedValue:RecordAt[%d] - invalid content %d - %d, %x", idx, i, len(d.content), d.content)
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

func (d *StandardMappedValue) LookupEncoder() ILookupEncoder {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedValue:LookupEncoder - not decoded"))
	}

	return d.lookup
}

func (d *StandardMappedValue) CompressEncoder() ICompressEncoder {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedValue:CompressEncoder - not decoded"))
	}

	return d.compression
}

func (d *StandardMappedValue) Value() []byte {

	if !d.decoded {
		panic(fmt.Sprintf("StandardMappedValue:Value - not decoded"))
	}

	if d.IsPrimitive() {
		return d.content
	} else {
		return nil
	}
}

////////////////////////////////////////
// encoding, decoding, and buf

func (d *StandardMappedValue) ValueMagic() byte {
	return d.buf[0]
}

func (d *StandardMappedValue) Buf() []byte {
	return d.buf
}

func (d *StandardMappedValue) EstBufSize() int {
	return len(d.buf)
}

func (d *StandardMappedValue) IsEncoded() bool {
	return true
}

func (d *StandardMappedValue) Encode(IContext) error {
	return fmt.Errorf("StandardMappedValue:Encode - encode not supported")
}

func (d *StandardMappedValue) IsDecoded() bool {
	return d.decoded
}

func (d *StandardMappedValue) Decode(IContext) (int, error) {

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
			return 0, fmt.Errorf("StandardMappedValue::Decode - invalid buf 1, data array no size, %d, %x", len(d.buf), d.buf)
		}
		format_is_set = true
		d.size = uint16(d.buf[pos])
		d.data_array = make([]IValue, d.size)
		pos += 1
	case 0x02:
		if len(d.buf) < 2+int(pos) {
			return 0, fmt.Errorf("StandardMappedValue::Decode - invalid buf 2, data array no size, %d, %x", len(d.buf), d.buf)
		}
		format_is_set = true
		d.size = binary.BigEndian.Uint16(d.buf[pos:])
		d.data_array = make([]IValue, d.size)
		pos += 2
	case 0x03:
		return 0, fmt.Errorf("StandardMappedValue::Decode - invalid magic - data array: %x", d.buf[0])
	}

	// process record list
	record_list_bits := (d.buf[0] >> 4) & 0x03
	switch record_list_bits {
	case 0x00:
		break
	case 0x01:
		if format_is_set {
			return 0, fmt.Errorf("StandardMappedValue::Decode - invalid magic [%x] - format set prior to record list", d.buf[0])
		}
		if len(d.buf) < 1+int(pos) {
			return 0, fmt.Errorf("StandardMappedValue::Decode - invalid buf 1, record list no size, %d, %x", len(d.buf), d.buf)
		}
		format_is_set = true
		d.size = uint16(d.buf[pos])
		d.record_list = make([]IRecord, d.size)
		pos += 1
	case 0x02:
		if format_is_set {
			return 0, fmt.Errorf("StandardMappedValue::Decode - invalid magic [%x] - format set prior to record list", d.buf[0])
		}
		if len(d.buf) < 2+int(pos) {
			return 0, fmt.Errorf("StandardMappedValue::Decode - invalid buf 2, record list no size, %d, %x", len(d.buf), d.buf)
		}
		format_is_set = true
		d.size = binary.BigEndian.Uint16(d.buf[pos:])
		d.record_list = make([]IRecord, d.size)
		pos += 2
	case 0x03:
		return 0, fmt.Errorf("StandardMappedValue::Decode - invalid magic [%x] - record list", d.buf[0])
	}

	// process lookup
	lookup_bit := (d.buf[0] >> 3) & 0x01
	switch lookup_bit {
	case 0x00:
		d.lookup = nil
	case 0x01:
		// pos      += 2
		return 0, fmt.Errorf("StandardMappedValue::Decode - invalid magic [%x] - lookup not supported", d.buf[0])
	}

	// process compression
	compression_bit := (d.buf[0] >> 2) & 0x01
	switch compression_bit {
	case 0x00:
		d.compression = nil
	case 0x01:
		// pos      += 2
		return 0, fmt.Errorf("StandardMappedValue::Decode - invalid magic [%x] - compression not supported", d.buf[0])
	}

	// process content
	length_bits := d.buf[0] & 0x03
	switch length_bits {
	case 0x00:
		break
	case 0x01:
		if len(d.buf) < 1+int(pos) {
			return 0, fmt.Errorf("StandardMappedValue::Decode - invalid buf 1, no length, %d, %x", len(d.buf), d.buf)
		}
		content_length = uint16(d.buf[pos])
		if len(d.buf) < 1+int(pos)+int(content_length) {
			return 0, fmt.Errorf("StandardMappedValue::Decode - invalid buf 1, missing content, %d, %x", len(d.buf), d.buf)
		}
		pos += 1 + int(content_length)
	case 0x02:
		if len(d.buf) < 2+int(pos) {
			return 0, fmt.Errorf("StandardMappedValue::Decode - invalid buf 2, no length, %d, %x", len(d.buf), d.buf)
		}
		content_length = binary.BigEndian.Uint16(d.buf[pos:])
		if len(d.buf) < 2+int(pos)+int(content_length) {
			return 0, fmt.Errorf("StandardMappedValue::Decode - invalid buf 2, missing content, %d, %x", len(d.buf), d.buf)
		}
		pos += 2 + int(content_length)
	case 0x03:
		return 0, fmt.Errorf("StandardMappedValue::Decode - invalid magic [%x] - length ", d.buf[0])
	}

	d.content = d.buf[pos-int(content_length) : pos]
	d.buf = d.buf[:pos]
	d.decoded = true

	return pos, nil
}

////////////////////////////////////////
// deep copy

func (d *StandardMappedValue) Copy() IEncodable {
	// make a deep copy of the buf
	buf := make([]byte, len(d.buf))
	copy(buf, d.buf)
	result, _, err := NewStandardMappedValue(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("StandardMappedValue:Copy - %s", err))
	}
	return result
}

func (d *StandardMappedValue) CopyConstruct() (IEncodable, error) {

	if d.IsPrimitive() {

		buf := make([]byte, len(d.Value()))
		copy(buf, d.Value())

		result := NewPrimitive(buf)

		return result, nil

	} else if d.IsValueArray() {

		result := NewValueArray()

		for i := uint16(0); i < d.Size(); i++ {

			data, err := d.ValueAt(i)
			if err != nil {
				return nil, err
			}

			data_copy, err := data.CopyConstruct()
			if err != nil {
				return nil, err
			}

			result.Append(data_copy.(IValue))
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

		return nil, fmt.Errorf("StandardMappedValue::CopyConstruct - unsupported type %x", d.ValueMagic())
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

func (d *Primitive) IsValueArray() bool {
	return false
}

func (d *Primitive) IsRecordList() bool {
	return false
}

func (d *Primitive) Size() uint16 {
	return uint16(0)
}

func (d *Primitive) ValueAt(idx uint16) (IValue, error) {

	return nil, fmt.Errorf("Primitive::ValueAt - not allowed for primitive data")
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

func (d *Primitive) Value() []byte {

	return d.data
}

////////////////////////////////////////
// encoding, decoding, and buf

func (d *Primitive) ValueMagic() byte {

	if !d.encoded {
		panic(fmt.Sprintf("Primitive::ValueMagic - not encoded"))
	}

	return d.magic
}

func (d *Primitive) Buf() []byte {

	if !d.encoded {
		panic(fmt.Sprintf("Primitive::ValueMagic - not encoded"))
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
// ValueArray
////////////////////////////////////////////////////////////////////////////////

type ValueArray struct {
	// buf
	encoded    bool
	buf        []byte
	estBufSize int
	// data array
	data_array []IValue
}

////////////////////////////////////////
// constructor

func NewValueArray() *ValueArray {
	return &ValueArray{encoded: false, data_array: []IValue{}, estBufSize: 1}
}

////////////////////////////////////////
// accessor to elements

func (d *ValueArray) IsNil() bool {
	return d.data_array == nil || len(d.data_array) == 0
}

func (d *ValueArray) IsPrimitive() bool {
	return d.IsNil() || false
}

func (d *ValueArray) IsValueArray() bool {
	return true
}

func (d *ValueArray) IsRecordList() bool {
	return false
}

func (d *ValueArray) Size() uint16 {
	return uint16(len(d.data_array))
}

func (d *ValueArray) ValueAt(idx uint16) (IValue, error) {

	if idx >= uint16(len(d.data_array)) {
		return nil, fmt.Errorf("ValueArray::ValueAt - idx [%d] bigger than size [%d]", idx, len(d.data_array))
	}

	return d.data_array[idx], nil
}

func (d *ValueArray) RecordAt(idx uint16) (IRecord, error) {

	return nil, fmt.Errorf("ValueArray::RecordAt - not allowed for data array")
}

func (d *ValueArray) LookupEncoder() ILookupEncoder {
	return nil
}

func (d *ValueArray) CompressEncoder() ICompressEncoder {
	return nil
}

func (d *ValueArray) Value() []byte {
	return nil
}

////////////////////////////////////////
// encoding, decoding, and buf

func (d *ValueArray) ValueMagic() byte {

	if !d.encoded {
		panic(fmt.Sprintf("ValueArray::ValueMagic - not encoded"))
	}

	return d.buf[0]
}

func (d *ValueArray) Buf() []byte {

	if !d.encoded {
		panic(fmt.Sprintf("ValueArray::Buf - not encoded"))
	}

	return d.buf
}

func (d *ValueArray) EstBufSize() int {
	if d.estBufSize <= 0 {
		return 1
	} else {
		return d.estBufSize
	}
}

func (d *ValueArray) IsEncoded() bool {
	return d.encoded
}

func (d *ValueArray) Encode(IContext) error {

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
		return fmt.Errorf("ValueArray::Encode - unexpected size %d", size)
	}

	// encode content
	content_buf := []byte{}
	for i := 0; i < len(d.data_array); i++ {
		err := d.data_array[i].Encode(nil)
		if err != nil {
			return fmt.Errorf("ValueArray::Encode - err [%v]", err)
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

		return fmt.Errorf("ValueArray::Encode - content length too big %d", content_len)

	}
}

func (d *ValueArray) IsDecoded() bool {
	return true
}

func (d *ValueArray) Decode(IContext) (int, error) {
	return 0, fmt.Errorf("ValueArray::Decode - decode not supported")
}

////////////////////////////////////////
// deep copy

func (d *ValueArray) Copy() IEncodable {

	c := NewValueArray()
	if d.data_array == nil {
		return c
	}

	// make a deep copy of the buf
	c.data_array = make([]IValue, len(d.data_array))
	for i := 0; i < len(d.data_array); i++ {
		c.data_array[i] = d.data_array[i].Copy().(IValue)
	}

	return c
}

func (d *ValueArray) CopyConstruct() (IEncodable, error) {

	c := NewValueArray()
	if d.data_array == nil {
		return c, nil
	}

	// make a deep copy of the buf
	c.data_array = make([]IValue, len(d.data_array))
	for i := 0; i < len(d.data_array); i++ {
		data_i, err := d.data_array[i].CopyConstruct()
		if err != nil {
			return nil, err
		} else {
			c.data_array[i] = data_i.(IValue)
		}
	}

	return c, nil
}

////////////////////////////////////////
// updater

func (d *ValueArray) Append(data IValue) *ValueArray {
	d.data_array = append(d.data_array, data)
	d.encoded = false
	d.estBufSize += data.EstBufSize()
	return d
}

func (d *ValueArray) DeleteAt(idx uint16) *ValueArray {

	if idx >= uint16(len(d.data_array)) {
		panic(fmt.Sprintf("ValueArray::ValueAt - idx [%d] bigger than size [%d]", idx, len(d.data_array)))
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

func (d *RecordList) IsValueArray() bool {
	return false
}

func (d *RecordList) IsRecordList() bool {
	return true
}

func (d *RecordList) Size() uint16 {
	return uint16(len(d.record_list))
}

func (d *RecordList) ValueAt(idx uint16) (IValue, error) {

	return nil, fmt.Errorf("RecordList::ValueAt - not allowed for record list")
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

func (d *RecordList) Value() []byte {
	return nil
}

////////////////////////////////////////
// encoding, decoding, and buf

func (d *RecordList) ValueMagic() byte {

	if !d.encoded {
		panic(fmt.Sprintf("RecordList::ValueMagic - not encoded"))
	}

	return d.buf[0]
}

func (d *RecordList) Buf() []byte {

	if !d.encoded {
		panic(fmt.Sprintf("RecordList::ValueMagic - not encoded"))
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
		panic(fmt.Sprintf("ValueArray::DeleteAt - idx [%d] bigger than size [%d]", idx, len(d.record_list)))
	}

	d.estBufSize -= d.record_list[idx].EstBufSize()
	d.record_list = append(d.record_list[:idx], d.record_list[idx+1:]...)
	d.encoded = false
	return d
}

////////////////////////////////////////////////////////////////////////////////
// utilities
////////////////////////////////////////////////////////////////////////////////

func estimateValueBufSize(d IValue) int {
	if collection.IsNil(d) {
		return 1
	} else {
		return d.EstBufSize()
	}
}
