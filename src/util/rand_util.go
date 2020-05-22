package util

import (
    "fmt"
    "bytes"
    //"unsafe"
    "crypto/rand"
    "encoding/binary"
)


func RandInt32() (int32) {
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

func RandUint32() (uint32) {
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

func RandInt64() (int64) {
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

func RandUint64() (uint64) {
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


func RandInt32Range(min, max int32) (int32) {
    data := RandInt32()
    result := data % (max - min) + min
    return result
}

func RandUint32Range(min, max uint32) (uint32) {
    data := RandUint32()
    result := data % (max - min) + min
    return result
}

func RandInt64Range(min, max int64) (int64) {
    data := RandInt64()
    result := data % (max - min) + min
    return result
}

func RandUint64Range(min, max uint64) (uint64) {
    data := RandUint64()
    result := data % (max - min) + min
    return result
}

