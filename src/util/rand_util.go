package util

import (
    "bytes"
    //"unsafe"
    "crypto/rand"
    "encoding/binary"
)


func RandInt32() (int32, error) {
    buf := make([]byte, 4)
    _, err := rand.Read(buf)
    if err != nil {
        return 0, err
    }

    var data int32
    err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
    if err != nil {
        return 0, err
    }

    return data, err
}

func RandUint32() (uint32, error) {
    buf := make([]byte, 4)
    _, err := rand.Read(buf)
    if err != nil {
        return 0, err
    }

    var data uint32
    err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
    if err != nil {
        return 0, err
    }

    return data, err
}

func RandInt64() (int64, error) {
    buf := make([]byte, 8)
    _, err := rand.Read(buf)
    if err != nil {
        return 0, err
    }

    var data int64
    err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
    if err != nil {
        return 0, err
    }

    return data, err
}

func RandUint64() (uint64, error) {
    buf := make([]byte, 8)
    _, err := rand.Read(buf)
    if err != nil {
        return 0, err
    }

    var data uint64
    err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &data)
    if err != nil {
        return 0, err
    }

    return data, err
}


func RandInt32Range(min, max int32) (int32, error) {
    data, err := RandInt32()
    if err != nil {
        return min, err
    }

    result := data % (max - min) + min
    return result, nil
}

func RandUint32Range(min, max uint32) (uint32, error) {
    data, err := RandUint32()
    if err != nil {
        return min, err
    }

    result := data % (max - min) + min
    return result, nil
}

func RandInt64Range(min, max int64) (int64, error) {
    data, err := RandInt64()
    if err != nil {
        return min, err
    }

    result := data % (max - min) + min
    return result, nil
}

func RandUint64Range(min, max uint64) (uint64, error) {
    data, err := RandUint64()
    if err != nil {
        return min, err
    }

    result := data % (max - min) + min
    return result, nil
}

