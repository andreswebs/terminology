package tbx

import (
	"fmt"
	"sort"

	"golang.org/x/text/language"
)

type ValidateResult struct {
	Concepts  int
	Languages []string
	Warnings  []Warning
	Errors    []Warning
}

func (g *Glossary) Validate(strict bool) ValidateResult {
	var res ValidateResult
	res.Concepts = len(g.Concepts)

	langSet := make(map[string]struct{})
	idSet := make(map[string]int)

	for i, c := range g.Concepts {
		if prev, dup := idSet[c.ID]; dup {
			res.Warnings = append(res.Warnings, Warning{
				Code:      "duplicate_id",
				Message:   fmt.Sprintf("concept ID %q already declared at index %d", c.ID, prev),
				ConceptID: c.ID,
				Line:      c.StartLine,
				Col:       c.StartCol,
			})
		}
		idSet[c.ID] = i
	}

	for _, c := range g.Concepts {
		for lang, ls := range c.Languages {
			langSet[lang] = struct{}{}

			_, err := language.Parse(lang)
			if err != nil {
				res.Warnings = append(res.Warnings, Warning{
					Code:      "invalid_lang_tag",
					Message:   fmt.Sprintf("concept %q: malformed BCP 47 tag %q", c.ID, lang),
					ConceptID: c.ID,
					Line:      ls.StartLine,
					Col:       ls.StartCol,
				})
			}

			if len(ls.Terms) == 0 {
				res.Warnings = append(res.Warnings, Warning{
					Code:      "missing_term",
					Message:   fmt.Sprintf("concept %q lang %q: langSec has no term", c.ID, lang),
					ConceptID: c.ID,
					Line:      ls.StartLine,
					Col:       ls.StartCol,
				})
			}
		}

		for _, cr := range c.CrossRefs {
			if _, found := idSet[cr.Target]; !found {
				w := Warning{
					Code:      "unresolved_crossref",
					Message:   fmt.Sprintf("concept %q references unknown ID %q", c.ID, cr.Target),
					ConceptID: c.ID,
					Line:      c.StartLine,
					Col:       c.StartCol,
				}
				if strict {
					res.Errors = append(res.Errors, w)
				} else {
					res.Warnings = append(res.Warnings, w)
				}
			}
		}

		for _, ls := range c.Languages {
			for _, t := range ls.Terms {
				for _, cr := range t.CrossRefs {
					if _, found := idSet[cr.Target]; !found {
						w := Warning{
							Code:      "unresolved_crossref",
							Message:   fmt.Sprintf("concept %q term %q references unknown ID %q", c.ID, t.Surface, cr.Target),
							ConceptID: c.ID,
							Line:      c.StartLine,
							Col:       c.StartCol,
						}
						if strict {
							res.Errors = append(res.Errors, w)
						} else {
							res.Warnings = append(res.Warnings, w)
						}
					}
				}
			}
		}
	}

	langs := make([]string, 0, len(langSet))
	for l := range langSet {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	res.Languages = langs

	return res
}
