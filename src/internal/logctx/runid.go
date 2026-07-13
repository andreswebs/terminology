package logctx

import (
	"crypto/rand"
	"encoding/hex"
)

// NewRunID returns a random 16-character hexadecimal identifier for a run.
func NewRunID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
