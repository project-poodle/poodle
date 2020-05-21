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
	"fmt"
	"strings"
	"testing"
)

var murmurTestCases = []struct {
	input []byte
	seed  murmurSeed
	want  uint32
}{
	{[]byte(""), 0, 0},
	{[]byte(""), 1, 0x514e28b7},
	{[]byte(""), 0xffffffff, 0x81f16f39},
	{[]byte("\xff\xff\xff\xff"), 0, 0x76293b50},
	{[]byte("!Ce\x87"), 0, 0xf55b516b},
	{[]byte("!Ce\x87"), 0x5082edee, 0x2362f9de},
	{[]byte("!Ce"), 0, 0x7e4a8634},
	{[]byte("!C"), 0, 0xa0f7b07a},
	{[]byte("!"), 0, 0x72661cf4},
	{[]byte("\x00\x00\x00\x00"), 0, 0x2362f9de},
	{[]byte("\x00\x00\x00"), 0, 0x85f0b427},
	{[]byte("\x00\x00"), 0, 0x30f4c306},
	{[]byte("Hello, world!"), 0x9747b28c, 0x24884CBA},
	{[]byte("ππππππππ"), 0x9747b28c, 0xD58063C1},
	{[]byte("abc"), 0, 0xb3dd93fa},
	{[]byte("abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq"), 0, 0xee925b90},
	{[]byte("The quick brown fox jumps over the lazy dog"), 0x9747b28c, 0x2fa826cd},
	{[]byte(strings.Repeat("a", 256)), 0x9747b28c, 0x37405bdc},
}

func TestMurmur(t *testing.T) {
	for _, tt := range murmurTestCases {
		got := tt.seed.hash(tt.input)
		if got != tt.want {
			t.Errorf("hash(%q, seed=0x%x): got 0x%x; want %x",
				tt.input, tt.seed, got, tt.want)
		}
	}
}

func BenchmarkMurmur(b *testing.B) {
	for _, size := range []int{1, 4, 8, 16, 32, 50, 500} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			s := []byte(strings.Repeat("a", size))
			b.SetBytes(int64(size))
			var seed murmurSeed
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				seed.hash(s)
			}
		})
	}
}