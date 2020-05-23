package util

import (
	//"fmt"
	"testing"
)

var dataTestCases = []struct {
	input       IData
	parent      bool
	want_data   []byte
	want_code   byte
}{
	{NewConstructedPrimitive(nil), true, nil, byte(0x00)},
	{NewConstructedPrimitive(nil), false, nil, byte(0x00)},
	{NewConstructedPrimitive([]byte("")), true, []byte{}, byte(0x00)},
	{NewConstructedPrimitive([]byte("")), false, []byte{0x00}, byte(0xff)},
	{NewConstructedPrimitive([]byte("a")), true, []byte{0x01, 'a'}, byte(0x01)},
	{NewConstructedPrimitive([]byte("abc")), false, []byte{0x01, 0x03, 'a', 'b', 'c'}, byte(0xff)},
	{NewConstructedDataArray().Append(NewConstructedPrimitive([]byte("abc"))), false, []byte{0x01<<6 | 0x01, 0x01, 0x05, 0x01, 0x03, 'a', 'b', 'c'}, byte(0xff)},
	{NewConstructedRecordList().Append(NewConstructedRecord().SetKeyData([]byte("abc"))), false, []byte{0x01<<4 | 0x01, 0x01, 0x05, 0x01<<6, 0x03, 'a', 'b', 'c'}, byte(0xff)},
}


func TestData(t *testing.T) {
	for _, tt := range dataTestCases {
	    got_data, got_code, err := tt.input.Encode(tt.parent)
	    if err != nil {
			t.Errorf("error occurred: %s", err)
	    }
		if !EqByteArray(got_data, tt.want_data) || got_code != tt.want_code {
			t.Errorf("(%v, parent=%t): got %v (%x); want %v (%x)",
				tt.input, tt.parent, got_data, got_code, tt.want_data, tt.want_code)
		}
	}
}


