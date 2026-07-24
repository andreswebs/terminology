package output_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/output"
)

type testEnvelope struct {
	SchemaVersion int    `json:"schema_version"`
	OK            bool   `json:"ok"`
	ConceptID     string `json:"concept_id"`
	SubjectField  string `json:"subject_field"`
}

func TestValidateFields_KnownPaths(t *testing.T) {
	paths, err := output.ValidateFields("concept_id,subject_field", testEnvelope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("got %d paths, want 2", len(paths))
	}
	if paths[0] != "concept_id" {
		t.Errorf("paths[0] = %q, want %q", paths[0], "concept_id")
	}
	if paths[1] != "subject_field" {
		t.Errorf("paths[1] = %q, want %q", paths[1], "subject_field")
	}
}

type testTerm struct {
	Term string `json:"term"`
}

type testTermGroup struct {
	Preferred *testTerm  `json:"preferred,omitempty"`
	Admitted  []testTerm `json:"admitted,omitempty"`
}

type testNestedEnvelope struct {
	SchemaVersion int                      `json:"schema_version"`
	OK            bool                     `json:"ok"`
	ConceptID     string                   `json:"concept_id"`
	Languages     map[string]testTermGroup `json:"languages"`
}

func TestValidateFields_WildcardPaths(t *testing.T) {
	paths, err := output.ValidateFields("languages.*.preferred.term", testNestedEnvelope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(paths))
	}
	if paths[0] != "languages.*.preferred.term" {
		t.Errorf("paths[0] = %q, want %q", paths[0], "languages.*.preferred.term")
	}
}

func TestProjectFields_FiltersTopLevelFields(t *testing.T) {
	input := `{"schema_version":1,"ok":true,"concept_id":"tzimtzum","subject_field":"kabbalah"}`
	got, err := output.ProjectFields([]byte(input), []string{"concept_id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if m["schema_version"] != float64(1) {
		t.Errorf("schema_version should be preserved, got %v", m["schema_version"])
	}
	if m["ok"] != true {
		t.Errorf("ok should be preserved, got %v", m["ok"])
	}
	if m["concept_id"] != "tzimtzum" {
		t.Errorf("concept_id = %v, want tzimtzum", m["concept_id"])
	}
	if _, has := m["subject_field"]; has {
		t.Error("subject_field should be filtered out")
	}
}

func TestProjectFields_WildcardProjection(t *testing.T) {
	input := `{
		"schema_version":1,"ok":true,
		"languages":{
			"he":{"preferred":{"term":"צמצום"},"admitted":[{"term":"zimzum"}]},
			"es":{"preferred":{"term":"tzimtzum"}}
		}
	}`
	got, err := output.ProjectFields([]byte(input), []string{"languages.*.preferred.term"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	langs, ok := m["languages"].(map[string]any)
	if !ok {
		t.Fatalf("languages is not a map: %T", m["languages"])
	}

	he, ok := langs["he"].(map[string]any)
	if !ok {
		t.Fatalf("he is not a map: %T", langs["he"])
	}
	pref, ok := he["preferred"].(map[string]any)
	if !ok {
		t.Fatalf("preferred is not a map: %T", he["preferred"])
	}
	if pref["term"] != "צמצום" {
		t.Errorf("he.preferred.term = %v, want צמצום", pref["term"])
	}
	if _, has := he["admitted"]; has {
		t.Error("admitted should be filtered out")
	}

	es, ok := langs["es"].(map[string]any)
	if !ok {
		t.Fatalf("es is not a map: %T", langs["es"])
	}
	esPref, ok := es["preferred"].(map[string]any)
	if !ok {
		t.Fatalf("es.preferred is not a map: %T", es["preferred"])
	}
	if esPref["term"] != "tzimtzum" {
		t.Errorf("es.preferred.term = %v, want tzimtzum", esPref["term"])
	}
}

func TestProjectFields_ArrayProjection(t *testing.T) {
	input := `{
		"schema_version":1,"ok":true,
		"results":[
			{"concept_id":"c001","subject_field":"kabbalah"},
			{"concept_id":"c002","subject_field":"philosophy"}
		]
	}`
	got, err := output.ProjectFields([]byte(input), []string{"results.concept_id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	results, ok := m["results"].([]any)
	if !ok {
		t.Fatalf("results is not an array: %T", m["results"])
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	r0, ok := results[0].(map[string]any)
	if !ok {
		t.Fatalf("results[0] is not a map: %T", results[0])
	}
	if r0["concept_id"] != "c001" {
		t.Errorf("results[0].concept_id = %v, want c001", r0["concept_id"])
	}
	if _, has := r0["subject_field"]; has {
		t.Error("results[0].subject_field should be filtered out")
	}
}

func TestValidateFields_ErrorHintContainsValidPaths(t *testing.T) {
	_, err := output.ValidateFields("nonexistent", testEnvelope{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "concept_id") {
		t.Errorf("error hint should enumerate valid paths, got: %v", err)
	}
}

func TestValidateFields_EmptyString(t *testing.T) {
	paths, err := output.ValidateFields("", testEnvelope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("empty string should produce nil/empty paths, got %v", paths)
	}
}

type testPtrTermGroup struct {
	Preferred *testTerm `json:"preferred,omitempty"`
}

type testPtrMapEnvelope struct {
	SchemaVersion int                          `json:"schema_version"`
	OK            bool                         `json:"ok"`
	Languages     map[string]*testPtrTermGroup `json:"languages"`
}

func TestValidateFields_MapWithPointerValues(t *testing.T) {
	paths, err := output.ValidateFields("languages.*.preferred.term", testPtrMapEnvelope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 || paths[0] != "languages.*.preferred.term" {
		t.Errorf("got %v, want [languages.*.preferred.term]", paths)
	}
}

func TestValidateFields_RejectUnknownPath(t *testing.T) {
	_, err := output.ValidateFields("concpet_id", testEnvelope{})
	if err == nil {
		t.Fatal("expected error for unknown path")
	}
	if !errors.Is(err, output.ErrInvalidField) {
		t.Errorf("error is not ErrInvalidField: %v", err)
	}
	if !strings.Contains(err.Error(), "concpet_id") {
		t.Errorf("error message should contain the bad path: %v", err)
	}
}
