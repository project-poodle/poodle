package util

import (
	"fmt"
	"math/big"
	"time"
	//"encoding/binary"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces

type IRecord interface {

	////////////////////////////////////////
	// accessor to elements
	Key() IData                      // key content
	Value() IData                    // value content
	Scheme() IData                   // scheme content
	Timestamp() *time.Time           // 8 bytes unix nano timestamp
	Signature() (*big.Int, *big.Int) // optional 2 * 32 bytes signature

	////////////////////////////////////////
	// encoding, decoding, and buf
	RecordMagic() byte // 1 byte Record Magic          - return 0xff if not encoded
	Buf() []byte       // full Record buffer           - return nil if not encoded
	// whether Record is encoded    - always return true for Mapped Record
	//                                  - return true for Constructed Record if encoded buf cache exists
	//                                  - return false for Constructed Record if no encoded buf cache
	IsEncoded() bool
	// encode Record                - for Constructed Record only, return error for Mapped Record
	//                                  - if successful, encoded buf is kept as part of Record object
	Encode() ([]byte, error)
	// whether Record is decoded    - always return true for Constructed Record
	//                                  - return true for Mapped Record if record elements are decoded
	//                                  - return false for Mapped Record if record elements are not decoded
	IsDecoded() bool
	// decode Record                - for Mapped Record only, return error for Constructed Record
	//                                  - if successful, individual key, value, scheme, timestamp, and signature pointers are decoded as part of Record object
	Decode() error

	////////////////////////////////////////
	// deep copy
	Copy() IRecord                   // make a deep copy of the record with same composition
	CopyConstruct() (IRecord, error) // copy from source record to constructed record recursively
}

////////////////////////////////////////////////////////////////////////////////
// Mapped Record
////////////////////////////////////////////////////////////////////////////////

type MappedRecord struct {
	// buf
	decoded bool   // whether this is decoded
	buf     []byte // original buf if not decoded, exact buf size if already decoded
	// elements
	key         IData      // key
	value       IData      // value
	scheme      IData      // scheme
	timestamp   *time.Time // timestamp
	signature_r *big.Int   // signature r
	signature_s *big.Int   // signature s
}

////////////////////////////////////////
// constructor

func NewMappedRecord(buf []byte) (*MappedRecord, error) {

	if buf == nil || len(buf) < 1 {
		return nil, fmt.Errorf("NewMappedRecord - empty buf")
	}

	// initialize record
	r := &MappedRecord{decoded: false, buf: buf}

	// decode
	err := r.Decode()
	if err != nil {
		return nil, err
	}

	return r, nil
}

////////////////////////////////////////
// accessor to elements

func (r *MappedRecord) Key() IData {

	if !r.decoded {
		// this should not happend
		panic(fmt.Sprintf("MappedRecord::Key - not decoded"))
	}

	return r.key
}

func (r *MappedRecord) Value() IData {

	if !r.decoded {
		// this should not happend
		panic(fmt.Sprintf("MappedRecord::Value - not decoded"))
	}

	return r.value
}

func (r *MappedRecord) Scheme() IData {

	if !r.decoded {
		// this should not happend
		panic(fmt.Sprintf("MappedRecord::Scheme - not decoded"))
	}

	return r.scheme
}

func (r *MappedRecord) Timestamp() *time.Time {

	if !r.decoded {
		// this should not happend
		panic(fmt.Sprintf("MappedRecord::Timestamp - not decoded"))
	}

	return r.timestamp
}

func (r *MappedRecord) Signature() (*big.Int, *big.Int) {

	if !r.decoded {
		// this should not happend
		panic(fmt.Sprintf("MappedRecord::Signature - not decoded"))
	}

	return r.signature_r, r.signature_s
}

////////////////////////////////////////
// encoding, decoding, and buf

func (r *MappedRecord) RecordMagic() byte {

	if !r.decoded {
		return 0xff
	} else {
		return r.buf[0]
	}
}

func (r *MappedRecord) Buf() []byte {
	return r.buf
}

func (r *MappedRecord) IsEncoded() bool {
	return true
}

func (r *MappedRecord) Encode() ([]byte, error) {
	return nil, fmt.Errorf("MappedRecord::Encode - cannot encode MappedRecord")
}

func (r *MappedRecord) IsDecoded() bool {
	return r.decoded
}

func (r *MappedRecord) Decode() (err error) {

	pos := 1

	var (
		key    IData
		value  IData
		scheme IData
	)
	// key
	encode := (r.buf[0] >> 6) & 0x03
	switch encode {
	case 0x00:
		key, err = NewSimpleMappedData(encode, r.buf[pos:])
	case 0x01:
		key, err = NewSimpleMappedData(encode, r.buf[pos:])
	case 0x02:
		key, err = NewSimpleMappedData(encode, r.buf[pos:])
	default:
		key, err = NewStandardMappedData(r.buf[pos:])
	}
	if err != nil {
		return err
	} else if len(key.Data()) > MAX_KEY_LENGTH {
		return fmt.Errorf("MappedRecord::Decode - key size %d larger than %d", len(key.Data()), MAX_KEY_LENGTH)
	} else {
		r.key = key
		pos += len(r.key.Buf())
	}

	// value
	if len(r.buf) < pos {
		return fmt.Errorf("MappedRecord::Decode - invalid buf, no value, %d, %x", len(r.buf), r.buf)
	}
	encode = (r.buf[0] >> 4) & 0x03
	switch encode {
	case 0x00:
		value, err = NewSimpleMappedData(encode, r.buf[pos:])
	case 0x01:
		value, err = NewSimpleMappedData(encode, r.buf[pos:])
	case 0x02:
		value, err = NewSimpleMappedData(encode, r.buf[pos:])
	default:
		value, err = NewStandardMappedData(r.buf[pos:])
	}
	if err != nil {
		return err
	} else if len(value.Data()) > MAX_VALUE_LENGTH {
		return fmt.Errorf("MappedRecord::Decode - value size %d larger than %d", len(value.Data()), MAX_VALUE_LENGTH)
	} else {
		r.value = value
		pos += len(r.value.Buf())
	}

	// scheme
	if len(r.buf) < pos {
		return fmt.Errorf("MappedRecord::Decode - invalid buf, no scheme, %d, %x", len(r.buf), r.buf)
	}
	encode = (r.buf[0] >> 2) & 0x03
	switch encode {
	case 0x00:
		scheme, err = NewSimpleMappedData(encode, r.buf[pos:])
	case 0x01:
		scheme, err = NewSimpleMappedData(encode, r.buf[pos:])
	case 0x02:
		scheme, err = NewSimpleMappedData(encode, r.buf[pos:])
	default:
		scheme, err = NewStandardMappedData(r.buf[pos:])
	}
	if err != nil {
		return err
	} else if len(scheme.Data()) > MAX_SCHEME_LENGTH {
		return fmt.Errorf("MappedRecord::Decode - scheme size %d larger than %d", len(scheme.Data()), MAX_SCHEME_LENGTH)
	} else {
		r.scheme = scheme
		pos += len(r.scheme.Buf())
	}

	// timestamp and signature bit
	encode = r.buf[0] & 0x01
	if encode == 0x01 {

		// timestamp
		if len(r.buf) < pos {
			return fmt.Errorf("MappedRecord::Decode - invalid buf, no timestamp, %d, %x", len(r.buf), r.buf)
		}
		r.timestamp, err = BytesToTime(r.buf[pos:])
		if err != nil {
			return err
		} else {
			pos += 8
		}

		// signature (optional)
		if len(r.buf) < pos {
			return fmt.Errorf("MappedRecord::Decode - invalid buf, no signature, %d, %x", len(r.buf), r.buf)
		} else if len(r.buf) < pos+64 { // 2 * 32 bytes signature
			// signature is optional - even if timestamp and signature bit is set
			// return err
			r.signature_r = nil
			r.signature_s = nil
		} else {
			r.signature_r = ByteArrayToBigInt(r.buf[pos : pos+32])
			r.signature_s = ByteArrayToBigInt(r.buf[pos+32 : pos+64])
			pos += 64
		}
	}

	// set buf length to exact length and return record
	r.buf = r.buf[:pos]

	r.decoded = true

	return nil
}

////////////////////////////////////////
// deep copy

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

func (r *MappedRecord) CopyConstruct() (IRecord, error) {

	result := NewRecord()
	var err error

	result.key, err = r.Key().CopyConstruct()
	if err != nil {
		return nil, err
	}

	result.value, err = r.Value().CopyConstruct()
	if err != nil {
		return nil, err
	}

	result.scheme, err = r.Scheme().CopyConstruct()
	if err != nil {
		return nil, err
	}

	result.timestamp = r.Timestamp()                       // timestamp is immutable
	result.signature_r, result.signature_s = r.Signature() // signature is immutable

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
// Constructed Record
////////////////////////////////////////////////////////////////////////////////

type Record struct {
	// buf
	encoded bool
	buf     []byte
	// elements
	key         IData
	value       IData
	scheme      IData
	timestamp   *time.Time
	signature_r *big.Int
	signature_s *big.Int
}

////////////////////////////////////////
// constructor

func NewRecord() *Record {
	return &Record{encoded: false}
}

////////////////////////////////////////
// accessor to elements

func (r *Record) Key() IData {
	return r.key
}

func (r *Record) Value() IData {
	return r.value
}

func (r *Record) Scheme() IData {
	return r.scheme
}

func (r *Record) Timestamp() *time.Time {
	return r.timestamp
}

func (r *Record) Signature() (*big.Int, *big.Int) {
	return r.signature_r, r.signature_s
}

////////////////////////////////////////
// encoding, decoding, and buf

func (r *Record) RecordMagic() byte {

	if !r.encoded {
		panic(fmt.Sprintf("Record::DataMagic - not encoded"))
	}

	return r.buf[0]
}

func (r *Record) Buf() []byte {

	if !r.encoded {
		panic(fmt.Sprintf("Record::Buf - not encoded"))
	}

	return r.buf
}

func (r *Record) IsEncoded() bool {
	return r.encoded
}

func (r *Record) Encode() ([]byte, error) {

	buf := []byte{0x00}

	// encode key
	if r.key != nil && !r.key.IsNil() {

		content_buf, magic, err := r.key.Encode(true)
		if err != nil {
			return nil, err
		}

		buf[0] |= (magic & 0x03) << 6
		buf = append(buf, content_buf...)
	}

	// encode value
	if r.value != nil && !r.value.IsNil() {

		content_buf, magic, err := r.value.Encode(true)
		if err != nil {
			return nil, err
		}

		buf[0] |= (magic & 0x03) << 4
		buf = append(buf, content_buf...)
	}

	// encode scheme
	if r.scheme != nil && !r.scheme.IsNil() {

		content_buf, magic, err := r.scheme.Encode(true)
		if err != nil {
			return nil, err
		}

		buf[0] |= (magic & 0x03) << 2
		buf = append(buf, content_buf...)
	}

	// encode timestamp
	if r.timestamp != nil {

		buf[0] |= 0x01
		buf = append(buf, TimeToBytes(r.timestamp)...)

		// encode signature
		if r.signature_r != nil && r.signature_s != nil {
			buf = append(buf, BigIntToByteArray(r.signature_r)...)
			buf = append(buf, BigIntToByteArray(r.signature_s)...)
		}
	}

	// record encoded buf
	r.buf = buf
	r.encoded = true

	return r.buf, nil
}

func (r *Record) IsDecoded() bool {
	return true
}

func (r *Record) Decode() error {
	return fmt.Errorf("Record::Decode - decode not supported")
}

////////////////////////////////////////
// deep copy

func (r *Record) Copy() IRecord {

	result := &Record{}
	if r.key != nil {
		result.key = r.key.Copy()
	}

	if r.value != nil {
		result.value = r.value.Copy()
	}

	if r.scheme != nil {
		result.scheme = r.scheme.Copy()
	}

	result.timestamp = r.timestamp
	result.signature_r = r.signature_r
	result.signature_s = r.signature_s

	return result
}

func (r *Record) CopyConstruct() (IRecord, error) {

	var err error

	result := &Record{}
	if r.key != nil {
		result.key, err = r.key.CopyConstruct()
		if err != nil {
			return nil, err
		}
	}

	if r.value != nil {
		result.value, err = r.value.CopyConstruct()
		if err != nil {
			return nil, err
		}
	}

	if r.scheme != nil {
		result.scheme, err = r.scheme.CopyConstruct()
		if err != nil {
			return nil, err
		}
	}

	result.timestamp = r.timestamp
	result.signature_r = r.signature_r
	result.signature_s = r.signature_s

	return result, nil
}

////////////////////////////////////////
// updater

func (r *Record) SetKey(key IData) *Record {
	r.key = key
	r.encoded = false
	return r
}

func (r *Record) SetK(key []byte) *Record {
	r.key = NewPrimitive(key)
	r.encoded = false
	return r
}

func (r *Record) SetValue(value IData) *Record {
	r.value = value
	r.encoded = false
	return r
}

func (r *Record) SetV(value []byte) *Record {
	r.value = NewPrimitive(value)
	r.encoded = false
	return r
}

func (r *Record) SetScheme(scheme IData) *Record {
	r.scheme = scheme
	r.encoded = false
	return r
}

func (r *Record) SetS(scheme []byte) *Record {
	r.scheme = NewPrimitive(scheme)
	r.encoded = false
	return r
}

func (r *Record) SetTimestamp(t *time.Time) *Record {
	r.timestamp = t
	r.encoded = false
	return r
}

func (r *Record) SetSignature(R, S *big.Int) *Record {
	r.signature_r = R
	r.signature_s = S
	r.encoded = false
	return r
}
