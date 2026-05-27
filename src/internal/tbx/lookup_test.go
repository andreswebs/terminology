package tbx

import (
	"testing"
)

func lookupGlossary() *Glossary {
	return &Glossary{
		Dialect: DialectLinguist,
		Style:   StyleDCT,
		Concepts: []Concept{
			{
				ID:           "tzimtzum",
				SubjectField: "kabbalah",
				Languages: map[string]LangSection{
					"he": {
						Lang: "he",
						Terms: []Term{
							{Surface: "צמצום", AdministrativeStatus: StatusPreferred},
						},
					},
					"en": {
						Lang: "en",
						Terms: []Term{
							{Surface: "tzimtzum", AdministrativeStatus: StatusPreferred},
						},
					},
					"es": {
						Lang: "es",
						Terms: []Term{
							{Surface: "tzimtzum", AdministrativeStatus: StatusPreferred},
						},
					},
				},
			},
			{
				ID:           "sefirot",
				SubjectField: "kabbalah",
				Languages: map[string]LangSection{
					"he": {
						Lang: "he",
						Terms: []Term{
							{Surface: "ספירות", AdministrativeStatus: StatusPreferred},
						},
					},
					"en": {
						Lang: "en",
						Terms: []Term{
							{Surface: "sefirot", AdministrativeStatus: StatusPreferred},
							{Surface: "sephiroth", AdministrativeStatus: StatusAdmitted},
						},
					},
				},
			},
		},
	}
}

func TestLookup_ExactMatch(t *testing.T) {
	t.Parallel()
	g := lookupGlossary()

	matches := g.Lookup("tzimtzum", "")
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	m := matches[0]
	if m.Concept.ID != "tzimtzum" {
		t.Errorf("ConceptID = %q, want %q", m.Concept.ID, "tzimtzum")
	}
	if m.TermLang != "en" && m.TermLang != "es" {
		t.Errorf("TermLang = %q, want en or es", m.TermLang)
	}
	if m.TermSurface != "tzimtzum" {
		t.Errorf("TermSurface = %q, want %q", m.TermSurface, "tzimtzum")
	}
}

func TestLookup_CaseInsensitive(t *testing.T) {
	t.Parallel()
	g := lookupGlossary()

	matches := g.Lookup("Tzimtzum", "")
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].Concept.ID != "tzimtzum" {
		t.Errorf("ConceptID = %q, want %q", matches[0].Concept.ID, "tzimtzum")
	}
}

func TestLookup_NFCNormalization(t *testing.T) {
	t.Parallel()
	g := &Glossary{
		Concepts: []Concept{
			{
				ID: "cafe",
				Languages: map[string]LangSection{
					"en": {
						Lang: "en",
						Terms: []Term{
							{Surface: "café", AdministrativeStatus: StatusPreferred},
						},
					},
				},
			},
		},
	}

	// Query with NFD form (e + combining acute)
	matches := g.Lookup("café", "")
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1 (NFD query against NFC term)", len(matches))
	}
	if matches[0].Concept.ID != "cafe" {
		t.Errorf("ConceptID = %q, want %q", matches[0].Concept.ID, "cafe")
	}
}

func TestLookup_LanguageFilter(t *testing.T) {
	t.Parallel()
	g := lookupGlossary()

	matches := g.Lookup("tzimtzum", "he")
	if len(matches) != 0 {
		t.Fatalf("got %d matches for lang=he with term 'tzimtzum', want 0", len(matches))
	}

	matches = g.Lookup("tzimtzum", "en")
	if len(matches) != 1 {
		t.Fatalf("got %d matches for lang=en, want 1", len(matches))
	}
	if matches[0].TermLang != "en" {
		t.Errorf("TermLang = %q, want %q", matches[0].TermLang, "en")
	}
}

func TestLookup_NoMatchReturnsEmptySlice(t *testing.T) {
	t.Parallel()
	g := lookupGlossary()

	matches := g.Lookup("nonexistent", "")
	if matches == nil {
		t.Fatal("got nil, want non-nil empty slice")
	}
	if len(matches) != 0 {
		t.Errorf("got %d matches, want 0", len(matches))
	}
}

func TestLookup_EmptyQueryReturnsEmptySlice(t *testing.T) {
	t.Parallel()
	g := lookupGlossary()

	matches := g.Lookup("", "")
	if matches == nil {
		t.Fatal("got nil, want non-nil empty slice")
	}
	if len(matches) != 0 {
		t.Errorf("got %d matches, want 0", len(matches))
	}
}

func TestLookup_MultipleConcepts(t *testing.T) {
	t.Parallel()
	g := &Glossary{
		Concepts: []Concept{
			{
				ID: "c1",
				Languages: map[string]LangSection{
					"en": {Lang: "en", Terms: []Term{{Surface: "light"}}},
				},
			},
			{
				ID: "c2",
				Languages: map[string]LangSection{
					"en": {Lang: "en", Terms: []Term{{Surface: "light"}}},
				},
			},
		},
	}

	matches := g.Lookup("light", "")
	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2", len(matches))
	}

	ids := map[string]bool{}
	for _, m := range matches {
		ids[m.Concept.ID] = true
	}
	if !ids["c1"] || !ids["c2"] {
		t.Errorf("expected both c1 and c2, got %v", ids)
	}
}

func TestLookup_MatchesAdmittedTerm(t *testing.T) {
	t.Parallel()
	g := lookupGlossary()

	matches := g.Lookup("sephiroth", "")
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].Concept.ID != "sefirot" {
		t.Errorf("ConceptID = %q, want %q", matches[0].Concept.ID, "sefirot")
	}
	if matches[0].TermSurface != "sephiroth" {
		t.Errorf("TermSurface = %q, want %q", matches[0].TermSurface, "sephiroth")
	}
}

func TestLookup_HebrewTerm(t *testing.T) {
	t.Parallel()
	g := lookupGlossary()

	matches := g.Lookup("צמצום", "")
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].Concept.ID != "tzimtzum" {
		t.Errorf("ConceptID = %q, want %q", matches[0].Concept.ID, "tzimtzum")
	}
	if matches[0].TermLang != "he" {
		t.Errorf("TermLang = %q, want %q", matches[0].TermLang, "he")
	}
}

func TestLookup_ConceptReturnedOnce(t *testing.T) {
	t.Parallel()
	g := lookupGlossary()

	// "tzimtzum" appears in both en and es, but concept should appear once
	// per matching (lang, term) pair — but the concept itself should be
	// deduplicated so we get one match per concept
	matches := g.Lookup("tzimtzum", "")
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1 (concept deduped)", len(matches))
	}
}
