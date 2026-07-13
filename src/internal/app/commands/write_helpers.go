package commands

import (
	"errors"
	"io/fs"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/terr"
	"github.com/andreswebs/terminology/internal/write"
)

func loadTBXForWrite(path string) (*tbx.Glossary, error) {
	g, _, err := tbx.Load(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, terr.Newf("io_error", 3, "", "%s", err)
		}
		return nil, err
	}
	return g, nil
}

func buildWriteResult(c tbx.Concept) output.WriteResult {
	return write.ConceptToWriteResult(c)
}
