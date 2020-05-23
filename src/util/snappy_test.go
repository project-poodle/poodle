package util

import (
	"fmt"
	//"os"
	"crypto/rand"
	"testing"

	"github.com/golang/snappy"
)

var setupSnappyData [][]byte
var setupSnappyCompressed [][]byte

func SetupSnappyData() ([][]byte, [][]byte) {
	if setupSnappyData != nil && setupSnappyCompressed != nil {
		return setupSnappyData, setupSnappyCompressed
	}

	var dataSize = 0
	var compressedSize = 0

	data := make([][]byte, 4*1024)
	compressed := make([][]byte, 4*1024)
	for i := 0; i < len(data); i++ {
		data[i] = make([]byte, 4*1024)
		for j := 0; j < len(data[i]); {
			if n, err := rand.Read(data[i]); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				j += n
			}
		}
		compressed[i] = snappy.Encode(nil, data[i])

		dataSize += len(data[i])
		compressedSize += len(compressed[i])
	}

	//fmt.Printf("\nData Size : %d\n", len(data))
	fmt.Printf("\nData Size : %d , Compressed Size : %d\n", dataSize, compressedSize)
	setupSnappyData = data
	setupSnappyCompressed = compressed
	return data, compressed
}

func TestSetupSnappy(t *testing.T) {
	SetupSnappyData()
}

func BenchmarkCompress(b *testing.B) {
	data, _ := SetupSnappyData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//fmt.Printf("%x\n", data[i%len(data)])
		snappy.Encode(nil, data[i%len(data)])
	}
}

func BenchmarkDecompress(b *testing.B) {
	_, compressed := SetupSnappyData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//fmt.Printf("%x\n", compressed[i%len(compressed)])
		_, err := snappy.Decode(nil, compressed[i%len(compressed)])
		if err != nil {
			fmt.Printf("%s : %x\n", err, compressed[i%len(compressed)])
		}
	}
}
