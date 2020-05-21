// This code is a re-implementation of the original code @ https://github.com/cespare/mph.
//
// The original code implements string interface, while our implementation is []byte (slice).
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
	"reflect"
	"unsafe"
)


// This file contains an optimized murmur3 32-bit implementation tailored for
// our specific use case. See https://en.wikipedia.org/wiki/MurmurHash.

// A murmurSeed is the initial state of a Murmur3 hash.
type murmurSeed uint32

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
func (ms murmurSeed) hash(b []byte) uint32 {
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