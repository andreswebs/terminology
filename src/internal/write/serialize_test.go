package write

import (
	"encoding/json"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
	_ "github.com/andreswebs/terminology/internal/tbx/linguist"
)

// Cycle 1 — exported serializer carries rich fields.
func TestConceptToWriteResult_RichFields(t *testing.T) {
	c := tbx.Concept{
		ID:           "tzimtzum",
		SubjectField: "kabbalah",
		Definitions:  []tbx.NoteText{{Plain: "concept-level def"}},
		Languages: map[string]tbx.LangSection{
			"en": {
				Lang: "en",
				Terms: []tbx.Term{
					{
						Surface:              "tzimtzum",
						AdministrativeStatus: tbx.StatusPreferred,
						Reading:              "tsimtsum",
						ReadingNote:          "Ashkenazi",
						Contexts:             []tbx.NoteText{{Plain: "some context"}},
						Notes:                []string{"a note"},
					},
				},
			},
		},
	}

	r := ConceptToWriteResult(c)

	if r.ConceptID != "tzimtzum" {
		t.Errorf("ConceptID = %q, want %q", r.ConceptID, "tzimtzum")
	}
	if r.SubjectField != "kabbalah" {
		t.Errorf("SubjectField = %q, want %q", r.SubjectField, "kabbalah")
	}
	if len(r.Definitions) != 1 || r.Definitions[0] != "concept-level def" {
		t.Errorf("Definitions = %v, want [concept-level def]", r.Definitions)
	}
	grp, ok := r.Languages["en"]
	if !ok || grp.Preferred == nil {
		t.Fatalf("expected en preferred term")
	}
	if grp.Preferred.Reading != "tsimtsum" || grp.Preferred.ReadingNote != "Ashkenazi" {
		t.Errorf("reading/reading_note not carried: %+v", grp.Preferred)
	}
	if len(grp.Preferred.Contexts) != 1 || grp.Preferred.Contexts[0] != "some context" {
		t.Errorf("contexts not carried: %v", grp.Preferred.Contexts)
	}
	if len(grp.Preferred.Notes) != 1 || grp.Preferred.Notes[0] != "a note" {
		t.Errorf("notes not carried: %v", grp.Preferred.Notes)
	}
}

// Cycle 2 — per-language definition accepted on input (FEAT-1).
func TestParseJSONInput_PerLanguageDefinition(t *testing.T) {
	input := `{
		"concept_id": "x",
		"languages": {
			"en": {"preferred": {"term": "foo"}, "definitions": ["EN def"]}
		}
	}`

	wr, err := ParseJSONInput([]byte(input))
	if err != nil {
		t.Fatalf("ParseJSONInput: %v", err)
	}
	c := ResultToConcept(wr)
	ls, ok := c.Languages["en"]
	if !ok {
		t.Fatalf("expected en language section")
	}
	if len(ls.Definitions) != 1 || ls.Definitions[0].Plain != "EN def" {
		t.Errorf("Languages[en].Definitions = %v, want [EN def]", ls.Definitions)
	}
}

// Cycle 3 — per-language definition emitted on output.
func TestConceptToWriteResult_PerLanguageDefinition(t *testing.T) {
	c := tbx.Concept{
		ID: "x",
		Languages: map[string]tbx.LangSection{
			"en": {
				Lang:        "en",
				Definitions: []tbx.NoteText{{Plain: "EN def"}},
				Terms:       []tbx.Term{{Surface: "foo", AdministrativeStatus: tbx.StatusPreferred}},
			},
		},
	}

	r := ConceptToWriteResult(c)
	grp := r.Languages["en"]
	if len(grp.Definitions) != 1 || grp.Definitions[0] != "EN def" {
		t.Errorf("Languages[en].Definitions = %v, want [EN def]", grp.Definitions)
	}
}

// Cycle 4 — full round-trip lossless for feedback-relevant fields.
func TestRoundTrip_ConceptToWriteResult_Lossless(t *testing.T) {
	c := tbx.Concept{
		ID:           "tzimtzum",
		SubjectField: "kabbalah",
		Definitions:  []tbx.NoteText{{Plain: "concept-level def"}},
		CrossRefs:    []tbx.CrossRef{{Target: "sefirot", Label: "see also"}},
		Languages: map[string]tbx.LangSection{
			"en": {
				Lang:        "en",
				Definitions: []tbx.NoteText{{Plain: "EN def"}},
				Terms: []tbx.Term{
					{
						Surface:              "tzimtzum",
						AdministrativeStatus: tbx.StatusPreferred,
						Reading:              "tsimtsum",
						ReadingNote:          "note",
						Contexts:             []tbx.NoteText{{Plain: "ctx"}},
						Notes:                []string{"term note"},
					},
					{
						Surface:              "contraction",
						AdministrativeStatus: tbx.StatusAdmitted,
					},
				},
			},
		},
	}

	rt := ResultToConcept(ptrTo(ConceptToWriteResult(c)))
	eq, err := ConceptsEqual(&c, &rt)
	if err != nil {
		t.Fatalf("ConceptsEqual: %v", err)
	}
	if !eq {
		t.Errorf("round-trip not lossless\noriginal: %+v\nround-trip: %+v", c, rt)
	}
}

// Cycle 5 — apply consumes serializer output (bilingual definitions).
func TestApply_ConsumesSerializer_BilingualDefinitions(t *testing.T) {
	c := tbx.Concept{
		ID: "tzimtzum",
		Languages: map[string]tbx.LangSection{
			"en": {
				Lang:        "en",
				Definitions: []tbx.NoteText{{Plain: "EN def"}},
				Terms:       []tbx.Term{{Surface: "tzimtzum", AdministrativeStatus: tbx.StatusPreferred}},
			},
			"pt": {
				Lang:        "pt",
				Definitions: []tbx.NoteText{{Plain: "PT def"}},
				Terms:       []tbx.Term{{Surface: "tzimtzum", AdministrativeStatus: tbx.StatusPreferred}},
			},
		},
	}

	data, err := json.Marshal(map[string]any{
		"concepts": []any{ConceptToWriteResult(c)},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	concepts, err := ParseApplyJSON(data)
	if err != nil {
		t.Fatalf("ParseApplyJSON: %v", err)
	}
	if len(concepts) != 1 {
		t.Fatalf("got %d concepts, want 1", len(concepts))
	}
	for _, lang := range []string{"en", "pt"} {
		ls, ok := concepts[0].Languages[lang]
		if !ok {
			t.Fatalf("missing %s language section", lang)
		}
		if len(ls.Definitions) != 1 {
			t.Errorf("%s definitions = %v, want one", lang, ls.Definitions)
		}
	}
}

func ptrTo[T any](v T) *T { return &v }
