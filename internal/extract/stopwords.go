package extract

import (
	"bufio"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// LoadStopwords reads a stopword file at path, one word per line, ignoring
// blank lines and comments, and returns the case-folded, NFC-normalized set.
func LoadStopwords(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	fold := cases.Fold()
	sw := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key := fold.String(norm.NFC.String(line))
		sw[key] = true
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sw, nil
}
