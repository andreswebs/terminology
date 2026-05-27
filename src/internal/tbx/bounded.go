package tbx

import (
	"fmt"
	"io"
	"os"
)

const (
	MaxTBXSize      int64 = 50 << 20 // 50 MB
	MaxMarkdownSize int64 = 10 << 20 // 10 MB
	MaxStdinSize    int64 = 10 << 20 // 10 MB
	MaxPayloadSize  int64 = 10 << 20 // 10 MB
)

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

func ReadFileBounded(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return ReadBounded(f, limit)
}
