// Package tbx models TBX glossaries and provides loading, saving, locking,
// validation, lookup, and dialect handling for terminology exchange files.
package tbx

import (
	"fmt"
	"io"
	"os"
)

// MaxTBXSize and the related limits cap the number of bytes read from the
// various inputs to guard against excessive memory use.
const (
	MaxTBXSize      int64 = 50 << 20 // 50 MB
	MaxMarkdownSize int64 = 10 << 20 // 10 MB
	MaxStdinSize    int64 = 10 << 20 // 10 MB
	MaxPayloadSize  int64 = 10 << 20 // 10 MB
)

// ReadBounded reads all bytes from r, failing with ErrInputTooLarge if the
// data exceeds limit.
func ReadBounded(r io.Reader, limit int64) ([]byte, error) {
	lr := io.LimitReader(r, limit+1)
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, ErrInputTooLarge.Wrap(
			fmt.Errorf("input size exceeds %d bytes", limit),
		)
	}
	return data, nil
}

// ReadFileBounded opens the file at path and reads its contents subject to
// limit, failing with ErrInputTooLarge if the file exceeds it.
func ReadFileBounded(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return ReadBounded(f, limit)
}
