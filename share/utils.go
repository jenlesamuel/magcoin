package share

import (
	"bytes"
	"encoding/binary"
)

func IntToBytes(n int) []byte {
	b := make([]byte, 4)

	binary.BigEndian.PutUint32(b, uint32(n))

	return b
}

func IntToBytes32(n int) []byte {
	b := make([]byte, 32)

	binary.BigEndian.PutUint32(b, uint32(n))

	return b
}

func Int32ToByte32(n int32) [32]byte {
	b := make([]byte, 32)

	binary.BigEndian.PutUint32(b, uint32(n))

	dst := [32]byte{}
	copy(dst[:], b)

	return dst
}

func Int64ToBytes(n int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))

	return b
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

func BytesSliceToByte20(slice []byte) [20]byte {
	b := [20]byte{}

	copy(b[:], slice)

	return b
}

func BytesToInt(b []byte) (int, error) {
	buff := bytes.NewReader(b)
	var num uint

	err := binary.Read(buff, binary.BigEndian, &num)
	if err != nil {
		return int(num), err
	}

	return int(num), nil
}

func BytesToInt64(b []byte) (int64, error) {
	buff := bytes.NewReader(b)
	var num uint64

	err := binary.Read(buff, binary.BigEndian, &num)
	if err != nil {
		return int64(num), err
	}

	return int64(num), nil
}

func Byte8ToInt64(b [8]byte) (int64, error) {
	reader := bytes.NewReader(b[:])
	var num int64

	if err := binary.Read(reader, binary.BigEndian, &num); err != nil {
		return num, err
	}

	return num, nil
}
