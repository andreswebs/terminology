package markdown

import (
	"strings"
	"testing"
)

func collect(src []byte) []Span {
	var spans []Span
	for s := range Spans(src) {
		spans = append(spans, s)
	}
	return spans
}

func TestHeadingsIncluded(t *testing.T) {
	t.Parallel()

	src := []byte("# Chapter One\n\nSome text.\n")
	spans := collect(src)

	var combined strings.Builder
	for _, s := range spans {
		combined.WriteString(s.Text)
	}
	if combined.String() != "Chapter OneSome text." {
		t.Errorf("combined text = %q, want %q", combined.String(), "Chapter OneSome text.")
	}
}

func TestIndentedCodeBlockExcluded(t *testing.T) {
	t.Parallel()

	src := []byte("text before\n\n    code line\n\ntext after\n")
	spans := collect(src)

	for _, s := range spans {
		if s.Text == "code line" {
			t.Errorf("span contains indented code block content: %q", s.Text)
		}
	}

	var combined strings.Builder
	for _, s := range spans {
		combined.WriteString(s.Text)
	}
	if combined.String() != "text beforetext after" {
		t.Errorf("combined text = %q, want %q", combined.String(), "text beforetext after")
	}
}

func TestEmptyInput(t *testing.T) {
	t.Parallel()

	spans := collect([]byte(""))
	if len(spans) != 0 {
		t.Errorf("expected no spans, got %d", len(spans))
	}

	spans = collect(nil)
	if len(spans) != 0 {
		t.Errorf("expected no spans for nil input, got %d", len(spans))
	}
}

func TestEmphasisAndLinksPreserved(t *testing.T) {
	t.Parallel()

	src := []byte("this is **bold** and [a link](http://example.com)")
	spans := collect(src)

	var combined strings.Builder
	for _, s := range spans {
		combined.WriteString(s.Text)
	}
	if combined.String() != "this is bold and a link" {
		t.Errorf("combined text = %q, want %q", combined.String(), "this is bold and a link")
	}

	// "bold" should have offset pointing into the original source
	for _, s := range spans {
		if s.Text == "bold" {
			wantOffset := 10 // "this is **" = 10 bytes
			if s.Offset != wantOffset {
				t.Errorf("'bold' Offset = %d, want %d", s.Offset, wantOffset)
			}
		}
	}
}

func TestHTMLBlockExcluded(t *testing.T) {
	t.Parallel()

	src := []byte("before\n\n<div>\nhidden content\n</div>\n\nafter\n")
	spans := collect(src)

	for _, s := range spans {
		if s.Text == "hidden content" {
			t.Errorf("span contains HTML block content: %q", s.Text)
		}
	}

	var combined strings.Builder
	for _, s := range spans {
		combined.WriteString(s.Text)
	}
	if combined.String() != "beforeafter" {
		t.Errorf("combined text = %q, want %q", combined.String(), "beforeafter")
	}
}

func TestSoftLineBreaksMerged(t *testing.T) {
	t.Parallel()

	src := []byte("hello\nworld\n")
	spans := collect(src)

	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1 (soft line break merged)", len(spans))
	}
	if spans[0].Text != "hello\nworld" {
		t.Errorf("Text = %q, want %q", spans[0].Text, "hello\nworld")
	}
	if spans[0].Line != 1 {
		t.Errorf("Line = %d, want 1", spans[0].Line)
	}
	if spans[0].Col != 1 {
		t.Errorf("Col = %d, want 1", spans[0].Col)
	}
	if spans[0].Offset != 0 {
		t.Errorf("Offset = %d, want 0", spans[0].Offset)
	}
}

func TestSoftLineBreaksMultiLine(t *testing.T) {
	t.Parallel()

	src := []byte("line one\nline two\nline three\n")
	spans := collect(src)

	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1 (consecutive soft breaks merged)", len(spans))
	}
	if spans[0].Text != "line one\nline two\nline three" {
		t.Errorf("Text = %q, want %q", spans[0].Text, "line one\nline two\nline three")
	}
}

func TestHardLineBreakNotMerged(t *testing.T) {
	t.Parallel()

	// Two trailing spaces = hard line break in markdown
	src := []byte("hello  \nworld\n")
	spans := collect(src)

	if len(spans) < 2 {
		t.Fatalf("got %d spans, want at least 2 (hard line break should not merge)", len(spans))
	}
}

func TestMultipleParagraphs(t *testing.T) {
	t.Parallel()

	src := []byte("first paragraph\n\nsecond paragraph\n")
	spans := collect(src)

	var combined strings.Builder
	for _, s := range spans {
		combined.WriteString(s.Text)
	}
	if combined.String() != "first paragraphsecond paragraph" {
		t.Errorf("combined text = %q, want %q", combined.String(), "first paragraphsecond paragraph")
	}

	// The second paragraph starts at byte offset 17
	var foundSecond bool
	for _, s := range spans {
		if s.Text == "second paragraph" {
			foundSecond = true
			if s.Offset != 17 {
				t.Errorf("second paragraph Offset = %d, want 17", s.Offset)
			}
			if s.Line != 3 {
				t.Errorf("second paragraph Line = %d, want 3", s.Line)
			}
			if s.Col != 1 {
				t.Errorf("second paragraph Col = %d, want 1", s.Col)
			}
		}
	}
	if !foundSecond {
		t.Error("did not find span with text \"second paragraph\"")
	}
}

func TestInlineCodeExcluded(t *testing.T) {
	t.Parallel()

	src := []byte("use the `getUserById` function")
	spans := collect(src)

	for _, s := range spans {
		if s.Text == "getUserById" {
			t.Errorf("span contains inline code content: %q", s.Text)
		}
	}

	var combined strings.Builder
	for _, s := range spans {
		combined.WriteString(s.Text)
	}
	if combined.String() != "use the  function" {
		t.Errorf("combined text = %q, want %q", combined.String(), "use the  function")
	}
}

func TestFencedCodeBlockExcluded(t *testing.T) {
	t.Parallel()

	src := []byte("before\n\n```go\nfunc main() {}\n```\n\nafter\n")
	spans := collect(src)

	for _, s := range spans {
		if s.Text == "func main() {}" {
			t.Errorf("span contains code block content: %q", s.Text)
		}
	}

	var combined strings.Builder
	for _, s := range spans {
		combined.WriteString(s.Text)
	}
	if combined.String() != "beforeafter" {
		t.Errorf("combined text = %q, want %q", combined.String(), "beforeafter")
	}
}

func TestPlainText(t *testing.T) {
	t.Parallel()

	spans := collect([]byte("hello world"))

	if len(spans) == 0 {
		t.Fatal("expected at least one span, got none")
	}

	var combined strings.Builder
	for _, s := range spans {
		combined.WriteString(s.Text)
	}
	if combined.String() != "hello world" {
		t.Errorf("combined text = %q, want %q", combined.String(), "hello world")
	}

	if spans[0].Offset != 0 {
		t.Errorf("first span Offset = %d, want 0", spans[0].Offset)
	}
	if spans[0].Line != 1 {
		t.Errorf("first span Line = %d, want 1", spans[0].Line)
	}
	if spans[0].Col != 1 {
		t.Errorf("first span Col = %d, want 1", spans[0].Col)
	}
}
