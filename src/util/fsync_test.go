package util

import (
    //"fmt"
    "os"
    "testing"
)

var setup_file *os.File
var setup_size int


func setupFsync() (* os.File){

    setup_file = Open("test1")
    setup_size = 1 * 1 * 1024 * 1024

    return setup_file
}


func BenchmarkWriteSync(b *testing.B) {
    f := setupFsync()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        //fmt.Printf("%x\n", data[i%len(data)])
        Write(f, setup_size)
        Sync(f)
    }
}

func BenchmarkWrite(b *testing.B) {
    f := setupFsync()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        //fmt.Printf("%x\n", data[i%len(data)])
        Write(f, setup_size)
    }

    Sync(f)
}

