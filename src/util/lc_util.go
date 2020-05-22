package util

import (
    "os"
    "fmt"
    "bufio"
    "log"
    "math/big"
    "strconv"
    "strings"
    "bytes"
    "io"
    "io/ioutil"
    "encoding/binary"
    "encoding/hex"
    "hash/crc32"
    "crypto/aes"
    "crypto/rand"
    "crypto/cipher"
    "crypto/ecdsa"
    //"github.com/boltdb/bolt"
    "github.com/howeyc/gopass"
)


////////////////////////////////////////////////////////////////////////////////
// Util Functions

func PromptText(prompt string) string {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print(prompt + " : ")
    text, _ := reader.ReadString('\n')
    return strings.TrimSpace(text)
}

func PromptPassphrase(repeat bool) ([]byte, error) {

    //terminal_state, _ := terminal.GetState(int(syscall.Stdin))
    //defer terminal.Restore(int(syscall.Stdin), terminal_state)
    //reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter Passphrase: ")
    bytePassword, err := gopass.GetPasswdMasked()
    //fmt.Println()
    if err != nil {
        return nil, fmt.Errorf("Error: %s", err)
    }

    password := string(bytePassword)
    if len(password) < 8 {
        return nil, fmt.Errorf("Error: passphrase length too short (%d)", len(password))
    }

    if repeat {
        fmt.Print("Re-Enter Passphrase: ")
        bytePassword2, err := gopass.GetPasswdMasked()
        //fmt.Println()
        if err != nil {
            return nil, fmt.Errorf("Error: %s", err)
        }

        if !bytes.Equal(bytePassword, bytePassword2) {
            return nil, fmt.Errorf("Error: passphrase NOT match.")
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

func LCSetLoggerFileTmp() error {
    // if filename exist
    var file *os.File
    if log_file == nil {
        var err error
        file, err = os.OpenFile(lc_get_log_file_tmp(), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
        if err != nil {
            return err
        }
        log_file = file
    }
    log.SetOutput(file)
    log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds)
    return nil
}

////////////////////////////////////////////////////////////////////////////////
// Key Functions

func LCGetEtcDir(cls int) string {
    return lc_get_etc_dir(cls)
}

func LCExistPubKey(cls int, id string) bool {
    filepath := lc_get_pub_key_path(cls, id)
    if _, err := os.Stat(filepath); err == nil {
        return true
    } else {
        return false
    }
}

func LCExistPrivKey(cls int, id string) bool {
    filepath := lc_get_priv_key_path(cls, id)
    if _, err := os.Stat(filepath); err == nil {
        return true
    } else {
        return false
    }
}

// ECDSA pub key (X) for node
func LCLoadPubKey(cls int, id string) (*ecdsa.PublicKey, error) {
    return lc_load_pub_key(cls, id)
}

// ECDSA priv key for node
func LCLoadPrivKey(cls int, id string, secret []byte) (*ecdsa.PrivateKey, error) {
    return lc_load_priv_key(cls, id, secret)
}

// ECDSA priv key for node
func LCLoadPrivKeyPrompt(cls int, id string) (*ecdsa.PrivateKey, error) {
    if !LCExistPrivKey(cls, id) {
        return nil, fmt.Errorf("priv key not exist [" + strconv.Itoa(cls) + "] : " + id)
    }

    secret, err := PromptPassphrase(false)
    if err != nil {
        return nil, err
    }

    // do not keep private key in fixed memory variables
    return lc_load_priv_key(cls, id, secret)
}

// ECDSA pub / priv keys for node
func LCSaveKeyPair(cls int, id string, pub_key, priv_key *big.Int, secret []byte) (error) {
    return lc_save_key_pair(cls, id, pub_key, priv_key, secret)
}

// ECDSA pub / priv keys for node
func LCSaveKeyPairPrompt(cls int, id string, pub_key, priv_key *big.Int) (error) {
    node_secret, err := PromptPassphrase(false)
    if err != nil {
        return err
    }

    return lc_save_key_pair(cls, id, pub_key, priv_key, node_secret)
}


////////////////////////////////////////////////////////////////////////////////
// Private Functions

func lc_get_etc_dir(cls int) string {
    switch cls {
    case CLS_NODE:
        return DEFAULT_ETC_DIR + "/node"
    case CLS_CLUSTER:
        return DEFAULT_ETC_DIR + "/cluster"
    case CLS_UNIVERSE:
        return DEFAULT_ETC_DIR + "/universe"
    case CLS_SERVICE:
        return DEFAULT_ETC_DIR + "/service"
    case CLS_FEDERATION:
        return DEFAULT_ETC_DIR + "/federation"
    default:
        panic("Unknown CLS : " + strconv.Itoa(cls))
    }
}

func lc_get_key_dir(cls int, id string) string {
    return lc_get_etc_dir(cls) + "/" + id
}

func lc_get_pub_key_path(cls int, id string) string {
    switch cls {
    case CLS_NODE:
        return lc_get_key_dir(cls, id) + "/node_id.pub"
    case CLS_CLUSTER:
        return lc_get_key_dir(cls, id) + "/cluster_id.pub"
    case CLS_UNIVERSE:
        return lc_get_key_dir(cls, id) + "/universe_id.pub"
    case CLS_SERVICE:
        return lc_get_key_dir(cls, id) + "/service_id.pub"
    case CLS_FEDERATION:
        return lc_get_key_dir(cls, id) + "/federation_id.pub"
    default:
        panic("Unknown CLS : " + strconv.Itoa(cls))
    }
}

func lc_get_priv_key_path(cls int, id string) string {
    switch cls {
    case CLS_NODE:
        return lc_get_key_dir(cls, id) + "/node_id"
    case CLS_CLUSTER:
        return lc_get_key_dir(cls, id) + "/cluster_id"
    case CLS_UNIVERSE:
        return lc_get_key_dir(cls, id) + "/universe_id"
    case CLS_SERVICE:
        return lc_get_key_dir(cls, id) + "/service_id"
    case CLS_FEDERATION:
        return lc_get_key_dir(cls, id) + "/federation_id"
    default:
        panic("Unknown CLS : " + strconv.Itoa(cls))
    }
}

// key data encoded as hex dump, with CRC32 at the end
func lc_load_key_data(filepath string) ([]byte, error) {

    if _, stat_err := os.Stat(filepath); os.IsNotExist(stat_err) {
        return nil, stat_err
    }

    input, err := ioutil.ReadFile(filepath)
    if err != nil {
        return nil, err
    }

    // TODO - decode
	output  := make([]byte, hex.DecodedLen(len(input)))
	n, err  := hex.Decode(output, input)
	if err != nil {
		return nil, err
	}

    data        := output[:n-4]
    compute_crc := make([]byte, 4)
    binary.BigEndian.PutUint32(compute_crc, crc32.ChecksumIEEE(data))

    content_crc := output[n-4:n]
    if !EqByteArray(compute_crc, content_crc) {
        return nil, fmt.Errorf("[%s] computed CRC does not match content CRC", filepath)
    }

    // fmt.Printf("%s\n", filepath)
    // fmt.Printf("    data : %x\n", data)
    // fmt.Printf("    compute crc : %x\n", compute_crc)
    // fmt.Printf("    content crc : %x\n", content_crc)

    return data, nil
}

// pub key are not encrypted
func lc_load_pub_key(cls int, id string) (*ecdsa.PublicKey, error) {

    filepath := lc_get_pub_key_path(cls, id)
    data, err := lc_load_key_data(filepath)
    if err != nil {
        return nil, err
    }

    return ECDSAGetPublicKey(data), nil
}

// priv key are always encrypted
func lc_load_priv_key(cls int, id string, secret []byte) (*ecdsa.PrivateKey, error) {

    filepath := lc_get_priv_key_path(cls, id)
    data, err := lc_load_key_data(filepath)
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

func lc_mkdir_if_not_exist(path string) error {

    if _, stat_err := os.Stat(path); os.IsNotExist(stat_err) {
        os.MkdirAll(path, 0755)
    }

    // check again for the path
    stat_info, _ := os.Stat(path)
    if (!stat_info.Mode().IsDir()) {
        return fmt.Errorf("Path is not a directory : %s", path)
    }

    return nil
}

func lc_save_pub_key(cls int, id string, pub_key *big.Int) error {

    dir := lc_get_key_dir(cls, id)
    if err := lc_mkdir_if_not_exist(dir); err != nil {
        return err
    }

    pub_key_filepath    := lc_get_pub_key_path(cls, id)

    pub_key_bytes       := ToByteArray32(pub_key)
    pub_key_crc         := make([]byte, 4)
    binary.BigEndian.PutUint32(pub_key_crc, crc32.ChecksumIEEE(pub_key_bytes))
    pub_data            := append(pub_key_bytes[:], pub_key_crc[:]...)

    pub_write_err   := ioutil.WriteFile(pub_key_filepath, []byte(fmt.Sprintf("%X", pub_data)), 0644)
    if pub_write_err != nil {
        return pub_write_err
    }

    return nil
}

func lc_save_key_pair(cls int, id string, pub_key, priv_key *big.Int, secret []byte) error {

    dir := lc_get_key_dir(cls, id)
    if err := lc_mkdir_if_not_exist(dir); err != nil {
        return err
    }

    pub_key_filepath    := lc_get_pub_key_path(cls, id)
    priv_key_filepath   := lc_get_priv_key_path(cls, id)

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

    pub_key_bytes       := ToByteArray32(pub_key)
    pub_key_crc         := make([]byte, 4)
    binary.BigEndian.PutUint32(pub_key_crc, crc32.ChecksumIEEE(pub_key_bytes))
    pub_data            := append(pub_key_bytes[:], pub_key_crc[:]...)

    priv_key_bytes      := ToByteArray32(priv_key)
    ciphertext          := aesgcm.Seal(nil, nonce, priv_key_bytes, nil)
    priv_encrypted      := append(nonce[:], ciphertext[:]...)
    priv_encrypted_crc  := make([]byte, 4)
    binary.BigEndian.PutUint32(priv_encrypted_crc, crc32.ChecksumIEEE(priv_encrypted))
    priv_data           := append(priv_encrypted[:], priv_encrypted_crc[:]...)

    //fmt.Printf("%x\n", priv_data)

    priv_write_err  := ioutil.WriteFile(priv_key_filepath, []byte(fmt.Sprintf("%X", priv_data)), 0600)
    if priv_write_err != nil {
        return priv_write_err
    }

    pub_write_err   := ioutil.WriteFile(pub_key_filepath, []byte(fmt.Sprintf("%X", pub_data)), 0644)
    if pub_write_err != nil {
        return pub_write_err
    }

    return nil
}

func lc_get_node_conf_path(node_id string) string {
    return lc_get_key_dir(CLS_NODE, node_id) + "/node.conf"
}

func lc_get_lib_dir(node_id string) string {
    return DEFAULT_LIB_DIR + "/" + node_id
}

func lc_get_transaction_log_dir(node_id string) string {
    return lc_get_lib_dir(node_id) + "/transaction"
}

func lc_get_domain_dir(node_id, domain_name string) string {
    return lc_get_lib_dir(node_id) + "/domain/" + domain_name
}

func lc_get_log_dir(node_id string) string {
    return DEFAULT_LOG_DIR + "/" + node_id
}

func lc_get_log_file_tmp() string {
    return "/tmp/" + strconv.Itoa(os.Getpid()) + ".log"
}