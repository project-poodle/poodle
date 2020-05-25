package util

import (
	"encoding/binary"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces
////////////////////////////////////////////////////////////////////////////////

type IKey interface {
	// this returns a uniquely identified key
	IsNil() bool // whether Key is nil
	Key() [][]byte
	SubKeyAt(idx int) []byte
	// encode, decode, and buf
	Buf() []byte     // return buf
	IsEncoded() bool // check if encoded
	Encode() error   // encode
	IsDecoded() bool // check if decoded
	Decode() error   // decode
	// copy
	Copy() IKey                   // copy
	CopyConstruct() (IKey, error) // copy construct
	// hash, equals and print
	Equal(IKey) bool // compare if two IKey equal to each other
	// takes a hash function, and return XOR hash of all sub keys, return 0 for empty key
	HashUint32(f func([]byte) uint32) uint32
	// print in readable format
	ToString() string
}

////////////////////////////////////////////////////////////////////////////////
// EmptyKey
////////////////////////////////////////////////////////////////////////////////

type EmptyKey struct {
}

func NewEmptyKey() *EmptyKey {
	return &EmptyKey{}
}

func (k *EmptyKey) IsNil() bool {
	return true
}

// returns empty array
func (k *EmptyKey) Key() [][]byte {
	return []([]byte){}
}

func (k *EmptyKey) SubKeyAt(idx int) []byte {
	panic(fmt.Sprintf("EmptyKey::SubKeyAt - empty key have no sub key index"))
}

func (k *EmptyKey) Buf() []byte {
	//result := []byte{}
	//return result
	return []byte{0x00}
}

func (k *EmptyKey) IsEncoded() bool {
	return true
}

func (k *EmptyKey) Encode() error {
	return nil
}

func (k *EmptyKey) IsDecoded() bool {
	return true
}

func (k *EmptyKey) Decode() error {
	return nil
}

func (k *EmptyKey) Copy() IKey {
	return NewEmptyKey()
}

func (k *EmptyKey) CopyConstruct() (IKey, error) {
	return NewEmptyKey(), nil
}

func (k *EmptyKey) Equal(o IKey) bool {
	return len(o.Key()) == 0
}

func (k *EmptyKey) HashUint32(f func([]byte) uint32) uint32 {
	hashValue := uint32(0)
	return hashValue
}

func (k *EmptyKey) ToString() string {
	return fmt.Sprintf("EmptyKey")
}

////////////////////////////////////////////////////////////////////////////////
// MappedKey
////////////////////////////////////////////////////////////////////////////////

type MappedKey struct {
	// buf
	decoded bool
	keys    [][]byte
	buf     []byte
}

////////////////////////////////////////
// constructor

func NewMappedKey(buf []byte) (*MappedKey, error) {

	result := &MappedKey{keys: []([]byte){}, buf: buf} // initialize with empty key and empty buf

	err := result.Decode()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (k *MappedKey) IsNil() bool {
	if !k.decoded {
		panic(fmt.Sprintf("MappedKey::Key - not decoded"))
	}

	return k.keys == nil || len(k.keys) == 0
}

func (k *MappedKey) Key() [][]byte {
	if !k.decoded {
		panic(fmt.Sprintf("MappedKey::Key - not decoded"))
	}

	return k.keys
}

func (k *MappedKey) SubKeyAt(idx int) []byte {
	if !k.decoded {
		panic(fmt.Sprintf("MappedKey::SubKeyAt - not decoded"))
	}

	if idx >= len(k.keys) {
		panic(fmt.Sprintf("MappedKey::SubKeyAt - index %d bigger than length %d", idx, len(k.keys)))
	}

	return k.keys[idx]
}

func (k *MappedKey) Buf() []byte {
	if k.IsNil() {
		return []byte{0x00}
	}
	return k.buf
}

func (k *MappedKey) IsEncoded() bool {
	return true
}

func (k *MappedKey) Encode() error {
	return nil
}

func (k *MappedKey) IsDecoded() bool {
	return k.keys == nil
}

func (k *MappedKey) Decode() error {

	k.keys = []([]byte){}
	pos := 0

	totalKeyLength, totalKeyN := binary.Uvarint(k.buf)
	if totalKeyN < 0 || (len(k.buf) != 0 && totalKeyN == 0 && k.buf[0] != 0) {
		//if n <= 0 {
		return fmt.Errorf("NewMappedKey - failed to read length")
	} else if totalKeyN == 0 && totalKeyLength == 0 {
		k.decoded = true
		return nil // return empty buf
	} else if totalKeyLength > MAX_KEY_LENGTH {
		return fmt.Errorf("NewMappedKey - length %d bigger than %d", totalKeyLength, MAX_KEY_LENGTH)
	} else if len(k.buf) < int(totalKeyLength)+totalKeyN {
		return fmt.Errorf("NewMappedKey - length %d bigger than buf length %d", len(k.buf), totalKeyLength)
	}
	pos += totalKeyN

	for pos < int(totalKeyLength)+totalKeyN {

		subKeylength, subKeyN := binary.Uvarint(k.buf[pos:])
		if subKeyN < 0 {
			// sub key cannot have zero length
			return fmt.Errorf("NewMappedKey - failed to read sub key length at pos %d", pos)
		} else if len(k.buf)-pos < int(subKeylength)+subKeyN {
			return fmt.Errorf("NewMappedKey - sub key length %d bigger than buf length %d at pos %d", subKeylength, len(k.buf), pos)
		}
		k.keys = append(k.keys, k.buf[pos+subKeyN:pos+subKeyN+int(subKeylength)])
		pos += subKeyN + int(subKeylength)
		// check if we have parsed all of key buffer
		if pos > int(totalKeyLength)+totalKeyN {
			return fmt.Errorf("NewMappedKey - position %d out of bound %d", pos, int(totalKeyLength)+totalKeyN)
		}
	}

	k.buf = k.buf[:pos] // set buf to exact key length
	k.decoded = true

	return nil
}

func (k *MappedKey) Copy() IKey {
	result, err := NewMappedKey(k.buf)
	if err != nil {
		panic(fmt.Errorf("MappedKey::Copy - unexpected failure %s", err))
	}

	return result
}

func (k *MappedKey) CopyConstruct() (IKey, error) {

	result := NewKey()

	for _, key := range k.Key() {
		result.Add(key)
	}

	return result, nil
}

func (k *MappedKey) Equal(o IKey) bool {
	if len(k.Key()) != len(o.Key()) {
		return false
	}

	for i, key := range k.keys {
		if !EqByteArray(key, o.SubKeyAt(i)) {
			return false
		}
	}

	return true
}

func (k *MappedKey) HashUint32(f func([]byte) uint32) uint32 {
	hashValue := uint32(0)
	for _, key := range k.keys {
		hashValue ^= f(key)
	}
	return hashValue
}

func (k *MappedKey) ToString() string {
	str := fmt.Sprintf("MappedKey")
	if k.decoded {
		for i, subKey := range k.keys {
			str += fmt.Sprintf("\n    subKey[%d] = %v", i, subKey)
		}
	}
	str += fmt.Sprintf("\n    buf = %v", k.buf[:MinInt(len(k.buf), 32)])
	return str
}

////////////////////////////////////////////////////////////////////////////////
// Key
////////////////////////////////////////////////////////////////////////////////

type Key struct {
	encoded bool
	keys    [][]byte
	buf     []byte
}

func NewKey() *Key {
	return &Key{keys: []([]byte){}, buf: nil}
}

func NewSimpleKey(simpleKey []byte) *Key {
	return &Key{keys: []([]byte){simpleKey}, buf: nil}
}

func NewStringKey(stringKey string) *Key {
	return &Key{keys: []([]byte){[]byte(stringKey)}, buf: nil}
}

func (k *Key) IsNil() bool {
	return k.keys == nil || len(k.keys) == 0
}

// returns empty array
func (k *Key) Key() [][]byte {
	return k.keys
}

func (k *Key) SubKeyAt(idx int) []byte {
	if idx >= len(k.keys) {
		panic(fmt.Sprintf("MappedKey::SubKeyAt - index %d bigger than length %d", idx, len(k.keys)))
	}

	return k.keys[idx]
}

func (k *Key) Buf() []byte {
	if !k.encoded {
		panic(fmt.Sprintf("Key::Buf - not encoded"))
	}

	return k.buf
}

func (k *Key) IsEncoded() bool {
	return k.encoded
}

func (k *Key) Encode() error {
	// TODO
	bufs := make([][]byte, len(k.keys))

	// calculate total key length
	totalLength := 0
	lenBuf := make([]byte, 10) // maximum 10 bytes
	for i, subKey := range k.keys {
		lenN := binary.PutUvarint(lenBuf, uint64(len(subKey)))
		if lenN <= 0 {
			panic(fmt.Sprintf("[%d] invalid uvarint length %d", lenN, len(subKey)))
		}
		bufs[i] = make([]byte, lenN+len(subKey))
		copy(bufs[i], lenBuf[:lenN])
		copy(bufs[i][lenN:], subKey)
		totalLength += len(bufs[i])
	}

	// encode total key length
	totalN := binary.PutUvarint(lenBuf, uint64(totalLength))
	buf := make([]byte, totalN+totalLength)
	copy(buf, lenBuf[:totalN])

	// encode each sub key
	pos := totalN
	for i := range bufs {
		copy(buf[pos:], bufs[i])
		pos += len(bufs[i])
	}

	if pos > MAX_KEY_LENGTH {
		return fmt.Errorf("Key::Encode - key length %d bigger than %d", pos, MAX_KEY_LENGTH)
	} else if pos != len(buf) {
		return fmt.Errorf("Key::Encode - buf length %d does not match pos %d", len(buf), pos)
	}

	// record encoded buf
	k.buf = buf[:pos]
	k.encoded = true

	return nil
}

func (k *Key) IsDecoded() bool {
	return true
}

func (k *Key) Decode() error {
	return nil
}

func (k *Key) Copy() IKey {

	result := NewKey()

	for _, key := range k.Key() {
		result.Add(key)
	}

	return result
}

func (k *Key) CopyConstruct() (IKey, error) {

	result := NewKey()

	for _, key := range k.Key() {
		result.Add(key)
	}

	return result, nil
}

func (k *Key) HashUint32(f func([]byte) uint32) uint32 {
	hashValue := uint32(0)
	for _, key := range k.keys {
		hashValue ^= f(key)
	}
	return hashValue
}

func (k *Key) Equal(o IKey) bool {
	if len(k.Key()) != len(o.Key()) {
		return false
	}

	for i, key := range k.keys {
		if !EqByteArray(key, o.SubKeyAt(i)) {
			return false
		}
	}

	return true
}

func (k *Key) ToString() string {
	str := fmt.Sprintf("Key")
	for i, subKey := range k.keys {
		str += fmt.Sprintf("\n    subKey[%d] = %v", i, subKey)
	}
	str += fmt.Sprintf("\n    buf = %v", k.buf[:MinInt(len(k.buf), 32)])
	return str
}

func (k *Key) Add(subKey []byte) *Key {

	if subKey == nil || len(subKey) == 0 {
		panic(fmt.Sprintf("Key::Add - subKey cannot have 0 length"))
	}

	k.keys = append(k.keys, subKey)
	return k
}
