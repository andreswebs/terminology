package tbx

import (
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// SearchOptions configures Glossary.Search.
type SearchOptions struct {
	Lang    string   // restrict the haystack to this language tag ("" = all)
	Include []string // widen haystack: "definitions","notes","contexts","subject_field"
}

// Search returns concepts whose normalized fields contain the normalized query
// as a substring, ordered by concept id. The default haystack is the concept
// id plus, per language section, each term's surface, reading, and reading
// note. opts.Include widens it to concept- and langSec-level definitions,
// notes, term contexts, and the subject field. An empty query matches nothing.
//
// Matching is diacritic- and separator-insensitive (see foldForSearch): romaji
// matches regardless of hyphens, spaces, or macrons, while CJK and kana survive
// normalization and match by substring. This is deliberately distinct from the
// boundary-aware prose matcher in internal/match.
func (g *Glossary) Search(query string, opts SearchOptions) []Concept {
	results := []Concept{}

	nq := foldForSearch(query)
	if nq == "" {
		return results
	}

	include := make(map[string]bool, len(opts.Include))
	for _, k := range opts.Include {
		include[k] = true
	}

	for _, c := range g.Concepts {
		if conceptMatchesSearch(c, nq, opts.Lang, include) {
			results = append(results, c)
		}
	}

	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results
}

func conceptMatchesSearch(c Concept, nq, lang string, include map[string]bool) bool {
	if strings.Contains(foldForSearch(c.ID), nq) {
		return true
	}
	if include["subject_field"] && strings.Contains(foldForSearch(c.SubjectField), nq) {
		return true
	}
	if include["definitions"] {
		for _, d := range c.Definitions {
			if strings.Contains(foldForSearch(d.Plain), nq) {
				return true
			}
		}
	}
	if include["notes"] {
		for _, n := range c.Notes {
			if strings.Contains(foldForSearch(n), nq) {
				return true
			}
		}
	}

	for tag, ls := range c.Languages {
		if lang != "" && tag != lang {
			continue
		}
		if langSectionMatchesSearch(ls, nq, include) {
			return true
		}
	}
	return false
}

func langSectionMatchesSearch(ls LangSection, nq string, include map[string]bool) bool {
	if include["definitions"] {
		for _, d := range ls.Definitions {
			if strings.Contains(foldForSearch(d.Plain), nq) {
				return true
			}
		}
	}
	for _, t := range ls.Terms {
		if strings.Contains(foldForSearch(t.Surface), nq) ||
			strings.Contains(foldForSearch(t.Reading), nq) ||
			strings.Contains(foldForSearch(t.ReadingNote), nq) {
			return true
		}
		if include["contexts"] {
			for _, ctx := range t.Contexts {
				if strings.Contains(foldForSearch(ctx.Plain), nq) {
					return true
				}
			}
		}
		if include["notes"] {
			for _, n := range t.Notes {
				if strings.Contains(foldForSearch(n), nq) {
					return true
				}
			}
		}
	}
	return false
}

var searchFolder = cases.Fold()

// foldForSearch normalizes s for diacritic- and separator-insensitive substring
// matching: NFKD-decompose, drop combining marks (so a macron folds to its base
// letter), keep only letters and numbers (dropping hyphens, spaces, and other
// separators), and case-fold. CJK ideographs and kana are letters, so they
// survive unchanged. It mirrors the reference norm() in terminology-search.py.
func foldForSearch(s string) string {
	if s == "" {
		return ""
	}
	decomposed := norm.NFKD.String(s)
	var b strings.Builder
	b.Grow(len(decomposed))
	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		}
	}
	return searchFolder.String(b.String())
}
