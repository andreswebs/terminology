//go:build perf

package tbx_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/andreswebs/terminology/internal/tbx"
)

func generateGlossary(n int, langsPerConcept int) *tbx.Glossary {
	concepts := make([]tbx.Concept, n)
	langs := []string{"en", "es", "he", "fr", "de"}
	if langsPerConcept > len(langs) {
		langsPerConcept = len(langs)
	}

	for i := range n {
		languages := make(map[string]tbx.LangSection, langsPerConcept)
		for j := range langsPerConcept {
			lang := langs[j]
			languages[lang] = tbx.LangSection{
				Lang: lang,
				Terms: []tbx.Term{
					{Surface: fmt.Sprintf("term-%s-%d", lang, i), AdministrativeStatus: tbx.StatusPreferred},
					{Surface: fmt.Sprintf("variant-%s-%d", lang, i), AdministrativeStatus: tbx.StatusAdmitted},
				},
			}
		}
		concepts[i] = tbx.Concept{
			ID:           fmt.Sprintf("concept-%05d", i),
			SubjectField: "test",
			Languages:    languages,
		}
	}

	return &tbx.Glossary{Concepts: concepts}
}

func TestPerf_Validate_10K(t *testing.T) {
	const budget = 500 * time.Millisecond

	g := generateGlossary(10000, 3)

	start := time.Now()
	_ = g.Validate(false)
	elapsed := time.Since(start)

	if elapsed > budget {
		t.Fatalf("validate 10k concepts took %s, budget %s", elapsed, budget)
	}
	t.Logf("validate 10k concepts: %s (budget %s)", elapsed, budget)
}

func TestPerf_Lookup_10K(t *testing.T) {
	const budget = 50 * time.Millisecond

	g := generateGlossary(10000, 3)

	start := time.Now()
	_ = g.Lookup("term-en-5000", "")
	elapsed := time.Since(start)

	if elapsed > budget {
		t.Fatalf("lookup in 10k-concept glossary took %s, budget %s", elapsed, budget)
	}
	t.Logf("lookup in 10k concepts: %s (budget %s)", elapsed, budget)
}
