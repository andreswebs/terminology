package tbx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Load reads the TBX file at path, detects its dialect, and decodes it into a
// Glossary, returning any non-fatal warnings encountered during decoding.
func Load(path string) (*Glossary, []Warning, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("opening TBX: %w", err)
	}
	defer func() { _ = f.Close() }()

	data, err := ReadBounded(f, MaxTBXSize)
	if err != nil {
		return nil, nil, err
	}

	if err := CheckDoctype(data); err != nil {
		return nil, nil, err
	}

	dialect, err := detectDialect(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}

	r, err := readerFor(dialect)
	if err != nil {
		return nil, nil, err
	}
	return r.Decode(bytes.NewReader(data))
}

// Save writes g to path atomically while holding the file lock.
func Save(path string, g *Glossary) error {
	lockPath := path + ".lock"
	unlock, err := acquireLock(lockPath)
	if err != nil {
		return err
	}
	defer unlock()

	return writeFile(path, g)
}

// SaveLocked writes g to path atomically, assuming the caller already holds
// the file lock.
func SaveLocked(path string, g *Glossary) error {
	return writeFile(path, g)
}

func writeFile(path string, g *Glossary) error {
	w, err := writerFor(g.Dialect)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".terminology-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if err := w.Encode(tmp, g); err != nil {
		return fmt.Errorf("encoding TBX: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("syncing temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

func detectDialect(r io.Reader) (Dialect, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = true
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return "", ErrUnsupportedDialect
		}
		if err != nil {
			return "", fmt.Errorf("detecting dialect: %w", err)
		}

		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local == "tbx" {
			for _, a := range se.Attr {
				if a.Name.Local == "type" {
					switch Dialect(a.Value) {
					case DialectLinguist:
						return DialectLinguist, nil
					default:
						return "", ErrUnsupportedDialect
					}
				}
			}
			return "", ErrUnsupportedDialect
		}
	}
}
