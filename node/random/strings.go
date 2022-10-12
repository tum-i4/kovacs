package random

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

func String(length int) string {
	byteSlice := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, byteSlice)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(byteSlice)
}
