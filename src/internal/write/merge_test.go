package write

import (
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
)

func TestReplaceConcept_OverwritesButKeepsID(t *testing.T) {
	existing := &tbx.Concept{ID: "keep-me", SubjectField: "old"}
	payload := &tbx.Concept{ID: "ignored", SubjectField: "new"}

	ReplaceConcept(existing, payload)

	if existing.ID != "keep-me" {
		t.Errorf("ID = %q, want %q (replace must preserve id)", existing.ID, "keep-me")
	}
	if existing.SubjectField != "new" {
		t.Errorf("SubjectField = %q, want %q", existing.SubjectField, "new")
	}
}

func TestMergeConcept_PopulatedFieldsOverlayEmptyFieldsPreserve(t *testing.T) {
	existing := &tbx.Concept{
		ID:           "c1",
		SubjectField: "law",
		Sources:      []string{"orig"},
	}
	payload := &tbx.Concept{
		SubjectField: "", // empty: must preserve existing
		Sources:      []string{"replacement"},
	}

	MergeConcept(existing, payload)

	if existing.SubjectField != "law" {
		t.Errorf("SubjectField = %q, want %q (empty payload preserves)", existing.SubjectField, "law")
	}
	if len(existing.Sources) != 1 || existing.Sources[0] != "replacement" {
		t.Errorf("Sources = %v, want [replacement] (non-empty payload replaces)", existing.Sources)
	}
}

func TestMergeConcept_MergesMatchingTermAndAppendsNew(t *testing.T) {
	existing := &tbx.Concept{
		ID: "c1",
		Languages: map[string]tbx.LangSection{
			"en": {Lang: "en", Terms: []tbx.Term{
				{Surface: "completeness", AdministrativeStatus: tbx.StatusPreferred, PartOfSpeech: "noun"},
			}},
		},
	}
	payload := &tbx.Concept{
		Languages: map[string]tbx.LangSection{
			"en": {Lang: "en", Terms: []tbx.Term{
				// same surface+status: field-merged
				{Surface: "completeness", AdministrativeStatus: tbx.StatusPreferred, Register: "technicalRegister"},
				// new surface: appended
				{Surface: "wholeness", AdministrativeStatus: tbx.StatusAdmitted},
			}},
		},
	}

	MergeConcept(existing, payload)

	terms := existing.Languages["en"].Terms
	if len(terms) != 2 {
		t.Fatalf("got %d terms, want 2", len(terms))
	}
	if terms[0].PartOfSpeech != "noun" {
		t.Errorf("merged term PartOfSpeech = %q, want %q (preserved)", terms[0].PartOfSpeech, "noun")
	}
	if terms[0].Register != "technicalRegister" {
		t.Errorf("merged term Register = %q, want %q (overlaid)", terms[0].Register, "technicalRegister")
	}
	if terms[1].Surface != "wholeness" {
		t.Errorf("appended term = %q, want %q", terms[1].Surface, "wholeness")
	}
}

func TestMergeConcept_AddsNewLanguage(t *testing.T) {
	existing := &tbx.Concept{ID: "c1", Languages: map[string]tbx.LangSection{
		"en": {Lang: "en", Terms: []tbx.Term{{Surface: "completeness"}}},
	}}
	payload := &tbx.Concept{Languages: map[string]tbx.LangSection{
		"he": {Lang: "he", Terms: []tbx.Term{{Surface: "שלמות"}}},
	}}

	MergeConcept(existing, payload)

	if _, ok := existing.Languages["en"]; !ok {
		t.Error("existing language en dropped")
	}
	if _, ok := existing.Languages["he"]; !ok {
		t.Error("new language he not added")
	}
}
