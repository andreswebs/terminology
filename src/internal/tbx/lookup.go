package tbx

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// LookupMatch is a term found by Lookup, identifying the concept, the language
// it belongs to, and the term's original surface form.
type LookupMatch struct {
	Concept     Concept
	TermLang    string
	TermSurface string
}

// Lookup returns the matches for term across the glossary, comparing
// case-insensitively on NFC-normalized surface forms. An empty lang matches any
// language; an empty term returns no matches.
func (g *Glossary) Lookup(term, lang string) []LookupMatch {
	results := []LookupMatch{}

	if term == "" {
		return results
	}

	folder := cases.Fold()
	query := folder.String(norm.NFC.String(term))

	for _, concept := range g.Concepts {
		if match, ok := matchConcept(concept, query, lang, folder); ok {
			results = append(results, match)
		}
	}

	return results
}

func matchConcept(c Concept, query, lang string, folder cases.Caser) (LookupMatch, bool) {
	for langTag, ls := range c.Languages {
		if lang != "" && langTag != lang {
			continue
		}
		for _, t := range ls.Terms {
			surface := folder.String(norm.NFC.String(t.Surface))
			if surface == query {
				return LookupMatch{
					Concept:     c,
					TermLang:    langTag,
					TermSurface: t.Surface,
				}, true
			}
		}
	}
	return LookupMatch{}, false
}
