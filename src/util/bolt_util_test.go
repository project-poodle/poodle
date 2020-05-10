package util

import (
    "fmt"
    "testing"
    "crypto/rand"
    //"github.com/boltdb/bolt"
)

var setup_bolt_data [][]byte = nil

func setupBoltData() ([][]byte) {
    if setup_bolt_data != nil {
        return setup_bolt_data
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

    //fmt.Printf("\nData Size : %d\n", len(data))
    setup_bolt_data = data
    return data
}


func BenchmarkBoltPut(b *testing.B) {
    //db, _ := bolt.Open("/tmp/bolt.db", 0600, nil)
    //defer db.Close()

    data := setupBoltData()
    //fmt.Printf("%s, %x\n", db, data)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        //fmt.Printf("%x\n", data[i%len(data)])
        BoltPut(([]byte)("TestBucket"), data[i%len(data)], data[i%len(data)])
        //put_err := Put(db, ([]byte)("TestBucket"), data[i%len(data)], data[i%len(data)])
        //if put_err != nil {
            //fmt.Printf("%s\n", put_err)
        //}
        //fmt.Printf("%x\n", data[i%len(data)])
    }
}

func BenchmarkBoltGet(b *testing.B) {
    //db, _ := bolt.Open("/tmp/bolt.db", 0600, nil)
    //defer db.Close()
    
    data := setupBoltData()
    //fmt.Printf("%s, %x\n", db, data)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        //fmt.Printf("%x\n", data[i%len(data)])
        BoltGet(([]byte)("TestBucket"), data[i%len(data)])
        //_, get_err := Get(db, ([]byte)("TestBucket"), data[i%len(data)])
        //if get_err != nil {
            //fmt.Printf("%s\n", get_err)
        //}
        //fmt.Printf("%x\n", data[i%len(data)])
    }
}

