package util

import (
    "os"
    "fmt"
    "log"
    "math/big"
    "time"
    "bytes"
    "io"
    "io/ioutil"
    "crypto/aes"
    "crypto/rand"
    "crypto/cipher"
    "crypto/ecdsa"
    //"github.com/boltdb/bolt"
    "github.com/howeyc/gopass"
    //"github.com/golang/protobuf/proto"
    //"../proto_config"
)


const (
    DEFAULT_NODE_SECRET     = "poodlefs"
    DEFAULT_CLUSTER_PORT    = 31416
    DEFAULT_CONTAINER_PORT  = 14159

    DEFAULT_REFRESH_NANOS   = 15 * 1e9
)


//var bolt_db *bolt.DB = nil

var node_secret string  = ""

var node_public_key                 *ecdsa.PublicKey    = nil
var node_public_key_load_time       int64               = -1

var node_private_key                *ecdsa.PrivateKey   = nil
var node_private_key_load_time      int64               = -1

var cluster_public_key              *ecdsa.PublicKey    = nil
var cluster_public_key_load_time    int64               = -1


/*
func TestEq(a, b []byte) bool {

    if a == nil && b == nil { 
        return true; 
    }

    if a == nil || b == nil { 
        return false; 
    }

    if len(a) != len(b) {
        return false
    }

    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }

    return true
}
*/

func lc_get_etc_directory() string {
    return "/etc/poodle/"
}


func lc_get_etc_config_filepath() string {
    return "/etc/poodle/poodle.conf"
}


func lc_get_log_file() string {
    return "/var/log/poodle.log"
}


func lc_get_bolt_db_file() string {
    return "/var/lib/poodle/bolt.db"
}


func lc_get_badger_db_inode_dir() string {
    return "/var/lib/poodle/inode/"
}


func lc_get_badger_db_container_dir() string {
    return "/var/lib/poodle/container/"
}


func lc_exist_key_file(name string, kind string) bool {
    filepath := lc_get_etc_directory() + name + "." + kind
    if _, stat_err := os.Stat(filepath); stat_err == nil {
        return true
    } else {
        return false
    }
}

// pub key are not encrypted
func lc_load_pub_key(name string) (*ecdsa.PublicKey, error) {
    filepath := lc_get_etc_directory() + name + ".pub"
    if _, stat_err := os.Stat(filepath); os.IsNotExist(stat_err) {
        return nil, stat_err
    }

    data, err := ioutil.ReadFile(filepath)
    if err != nil {
        return nil, err
    }

    return ECDSAGetPublicKey(data), nil
}

// priv key are always encrypted
func lc_load_priv_key(name string, secret []byte) (*ecdsa.PrivateKey, error) {
    filepath := lc_get_etc_directory() + name + ".key"
    if _, stat_err := os.Stat(filepath); os.IsNotExist(stat_err) {
        return nil, stat_err
    }

    data, err := ioutil.ReadFile(filepath)
    if err != nil {
        return nil, err
    }

    key         := []byte(secret)
    nonce       := data[:12]
    ciphertext  := data[12:]

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    aesgcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }

    return ECDSAGetPrivateKey(plaintext), nil
}

func lc_save_pub_key(name string, pub_key *big.Int) error {

    pub_key_filepath    := lc_get_etc_directory() + name + ".pub"

    pub_key_bytes   := ToByteArray32(pub_key)

    pub_write_err   := ioutil.WriteFile(pub_key_filepath, pub_key_bytes, 0644)
    if pub_write_err != nil {
        return pub_write_err
    }

    return nil
}

func lc_save_key_pair(name string, pub_key, priv_key *big.Int, secret []byte) error {

    priv_key_filepath   := lc_get_etc_directory() + name + ".key"
    pub_key_filepath    := lc_get_etc_directory() + name + ".pub"

    block, err := aes.NewCipher(secret)
    if err != nil {
        return err
    }

    // Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
    nonce := make([]byte, 12)
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return err
    }

    aesgcm, err := cipher.NewGCM(block)
    if err != nil {
        return err
    }

    priv_key_bytes  := ToByteArray32(priv_key)
    pub_key_bytes   := ToByteArray32(pub_key)

    ciphertext  := aesgcm.Seal(nil, nonce, priv_key_bytes, nil)
    priv_data   := append(nonce[:], ciphertext[:]...)
    //fmt.Printf("%x\n", priv_data)

    priv_write_err  := ioutil.WriteFile(priv_key_filepath,  priv_data, 0600)
    if priv_write_err != nil {
        return priv_write_err
    }

    pub_write_err   := ioutil.WriteFile(pub_key_filepath,   pub_key_bytes, 0644) 
    if pub_write_err != nil {
        return pub_write_err
    }

    return nil
}

func PromptPassword(repeat bool) ([]byte, error) {

    //terminal_state, _ := terminal.GetState(int(syscall.Stdin))
    //defer terminal.Restore(int(syscall.Stdin), terminal_state)
    //reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter Password: ")
    bytePassword, err := gopass.GetPasswdMasked()
    //fmt.Println()
    if err != nil {
        return nil, fmt.Errorf("Error: %s", err)
    }

    password := string(bytePassword)
    if len(password) < 8 {
        return nil, fmt.Errorf("Error: password length too short (%d)", len(password))
    }

    if repeat {
        fmt.Print("Re-Enter Password: ")
        bytePassword2, err := gopass.GetPasswdMasked()
        //fmt.Println()
        if err != nil {
            return nil, fmt.Errorf("Error: %s", err)
        }

        if !bytes.Equal(bytePassword, bytePassword2) {
            return nil, fmt.Errorf("Error: password NOT match.")
        }
    }

    return bytePassword, nil
}

var log_file *os.File = nil

func LCSetLoggerStdout() error {
    log.SetOutput(os.Stdout)
    log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
    return nil
}

func LCSetLoggerFile() error {
    // if filename exist
    var file *os.File
    if log_file == nil {
        var err error
        file, err = os.OpenFile(lc_get_log_file(), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
        if err != nil {
            return err
        }
        log_file = file
    }
    log.SetOutput(file)
    log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds)
    return nil
}

func LCExistClusterPubKey() bool {
    return lc_exist_key_file("cluster", "pub")
}

func LCExistClusterPrivKey() bool {
    return lc_exist_key_file("cluster", "key")
}

// ECDSA pub key (X)
func LCLoadClusterPubKey() (*ecdsa.PublicKey, error) {
    pub_key, err := lc_load_pub_key("cluster")
    if err != nil {
        return nil, err
    }

    cluster_public_key              = pub_key
    cluster_public_key_load_time    = time.Now().UnixNano()

    return pub_key, nil
}

func LCGetClusterPubKey() (*ecdsa.PublicKey, error) {
    if cluster_public_key != nil {
        t := time.Now().UnixNano()
        rand_interval, err := RandInt64Range(0,  DEFAULT_REFRESH_NANOS)
        if err == nil && t - cluster_public_key_load_time >= rand_interval {
            // consider cache expired, reload needed
            return LCLoadClusterPubKey()
        } else {
            // return cache value if within 15 seconds
            return cluster_public_key, nil
        }
    }

    return LCLoadClusterPubKey()
}

// ECDSA priv key
func LCLoadClusterPrivKey(cluster_secret []byte) (*ecdsa.PrivateKey, error) {
    // NEVER keep cluster private key in fixed memory variable
    return lc_load_priv_key("cluster", cluster_secret)
}

// ECDSA priv key
func LCLoadClusterPrivKeyPrompt() (*ecdsa.PrivateKey, error) {
    if !LCExistClusterPrivKey() {
        return nil, fmt.Errorf("cluster priv key not exist.")
    }

    cluster_secret, err := PromptPassword(false)
    if err != nil {
        return nil, err
    }

    // NEVER keep cluster private key in fixed memory variables
    return lc_load_priv_key("cluster", cluster_secret)
}

// ECDSA pub key
func LCSaveClusterPubKey(pub_key *big.Int) (error) {
    err := lc_save_pub_key("cluster", pub_key)
    if err != nil {
        return err
    }

    cluster_public_key              = ECDSAGetPublicKey(pub_key.Bytes())
    cluster_public_key_load_time    = time.Now().UnixNano()

    return nil
}

// ECDSA pub / priv keys
func LCSaveClusterKeyPair(pub_key, priv_key *big.Int, cluster_secret []byte) (error) {
    err := lc_save_key_pair("cluster", pub_key, priv_key, cluster_secret)
    if err != nil {
        return err
    }

    cluster_public_key              = ECDSAGetPublicKey(pub_key.Bytes())
    cluster_public_key_load_time    = time.Now().UnixNano()

    return nil
}


func SetupNodeSecret(ns_cli string, use_default bool, prompt bool) {
    if ns_cli != "" {
        SetNodeSecret(ns_cli)
        return
    }

    ns_env := os.Getenv("NODE_SECRET")
    if ns_env != "" {
        SetNodeSecret(ns_env)
        return
    }

    if prompt {
        ns_prompt := PromptNodeSecret()
        SetNodeSecret(ns_prompt)
    }

    if use_default {
        SetNodeSecret(DEFAULT_NODE_SECRET)
    }
}


func SetNodeSecret(secret string) {
    node_secret = secret
}

func GetNodeSecret() string {
    return node_secret
}

func PromptNodeSecret() string {
    node_secret, err := PromptPassword(false)
    if err != nil {
        return ""
    }
    return string(node_secret)
}

func LCExistNodePubKey() bool {
    return lc_exist_key_file("node", "pub")
}

func LCExistNodePrivKey() bool {
    return lc_exist_key_file("node", "key")
}

// ECDSA pub key (X)
func LCLoadNodePubKey() (*ecdsa.PublicKey, error) {
    pub_key, err := lc_load_pub_key("node")
    if err != nil {
        return nil, err
    }

    node_public_key             = pub_key
    node_public_key_load_time   = time.Now().UnixNano()

    return pub_key, nil
}

func LCGetNodePubKey() (*ecdsa.PublicKey,error) {
    if node_public_key != nil {
        t := time.Now().UnixNano()
        rand_interval, err := RandInt64Range(0,  DEFAULT_REFRESH_NANOS)
        if err == nil && t - node_public_key_load_time >= rand_interval {
            // consider cache expired, reload needed
            return LCLoadNodePubKey()
        } else {
            // return cache value if within 15 seconds
            return node_public_key, nil
        }
    }

    return LCLoadNodePubKey()
}


// ECDSA priv key
func LCLoadNodePrivKey(node_secret []byte) (*ecdsa.PrivateKey, error) {
    //fmt.Printf("here1 : %x\n", node_secret)
    priv_key, err := lc_load_priv_key("node", node_secret)
    if err != nil {
        return nil, err
    }

    node_private_key            = priv_key
    node_private_key_load_time  = time.Now().UnixNano()

    return priv_key, nil
}

// ECDSA priv key
func LCLoadNodePrivKeyPrompt() (*ecdsa.PrivateKey, error) {
    if !LCExistNodePrivKey() {
        return nil, fmt.Errorf("node priv key not exist.")
    }
    
    node_secret, err := PromptPassword(false)
    if err != nil {
        return nil, err
    }

    node_secret = AESPad(node_secret)

    //fmt.Printf("here2 : %x\n", node_secret)
    priv_key, err := lc_load_priv_key("node", node_secret)
    if err != nil {
        return nil, err
    }

    node_private_key            = priv_key
    node_private_key_load_time  = time.Now().UnixNano()

    return priv_key, nil
}

func LCGetNodePrivKey() (*ecdsa.PrivateKey, error) {
    if node_private_key != nil {
        t := time.Now().UnixNano()
        rand_interval, err := RandInt64Range(0,  DEFAULT_REFRESH_NANOS)
        if err == nil && t - node_private_key_load_time >= rand_interval {
            // consider cache expired, reload needed
            return LCLoadNodePrivKey([]byte(GetNodeSecret()))
        } else {
            // return cache value if within 15 seconds
            return node_private_key, nil
        }
    }

    return LCLoadNodePrivKey([]byte(GetNodeSecret()))
}


// ECDSA pub key
func LCSaveNodePubKey(pub_key *big.Int) (error) {
    err := lc_save_pub_key("node", pub_key)
    if err != nil {
        return err
    }

    node_public_key = ECDSAGetPublicKey(pub_key.Bytes())
    node_public_key_load_time   = time.Now().UnixNano()

    return nil
}

// ECDSA pub / priv keys
func LCSaveNodeKeyPair(pub_key, priv_key *big.Int, node_secret []byte) (error) {
    err := lc_save_key_pair("node", pub_key, priv_key, node_secret)
    if err != nil {
        return err
    }

    node_private_key            = ECDSAGetPrivateKey(priv_key.Bytes())
    node_private_key_load_time  = time.Now().UnixNano()

    node_public_key             = ECDSAGetPublicKey(pub_key.Bytes())
    node_public_key_load_time   = time.Now().UnixNano()

    return nil
}
