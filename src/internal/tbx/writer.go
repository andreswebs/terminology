package tbx

import "io"

// Writer encodes a glossary to a TBX output stream.
type Writer interface {
	Encode(w io.Writer, g *Glossary) error
}
