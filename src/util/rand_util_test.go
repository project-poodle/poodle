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

	d := RandInt32()

	fmt.Printf("RandInt32 : %x\n", d)
}
