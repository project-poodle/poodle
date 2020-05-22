package util

import (
    "fmt"
    "time"
    "math/big"
    "encoding/binary"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces

type IRecord interface {
    RecordMagic()       byte                            // 1 byte magic
    Key()               (IData, error)                  // key content
    Value()             (IData, error)                  // value content
    Scheme()            (IData, error)                  // scheme content
    Timestamp()         (*time.Time, error)             // 8 bytes unix timestamp
    Signature()         (*big.Int, *big.Int, error)     // 2 * 32 bytes signature
    Buf()               []byte                          // full Record buffer
    Copy()              IRecord                         // make a deep copy of the record
}

type IData interface {
    // basic methods
    IsNil()             bool                    // whether Data is nil
    IsEncoded()         bool                    // whether Data is encoded
    Content()           []byte                  // data content
    Buf()               []byte                  // full Data buffer
    Copy()              IData                   // make a deep copy of the data
    // encoding methods
    DataMagic()         byte                    // 1 byte Data encode magic
    IsDataArray()       bool                    // whether this is data array
    IsRecordList()      bool                    // whether this is record list
    Size()              uint16                  // size of the data array or record list
    DataAt(i uint16)    (IData, error)          // get i-th Data element - for Array only
    RecordAt(i uint16)  (IRecord, error)        // get i-th Record element - for Composite only
    GetLookup()         []byte                  // get Lookup scheme (2 bytes)
    GetCompression()    []byte                  // get compression scheme (2 bytes)
}


////////////////////////////////////////////////////////////////////////////////
// Mapped Record

type MappedRecord struct {
    buf             []byte
    key             IData           // key
    value           IData           // value
    scheme          IData           // scheme
    timestamp       *time.Time      // timestamp
    signature_r     *big.Int        // signature r
    signature_s     *big.Int        // signature s
}

func NewMappedRecord(buf []byte) (*MappedRecord, error) {
    if (buf == nil || len(buf)<1) {
        return nil, fmt.Errorf("NewMappedRecord - invalid empty buf")
    }
    r, length := &MappedRecord{buf: buf}, 1
    if _, err := r.Key(); err != nil {
        return nil, err
    } else {
        length += len(r.key.Buf())
    }
    if _, err := r.Value(); err != nil {
        return nil, err
    } else {
        length += len(r.value.Buf())
    }
    if _, err := r.Scheme(); err != nil {
        return nil, err
    } else {
        length += len(r.scheme.Buf())
    }
    if _, err := r.Timestamp(); err != nil {
        return nil, err
    } else if r.timestamp != nil {
        length += 8
    }
    if _, _, err := r.Signature(); err != nil {
        return nil, err
    } else if r.signature_r != nil && r.signature_s != nil {
        length += 64
    }
    // set buf length to exact length
    r.buf = buf[:length]
    return r, nil
}

func (r *MappedRecord) RecordMagic() byte {
    return r.buf[0]
}

func (r *MappedRecord) Key() (IData, error) {
    if r.key != nil {
        return r.key, nil
    }

    pos     := 1
    if len(r.buf) < pos {
        return nil, fmt.Errorf("MappedRecord::Key - invalid buf, no key, %d, %x", len(r.buf), r.buf)
    }
    err     := (error)(nil)
    encode  := (r.buf[0] >> 6) & 0x03
    switch encode {
    case 0x00:
        r.key, err  = NewSimpleMappedData(encode, r.buf[pos:])
    case 0x01:
        r.key, err  = NewSimpleMappedData(encode, r.buf[pos:])
    case 0x02:
        r.key, err  = NewSimpleMappedData(encode, r.buf[pos:])
    default:
        r.key, err  = NewEncodedMappedData(r.buf[pos:])
    }

    return r.key, err
}

func (r *MappedRecord) Value() (IData, error) {
    if r.value != nil {
        return r.value, nil
    }

    key, err := r.Key()
    if err != nil {
        return nil, err
    }

    pos     := 1+len(key.Buf())
    if len(r.buf) < pos {
        return nil, fmt.Errorf("MappedRecord::Value - invalid buf, no value, %d, %x", len(r.buf), r.buf)
    }
    encode  := (r.buf[0] >> 4) & 0x03
    switch encode {
    case 0x00:
        r.value, err    = NewSimpleMappedData(encode, r.buf[pos:])
    case 0x01:
        r.value, err    = NewSimpleMappedData(encode, r.buf[pos:])
    case 0x02:
        r.value, err    = NewSimpleMappedData(encode, r.buf[pos:])
    default:
        r.value, err    = NewEncodedMappedData(r.buf[pos:])
    }

    return r.value, err
}

func (r *MappedRecord) Scheme() (IData, error) {
    if r.scheme != nil {
        return r.scheme, nil
    }

    key, err := r.Key()
    if err != nil {
        return nil, err
    }

    value, err := r.Value()
    if err != nil {
        return nil, err
    }

    pos     := 1 + len(key.Buf()) + len(value.Buf())
    if len(r.buf) < pos {
        return nil, fmt.Errorf("MappedRecord::Scheme - invalid buf, no scheme, %d, %x", len(r.buf), r.buf)
    }
    encode  := (r.buf[0] >> 2) & 0x03
    switch encode {
    case 0x00:
        r.scheme, err   = NewSimpleMappedData(encode, r.buf[pos:])
    case 0x01:
        r.scheme, err   = NewSimpleMappedData(encode, r.buf[pos:])
    case 0x02:
        r.scheme, err   = NewSimpleMappedData(encode, r.buf[pos:])
    default:
        r.scheme, err   = NewEncodedMappedData(r.buf[pos:])
    }

    return r.scheme, err
}

func (r *MappedRecord) Timestamp() (*time.Time, error) {
    if r.timestamp != nil {
        return r.timestamp, nil
    }

    encode  := r.buf[0] & 0x01
    if encode == 0x00 {
        return nil, nil
    }

    key, err := r.Key()
    if err != nil {
        return nil, err
    }

    value, err := r.Value()
    if err != nil {
        return nil, err
    }

    scheme, err := r.Scheme()
    if err != nil {
        return nil, err
    }

    // timestamp position
    pos     := 1 + len(key.Buf()) + len(value.Buf()) + len(scheme.Buf())
    if len(r.buf) < pos {
        return nil, fmt.Errorf("MappedRecord::Timestamp - invalid buf, no timestamp, %d, %x", len(r.buf), r.buf)
    }

    r.timestamp, err = BytesToTime(r.buf[pos:])
    if err != nil {
        return nil, err
    } else {
        return r.timestamp, nil
    }
}

func (r *MappedRecord) Signature() (*big.Int, *big.Int, error) {
    if r.signature_r != nil && r.signature_s != nil {
        return r.signature_r, r.signature_s, nil
    }

    encode  := r.buf[0] & 0x01
    if encode == 0x00 {
        return nil, nil, nil
    }

    key, err := r.Key()
    if err != nil {
        return nil, nil, err
    }

    value, err := r.Value()
    if err != nil {
        return nil, nil, err
    }

    scheme, err := r.Scheme()
    if err != nil {
        return nil, nil, err
    }

    pos     := 9 + len(key.Buf()) + len(value.Buf()) + len(scheme.Buf())
    if len(r.buf) < pos {
        return nil, nil, fmt.Errorf("MappedRecord::Signature - invalid buf, no signature, %d, %x", len(r.buf), r.buf)
    }

    if len(r.buf) < pos + 64 { // 2 * 32 bytes signature
        // signature is optional - even if timestamp and signature bit is set
        return nil, nil, nil
    } else {
        r.signature_r = ToBigInt(r.buf[pos:pos+32])
        r.signature_s = ToBigInt(r.buf[pos+32:pos+64])
        return r.signature_r, r.signature_s, nil
    }
}

func (r *MappedRecord) Buf() []byte {
    return r.buf
}

func (r *MappedRecord) Copy() IRecord {
    buf := make([]byte, len(r.buf))
    copy(buf, r.buf)
    copy, err := NewMappedRecord(buf)
    if err != nil {
        // this should not happen
        panic(fmt.Sprintf("MappedRecord:Copy - %s", err))
    }
    return copy
}


////////////////////////////////////////////////////////////////////////////////
// Constructed Record

type ConstructedRecord struct {
    key         IData
    value       IData
    scheme      IData
    timestamp   *time.Time
    signature_r *big.Int
    signature_s *big.Int
}

////////////////////////////////////////////////////////////////////////////////
// Simple Mapped Data

type SimpleMappedData struct {
    encode      byte
    length      uint16
    content     []byte
    buf         []byte
}

func NewSimpleMappedData(encode byte, buf []byte) (*SimpleMappedData, error) {
    switch encode {
    case 0x00:
        return &SimpleMappedData{encode: encode, length: 0, content: nil, buf: nil}, nil
    case 0x01:
        if len(buf) < 1 {
            return nil, fmt.Errorf("NewSimpleMappedData - invalid buf 1, no length, %d, %x", len(buf), buf)
        }
        length := uint16(buf[0])
        if len(buf) < 1 + int(length) {
            return nil, fmt.Errorf("NewSimpleMappedData - invalid buf 1, missing content %d, %x", len(buf), buf)
        }
        return &SimpleMappedData{encode: encode, length: length, content: buf[1:1+length], buf: buf[0:1+length]}, nil
    case 0x02:
        if len(buf) < 2 {
            return nil, fmt.Errorf("NewSimpleMappedData - invalid buf 2, no length %d, %x", len(buf), buf)
        }
        length := uint16(binary.BigEndian.Uint16(buf[0:1]))
        if len(buf) < 2 + int(length) {
            return nil, fmt.Errorf("NewSimpleMappedData - invalid buf 2, missing content %d, %x", len(buf), buf)
        }
        return &SimpleMappedData{encode: encode, length: length, content: buf[2:2+length], buf: buf[0:2+length]}, nil
    default:
        return nil, fmt.Errorf("NewSimpleMappedData - invalid encode [%b]", encode)
    }
}

func (d *SimpleMappedData) IsNil() bool {
    return (d.encode >> 6) == 0
}

func (d *SimpleMappedData) IsEncoded() bool {
    return false
}

func (d *SimpleMappedData) Content() []byte {
    return d.content
}

func (d *SimpleMappedData) Buf() []byte {
    return d.buf
}

func (d *SimpleMappedData) Copy() IData {
    // make a deep copy of the buf
    buf := make([]byte, len(d.buf))
    copy(buf, d.buf)
    copy, err := NewSimpleMappedData(d.encode, buf)
    if err != nil {
        // this should not happen
        panic(fmt.Sprintf("SimpleMappedData:Copy - %s", err))
    }
    return copy
}

func (d *SimpleMappedData) DataMagic() byte {
    return byte(0)
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
    return nil, fmt.Errorf("SimpleMappedData::RecordAt - no composite element")
}

func (d *SimpleMappedData) GetLookup() []byte {
    return nil
}

func (d *SimpleMappedData) GetCompression() []byte {
    return nil
}

////////////////////////////////////////////////////////////////////////////////
// Encoded Mapped Data

type EncodedMappedData struct {
    data_magic      byte
    data_array      []IData
    record_list     []IRecord
    size            uint16
    lookup          []byte
    compression     []byte
    content         []byte
    buf             []byte
}

func NewEncodedMappedData(buf []byte) (*EncodedMappedData, error){
    if len(buf) < 1 {
        return nil, fmt.Errorf("NewEncodedMappedData - invalid empty buf")
    }
    d               := &EncodedMappedData{data_magic: buf[0]}
    buf_length      := uint16(0)
    content_length  := uint16(0)
    encode_is_set   := false
    // process data array
    data_array      := (d.data_magic >> 6) & 0x03
    switch data_array {
    case 0x00:
        break
    case 0x01:
        encode_is_set       = true
        if len(buf) < 1 + int(buf_length) {
            return nil, fmt.Errorf("NewEncodedMappedData - invalid buf 1, data array no size, %d, %x", len(buf), buf)
        }
        d.size              = uint16(buf[1])
        d.data_array        = make([]IData, d.size)
        buf_length          += 1
    case 0x02:
        encode_is_set       = true
        if len(buf) < 2 + int(buf_length) {
            return nil, fmt.Errorf("NewEncodedMappedData - invalid buf 2, data array no size, %d, %x", len(buf), buf)
        }
        d.size              = binary.BigEndian.Uint16(buf[1:2])
        d.data_array        = make([]IData, d.size)
        buf_length          += 2
    case 0x03:
        return nil, fmt.Errorf("NewEncodedMappedData - invalid magic - data array: %b", d.data_magic)
    }
    // process record list
    record_list     := (d.data_magic >> 4) & 0x03
    switch record_list {
    case 0x00:
        break
    case 0x01:
        if encode_is_set {
            return nil, fmt.Errorf("NewEncodedMappedData - invalid magic [%b] - encode set prior to record list", d.data_magic)
        }
        encode_is_set       = true
        if len(buf) < 1 + int(buf_length) {
            return nil, fmt.Errorf("NewEncodedMappedData - invalid buf 1, record list no size, %d, %x", len(buf), buf)
        }
        d.size              = uint16(buf[1])
        d.record_list       = make([]IRecord, d.size)
        buf_length          += 1
    case 0x02:
        if encode_is_set {
            return nil, fmt.Errorf("NewEncodedMappedData - invalid magic [%b] - encode set prior to record list", d.data_magic)
        }
        encode_is_set       = true
        if len(buf) < 2 + int(buf_length) {
            return nil, fmt.Errorf("NewEncodedMappedData - invalid buf 2, record list no size, %d, %x", len(buf), buf)
        }
        d.size              = binary.BigEndian.Uint16(buf[1:2])
        d.record_list       = make([]IRecord, d.size)
        buf_length          += 2
    case 0x03:
        return nil, fmt.Errorf("NewEncodedMappedData - invalid magic [%b] - record list", d.data_magic)
    }
    // process lookup
    lookup_bit      := (d.data_magic >> 3) & 0x01
    switch lookup_bit {
    case 0x00:
        d.lookup            = nil
    case 0x01:
        // buf_length      += 2
        return nil, fmt.Errorf("NewEncodedMappedData - invalid magic [%b] - lookup not supported", d.data_magic)
    }
    // process compression
    compression_bit := (d.data_magic >> 2) & 0x01
    switch compression_bit {
    case 0x00:
        d.compression       = nil
    case 0x01:
        // buf_length      += 2
        return nil, fmt.Errorf("NewEncodedMappedData - invalid magic [%b] - compression not supported", d.data_magic)
    }
    // process length
    length_bit      := d.data_magic & 0x03
    switch length_bit {
    case 0x00:
        break
    case 0x01:
        if len(buf) < 1 + int(buf_length) {
            return nil, fmt.Errorf("NewEncodedMappedData - invalid buf 1, no length, %d, %x", len(buf), buf)
        }
        content_length      =   uint16(buf[buf_length])
        if len(buf) < 1 + int(buf_length) + int(content_length) {
            return nil, fmt.Errorf("NewEncodedMappedData - invalid buf 1, missing content, %d, %x", len(buf), buf)
        }
        buf_length          +=  1 + content_length
    case 0x02:
        if len(buf) < 2 + int(buf_length) {
            return nil, fmt.Errorf("NewEncodedMappedData - invalid buf 2, no length, %d, %x", len(buf), buf)
        }
        content_length      =   binary.BigEndian.Uint16(buf[buf_length:buf_length+1])
        if len(buf) < 2 + int(buf_length) + int(content_length) {
            return nil, fmt.Errorf("NewEncodedMappedData - invalid buf 2, missing content, %d, %x", len(buf), buf)
        }
        buf_length          +=  2 + content_length
    case 0x03:
        return nil, fmt.Errorf("NewEncodedMappedData - invalid magic [%b] - length ", d.data_magic)
    }

    d.content   = buf[buf_length-content_length:buf_length]
    d.buf       = buf[:buf_length]

    return d, nil
}

func (d *EncodedMappedData) IsNil() bool {
    return d.data_magic == 0x00
}

func (d *EncodedMappedData) IsEncoded() bool {
    return true
}

func (d *EncodedMappedData) Content() []byte {
    return d.content
}

func (d *EncodedMappedData) Buf() []byte {
    return d.buf
}

func (d *EncodedMappedData) Copy() IData {
    // make a deep copy of the buf
    buf := make([]byte, len(d.buf))
    copy(buf, d.buf)
    copy, err := NewEncodedMappedData(buf)
    if err != nil {
        // this should not happen
        panic(fmt.Sprintf("EncodedMappedData:Copy - %s", err))
    }
    return copy
}
func (d *EncodedMappedData) DataMagic() byte {
    return d.data_magic
}

func (d *EncodedMappedData) IsDataArray() bool {
    return d.data_array != nil
}

func (d *EncodedMappedData) IsRecordList() bool {
    return d.record_list != nil
}

func (d *EncodedMappedData) Size() uint16 {
    return d.size
}

func (d *EncodedMappedData) DataAt(idx uint16) (IData, error) {

    if idx >= d.size {
        return nil, fmt.Errorf("EncodedMappedData::DataAt - idx [%d] bigger than size [%d]", idx, d.size)
    }

    if (d.data_array[idx] != nil) {
        return d.data_array[idx], nil
    }

    err := (error)(nil)
    pos := 0
    for i := uint16(0); i <= idx; i++ {
        if d.data_array[i] == nil {
            if len(d.content) < pos {
                return nil, fmt.Errorf("EncodedMappedData:DataAt[%d] - invalid content %d - %d, %x", idx, i, len(d.content), d.content)
            }
            d.data_array[i], err = NewEncodedMappedData(d.content[pos:])
            if err != nil {
                return nil, err
            }
        }
        pos += len(d.data_array[i].Buf())
    }

    return d.data_array[idx], nil
}

func (d *EncodedMappedData) RecordAt(idx uint16) (IRecord, error) {

    if idx >= d.size {
        return nil, fmt.Errorf("EncodedMappedData::RecordAt - idx [%d] bigger than size [%d]", idx, d.size)
    }

    if (d.record_list[idx] != nil) {
        return d.record_list[idx], nil
    }

    err := (error)(nil)
    pos := 0
    for i := uint16(0); i <= idx; i++ {
        if d.record_list[i] == nil {
            if len(d.content) < pos {
                return nil, fmt.Errorf("EncodedMappedData:RecordAt[%d] - invalid content %d - %d, %x", idx, i, len(d.content), d.content)
            }
            d.record_list[i], err = NewMappedRecord(d.content[pos:])
            if err != nil {
                return nil, err
            }
        }
        pos += len(d.record_list[i].Buf())
    }

    return d.record_list[idx], nil
}

func (d *EncodedMappedData) GetLookup() []byte {
    return d.lookup
}

func (d *EncodedMappedData) GetCompression() []byte {
    return d.compression
}


////////////////////////////////////////////////////////////////////////////////
// Constructed Data

type ConstructedDataArray struct {
    data_array          []IData
}

type ConstructedRecordList struct {
    record_list         []IData
}

type ConstructedPrimitive struct {
    data                []byte
}
