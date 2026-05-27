package extract

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type Options struct {
	Script    string
	BaseLang  string
	MinFreq   int
	Stopwords map[string]bool
}

func langToScript(lang string) *unicode.RangeTable {
	switch lang {
	case "he", "yi":
		return unicode.Hebrew
	case "ar", "fa", "ur":
		return unicode.Arabic
	case "ru", "uk", "bg", "sr", "mk", "be":
		return unicode.Cyrillic
	case "el":
		return unicode.Greek
	case "zh", "ja":
		return unicode.Han
	default:
		return unicode.Latin
	}
}

func ForeignScriptTokens(spans []Span, opts Options) []Candidate {
	var scriptFilter *unicode.RangeTable
	if opts.Script != "" && opts.Script != "any" {
		scriptFilter = scriptByName(opts.Script)
	}

	var baseScript *unicode.RangeTable
	if opts.BaseLang != "" {
		baseScript = langToScript(opts.BaseLang)
	}

	agg := make(map[string]*Candidate)

	for _, span := range spans {
		dominant := baseScript
		if dominant == nil {
			dominant = dominantScript(span.Text)
		}
		if dominant == nil {
			continue
		}

		words := tokenize(span.Text)
		for _, w := range words {
			if w.isSep {
				continue
			}
			cleaned := stripTrailingPunct(w.text)
			if cleaned == "" {
				continue
			}
			ws := dominantScript(cleaned)
			if ws == nil || ws == dominant {
				continue
			}
			if scriptFilter != nil && ws != scriptFilter {
				continue
			}

			key := norm.NFC.String(cleaned)
			c, ok := agg[key]
			if !ok {
				c = &Candidate{
					Term:      key,
					Heuristic: "foreign_script",
				}
				agg[key] = c
			}
			c.Frequency++
			line, col := span.lineColAt(w.offset)
			c.Locations = append(c.Locations, Location{
				Line:   line,
				Col:    col,
				Offset: span.Offset + w.offset,
			})
		}
	}

	result := make([]Candidate, 0, len(agg))
	for _, c := range agg {
		result = append(result, *c)
	}
	return result
}

func scriptByName(name string) *unicode.RangeTable {
	switch name {
	case "latin":
		return unicode.Latin
	case "hebrew":
		return unicode.Hebrew
	case "cyrillic":
		return unicode.Cyrillic
	case "arabic":
		return unicode.Arabic
	default:
		return nil
	}
}

func dominantScript(text string) *unicode.RangeTable {
	counts := make(map[string]int)
	var tables = map[string]*unicode.RangeTable{
		"Latin":    unicode.Latin,
		"Hebrew":   unicode.Hebrew,
		"Cyrillic": unicode.Cyrillic,
		"Arabic":   unicode.Arabic,
		"Greek":    unicode.Greek,
		"Han":      unicode.Han,
		"Hiragana": unicode.Hiragana,
		"Katakana": unicode.Katakana,
	}

	for i := 0; i < len(text); {
		r, sz := utf8.DecodeRuneInString(text[i:])
		i += sz
		if unicode.In(r, unicode.Common, unicode.Inherited) {
			continue
		}
		for name, table := range tables {
			if unicode.Is(table, r) {
				counts[name]++
				break
			}
		}
	}

	var bestName string
	var bestCount int
	for name, count := range counts {
		if count > bestCount {
			bestName = name
			bestCount = count
		}
	}
	if bestName == "" {
		return nil
	}
	return tables[bestName]
}
