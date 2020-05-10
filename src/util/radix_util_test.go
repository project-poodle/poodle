package util

import (
    "log"
    "fmt"
    "time"
    //"unsafe"
    "testing"
    "strings"
    //"crypto/aes"
    //"crypto/rand"
    //"encoding/binary"
    //"github.com/boltdb/bolt"
    //"github.com/golang/protobuf/proto"
    "crypto/ecdsa"
    "../proto_cluster"
)

var cluster_priv_key    *ecdsa.PrivateKey
var cluster_pub_key     *ecdsa.PublicKey

func setup_env() {
    cluster_priv_key    = ECDSAGenerateKey()
    cluster_pub_key     = &cluster_priv_key.PublicKey
}

type NodeData struct {                  // impements IRadixData, IRadixExtension
    key             []byte
    timestamp       int64
    action          int32
    r               []byte
    s               []byte
}

func NewNodeData(key []byte, timestamp int64, action int32, r, s []byte) *NodeData {
    result              := new(NodeData)
    result.key          = key
    result.timestamp    = timestamp
    result.action       = action
    result.r            = r
    result.s            = s
    return result
}

func (n *NodeData) Key() []byte {
    return n.key
}

func (n *NodeData) Value() []byte {
    buf := make([]byte, 8 + 4 + 32 + 32)
    copy(buf, append(Int64ToByteArray(n.timestamp), Int32ToByteArray(n.action)...))
    copy(buf[12:], append(n.r, n.s...))
    return buf
}

func (n *NodeData) Extension() IRadixExtension {
    return n
}

func (n *NodeData) CanReplaceBy(data IRadixData) bool {
    // TODO
    return true
}

func (n *NodeData) Print(indent int) {
    padding := strings.Repeat(RADIX_PRINT_INDENT, int(indent))
    fmt.Printf(padding + "KEY           : %x\n", n.key)
    fmt.Printf(padding + "Timestamp     : %d\n", n.timestamp)
    fmt.Printf(padding + "Action        : %d\n", n.action)
    fmt.Printf(padding + "R             : %x\n", n.r)
    fmt.Printf(padding + "S             : %x\n", n.s)
}

func TestRecordUpsert(t *testing.T) {

    setup_env()

    root := NewRadixRoot(4)

    for i:=0; i<2000; i++ {

        node_priv_key       := ECDSAGenerateKey()
        node_pub_key        := node_priv_key.PublicKey

        buf := make([]byte, 32 + 8 + 8)
        copy(buf, ToByteArray32(node_pub_key.X))
        buf[40]     = byte(proto_cluster.NodeActionEnum_ALLOW)

        //log.Printf("%x\n", buf)

        hash := SumSHA256(buf)

        sig_r, sig_s, err := ECDSASign(cluster_priv_key, hash[:])
        if err != nil {
            log.Fatalf("Signature error : %s\n", err)
        }

        //copy(buf[48:], ToByteArray32(sig_r))
        //copy(buf[80:], ToByteArray32(sig_s))

        data := NewNodeData(node_pub_key.X.Bytes(), time.Now().UnixNano(), 1, ToByteArray32(sig_r), ToByteArray32(sig_s))
        //data.Print(0)

        ret_code := root.Upsert(data)
        log.Printf("Upsert return Code : %d", ret_code)
        //root.Print(0)
    }

    root.Print(0)
}
