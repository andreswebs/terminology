// Package extract discovers candidate terminology in source text using
// heuristics such as capitalized phrases, foreign-script tokens, and
// high-frequency tokens.
package extract

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// CapitalizedPhrases returns candidates built from runs of consecutive
// capitalized words found in spans, excluding words at sentence start.
func CapitalizedPhrases(spans []Span, _ string) []Candidate {
	type entry struct {
		candidate Candidate
	}
	agg := make(map[string]*entry)

	for _, span := range spans {
		phrases := findCapitalizedPhrases(span)
		for _, p := range phrases {
			key := norm.NFC.String(p.term)
			e, ok := agg[key]
			if !ok {
				e = &entry{
					candidate: Candidate{
						Term:      key,
						Heuristic: "capitalized_phrase",
					},
				}
				agg[key] = e
			}
			e.candidate.Frequency++
			line, col := span.lineColAt(p.offset)
			e.candidate.Locations = append(e.candidate.Locations, Location{
				Line:   line,
				Col:    col,
				Offset: span.Offset + p.offset,
			})
		}
	}

	result := make([]Candidate, 0, len(agg))
	for _, e := range agg {
		result = append(result, e.candidate)
	}
	return result
}

type phrase struct {
	term   string
	col    int
	offset int
}

func findCapitalizedPhrases(span Span) []phrase {
	text := span.Text
	words := tokenize(text)

	var phrases []phrase
	var current []token
	afterSentenceStart := true

	for _, w := range words {
		if w.isSep {
			if endsSentence(w.text) {
				afterSentenceStart = true
				if len(current) > 0 {
					phrases = append(phrases, buildPhrase(current))
					current = nil
				}
			}
			continue
		}

		r, _ := utf8.DecodeRuneInString(w.text)
		isCap := unicode.IsUpper(r)

		if afterSentenceStart {
			afterSentenceStart = endsSentence(w.text)
			if len(current) > 0 {
				phrases = append(phrases, buildPhrase(current))
				current = nil
			}
			continue
		}

		if isCap {
			current = append(current, w)
		} else {
			if len(current) > 0 {
				phrases = append(phrases, buildPhrase(current))
				current = nil
			}
		}

		if endsSentence(w.text) {
			afterSentenceStart = true
			if len(current) > 0 {
				phrases = append(phrases, buildPhrase(current))
				current = nil
			}
		}
	}

	if len(current) > 0 {
		phrases = append(phrases, buildPhrase(current))
	}

	return phrases
}

type token struct {
	text   string
	offset int
	isSep  bool
}

func tokenize(text string) []token {
	var tokens []token
	i := 0
	for i < len(text) {
		r, _ := utf8.DecodeRuneInString(text[i:])
		isSep := unicode.IsSpace(r)
		start := i
		for i < len(text) {
			r, sz := utf8.DecodeRuneInString(text[i:])
			if unicode.IsSpace(r) != isSep {
				break
			}
			i += sz
		}
		tokens = append(tokens, token{text: text[start:i], offset: start, isSep: isSep})
	}
	return tokens
}

func endsSentence(text string) bool {
	for i := len(text) - 1; i >= 0; i-- {
		switch text[i] {
		case '.', '!', '?':
			return true
		case ' ', '\t', '\n', '\r':
			continue
		default:
			return false
		}
	}
	return false
}

func buildPhrase(tokens []token) phrase {
	if len(tokens) == 0 {
		return phrase{}
	}
	start := tokens[0].offset
	var combined strings.Builder
	for i, tok := range tokens {
		if i > 0 {
			combined.WriteString(" ")
		}
		combined.WriteString(stripTrailingPunct(tok.text))
	}
	return phrase{
		term:   combined.String(),
		col:    start,
		offset: start,
	}
}

func stripTrailingPunct(s string) string {
	runes := []rune(s)
	for len(runes) > 0 && unicode.IsPunct(runes[len(runes)-1]) {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}
