package write

import (
	"bytes"
	"regexp"

	"github.com/andreswebs/terminology/internal/tbx"
)

var transacGrpRe = regexp.MustCompile(`(?s) *<transacGrp>.*?</transacGrp>\n`)

// ConceptsEqual reports whether two concepts are semantically equal, comparing
// their canonical TBX encodings while ignoring transaction records.
func ConceptsEqual(a, b *tbx.Concept) (bool, error) {
	canonA, err := canonicalize(a)
	if err != nil {
		return false, err
	}
	canonB, err := canonicalize(b)
	if err != nil {
		return false, err
	}
	return bytes.Equal(canonA, canonB), nil
}

func canonicalize(c *tbx.Concept) ([]byte, error) {
	g := &tbx.Glossary{
		Dialect:    tbx.DialectLinguist,
		SourceDesc: "equality-check",
		Concepts:   []tbx.Concept{*c},
	}

	w, err := tbx.WriterForDialect(tbx.DialectLinguist)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := w.Encode(&buf, g); err != nil {
		return nil, err
	}

	stripped := transacGrpRe.ReplaceAll(buf.Bytes(), nil)
	return stripped, nil
}
