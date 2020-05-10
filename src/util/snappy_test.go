package file

import (
    "fmt"
    //"os"
    "testing"
    "crypto/rand"
    "github.com/golang/snappy"
)

var setup_snappy_data [][]byte = nil
var setup_snappy_compressed [][]byte = nil

func setupSnappyData() ([][]byte, [][]byte) {
    if setup_snappy_data != nil && setup_snappy_compressed != nil {
        return setup_snappy_data, setup_snappy_compressed
    }

    var data_size = 0
    var compressed_size = 0

    data        := make([][]byte, 64 * 1024)
    compressed  := make([][]byte, 64 * 1024)
    for i:=0; i<len(data); i++ {
        data[i] = make([]byte, 4 * 1024)
        for j:=0; j<len(data[i]); {
            if n, err := rand.Read(data[i]); err != nil {
                fmt.Printf("%s\n", err)
            } else {
                j += n
            }
        }
        compressed[i]   = snappy.Encode(nil, data[i])

        data_size       += len(data[i])
        compressed_size += len(compressed[i])
    }

    //fmt.Printf("\nData Size : %d\n", len(data))
    fmt.Printf("\nData Size : %d , Compressed Size : %d\n", data_size, compressed_size)
    setup_snappy_data = data
    setup_snappy_compressed = compressed
    return data, compressed
}

func TestSetup(t *testing.T) {
    setupSnappyData()
}

func BenchmarkCompress(b *testing.B) {
    data, _ := setupSnappyData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        //fmt.Printf("%x\n", data[i%len(data)])
        snappy.Encode(nil, data[i % len(data)])
    }
}

func BenchmarkDecompress(b *testing.B) {
    _, compressed := setupSnappyData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        //fmt.Printf("%x\n", compressed[i%len(compressed)])
        _, err := snappy.Decode(nil, compressed[i % len(compressed)])
        if (err != nil) {
            fmt.Printf("%s : %x\n", err, compressed[i%len(compressed)])
        }
    }
}
