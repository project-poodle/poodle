package util

import (
	"fmt"
	"math/big"
	"time"
	//"encoding/binary"
	"../collection"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces

type IRecord interface {

	////////////////////////////////////////
	// embeded interfaces
	//collection.IPrintable
	IEncodable

	////////////////////////////////////////
	// accessor to elements
	Key() IKey                       // key content
	Value() IValue                   // value content
	Scheme() IValue                  // scheme content
	Timestamp() *time.Time           // 8 bytes unix nano timestamp
	Signature() (*big.Int, *big.Int) // optional 2 * 32 bytes signature

	////////////////////////////////////////
	// encoding, decoding, and buf
	RecordMagic() byte // 1 byte Record Magic          - return 0xff if not encoded
}

////////////////////////////////////////////////////////////////////////////////
// Mapped Record
////////////////////////////////////////////////////////////////////////////////

type MappedRecord struct {
	// buf
	decoded bool   // whether this is decoded
	buf     []byte // original buf if not decoded, exact buf size if already decoded
	// elements
	key         IKey       // key
	value       IValue     // value
	scheme      IValue     // scheme
	timestamp   *time.Time // timestamp
	signature_r *big.Int   // signature r
	signature_s *big.Int   // signature s
}

////////////////////////////////////////
// constructor

func NewMappedRecord(buf []byte) (*MappedRecord, int, error) {

	if buf == nil || len(buf) < 1 {
		return nil, 0, fmt.Errorf("NewMappedRecord - empty buf")
	}

	// initialize record
	r := &MappedRecord{decoded: false, buf: buf}

	// decode
	length, err := r.Decode(nil)
	if err != nil {
		return nil, length, err
	}

	return r, length, nil
}

////////////////////////////////////////
// accessor to elements

func (r *MappedRecord) Key() IKey {

	if !r.decoded {
		// this should not happend
		panic(fmt.Sprintf("MappedRecord::Key - not decoded"))
	}

	return r.key
}

func (r *MappedRecord) Value() IValue {

	if !r.decoded {
		// this should not happend
		panic(fmt.Sprintf("MappedRecord::Value - not decoded"))
	}

	return r.value
}

func (r *MappedRecord) Scheme() IValue {

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

func (r *MappedRecord) EstBufSize() int {
	return len(r.buf)
}

func (r *MappedRecord) IsEncoded() bool {
	return true
}

func (r *MappedRecord) Encode(IContext) error {
	return fmt.Errorf("MappedRecord::Encode - cannot encode MappedRecord")
}

func (r *MappedRecord) IsDecoded() bool {
	return r.decoded
}

func (r *MappedRecord) Decode(IContext) (length int, err error) {

	pos := 1

	var (
		key    IKey
		value  IValue
		scheme IValue
	)
	// key
	hasKey := (r.buf[0] >> 6) & 0x01
	if hasKey != 0 {
		key, length, err = NewMappedKey(r.buf[pos:])
	}
	if err != nil {
		return 0, fmt.Errorf("MappedRecord::Decode - key error [%v]", err)
	} else if len(key.Buf()) > MAX_KEY_LENGTH {
		return 0, fmt.Errorf("MappedRecord::Decode - key size %d larger than %d", len(key.Buf()), MAX_KEY_LENGTH)
	} else {
		r.key = key
		pos += len(r.key.Buf())
	}

	// value
	if len(r.buf) < pos {
		return 0, fmt.Errorf("MappedRecord::Decode - invalid buf, no value, %d, %x", len(r.buf), r.buf)
	}
	hasValue := (r.buf[0] >> 5) & 0x01
	if hasValue != 0 {
		value, length, err = NewStandardMappedValue(r.buf[pos:])
	}
	if err != nil {
		return 0, fmt.Errorf("MappedRecord::Decode - value error [%v]", err)
	} else if len(value.Buf()) > MAX_VALUE_LENGTH {
		return 0, fmt.Errorf("MappedRecord::Decode - value size %d larger than %d", len(value.Buf()), MAX_VALUE_LENGTH)
	} else {
		r.value = value
		pos += len(r.value.Buf())
	}

	// scheme
	if len(r.buf) < pos {
		return 0, fmt.Errorf("MappedRecord::Decode - invalid buf, no scheme, %d, %x", len(r.buf), r.buf)
	}
	hasScheme := (r.buf[0] >> 4) & 0x01
	if hasScheme != 0 {
		scheme, length, err = NewStandardMappedValue(r.buf[pos:])
	}
	if err != nil {
		return 0, fmt.Errorf("MappedRecord::Decode - scheme error [%v]", err)
	} else if len(scheme.Buf()) > MAX_SCHEME_LENGTH {
		return 0, fmt.Errorf("MappedRecord::Decode - scheme size %d larger than %d", len(scheme.Buf()), MAX_SCHEME_LENGTH)
	} else {
		r.scheme = scheme
		pos += len(r.scheme.Buf())
	}

	// timestamp bit
	hasTimestamp := (r.buf[0] >> 2) & 0x01
	if hasTimestamp != 0 {
		if len(r.buf) < pos {
			return 0, fmt.Errorf("MappedRecord::Decode - invalid buf, no timestamp, %d, %x", len(r.buf), r.buf)
		}
		r.timestamp, err = collection.BytesToTime(r.buf[pos:])
		if err != nil {
			return 0, fmt.Errorf("MappedRecord::Decode - timestamp error [%v]", err)
		} else {
			pos += 8
		}
	}

	// signature (optional)
	hasSignature := (r.buf[0] >> 1) & 0x01
	if hasSignature != 0 {
		if len(r.buf) < pos {
			return 0, fmt.Errorf("MappedRecord::Decode - invalid buf, no signature, %d, %x", len(r.buf), r.buf)
		} else if len(r.buf) < pos+64 { // 2 * 32 bytes signature
			// signature is optional - even if signature bit is set
			r.signature_r = nil
			r.signature_s = nil
		} else {
			r.signature_r = collection.ByteArrayToBigInt(r.buf[pos : pos+32])
			r.signature_s = collection.ByteArrayToBigInt(r.buf[pos+32 : pos+64])
			pos += 64
		}
	}

	// set buf length to exact length and return record
	r.buf = r.buf[:pos]

	r.decoded = true

	return pos, nil
}

////////////////////////////////////////
// deep copy

func (r *MappedRecord) Copy() IEncodable {
	buf := make([]byte, len(r.buf))
	copy(buf, r.buf)
	copy, _, err := NewMappedRecord(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("MappedRecord:Copy - %s", err))
	}
	return copy
}

func (r *MappedRecord) CopyConstruct() (IEncodable, error) {

	result := NewRecord()
	var err error

	key, err := r.Key().CopyConstruct()
	result.key = key.(IKey)
	if err != nil {
		return nil, fmt.Errorf("MappedRecord::CopyConstruct - key error [%v]", err)
	}

	value, err := r.Value().CopyConstruct()
	result.value = value.(IValue)
	if err != nil {
		return nil, fmt.Errorf("MappedRecord::CopyConstruct - value error [%v]", err)
	}

	scheme, err := r.Scheme().CopyConstruct()
	result.scheme = scheme.(IValue)
	if err != nil {
		return nil, fmt.Errorf("MappedRecord::CopyConstruct - scheme error [%v]", err)
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
	encoded       bool
	buf           []byte
	estKeySize    int
	estDataSize   int
	estSchemeSize int
	// elements
	key         IKey
	value       IValue
	scheme      IValue
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

func (r *Record) Key() IKey {
	return r.key
}

func (r *Record) Value() IValue {
	return r.value
}

func (r *Record) Scheme() IValue {
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

func (r *Record) EstBufSize() int {

	// initialize
	result := 1

	// key size
	if r.estKeySize <= 0 {
		result += 1
	} else {
		result += r.estKeySize
	}

	// data size
	if r.estDataSize <= 0 {
		result += 1
	} else {
		result += r.estDataSize
	}

	// schema size
	if r.estSchemeSize <= 0 {
		result += 1
	} else {
		result += r.estSchemeSize
	}

	// timestamp
	if r.timestamp != nil {
		result += 8
	}

	// signature
	if r.signature_r != nil && r.signature_s != nil {
		result += 32 * 2
	}

	return result
}

func (r *Record) IsEncoded() bool {
	return r.encoded
}

func (r *Record) Encode(IContext) error {

	buf := []byte{0x00}

	// encode key
	if r.key != nil && !r.key.IsEmpty() {

		err := r.key.Encode(nil)
		if err != nil {
			return fmt.Errorf("Record::Encode - key error [%v]", err)
		}

		buf[0] |= byte(0x01) << 6
		buf = append(buf, r.key.Buf()...)
	}

	// encode value
	if r.value != nil && !r.value.IsNil() {

		err := r.value.Encode(nil)
		if err != nil {
			return fmt.Errorf("Record::Encode - value error [%v]", err)
		}

		buf[0] |= byte(0x01) << 5
		buf = append(buf, r.value.Buf()...)
	}

	// encode scheme
	if r.scheme != nil && !r.scheme.IsNil() {

		err := r.scheme.Encode(nil)
		if err != nil {
			return fmt.Errorf("Record::Encode - scheme error [%v]", err)
		}

		buf[0] |= byte(0x01) << 4
		buf = append(buf, r.scheme.Buf()...)
	}

	// encode timestamp
	if r.timestamp != nil {

		buf[0] |= byte(0x01) << 2
		buf = append(buf, collection.TimeToBytes(r.timestamp)...)

		// encode signature
		if r.signature_r != nil && r.signature_s != nil {
			buf[0] |= byte(0x01) << 1
			buf = append(buf, collection.BigIntToByteArray(r.signature_r)...)
			buf = append(buf, collection.BigIntToByteArray(r.signature_s)...)
		}
	}

	// record encoded buf
	r.buf = buf
	r.encoded = true

	return nil
}

func (r *Record) IsDecoded() bool {
	return true
}

func (r *Record) Decode(IContext) (int, error) {
	return 0, fmt.Errorf("Record::Decode - decode not supported")
}

////////////////////////////////////////
// deep copy

func (r *Record) Copy() IEncodable {

	result := &Record{}
	if r.key != nil {
		key := r.key.Copy()
		result.key = key.(IKey)
	}

	if r.value != nil {
		value := r.value.Copy()
		result.value = value.(IValue)
	}

	if r.scheme != nil {
		scheme := r.scheme.Copy()
		result.scheme = scheme.(IValue)
	}

	result.timestamp = r.timestamp
	result.signature_r = r.signature_r
	result.signature_s = r.signature_s

	return result
}

func (r *Record) CopyConstruct() (IEncodable, error) {

	result := &Record{}
	if r.key != nil {
		key, err := r.key.CopyConstruct()
		if err != nil {
			return nil, err
		} else {
			result.key = key.(IKey)
		}
	}

	if r.value != nil {
		value, err := r.value.CopyConstruct()
		if err != nil {
			return nil, err
		} else {
			result.value = value.(IValue)
		}
	}

	if r.scheme != nil {
		scheme, err := r.scheme.CopyConstruct()
		if err != nil {
			return nil, err
		} else {
			result.scheme = scheme.(IValue)
		}
	}

	result.timestamp = r.timestamp
	result.signature_r = r.signature_r
	result.signature_s = r.signature_s

	return result, nil
}

////////////////////////////////////////
// updater

func (r *Record) SetKey(key IKey) *Record {
	r.key = key
	r.encoded = false
	r.estKeySize = r.key.EstBufSize()
	return r
}

func (r *Record) SetK(key []byte) *Record {
	r.key = NewKey().Add(key)
	r.encoded = false
	r.estKeySize = r.key.EstBufSize()
	return r
}

func (r *Record) SetValue(value IValue) *Record {
	r.value = value
	r.encoded = false
	r.estDataSize = r.value.EstBufSize()
	return r
}

func (r *Record) SetV(value []byte) *Record {
	r.value = NewPrimitive(value)
	r.encoded = false
	r.estDataSize = r.value.EstBufSize()
	return r
}

func (r *Record) SetScheme(scheme IValue) *Record {
	r.scheme = scheme
	r.encoded = false
	r.estSchemeSize = r.scheme.EstBufSize()
	return r
}

func (r *Record) SetS(scheme []byte) *Record {
	r.scheme = NewPrimitive(scheme)
	r.encoded = false
	r.estSchemeSize = r.scheme.EstBufSize()
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
