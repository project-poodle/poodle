package util

import (
    //"log"
    "fmt"
    //"time"
    //"unsafe"
    "testing"
    //"crypto/aes"
    //"crypto/rand"
    //"github.com/boltdb/bolt"
    //"github.com/golang/protobuf/proto"
    //"crypto/ecdsa"
    //"../proto_cluster"
)

func TestRandInt32(t *testing.T) {
    d, err := RandInt32()

    if err != nil {
        t.Fatal("RandInt32 : ", err)
    }

    fmt.Printf("RandInt32 : %x\n", d)
}
