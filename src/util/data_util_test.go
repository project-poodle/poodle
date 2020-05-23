package util

import (
	"fmt"
	"testing"
	//"strconv"
)

var dataTestCases = []struct {
	input    IData
	parent   bool
	wantData []byte
	wantCode byte
}{
	{NewPrimitive(nil), true, nil, byte(0x00)},
	{NewPrimitive(nil), false, nil, byte(0x00)},
	{NewPrimitive([]byte("")), true, []byte{}, byte(0x00)},
	{NewPrimitive([]byte("")), false, []byte{0x00}, byte(0xff)},
	{NewPrimitive([]byte("a")), true, []byte{0x01, 'a'}, byte(0x01)},
	{NewPrimitive([]byte("abc")), false, []byte{0x01, 0x03, 'a', 'b', 'c'}, byte(0xff)},
	{NewDataArray().Append(NewPrimitive([]byte("abc"))), false, []byte{0x01<<6 | 0x01, 0x01, 0x05, 0x01, 0x03, 'a', 'b', 'c'}, byte(0xff)},
	{NewRecordList().Append(NewRecord().SetK([]byte("abc"))), false, []byte{0x01<<4 | 0x01, 0x01, 0x05, 0x01 << 6, 0x03, 'a', 'b', 'c'}, byte(0xff)},
	{NewRecordList().Append(NewRecord().SetK([]byte("ab")).SetV([]byte("cd")).SetS([]byte("ef"))), false, []byte{0x01<<4 | 0x01, 0x01, 0x0a, (0x01 << 6) | (0x01 << 4) | (0x01 << 2), 0x02, 'a', 'b', 0x02, 'c', 'd', 0x02, 'e', 'f'}, byte(0xff)},
}

func TestData(t *testing.T) {
	for _, tt := range dataTestCases {
		gotData, gotCode, err := tt.input.Encode(tt.parent)
		if err != nil {
			t.Errorf("error occurred: %s", err)
		}
		if !EqByteArray(gotData, tt.wantData) || gotCode != tt.wantCode {
			t.Errorf("(%v, parent=%t): got %v (%x); want %v (%x)",
				tt.input, tt.parent, gotData, gotCode, tt.wantData, tt.wantCode)
		}
	}
}

func generateRandomPrimitive(length int) IData {
	data := make([]byte, int(RandUint32Range(0, uint32(length))))
	for i := range data {
		data[i] = RandUint8()
	}
	return NewPrimitive(data)
}

func generateRandomData(depth, breadth, length int) IData {
	if depth < 0 {
		return nil
	} else if depth == 0 {
		return generateRandomPrimitive(length)
	}

	switch RandUint32Range(1, 3) {
	case 1:
		return generateRandomPrimitive(length)
	case 2:
		result := NewDataArray()
		size := RandUint16Range(0, uint16(breadth))
		for i := 0; i < int(size); i++ {
			result.Append(generateRandomData(depth-int(RandUint32Range(1, 2)), breadth, length))
		}
		return result
	case 3:
		result := NewRecordList()
		size := RandUint16Range(0, uint16(breadth))
		for i := 0; i < int(size); i++ {
			result.Append(generateRandomRecord(depth-int(RandUint32Range(1, 2)), breadth, length))
		}
		return result
	default:
		panic(fmt.Sprintf("unexpected switch"))
	}
}

func generateRandomRecord(depth, breadth, length int) IRecord {
	result := NewRecord()
	result.SetKey(generateRandomPrimitive(length))
	result.SetValue(generateRandomData(depth-1, breadth, length))
	result.SetScheme(generateRandomPrimitive(length))
	return result
}

func TestDataRandom(t *testing.T) {
	randStart := RandUint32() % 1000000
	randRange := RandUint32()%2000 + 500
	for i := int(randStart); i < int(randStart+randRange); i++ {
		d := generateRandomData(2, 15, 300)
		buf, _, err := d.Encode(false)
		if err != nil {
			t.Errorf("error occurred: %s", err)
			//fmt.Printf("    %#v\n", d)
			continue
		}
		mapped, err := NewStandardMappedData(buf)
		if err != nil {
			t.Errorf("error occurred: %s", err)
			//fmt.Printf("    %#v\n", d)
			//fmt.Printf("    %#v\n", buf)
			continue
		}
		// fmt.Printf("    %#v\n", mapped)
		if !testDataEqual(d, mapped, t) {
			t.Errorf("data not match: %#v, %#v", d, mapped)
		}
	}
}

func testDataEqual(d1, d2 IData, t *testing.T) bool {
	if !d1.IsDecoded() {
		err := d1.Decode(0xff)
		if err != nil {
			t.Errorf("cannot decode d1 - %s", err)
		}
	}

	if !d2.IsDecoded() {
		err := d2.Decode(0xff)
		if err != nil {
			t.Errorf("cannot decode d2 - %s", err)
		}
	}

	if d1.DataMagic() != d2.DataMagic() {
		t.Errorf("data magic not match %x vs %x", d1, d2)
		return false
	}

	if d1.IsNil() {
		return d2.IsNil()
	}

	if d1.IsPrimitive() {
		return d2.IsPrimitive() && EqByteArray(d1.Data(), d2.Data())
	}

	if d1.IsDataArray() {
		if d1.Size() != d2.Size() {
			t.Errorf("data array size mismatch: %d vs %d", d1, d2)
		}

		for i := uint16(0); i < d1.Size(); i++ {
			d1i, err := d1.DataAt(i)
			if err != nil {
				t.Errorf("cannot decode d1 [%d]", i)
			}
			d2i, err := d2.DataAt(i)
			if err != nil {
				t.Errorf("cannot decode d2 [%d]", i)
			}
			if !testDataEqual(d1i, d2i, t) {
				t.Errorf("data mismatch at idx %d", i)
			}
		}
	}

	if d1.IsRecordList() {
		if d1.Size() != d2.Size() {
			t.Errorf("record list size mismatch: %d vs %d", d1, d2)
		}
		for i := uint16(0); i < d1.Size(); i++ {
			r1i, err := d1.RecordAt(i)
			if err != nil {
				t.Errorf("cannot decode d1 [%d]", i)
			}
			r2i, err := d2.RecordAt(i)
			if err != nil {
				t.Errorf("cannot decode d2 [%d]", i)
			}
			if !testRecordEqual(r1i, r2i, t) {
				t.Errorf("data mismatch at idx %d", i)
			}
		}
	}

	return true
}

func testRecordEqual(r1, r2 IRecord, t *testing.T) bool {
	if !r1.IsDecoded() {
		err := r1.Decode()
		if err != nil {
			t.Errorf("cannot decode r1 - %s", err)
		}
	}

	if !r2.IsDecoded() {
		err := r2.Decode()
		if err != nil {
			t.Errorf("cannot decode r2 - %s", err)
		}
	}

	if !testDataEqual(r1.Key(), r2.Key(), t) {
		t.Errorf("r1 r2 key mismatch")
	}

	if !testDataEqual(r1.Value(), r2.Value(), t) {
		t.Errorf("r1 r2 value mismatch")
	}

	if !testDataEqual(r1.Scheme(), r2.Scheme(), t) {
		t.Errorf("r1 r2 scheme mismatch")
	}

	if (r1.Timestamp() == nil) != (r2.Timestamp() == nil) {
		t.Errorf("r1 r2 timestamp mismatch")
	}

	if r1.Timestamp() != nil {
		if r1.Timestamp().UnixNano() != r2.Timestamp().UnixNano() {
			t.Errorf("r1 r2 timestamp mismatch")
		}
	}

	return true
}
