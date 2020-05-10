package util

import (
    "log"
    "fmt"
    "time"
    "unsafe"
    "testing"
    "crypto/aes"
    //"crypto/rand"
    //"github.com/boltdb/bolt"
    "github.com/golang/protobuf/proto"
    "crypto/ecdsa"
    "../proto_cluster"
)

func TestConfigNodeList(t *testing.T) {

    secret              := "poodlefs"
    padded_secret       := make([]byte, aes.BlockSize)
    copy(padded_secret[:], secret)

    cluster_priv_key, err       := LCLoadClusterPrivKey(padded_secret)
    if err != nil {
        panic(err)
    }

    node_list   := new(proto_cluster.NodeActionConfigList)

    allowed_count   := 40960
    denied_count    := 2048

    node_list.Nodes = make([]*proto_cluster.NodeActionConfig, allowed_count + denied_count)
    for i:=0; i<allowed_count; i++ {
        node_priv_key    := ECDSAGenerateKey()
        node_pub_key     := node_priv_key.Public().(*ecdsa.PublicKey)

        node_msg                := new(proto_cluster.NodeActionConfig)
        node_msg.Node           = new(proto_cluster.NodeAction)
        node_msg.Node.NodeId    = node_pub_key.X.Bytes()
        node_msg.Node.Timestamp = time.Now().UnixNano()
        node_msg.Node.Action    = proto_cluster.NodeActionEnum_ALLOW
        node_bytes, err         := proto.Marshal(node_msg.Node)
        if err != nil {
            panic(err)
        }

        node_hash               := SumSHA256(node_bytes)
        r, s, err               := ECDSASign(cluster_priv_key, node_hash[:])
        if err != nil {
            panic(err)
        }

        node_msg.VerifyR        = r.Bytes()
        node_msg.VerifyS        = s.Bytes()
        node_list.Nodes[i]      = node_msg
    }

    for i:=allowed_count; i<allowed_count+denied_count; i++ {
        node_priv_key    := ECDSAGenerateKey()
        node_pub_key     := node_priv_key.Public().(*ecdsa.PublicKey)

        node_msg                := new(proto_cluster.NodeActionConfig)
        node_msg.Node           = new(proto_cluster.NodeAction)
        node_msg.Node.NodeId    = node_pub_key.X.Bytes()
        node_msg.Node.Timestamp = time.Now().UnixNano()
        node_msg.Node.Action    = proto_cluster.NodeActionEnum_BLOCK
        node_bytes, err         := proto.Marshal(node_msg.Node)
        if err != nil {
            panic(err)
        }

        node_hash               := SumSHA256(node_bytes)
        r, s, err               := ECDSASign(cluster_priv_key, node_hash[:])
        if err != nil {
            panic(err)
        }

        node_msg.VerifyR        = r.Bytes()
        node_msg.VerifyS        = s.Bytes()
        node_list.Nodes[i]      = node_msg
    }

    //data, err := proto.Marshal(msg)
    //if err != nil {
    //    log.Fatal("marshaling error : ", err)
    //}

    //fmt.Printf("Marshalling : %x\n", data)

    set_err := SaveConfigNodeList(node_list)
    if set_err != nil {
        log.Fatal("SaveConfigNodeList : ", set_err)
    }

    get_nodes, get_err := LoadConfigNodeList()
    if get_err != nil {
        log.Fatal("LoadConfigNodeList : ", get_err)
    }

    fmt.Printf("Nodes : %x\n", unsafe.Sizeof(*get_nodes))
    //fmt.Printf("Nodes : %x\n", get_nodes)
}

