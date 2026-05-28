package match

import (
	"testing"

	"github.com/andreswebs/terminology/internal/markdown"
	"github.com/andreswebs/terminology/internal/tbx"
)

func glossary(concepts ...tbx.Concept) *tbx.Glossary {
	return &tbx.Glossary{Concepts: concepts}
}

func concept(id string, langs map[string]tbx.LangSection) tbx.Concept {
	return tbx.Concept{ID: id, Languages: langs}
}

func langSec(lang string, terms ...tbx.Term) tbx.LangSection {
	return tbx.LangSection{Lang: lang, Terms: terms}
}

func term(surface string, status tbx.Status) tbx.Term {
	return tbx.Term{Surface: surface, AdministrativeStatus: status}
}

func spans(text string) []markdown.Span {
	return []markdown.Span{{Text: text, Offset: 0, Line: 1, Col: 1}}
}

func TestScan_SingleTermMatch(t *testing.T) {
	g := glossary(concept("tzimtzum", map[string]tbx.LangSection{
		"en": langSec("en", term("tzimtzum", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	text := []byte("the concept of tzimtzum in Kabbalah")
	matches := m.Scan(text, spans("the concept of tzimtzum in Kabbalah"), 80)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	got := matches[0]
	if got.ConceptID != "tzimtzum" {
		t.Errorf("ConceptID = %q, want %q", got.ConceptID, "tzimtzum")
	}
	if got.Term != "tzimtzum" {
		t.Errorf("Term = %q, want %q", got.Term, "tzimtzum")
	}
	if got.Lang != "en" {
		t.Errorf("Lang = %q, want %q", got.Lang, "en")
	}
	if got.Status != "preferred" {
		t.Errorf("Status = %q, want %q", got.Status, "preferred")
	}
	if got.Line != 1 {
		t.Errorf("Line = %d, want 1", got.Line)
	}
	if got.Column != 16 {
		t.Errorf("Column = %d, want 16", got.Column)
	}
}

func TestScan_CaseInsensitive(t *testing.T) {
	g := glossary(concept("tzimtzum", map[string]tbx.LangSection{
		"en": langSec("en", term("Tzimtzum", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	matches := m.Scan(
		[]byte("TZIMTZUM appears here"),
		spans("TZIMTZUM appears here"),
		80,
	)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].Term != "Tzimtzum" {
		t.Errorf("Term = %q, want original surface %q", matches[0].Term, "Tzimtzum")
	}
}

func TestScan_HebrewNiqqud(t *testing.T) {
	g := glossary(concept("shalom", map[string]tbx.LangSection{
		"he": langSec("he", term("שָׁלוֹם", tbx.StatusPreferred)),
	}))

	m, err := New(g, "he", PolicyFor("he"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	matches := m.Scan(
		[]byte("the word שלום means peace"),
		spans("the word שלום means peace"),
		80,
	)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].Term != "שָׁלוֹם" {
		t.Errorf("Term = %q, want original surface", matches[0].Term)
	}
	if matches[0].Lang != "he" {
		t.Errorf("Lang = %q, want %q", matches[0].Lang, "he")
	}
}

func TestScan_MultiWordTerm(t *testing.T) {
	g := glossary(concept("tp", map[string]tbx.LangSection{
		"es": langSec("es", term("tzimtzum primordial", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	matches := m.Scan(
		[]byte("the tzimtzum primordial concept"),
		spans("the tzimtzum primordial concept"),
		80,
	)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].Term != "tzimtzum primordial" {
		t.Errorf("Term = %q, want %q", matches[0].Term, "tzimtzum primordial")
	}
}

func TestScan_StatusTagging(t *testing.T) {
	g := glossary(concept("c001", map[string]tbx.LangSection{
		"en": langSec("en",
			term("tzimtzum", tbx.StatusPreferred),
			term("contraction", tbx.StatusDeprecated),
			term("tsimtsum", tbx.StatusAdmitted),
		),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	matches := m.Scan(
		[]byte("tzimtzum and contraction and tsimtsum"),
		spans("tzimtzum and contraction and tsimtsum"),
		80,
	)

	if len(matches) != 3 {
		t.Fatalf("got %d matches, want 3", len(matches))
	}

	wantStatus := map[string]string{
		"tzimtzum":    "preferred",
		"contraction": "deprecated",
		"tsimtsum":    "admitted",
	}
	for _, got := range matches {
		want, ok := wantStatus[got.Term]
		if !ok {
			t.Errorf("unexpected term %q", got.Term)
			continue
		}
		if got.Status != want {
			t.Errorf("Term %q: Status = %q, want %q", got.Term, got.Status, want)
		}
	}
}

func TestScan_LanguageFilter(t *testing.T) {
	g := glossary(concept("c001", map[string]tbx.LangSection{
		"en": langSec("en", term("tzimtzum", tbx.StatusPreferred)),
		"he": langSec("he", term("צמצום", tbx.StatusPreferred)),
	}))

	m, err := New(g, "he", PolicyFor("he"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	matches := m.Scan(
		[]byte("tzimtzum and צמצום"),
		spans("tzimtzum and צמצום"),
		80,
	)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1 (Hebrew only)", len(matches))
	}
	if matches[0].Lang != "he" {
		t.Errorf("Lang = %q, want %q", matches[0].Lang, "he")
	}
}

func TestScan_LongestMatchWins(t *testing.T) {
	g := glossary(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("tzimtzum", tbx.StatusPreferred)),
		}),
		concept("c002", map[string]tbx.LangSection{
			"en": langSec("en", term("tzimtzum primordial", tbx.StatusPreferred)),
		}),
	)

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	matches := m.Scan(
		[]byte("the tzimtzum primordial concept"),
		spans("the tzimtzum primordial concept"),
		80,
	)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].Term != "tzimtzum primordial" {
		t.Errorf("Term = %q, want %q", matches[0].Term, "tzimtzum primordial")
	}
}

func TestScan_ContextWindow(t *testing.T) {
	g := glossary(concept("c001", map[string]tbx.LangSection{
		"en": langSec("en", term("tzimtzum", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	text := "the concept of tzimtzum in Kabbalah"
	matches := m.Scan([]byte(text), spans(text), 80)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	ctx := matches[0].Context
	if len(ctx) == 0 {
		t.Error("Context is empty")
	}
	if ctx != text {
		t.Errorf("Context = %q, want full text (short enough to include all)", ctx)
	}
}

func TestScan_ContextWindowCustomSize(t *testing.T) {
	g := glossary(concept("c001", map[string]tbx.LangSection{
		"en": langSec("en", term("tzimtzum", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	text := "aaaaaaaaaa bbbbbbbbbb tzimtzum cccccccccc dddddddddd"
	matches := m.Scan([]byte(text), spans(text), 20)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	ctx := matches[0].Context
	if ctx == text {
		t.Error("Context should be truncated with contextSize=20, but got full text")
	}
	if len(ctx) == 0 {
		t.Error("Context is empty")
	}
	// "tzimtzum" is 8 chars, half-window is 10, so context should be ~28 chars + ellipses
	if len(ctx) > 40 {
		t.Errorf("Context too long (%d chars): %q", len(ctx), ctx)
	}
}

func TestScan_ContextWindowZeroUsesDefault(t *testing.T) {
	g := glossary(concept("c001", map[string]tbx.LangSection{
		"en": langSec("en", term("tzimtzum", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	text := "the concept of tzimtzum in Kabbalah"
	matchesDefault := m.Scan([]byte(text), spans(text), 0)
	matchesExplicit := m.Scan([]byte(text), spans(text), 80)

	if len(matchesDefault) != 1 || len(matchesExplicit) != 1 {
		t.Fatalf("expected 1 match each, got %d and %d", len(matchesDefault), len(matchesExplicit))
	}
	if matchesDefault[0].Context != matchesExplicit[0].Context {
		t.Errorf("contextSize=0 should use default (80), got %q vs %q",
			matchesDefault[0].Context, matchesExplicit[0].Context)
	}
}

func TestScan_NoMatches(t *testing.T) {
	g := glossary(concept("c001", map[string]tbx.LangSection{
		"en": langSec("en", term("tzimtzum", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	matches := m.Scan(
		[]byte("no matching terms here"),
		spans("no matching terms here"),
		80,
	)

	if len(matches) != 0 {
		t.Errorf("got %d matches, want 0", len(matches))
	}
}

func TestScan_MultipleSpans(t *testing.T) {
	g := glossary(concept("c001", map[string]tbx.LangSection{
		"en": langSec("en", term("tzimtzum", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	multiSpans := []markdown.Span{
		{Text: "first paragraph mentions tzimtzum here", Offset: 0, Line: 1, Col: 1},
		{Text: "second paragraph also has tzimtzum", Offset: 50, Line: 5, Col: 1},
	}

	matches := m.Scan([]byte(""), multiSpans, 80)

	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2", len(matches))
	}
	if matches[0].Line != 1 {
		t.Errorf("first match Line = %d, want 1", matches[0].Line)
	}
	if matches[1].Line != 5 {
		t.Errorf("second match Line = %d, want 5", matches[1].Line)
	}
}

func TestScan_EmptyGlossary(t *testing.T) {
	g := glossary()

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	matches := m.Scan(
		[]byte("some text"),
		spans("some text"),
		80,
	)

	if len(matches) != 0 {
		t.Errorf("got %d matches, want 0", len(matches))
	}
}

func TestScan_SortedByPosition(t *testing.T) {
	g := glossary(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("beta", tbx.StatusPreferred)),
		}),
		concept("c002", map[string]tbx.LangSection{
			"en": langSec("en", term("alpha", tbx.StatusPreferred)),
		}),
	)

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	text := "alpha comes before beta"
	matches := m.Scan([]byte(text), spans(text), 80)

	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2", len(matches))
	}
	if matches[0].Term != "alpha" {
		t.Errorf("first match = %q, want %q", matches[0].Term, "alpha")
	}
	if matches[1].Term != "beta" {
		t.Errorf("second match = %q, want %q", matches[1].Term, "beta")
	}
}

func TestScan_UnspecifiedStatus(t *testing.T) {
	g := glossary(concept("c001", map[string]tbx.LangSection{
		"en": langSec("en", term("tzimtzum", tbx.StatusUnspecified)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	text := "the tzimtzum concept"
	matches := m.Scan([]byte(text), spans(text), 80)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].Status != "unspecified" {
		t.Errorf("Status = %q, want %q", matches[0].Status, "unspecified")
	}
}

func TestScan_HebrewTermBoundary(t *testing.T) {
	g := glossary(concept("c001", map[string]tbx.LangSection{
		"he": langSec("he", term("צמצום", tbx.StatusPreferred)),
	}))

	m, err := New(g, "he", PolicyFor("he"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	text := "la noción de צמצום aparece"
	matches := m.Scan([]byte(text), spans(text), 80)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].Term != "צמצום" {
		t.Errorf("Term = %q, want %q", matches[0].Term, "צמצום")
	}
}

func TestScan_NiqqudMatchWithoutLangFilter(t *testing.T) {
	g := glossary(concept("sefirah", map[string]tbx.LangSection{
		"he": langSec("he", term("סְפִירָה", tbx.StatusPreferred)),
		"en": langSec("en", term("sefirah", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	text := "the term סְפִירָה with niqqud and ספירה without"
	matches := m.Scan([]byte(text), spans(text), 80)

	heMatches := 0
	for _, got := range matches {
		if got.Lang == "he" {
			heMatches++
		}
	}
	if heMatches != 2 {
		t.Errorf("got %d Hebrew matches, want 2 (both niqqud and plain forms)", heMatches)
	}
}

func TestScan_MultiWordAcrossLineBreak(t *testing.T) {
	g := glossary(concept("tp", map[string]tbx.LangSection{
		"es": langSec("es", term("tzimtzum primordial", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	src := []byte("The concept of tzimtzum\nprimordial extends the basic idea.\n")
	var spanSlice []markdown.Span
	for sp := range markdown.Spans(src) {
		spanSlice = append(spanSlice, sp)
	}

	matches := m.Scan(src, spanSlice, 80)

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].Term != "tzimtzum primordial" {
		t.Errorf("Term = %q, want %q", matches[0].Term, "tzimtzum primordial")
	}
	if matches[0].Line != 1 {
		t.Errorf("Line = %d, want 1", matches[0].Line)
	}
}

func TestScan_BoundaryRejection(t *testing.T) {
	g := glossary(concept("c001", map[string]tbx.LangSection{
		"en": langSec("en", term("art", tbx.StatusPreferred)),
	}))

	m, err := New(g, "", Baseline)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	matches := m.Scan(
		[]byte("the earth is beautiful"),
		spans("the earth is beautiful"),
		80,
	)

	if len(matches) != 0 {
		t.Errorf("got %d matches, want 0 ('art' inside 'earth' should be rejected)", len(matches))
	}
}

func TestScanText_ExtractsSpansAndMatches(t *testing.T) {
	g := glossary(concept("tzimtzum", map[string]tbx.LangSection{
		"en": langSec("en", term("tzimtzum", tbx.StatusPreferred)),
	}))

	// Raw markdown: ScanText must strip the heading/emphasis syntax via
	// span extraction before matching, which Matcher.Scan alone does not do.
	doc := []byte("# Kabbalah\n\nThe concept of *tzimtzum* matters.\n")

	matches, err := ScanText(g, doc, "", 80)
	if err != nil {
		t.Fatalf("ScanText: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("ScanText found %d matches, want 1", len(matches))
	}
	if matches[0].ConceptID != "tzimtzum" {
		t.Errorf("ConceptID = %q, want %q", matches[0].ConceptID, "tzimtzum")
	}
	if matches[0].Status != "preferred" {
		t.Errorf("Status = %q, want %q", matches[0].Status, "preferred")
	}
}

func TestScanText_NoMatchesOnEmptyGlossary(t *testing.T) {
	matches, err := ScanText(glossary(), []byte("any text here"), "", 80)
	if err != nil {
		t.Fatalf("ScanText: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("ScanText found %d matches, want 0", len(matches))
	}
}
