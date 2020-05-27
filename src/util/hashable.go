package util

import (
	"fmt"
	"reflect"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces
////////////////////////////////////////////////////////////////////////////////

type IObject interface {
}

type IHashable interface {

	////////////////////////////////////////
	// compare if two hashable objects equal to each other
	Equal(IObject) bool

	////////////////////////////////////////
	// takes a hash function, and return uint32 hash value of the object
	//     - implementation may use the hash function directly, or may
	//       use input hash function to compute sub values, and combine as
	//       the object hash value
	HashUint32(f func([]byte) uint32) uint32
}

////////////////////////////////////////////////////////////////////////////////
// HashableSlice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type HashableSlice struct {
	slice []IHashable
}

func NewHashableSlice(s []IHashable) *HashableSlice {
	return &HashableSlice{slice: s}
}

// return XOR of hash of each IHashable object in the slice
func (s *HashableSlice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*HashableSlice)(nil)) {
		return false
	}

	// convert to HashableSlice
	th := t.(*HashableSlice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	for i := range s.slice {
		if (s.slice[i] == nil) != (th.slice[i] == nil) {
			return false
		}
		if s.slice[i] == nil {
			continue
		} else if !s.slice[i].Equal(th.slice[i]) {
			return false
		}
	}

	return true
}

// return XOR of hash of each IHashable object in the slice
func (s *HashableSlice) HashUint32(f func([]byte) uint32) uint32 {

	// compute initial hash with empty byte array
	empty_hash := f([]byte{})

	if s.slice == nil {
		return empty_hash
	}

	h := empty_hash
	for _, obj := range s.slice {
		if obj == nil {
			h ^= empty_hash
		} else {
			if !reflect.TypeOf(obj).Implements(reflect.TypeOf((IHashable)(nil))) {
				panic(fmt.Sprintf("HashableSlice::HashUint32 - unexpected object type [%s]", reflect.TypeOf(obj)))
			}
			h ^= obj.(IHashable).HashUint32(f)
		}
	}

	return h
}

////////////////////////////////////////////////////////////////////////////////
// HashableByteSlice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type HashableByteSlice struct {
	slice []byte
}

func NewHashableByteSlice(s []byte) *HashableByteSlice {
	return &HashableByteSlice{slice: s}
}

// return if two hashable byte array equals
func (s *HashableByteSlice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*HashableByteSlice)(nil)) {
		return false
	}

	// convert to HashableByteSlice
	th := t.(*HashableByteSlice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualByteArray(s.slice, th.slice)
}

// return hash of byte slice
func (s *HashableByteSlice) HashUint32(f func([]byte) uint32) uint32 {

	if s.slice == nil {
		return f([]byte{})
	}

	return f(s.slice)
}

////////////////////////////////////////////////////////////////////////////////
// HashableInt16Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type HashableInt16Slice struct {
	slice []int16
}

func NewHashableInt16Slice(s []int16) *HashableInt16Slice {
	return &HashableInt16Slice{slice: s}
}

// return if two hashable byte array equals
func (s *HashableInt16Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*HashableInt16Slice)(nil)) {
		return false
	}

	// convert to HashableInt16Slice
	th := t.(*HashableInt16Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualInt16Array(s.slice, th.slice)
}

// return hash of byte slice
func (s *HashableInt16Slice) HashUint32(f func([]byte) uint32) uint32 {

	if s.slice == nil {
		return f([]byte{})
	}

	// convert int16 array to byte array
	b := (*(*[]byte)(unsafe.Pointer(&s.slice)))[:len(s.slice)*2]
	return f(b)
}

////////////////////////////////////////////////////////////////////////////////
// HashableUint16Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type HashableUint16Slice struct {
	slice []uint16
}

func NewHashableUint16Slice(s []uint16) *HashableUint16Slice {
	return &HashableUint16Slice{slice: s}
}

// return if two hashable byte array equals
func (s *HashableUint16Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*HashableUint16Slice)(nil)) {
		return false
	}

	// convert to HashableUint16Slice
	th := t.(*HashableUint16Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualUint16Array(s.slice, th.slice)
}

// return hash of byte slice
func (s *HashableUint16Slice) HashUint32(f func([]byte) uint32) uint32 {

	if s.slice == nil {
		return f([]byte{})
	}

	// convert uint16 array to byte array
	b := (*(*[]byte)(unsafe.Pointer(&s.slice)))[:len(s.slice)*2]
	return f(b)
}

////////////////////////////////////////////////////////////////////////////////
// HashableInt32Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type HashableInt32Slice struct {
	slice []int32
}

func NewHashableInt32Slice(s []int32) *HashableInt32Slice {
	return &HashableInt32Slice{slice: s}
}

// return if two hashable byte array equals
func (s *HashableInt32Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*HashableInt32Slice)(nil)) {
		return false
	}

	// convert to HashableInt32Slice
	th := t.(*HashableInt32Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualInt32Array(s.slice, th.slice)
}

// return hash of byte slice
func (s *HashableInt32Slice) HashUint32(f func([]byte) uint32) uint32 {

	if s.slice == nil {
		return f([]byte{})
	}

	// convert int32 array to byte array
	b := (*(*[]byte)(unsafe.Pointer(&s.slice)))[:len(s.slice)*4]
	return f(b)
}

////////////////////////////////////////////////////////////////////////////////
// HashableUint32Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type HashableUint32Slice struct {
	slice []uint32
}

func NewHashableUint32Slice(s []uint32) *HashableUint32Slice {
	return &HashableUint32Slice{slice: s}
}

// return if two hashable byte array equals
func (s *HashableUint32Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*HashableUint32Slice)(nil)) {
		return false
	}

	// convert to HashableUint32Slice
	th := t.(*HashableUint32Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualUint32Array(s.slice, th.slice)
}

// return hash of byte slice
func (s *HashableUint32Slice) HashUint32(f func([]byte) uint32) uint32 {

	if s.slice == nil {
		return f([]byte{})
	}

	// convert uint32 array to byte array
	b := (*(*[]byte)(unsafe.Pointer(&s.slice)))[:len(s.slice)*4]
	return f(b)
}

////////////////////////////////////////////////////////////////////////////////
// HashableInt64Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type HashableInt64Slice struct {
	slice []int64
}

func NewHashableInt64Slice(s []int64) *HashableInt64Slice {
	return &HashableInt64Slice{slice: s}
}

// return if two hashable byte array equals
func (s *HashableInt64Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*HashableInt64Slice)(nil)) {
		return false
	}

	// convert to HashableInt64Slice
	th := t.(*HashableInt64Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualInt64Array(s.slice, th.slice)
}

// return hash of byte slice
func (s *HashableInt64Slice) HashUint32(f func([]byte) uint32) uint32 {

	if s.slice == nil {
		return f([]byte{})
	}

	// convert int64 array to byte array
	b := (*(*[]byte)(unsafe.Pointer(&s.slice)))[:len(s.slice)*8]
	return f(b)
}

////////////////////////////////////////////////////////////////////////////////
// HashableUint64Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type HashableUint64Slice struct {
	slice []uint64
}

func NewHashableUint64Slice(s []uint64) *HashableUint64Slice {
	return &HashableUint64Slice{slice: s}
}

// return if two hashable byte array equals
func (s *HashableUint64Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*HashableUint64Slice)(nil)) {
		return false
	}

	// convert to HashableUint64Slice
	th := t.(*HashableUint64Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualUint64Array(s.slice, th.slice)
}

// return hash of byte slice
func (s *HashableUint64Slice) HashUint32(f func([]byte) uint32) uint32 {

	if s.slice == nil {
		return f([]byte{})
	}

	// convert uint64 array to byte array
	b := (*(*[]byte)(unsafe.Pointer(&s.slice)))[:len(s.slice)*8]
	return f(b)
}
