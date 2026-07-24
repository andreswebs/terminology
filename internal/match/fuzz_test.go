package match

import (
	"testing"

	"github.com/andreswebs/terminology/internal/markdown"
	"github.com/andreswebs/terminology/internal/tbx"
)

func FuzzMatcherScan(f *testing.F) {
	f.Add("the concept of tzimtzum in Kabbalah")
	f.Add("צמצום appears here")
	f.Add("TZIMTZUM and contraction")
	f.Add("")
	f.Add("a")
	f.Add("\n\n\n")
	f.Add("the שָׁלוֹם word with niqqud")
	f.Add("tzimtzum primordial spans\nmultiple lines here")
	f.Add(string([]byte{0x00, 0xFF, 0xFE}))

	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID: "tzimtzum",
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{
						{Surface: "tzimtzum", AdministrativeStatus: tbx.StatusPreferred},
						{Surface: "contraction", AdministrativeStatus: tbx.StatusDeprecated},
					}},
					"he": {Lang: "he", Terms: []tbx.Term{
						{Surface: "צמצום", AdministrativeStatus: tbx.StatusPreferred},
					}},
				},
			},
			{
				ID: "sefirah",
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{
						{Surface: "sefirah", AdministrativeStatus: tbx.StatusPreferred},
					}},
					"he": {Lang: "he", Terms: []tbx.Term{
						{Surface: "סְפִירָה", AdministrativeStatus: tbx.StatusPreferred},
					}},
				},
			},
		},
	}

	m, err := New(g, "", Baseline)
	if err != nil {
		f.Fatalf("New: %v", err)
	}

	f.Fuzz(func(t *testing.T, text string) {
		spans := []markdown.Span{{Text: text, Offset: 0, Line: 1, Col: 1}}
		matches := m.Scan([]byte(text), spans, 80)

		for _, got := range matches {
			if got.ConceptID == "" {
				t.Fatal("match with empty ConceptID")
			}
			if got.Term == "" {
				t.Fatal("match with empty Term")
			}
			if got.Line < 1 {
				t.Fatalf("match with invalid Line: %d", got.Line)
			}
			if got.Column < 1 {
				t.Fatalf("match with invalid Column: %d", got.Column)
			}
		}
	})
}
