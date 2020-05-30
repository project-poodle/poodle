package collection

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// utilities
////////////////////////////////////////////////////////////////////////////////

func µ(a ...interface{}) []interface{} {
	return a
}

func Ternary(statement bool, a, b interface{}) interface{} {
	if statement {
		return a
	}
	return b
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////
// Min & Max
////////////////////////////////////////////////////////////////////////////////

func MinInt(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinUnt(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxUint(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinInt8(a, b int8) int8 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxInt8(a, b int8) int8 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinUnt8(a, b uint8) uint8 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinInt16(a, b int16) int16 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxInt16(a, b int16) int16 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinUnt16(a, b uint16) uint16 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxUint16(a, b uint16) uint16 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinInt32(a, b int32) int32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxInt32(a, b int32) int32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinUint32(a, b uint32) uint32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxUint32(a, b uint32) uint32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinUint64(a, b uint64) uint64 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	} else {
		return b
	}
}

////////////////////////////////////////////////////////////////////////////////
// Equal of primitive array
////////////////////////////////////////////////////////////////////////////////

func EqualByteArray(a, b []byte) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func EqualInt8Array(a, b []int8) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func EqualInt16Array(a, b []int16) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func EqualUint16Array(a, b []uint16) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func EqualInt32Array(a, b []int32) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func EqualUint32Array(a, b []uint32) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func EqualInt64Array(a, b []int64) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func EqualUint64Array(a, b []uint64) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func EqualIntArray(a, b []int) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

////////////////////////////////////////////////////////////////////////////////
// Type Conversions
////////////////////////////////////////////////////////////////////////////////

func Int64ToTime(nano int64) *time.Time {
	t := time.Unix(0, nano)
	return &t
}

func BytesToTime(buf []byte) (*time.Time, error) {
	if len(buf) < 8 {
		return nil, fmt.Errorf("BytesToTime - buf length less than 8 bytes [%x]", buf)
	}
	nano := binary.BigEndian.Uint64(buf[:8])
	return Int64ToTime(int64(nano)), nil
}

func TimeToBytes(t *time.Time) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(t.UnixNano()))
	return buf
}

func ByteArrayToBigInt(data []byte) *big.Int {
	result := new(big.Int)
	result.SetBytes(data)
	return result
}

func BigIntToByteArray(d *big.Int) []byte {
	input := d.Bytes()
	if len(input) == 32 {
		return input
	} else if len(input) > 32 {
		return input[len(input)-32:]
	} else {
		buf := make([]byte, 32-len(input))
		return append(buf, input[:]...)
	}
}

func Int64ToByteArray(input int64) []byte {
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, uint64(input))
	return result
}

func ByteArrayToInt64(buf []byte) int64 {
	var data uint64
	err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		return 0
	}
	return int64(data)
}

func Int32ToByteArray(input int32) []byte {
	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, uint32(input))
	return result
}

func ByteArrayToInt32(buf []byte) int32 {
	var data uint32
	err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		return 0
	}
	return int32(data)
}

func Uint32ToByteArray(input uint32) []byte {
	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, uint32(input))
	return result
}

func ByteArrayToUint32(buf []byte) uint32 {
	var data uint32
	err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		return 0
	}
	return data
}
