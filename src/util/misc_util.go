package util

import (
	"encoding/binary"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// utilities
////////////////////////////////////////////////////////////////////////////////

func µ(a ...interface{}) []interface{} {
	return a
}

func Ternary(statement bool, a, b interface{}) interface{} {
	if statement {
		return a
	}
	return b
}

////////////////////////////////////////////////////////////////////////////////
// Encode and Decode Varchar
////////////////////////////////////////////////////////////////////////////////

func EncodeUvarint(data uint64) []byte {

	lenBuf := make([]byte, 10) // maximum 10 bytes
	lenN := binary.PutUvarint(lenBuf, data)

	if lenN < 0 {
		panic(fmt.Sprintf("EncodeUvarint - invalid uvarint length [%d], input [%d]", lenN, data))
	} else if lenN == 0 && data != 0 {
		panic(fmt.Sprintf("EncodeUvarint - nvalid uvarint encode length [%d], input [%d]", lenN, data))
	}

	return lenBuf[:lenN]
}

func DecodeUvarint(buf []byte) (uint64, int, error) {

	// empty input
	if buf == nil || len(buf) == 0 {
		return 0, 0, nil
	}

	bufLength, bufN := binary.Uvarint(buf)
	if bufN < 0 {
		// sub key cannot have zero length
		return 0, 0, fmt.Errorf("DecodeUvarint - failed to read input length [%d]", bufN)
	} else if bufN == 0 && len(buf) != 0 && buf[0] != 0 {
		return 0, 0, fmt.Errorf("DecodeUvarint - unexpected error - buf len [%d], first byte [%d]", len(buf), buf[0])
	} else if len(buf) < bufN+int(bufLength) {
		return 0, 0, fmt.Errorf("DecodeUvarint - varchar length [%d] bigger than remaining buf size [%d]", bufLength, len(buf)-bufN)
	}

	return bufLength, bufN, nil
}

func EncodeVarchar(data []byte) []byte {

	lenBuf := make([]byte, 10) // maximum 10 bytes
	lenN := binary.PutUvarint(lenBuf, uint64(len(data)))
	if lenN < 0 {
		panic(fmt.Sprintf("EncodeVarchar - invalid uvarint length [%d], input [%d]", lenN, len(data)))
	} else if lenN == 0 && len(data) != 0 {
		panic(fmt.Sprintf("EncodeVarchar - nvalid uvarint encode length [%d], input [%d]", lenN, len(data)))
	}

	buf := make([]byte, lenN+len(data))
	copy(buf, lenBuf[:lenN])
	copy(buf[lenN:], data)

	return buf
}

func DecodeVarchar(buf []byte) ([]byte, int, error) {

	// empty input
	if buf == nil || len(buf) == 0 {
		return []byte{}, 0, nil
	}

	bufLength, bufN := binary.Uvarint(buf)
	if bufN < 0 {
		// sub key cannot have zero length
		return nil, 0, fmt.Errorf("DecodeVarchar - failed to read input length [%d]", bufN)
	} else if bufN == 0 && len(buf) != 0 && buf[0] != 0 {
		return nil, 0, fmt.Errorf("DecodeVarchar - unexpected error - buf len [%d], first byte [%d]", len(buf), buf[0])
	} else if len(buf) < bufN+int(bufLength) {
		return nil, 0, fmt.Errorf("DecodeVarchar - varchar length [%d] bigger than remaining buf size [%d]", bufLength, len(buf)-bufN)
	}

	return buf[bufN : bufN+int(bufLength)], bufN + int(bufLength), nil
}
