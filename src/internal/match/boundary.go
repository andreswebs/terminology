package match

import (
	"unicode"
	"unicode/utf8"
)

func validBoundary(orig []byte, start, end int) bool {
	if start > 0 {
		r, _ := utf8.DecodeLastRune(orig[:start])
		if unicode.Is(unicode.Letter, r) || unicode.Is(unicode.Number, r) {
			return false
		}
	}
	if end < len(orig) {
		r, _ := utf8.DecodeRune(orig[end:])
		if unicode.Is(unicode.Letter, r) || unicode.Is(unicode.Number, r) {
			return false
		}
	}
	return true
}
