package main

import (
    "os"
    "flag"
    "fmt"
    "../util"
)

func main() {

    typePtr     := flag.String("type", "node", "Generate Key for Node, Cluster, Service, or Universe [node|cluster|service|universe].")
    forcePtr    := flag.Bool("force", false, "Whether to overwrite existing keys.")
    flag.Parse()

    if *typePtr != "node" && *typePtr != "cluster" {
        flag.Usage()
        os.Exit(2)
    }

    if *typePtr == "node" {
        process_node_key(*forcePtr)
    } else {
        process_cluster_key(*forcePtr)
    }
}
