package match

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

type Canonical struct {
	Bytes []byte
	Map   []int
}

func Normalize(src []byte, p Policy) Canonical {
	if len(src) == 0 {
		return Canonical{}
	}

	var nf norm.Form
	if p.FoldDiacritics {
		nf = norm.NFD
	} else if p.Normalize == NFC {
		nf = norm.NFC
	} else {
		nf = norm.NFKD
	}

	normBytes, normMap := applyNormForm(src, nf)

	folder := cases.Fold()
	var out []byte
	var finalMap []int
	inWS := false

	i := 0
	for i < len(normBytes) {
		r, size := utf8.DecodeRune(normBytes[i:])
		srcOff := normMap[i]

		if p.FoldDiacritics && unicode.Is(unicode.Mn, r) {
			i += size
			continue
		}

		if p.StripNiqqud && isNiqqud(r) {
			i += size
			continue
		}

		if unicode.IsSpace(r) {
			if !inWS {
				inWS = true
				out = append(out, ' ')
				finalMap = append(finalMap, srcOff)
			}
			i += size
			continue
		}
		inWS = false

		if p.CaseFold {
			var buf [utf8.UTFMax]byte
			n := utf8.EncodeRune(buf[:], r)
			folded := folder.Bytes(buf[:n])
			for _, b := range folded {
				out = append(out, b)
				finalMap = append(finalMap, srcOff)
			}
		} else {
			for j := range size {
				out = append(out, normBytes[i+j])
				finalMap = append(finalMap, srcOff)
			}
		}

		i += size
	}

	return Canonical{Bytes: out, Map: finalMap}
}

func applyNormForm(src []byte, f norm.Form) ([]byte, []int) {
	var dst []byte
	var m []int

	pos := 0
	for pos < len(src) {
		segStart := pos
		_, sz := utf8.DecodeRune(src[pos:])
		pos += sz
		for pos < len(src) {
			p := f.Properties(src[pos:])
			if p.BoundaryBefore() {
				break
			}
			_, sz = utf8.DecodeRune(src[pos:])
			pos += sz
		}

		segNorm := f.Bytes(src[segStart:pos])
		for _, b := range segNorm {
			dst = append(dst, b)
			m = append(m, segStart)
		}
	}

	return dst, m
}

func isNiqqud(r rune) bool {
	return r >= 0x0591 && r <= 0x05C7
}
