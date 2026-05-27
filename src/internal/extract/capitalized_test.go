package extract

import (
	"testing"
)

func TestCapitalizedPhrases_SingleWordMidSentence(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the Holy Temple was destroyed", Line: 1, Col: 1, Offset: 0},
	}

	candidates := CapitalizedPhrases(spans, "en")

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Term != "Holy Temple" {
		t.Errorf("term = %q, want %q", candidates[0].Term, "Holy Temple")
	}
	if candidates[0].Frequency != 1 {
		t.Errorf("frequency = %d, want 1", candidates[0].Frequency)
	}
	if candidates[0].Heuristic != "capitalized_phrase" {
		t.Errorf("heuristic = %q, want %q", candidates[0].Heuristic, "capitalized_phrase")
	}
}

func TestCapitalizedPhrases_SentenceStartExcluded(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "The cat sat. The dog ran.", Line: 1, Col: 1, Offset: 0},
	}

	candidates := CapitalizedPhrases(spans, "en")

	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0 (sentence-start words should be excluded)", len(candidates))
		for _, c := range candidates {
			t.Logf("  unexpected: %q", c.Term)
		}
	}
}

func TestCapitalizedPhrases_MultiWord(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the Dead Sea Scrolls were found", Line: 1, Col: 1, Offset: 0},
	}

	candidates := CapitalizedPhrases(spans, "en")

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Term != "Dead Sea Scrolls" {
		t.Errorf("term = %q, want %q", candidates[0].Term, "Dead Sea Scrolls")
	}
}

func TestCapitalizedPhrases_FrequencyAggregation(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the Holy Temple was great", Line: 1, Col: 1, Offset: 0},
		{Text: "we visited the Holy Temple again", Line: 2, Col: 1, Offset: 30},
		{Text: "the Holy Temple is ancient", Line: 3, Col: 1, Offset: 60},
	}

	candidates := CapitalizedPhrases(spans, "en")

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Frequency != 3 {
		t.Errorf("frequency = %d, want 3", candidates[0].Frequency)
	}
	if len(candidates[0].Locations) != 3 {
		t.Errorf("locations count = %d, want 3", len(candidates[0].Locations))
	}
}

func TestCapitalizedPhrases_NFCNormalization(t *testing.T) {
	t.Parallel()

	// "é" in NFC (U+00E9) vs NFD (U+0065 U+0301)
	nfc := "the René was here"
	nfd := "the René was here"

	spans := []Span{
		{Text: nfc, Line: 1, Col: 1, Offset: 0},
		{Text: nfd, Line: 2, Col: 1, Offset: 30},
	}

	candidates := CapitalizedPhrases(spans, "en")

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1 (NFC/NFD should aggregate)", len(candidates))
	}
	if candidates[0].Frequency != 2 {
		t.Errorf("frequency = %d, want 2", candidates[0].Frequency)
	}
}

func TestCapitalizedPhrases_ParagraphStartExcluded(t *testing.T) {
	t.Parallel()

	// Each span is a new paragraph; first word should be treated as sentence start
	spans := []Span{
		{Text: "Hello world", Line: 1, Col: 1, Offset: 0},
		{Text: "Goodbye world", Line: 3, Col: 1, Offset: 20},
	}

	candidates := CapitalizedPhrases(spans, "en")

	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0 (paragraph-start words should be excluded)", len(candidates))
		for _, c := range candidates {
			t.Logf("  unexpected: %q", c.Term)
		}
	}
}

func TestCapitalizedPhrases_LocationTracking(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the Holy Temple stands", Line: 5, Col: 3, Offset: 100},
	}

	candidates := CapitalizedPhrases(spans, "en")

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}

	loc := candidates[0].Locations[0]
	if loc.Line != 5 {
		t.Errorf("line = %d, want 5", loc.Line)
	}
	// "the " = 4 bytes, so "Holy" starts at col offset 4, plus span col 3
	wantCol := 3 + 4
	if loc.Col != wantCol {
		t.Errorf("col = %d, want %d", loc.Col, wantCol)
	}
	wantOffset := 100 + 4
	if loc.Offset != wantOffset {
		t.Errorf("offset = %d, want %d", loc.Offset, wantOffset)
	}
}

func TestCapitalizedPhrases_EmptyInput(t *testing.T) {
	t.Parallel()

	candidates := CapitalizedPhrases(nil, "en")
	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0 for nil input", len(candidates))
	}

	candidates = CapitalizedPhrases([]Span{}, "en")
	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0 for empty input", len(candidates))
	}
}

func TestCapitalizedPhrases_AllLowercase(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "no capitalized words here at all", Line: 1, Col: 1, Offset: 0},
	}

	candidates := CapitalizedPhrases(spans, "en")

	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0", len(candidates))
	}
}

func TestCapitalizedPhrases_MixedSentences(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "We saw the Holy Temple. Then we left.", Line: 1, Col: 1, Offset: 0},
	}

	candidates := CapitalizedPhrases(spans, "en")

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Term != "Holy Temple" {
		t.Errorf("term = %q, want %q", candidates[0].Term, "Holy Temple")
	}
}

func TestCapitalizedPhrases_CapitalizedAfterExclamation(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "look! Capital here", Line: 1, Col: 1, Offset: 0},
	}

	candidates := CapitalizedPhrases(spans, "en")

	// "Capital" follows "!" — should be treated as sentence start, excluded
	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0 (word after ! is sentence start)", len(candidates))
		for _, c := range candidates {
			t.Logf("  unexpected: %q", c.Term)
		}
	}
}

func TestCapitalizedPhrases_MultipleDistinctPhrases(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the Dead Sea Scrolls and the Holy Temple were found", Line: 1, Col: 1, Offset: 0},
	}

	candidates := CapitalizedPhrases(spans, "en")

	if len(candidates) != 2 {
		t.Fatalf("got %d candidates, want 2", len(candidates))
	}

	found := make(map[string]bool)
	for _, c := range candidates {
		found[c.Term] = true
	}
	if !found["Dead Sea Scrolls"] {
		t.Error("missing candidate: Dead Sea Scrolls")
	}
	if !found["Holy Temple"] {
		t.Error("missing candidate: Holy Temple")
	}
}
