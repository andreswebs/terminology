package extract

import (
	"testing"
)

func TestHighFrequencyTokens_NFCNormalization(t *testing.T) {
	t.Parallel()

	// U+00E9 (precomposed) and U+0065 U+0301 (decomposed) should aggregate
	spans := []Span{
		{Text: "café café café", Line: 1, Col: 1, Offset: 0},
	}

	candidates := HighFrequencyTokens(spans, Options{MinFreq: 3})

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Frequency != 3 {
		t.Errorf("frequency = %d, want 3", candidates[0].Frequency)
	}
}

func TestHighFrequencyTokens_EmptyInput(t *testing.T) {
	t.Parallel()

	candidates := HighFrequencyTokens(nil, Options{MinFreq: 3})
	if len(candidates) != 0 {
		t.Errorf("got %d candidates from nil input, want 0", len(candidates))
	}

	candidates = HighFrequencyTokens([]Span{}, Options{MinFreq: 3})
	if len(candidates) != 0 {
		t.Errorf("got %d candidates from empty input, want 0", len(candidates))
	}
}

func TestHighFrequencyTokens_MultiSpanAggregation(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the temple stands", Line: 1, Col: 1, Offset: 0},
		{Text: "the temple rises", Line: 2, Col: 1, Offset: 20},
		{Text: "the temple endures", Line: 3, Col: 1, Offset: 40},
	}

	candidates := HighFrequencyTokens(spans, Options{MinFreq: 3})

	found := false
	for _, c := range candidates {
		if c.Term == "temple" {
			found = true
			if c.Frequency != 3 {
				t.Errorf("frequency = %d, want 3", c.Frequency)
			}
			if len(c.Locations) != 3 {
				t.Errorf("locations count = %d, want 3", len(c.Locations))
			}
		}
	}
	if !found {
		t.Errorf("expected candidate 'temple'")
	}
}

func TestHighFrequencyTokens_DefaultMinFreq(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "temple temple temple", Line: 1, Col: 1, Offset: 0},
	}

	candidates := HighFrequencyTokens(spans, Options{})

	found := false
	for _, c := range candidates {
		if c.Term == "temple" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'temple' with default MinFreq of 3")
	}
}

func TestHighFrequencyTokens_StopwordsExclusion(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the the the temple temple temple", Line: 1, Col: 1, Offset: 0},
	}

	stopwords := map[string]bool{"the": true}
	candidates := HighFrequencyTokens(spans, Options{MinFreq: 3, Stopwords: stopwords})

	for _, c := range candidates {
		if c.Term == "the" {
			t.Errorf("'the' should be excluded by stopwords")
		}
	}
	found := false
	for _, c := range candidates {
		if c.Term == "temple" {
			found = true
			if c.Frequency != 3 {
				t.Errorf("frequency of 'temple' = %d, want 3", c.Frequency)
			}
		}
	}
	if !found {
		t.Errorf("expected candidate 'temple' to remain")
	}
}

func TestHighFrequencyTokens_CaseFoldAggregation(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "Temple temple TEMPLE", Line: 1, Col: 1, Offset: 0},
	}

	candidates := HighFrequencyTokens(spans, Options{MinFreq: 3})

	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Frequency != 3 {
		t.Errorf("frequency = %d, want 3", candidates[0].Frequency)
	}
	if candidates[0].Term != "temple" {
		t.Errorf("term = %q, want %q", candidates[0].Term, "temple")
	}
}

func TestHighFrequencyTokens_BelowThresholdExcluded(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "the temple is rare and temple is common", Line: 1, Col: 1, Offset: 0},
	}

	candidates := HighFrequencyTokens(spans, Options{MinFreq: 3})

	for _, c := range candidates {
		if c.Term == "temple" {
			t.Errorf("'temple' (frequency 2) should not appear with MinFreq 3")
		}
		if c.Term == "rare" {
			t.Errorf("'rare' (frequency 1) should not appear with MinFreq 3")
		}
	}
}

func TestHighFrequencyTokens_BasicFrequency(t *testing.T) {
	t.Parallel()

	spans := []Span{
		{Text: "temple is a great temple and the temple stands temple ahead", Line: 1, Col: 1, Offset: 0},
	}

	candidates := HighFrequencyTokens(spans, Options{MinFreq: 3})

	found := false
	for _, c := range candidates {
		if c.Term == "temple" {
			if c.Frequency != 4 {
				t.Errorf("frequency of 'temple' = %d, want 4", c.Frequency)
			}
			if c.Heuristic != "high_frequency" {
				t.Errorf("heuristic = %q, want %q", c.Heuristic, "high_frequency")
			}
			found = true
		}
	}
	if !found {
		t.Errorf("expected candidate 'temple' with frequency >= 3")
	}
}
