package check

import (
	"sort"

	"github.com/andreswebs/terminology/internal/markdown"
	"github.com/andreswebs/terminology/internal/match"
	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
)

type Result struct {
	Violations      []output.CheckViolation
	Warnings        []output.CheckWarning
	ConceptsChecked int
}

func Check(g *tbx.Glossary, srcText, tgtText []byte,
	srcLang, tgtLang string, contextSize int, strict bool) (*Result, error) {

	srcMatcher, err := match.New(g, srcLang, match.PolicyFor(srcLang))
	if err != nil {
		return nil, err
	}

	var srcSpans []markdown.Span
	for s := range markdown.Spans(srcText) {
		srcSpans = append(srcSpans, s)
	}

	srcMatches := srcMatcher.Scan(srcText, srcSpans, contextSize)

	type srcInfo struct {
		count      int
		sourceTerm string
	}
	srcByConcept := make(map[string]*srcInfo)
	for _, m := range srcMatches {
		if m.Status == "deprecated" || m.Status == "superseded" {
			continue
		}
		info, ok := srcByConcept[m.ConceptID]
		if !ok {
			srcByConcept[m.ConceptID] = &srcInfo{count: 1, sourceTerm: m.Term}
		} else {
			info.count++
		}
	}

	if len(srcByConcept) == 0 {
		return &Result{}, nil
	}

	tgtMatcher, err := match.New(g, tgtLang, match.PolicyFor(tgtLang))
	if err != nil {
		return nil, err
	}

	var tgtSpans []markdown.Span
	for s := range markdown.Spans(tgtText) {
		tgtSpans = append(tgtSpans, s)
	}

	tgtMatches := tgtMatcher.Scan(tgtText, tgtSpans, contextSize)

	tgtPreferred := make(map[string]int)
	type posHit struct {
		conceptID string
		variant   string
		line      int
		column    int
		context   string
	}
	var forbiddenHits []posHit
	var admittedHits []posHit

	for _, m := range tgtMatches {
		if _, inSrc := srcByConcept[m.ConceptID]; !inSrc {
			continue
		}

		switch m.Status {
		case "preferred", "unspecified":
			tgtPreferred[m.ConceptID]++
		case "deprecated", "superseded":
			forbiddenHits = append(forbiddenHits, posHit{
				conceptID: m.ConceptID,
				variant:   m.Term,
				line:      m.Line,
				column:    m.Column,
				context:   m.Context,
			})
		case "admitted":
			admittedHits = append(admittedHits, posHit{
				conceptID: m.ConceptID,
				variant:   m.Term,
				line:      m.Line,
				column:    m.Column,
				context:   m.Context,
			})
		}
	}

	var violations []output.CheckViolation
	var warnings []output.CheckWarning

	for _, fh := range forbiddenHits {
		violations = append(violations, output.CheckViolation{
			Type:      "forbidden_variant",
			ConceptID: fh.conceptID,
			Variant:   fh.variant,
			Line:      fh.line,
			Column:    fh.column,
			Context:   fh.context,
		})
	}

	for _, ah := range admittedHits {
		if strict {
			violations = append(violations, output.CheckViolation{
				Type:      "admitted_variant",
				ConceptID: ah.conceptID,
				Variant:   ah.variant,
				Line:      ah.line,
				Column:    ah.column,
				Context:   ah.context,
			})
		} else {
			warnings = append(warnings, output.CheckWarning{
				Type:      "admitted_variant",
				ConceptID: ah.conceptID,
				Variant:   ah.variant,
				Line:      ah.line,
				Column:    ah.column,
				Context:   ah.context,
			})
		}
	}

	conceptsWithBothLangs := 0
	for conceptID, info := range srcByConcept {
		preferred := preferredTarget(g, conceptID, tgtLang)
		if preferred == "" {
			continue
		}
		conceptsWithBothLangs++

		if tgtPreferred[conceptID] == 0 {
			violations = append(violations, output.CheckViolation{
				Type:              "missing",
				ConceptID:         conceptID,
				SourceTerm:        info.sourceTerm,
				ExpectedTarget:    preferred,
				SourceOccurrences: info.count,
			})
		}
	}

	sortViolations(violations)
	sortWarnings(warnings)

	return &Result{
		Violations:      violations,
		Warnings:        warnings,
		ConceptsChecked: conceptsWithBothLangs,
	}, nil
}

func preferredTarget(g *tbx.Glossary, conceptID, lang string) string {
	for _, c := range g.Concepts {
		if c.ID != conceptID {
			continue
		}
		ls, ok := c.Languages[lang]
		if !ok {
			return ""
		}
		for _, t := range ls.Terms {
			if t.AdministrativeStatus == tbx.StatusPreferred {
				return t.Surface
			}
		}
		if len(ls.Terms) > 0 {
			return ls.Terms[0].Surface
		}
		return ""
	}
	return ""
}

func isPositionalWarning(w output.CheckWarning) bool {
	return w.Line != 0 || w.Column != 0
}

func sortWarnings(ws []output.CheckWarning) {
	sort.SliceStable(ws, func(i, j int) bool {
		iPosn := isPositionalWarning(ws[i])
		jPosn := isPositionalWarning(ws[j])

		if iPosn && !jPosn {
			return true
		}
		if !iPosn && jPosn {
			return false
		}

		if iPosn && jPosn {
			if ws[i].Line != ws[j].Line {
				return ws[i].Line < ws[j].Line
			}
			if ws[i].Column != ws[j].Column {
				return ws[i].Column < ws[j].Column
			}
			return ws[i].ConceptID < ws[j].ConceptID
		}

		return ws[i].ConceptID < ws[j].ConceptID
	})
}

func sortViolations(vs []output.CheckViolation) {
	sort.SliceStable(vs, func(i, j int) bool {
		iPositional := vs[i].Type != "missing"
		jPositional := vs[j].Type != "missing"

		if iPositional && !jPositional {
			return true
		}
		if !iPositional && jPositional {
			return false
		}

		if iPositional && jPositional {
			if vs[i].Line != vs[j].Line {
				return vs[i].Line < vs[j].Line
			}
			if vs[i].Column != vs[j].Column {
				return vs[i].Column < vs[j].Column
			}
			return vs[i].ConceptID < vs[j].ConceptID
		}

		return vs[i].ConceptID < vs[j].ConceptID
	})
}
