package tbx

import "io"

// Reader decodes a glossary from a TBX input stream, returning any non-fatal
// warnings.
type Reader interface {
	Decode(r io.Reader) (*Glossary, []Warning, error)
}
