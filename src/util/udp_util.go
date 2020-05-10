package util

import (
    "fmt"
    "log"
    "net"
    "time"
    //"math"
    "bytes"
    //"strconv"
    "encoding/binary"
    //"crypto/rand"
    "crypto/ecdsa"
    //"math/big"
)



var listen_udp_conns = map[int32]*net.UDPConn{}


func AbsInt32(x int32) int32 {
    if x < 0 {
        return -x
    }
    return x
}

func AbsInt64(x int64) int64 {
    if x < 0 {
        return -x
    }
    return x
}


type PUDP struct {
    Version     byte        // 4 bits (first byte, 0-3)     - version number
    PHL         byte        // 4 bits (frist byte, 4-7)     - poodle header length (in 4 bytes - does not include signature or timestamp)
    Algo_EC     byte        // 2 bits (second byte, 0-1)    - Erasure Coding Algorithm  - 0 means no EC
    Algo_SIG    byte        // 2 bits (second byte, 2-3)    - Signature Algorithm       - 0 means no signature
    Algo_ENC    byte        // 2 bits (second byte, 4-5)    - Encryption Algorithm      - 0 means no encryption
    Reserved_1  byte        // 2 bits (second byte, 6-7)    - 2 bits reserved for future expansion
    MSG_ID      uint16      // 3rd-4th byte - message id    - 16 bit message identity

    EC_Base     byte        // 1 byte - base count          - present only if fragmented ec encoding
    EC_Parity   byte        // 1 byte - paraty count        - present only if fragmented ec encoding
    EC_Offset   byte        // 1 byte - current position    - present only if fragmented ec encoding
    Reserved_2  byte        // 1 byte - reserved            - present only if fragmented ec encoding
    Msg_Length  uint32      // length of entire msg         - present only if fragmented ec encoding, represent total msg length (in bytes)

    Pub_Key     []byte      // 256 bits public key X        - present if Algo_SIG != 0
    Signature_r []byte      // 256 bits signature r         - present if Algo_SIG != 0
    Signature_s []byte      // 256 bits signature s         - present if Algo_SIG != 0

    Timestamp   int64       // 64 bit nano timestamp        - present in all messages - not included in header len - included in signature calculation

    Data        []byte      // data
}

func New_PUDP(data []byte) (*PUDP, error) {
    var pudp    = new(PUDP)

    r := bytes.NewReader(data)
    offset := 0

    var merged uint32
    err := binary.Read(r, binary.BigEndian, &merged)  // 32 bits
    if err != nil {
        return nil, err
    }
    offset += 4

    pudp.Version    = byte(merged >> 28 & 0xf)     // bit 0-3 is version
    pudp.PHL        = byte(merged >> 24 & 0xf)     // bit 4-7 is header length
    pudp.Algo_EC    = byte(merged >> 22 & 0x3)     // EC
    pudp.Algo_SIG   = byte(merged >> 20 & 0x3)     // signature
    pudp.Algo_ENC   = byte(merged >> 18 & 0x3)     // encryption
    pudp.Reserved_1 = byte(merged >> 16 & 0x3)     // reserved
    pudp.MSG_ID     = uint16(merged & 0xFFFF)      // msg ID

    if pudp.Algo_EC != 0 {
        err = binary.Read(r, binary.BigEndian, &merged)
        if err != nil {
            return nil, err
        }
        offset += 4

        pudp.EC_Base    = byte(merged >> 24 & 0xff)
        pudp.EC_Parity  = byte(merged >> 16 & 0xff)
        pudp.EC_Offset  = byte(merged >> 8 & 0xff)
        pudp.Reserved_2 = byte(merged & 0xff)

        err = binary.Read(r, binary.BigEndian, &merged)
        if err != nil {
            return nil, err
        }
        offset += 4

        pudp.Msg_Length = merged
    }

    r = bytes.NewReader(data[pudp.PHL * 4:])

    if pudp.Algo_SIG != 0 {
        var pub_key [32]byte
        err = binary.Read(r, binary.BigEndian, &pub_key)
        if err != nil {
            return nil, err
        } else {
            pudp.Pub_Key = pub_key[:]
        }
        offset += 32

        var sig_r [32]byte
        err = binary.Read(r, binary.BigEndian, &sig_r)
        if err != nil {
            return nil, err
        } else {
            pudp.Signature_r = sig_r[:]
        }
        offset += 32

        var sig_s [32]byte
        err = binary.Read(r, binary.BigEndian, &sig_s)
        if err != nil {
            return nil, err
        } else {
            pudp.Signature_s = sig_s[:]
        }
        offset += 32

        public_key := ECDSAGetPublicKey(pudp.Pub_Key)
        hash := SumSHA256(data[offset:])
        if !ECDSAVerify(public_key, hash[:], ToBigInt(sig_r[:]), ToBigInt(sig_s[:])) {
            return nil, fmt.Errorf("UDP ECDSA Verification Failed : %d", pudp.MSG_ID)
        }
    }

    err = binary.Read(r, binary.BigEndian, &pudp.Timestamp)
    if err != nil {
        return nil, err
    }
    offset += 8

    delta := AbsInt64(time.Now().UnixNano() - pudp.Timestamp)
    if delta >= 5*1e9 {
        // delta is more than 5 second
        return nil, fmt.Errorf("UDP Timestamp Out of Range : %d (%d)", pudp.MSG_ID, delta)
    } else if delta >= 1e9 {
        // delta is between 1 second and 5 second
        range_rand, err := RandInt64Range(0, 4 * 1e9)
        if err != nil {
            return nil, err
        }
        // if delta is greater than a random threshold
        if delta - 1e9 > range_rand {
            return nil, fmt.Errorf("UDP Timestamp Out of Range : %d (%d)", pudp.MSG_ID, delta)
        }
    }
    // delta is less than 1 second - ok

    pudp.Data = data[offset:]

    return pudp, nil
}



func (pudp *PUDP) Marshal() ([]byte, error) {


    buf_content := new(bytes.Buffer)
    err := binary.Write(buf_content, binary.BigEndian, pudp.Timestamp)
    if err != nil {
        log.Printf("here1")
        return nil, err
    }

    if pudp.Data != nil {
        err = binary.Write(buf_content, binary.BigEndian, pudp.Data)
        if err != nil {
            log.Printf("here2")
            return nil, err
        }
    }
    //fmt.Printf("Buf length 4: %d, %d\n", buf.Len(), len(buf.Bytes()))

    buf := new(bytes.Buffer)

    if pudp.Version <=1 {
        if pudp.Algo_EC == 0 {
            pudp.PHL = 1
        } else if pudp.Algo_EC == 1 {
            pudp.PHL = 3
        }
    }

    merged := (uint32(pudp.Version) & 0xf << 28) | (uint32(pudp.PHL) & 0xf << 24) | (uint32(pudp.Algo_EC) & 0x3 << 22) | (uint32(pudp.Algo_SIG) & 0x3 << 20) | (uint32(pudp.Algo_ENC) & 0x3 << 18) | (uint32(pudp.Reserved_1) & 0x3 << 16) | (uint32(pudp.MSG_ID) & 0XFFFF)
    //fmt.Printf("Merged: %x\n", merged)
    err = binary.Write(buf, binary.BigEndian, merged)
    if err != nil {
        log.Printf("here3")
        return nil, err
    }
    //fmt.Printf("Buf length 1: %d, %d\n", buf.Len(), len(buf.Bytes()))

    if pudp.Algo_EC == 1 {
        merged = (uint32(pudp.EC_Base) & 0xFF << 24) | (uint32(pudp.EC_Parity) & 0xFF << 16) | (uint32(pudp.EC_Offset) & 0xFF << 8) | (uint32(pudp.Reserved_2) & 0xFF)
        err = binary.Write(buf, binary.BigEndian, merged)
        if err != nil {
            log.Printf("here4")
            return nil, err
        }

        err = binary.Write(buf, binary.BigEndian, pudp.Msg_Length)
        if err != nil {
            log.Printf("here5")
            return nil, err
        }
        //fmt.Printf("Buf length 2: %d, %d\n", buf.Len(), len(buf.Bytes()))
    }

    if pudp.Algo_SIG == 1 {

        var node_priv_key *ecdsa.PrivateKey
        if GetNodeSecret() == "" {
            node_priv_key, err = LCLoadNodePrivKeyPrompt()
        } else {
            node_priv_key, err = LCLoadNodePrivKey(AESPad([]byte(GetNodeSecret())))
        }
        if err != nil {
            return nil, err
        }
        hash := SumSHA256(buf_content.Bytes())
        r, s, err := ECDSASign(node_priv_key, hash[:])
        if err != nil {
            return nil, err
        }

        pudp.Pub_Key        = ToByteArray32(node_priv_key.PublicKey.X)
        pudp.Signature_r    = ToByteArray32(r)
        pudp.Signature_s    = ToByteArray32(s)

        err = binary.Write(buf, binary.BigEndian, pudp.Pub_Key)
        if err != nil {
            log.Printf("here6")
            return nil, err
        }

        err = binary.Write(buf, binary.BigEndian, pudp.Signature_r)
        if err != nil {
            log.Printf("here7")
            return nil, err
        }

        err = binary.Write(buf, binary.BigEndian, pudp.Signature_s)
        if err != nil {
            log.Printf("here8")
            return nil, err
        }

        //fmt.Printf("Buf length 3: %d, %d\n", buf.Len(), len(buf.Bytes()))
    }


    err = binary.Write(buf, binary.BigEndian, buf_content.Bytes())
    if err != nil {
        log.Printf("here9")
        return nil, err
    }

    return buf.Bytes(), nil
}




func UDPGetListenConn(port int32) (*net.UDPConn, error) {
    if conn, ok := listen_udp_conns[port]; ok {
        return conn, nil
    } else {
        sAddr, err := net.ResolveUDPAddr("udp", ":" + fmt.Sprintf("%d", port))
        if err != nil {
            return nil, err
        }

        conn, err := net.ListenUDP("udp", sAddr)
        if err != nil {
            return nil, err
        }

        log.Printf("Listen on UDP Address [%s]\n", sAddr)
        listen_udp_conns[port] = conn
        return conn, nil
    }
}

func udp_recv_message(conn *net.UDPConn) (*net.UDPAddr, []byte, error) {
    buffer := make([]byte, 64*1024)
    n, addr, err := conn.ReadFromUDP(buffer)
    if err != nil {
        return nil, nil, fmt.Errorf("UDP Read Error [%s]: %s\n", addr, err)
    }
    return addr, buffer[:n], nil
}



func UDPListen(port int32, handle_packet func(conn *net.UDPConn, addr *net.UDPAddr, packet []byte)) error {
    conn, err := UDPGetListenConn(port)
    if err != nil {
        return err
    }
    // listen to incoming udp packet
    defer delete(listen_udp_conns, port)
    defer conn.Close()

    for true {

        recv_addr, buf, err := udp_recv_message(conn)
        if err != nil {
            return err
        }

        handle_packet(conn, recv_addr, buf)
        //err = UDPReceive(port, handle_packet)
    }

    return nil
}

// local_port <= 0 means do not bind to udp
func UDPSend(local_conn *net.UDPConn, remote_addr *net.UDPAddr,  packet[] byte) error {
    //simple write
    var conn *net.UDPConn
    var err error
    if local_conn == nil {
        conn, err = net.DialUDP("udp", nil, remote_addr)
        if err != nil {
            return err
        }
    } else {
        conn = local_conn
    }

    // listen to incoming udp packet
    //defer delete(udp_conns, port)
    //defer conn.Close()

    n, err := conn.WriteTo(packet, remote_addr)
    if err != nil {
        return err
    }

    if n != len(packet) {
        return fmt.Errorf("UDP Send length mismatch: %d != %d", len(packet), n)
    }

    return nil
}




func UDPListenRandomPort() (int32, *net.UDPConn, error) {

    retry_count := 64
    for i:=0; i<retry_count; i++ {

        port, err := RandInt32Range(1024, 65535)
        if err != nil {
            log.Printf("%s\n", err)
            continue
        }

        //log.Printf("%d\n", port)

        if _, ok := listen_udp_conns[port]; ok {
            continue
        }

        sAddr, err := net.ResolveUDPAddr("udp", ":" + fmt.Sprintf("%d", port))
        if err != nil {
            continue
        }

        conn, err := net.ListenUDP("udp", sAddr)
        if err != nil {
            continue
        }

        // listen_udp_conns[port] = conn // do not persist this port
        return port, conn, nil
    }

    return -1, nil, fmt.Errorf("cannot find UDP port to listen to [retry=%d]", retry_count)
}



func UDPReceiveMessage(conn *net.UDPConn, timeout time.Duration, handle_packet func(addr net.Addr, packet []byte, err error)) error {

    buf_chan    := make(chan []byte)
    err_chan    := make(chan error)

    var recv_addr net.Addr

    go func() {
        var buf []byte
        var err error
        recv_addr, buf, err = udp_recv_message(conn)
        if err != nil {
            err_chan <- err
        }
        buf_chan <- buf
    }()

    select {
    case <- time.After(timeout):
        return fmt.Errorf("UDP Receive Timeout")
    case err := <- err_chan:
        handle_packet(nil, nil, err)
    case packet := <- buf_chan:
        handle_packet(recv_addr, packet, nil)
    }

    return nil
}


var next_message_id uint16 = 65535

func UDPGetNextMsgID() uint16 {
    next_message_id++
    if next_message_id == 0 {
        next_message_id++
    }
    return next_message_id
}

func UDPRequestResponse(remote_addr *net.UDPAddr, request []byte, retry int, timeout time.Duration) ([]byte, error) {

    _, conn, err := UDPListenRandomPort()
    if err != nil {
        return nil, fmt.Errorf("Error: %s", err)
    }
    defer conn.Close()

    msg_id := UDPGetNextMsgID()

    var last_err error

    for i:=0; i<retry; i++ {

        //sAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(DEFAULT_CLUSTER_PORT))
        //if err != nil {
        //    last_err = fmt.Errorf("Error: %s", err)
        //    continue
        //}

        pudp_request := new(PUDP)
        // fields pudp_request
        pudp_request.Version        = 1
        pudp_request.PHL            = 1
        pudp_request.Algo_EC        = 0
        pudp_request.Algo_SIG       = 1
        pudp_request.Algo_ENC       = 0
        pudp_request.Reserved_1     = 0
        pudp_request.MSG_ID         = msg_id
        // timestamp
        pudp_request.Timestamp      = time.Now().UnixNano()
        pudp_request.Data           = request
        // marshall to buf
        buf, err := pudp_request.Marshal()
        if err != nil {
            last_err = fmt.Errorf("Error: %s", err)
            continue
        }

        if err = UDPSend(conn, remote_addr, buf); err != nil {
            last_err = fmt.Errorf("Error: %s", err)
            continue
        }

        var response        []byte
        var response_err    error
        if err = UDPReceiveMessage(conn, time.Duration(int(timeout.Seconds()) * (i+1)) * time.Second, func(addr net.Addr, packet []byte, err error) {
            if err != nil {
                response_err = err
                return
            }

            pudp_response, err   := New_PUDP(packet)
            if err != nil {
                response_err = err
                return
            }

            response        = pudp_response.Data
            response_err    = nil            
        }); err != nil {
            last_err = fmt.Errorf("Error: %s", err)
            continue
        }

        if response_err != nil {
            last_err = response_err
            continue
        }

        return response, nil
    }

    return nil, last_err
}

