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

type TestKey struct {
    key     []byte
}

func NewTestKey(s string) (*TestKey) {
    return &TestKey{key: []byte(s)}
}

func (t *TestKey) Key() []byte {
    return t.key
}

func (t *TestKey) Equal(k IKey) bool {
    if tmp, ok := k.(*TestKey); ok {
        return TestEq(t.key, tmp.key)
    } else {
        return false
    }
}

var murmurTestCases = []struct {
	input IKey
	seed  MurmurSeed
	want  uint32
}{
	{NewTestKey(""), 0, 0},
	{NewTestKey(""), 1, 0x514e28b7},
	{NewTestKey(""), 0xffffffff, 0x81f16f39},
	{NewTestKey("\xff\xff\xff\xff"), 0, 0x76293b50},
	{NewTestKey("!Ce\x87"), 0, 0xf55b516b},
	{NewTestKey("!Ce\x87"), 0x5082edee, 0x2362f9de},
	{NewTestKey("!Ce"), 0, 0x7e4a8634},
	{NewTestKey("!C"), 0, 0xa0f7b07a},
	{NewTestKey("!"), 0, 0x72661cf4},
	{NewTestKey("\x00\x00\x00\x00"), 0, 0x2362f9de},
	{NewTestKey("\x00\x00\x00"), 0, 0x85f0b427},
	{NewTestKey("\x00\x00"), 0, 0x30f4c306},
	{NewTestKey("Hello, world!"), 0x9747b28c, 0x24884CBA},
	{NewTestKey("ππππππππ"), 0x9747b28c, 0xD58063C1},
	{NewTestKey("abc"), 0, 0xb3dd93fa},
	{NewTestKey("abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq"), 0, 0xee925b90},
	{NewTestKey("The quick brown fox jumps over the lazy dog"), 0x9747b28c, 0x2fa826cd},
	{NewTestKey(strings.Repeat("a", 256)), 0x9747b28c, 0x37405bdc},
}

func TestMurmur(t *testing.T) {
	for _, tt := range murmurTestCases {
		got := tt.seed.hash(tt.input.Key())
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
	testTable(t, []IKey{NewTestKey("foo"), NewTestKey("foo2"), NewTestKey("bar"), NewTestKey("baz")}, []IKey{NewTestKey("quux")})
}

func TestBuild_stress(t *testing.T) {
	var keys, extra []IKey
	for i := 0; i < 20000; i++ {
		s := strconv.Itoa(i)
		if i < 10000 {
			keys = append(keys, NewTestKey(s))
		} else {
			extra = append(extra, NewTestKey(s))
		}
	}
	testTable(t, keys, extra)
}

func testTable(t *testing.T, keys []IKey, extra []IKey) {
	table := MPHBuild(keys, 99, true)
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
	words      []IKey
	wordsOnce  sync.Once
	benchTable *MPHTable
)

func BenchmarkMPHBuild(b *testing.B) {
	wordsOnce.Do(loadBenchTable)
	if len(words) == 0 {
		b.Skip("unable to load dictionary file")
	}
	for i := 0; i < b.N; i++ {
		MPHBuild(words, uint32(i), false)
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
		m[string(word.Key())] = uint32(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := i % len(words)
		n, ok := m[string(words[j].Key())]
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
		benchTable = MPHBuild(words, 0, false)
	}
}

func loadDict(dict string) ([]IKey, error) {
	f, err := os.Open(dict)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var words []IKey
	for scanner.Scan() {
		words = append(words, NewTestKey(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return words, nil
}