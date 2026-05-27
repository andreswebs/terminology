package extract

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

func HighFrequencyTokens(spans []Span, opts Options) []Candidate {
	minFreq := opts.MinFreq
	if minFreq <= 0 {
		minFreq = 3
	}

	type entry struct {
		candidate Candidate
	}
	agg := make(map[string]*entry)
	fold := cases.Fold()

	for _, span := range spans {
		words := tokenize(span.Text)
		for _, w := range words {
			if w.isSep {
				continue
			}
			cleaned := stripTrailingPunct(w.text)
			if cleaned == "" {
				continue
			}
			key := fold.String(norm.NFC.String(cleaned))

			if opts.Stopwords != nil && opts.Stopwords[key] {
				continue
			}

			e, ok := agg[key]
			if !ok {
				e = &entry{
					candidate: Candidate{
						Term:      key,
						Heuristic: "high_frequency",
					},
				}
				agg[key] = e
			}
			e.candidate.Frequency++
			line, col := span.lineColAt(w.offset)
			e.candidate.Locations = append(e.candidate.Locations, Location{
				Line:   line,
				Col:    col,
				Offset: span.Offset + w.offset,
			})
		}
	}

	result := make([]Candidate, 0, len(agg))
	for _, e := range agg {
		if e.candidate.Frequency >= minFreq {
			result = append(result, e.candidate)
		}
	}
	return result
}
