//go:build perf

package match_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/andreswebs/terminology/internal/markdown"
	"github.com/andreswebs/terminology/internal/match"
	"github.com/andreswebs/terminology/internal/tbx"
)

func perfGlossary(n int) *tbx.Glossary {
	concepts := make([]tbx.Concept, n)
	for i := range n {
		concepts[i] = tbx.Concept{
			ID: fmt.Sprintf("concept-%05d", i),
			Languages: map[string]tbx.LangSection{
				"en": {
					Lang: "en",
					Terms: []tbx.Term{
						{Surface: fmt.Sprintf("terminology%d", i), AdministrativeStatus: tbx.StatusPreferred},
					},
				},
			},
		}
	}
	return &tbx.Glossary{Concepts: concepts}
}

func generateMarkdown(targetBytes int, glossary *tbx.Glossary, termDensity int) []byte {
	var terms []string
	for _, c := range glossary.Concepts {
		if ls, ok := c.Languages["en"]; ok && len(ls.Terms) > 0 {
			terms = append(terms, ls.Terms[0].Surface)
		}
	}
	if termDensity <= 0 {
		termDensity = 20
	}

	var b strings.Builder
	fillers := []string{
		"The quick brown fox jumps over the lazy dog. ",
		"Academic scholarship continues to advance our understanding. ",
		"This paragraph contains ordinary prose without special vocabulary. ",
		"Further analysis reveals patterns in the underlying structure. ",
	}
	lineCount := 0
	termIdx := 0
	for b.Len() < targetBytes {
		b.WriteString(fillers[lineCount%len(fillers)])
		lineCount++
		if len(terms) > 0 && lineCount%termDensity == 0 {
			b.WriteString(terms[termIdx%len(terms)])
			b.WriteString(" is discussed here. ")
			termIdx++
		}
		if lineCount%10 == 0 {
			b.WriteString("\n\n")
		}
	}
	return []byte(b.String())
}

func collectSpans(src []byte) []markdown.Span {
	var spans []markdown.Span
	for s := range markdown.Spans(src) {
		spans = append(spans, s)
	}
	return spans
}

func TestPerf_Scan_200Concepts_100KB(t *testing.T) {
	const budget = 100 * time.Millisecond

	g := perfGlossary(200)
	src := generateMarkdown(100*1024, g, 5)
	spans := collectSpans(src)

	m, err := match.New(g, "en", match.PolicyFor("en"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	start := time.Now()
	_ = m.Scan(src, spans, 80)
	elapsed := time.Since(start)

	if elapsed > budget {
		t.Fatalf("scan 200 concepts × 100KB took %s, budget %s", elapsed, budget)
	}
	t.Logf("scan 200 concepts × 100KB: %s (budget %s)", elapsed, budget)
}

func TestPerf_Scan_5000Concepts_5MB(t *testing.T) {
	const budget = 5 * time.Second

	g := perfGlossary(5000)
	src := generateMarkdown(5*1024*1024, g, 50)
	spans := collectSpans(src)

	m, err := match.New(g, "en", match.PolicyFor("en"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	start := time.Now()
	_ = m.Scan(src, spans, 80)
	elapsed := time.Since(start)

	if elapsed > budget {
		t.Fatalf("scan 5000 concepts × 5MB took %s, budget %s", elapsed, budget)
	}
	t.Logf("scan 5000 concepts × 5MB: %s (budget %s)", elapsed, budget)
}
