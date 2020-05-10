package util

import (
    //"log"
    "fmt"
    "time"
    //"unsafe"
    "testing"
    //"crypto/aes"
    //"crypto/rand"
    //"github.com/boltdb/bolt"
    //"github.com/golang/protobuf/proto"
    //"crypto/ecdsa"
    //"../proto_config"
)




func prepare_PUDP() *PUDP {

    pudp := new(PUDP)

    pudp.Version        = 1
    pudp.PHL            = 100
    pudp.Algo_EC        = 0
    pudp.Algo_SIG       = 0
    pudp.Algo_ENC       = 0
    pudp.Reserved_1     = 0
    pudp.MSG_ID         = 512

    pudp.Timestamp      = time.Now().UnixNano()
    return pudp
}

func prepare_PUDP_buffer() []byte {
    pudp := prepare_PUDP()
    buf, err := pudp.Marshal()
    if err != nil {
        panic(err)
    }
    //fmt.Printf("PDUP Marshal : %x\n", buf)
    return buf
}

func TestMarshal(t *testing.T) {
    pudp        := prepare_PUDP()
    //buf     = prepare_PUDP_buffer()
    buf, err    := pudp.Marshal()
    if err != nil {
        panic(err)
    }
    fmt.Printf("PDUP Marshal : %x\n", buf)
}

func TestNew_PDUP(t *testing.T) {
    //pudp    = prepare_PDUP()
    buf         := prepare_PUDP_buffer()
    pudp, err   := New_PUDP(buf)
    if err != nil {
        panic(err)
    }
    fmt.Printf("PDUP : %x\n", pudp)
}

