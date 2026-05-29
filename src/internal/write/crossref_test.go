package write

import (
	"slices"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
)

func TestCrossRefsTo(t *testing.T) {
	conceptLevel := tbx.Concept{
		ID:        "refs-via-concept",
		CrossRefs: []tbx.CrossRef{{Target: "target"}},
	}
	termLevel := tbx.Concept{
		ID: "refs-via-term",
		Languages: map[string]tbx.LangSection{
			"en": {Lang: "en", Terms: []tbx.Term{
				{Surface: "x", CrossRefs: []tbx.CrossRef{{Target: "target"}}},
			}},
		},
	}
	unrelated := tbx.Concept{ID: "unrelated", CrossRefs: []tbx.CrossRef{{Target: "elsewhere"}}}
	// self-reference must be excluded
	target := tbx.Concept{ID: "target", CrossRefs: []tbx.CrossRef{{Target: "target"}}}

	g := makeGlossary(conceptLevel, termLevel, unrelated, target)

	got := CrossRefsTo(g, "target")
	slices.Sort(got)
	want := []string{"refs-via-concept", "refs-via-term"}
	if !slices.Equal(got, want) {
		t.Errorf("CrossRefsTo(target) = %v, want %v", got, want)
	}
}

func TestCrossRefsTo_None(t *testing.T) {
	g := makeGlossary(makeConcept("a", "en", "alpha"))
	if got := CrossRefsTo(g, "a"); len(got) != 0 {
		t.Errorf("CrossRefsTo(a) = %v, want empty", got)
	}
}
