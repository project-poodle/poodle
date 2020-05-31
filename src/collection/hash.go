// This Murmur3 hash implementation is originated from the original code
// @ https://github.com/cespare/mph
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

package collection

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"unsafe"
)

const (
	DEFAULT_HASH_MIN_CAPACITY     = 10
	DEFAULT_HASH_INITIAL_CAPACITY = 10
	DEFAULT_HASH_LOAD_FACTOR      = 0.75
	DEFAULT_HASH_MULTIPLIER       = 1.5
)

type Hash struct {
	seed       MurmurSeed
	loadFactor float32
	multiplier float32
	array      []*HashNode
	size       int
}

type HashNode struct {
	key   IHashable
	value IObject
	link  *HashNode
}

type HashIterator struct {
	hash     *Hash
	pos      int
	currNode *HashNode
}

func NewHash() *Hash {
	return NewCustomHash(
		DEFAULT_HASH_INITIAL_CAPACITY,
		DEFAULT_HASH_LOAD_FACTOR,
		DEFAULT_HASH_MULTIPLIER)
}

func NewCustomHash(iniitalCapacity int, loadFactor, multiplier float32) *Hash {

	// validity check
	if iniitalCapacity < DEFAULT_HASH_MIN_CAPACITY {
		iniitalCapacity = DEFAULT_HASH_MIN_CAPACITY
	}

	// validity check
	if multiplier > 5.0 {
		multiplier = 5.0
	} else if multiplier < 1.2 {
		multiplier = 1.2
	}

	// validity check
	if loadFactor > 1.0 {
		loadFactor = 1.0
	} else if loadFactor < 0.5 {
		loadFactor = 0.5
	}

	// compose Hash
	return &Hash{
		seed:       MurmurSeed(randUint32()),
		loadFactor: loadFactor,
		multiplier: multiplier,
		array:      make([]*HashNode, iniitalCapacity)}
}

func (h *Hash) Get(k IHashable) *HashNode {

	// compute hash value
	hashValue := k.HashUint32(h.seed.Hash) % uint32(len(h.array))

	// iterate linked list
	for currNode := h.array[hashValue]; currNode != nil; currNode = currNode.link {
		if currNode.key.Equal(k) {
			return currNode
		}
	}

	// if not found
	return nil
}

func (h *Hash) Put(k IHashable, v IObject) *HashNode {

	// compute hash value
	hashValue := k.HashUint32(h.seed.Hash) % uint32(len(h.array))

	// iterate linked list
	for currNode := h.array[hashValue]; currNode != nil; {
		if currNode.key.Equal(k) {
			// return previous key/value
			returnNode := &HashNode{key: currNode.key, value: currNode.value}
			currNode.key = k
			currNode.value = v
			return returnNode
		}
		currNode = currNode.link
	}

	// if not found, create new HashNode and insert to the head
	prevHead := h.array[hashValue]
	h.array[hashValue] = &HashNode{key: k, value: v, link: prevHead}
	h.size += 1

	// if size to capacity ratio is greater than loadFactor
	if float32(h.size)/float32(len(h.array)) >= h.loadFactor {
		newCapacity := int(float32(len(h.array)) * h.multiplier)
		h.resize(newCapacity)
	}

	return nil
}

func (h *Hash) Remove(k IHashable) *HashNode {

	// compute hash value
	hashValue := k.HashUint32(h.seed.Hash) % uint32(len(h.array))

	// iterate linked list
	prevNode := (*HashNode)(nil)
	currNode := h.array[hashValue]
	for currNode != nil {
		if currNode.key.Equal(k) {
			// return previous key/value
			if prevNode == nil {
				h.array[hashValue] = currNode.link // clear currNode
			} else {
				prevNode.link = currNode.link // skip currNode
				currNode.link = nil
			}
			h.size -= 1

			// if size to capacity ratio is less than loadFactor
			if float32(h.size)/float32(len(h.array)) < h.loadFactor/float32(math.Pow(float64(h.multiplier), 4)) {
				newCapacity := int(float32(len(h.array)) / float32(math.Pow(float64(h.multiplier), 2)))
				h.resize(newCapacity)
			}

			return currNode
		}

		prevNode = currNode
		currNode = currNode.link
	}

	// if not found
	return nil
}

func (h *Hash) Size() int {
	return h.size
}

func (h *Hash) resize(newCapacity int) {

	if newCapacity < DEFAULT_HASH_MIN_CAPACITY {
		newCapacity = DEFAULT_HASH_MIN_CAPACITY
	}

	// same capacity, no change needed
	if len(h.array) == newCapacity {
		return
	}

	newArray := make([]*HashNode, newCapacity)
	for _, node := range h.array {
		currNode := node
		for currNode != nil {
			hashValue := currNode.key.HashUint32(h.seed.Hash) % uint32(len(newArray))
			prevNode := newArray[hashValue]
			newArray[hashValue] = &HashNode{key: currNode.key, value: currNode.value, link: prevNode}
			currNode = currNode.link
		}
	}

	// we have complect resize successfully, just
	h.array = newArray
}

func (i *HashIterator) Next() IObject {
	i.advance()
	if i.pos >= len(i.hash.array) || i.currNode == nil {
		return nil
	}
	resultNode := &HashNode{key: i.currNode.key, value: i.currNode.value}
	i.currNode = i.currNode.link
	return resultNode
}

func (i *HashIterator) HasNext() bool {
	i.advance()
	return i.pos < len(i.hash.array) && i.currNode != nil
}

func (i *HashIterator) advance() {
	// advance to the next non-nil node
	for i.pos < len(i.hash.array) {
		if i.currNode != nil {
			return
		}
		i.pos += 1
	}
	if i.pos < len(i.hash.array) {
		i.currNode = i.hash.array[i.pos]
	}
}

func (h *Hash) Iterator() IIterator {
	return &HashIterator{hash: h, pos: 0, currNode: nil}
}

func (h *Hash) Print(w io.Writer, indent int) {
	fmt.Fprintf(w, "%"+strconv.Itoa(indent)+"s%s\n", "", h.ToString())
	for _, hn := range h.array {
		if !IsNil(hn) {
			hn.Print(w, indent+4)
		}
	}
}

func (h *Hash) ToString() string {
	return fmt.Sprintf("Hash: size=%d, capacity=%d, seed=%d, lf=%.2f, m=%.2f, ",
		h.size,
		len(h.array),
		h.seed,
		h.loadFactor,
		h.multiplier)
}

func (h *HashNode) Print(w io.Writer, indent int) {
	fmt.Fprintf(w, "%"+strconv.Itoa(indent)+"s%s\n", "", h.ToString())
	if !IsNil(h.link) {
		h.link.Print(w, indent+4)
	}
}

func (h *HashNode) ToString() string {
	if IsNil(h.link) {
		return fmt.Sprintf("k=%v, v=%v, l=[%v]", h.key, h.value, nil)
	} else {
		return fmt.Sprintf("k=%v, v=%v, l=[%v]", h.key, h.value, h.link.key)
	}
}

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
func (ms MurmurSeed) Hash(b []byte) uint32 {
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
