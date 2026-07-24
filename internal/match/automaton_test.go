package match

import (
	"testing"
)

func TestSearch_SinglePattern(t *testing.T) {
	a := buildAutomaton([][]byte{[]byte("hello")})
	matches := a.Search([]byte("say hello world"))

	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	m := matches[0]
	if m.PatternID != 0 {
		t.Errorf("PatternID = %d, want 0", m.PatternID)
	}
	if m.Start != 4 {
		t.Errorf("Start = %d, want 4", m.Start)
	}
	if m.End != 9 {
		t.Errorf("End = %d, want 9", m.End)
	}
}

func TestSearch_NoMatches(t *testing.T) {
	a := buildAutomaton([][]byte{[]byte("xyz")})
	matches := a.Search([]byte("hello world"))

	if len(matches) != 0 {
		t.Fatalf("got %d matches, want 0", len(matches))
	}
}

func TestSearch_EmptyPatterns(t *testing.T) {
	a := buildAutomaton(nil)
	matches := a.Search([]byte("hello world"))

	if len(matches) != 0 {
		t.Fatalf("got %d matches, want 0", len(matches))
	}
}

func TestSearch_MultipleOccurrences(t *testing.T) {
	a := buildAutomaton([][]byte{[]byte("ab")})
	matches := a.Search([]byte("ab cd ab ef ab"))

	if len(matches) != 3 {
		t.Fatalf("got %d matches, want 3", len(matches))
	}
	starts := make([]int, len(matches))
	for i, m := range matches {
		starts[i] = m.Start
	}
	want := []int{0, 6, 12}
	for i, s := range starts {
		if s != want[i] {
			t.Errorf("match[%d].Start = %d, want %d", i, s, want[i])
		}
	}
}

func TestSearch_ReusableAcrossSearches(t *testing.T) {
	a := buildAutomaton([][]byte{[]byte("cat")})

	m1 := a.Search([]byte("the cat sat"))
	if len(m1) != 1 {
		t.Fatalf("first search: got %d matches, want 1", len(m1))
	}

	m2 := a.Search([]byte("no match here"))
	if len(m2) != 0 {
		t.Fatalf("second search: got %d matches, want 0", len(m2))
	}

	m3 := a.Search([]byte("cat cat"))
	if len(m3) != 2 {
		t.Fatalf("third search: got %d matches, want 2", len(m3))
	}
}

func TestSearch_OverlappingPatterns(t *testing.T) {
	a := buildAutomaton([][]byte{[]byte("abc"), []byte("abcdef")})
	matches := a.Search([]byte("xabcdefy"))

	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2", len(matches))
	}

	if matches[0].Start != 1 || matches[0].End != 7 {
		t.Errorf("first match (longest) = [%d,%d), want [1,7)", matches[0].Start, matches[0].End)
	}
	if matches[1].Start != 1 || matches[1].End != 4 {
		t.Errorf("second match (shorter) = [%d,%d), want [1,4)", matches[1].Start, matches[1].End)
	}
}

func TestSearch_SortedByStartThenLongest(t *testing.T) {
	a := buildAutomaton([][]byte{[]byte("cd"), []byte("ab"), []byte("abcd")})
	matches := a.Search([]byte("abcd"))

	if len(matches) != 3 {
		t.Fatalf("got %d matches, want 3", len(matches))
	}
	if matches[0].Start != 0 || matches[0].End != 4 {
		t.Errorf("matches[0] = [%d,%d), want [0,4)", matches[0].Start, matches[0].End)
	}
	if matches[1].Start != 0 || matches[1].End != 2 {
		t.Errorf("matches[1] = [%d,%d), want [0,2)", matches[1].Start, matches[1].End)
	}
	if matches[2].Start != 2 || matches[2].End != 4 {
		t.Errorf("matches[2] = [%d,%d), want [2,4)", matches[2].Start, matches[2].End)
	}
}

func TestSearch_MultiplePatterns(t *testing.T) {
	a := buildAutomaton([][]byte{[]byte("hello"), []byte("world")})
	matches := a.Search([]byte("hello world"))

	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2", len(matches))
	}

	byStart := make(map[int]rawMatch)
	for _, m := range matches {
		byStart[m.Start] = m
	}

	hello := byStart[0]
	if hello.PatternID != 0 {
		t.Errorf("hello PatternID = %d, want 0", hello.PatternID)
	}
	if hello.End != 5 {
		t.Errorf("hello End = %d, want 5", hello.End)
	}

	world := byStart[6]
	if world.PatternID != 1 {
		t.Errorf("world PatternID = %d, want 1", world.PatternID)
	}
	if world.End != 11 {
		t.Errorf("world End = %d, want 11", world.End)
	}
}
