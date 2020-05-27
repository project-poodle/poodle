// This code is a re-implementation of the original code @ https://github.com/cespare/mph.
//
// The original code implements of key is string, this implementation is []byte (slice).
//
// This implementation also changes lookup verification to a hash value, compared to
// the original implementation of exact key.  In this implementation, the Lookup return
// of 'ok' does not mean exact match, rather an bloom filter.
//
// The original LICENSE.txt as below:
//
// Copyright (c) 2016 Caleb Spare
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package util

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
	"unsafe"
)

// A Table is an immutable hash table that provides constant-time lookups of key
// indices using a minimal perfect hash.
type MPHTable struct {
	level0     []uint32 // power of 2 size
	level0Mask int      // len(Level0) - 1
	level1     []uint32 // power of 2 size >= len(keys)
	level1Mask int      // len(Level1) - 1
	verifyKey  []IKey   // verify key - if not nil, verify lookup by exact key
	verifySeed uint32   // if verify key is nil, this is murmur seed for verify hash
	verifyHash []uint32 // if verify key is nil, this is verify lookup by hash (bloom filter)
}

// parse MPHTable from []byte (deserialize)
func NewMPHTable(buf []byte) (*MPHTable, int, error) {

	if len(buf) < 1 {
		return nil, 0, fmt.Errorf("NewMPHTable - no magic")
	}

	if buf[0] != 0 && buf[0] != 1 {
		return nil, 0, fmt.Errorf("NewMPHTable - unrecognized magic %d", buf[0])
	}

	// initialize
	t := &MPHTable{}
	pos := 1

	////////////////////////////////////////
	// level 0 size
	if len(buf) < pos+4 {
		return nil, pos, fmt.Errorf("NewMPHTable - missing level 0 length %d", len(buf))
	}
	level0_len := binary.BigEndian.Uint32(buf[pos:])
	pos += 4

	// parse each level 0 data
	if len(buf) < pos+4*int(level0_len) {
		return nil, pos, fmt.Errorf("NewMPHTable - missing level 0 data %d", level0_len)
	}
	t.level0 = make([]uint32, level0_len)
	for i := 0; i < int(level0_len); i++ {
		t.level0[i] = binary.BigEndian.Uint32(buf[pos:])
		pos += 4
	}

	// parse level 0 mask
	if len(buf) < pos+4 {
		return nil, pos, fmt.Errorf("NewMPHTable - missing level 0 mask %d", len(buf))
	}
	t.level0Mask = int(binary.BigEndian.Uint32(buf[pos:]))
	pos += 4

	////////////////////////////////////////
	// level 1 size
	if len(buf) < pos+4 {
		return nil, pos, fmt.Errorf("NewMPHTable - missing level 1 length %d", len(buf))
	}
	level1_len := binary.BigEndian.Uint32(buf[pos:])
	pos += 4

	// parse each level 0 data
	if len(buf) < pos+4*int(level1_len) {
		return nil, pos, fmt.Errorf("NewMPHTable - missing level 1 data %d", level1_len)
	}
	t.level1 = make([]uint32, level1_len)
	for i := 0; i < int(level1_len); i++ {
		t.level1[i] = binary.BigEndian.Uint32(buf[pos:])
		pos += 4
	}

	// parse level 0 mask
	if len(buf) < pos+4 {
		return nil, pos, fmt.Errorf("NewMPHTable - missing level 1 mask %d", len(buf))
	}
	t.level1Mask = int(binary.BigEndian.Uint32(buf[pos:]))
	pos += 4

	////////////////////////////////////////
	// verify key, or verify hash
	if buf[0] == 1 {

		// verify by key
		t.verifyHash = nil

		// verify key array size
		if len(buf) < pos+4 {
			return nil, pos, fmt.Errorf("NewMPHTable - missing verify key size %d", len(buf))
		}
		verifyKeySize := binary.BigEndian.Uint32(buf[pos:])
		pos += 4

		t.verifyKey = make([]IKey, verifyKeySize)

		// iterate each key
		for i := 0; i < int(verifyKeySize); i++ {

			// verify key length
			verifyKey, verifyKeyN, err := NewMappedKey(buf[pos:])
			if err != nil {
				return nil, pos, fmt.Errorf("NewMPHTable - load key failed %s", err)
			}

			// verify key data content
			if len(buf) < pos+verifyKeyN {
				return nil, pos, fmt.Errorf("NewMPHTable - missing verify key data [%d] %d", i, len(buf))
			}

			t.verifyKey[i] = verifyKey
			pos += verifyKeyN
		}

	} else {

		// verify by hash
		t.verifyKey = nil

		// verify hash seed
		if len(buf) < pos+4 {
			return nil, pos, fmt.Errorf("NewMPHTable - missing verify seed %d", len(buf))
		}
		t.verifySeed = binary.BigEndian.Uint32(buf[pos:])
		pos += 4

		// verify hash array size
		if len(buf) < pos+4 {
			return nil, pos, fmt.Errorf("NewMPHTable - missing verify hash length %d", len(buf))
		}
		verify_hash_len := binary.BigEndian.Uint32(buf[pos:])
		pos += 4

		// verify hash array data
		if len(buf) < pos+4*int(verify_hash_len) {
			return nil, pos, fmt.Errorf("NewMPHTable - missing verify hash data %d", verify_hash_len)
		}
		t.verifyHash = make([]uint32, verify_hash_len)

		// iterate each verify hash
		for i := 0; i < int(verify_hash_len); i++ {
			t.verifyHash[i] = binary.BigEndian.Uint32(buf[pos:])
			pos += 4
		}
	}

	return t, pos, nil
}

// serialize to []byte
func (t *MPHTable) Encode() ([]byte, error) {

	buf_len := 1 + 4 + 4*len(t.level0) + 4 + 4 + 4*len(t.level1) + 4
	buf := make([]byte, buf_len)
	pos := 0
	if t.verifyKey != nil {
		buf[0] = 1
		pos += 1
	} else {
		buf[0] = 0
		pos += 1
	}

	// level 0
	binary.BigEndian.PutUint32(buf[pos:], uint32(len(t.level0)))
	pos += 4
	for i := 0; i < len(t.level0); i++ {
		binary.BigEndian.PutUint32(buf[pos:], t.level0[i])
		pos += 4
	}
	binary.BigEndian.PutUint32(buf[pos:], uint32(t.level0Mask))
	pos += 4

	// level 1
	binary.BigEndian.PutUint32(buf[pos:], uint32(len(t.level1)))
	pos += 4
	for i := 0; i < len(t.level1); i++ {
		binary.BigEndian.PutUint32(buf[pos:], t.level1[i])
		pos += 4
	}
	binary.BigEndian.PutUint32(buf[pos:], uint32(t.level1Mask))
	pos += 4

	// verify key, or verify hash
	if t.verifyKey != nil {

		// key array size
		key_len_buf := make([]byte, 4)
		binary.BigEndian.PutUint32(key_len_buf, uint32(len(t.verifyKey)))
		buf = append(buf, key_len_buf...)

		// iterate each key
		for i := 0; i < len(t.verifyKey); i++ {

			key_data := t.verifyKey[i]
			if !key_data.IsEncoded() {
				err := key_data.Encode(nil)
				if err != nil {
					return nil, fmt.Errorf("MPHTable::Encode - %s", err)
				}
			}

			// append
			buf = append(buf, key_data.Buf()...)
		}
	} else {

		verify_hash_buf := make([]byte, 4+4+4*len(t.verifyHash))

		// verify seed
		binary.BigEndian.PutUint32(verify_hash_buf[0:], t.verifySeed)

		// verify hash array size
		binary.BigEndian.PutUint32(verify_hash_buf[4:], uint32(len(t.verifyHash)))

		// iterate each verify hash
		for i := 0; i < len(t.verifyHash); i++ {
			binary.BigEndian.PutUint32(verify_hash_buf[8+i*4:], t.verifyHash[i])
		}

		// append
		buf = append(buf, verify_hash_buf...)
	}

	return buf, nil
}

// Build builds a Table from keys using the "Hash, displace, and compress"
// algorithm described in http://cmph.sourceforge.net/papers/esa09.pdf.
func MPHBuild(keys []IKey, verify_by_key bool) *MPHTable {
	var (
		level0        = make([]uint32, nextPow2(len(keys)/4))
		level0Mask    = len(level0) - 1
		level1        = make([]uint32, nextPow2(len(keys)))
		level1Mask    = len(level1) - 1
		sparseBuckets = make([][]int, len(level0))
		zeroSeed      = MurmurSeed(0)
		keyArray      = make([]IKey, len(keys))
		verifySeed    = uint32(RandUint64Range(2^16, 2^32-1))
	)
	for i, s := range keys {
		keyArray[i] = s
		n := int(keyArray[i].HashUint32(zeroSeed.hash)) & level0Mask
		sparseBuckets[n] = append(sparseBuckets[n], i)
	}
	var buckets []indexBucket
	for n, vals := range sparseBuckets {
		if len(vals) > 0 {
			buckets = append(buckets, indexBucket{n, vals})
		}
	}
	sort.Sort(bySize(buckets))

	occ := make([]bool, len(level1))
	var tmpOcc []int
	for _, bucket := range buckets {
		var seed MurmurSeed
	trySeed:
		tmpOcc = tmpOcc[:0]
		for _, i := range bucket.vals {
			n := int(keyArray[i].HashUint32(seed.hash)) & level1Mask
			if occ[n] {
				for _, n := range tmpOcc {
					occ[n] = false
				}
				seed++
				goto trySeed
			}
			occ[n] = true
			tmpOcc = append(tmpOcc, n)
			level1[n] = uint32(i)
		}
		level0[int(bucket.n)] = uint32(seed)
	}

	if verify_by_key {
		// verify by exact key
		return &MPHTable{
			level0:     level0,
			level0Mask: level0Mask,
			level1:     level1,
			level1Mask: level1Mask,
			verifyKey:  keyArray,
		}
	} else {
		// verify by hash (bloom filter)
		verifyHash := make([]uint32, len(keyArray))
		for i := 0; i < len(keyArray); i++ {
			// verify by bloom filter key
			// bloom filter key needs to be consistent for a given key, and can be different from actual key bytes
			// different bloom key is a security measure, and is not required
			verifyHash[i] = keys[i].HashUint32((MurmurSeed)(verifySeed).hash)
		}

		return &MPHTable{
			level0:     level0,
			level0Mask: level0Mask,
			level1:     level1,
			level1Mask: level1Mask,
			verifySeed: verifySeed,
			verifyHash: verifyHash,
		}
	}
}

func nextPow2(n int) int {
	for i := 1; ; i *= 2 {
		if i >= n {
			return i
		}
	}
}

// Lookup searches for s in t and returns its index and whether it was found.
func (t *MPHTable) Lookup(s IKey) (n uint32, ok bool) {
	i0 := int(s.HashUint32(MurmurSeed(0).hash)) & t.level0Mask
	seed := t.level0[i0]
	i1 := int(s.HashUint32(MurmurSeed(seed).hash)) & t.level1Mask
	n = t.level1[i1]
	if t.verifyKey != nil {
		return n, s.Equal(t.verifyKey[int(n)])
	} else {
		// verify by bloom filter key
		// bloom filter key needs to be consistent for a given key, and can be different from actual key bytes
		// different bloom key is a security measure, and is not required
		verify_hash := s.HashUint32((MurmurSeed)(t.verifySeed).hash)
		return n, verify_hash == t.verifyHash[int(n)]
	}
}

type indexBucket struct {
	n    int
	vals []int
}

type bySize []indexBucket

func (s bySize) Len() int           { return len(s) }
func (s bySize) Less(i, j int) bool { return len(s[i].vals) > len(s[j].vals) }
func (s bySize) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// Below code contains an optimized murmur3 32-bit implementation tailored for
// our specific use case. See https://en.wikipedia.org/wiki/MurmurHash.

// A murmurSeed is the initial state of a Murmur3 hash.
type MurmurSeed uint32

const (
	c1      = 0xcc9e2d51
	c2      = 0x1b873593
	r1Left  = 15
	r1Right = 32 - r1Left
	r2Left  = 13
	r2Right = 32 - r2Left
	m       = 5
	n       = 0xe6546b64
)

// hash computes the 32-bit Murmur3 hash of s using ms as the seed.
func (ms MurmurSeed) hash(b []byte) uint32 {
	h := uint32(ms)
	l := len(b)
	numBlocks := l / 4
	var blocks []uint32
	header := (*reflect.SliceHeader)(unsafe.Pointer(&blocks))
	header.Data = (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data
	header.Len = numBlocks
	header.Cap = numBlocks
	for _, k := range blocks {
		k *= c1
		k = (k << r1Left) | (k >> r1Right)
		k *= c2
		h ^= k
		h = (h << r2Left) | (h >> r2Right)
		h = h*m + n
	}

	var k uint32
	ntail := l & 3
	itail := l - ntail
	switch ntail {
	case 3:
		k ^= uint32(b[itail+2]) << 16
		fallthrough
	case 2:
		k ^= uint32(b[itail+1]) << 8
		fallthrough
	case 1:
		k ^= uint32(b[itail])
		k *= c1
		k = (k << r1Left) | (k >> r1Right)
		k *= c2
		h ^= k
	}

	h ^= uint32(l)
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}
