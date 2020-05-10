package util

import (
    //"os"
    //"flag"
    "fmt"
    //"log"
    "bytes"
    //"math"
    //"net"
    //"time"
    //"reflect"
    "strings"
    "encoding/binary"
)


//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////
// constants

var RADIX_VALID_BITS                    = [...]int{1, 2, 4, 8}              // valid radix bits
var RADIX_VALID_SPANS                   = [...]int{2, 4, 16, 256}           // valid radix spans
var RADIX_PRINT_INDENT                  = "    "


//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////
// utility functions

func is_valid_radix_bits(bits int) bool {
    for _, valid := range RADIX_VALID_BITS {
        if bits == valid {
            return true
        }
    }

    return false
}

func max_int(x, y int) int {
    if (x > y) {
        return x
    } else {
        return y
    }
}

func min_int(x, y int) int {
    if (x < y) {
        return x
    } else {
        return y
    }
}


//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////
// interfaces

type IRadixExtension interface {
    CanReplaceBy(data IRadixData)   bool                // whether to replace
    Print(indent int)                                   // print data
}

type IRadixData interface {
    Key()                           []byte              // return key
    Value()                         []byte              // return value
    Extension()                     IRadixExtension     // generic interface - extensions are not included in HASH or Merkle Tree computation
    Print(indent int)                                   // print data
}

type IRadixNode interface {

    RadixBits()                     int                 // radix bits

    Parent()                        IRadixNode          // ParentNode
    IsLeaf()                        bool                // whether this node is leaf

    PrefixLength()                  int                 // prefix length: # of bits
    PrefixBuf()                     []byte              // represent prefix buffer

    Hash()                          []byte              // SHA256 hash of the current block - 32 bytes
    Upsert(data IRadixData)         int                 // <0 : failure,  >=0 : successful,  >0 : changes were introduced, recompute of hash required

    Print(indent int)                                   // print data
}


//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////
// implementations

//////////////////////////////////////////////////////////////////////
// RadixData

type RadixData struct {             // impements IRadixData, IRadixNode
    parent          IRadixNode
    key_size        int
    buf             []byte
    ext             IRadixExtension
}

func NewRadixData(parent IRadixNode, data IRadixData) *RadixData {
    if parent == nil {
        panic(fmt.Sprintf("Parent cannot be nil"))
    }

    result          := new(RadixData)
    result.parent   = parent
    result.key_size = int(len(data.Key()))
    result.buf      = append(data.Key(), data.Value()...)
    result.ext      = data.Extension()
    return result
}

func (rd *RadixData) Key() []byte {
    return rd.buf[0:rd.key_size]
}

func (rd *RadixData) Value() []byte {
    return rd.buf[rd.key_size:]
}

func (rd *RadixData) Data() []byte {
    return rd.buf
}

func (rd *RadixData) Extension() IRadixExtension {
    return rd.ext
}

func (rd *RadixData) RadixBits() int {
    return rd.parent.RadixBits()
}

func (rd *RadixData) Parent() IRadixNode {
    return rd.parent
}

func (rd *RadixData) IsLeaf() bool {
    return true
}

func (rd *RadixData) PrefixLength() int {
    return rd.key_size
}

func (rd *RadixData) PrefixBuf() []byte {
    return rd.Key()
}

func (rd *RadixData) Hash() []byte {
    buf := SumSHA256(rd.buf)
    return buf[:]
}

func (rd *RadixData) Print(indent int) {
    padding := strings.Repeat(RADIX_PRINT_INDENT, int(indent))
    fmt.Printf(padding + "KEY           : %x\n", rd.Key())
    fmt.Printf(padding + "Value         : %x\n", rd.Value())
    fmt.Printf(padding + "Ext           : \n")
    if rd.ext != nil {
        rd.ext.Print(indent + 1)
    }
}

// <0 : failure,  >=0 : successful,  >0 : changes were introduced, recompute of hash required
func (rd *RadixData) Upsert(data IRadixData) int {
    if !bytes.Equal(rd.Key(), data.Key()) {
        return -1
    }

    if bytes.Equal(rd.Value(), data.Value()) {
        return 0
    }

    if rd.ext == nil || rd.ext.CanReplaceBy(data) {
        rd.buf  = append(data.Key(), data.Value()...)
        rd.ext  = data.Extension()
        return 1
    } else {
        return 0
    }
}

//////////////////////////////////////////////////////////////////////
// RadixNode

type RadixNode struct {
    parent          IRadixNode
    radix_bits      int
    radix_span      int
    buf             []byte              // 32*radix_span hash, 32 bytes hash for own data hash, 4 byte prefix_len, n bytes of prefix_buf
    prefix_len      int                 // prefix bits
    prefix_buf      []byte              // prefix buf
    data            *RadixData          // contains data only if key is the same as this prefix
    children        []IRadixNode        // children
}

func NewRadixRoot(radix_bits int) *RadixNode {
    if !is_valid_radix_bits(radix_bits) {
        panic(fmt.Sprintf("Invalid radix bit %d. valud bits are %v", radix_bits, RADIX_VALID_BITS))
    }

    result := new(RadixNode)
    result.parent               = nil
    result.radix_bits           = radix_bits
    result.radix_span           = int(1<<radix_bits)
    result.prefix_len           = 0
    result.prefix_buf           = []byte{}
    result.buf                  = make([]byte, 32*result.radix_span + 32 + 4)
    prefix_len_bs               := make([]byte, 4)
    binary.BigEndian.PutUint32(prefix_len_bs, uint32(result.prefix_len))
    copy(result.buf[32*result.radix_span + 32:], prefix_len_bs)
    result.children             = make([]IRadixNode, result.radix_span)
    return result
}

func NewRadixNode(parent IRadixNode, prefix_len int, prefix_buf []byte) *RadixNode {
    if parent == nil {
        panic(fmt.Sprintf("Parent cannot be nil"))
    }

    result := new(RadixNode)
    result.parent               = parent
    result.radix_bits           = parent.RadixBits()
    result.radix_span           = int(1<<result.radix_bits)
    result.prefix_len           = prefix_len
    prefix_len_bytes            := int((prefix_len + 7) / 8)
    result.prefix_buf           = make([]byte, prefix_len_bytes)
    copy(result.prefix_buf, prefix_buf[:min_int(len(prefix_buf), prefix_len_bytes)])
    result.buf                  = make([]byte, 32*int(result.radix_span) + 32 + 4 + len(result.prefix_buf))
    prefix_len_bs               := make([]byte, 4)
    binary.BigEndian.PutUint32(prefix_len_bs, uint32(result.prefix_len))
    copy(result.buf[32*result.radix_span + 32:], prefix_len_bs)
    copy(result.buf[32*result.radix_span + 32 + 4:], result.prefix_buf)
    result.children             = make([]IRadixNode, result.radix_span)
    return result
}

func (rn *RadixNode) Parent() IRadixNode {
    return rn.parent
}

func (rn *RadixNode) IsLeaf() bool {
    return false
}

func (rn *RadixNode) RadixBits() int {
    return rn.radix_bits
}

func (rn *RadixNode) PrefixLength() int {
    return rn.prefix_len
}

func (rn *RadixNode) PrefixBuf() []byte {
    return rn.prefix_buf
}

func (rn *RadixNode) Hash() []byte {
    buf := SumSHA256(rn.buf)
    return buf[:]
}

func (rn *RadixNode) Print(indent int) {
    padding := strings.Repeat(RADIX_PRINT_INDENT, int(indent))
    fmt.Printf(padding + strings.Repeat("=", 60) + "\n")
    fmt.Printf(padding + "Prefix    : %d [%x]\n", rn.PrefixLength(), rn.PrefixBuf())
    fmt.Printf(padding + "Data      :\n")
    if rn.data != nil {
        rn.data.Print(indent+1)
    }
    for i:=int(0); i<rn.radix_span; i++ {
        child := rn.children[i]
        if child != nil {
            fmt.Printf(padding + strings.Repeat("-", 60) + "\n")
            fmt.Printf(padding + "Hash[%x]      : %x\n", i, rn.buf[32*i:32*(i+1)])
            fmt.Printf(padding + "Child[%x]     :\n", i)
            child.Print(indent + 1)
        }
    }
}

func (rn *RadixNode) key_satisfy_prefix(key []byte) bool {
    if len(key) * 8 < int(rn.prefix_len) {
        return false
    }

    for i:=0; i<len(rn.prefix_buf); i++ {
        curr_bits_len := int(rn.prefix_len) - i*8
        if curr_bits_len >= 8 {
            if key[i] != rn.prefix_buf[i] {
                return false
            } else {
                continue
            }
        } else if curr_bits_len >= 1 {
            mask := byte(0xFF) << uint8(8 - curr_bits_len)
            if (key[i] & mask) != (rn.prefix_buf[i] & mask) {
                return false
            } else {
                return true
            }
        } else {
            break
        }
    }

    return true
}

func (rn *RadixNode) key_get_idx(key []byte) int {
    var result uint8 = 0
    for i:=0; i<len(key); i++ {
        curr_bits_len := int(rn.prefix_len) - i*8
        if curr_bits_len >= 8 {
            continue
        } else if curr_bits_len >= 1 {
            result = key[i] << curr_bits_len >> (8 - rn.radix_bits)
        } else if curr_bits_len + int(rn.radix_bits) >= 1 {
            result |= key[i] >> uint8(8 - (curr_bits_len + int(rn.radix_bits)))
        }
    }

    return int(result)
}

func (rn *RadixNode) key_make_prefix(idx int) (prefix_len int, prefix_buf []byte) {
    prefix_len              = rn.prefix_len + rn.radix_bits
    prefix_len_bytes        := int((prefix_len + 7) / 8)
    prefix_buf              = make([]byte, prefix_len_bytes)
    copy(prefix_buf, rn.prefix_buf)

    mask := (uint8)(0xFF) >> (8 - rn.radix_bits)
    prefix_bits := ((uint8)(idx) & mask) << (prefix_len % 8)
    prefix_buf[len(prefix_buf)-1] = prefix_buf[len(prefix_buf)-1] & ((uint8)(0xFF) << (prefix_len % 8)) | prefix_bits

    return prefix_len, prefix_buf
}

func (rn *RadixNode) Lookup(key []byte) IRadixData {
    // check for key validity
    if !rn.key_satisfy_prefix(key) {
        return nil
    }

    // if the key for this RadixNode
    if len(key) * 8 == int(rn.prefix_len) && bytes.Equal(key, rn.prefix_buf) {
        return rn.data
    }

    // we are here if we are to check children
    idx := rn.key_get_idx(key)

    child := rn.children[idx]
    if child == nil {
        return nil
    } else if child.IsLeaf() {
        if bytes.Equal((child.(*RadixData)).Key(), key) {
            return child.(*RadixData)
        } else {
            return nil
        }
    } else {
        return child.(*RadixNode).Lookup(key)
    }
}

// <0 : failure,  >=0 : successful,  >0 : changes were introduced, recompute of hash required
func (rn *RadixNode) Upsert(data IRadixData) int {
    // check for key validity
    if !rn.key_satisfy_prefix(data.Key()) {
        return -1
    }

    // if the key for this RadixNode
    if len(data.Key()) * 8 == int(rn.prefix_len) && bytes.Equal(data.Key(), rn.prefix_buf) {
        if rn.data == nil {
            rn.data = NewRadixData(rn, data)
            copy(rn.buf[32 * rn.radix_span:], rn.data.Hash())
            return 1
        } else {
            result := rn.data.Upsert(data)
            if result > 0 {
                copy(rn.buf[32 * rn.radix_span:], rn.data.Hash())
                return 1
            } else {
                return 0
            }
        }
    }

    // we are here if we are to check children
    idx := rn.key_get_idx(data.Key())

    child := rn.children[idx]
    if child == nil {
        // if child does not exist
        child := NewRadixData(rn, data)
        rn.children[idx] = child
        copy(rn.buf[32 * idx:], child.Hash())
        return 1
    } else if child.IsLeaf() {
        // if child is a leaf node
        child_rd := child.(*RadixData)
        if bytes.Equal(data.Key(), child_rd.Key()) {
            result := child_rd.Upsert(data)
            if result > 0 {
                copy(rn.buf[32 * idx:], child_rd.Hash())
            }
            return result
        }

        new_prefix_len, new_prefix_buf := rn.key_make_prefix(idx)

        new_rn          := NewRadixNode(rn, new_prefix_len, new_prefix_buf)
        result_child    := new_rn.Upsert(child_rd)
        result_data     := new_rn.Upsert(data)

        if result_child < 0 || result_data < 0 {
            return -1
        }

        rn.children[idx]    = new_rn
        copy(rn.buf[32 * idx:], new_rn.Hash())
        return 1

    } else {

        child_rn := child.(*RadixNode)

        result := child_rn.Upsert(data)
        if result > 0 {
            copy(rn.buf[32 * idx:], child_rn.Hash())
        }

        return result
    }
}

// return # of records deleted
func (rn *RadixNode) Delete(cond_func func(data IRadixData) bool) int {

    found_count := 0
    for i:=int(0); i<rn.radix_span; i++ {
        child := rn.children[i]
        if child == nil {
            continue
        } else if child.IsLeaf() {
            child_rd := child.(*RadixData)
            if cond_func(child_rd) {
                rn.children[i] = nil
                copy(rn.buf[32 * i:], make([]byte, 32))
                found_count += 1
            }
        } else {
            child_rn := child.(*RadixNode)
            found_count += child_rn.Delete(cond_func)
            if found_count != 0 {
                copy(rn.buf[32 * i:], child_rn.Hash())
            }
        }
    }

    return found_count
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////
