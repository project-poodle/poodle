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
	{NewPrimitive(nil), true, nil, byte(0x00)},
	{NewPrimitive(nil), false, nil, byte(0x00)},
	{NewPrimitive([]byte("")), true, []byte{}, byte(0x00)},
	{NewPrimitive([]byte("")), false, []byte{0x00}, byte(0xff)},
	{NewPrimitive([]byte("a")), true, []byte{0x01, 'a'}, byte(0x01)},
	{NewPrimitive([]byte("abc")), false, []byte{0x01, 0x03, 'a', 'b', 'c'}, byte(0xff)},
	{NewDataArray().Append(NewPrimitive([]byte("abc"))), false, []byte{0x01<<6 | 0x01, 0x01, 0x05, 0x01, 0x03, 'a', 'b', 'c'}, byte(0xff)},
	{NewRecordList().Append(NewRecord().SetK([]byte("abc"))), false, []byte{0x01<<4 | 0x01, 0x01, 0x05, 0x01<<6, 0x03, 'a', 'b', 'c'}, byte(0xff)},
	{NewRecordList().Append(NewRecord().SetK([]byte("ab")).SetV([]byte("cd")).SetS([]byte("ef"))), false, []byte{0x01<<4 | 0x01, 0x01, 0x0a, (0x01<<6)|(0x01<<4)|(0x01<<2), 0x02, 'a', 'b', 0x02, 'c', 'd', 0x02, 'e', 'f'}, byte(0xff)},
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


