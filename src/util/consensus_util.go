package util

import (
    "fmt"
    "encoding/binary"
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
// LedgerTime

type LedgerTime struct {
    epoch               uint32
}

func NewLedgerTime(t uint32) *LedgerTime {
    return &LedgerTime{epoch: t}
}

func (t *LedgerTime) Buf() []byte {
    buf := make([]byte, 4)
    binary.BigEndian.PutUint32(buf, t.epoch)
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

func (t *RaftTime) Buf() []byte {
    buf := make([]byte, 12)
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
    Buf                 []byte
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
    c.Buf = buf[:pos]

    return c, nil
}

func (c *ConsensusID) Copy() (*ConsensusID) {
    // make a deep copy of the buf
    buf := make([]byte, len(c.Buf))
    copy(buf, c.Buf)
    copy, err := NewConsensusID(buf)
    if err != nil {
        // this should not happen
        panic(fmt.Sprintf("ConsensusID:Copy - %s", err))
    }
    return copy
}