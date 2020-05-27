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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

var murmurTestCases = []struct {
	input IKey
	seed  MurmurSeed
	want  uint32
}{
	{NewStringKey(""), 0, 0},
	{NewStringKey(""), 1, 0x514e28b7},
	{NewStringKey(""), 0xffffffff, 0x81f16f39},
	{NewStringKey("\xff\xff\xff\xff"), 0, 0x76293b50},
	{NewStringKey("!Ce\x87"), 0, 0xf55b516b},
	{NewStringKey("!Ce\x87"), 0x5082edee, 0x2362f9de},
	{NewStringKey("!Ce"), 0, 0x7e4a8634},
	{NewStringKey("!C"), 0, 0xa0f7b07a},
	{NewStringKey("!"), 0, 0x72661cf4},
	{NewStringKey("\x00\x00\x00\x00"), 0, 0x2362f9de},
	{NewStringKey("\x00\x00\x00"), 0, 0x85f0b427},
	{NewStringKey("\x00\x00"), 0, 0x30f4c306},
	{NewStringKey("Hello, world!"), 0x9747b28c, 0x24884CBA},
	{NewStringKey("ππππππππ"), 0x9747b28c, 0xD58063C1},
	{NewStringKey("abc"), 0, 0xb3dd93fa},
	{NewStringKey("abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq"), 0, 0xee925b90},
	{NewStringKey("The quick brown fox jumps over the lazy dog"), 0x9747b28c, 0x2fa826cd},
	{NewStringKey(strings.Repeat("a", 256)), 0x9747b28c, 0x37405bdc},
}

func TestMurmur(t *testing.T) {
	for _, tt := range murmurTestCases {
		got := tt.input.HashUint32(tt.seed.hash)
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
	testTable(t, []IKey{NewStringKey("foo"), NewStringKey("foo2"), NewStringKey("bar"), NewStringKey("baz")}, []IKey{NewStringKey("quux")})
}

func TestBuild_stress(t *testing.T) {
	var keys, extra []IKey
	rand_start := RandUint32() % 1000000
	rand_range := RandUint32()%50000 + 10000
	for i := int(rand_start); i < int(rand_start+rand_range); i++ {
		s := strconv.Itoa(i)
		if i < int(rand_start)+int(rand_range)/2 {
			keys = append(keys, NewStringKey(s))
		} else {
			extra = append(extra, NewStringKey(s))
		}
	}
	testTable(t, keys, extra)
}

func testTable(t *testing.T, keys []IKey, extra []IKey) {
	table := MPHBuild(keys, true)
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

func (t *MPHTable) Print() {
	fmt.Printf("MPHTable\n")
	fmt.Printf("    %v\n", t.level0)
	fmt.Printf("    %d\n", t.level0Mask)
	fmt.Printf("    %v\n", t.level1)
	fmt.Printf("    %d\n", t.level1Mask)
	if t.verifyKey != nil {
		for i := 0; i < len(t.verifyKey); i++ {
			fmt.Printf("        %#v\n", t.verifyKey[i])
		}
	} else {
		fmt.Printf("        %d\n", t.verifySeed)
		fmt.Printf("        %#v\n", t.verifyHash)
	}
}

func TestSerializeKey(t *testing.T) {
	var keys []IKey
	rand_start := RandUint32() % 1000000
	rand_range := RandUint32()%50000 + 10000
	for i := int(rand_start); i < int(rand_start+rand_range); i++ {
		s := strconv.Itoa(i)
		keys = append(keys, NewStringKey(s))
	}

	// build table
	table := MPHBuild(keys, true)

	// serialize, then deserialize
	buf, err := table.Encode()
	if err != nil {
		t.Errorf("Parse MPHTable encode failed %s", err)
		return
	}
	loaded_table, _, err := NewMPHTable(buf)
	if err != nil {
		t.Errorf("Parse MPHTable failed %s", err)
		return
	}

	// test level 0
	if !EqualUint32Array(table.level0, loaded_table.level0) {
		t.Errorf("Level 0 data mismatch")
	}
	if table.level0Mask != loaded_table.level0Mask {
		t.Errorf("Level 0 mask mismatch")
	}

	// test level 1
	if !EqualUint32Array(table.level1, loaded_table.level1) {
		//table.Print()
		//loaded_table.Print()
		t.Errorf("Level 1 data mismatch")
	}
	if table.level1Mask != loaded_table.level1Mask {
		t.Errorf("Level 1 mask mismatch")
	}

	// check verify key are not null
	if table.verifyKey == nil {
		t.Errorf("Unexpected nil verify key in orig table")
	}
	if loaded_table.verifyKey == nil {
		t.Errorf("Unexpected nil verify key in loaded table")
	}

	// test keys
	if len(table.verifyKey) != len(loaded_table.verifyKey) {
		t.Errorf("Verify Key length mismatch")
	}
	for i := 0; i < len(table.verifyKey); i++ {
		if !table.verifyKey[i].Equal(loaded_table.verifyKey[i]) {
			t.Errorf("Verify Key [%d] mismatch", i)
		}
	}

	// check verify hash are null
	if table.verifyHash != nil {
		t.Errorf("Unexpected not nil verify hash in orig table")
	}
	if loaded_table.verifyHash != nil {
		t.Errorf("Unexpected not nil verify hash in loaded table")
	}
}

func TestSerializeHash(t *testing.T) {
	var keys []IKey
	rand_start := RandUint32() % 1000000
	rand_range := RandUint32()%50000 + 10000
	for i := int(rand_start); i < int(rand_start+rand_range); i++ {
		s := strconv.Itoa(i)
		keys = append(keys, NewStringKey(s))
	}

	// build table
	table := MPHBuild(keys, false)

	// serialize, then deserialize
	buf, err := table.Encode()
	if err != nil {
		t.Errorf("Parse MPHTable encode failed %s", err)
	}
	loaded_table, _, err := NewMPHTable(buf)
	if err != nil {
		t.Errorf("Parse MPHTable failed %s", err)
	}

	// test level 0
	if !EqualUint32Array(table.level0, loaded_table.level0) {
		t.Errorf("Level 0 data mismatch")
	}
	if table.level0Mask != loaded_table.level0Mask {
		t.Errorf("Level 0 mask mismatch")
	}

	// test level 1
	if !EqualUint32Array(table.level1, loaded_table.level1) {
		t.Errorf("Level 1 data mismatch")
	}
	if table.level1Mask != loaded_table.level1Mask {
		t.Errorf("Level 1 mask mismatch")
	}

	// check verify key are null
	if table.verifyKey != nil {
		t.Errorf("Unexpected not nil verify key in orig table")
	}
	if loaded_table.verifyKey != nil {
		t.Errorf("Unexpected not nil verify key in loaded table")
	}

	// test verify hash
	if table.verifySeed != loaded_table.verifySeed {
		t.Errorf("Verify Seed mismatch")
	}
	if !EqualUint32Array(table.verifyHash, loaded_table.verifyHash) {
		t.Errorf("Verify Hash mismatch")
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
		MPHBuild(words, false)
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
		if !word.IsEncoded() {
			err := word.Encode(nil)
			if err != nil {
				b.Errorf("encode failed %s", word)
			}
		}
		m[string(word.SubKeyAt(0))] = uint32(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := i % len(words)
		n, ok := m[string(words[j].SubKeyAt(0))]
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
		benchTable = MPHBuild(words, false)
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
		words = append(words, NewStringKey(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return words, nil
}
