package tbx

import "io"

type Reader interface {
	Decode(r io.Reader) (*Glossary, []Warning, error)
}
