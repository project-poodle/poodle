package util

import (
	"testing"

	"../collection"
)

var keyTestCases = []struct {
	input    IKey
	wantData []byte
}{
	{NewEmptyKey(), []byte{0x00}},
	{µ(NewMappedKey(nil))[0].(IKey), []byte{0x00}},
	{NewKey(), []byte{0x00}},
	{NewKey().Add([]byte{'1'}), []byte{0x02, 0x01, '1'}},
	{NewKey().Add([]byte("a")), []byte{0x02, 0x01, 'a'}},
	{NewKey().Add([]byte("a")).Add([]byte("bc")), []byte{0x05, 0x01, 'a', 0x02, 'b', 'c'}},
}

func TestKey(t *testing.T) {
	for _, tt := range keyTestCases {
		err := tt.input.Encode(nil)
		if err != nil {
			t.Errorf("error occurred: %s", err)
		}
		gotData := tt.input.Buf()
		if !collection.EqualByteSlice(gotData, tt.wantData) {
			t.Errorf("(%v): got %v; want %v",
				tt.input, gotData, tt.wantData)
		}
	}
}

func generateRandomKey(sub, length int) IKey {
	size := int(RandUint32Range(0, uint32(sub)))
	key := NewKey()
	for i := 0; i < size; i++ {
		subKey := make([]byte, int(RandUint32Range(1, uint32(length))))
		//fmt.Printf("%d\n", len(subKey))
		for i := range subKey {
			subKey[i] = RandUint8()
		}
		key.Add(subKey)
	}
	return key
}

func TestKeyRandom(t *testing.T) {
	randStart := RandUint32() % 1000000
	randRange := RandUint32()%500 + 100
	for i := int(randStart); i < int(randStart+randRange); i++ {
		d := generateRandomKey(10, 320)
		err := d.Encode(nil)
		if err != nil {
			t.Errorf("error occurred: %s", err)
			//fmt.Printf("    %#v\n", d)
			continue
		}
		mapped, _, err := NewMappedKey(d.Buf())
		if err != nil {
			t.Errorf("error occurred: %s", err)
			//fmt.Printf("    %#v\n", d)
			//fmt.Printf("    %#v\n", buf)
			continue
		}
		// fmt.Printf("    %#v\n", mapped)
		if !testKeyEqual(d, mapped, t) {
			t.Errorf("key not match: \n%s, \n%s", d.ToString(), mapped.ToString())
		}
	}
}

func testKeyEqual(k1, k2 IKey, t *testing.T) bool {
	if !k1.IsDecoded() {
		_, err := k1.Decode(nil)
		if err != nil {
			t.Errorf("cannot decode k1 - %s", err)
		}
	}

	if !k2.IsDecoded() {
		_, err := k2.Decode(nil)
		if err != nil {
			t.Errorf("cannot decode k2 - %s", err)
		}
	}

	if k1.IsEmpty() {
		return k2.IsEmpty()
	}

	return k1.Equal(k2)
}
