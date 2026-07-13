package match

import (
	"unicode/utf8"

	"github.com/andreswebs/terminology/internal/markdown"
	"github.com/andreswebs/terminology/internal/tbx"
)

type termPattern struct {
	ConceptID string
	Term      string
	Lang      string
	Status    string
}

// Match is a single glossary term occurrence found in scanned text, including
// its concept, language, status, position, and surrounding context.
type Match struct {
	ConceptID string
	Term      string
	Lang      string
	Status    string
	Line      int
	Column    int
	Context   string
}

// Matcher scans text for glossary terms using a prebuilt automaton and the
// normalization policy the terms were compiled under.
type Matcher struct {
	policy   Policy
	machine  *automaton
	patterns []termPattern
}

// New builds a Matcher for the terms in g. When lang is non-empty only that
// language's terms are included; otherwise all languages are used and their
// per-language policies are merged. The policy argument sets the base policy.
func New(g *tbx.Glossary, lang string, policy Policy) (*Matcher, error) {
	var patterns []termPattern
	var canonical [][]byte

	merged := policy
	for _, c := range g.Concepts {
		for ltag, ls := range c.Languages {
			if lang != "" && ltag != lang {
				continue
			}
			patPolicy := policy
			if lang == "" {
				patPolicy = PolicyFor(ltag)
				merged = mergePolicy(merged, patPolicy)
			}
			for _, t := range ls.Terms {
				cn := Normalize([]byte(t.Surface), patPolicy)
				patterns = append(patterns, termPattern{
					ConceptID: c.ID,
					Term:      t.Surface,
					Lang:      ltag,
					Status:    statusLabel(t.AdministrativeStatus),
				})
				canonical = append(canonical, cn.Bytes)
			}
		}
	}

	return &Matcher{
		policy:   merged,
		machine:  buildAutomaton(canonical),
		patterns: patterns,
	}, nil
}

// ScanText is the high-level entry point for a single scan: it builds a
// matcher for lang over g, extracts the markdown spans from data, and scans
// them, applying the language-specific policy. Callers needing the lower-level
// pieces can use New and Matcher.Scan directly.
func ScanText(g *tbx.Glossary, data []byte, lang string, contextSize int) ([]Match, error) {
	matcher, err := New(g, lang, PolicyFor(lang))
	if err != nil {
		return nil, err
	}

	var spans []markdown.Span
	for s := range markdown.Spans(data) {
		spans = append(spans, s)
	}

	return matcher.Scan(data, spans, contextSize), nil
}

// Scan searches the given spans for glossary term matches, returning one Match
// per valid occurrence with contextSize bytes of surrounding context (defaulting
// to 80 when not positive).
func (m *Matcher) Scan(_ []byte, spans []markdown.Span, contextSize int) []Match {
	if contextSize <= 0 {
		contextSize = 80
	}
	var matches []Match

	for _, sp := range spans {
		spBytes := []byte(sp.Text)
		cn := Normalize(spBytes, m.policy)

		raw := m.machine.Search(cn.Bytes)
		raw = longestMatchPerStart(raw)

		for _, rm := range raw {
			origStart := cn.Map[rm.Start]
			lastSrcOff := cn.Map[rm.End-1]
			_, runeSize := utf8.DecodeRune(spBytes[lastSrcOff:])
			origEnd := lastSrcOff + runeSize

			if !validBoundary(spBytes, origStart, origEnd) {
				continue
			}

			pat := m.patterns[rm.PatternID]

			line, col := offsetToLineCol(sp, origStart)

			matches = append(matches, Match{
				ConceptID: pat.ConceptID,
				Term:      pat.Term,
				Lang:      pat.Lang,
				Status:    pat.Status,
				Line:      line,
				Column:    col,
				Context:   extractContext(sp.Text, origStart, origEnd, contextSize),
			})
		}
	}

	return matches
}

func offsetToLineCol(sp markdown.Span, byteOffset int) (int, int) {
	text := sp.Text
	line := sp.Line
	col := sp.Col

	for i := 0; i < byteOffset && i < len(text); i++ {
		if text[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

func extractContext(text string, start, end, contextSize int) string {
	window := contextSize / 2

	ctxStart := max(start-window, 0)
	ctxEnd := min(end+window, len(text))

	var prefix, suffix string
	if ctxStart > 0 {
		prefix = "..."
	}
	if ctxEnd < len(text) {
		suffix = "..."
	}

	return prefix + text[ctxStart:ctxEnd] + suffix
}

func statusLabel(s tbx.Status) string {
	switch s {
	case tbx.StatusPreferred:
		return "preferred"
	case tbx.StatusAdmitted:
		return "admitted"
	case tbx.StatusDeprecated:
		return "deprecated"
	case tbx.StatusSuperseded:
		return "superseded"
	default:
		return "unspecified"
	}
}
