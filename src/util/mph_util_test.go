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
	"bufio"
	"os"
	"strconv"
	"sync"
	"testing"
)

var murmurTestCases = []struct {
	input []byte
	seed  MurmurSeed
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
			var seed MurmurSeed
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				seed.hash(s)
			}
		})
	}
}

func TestBuild_simple(t *testing.T) {
	testTable(t, [][]byte{[]byte("foo"), []byte("foo2"), []byte("bar"), []byte("baz")}, [][]byte{[]byte("quux")})
}

func TestBuild_stress(t *testing.T) {
	var keys, extra [][]byte
	for i := 0; i < 20000; i++ {
		s := strconv.Itoa(i)
		if i < 10000 {
			keys = append(keys, []byte(s))
		} else {
			extra = append(extra, []byte(s))
		}
	}
	testTable(t, keys, extra)
}

func testTable(t *testing.T, keys [][]byte, extra [][]byte) {
	table := MPHBuild(keys, 99)
	for i, key := range keys {
		n, ok := table.Lookup(key)
		if !ok {
			t.Errorf("Lookup(%s): got !ok; want ok", key)
			continue
		}
		if int(n) != i {
			t.Errorf("Lookup(%s): got n=%d; want %d", key, n, i)
		}
	}
	for _, key := range extra {
		if _, ok := table.Lookup(key); ok {
			t.Errorf("Lookup(%s): got ok; want !ok", key)
		}
	}
}

var (
	words      [][]byte
	wordsOnce  sync.Once
	benchTable *MPHTable
)

func BenchmarkMPHBuild(b *testing.B) {
	wordsOnce.Do(loadBenchTable)
	if len(words) == 0 {
		b.Skip("unable to load dictionary file")
	}
	for i := 0; i < b.N; i++ {
		MPHBuild(words, uint32(i))
	}
}

func BenchmarkMPHTable(b *testing.B) {
	wordsOnce.Do(loadBenchTable)
	if len(words) == 0 {
		b.Skip("unable to load dictionary file")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := i % len(words)
		n, ok := benchTable.Lookup(words[j])
		if !ok {
			b.Fatal("missing key")
		}
		if n != uint32(j) {
			b.Fatal("bad result index")
		}
	}
}

// For comparison against BenchmarkTable.
func BenchmarkMPHTableMap(b *testing.B) {
	wordsOnce.Do(loadBenchTable)
	if len(words) == 0 {
		b.Skip("unable to load dictionary file")
	}
	m := make(map[string]uint32)
	for i, word := range words {
		m[string(word)] = uint32(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := i % len(words)
		n, ok := m[string(words[j])]
		if !ok {
			b.Fatal("missing key")
		}
		if n != uint32(j) {
			b.Fatal("bad result index")
		}
	}
}

func loadBenchTable() {
	for _, dict := range []string{"/usr/share/dict/words", "/usr/dict/words"} {
		var err error
		words, err = loadDict(dict)
		if err == nil {
			break
		}
	}
	if len(words) > 0 {
		benchTable = MPHBuild(words, 0)
	}
}

func loadDict(dict string) ([][]byte, error) {
	f, err := os.Open(dict)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var words [][]byte
	for scanner.Scan() {
		words = append(words, []byte(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return words, nil
}