package tbx

import "io"

type Writer interface {
	Encode(w io.Writer, g *Glossary) error
}
