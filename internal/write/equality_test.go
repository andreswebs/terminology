package write

import (
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
	_ "github.com/andreswebs/terminology/internal/tbx/linguist"
)

func baseConcept() tbx.Concept {
	return tbx.Concept{
		ID:           "tzimtzum",
		SubjectField: "kabbalah",
		Languages: map[string]tbx.LangSection{
			"en": {
				Lang: "en",
				Terms: []tbx.Term{
					{Surface: "tzimtzum", AdministrativeStatus: tbx.StatusPreferred},
				},
			},
			"he": {
				Lang: "he",
				Terms: []tbx.Term{
					{Surface: "צמצום", AdministrativeStatus: tbx.StatusPreferred},
				},
			},
		},
	}
}

func TestConceptsEqual_IdenticalConcepts(t *testing.T) {
	a := baseConcept()
	b := baseConcept()

	equal, err := ConceptsEqual(&a, &b)
	if err != nil {
		t.Fatalf("ConceptsEqual: %v", err)
	}
	if !equal {
		t.Error("expected identical concepts to be equal")
	}
}

func TestConceptsEqual_DifferingTransactions_ConceptLevel(t *testing.T) {
	a := baseConcept()
	a.Transactions = []tbx.Transaction{
		{Type: "origination", Date: "2025-01-01T00:00:00Z", Responsibility: "alice"},
	}

	b := baseConcept()
	b.Transactions = []tbx.Transaction{
		{Type: "modification", Date: "2025-06-01T00:00:00Z", Responsibility: "bob"},
		{Type: "origination", Date: "2025-01-01T00:00:00Z", Responsibility: "alice"},
	}

	equal, err := ConceptsEqual(&a, &b)
	if err != nil {
		t.Fatalf("ConceptsEqual: %v", err)
	}
	if !equal {
		t.Error("expected concepts differing only in concept-level transactions to be equal")
	}
}

func TestConceptsEqual_DifferingContent(t *testing.T) {
	a := baseConcept()

	b := baseConcept()
	b.SubjectField = "mysticism"

	equal, err := ConceptsEqual(&a, &b)
	if err != nil {
		t.Fatalf("ConceptsEqual: %v", err)
	}
	if equal {
		t.Error("expected concepts with different subject fields to not be equal")
	}
}

func TestConceptsEqual_DifferingTerms(t *testing.T) {
	a := baseConcept()

	b := baseConcept()
	enSec := b.Languages["en"]
	enSec.Terms = append(enSec.Terms, tbx.Term{
		Surface:              "contraction",
		AdministrativeStatus: tbx.StatusDeprecated,
	})
	b.Languages["en"] = enSec

	equal, err := ConceptsEqual(&a, &b)
	if err != nil {
		t.Fatalf("ConceptsEqual: %v", err)
	}
	if equal {
		t.Error("expected concepts with different terms to not be equal")
	}
}

func TestConceptsEqual_EmptyConcepts(t *testing.T) {
	a := tbx.Concept{ID: "empty"}
	b := tbx.Concept{ID: "empty"}

	equal, err := ConceptsEqual(&a, &b)
	if err != nil {
		t.Fatalf("ConceptsEqual: %v", err)
	}
	if !equal {
		t.Error("expected empty concepts to be equal")
	}
}

func TestConceptsEqual_FieldOrderingIrrelevant(t *testing.T) {
	a := baseConcept()
	a.Languages = map[string]tbx.LangSection{
		"he": {
			Lang: "he",
			Terms: []tbx.Term{
				{Surface: "צמצום", AdministrativeStatus: tbx.StatusPreferred},
			},
		},
		"en": {
			Lang: "en",
			Terms: []tbx.Term{
				{Surface: "tzimtzum", AdministrativeStatus: tbx.StatusPreferred},
			},
		},
	}

	b := baseConcept()

	equal, err := ConceptsEqual(&a, &b)
	if err != nil {
		t.Fatalf("ConceptsEqual: %v", err)
	}
	if !equal {
		t.Error("expected concepts with different language map insertion order to be equal (canonical writer sorts)")
	}
}

func TestConceptsEqual_TransactionsAtAllLevels(t *testing.T) {
	a := tbx.Concept{
		ID:           "sefirot",
		SubjectField: "kabbalah",
		Definitions:  []tbx.NoteText{{Plain: "The ten emanations"}},
		Transactions: []tbx.Transaction{
			{Type: "origination", Date: "2025-01-01T00:00:00Z", Responsibility: "alice"},
		},
		Languages: map[string]tbx.LangSection{
			"en": {
				Lang: "en",
				Terms: []tbx.Term{
					{
						Surface:              "sefirot",
						AdministrativeStatus: tbx.StatusPreferred,
						Transactions: []tbx.Transaction{
							{Type: "origination", Date: "2025-01-01T00:00:00Z"},
							{Type: "modification", Date: "2025-06-01T00:00:00Z"},
						},
					},
				},
			},
		},
	}

	b := tbx.Concept{
		ID:           "sefirot",
		SubjectField: "kabbalah",
		Definitions:  []tbx.NoteText{{Plain: "The ten emanations"}},
		Languages: map[string]tbx.LangSection{
			"en": {
				Lang: "en",
				Terms: []tbx.Term{
					{
						Surface:              "sefirot",
						AdministrativeStatus: tbx.StatusPreferred,
					},
				},
			},
		},
	}

	equal, err := ConceptsEqual(&a, &b)
	if err != nil {
		t.Fatalf("ConceptsEqual: %v", err)
	}
	if !equal {
		t.Error("expected concepts differing only in transactions (concept + term level) to be equal")
	}
}

func TestConceptsEqual_DifferingTransactions_TermLevel(t *testing.T) {
	a := baseConcept()
	enSec := a.Languages["en"]
	enSec.Terms[0].Transactions = []tbx.Transaction{
		{Type: "origination", Date: "2025-01-01T00:00:00Z"},
	}
	a.Languages["en"] = enSec

	b := baseConcept()

	equal, err := ConceptsEqual(&a, &b)
	if err != nil {
		t.Fatalf("ConceptsEqual: %v", err)
	}
	if !equal {
		t.Error("expected concepts differing only in term-level transactions to be equal")
	}
}
