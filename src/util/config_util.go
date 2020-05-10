package util

import (
    //"os"
    "log"
    "sort"
    "math/big"
    "bytes"
    //"fmt"
    //"time"
    //"io"
    //"io/ioutil"
    //"crypto/aes"
    //"crypto/rand"
    //"crypto/cipher"
    //"crypto/ecdsa"
    //"github.com/boltdb/bolt"
    "github.com/golang/protobuf/proto"
    "../proto_cluster"
)


type ConfigNodes struct {
    allowed_nodes   map[string]proto_cluster.NodeAction
    denied_nodes    map[string]proto_cluster.NodeAction
    unknown_nodes   map[string]proto_cluster.NodeAction
}

var config_nodes *ConfigNodes = new(ConfigNodes)



func LoadConfigNodeList() (*proto_cluster.NodeActionConfigList, error) {

    cluster_pub_key, err := LCLoadClusterPubKey()
    if err != nil {
        return nil, err
    }

    data, data_err  := BoltGet([]byte("config"), []byte("node_list"))
    if data_err != nil {
        return nil, data_err
    }

    node_list       := new(proto_cluster.NodeActionConfigList)
    unmarshal_err   := proto.Unmarshal(data, node_list)
    if unmarshal_err != nil {
        return nil, unmarshal_err
    }

    for i:=0; i<len(node_list.Nodes); i++ {
        msg         := node_list.Nodes[i]

        node                := msg.Node
        node_data, node_err := proto.Marshal(node)
        if node_err != nil {
            log.Printf("LoadConfigNodeList : %s\n", node_err)
            continue
        }

        node_hash   := SumSHA256(node_data)

        verify_r    := new(big.Int)
        verify_r.SetBytes(msg.VerifyR)
        verify_s    := new(big.Int)
        verify_s.SetBytes(msg.VerifyS)

        verified := ECDSAVerify(cluster_pub_key, node_hash[:], verify_r, verify_s)
        if !verified {
            log.Printf("LoadConfigNodeList : not verified %s\n", node)
            continue
        }

    }

    return node_list, unmarshal_err
}


func SaveConfigNodeList(nodes *proto_cluster.NodeActionConfigList) error {

    sort.Slice(nodes.Nodes, func(i, j int) bool {
        if nodes.Nodes[i].Node.Action > nodes.Nodes[j].Node.Action {
            return true
        }
        if nodes.Nodes[j].Node.Action > nodes.Nodes[i].Node.Action {
            return false
        }
        return bytes.Compare(nodes.Nodes[i].Node.NodeId, nodes.Nodes[j].Node.NodeId) > 0
    })

    data, data_err := proto.Marshal(nodes)
    if data_err != nil {
        return data_err
    }

    return BoltPut([]byte("config"), []byte("node_list"), data)
}




