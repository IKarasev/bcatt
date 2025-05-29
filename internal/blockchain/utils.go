package blockchain

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/rs/xid"
)

func reverseSlice[T any](d []T) {
	for i, j := 0, len(d)-1; i < j; i, j = i+1, j-1 {
		d[i], d[j] = d[j], d[i]
	}
}

func trimLeadingZeros(b []byte) []byte {
	i := 0
	for ; i < len(b) && b[i] == 0; i++ {
	}
	return b[i:]
}

func GenUniqueIdString() string {
	return xid.New().String()
}

func SliceHasDuplicates[T comparable](a []T) bool {
	tmp := make(map[T]interface{})
	for _, v := range a {
		if _, ok := tmp[v]; ok {
			return true
		}
		tmp[v] = nil
	}
	return false
}

func UintToBytes(a uint) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(a))
	return b
}

func IntToBytes[T int | int64](a T) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(a))
	return b
}

func BytesToString(b []byte) string {
	return hex.EncodeToString(b)
}

func StringToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

func IntToBytesString[T int | int64](i T) string {
	return hex.EncodeToString(IntToBytes(i))
}
