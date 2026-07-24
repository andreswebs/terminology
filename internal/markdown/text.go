package markdown

import (
	"iter"
	"sort"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gmtext "github.com/yuin/goldmark/text"
)

// Span is a run of markdown text content together with its byte offset and
// line/column position in the source.
type Span struct {
	Text   string
	Offset int
	Line   int
	Col    int
}

// Spans returns an iterator over the plain-text spans of src, skipping code
// blocks, code spans, and raw HTML.
func Spans(src []byte) iter.Seq[Span] {
	reader := gmtext.NewReader(src)
	parser := goldmark.DefaultParser()
	doc := parser.Parse(reader)

	lineOffsets := buildLineOffsets(src)

	return func(yield func(Span) bool) {
		walkText(doc, src, lineOffsets, yield)
	}
}

func walkText(n ast.Node, src []byte, lineOffsets []int, yield func(Span) bool) bool {
	switch n.Kind() {
	case ast.KindFencedCodeBlock, ast.KindCodeBlock, ast.KindCodeSpan, ast.KindHTMLBlock:
		return true
	}

	if n.Kind() == ast.KindText {
		txt := n.(*ast.Text)
		seg := txt.Segment
		val := seg.Value(src)
		if len(val) > 0 {
			line, col := position(lineOffsets, seg.Start)
			if !yield(Span{
				Text:   string(val),
				Offset: seg.Start,
				Line:   line,
				Col:    col,
			}) {
				return false
			}
		}
		return true
	}

	child := n.FirstChild()
	for child != nil {
		if child.Kind() != ast.KindText {
			if !walkText(child, src, lineOffsets, yield) {
				return false
			}
			child = child.NextSibling()
			continue
		}

		first := child.(*ast.Text)
		last := first
		for last.SoftLineBreak() && last.NextSibling() != nil && last.NextSibling().Kind() == ast.KindText {
			last = last.NextSibling().(*ast.Text)
		}

		merged := src[first.Segment.Start:last.Segment.Stop]
		if len(merged) > 0 {
			line, col := position(lineOffsets, first.Segment.Start)
			if !yield(Span{
				Text:   string(merged),
				Offset: first.Segment.Start,
				Line:   line,
				Col:    col,
			}) {
				return false
			}
		}

		child = last.NextSibling()
	}
	return true
}

func buildLineOffsets(src []byte) []int {
	offsets := []int{0}
	for i := range src {
		if src[i] == '\n' {
			offsets = append(offsets, i+1)
		}
	}
	return offsets
}

func position(lineOffsets []int, offset int) (line, col int) {
	i := max(sort.SearchInts(lineOffsets, offset+1)-1, 0)
	return i + 1, offset - lineOffsets[i] + 1
}
