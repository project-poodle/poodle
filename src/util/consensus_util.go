package util

import (
    "fmt"
    "encoding/binary"
)

const (
    CONSENSUS_TIME_LEDGER   = 0x01
    CONSENSUS_TIME_RAFT     = 0x02
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces

type IConsensusTime interface {
    Buf()                   []byte
    GT(ict IConsensusTime)  (bool, error)
    GE(ict IConsensusTime)  (bool, error)
    EQ(ict IConsensusTime)  (bool, error)
    NE(ict IConsensusTime)  (bool, error)
    LT(ict IConsensusTime)  (bool, error)
    LE(ict IConsensusTime)  (bool, error)
}

////////////////////////////////////////////////////////////////////////////////
// Factory

func NewConsensusTime(buf []byte) (IConsensusTime, error) {
    if len(buf) < 1 {
        return nil, fmt.Errorf("NewConsensusTime - no magic")
    }

    switch buf[0] {

    case CONSENSUS_TIME_LEDGER:

        if len(buf) < 1 + 4 {
            return nil, fmt.Errorf("NewConsensusTime - missing ledger time - %x", buf)
        }
        return NewLedgerTime(binary.BigEndian.Uint32(buf[1:])), nil

    case CONSENSUS_TIME_RAFT:

        if len(buf) < 1 + 12 {
            return nil, fmt.Errorf("NewConsensusTime - missing raft time - %x", buf)
        }
        return NewRaftTime(binary.BigEndian.Uint32(buf[1:]),
                            binary.BigEndian.Uint32(buf[5:]),
                            binary.BigEndian.Uint32(buf[9:])), nil

    default:

        return nil, fmt.Errorf("NewConsensusTime - unsupported magic - %x", buf[0])
    }
}

////////////////////////////////////////////////////////////////////////////////
// LedgerTime

type LedgerTime struct {
    epoch               uint32
}

func NewLedgerTime(t uint32) *LedgerTime {
    return &LedgerTime{epoch: t}
}

func (t *LedgerTime) Buf() []byte {
    buf     := make([]byte, 1 + 4)
    buf[0]  = CONSENSUS_TIME_LEDGER
    binary.BigEndian.PutUint32(buf[1:], t.epoch)
    return buf
}

func (t *LedgerTime) GT(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*LedgerTime); ok {
        return t.epoch > tmp.epoch, nil
    } else {
        return false, fmt.Errorf("LedgerTime::GT - incompatible type %x", ict)
    }
}

func (t *LedgerTime) GE(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*LedgerTime); ok {
        return t.epoch >= tmp.epoch, nil
    } else {
        return false, fmt.Errorf("LedgerTime::GE - incompatible type %x", ict)
    }
}

func (t *LedgerTime) EQ(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*LedgerTime); ok {
        return t.epoch == tmp.epoch, nil
    } else {
        return false, fmt.Errorf("LedgerTime::EQ - incompatible type %x", ict)
    }
}

func (t *LedgerTime) NE(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*LedgerTime); ok {
        return t.epoch != tmp.epoch, nil
    } else {
        return false, fmt.Errorf("LedgerTime::NE - incompatible type %x", ict)
    }
}

func (t *LedgerTime) LT(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*LedgerTime); ok {
        return t.epoch < tmp.epoch, nil
    } else {
        return false, fmt.Errorf("LedgerTime::LT - incompatible type %x", ict)
    }
}

func (t *LedgerTime) LE(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*LedgerTime); ok {
        return t.epoch <= tmp.epoch, nil
    } else {
        return false, fmt.Errorf("LedgerTime::LE - incompatible type %x", ict)
    }
}

////////////////////////////////////////////////////////////////////////////////
// RaftTime

type RaftTime struct {
    term                uint32
    millis              uint32
    count               uint32
}

func NewRaftTime(term, millis, count uint32) *RaftTime {
    return &RaftTime{term: term, millis: millis, count: count}
}

func (t *RaftTime) Buf() []byte {
    buf     := make([]byte, 1 + 12)
    buf[0]  = CONSENSUS_TIME_RAFT
    binary.BigEndian.PutUint32(buf, t.term)
    binary.BigEndian.PutUint32(buf[4:], t.millis)
    binary.BigEndian.PutUint32(buf[8:], t.count)
    return buf
}

func (t *RaftTime) GT(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*RaftTime); ok {
        if t.term > tmp.term {
            return true, nil
        } else if (t.term < tmp.term) {
            return false, nil
        }
        if t.millis > tmp.millis {
            return true, nil
        } else if (t.millis < tmp.millis) {
            return false, nil
        }
        if t.count > tmp.count {
            return true, nil
        } else if (t.count < tmp.count) {
            return false, nil
        }
        return false, nil
    } else {
        return false, fmt.Errorf("RaftTime::GT - incompatible type %x", ict)
    }
}

func (t *RaftTime) GE(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*RaftTime); ok {
        if t.term > tmp.term {
            return true, nil
        } else if (t.term < tmp.term) {
            return false, nil
        }
        if t.millis > tmp.millis {
            return true, nil
        } else if (t.millis < tmp.millis) {
            return false, nil
        }
        if t.count > tmp.count {
            return true, nil
        } else if (t.count < tmp.count) {
            return false, nil
        }
        return true, nil
    } else {
        return false, fmt.Errorf("RaftTime::GE - incompatible type %x", ict)
    }
}

func (t *RaftTime) EQ(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*RaftTime); ok {
        if t.term == tmp.term && t.millis == tmp.millis && t.count == tmp.count {
            return true, nil
        } else {
            return false, nil
        }
    } else {
        return false, fmt.Errorf("RaftTime::EQ - incompatible type %x", ict)
    }
}

func (t *RaftTime) NE(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*RaftTime); ok {
        if t.term != tmp.term || t.millis != tmp.millis || t.count != tmp.count {
            return true, nil
        } else {
            return false, nil
        }
    } else {
        return false, fmt.Errorf("RaftTime::NE - incompatible type %x", ict)
    }
}

func (t *RaftTime) LT(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*RaftTime); ok {
        if t.term < tmp.term {
            return true, nil
        } else if (t.term < tmp.term) {
            return false, nil
        }
        if t.millis < tmp.millis {
            return true, nil
        } else if (t.millis < tmp.millis) {
            return false, nil
        }
        if t.count < tmp.count {
            return true, nil
        } else if (t.count < tmp.count) {
            return false, nil
        }
        return false, nil
    } else {
        return false, fmt.Errorf("RaftTime::LT - incompatible type %x", ict)
    }
}

func (t *RaftTime) LE(ict IConsensusTime) (bool, error) {
    if tmp, ok := ict.(*RaftTime); ok {
        if t.term < tmp.term {
            return true, nil
        } else if (t.term < tmp.term) {
            return false, nil
        }
        if t.millis < tmp.millis {
            return true, nil
        } else if (t.millis < tmp.millis) {
            return false, nil
        }
        if t.count < tmp.count {
            return true, nil
        } else if (t.count < tmp.count) {
            return false, nil
        }
        return true, nil
    } else {
        return false, fmt.Errorf("RaftTime::LE - incompatible type %x", ict)
    }
}

////////////////////////////////////////////////////////////////////////////////
// ConsensusID

type ConsensusID struct {
    ConsensusMagic      byte
    UniverseID          []byte
    ClusterID           []byte
    FederationID        []byte
    ServiceID           []byte
    ShardStart          []byte
    ShardEnd            []byte
    buf                 []byte
}

func NewConsensusID(buf []byte) (*ConsensusID, error) {
    if buf == nil || len(buf) == 0 || buf[0] == 0x00 {
        return nil, fmt.Errorf("NewConsensusID - invalid magic [%b]", buf[0])
    }

    if (buf[0] & 0x7) != 0 {
        return nil, fmt.Errorf("NewConsensusID - invalid magic [%b] - reserved bits set", buf[0])
    }

    c := &ConsensusID{ConsensusMagic: buf[0]}

    pos := 1
    if (buf[0] >> 7) & 0x01 != 0 {
        if len(buf) < pos + 32 {
            return nil, fmt.Errorf("NewConsensusID - insufficient length [%x]", buf)
        }
        c.UniverseID  = buf[pos:pos+32]
        pos += 32
    }

    if (buf[0] >> 6) & 0x01 != 0 {
        if len(buf) < pos + 32 {
            return nil, fmt.Errorf("NewConsensusID - insufficient length [%x]", buf)
        }
        c.ClusterID  = buf[pos:pos+32]
        pos += 32
    }

    if (buf[0] >> 5) & 0x01 != 0 {
        if len(buf) < pos + 32 {
            return nil, fmt.Errorf("NewConsensusID - insufficient length [%x]", buf)
        }
        c.FederationID  = buf[pos:pos+32]
        pos += 32
    }

    if (buf[0] >> 4) & 0x01 != 0 {
        if len(buf) < pos + 32 {
            return nil, fmt.Errorf("NewConsensusID - insufficient length [%x]", buf)
        }
        c.ServiceID  = buf[pos:pos+32]
        pos += 32
    }

    if (buf[0] >> 3) & 0x01 != 0 {
        if len(buf) < pos + 64 {
            return nil, fmt.Errorf("NewConsensusID - insufficient length [%x]", buf)
        }
        c.ShardStart  = buf[pos:pos+32]
        pos += 32
        c.ShardEnd  = buf[pos:pos+32]
        pos += 32
    }

    // set buf length to be the exact length
    c.buf = buf[:pos]

    return c, nil
}

func (c *ConsensusID) Buf() ([]byte) {
    return c.buf
}

func (c *ConsensusID) Copy() (*ConsensusID) {
    // make a deep copy of the buf
    buf := make([]byte, len(c.buf))
    copy(buf, c.buf)
    copy, err := NewConsensusID(buf)
    if err != nil {
        // this should not happen
        panic(fmt.Sprintf("ConsensusID:Copy - %s", err))
    }
    return copy
}