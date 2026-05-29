package write

import (
	"errors"
	"testing"

	"github.com/andreswebs/terminology/internal/terr"
)

func TestConceptIndex_Found(t *testing.T) {
	g := makeGlossary(makeConcept("alpha", "en", "a"), makeConcept("bravo", "en", "b"))

	idx, err := ConceptIndex(g, "bravo")
	if err != nil {
		t.Fatalf("ConceptIndex(bravo) error = %v, want nil", err)
	}
	if idx != 1 {
		t.Errorf("ConceptIndex(bravo) = %d, want 1", idx)
	}
}

func TestConceptIndex_NotFound(t *testing.T) {
	g := makeGlossary(makeConcept("alpha", "en", "a"))

	idx, err := ConceptIndex(g, "missing")
	if idx != -1 {
		t.Errorf("ConceptIndex(missing) idx = %d, want -1", idx)
	}
	var coded terr.Coded
	if !errors.As(err, &coded) || coded.Code() != "not_found" {
		t.Errorf("ConceptIndex(missing) error = %v, want code %q", err, "not_found")
	}
}
