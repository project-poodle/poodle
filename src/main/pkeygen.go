package main

import (
    "os"
    "flag"
    //"bufio"
    "fmt"
    //"strings"
    //"syscall"
    "crypto/aes"
    "../util"
)

func check_path_exist(filepath string) bool {
    _, stat_err := os.Stat(filepath)
    if os.IsNotExist(stat_err) {
        return false
    } else {
        return true
    }
}

func generate_key_pair(cls int) (string, []byte, error) {

    priv_key        := util.ECDSAGenerateKey()
    pub_key         := priv_key.PublicKey

    secret, err     := util.PromptPassphrase(true)
    if err != nil {
        return "", nil, err
    }

    padded_secret   := make([]byte, aes.BlockSize)
    copy(padded_secret, secret)

    id := fmt.Sprintf("%x", pub_key.X.Bytes())
    err = util.LCSaveKeyPair(cls, id, pub_key.X, priv_key.D, padded_secret)
    if err != nil {
        return "", nil, err
    }

    return id, padded_secret, nil
}

func process_key(cls int, class string, alias string, force bool) {

    etc_dir         := util.LCGetEtcDir(cls)
    if alias != "" {
        alias_path  := etc_dir + "/" + alias
        if (check_path_exist(alias_path) && !force) {
            input := util.PromptText("Alias [" + alias + "] already exist.  Overwrite? [y/N]")
            if (input != "y" && input != "Y") {
                os.Exit(1)
            }
        }
    }

    fmt.Printf("Generating [%s] pub/priv key pair.\n", class)
    id, padded_secret, err := generate_key_pair(cls)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        os.Exit(1)
    }

    pub_key, err := util.LCLoadPubKey(cls, id)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        os.Exit(1)
    }

    _, err = util.LCLoadPrivKey(cls, id, padded_secret)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        os.Exit(1)
    }

    fmt.Printf("Generated [%s] pub/priv key pair.  Pub key : %x\n", class, pub_key.X.Bytes())

    if alias != "" {
        alias_path  := etc_dir + "/" + alias
        if check_path_exist(alias_path) {
            err := os.Remove(alias_path)
            if err != nil {
                fmt.Printf("ERROR: %s\n", err)
                os.Exit(1)
            }
        }
        key_path    := etc_dir + "/" + id
        err := os.Symlink(key_path, alias_path)
        if err != nil {
            fmt.Printf("ERROR: %s\n", err)
            os.Exit(1)
        }
        fmt.Printf("Created [%s] alias : %s\n", class, alias)
    }
}

func main() {

    classPtr    := flag.String("class", "", "Generate Key for Specified Class [node|cluster|universe|service|federation].")
    aliasPtr    := flag.String("alias", "", "Alias of Generated Key")
    forcePtr    := flag.Bool("force", false, "Whether to Overwrite Existing Alias.")
    flag.Parse()

    cls := 0
    switch *classPtr {
    case "node":
        cls = util.CLS_NODE
    case "cluster":
        cls = util.CLS_CLUSTER
    case "universe":
        cls = util.CLS_UNIVERSE
    case "service":
        cls = util.CLS_SERVICE
    case "federation":
        cls = util.CLS_FEDERATION
    default:
        flag.Usage()
        os.Exit(2)
    }

    process_key(cls, *classPtr, *aliasPtr, *forcePtr)
}
