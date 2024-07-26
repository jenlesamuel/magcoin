package share

import (
	"bytes"
	"encoding/binary"
)

func Uint32ToByte4(n uint32) [4]byte {
	b := make([]byte, 4)

	binary.BigEndian.PutUint32(b, n)

	dst := [4]byte{}
	copy(dst[:], b)

	return dst
}

func Int32ToByte32(n int32) [32]byte {
	b := make([]byte, 32)

	binary.BigEndian.PutUint32(b, uint32(n))

	dst := [32]byte{}
	copy(dst[:], b)

	return dst
}

func IntToByte1(n int) [1]byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(n))

	dst := [1]byte{}
	copy(dst[:], b)

	return dst
}

func Int64ToByte8(n int64) [8]byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))

	dst := [8]byte{}
	copy(dst[:], b)

	return dst
}

func SliceToByte32(slice []byte) [32]byte {
	b := [32]byte{}

	copy(b[:], slice)

	return b
}

func SliceToByte4(slice []byte) [4]byte {
	b := [4]byte{}

	copy(b[:], slice)

	return b
}

func SliceToByte20(slice []byte) [20]byte {
	b := [20]byte{}

	copy(b[:], slice)

	return b
}

func IsZeroArray(b [32]byte) bool {
	b2 := [32]byte{}

	return bytes.Equal(b[:], b2[:])
}

func SliceToInt32(b []byte) (uint32, error) {
	buff := bytes.NewReader(b)
	var num uint32

	err := binary.Read(buff, binary.BigEndian, &num)
	if err != nil {
		return num, err
	}

	return num, nil
}

func Byte8ToInt64(b [8]byte) (int64, error) {
	reader := bytes.NewReader(b[:])
	var num int64

	if err := binary.Read(reader, binary.BigEndian, &num); err != nil {
		return num, err
	}

	return num, nil
}
