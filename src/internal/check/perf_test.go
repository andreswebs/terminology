//go:build perf

package check_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/andreswebs/terminology/internal/check"
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
						{Surface: fmt.Sprintf("source%d", i), AdministrativeStatus: tbx.StatusPreferred},
					},
				},
				"es": {
					Lang: "es",
					Terms: []tbx.Term{
						{Surface: fmt.Sprintf("target%d", i), AdministrativeStatus: tbx.StatusPreferred},
					},
				},
			},
		}
	}
	return &tbx.Glossary{Concepts: concepts}
}

func generateMarkdown(targetBytes int, lang string, glossary *tbx.Glossary, termDensity int) []byte {
	var terms []string
	for _, c := range glossary.Concepts {
		if ls, ok := c.Languages[lang]; ok && len(ls.Terms) > 0 {
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

func TestPerf_Check_5000Concepts_5MB(t *testing.T) {
	const budget = 10 * time.Second

	g := perfGlossary(5000)
	srcText := generateMarkdown(5*1024*1024, "en", g, 50)
	tgtText := generateMarkdown(5*1024*1024, "es", g, 50)

	start := time.Now()
	_, err := check.Check(g, srcText, tgtText, "en", "es", 80, false)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if elapsed > budget {
		t.Fatalf("check 5000 concepts × 5MB × 2 files took %s, budget %s", elapsed, budget)
	}
	t.Logf("check 5000 concepts × 5MB × 2: %s (budget %s)", elapsed, budget)
}
