package lineindex

import (
	"strings"
	"testing"
)

func TestEmptyInput(t *testing.T) {
	t.Parallel()

	idx, err := New(strings.NewReader(""))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	line, col := idx.Position(0)
	if line != 1 || col != 1 {
		t.Errorf("Position(0) = (%d, %d), want (1, 1)", line, col)
	}
}

func TestSingleLine(t *testing.T) {
	t.Parallel()

	idx, err := New(strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	cases := []struct {
		offset   int
		wantLine int
		wantCol  int
	}{
		{0, 1, 1},
		{4, 1, 5},
	}

	for _, tc := range cases {
		line, col := idx.Position(tc.offset)
		if line != tc.wantLine || col != tc.wantCol {
			t.Errorf("Position(%d) = (%d, %d), want (%d, %d)",
				tc.offset, line, col, tc.wantLine, tc.wantCol)
		}
	}
}

func TestMultipleLines(t *testing.T) {
	t.Parallel()

	idx, err := New(strings.NewReader("line1\nline2\nline3"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	cases := []struct {
		offset   int
		wantLine int
		wantCol  int
	}{
		{0, 1, 1},
		{4, 1, 5},
		{5, 1, 6},
		{6, 2, 1},
		{11, 2, 6},
		{12, 3, 1},
		{16, 3, 5},
	}

	for _, tc := range cases {
		line, col := idx.Position(tc.offset)
		if line != tc.wantLine || col != tc.wantCol {
			t.Errorf("Position(%d) = (%d, %d), want (%d, %d)",
				tc.offset, line, col, tc.wantLine, tc.wantCol)
		}
	}
}

func TestOffsetPastEnd(t *testing.T) {
	t.Parallel()

	idx, err := New(strings.NewReader("ab\ncd"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	line, _ := idx.Position(999)
	if line != 2 {
		t.Errorf("Position(999) line = %d, want 2", line)
	}
}

func TestNegativeOffset(t *testing.T) {
	t.Parallel()

	idx, err := New(strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	line, col := idx.Position(-1)
	if line != 1 || col != 1 {
		t.Errorf("Position(-1) = (%d, %d), want (1, 1)", line, col)
	}
}

func TestTrailingNewline(t *testing.T) {
	t.Parallel()

	idx, err := New(strings.NewReader("abc\n"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	cases := []struct {
		offset   int
		wantLine int
		wantCol  int
	}{
		{0, 1, 1},
		{3, 1, 4},
		{4, 2, 1},
	}

	for _, tc := range cases {
		line, col := idx.Position(tc.offset)
		if line != tc.wantLine || col != tc.wantCol {
			t.Errorf("Position(%d) = (%d, %d), want (%d, %d)",
				tc.offset, line, col, tc.wantLine, tc.wantCol)
		}
	}
}

func TestLargeInput(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	lineLen := 80
	numLines := 10000
	for i := range numLines {
		for range lineLen - 1 {
			b.WriteByte('x')
		}
		if i < numLines-1 {
			b.WriteByte('\n')
		}
	}

	idx, err := New(strings.NewReader(b.String()))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	line, col := idx.Position(0)
	if line != 1 || col != 1 {
		t.Errorf("Position(0) = (%d, %d), want (1, 1)", line, col)
	}

	midOffset := 5000 * lineLen
	line, col = idx.Position(midOffset)
	if line != 5001 || col != 1 {
		t.Errorf("Position(%d) = (%d, %d), want (5001, 1)", midOffset, line, col)
	}

	lastLineStart := (numLines - 1) * lineLen
	line, col = idx.Position(lastLineStart)
	if line != numLines || col != 1 {
		t.Errorf("Position(%d) = (%d, %d), want (%d, 1)", lastLineStart, line, col, numLines)
	}
}

func TestWindowsLineEndings(t *testing.T) {
	t.Parallel()

	idx, err := New(strings.NewReader("ab\r\ncd"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	cases := []struct {
		offset   int
		wantLine int
		wantCol  int
	}{
		{0, 1, 1},
		{1, 1, 2},
		{2, 1, 3},
		{3, 1, 4},
		{4, 2, 1},
	}

	for _, tc := range cases {
		line, col := idx.Position(tc.offset)
		if line != tc.wantLine || col != tc.wantCol {
			t.Errorf("Position(%d) = (%d, %d), want (%d, %d)",
				tc.offset, line, col, tc.wantLine, tc.wantCol)
		}
	}
}
