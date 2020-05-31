package collection

import (
	"fmt"
	"reflect"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces
////////////////////////////////////////////////////////////////////////////////

type IComparable interface {

	////////////////////////////////////////
	// compare if two comparable objects equal to each other
	Equal(IObject) bool

	////////////////////////////////////////
	// compare two comparable objects equal to each other
	Compare(IComparable) int
}

type ISortedMapIterator interface {

	////////////////////////////////////////
	// return the next object
	Next() (IComparable, IObject)

	////////////////////////////////////////
	// whether next object exist
	HasNext() bool
}

////////////////////////////////////////////////////////////////////////////////
// ComparableSlice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type ComparableSlice struct {
	slice []IComparable
}

func NewComparableSlice(s []IComparable) *ComparableSlice {
	return &ComparableSlice{slice: s}
}

// return XOR of hash of each IComparable object in the slice
func (s *ComparableSlice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*ComparableSlice)(nil)) {
		return false
	}

	// convert to ComparableSlice
	th := t.(*ComparableSlice)

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

// compare each element in the slice
func (s *ComparableSlice) Compare(c IComparable) int {

	if IsNil(s.slice) && IsNil(c) {
		return 0
	} else if IsNil(s.slice) {
		return -1
	} else if IsNil(c) {
		return 1
	}

	// we are here if both s.slice and c are not nil
	if reflect.TypeOf(s) != reflect.TypeOf(c) {
		panic(fmt.Sprintf("ComparableSlice::Compare - target is not ComparableSlice [%v]", reflect.TypeOf(c)))
	}

	t := c.(*ComparableSlice)

	for i, _ := range s.slice {

		if len(t.slice) < i {
			return 1
		}

		compare := s.slice[i].Compare(t.slice[i])
		if compare != 0 {
			return compare
		} else {
			continue
		}
	}

	// we are here if all the previous elements in the slice are equal
	if len(t.slice) > len(s.slice) {
		return -1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// ComparableByteSlice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type ComparableByteSlice struct {
	slice []byte
}

func NewComparableByteSlice(s []byte) *ComparableByteSlice {
	return &ComparableByteSlice{slice: s}
}

// return if two hashable byte array equals
func (s *ComparableByteSlice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*ComparableByteSlice)(nil)) {
		return false
	}

	// convert to ComparableByteSlice
	th := t.(*ComparableByteSlice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualByteArray(s.slice, th.slice)
}

// compare each element in the slice
func (s *ComparableByteSlice) Compare(c IComparable) int {

	if IsNil(s.slice) && IsNil(c) {
		return 0
	} else if IsNil(s.slice) {
		return -1
	} else if IsNil(c) {
		return 1
	}

	// we are here if both s.slice and c are not nil
	if reflect.TypeOf(s) != reflect.TypeOf(c) {
		panic(fmt.Sprintf("ComparableSlice::Compare - target is not ComparableByteSlice [%v]", reflect.TypeOf(c)))
	}

	t := c.(*ComparableByteSlice)

	for i, _ := range s.slice {

		if len(t.slice) < i {
			return 1
		}

		if s.slice[i] < t.slice[i] {
			return -1
		} else if s.slice[i] > t.slice[i] {
			return 1
		} else {
			continue
		}
	}

	// we are here if all the previous elements in the slice are equal
	if len(t.slice) > len(s.slice) {
		return -1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// ComparableInt16Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type ComparableInt16Slice struct {
	slice []int16
}

func NewComparableInt16Slice(s []int16) *ComparableInt16Slice {
	return &ComparableInt16Slice{slice: s}
}

// return if two hashable byte array equals
func (s *ComparableInt16Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*ComparableInt16Slice)(nil)) {
		return false
	}

	// convert to ComparableInt16Slice
	th := t.(*ComparableInt16Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualInt16Array(s.slice, th.slice)
}

// compare each element in the slice
func (s *ComparableInt16Slice) Compare(c IComparable) int {

	if IsNil(s.slice) && IsNil(c) {
		return 0
	} else if IsNil(s.slice) {
		return -1
	} else if IsNil(c) {
		return 1
	}

	// we are here if both s.slice and c are not nil
	if reflect.TypeOf(s) != reflect.TypeOf(c) {
		panic(fmt.Sprintf("ComparableSlice::Compare - target is not ComparableInt16Slice [%v]", reflect.TypeOf(c)))
	}

	t := c.(*ComparableInt16Slice)

	for i, _ := range s.slice {

		if len(t.slice) < i {
			return 1
		}

		if s.slice[i] < t.slice[i] {
			return -1
		} else if s.slice[i] > t.slice[i] {
			return 1
		} else {
			continue
		}
	}

	// we are here if all the previous elements in the slice are equal
	if len(t.slice) > len(s.slice) {
		return -1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// ComparableUint16Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type ComparableUint16Slice struct {
	slice []uint16
}

func NewComparableUint16Slice(s []uint16) *ComparableUint16Slice {
	return &ComparableUint16Slice{slice: s}
}

// return if two hashable byte array equals
func (s *ComparableUint16Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*ComparableUint16Slice)(nil)) {
		return false
	}

	// convert to ComparableUint16Slice
	th := t.(*ComparableUint16Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualUint16Array(s.slice, th.slice)
}

// compare each element in the slice
func (s *ComparableUint16Slice) Compare(c IComparable) int {

	if IsNil(s.slice) && IsNil(c) {
		return 0
	} else if IsNil(s.slice) {
		return -1
	} else if IsNil(c) {
		return 1
	}

	// we are here if both s.slice and c are not nil
	if reflect.TypeOf(s) != reflect.TypeOf(c) {
		panic(fmt.Sprintf("ComparableSlice::Compare - target is not ComparableUint16Slice [%v]", reflect.TypeOf(c)))
	}

	t := c.(*ComparableUint16Slice)

	for i, _ := range s.slice {

		if len(t.slice) < i {
			return 1
		}

		if s.slice[i] < t.slice[i] {
			return -1
		} else if s.slice[i] > t.slice[i] {
			return 1
		} else {
			continue
		}
	}

	// we are here if all the previous elements in the slice are equal
	if len(t.slice) > len(s.slice) {
		return -1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// ComparableInt32Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type ComparableInt32Slice struct {
	slice []int32
}

func NewComparableInt32Slice(s []int32) *ComparableInt32Slice {
	return &ComparableInt32Slice{slice: s}
}

// return if two hashable byte array equals
func (s *ComparableInt32Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*ComparableInt32Slice)(nil)) {
		return false
	}

	// convert to ComparableInt32Slice
	th := t.(*ComparableInt32Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualInt32Array(s.slice, th.slice)
}

// compare each element in the slice
func (s *ComparableInt32Slice) Compare(c IComparable) int {

	if IsNil(s.slice) && IsNil(c) {
		return 0
	} else if IsNil(s.slice) {
		return -1
	} else if IsNil(c) {
		return 1
	}

	// we are here if both s.slice and c are not nil
	if reflect.TypeOf(s) != reflect.TypeOf(c) {
		panic(fmt.Sprintf("ComparableSlice::Compare - target is not ComparableInt32Slice [%v]", reflect.TypeOf(c)))
	}

	t := c.(*ComparableInt32Slice)

	for i, _ := range s.slice {

		if len(t.slice) < i {
			return 1
		}

		if s.slice[i] < t.slice[i] {
			return -1
		} else if s.slice[i] > t.slice[i] {
			return 1
		} else {
			continue
		}
	}

	// we are here if all the previous elements in the slice are equal
	if len(t.slice) > len(s.slice) {
		return -1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// ComparableUint32Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type ComparableUint32Slice struct {
	slice []uint32
}

func NewComparableUint32Slice(s []uint32) *ComparableUint32Slice {
	return &ComparableUint32Slice{slice: s}
}

// return if two hashable byte array equals
func (s *ComparableUint32Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*ComparableUint32Slice)(nil)) {
		return false
	}

	// convert to ComparableUint32Slice
	th := t.(*ComparableUint32Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualUint32Array(s.slice, th.slice)
}

// compare each element in the slice
func (s *ComparableUint32Slice) Compare(c IComparable) int {

	if IsNil(s.slice) && IsNil(c) {
		return 0
	} else if IsNil(s.slice) {
		return -1
	} else if IsNil(c) {
		return 1
	}

	// we are here if both s.slice and c are not nil
	if reflect.TypeOf(s) != reflect.TypeOf(c) {
		panic(fmt.Sprintf("ComparableSlice::Compare - target is not ComparableUint32Slice [%v]", reflect.TypeOf(c)))
	}

	t := c.(*ComparableUint32Slice)

	for i, _ := range s.slice {

		if len(t.slice) < i {
			return 1
		}

		if s.slice[i] < t.slice[i] {
			return -1
		} else if s.slice[i] > t.slice[i] {
			return 1
		} else {
			continue
		}
	}

	// we are here if all the previous elements in the slice are equal
	if len(t.slice) > len(s.slice) {
		return -1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// ComparableInt64Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type ComparableInt64Slice struct {
	slice []int64
}

func NewComparableInt64Slice(s []int64) *ComparableInt64Slice {
	return &ComparableInt64Slice{slice: s}
}

// return if two hashable byte array equals
func (s *ComparableInt64Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*ComparableInt64Slice)(nil)) {
		return false
	}

	// convert to ComparableInt64Slice
	th := t.(*ComparableInt64Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualInt64Array(s.slice, th.slice)
}

// compare each element in the slice
func (s *ComparableInt64Slice) Compare(c IComparable) int {

	if IsNil(s.slice) && IsNil(c) {
		return 0
	} else if IsNil(s.slice) {
		return -1
	} else if IsNil(c) {
		return 1
	}

	// we are here if both s.slice and c are not nil
	if reflect.TypeOf(s) != reflect.TypeOf(c) {
		panic(fmt.Sprintf("ComparableSlice::Compare - target is not ComparableInt64Slice [%v]", reflect.TypeOf(c)))
	}

	t := c.(*ComparableInt64Slice)

	for i, _ := range s.slice {

		if len(t.slice) < i {
			return 1
		}

		if s.slice[i] < t.slice[i] {
			return -1
		} else if s.slice[i] > t.slice[i] {
			return 1
		} else {
			continue
		}
	}

	// we are here if all the previous elements in the slice are equal
	if len(t.slice) > len(s.slice) {
		return -1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// ComparableUint64Slice
////////////////////////////////////////////////////////////////////////////////

// Make slice hashable
type ComparableUint64Slice struct {
	slice []uint64
}

func NewComparableUint64Slice(s []uint64) *ComparableUint64Slice {
	return &ComparableUint64Slice{slice: s}
}

// return if two hashable byte array equals
func (s *ComparableUint64Slice) Equal(t IObject) bool {
	if (s == nil) != (t == nil) {
		return false
	}

	if s == nil {
		return true
	}

	if reflect.TypeOf(t) != reflect.TypeOf((*ComparableUint64Slice)(nil)) {
		return false
	}

	// convert to ComparableUint64Slice
	th := t.(*ComparableUint64Slice)

	if len(s.slice) != len(th.slice) {
		return false
	}

	return EqualUint64Array(s.slice, th.slice)
}

// compare each element in the slice
func (s *ComparableUint64Slice) Compare(c IComparable) int {

	if IsNil(s.slice) && IsNil(c) {
		return 0
	} else if IsNil(s.slice) {
		return -1
	} else if IsNil(c) {
		return 1
	}

	// we are here if both s.slice and c are not nil
	if reflect.TypeOf(s) != reflect.TypeOf(c) {
		panic(fmt.Sprintf("ComparableSlice::Compare - target is not ComparableUint64Slice [%v]", reflect.TypeOf(c)))
	}

	t := c.(*ComparableUint64Slice)

	for i, _ := range s.slice {

		if len(t.slice) < i {
			return 1
		}

		if s.slice[i] < t.slice[i] {
			return -1
		} else if s.slice[i] > t.slice[i] {
			return 1
		} else {
			continue
		}
	}

	// we are here if all the previous elements in the slice are equal
	if len(t.slice) > len(s.slice) {
		return -1
	}

	return 0
}
