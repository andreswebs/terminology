package logctx

import (
	"crypto/rand"
	"encoding/hex"
)

func NewRunID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
