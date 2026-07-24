package match

import (
	"testing"
)

func TestLongestMatchPerStart_Empty(t *testing.T) {
	got := longestMatchPerStart(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %d matches", len(got))
	}
}

func TestLongestMatchPerStart_SameStartKeepsLongest(t *testing.T) {
	input := []rawMatch{
		{PatternID: 1, Start: 0, End: 20},
		{PatternID: 0, Start: 0, End: 8},
	}
	got := longestMatchPerStart(input)
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got))
	}
	if got[0].PatternID != 1 || got[0].End != 20 {
		t.Errorf("expected longest match (PatternID=1, End=20), got PatternID=%d End=%d", got[0].PatternID, got[0].End)
	}
}

func TestLongestMatchPerStart_MultipleGroups(t *testing.T) {
	input := []rawMatch{
		{PatternID: 1, Start: 0, End: 10},
		{PatternID: 0, Start: 0, End: 5},
		{PatternID: 2, Start: 20, End: 25},
	}
	got := longestMatchPerStart(input)
	if len(got) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(got))
	}
	if got[0].PatternID != 1 || got[0].Start != 0 {
		t.Errorf("first match: expected PatternID=1 Start=0, got PatternID=%d Start=%d", got[0].PatternID, got[0].Start)
	}
	if got[1].PatternID != 2 || got[1].Start != 20 {
		t.Errorf("second match: expected PatternID=2 Start=20, got PatternID=%d Start=%d", got[1].PatternID, got[1].Start)
	}
}

func TestLongestMatchPerStart_SingleMatch(t *testing.T) {
	input := []rawMatch{{PatternID: 0, Start: 5, End: 12}}
	got := longestMatchPerStart(input)
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got))
	}
	if got[0] != input[0] {
		t.Errorf("expected match unchanged, got %+v", got[0])
	}
}

func TestLongestMatchPerStart_NoOverlaps(t *testing.T) {
	input := []rawMatch{
		{PatternID: 0, Start: 0, End: 5},
		{PatternID: 1, Start: 10, End: 15},
	}
	got := longestMatchPerStart(input)
	if len(got) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(got))
	}
	if got[0].Start != 0 || got[1].Start != 10 {
		t.Errorf("unexpected starts: %d, %d", got[0].Start, got[1].Start)
	}
}
