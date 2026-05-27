package extract

import (
	"testing"
)

func TestForeignScriptTokens_ScriptFilter(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the concept of צמצום and Ψυχή in philosophy", Line: 1, Col: 1, Offset: 0},
	}

	candidates := ForeignScriptTokens(spans, Options{Script: "hebrew"})

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1 (only Hebrew)", len(candidates))
	}
	if candidates[0].Term != "צמצום" {
		t.Errorf("term = %q, want %q", candidates[0].Term, "צמצום")
	}
}

func TestForeignScriptTokens_ScriptFilterAny(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the concept of צמצום and Ψυχή in philosophy", Line: 1, Col: 1, Offset: 0},
	}

	candidates := ForeignScriptTokens(spans, Options{Script: "any"})

	if len(candidates) != 2 {
		t.Fatalf("got %d candidates, want 2", len(candidates))
	}
}

func TestForeignScriptTokens_CommonInheritedIgnored(t *testing.T) {
	t.Parallel()

	// Digits and punctuation are Common — they should not affect script detection
	spans := []Span{
		{Text: "the year 2024 and price $100 are not foreign", Line: 1, Col: 1, Offset: 0},
	}

	candidates := ForeignScriptTokens(spans, Options{})

	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0 (digits/punctuation are Common script)", len(candidates))
		for _, c := range candidates {
			t.Logf("  unexpected: %q", c.Term)
		}
	}
}

func TestForeignScriptTokens_FrequencyAggregation(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the concept of צמצום is important", Line: 1, Col: 1, Offset: 0},
		{Text: "another mention of צמצום here", Line: 2, Col: 1, Offset: 40},
	}

	candidates := ForeignScriptTokens(spans, Options{})

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Frequency != 2 {
		t.Errorf("frequency = %d, want 2", candidates[0].Frequency)
	}
	if len(candidates[0].Locations) != 2 {
		t.Errorf("locations = %d, want 2", len(candidates[0].Locations))
	}
}

func TestForeignScriptTokens_EmptyInput(t *testing.T) {
	t.Parallel()

	candidates := ForeignScriptTokens(nil, Options{})
	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0 for nil input", len(candidates))
	}

	candidates = ForeignScriptTokens([]Span{}, Options{})
	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0 for empty input", len(candidates))
	}
}

func TestForeignScriptTokens_AllSameScript(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "only latin words here nothing foreign", Line: 1, Col: 1, Offset: 0},
	}

	candidates := ForeignScriptTokens(spans, Options{})

	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0", len(candidates))
	}
}

func TestForeignScriptTokens_LocationTracking(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the concept of צמצום is key", Line: 5, Col: 3, Offset: 100},
	}

	candidates := ForeignScriptTokens(spans, Options{})

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}

	loc := candidates[0].Locations[0]
	if loc.Line != 5 {
		t.Errorf("line = %d, want 5", loc.Line)
	}
}

func TestForeignScriptTokens_LatinInHebrew(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "המושג של API בתכנות", Line: 1, Col: 1, Offset: 0},
	}

	candidates := ForeignScriptTokens(spans, Options{})

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Term != "API" {
		t.Errorf("term = %q, want %q", candidates[0].Term, "API")
	}
}

func TestForeignScriptTokens_BaseLangHebrew(t *testing.T) {
	t.Parallel()

	// When BaseLang is "he", Hebrew is the base script.
	// Latin tokens should be flagged as foreign, not Hebrew ones.
	spans := []Span{
		{Text: "הצמצום הוא concept מרכזי", Line: 1, Col: 1, Offset: 0},
		{Text: "The Kabbalah teaches about divine light", Line: 3, Col: 1, Offset: 50},
	}

	candidates := ForeignScriptTokens(spans, Options{BaseLang: "he"})

	terms := make(map[string]bool)
	for _, c := range candidates {
		terms[c.Term] = true
	}

	// Latin tokens should be foreign
	for _, want := range []string{"concept", "The", "Kabbalah", "teaches", "about", "divine", "light"} {
		if !terms[want] {
			t.Errorf("expected Latin token %q to be flagged as foreign_script", want)
		}
	}
	// Hebrew tokens should NOT be foreign
	for _, notWant := range []string{"הצמצום", "הוא", "מרכזי"} {
		if terms[notWant] {
			t.Errorf("Hebrew token %q should not be flagged as foreign when BaseLang=he", notWant)
		}
	}
}

func TestForeignScriptTokens_BaseLangSpanish(t *testing.T) {
	t.Parallel()

	// When BaseLang is "es" (Latin script), Hebrew tokens should be foreign.
	spans := []Span{
		{Text: "El concepto de צמצום es central", Line: 1, Col: 1, Offset: 0},
	}

	candidates := ForeignScriptTokens(spans, Options{BaseLang: "es"})

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Term != "צמצום" {
		t.Errorf("term = %q, want %q", candidates[0].Term, "צמצום")
	}
}

func TestForeignScriptTokens_BaseLangEmpty_FallsBackToDominant(t *testing.T) {
	t.Parallel()

	// Without BaseLang, existing per-span dominant-script behavior is preserved.
	spans := []Span{
		{Text: "the concept of צמצום in Kabbalah", Line: 1, Col: 1, Offset: 0},
	}

	candidates := ForeignScriptTokens(spans, Options{})

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Term != "צמצום" {
		t.Errorf("term = %q, want %q", candidates[0].Term, "צמצום")
	}
}

func TestForeignScriptTokens_HebrewInLatin(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the concept of צמצום in Kabbalah", Line: 1, Col: 1, Offset: 0},
	}

	candidates := ForeignScriptTokens(spans, Options{})

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Term != "צמצום" {
		t.Errorf("term = %q, want %q", candidates[0].Term, "צמצום")
	}
	if candidates[0].Heuristic != "foreign_script" {
		t.Errorf("heuristic = %q, want %q", candidates[0].Heuristic, "foreign_script")
	}
	if candidates[0].Frequency != 1 {
		t.Errorf("frequency = %d, want 1", candidates[0].Frequency)
	}
}
