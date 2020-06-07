package util

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"

	"../collection"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces
////////////////////////////////////////////////////////////////////////////////

type IKey interface {

	////////////////////////////////////////
	// embeded interfaces
	collection.IComparable
	collection.IHashable
	collection.IPrintable
	IEncodable

	////////////////////////////////////////
	// accessor to elements
	IsEmpty() bool // whether Key is nil
	Key() [][]byte
	SubKeyAt(idx int) []byte
}

////////////////////////////////////////////////////////////////////////////////
// EmptyKey
////////////////////////////////////////////////////////////////////////////////

type EmptyKey struct {
}

func NewEmptyKey() *EmptyKey {
	return &EmptyKey{}
}

func (k *EmptyKey) IsEmpty() bool {
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

func (k *EmptyKey) EstBufSize() int {
	return 1
}

func (k *EmptyKey) IsEncoded() bool {
	return true
}

func (k *EmptyKey) Encode(IContext) error {
	return nil
}

func (k *EmptyKey) IsDecoded() bool {
	return true
}

func (k *EmptyKey) Decode(IContext) (int, error) {
	return 0, nil
}

func (k *EmptyKey) Copy() IEncodable {
	return NewEmptyKey()
}

func (k *EmptyKey) CopyConstruct() (IEncodable, error) {
	return NewEmptyKey(), nil
}

func (k *EmptyKey) Equal(o collection.IObject) bool {

	if o == nil {
		return false
	}

	if !reflect.TypeOf(o).Implements(reflect.TypeOf((*IKey)(nil)).Elem()) {
		return false
	}

	return len(o.(IKey).Key()) == 0
}

func (k *EmptyKey) Compare(c collection.IComparable) int {

	if collection.IsNil(c) {
		return 1
	}

	if !reflect.TypeOf(c).Implements(reflect.TypeOf((*IKey)(nil)).Elem()) {
		panic(fmt.Sprintf("EmptyKey::Compare - target is not IKey [%v]", reflect.TypeOf(c)))
	}

	t := c.(IKey)
	if len(t.Key()) == 0 {
		return 0
	} else {
		return -1
	}
}

func (k *EmptyKey) HashUint32(f func([]byte) uint32) uint32 {
	hashValue := uint32(0)
	return hashValue
}

func (k *EmptyKey) Print(w io.Writer, indent int) {
	fmt.Fprintf(w, "EmptyKey")
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

func NewMappedKey(buf []byte) (*MappedKey, int, error) {

	result := &MappedKey{keys: []([]byte){}, buf: Ternary(buf == nil, []byte{}, buf).([]byte)} // initialize with empty key and empty buf

	keyN, err := result.Decode(nil)
	if err != nil {
		return nil, keyN, err
	}

	return result, keyN, nil
}

func (k *MappedKey) IsEmpty() bool {
	if k.buf == nil || len(k.buf) == 0 {
		return true
	}

	if !k.decoded {
		panic(fmt.Sprintf("MappedKey::IsEmpty - not decoded"))
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
	if k.IsEmpty() {
		return []byte{0x00}
	}
	return k.buf
}

func (k *MappedKey) EstBufSize() int {
	if k.IsEmpty() {
		return 1
	}
	return len(k.buf)
}

func (k *MappedKey) IsEncoded() bool {
	return true
}

func (k *MappedKey) Encode(IContext) error {
	return nil
}

func (k *MappedKey) IsDecoded() bool {
	return k.keys == nil
}

func (k *MappedKey) Decode(IContext) (int, error) {

	k.keys = []([]byte){}

	totalKey, totalKeyN, err := DecodeVarchar(k.buf)
	if err != nil {
		return 0, fmt.Errorf("MappedKey::Decode - %s", err)
	}

	if totalKey == nil {
		k.decoded = true
		return 0, nil // return empty buf successfully
	}

	if len(totalKey) > MAX_KEY_LENGTH {
		return 0, fmt.Errorf("MappedKey::Decode - length %d bigger than %d", len(totalKey), MAX_KEY_LENGTH)
	}

	pos := 0
	for pos < len(totalKey) {

		subKey, subKeyN, err := DecodeVarchar(totalKey[pos:])
		if err != nil {
			return 0, fmt.Errorf("MappedKey::Decode - position [%d] - %s", pos, err)
		} else if pos+subKeyN > len(totalKey) {
			return 0, fmt.Errorf("MappedKey::Decode - sub key [%d / %d] at [%d] too long [%d]", subKeyN, len(subKey), pos, len(totalKey))
		} else if len(subKey) == 0 {
			// sub key cannot have zero length
			return 0, fmt.Errorf("MappedKey::Decode - zero sub key length at [%d]", pos)
		}
		k.keys = append(k.keys, subKey)
		pos += subKeyN
	}

	// check if we have parsed all of key buffer
	if pos > len(totalKey) {
		return 0, fmt.Errorf("MappedKey::Decode - position [%d] out of bound %d", pos, len(totalKey)+totalKeyN)
	}

	// we are here when len(totalKey) == pos
	if pos != len(totalKey) {
		panic(fmt.Sprintf("MappedKey::Decode - position [%d] does not match key length %d", pos, len(totalKey)))
	}

	k.buf = k.buf[:totalKeyN] // set buf to exact key length
	k.decoded = true

	return totalKeyN, nil
}

func (k *MappedKey) Copy() IEncodable {
	result, _, err := NewMappedKey(k.buf)
	if err != nil {
		panic(fmt.Errorf("MappedKey::Copy - unexpected failure %s", err))
	}

	return result
}

func (k *MappedKey) CopyConstruct() (IEncodable, error) {

	result := NewKey()

	for _, key := range k.Key() {
		result.Add(key)
	}

	return result, nil
}

func (k *MappedKey) Equal(o collection.IObject) bool {
	if o == nil {
		return false
	}

	if !reflect.TypeOf(o).Implements(reflect.TypeOf((*IKey)(nil)).Elem()) {
		return false
	}

	obj := o.(IKey)
	if len(k.Key()) != len(obj.Key()) {
		return false
	}

	for i, key := range k.keys {
		if !collection.EqualByteSlice(key, obj.SubKeyAt(i)) {
			return false
		}
	}

	return true
}

func (k *MappedKey) Compare(c collection.IComparable) int {

	if collection.IsNil(c) {
		return 1
	}

	if !reflect.TypeOf(c).Implements(reflect.TypeOf((*IKey)(nil)).Elem()) {
		panic(fmt.Sprintf("MappedKey::Compare - target is not IKey [%v]", reflect.TypeOf(c)))
	}

	t := c.(IKey)
	for idx, subKey := range k.Key() {
		if len(t.Key()) < idx {
			return 1
		}
		r := collection.CompareByteSlice(subKey, t.SubKeyAt(idx))
		if r != 0 {
			return r
		} else {
			continue
		}
	}

	// we are here if all subKey matches
	if len(t.Key()) > len(k.Key()) {
		return -1
	} else {
		return 0
	}
}

func (k *MappedKey) HashUint32(f func([]byte) uint32) uint32 {
	hashValue := uint32(0)
	for _, key := range k.keys {
		hashValue ^= f(key)
	}
	return hashValue
}

func (k *MappedKey) Print(w io.Writer, indent int) {
	fmt.Fprintf(w, "MappedKey:\n")
	if k.decoded {
		for i, subKey := range k.keys {
			fmt.Fprintf(w, "    subKey[%d] = %v\n", i, subKey)
		}
	}
	fmt.Fprintf(w, "    buf = %v\n", k.buf[:collection.MinInt(len(k.buf), 32)])
}

func (k *MappedKey) ToString() string {
	str := fmt.Sprintf("MappedKey")
	if k.decoded {
		for i, subKey := range k.keys {
			str += fmt.Sprintf("\n    subKey[%d] = %v", i, subKey)
		}
	}
	str += fmt.Sprintf("\n    buf = %v", k.buf[:collection.MinInt(len(k.buf), 32)])
	return str
}

////////////////////////////////////////////////////////////////////////////////
// Key
////////////////////////////////////////////////////////////////////////////////

type Key struct {
	encoded    bool
	keys       [][]byte
	buf        []byte
	estBufSize int
}

func NewKey() *Key {
	return &Key{keys: []([]byte){}, buf: nil, estBufSize: 1}
}

func NewSimpleKey(simpleKey []byte) *Key {
	return &Key{keys: []([]byte){simpleKey}, buf: nil, estBufSize: 2 + len(simpleKey)}
}

func NewStringKey(stringKey string) *Key {
	return &Key{keys: []([]byte){[]byte(stringKey)}, buf: nil, estBufSize: 2 + len(stringKey)}
}

func (k *Key) IsEmpty() bool {
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

func (k *Key) EstBufSize() int {
	return k.estBufSize
}

func (k *Key) IsEncoded() bool {
	return k.encoded
}

func (k *Key) Encode(IContext) error {
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
	k.estBufSize = pos
	k.encoded = true

	return nil
}

func (k *Key) IsDecoded() bool {
	return true
}

func (k *Key) Decode(IContext) (int, error) {
	return 0, fmt.Errorf("Key::Decode - not supported")
}

func (k *Key) Copy() IEncodable {

	result := NewKey()

	for _, key := range k.Key() {
		result.Add(key)
	}

	return result
}

func (k *Key) CopyConstruct() (IEncodable, error) {

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

func (k *Key) Equal(o collection.IObject) bool {
	if o == nil {
		return false
	}

	if !reflect.TypeOf(o).Implements(reflect.TypeOf((*IKey)(nil)).Elem()) {
		return false
	}

	obj := o.(IKey)
	if len(k.Key()) != len(obj.Key()) {
		return false
	}

	for i, key := range k.keys {
		if !collection.EqualByteSlice(key, obj.SubKeyAt(i)) {
			return false
		}
	}

	return true
}

func (k *Key) Compare(c collection.IComparable) int {

	if collection.IsNil(c) {
		return 1
	}

	if !reflect.TypeOf(c).Implements(reflect.TypeOf((*IKey)(nil)).Elem()) {
		panic(fmt.Sprintf("Key::Compare - target is not IKey [%v]", reflect.TypeOf(c)))
	}

	t := c.(IKey)
	for idx, subKey := range k.Key() {
		if len(t.Key()) < idx {
			return 1
		}
		r := collection.CompareByteSlice(subKey, t.SubKeyAt(idx))
		if r != 0 {
			return r
		} else {
			continue
		}
	}

	// we are here if all subKey matches
	if len(t.Key()) > len(k.Key()) {
		return -1
	} else {
		return 0
	}
}

func (k *Key) Print(w io.Writer, indent int) {
	fmt.Fprintf(w, "Key:\n")
	for i, subKey := range k.keys {
		fmt.Fprintf(w, "    subKey[%d] = %v\n", i, subKey)
	}
	fmt.Fprintf(w, "    buf = %v\n", k.buf[:collection.MinInt(len(k.buf), 32)])

}

func (k *Key) ToString() string {

	str := fmt.Sprintf("Key")

	for i, subKey := range k.keys {
		str += fmt.Sprintf("\n    subKey[%d] = %v", i, subKey)
	}
	str += fmt.Sprintf("\n    buf = %v", k.buf[:collection.MinInt(len(k.buf), 32)])

	return str
}

func (k *Key) Add(subKey []byte) *Key {

	if subKey == nil || len(subKey) == 0 {
		panic(fmt.Sprintf("Key::Add - subKey cannot have 0 length"))
	}

	k.keys = append(k.keys, subKey)
	k.estBufSize += 1 + len(subKey)
	return k
}
