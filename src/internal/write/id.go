package write

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func DeriveID(term string) (string, error) {
	// NFKD normalize
	b := norm.NFKD.Bytes([]byte(term))

	// Drop combining marks
	var stripped []byte
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if !unicode.Is(unicode.Mn, r) {
			stripped = append(stripped, b[i:i+size]...)
		}
		i += size
	}

	// Unicode default casefold
	folded := cases.Fold().Bytes(stripped)

	// Replace non-[a-z0-9] runs with single hyphen
	slug := nonAlnum.ReplaceAllString(string(folded), "-")

	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	// Truncate to 64 codepoints on hyphen boundary
	slug = truncate(slug, 64)

	if slug == "" {
		return "", ErrInvalidID
	}

	return slug, nil
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}

	runes = runes[:max]
	result := string(runes)

	if result[len(result)-1] == '-' {
		return strings.TrimRight(result, "-")
	}

	if idx := strings.LastIndex(result, "-"); idx > 0 {
		return result[:idx]
	}
	return result
}
