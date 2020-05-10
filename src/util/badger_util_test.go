package util

import (
    "fmt"
    "testing"
    "crypto/rand"
    //"github.com/boltdb/bolt"
)

var setup_badger_data [][]byte = nil

func setupBadgerData() ([][]byte) {

    if setup_badger_data != nil {
        return setup_badger_data
    }

    data := make([][]byte, 8192)
    for i:=0; i<len(data); i++ {
        data[i] = make([]byte, 256)
        for j:=0; j<len(data[i]); {
            if n, err := rand.Read(data[i]); err != nil {
                fmt.Printf("%s\n", err)
            } else {
                j += n
            }
        }
    }

    setup_badger_data = data
    return data
}


func BenchmarkBadgerPut_iNode(b *testing.B) {
    data := setupBadgerData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        BadgerPut_iNode(data[i%len(data)], data[i%len(data)])
    }
}

func BenchmarkBadgerGet_iNode(b *testing.B) {
    data := setupBadgerData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        BadgerGet_iNode(data[i%len(data)])
    }
}


func BenchmarkBadgerPut_Container(b *testing.B) {
    data := setupBadgerData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        BadgerPut_Container(data[i%len(data)], data[i%len(data)])
    }
}

func BenchmarkBadgerGet_Container(b *testing.B) {
    data := setupBadgerData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        BadgerGet_Container(data[i%len(data)])
    }
}

