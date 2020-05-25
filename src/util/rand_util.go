package util

import (
	"bytes"
	"fmt"

	//"unsafe"
	"crypto/rand"
	"encoding/binary"
)

func RandInt8() int8 {
	buf := make([]byte, 1)
	_, err := rand.Read(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandInt8 - %s", err))
	}

	var data int8
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandInt8 - %s", err))
	}

	return data
}

func RandUint8() uint8 {
	buf := make([]byte, 1)
	_, err := rand.Read(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandUint32 - %s", err))
	}

	var data uint8
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandUint32 - %s", err))
	}

	return data
}

func RandInt16() int16 {
	buf := make([]byte, 2)
	_, err := rand.Read(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandInt16 - %s", err))
	}

	var data int16
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandInt16 - %s", err))
	}

	return data
}

func RandUint16() uint16 {
	buf := make([]byte, 2)
	_, err := rand.Read(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandUint32 - %s", err))
	}

	var data uint16
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandUint32 - %s", err))
	}

	return data
}

func RandInt32() int32 {
	buf := make([]byte, 4)
	_, err := rand.Read(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandInt32 - %s", err))
	}

	var data int32
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandInt32 - %s", err))
	}

	return data
}

func RandUint32() uint32 {
	buf := make([]byte, 4)
	_, err := rand.Read(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandUint32 - %s", err))
	}

	var data uint32
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandUint32 - %s", err))
	}

	return data
}

func RandInt64() int64 {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandInt64 - %s", err))
	}

	var data int64
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandInt64 - %s", err))
	}

	return data
}

func RandUint64() uint64 {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandUint64 - %s", err))
	}

	var data uint64
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("RandUint64 - %s", err))
	}

	return data
}

func RandInt8Range(min, max int8) int8 {
	data := RandInt8()
	result := data%(max-min) + min
	return result
}

func RandUint8Range(min, max uint8) uint8 {
	data := RandUint8()
	result := data%(max-min) + min
	return result
}

func RandInt16Range(min, max int16) int16 {
	data := RandInt16()
	result := data%(max-min) + min
	return result
}

func RandUint16Range(min, max uint16) uint16 {
	data := RandUint16()
	result := data%(max-min) + min
	return result
}

func RandInt32Range(min, max int32) int32 {
	data := RandInt32()
	result := data%(max-min) + min
	return result
}

func RandUint32Range(min, max uint32) uint32 {
	data := RandUint32()
	result := data%(max-min) + min
	return result
}

func RandInt64Range(min, max int64) int64 {
	data := RandInt64()
	result := data%(max-min) + min
	return result
}

func RandUint64Range(min, max uint64) uint64 {
	data := RandUint64()
	result := data%(max-min) + min
	return result
}
