package utils

import (
	"encoding/binary"
)

func Int64ToBytes(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func BytesToInt64(xs []byte) int64 {
	return int64(binary.BigEndian.Uint64(xs))
}
