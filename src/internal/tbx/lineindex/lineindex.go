package lineindex

import (
	"io"
	"sort"
)

type Index struct {
	offsets []int
	size    int
}

func New(r io.Reader) (*Index, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var offsets []int
	offsets = append(offsets, 0)
	for i, b := range data {
		if b == '\n' {
			offsets = append(offsets, i+1)
		}
	}

	return &Index{offsets: offsets, size: len(data)}, nil
}

func (idx *Index) Position(offset int) (line, col int) {
	if offset < 0 {
		return 1, 1
	}

	i := max(sort.SearchInts(idx.offsets, offset+1)-1, 0)

	line = i + 1
	col = offset - idx.offsets[i] + 1
	return line, col
}
