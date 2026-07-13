package output_test

import (
	"encoding/json"
	"testing"

	"github.com/andreswebs/terminology/internal/output"
)

func TestValidateEnvelope_JSONShape(t *testing.T) {
	env := output.ValidateEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Concepts:      2,
		Languages:     []string{"en", "he"},
		Warnings:      []output.ValidateWarning{},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["schema_version"] != float64(1) {
		t.Errorf("schema_version = %v, want 1", got["schema_version"])
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}
	if got["concepts"] != float64(2) {
		t.Errorf("concepts = %v, want 2", got["concepts"])
	}

	langs, ok := got["languages"].([]any)
	if !ok {
		t.Fatalf("languages is not an array: %T", got["languages"])
	}
	if len(langs) != 2 || langs[0] != "en" || langs[1] != "he" {
		t.Errorf("languages = %v, want [en he]", langs)
	}

	warnings, ok := got["warnings"].([]any)
	if !ok {
		t.Fatalf("warnings is not an array: %T", got["warnings"])
	}
	if len(warnings) != 0 {
		t.Errorf("warnings length = %d, want 0", len(warnings))
	}
}

func TestValidateWarning_OmitemptyZeroValues(t *testing.T) {
	w := output.ValidateWarning{
		Code:    "duplicate_id",
		Message: "concept ID 'foo' already exists",
	}

	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	for _, absent := range []string{`"concept_id"`, `"line"`, `"column"`} {
		if contains(raw, absent) {
			t.Errorf("JSON contains %s but should omit zero-valued optional field", absent)
		}
	}

	for _, present := range []string{`"code"`, `"message"`} {
		if !contains(raw, present) {
			t.Errorf("JSON missing required field %s", present)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLookupEnvelope_JSONShape(t *testing.T) {
	env := output.LookupEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Results: []output.WriteResult{
			{
				ConceptID:    "tzimtzum",
				SubjectField: "kabbalah",
				Languages: map[string]output.WriteTermGroup{
					"he": {Preferred: &output.WriteTerm{Term: "צמצום"}},
					"es": {Preferred: &output.WriteTerm{Term: "tzimtzum"}},
				},
			},
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["schema_version"] != float64(1) {
		t.Errorf("schema_version = %v, want 1", got["schema_version"])
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}

	results, ok := got["results"].([]any)
	if !ok {
		t.Fatalf("results is not an array: %T", got["results"])
	}
	if len(results) != 1 {
		t.Fatalf("results length = %d, want 1", len(results))
	}

	r := results[0].(map[string]any)
	if r["concept_id"] != "tzimtzum" {
		t.Errorf("concept_id = %v, want tzimtzum", r["concept_id"])
	}
	if r["subject_field"] != "kabbalah" {
		t.Errorf("subject_field = %v, want kabbalah", r["subject_field"])
	}

	langs, ok := r["languages"].(map[string]any)
	if !ok {
		t.Fatalf("languages is not an object: %T", r["languages"])
	}

	he, ok := langs["he"].(map[string]any)
	if !ok {
		t.Fatalf("he is not an object: %T", langs["he"])
	}
	pref, ok := he["preferred"].(map[string]any)
	if !ok {
		t.Fatalf("preferred is not an object: %T", he["preferred"])
	}
	if pref["term"] != "צמצום" {
		t.Errorf("he preferred term = %v, want צמצום", pref["term"])
	}
}

func TestLookupEnvelope_EmptyResultsSerializesAsArray(t *testing.T) {
	env := output.LookupEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Results:       []output.WriteResult{},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	if !contains(raw, `"results":[]`) {
		t.Errorf("expected empty results array, got: %s", raw)
	}
}

func TestLookupEnvelope_NilResultsSerializesAsArray(t *testing.T) {
	env := output.LookupEnvelope{
		SchemaVersion: 1,
		OK:            true,
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	results, ok := got["results"].([]any)
	if !ok {
		t.Fatalf("results should be an array, got %T", got["results"])
	}
	if len(results) != 0 {
		t.Errorf("results length = %d, want 0", len(results))
	}
}

func TestExportEnvelope_EmptyConceptsSerializesAsArray(t *testing.T) {
	env := output.ExportEnvelope{SchemaVersion: 1, OK: true}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if raw := string(data); !contains(raw, `"concepts":[]`) {
		t.Errorf("expected empty concepts array, got: %s", raw)
	}
}

func TestShowEnvelope_NilLanguagesSerializesAsObject(t *testing.T) {
	env := output.ShowEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Concept:       output.WriteResult{ConceptID: "test"},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if raw := string(data); !contains(raw, `"languages":{}`) {
		t.Errorf("expected empty languages object, got: %s", raw)
	}
}

func TestListEnvelope_EmptyConceptsSerializesAsArray(t *testing.T) {
	env := output.ListEnvelope{SchemaVersion: 1, OK: true}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if raw := string(data); !contains(raw, `"concepts":[]`) {
		t.Errorf("expected empty concepts array, got: %s", raw)
	}
}

func TestExtractEnvelope_JSONShape(t *testing.T) {
	env := output.ExtractEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Candidates: []output.ExtractCandidate{
			{
				Term:      "Holy Temple",
				Frequency: 5,
				Heuristic: "capitalized_phrase",
				Locations: []output.ExtractLocation{
					{File: "ch1.md", Line: 12, Col: 5},
				},
			},
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["schema_version"] != float64(1) {
		t.Errorf("schema_version = %v, want 1", got["schema_version"])
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}

	candidates, ok := got["candidates"].([]any)
	if !ok {
		t.Fatalf("candidates is not an array: %T", got["candidates"])
	}
	if len(candidates) != 1 {
		t.Fatalf("candidates length = %d, want 1", len(candidates))
	}

	c := candidates[0].(map[string]any)
	if c["term"] != "Holy Temple" {
		t.Errorf("term = %v, want Holy Temple", c["term"])
	}
	if c["frequency"] != float64(5) {
		t.Errorf("frequency = %v, want 5", c["frequency"])
	}
	if c["heuristic"] != "capitalized_phrase" {
		t.Errorf("heuristic = %v, want capitalized_phrase", c["heuristic"])
	}

	locs, ok := c["locations"].([]any)
	if !ok {
		t.Fatalf("locations is not an array: %T", c["locations"])
	}
	if len(locs) != 1 {
		t.Fatalf("locations length = %d, want 1", len(locs))
	}

	loc := locs[0].(map[string]any)
	if loc["file"] != "ch1.md" {
		t.Errorf("file = %v, want ch1.md", loc["file"])
	}
	if loc["line"] != float64(12) {
		t.Errorf("line = %v, want 12", loc["line"])
	}
	if loc["col"] != float64(5) {
		t.Errorf("col = %v, want 5", loc["col"])
	}
}

func TestExtractEnvelope_NilCandidatesSerializesAsArray(t *testing.T) {
	env := output.ExtractEnvelope{
		SchemaVersion: 1,
		OK:            true,
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	candidates, ok := got["candidates"].([]any)
	if !ok {
		t.Fatalf("candidates should be an array, got %T", got["candidates"])
	}
	if len(candidates) != 0 {
		t.Errorf("candidates length = %d, want 0", len(candidates))
	}
}

func TestExtractEnvelope_EmptyCandidatesSerializesAsArray(t *testing.T) {
	env := output.ExtractEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Candidates:    []output.ExtractCandidate{},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	if !contains(raw, `"candidates":[]`) {
		t.Errorf("expected empty candidates array, got: %s", raw)
	}
}

func TestExtractCandidate_OmitemptyLocations(t *testing.T) {
	c := output.ExtractCandidate{
		Term:      "test",
		Frequency: 3,
		Heuristic: "capitalized_phrase",
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	if contains(raw, `"locations"`) {
		t.Errorf("JSON contains locations but should omit empty: %s", raw)
	}
}

func TestExtractLocation_OmitemptyLineCol(t *testing.T) {
	loc := output.ExtractLocation{
		File: "ch1.md",
	}

	data, err := json.Marshal(loc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	if contains(raw, `"line"`) {
		t.Errorf("JSON contains line but should omit zero: %s", raw)
	}
	if contains(raw, `"col"`) {
		t.Errorf("JSON contains col but should omit zero: %s", raw)
	}
	if !contains(raw, `"file"`) {
		t.Errorf("JSON missing required file field: %s", raw)
	}
}

func TestValidateWarning_AllFieldsPopulated(t *testing.T) {
	w := output.ValidateWarning{
		Code:      "unresolved_crossref",
		Message:   "concept 'foo' references unknown ID 'bar'",
		ConceptID: "foo",
		Line:      42,
		Col:       12,
	}

	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["code"] != "unresolved_crossref" {
		t.Errorf("code = %v, want unresolved_crossref", got["code"])
	}
	if got["concept_id"] != "foo" {
		t.Errorf("concept_id = %v, want foo", got["concept_id"])
	}
	if got["line"] != float64(42) {
		t.Errorf("line = %v, want 42", got["line"])
	}
	if got["column"] != float64(12) {
		t.Errorf("column = %v, want 12", got["column"])
	}
}

func TestScanEnvelope_Registered(t *testing.T) {
	v, ok := output.EnvelopeFor("scan")
	if !ok {
		t.Fatal("scan envelope not registered")
	}
	if _, ok := v.(output.ScanEnvelope); !ok {
		t.Errorf("scan envelope type = %T, want output.ScanEnvelope", v)
	}
}

func TestScanEnvelope_JSONShape(t *testing.T) {
	env := output.ScanEnvelope{
		SchemaVersion: 1,
		OK:            true,
		File:          "source/ch1.md",
		Matches: []output.ScanMatch{
			{
				ConceptID: "c001",
				Term:      "tzimtzum",
				Lang:      "es",
				Status:    "preferred",
				Line:      14,
				Column:    23,
				Context:   "...El concepto de tzimtzum es central...",
			},
		},
		Summary: output.ScanSummary{
			TotalMatches:   1,
			UniqueConcepts: 1,
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["schema_version"] != float64(1) {
		t.Errorf("schema_version = %v, want 1", got["schema_version"])
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}
	if got["file"] != "source/ch1.md" {
		t.Errorf("file = %v, want source/ch1.md", got["file"])
	}

	matches, ok := got["matches"].([]any)
	if !ok {
		t.Fatalf("matches is not an array: %T", got["matches"])
	}
	if len(matches) != 1 {
		t.Fatalf("matches length = %d, want 1", len(matches))
	}

	m := matches[0].(map[string]any)
	if m["concept_id"] != "c001" {
		t.Errorf("concept_id = %v, want c001", m["concept_id"])
	}
	if m["term"] != "tzimtzum" {
		t.Errorf("term = %v, want tzimtzum", m["term"])
	}
	if m["lang"] != "es" {
		t.Errorf("lang = %v, want es", m["lang"])
	}
	if m["status"] != "preferred" {
		t.Errorf("status = %v, want preferred", m["status"])
	}
	if m["line"] != float64(14) {
		t.Errorf("line = %v, want 14", m["line"])
	}
	if m["column"] != float64(23) {
		t.Errorf("column = %v, want 23", m["column"])
	}
	if m["context"] != "...El concepto de tzimtzum es central..." {
		t.Errorf("context = %v, want expected string", m["context"])
	}

	summary, ok := got["summary"].(map[string]any)
	if !ok {
		t.Fatalf("summary is not an object: %T", got["summary"])
	}
	if summary["total_matches"] != float64(1) {
		t.Errorf("total_matches = %v, want 1", summary["total_matches"])
	}
	if summary["unique_concepts"] != float64(1) {
		t.Errorf("unique_concepts = %v, want 1", summary["unique_concepts"])
	}
}

func TestScanEnvelope_NilMatchesSerializesAsArray(t *testing.T) {
	env := output.ScanEnvelope{
		SchemaVersion: 1,
		OK:            true,
		File:          "test.md",
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	matches, ok := got["matches"].([]any)
	if !ok {
		t.Fatalf("matches should be an array, got %T", got["matches"])
	}
	if len(matches) != 0 {
		t.Errorf("matches length = %d, want 0", len(matches))
	}
}

func TestScanEnvelope_EmptyMatchesSerializesAsArray(t *testing.T) {
	env := output.ScanEnvelope{
		SchemaVersion: 1,
		OK:            true,
		File:          "test.md",
		Matches:       []output.ScanMatch{},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	if !contains(raw, `"matches":[]`) {
		t.Errorf("expected empty matches array, got: %s", raw)
	}
}

func TestCheckEnvelope_Registered(t *testing.T) {
	v, ok := output.EnvelopeFor("check")
	if !ok {
		t.Fatal("check envelope not registered")
	}
	if _, ok := v.(output.CheckEnvelope); !ok {
		t.Errorf("check envelope type = %T, want output.CheckEnvelope", v)
	}
}

func TestCheckEnvelope_JSONShape_MissingViolation(t *testing.T) {
	env := output.CheckEnvelope{
		SchemaVersion: 1,
		OK:            false,
		Source:        "source/ch1.md",
		Target:        "target/ch1.md",
		Violations: []output.CheckViolation{
			{
				Type:              "missing",
				ConceptID:         "c001",
				SourceTerm:        "tzimtzum",
				ExpectedTarget:    "צמצום",
				SourceOccurrences: 5,
			},
		},
		Warnings: []output.CheckWarning{},
		Summary: output.CheckSummary{
			Violations:      1,
			Warnings:        0,
			ConceptsChecked: 12,
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["schema_version"] != float64(1) {
		t.Errorf("schema_version = %v, want 1", got["schema_version"])
	}
	if got["ok"] != false {
		t.Errorf("ok = %v, want false", got["ok"])
	}
	if got["source"] != "source/ch1.md" {
		t.Errorf("source = %v, want source/ch1.md", got["source"])
	}
	if got["target"] != "target/ch1.md" {
		t.Errorf("target = %v, want target/ch1.md", got["target"])
	}

	violations, ok := got["violations"].([]any)
	if !ok {
		t.Fatalf("violations is not an array: %T", got["violations"])
	}
	if len(violations) != 1 {
		t.Fatalf("violations length = %d, want 1", len(violations))
	}

	v := violations[0].(map[string]any)
	if v["type"] != "missing" {
		t.Errorf("type = %v, want missing", v["type"])
	}
	if v["concept_id"] != "c001" {
		t.Errorf("concept_id = %v, want c001", v["concept_id"])
	}
	if v["source_term"] != "tzimtzum" {
		t.Errorf("source_term = %v, want tzimtzum", v["source_term"])
	}
	if v["expected_target"] != "צמצום" {
		t.Errorf("expected_target = %v, want צמצום", v["expected_target"])
	}
	if v["source_occurrences"] != float64(5) {
		t.Errorf("source_occurrences = %v, want 5", v["source_occurrences"])
	}

	summary, ok := got["summary"].(map[string]any)
	if !ok {
		t.Fatalf("summary is not an object: %T", got["summary"])
	}
	if summary["violations"] != float64(1) {
		t.Errorf("summary violations = %v, want 1", summary["violations"])
	}
	if summary["warnings"] != float64(0) {
		t.Errorf("summary warnings = %v, want 0", summary["warnings"])
	}
	if summary["concepts_checked"] != float64(12) {
		t.Errorf("summary concepts_checked = %v, want 12", summary["concepts_checked"])
	}
}

func TestCheckEnvelope_JSONShape_ForbiddenVariantViolation(t *testing.T) {
	env := output.CheckEnvelope{
		SchemaVersion: 1,
		OK:            false,
		Source:        "source/ch1.md",
		Target:        "target/ch1.md",
		Violations: []output.CheckViolation{
			{
				Type:      "forbidden_variant",
				ConceptID: "razon-historica",
				Variant:   "razón histórica",
				Line:      142,
				Column:    12,
				Context:   "...la razón histórica de su pensamiento...",
			},
		},
		Warnings: []output.CheckWarning{},
		Summary: output.CheckSummary{
			Violations:      1,
			Warnings:        0,
			ConceptsChecked: 5,
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	violations := got["violations"].([]any)
	v := violations[0].(map[string]any)
	if v["type"] != "forbidden_variant" {
		t.Errorf("type = %v, want forbidden_variant", v["type"])
	}
	if v["variant"] != "razón histórica" {
		t.Errorf("variant = %v, want razón histórica", v["variant"])
	}
	if v["line"] != float64(142) {
		t.Errorf("line = %v, want 142", v["line"])
	}
	if v["column"] != float64(12) {
		t.Errorf("column = %v, want 12", v["column"])
	}
	if v["context"] != "...la razón histórica de su pensamiento..." {
		t.Errorf("context = %v, want expected string", v["context"])
	}
}

func TestCheckViolation_OmitemptyMissingType(t *testing.T) {
	v := output.CheckViolation{
		Type:              "missing",
		ConceptID:         "c001",
		SourceTerm:        "tzimtzum",
		ExpectedTarget:    "צמצום",
		SourceOccurrences: 3,
	}

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	for _, absent := range []string{`"variant"`, `"line"`, `"column"`, `"context"`} {
		if contains(raw, absent) {
			t.Errorf("JSON contains %s but should omit for missing violation type", absent)
		}
	}
	for _, present := range []string{`"source_term"`, `"expected_target"`, `"source_occurrences"`} {
		if !contains(raw, present) {
			t.Errorf("JSON missing field %s for missing violation type", present)
		}
	}
}

func TestCheckViolation_OmitemptyForbiddenVariantType(t *testing.T) {
	v := output.CheckViolation{
		Type:      "forbidden_variant",
		ConceptID: "c001",
		Variant:   "contraction",
		Line:      17,
		Column:    4,
		Context:   "...some context...",
	}

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	for _, absent := range []string{`"source_term"`, `"expected_target"`, `"source_occurrences"`} {
		if contains(raw, absent) {
			t.Errorf("JSON contains %s but should omit for forbidden_variant type", absent)
		}
	}
	for _, present := range []string{`"variant"`, `"line"`, `"column"`, `"context"`} {
		if !contains(raw, present) {
			t.Errorf("JSON missing field %s for forbidden_variant type", present)
		}
	}
}

func TestCheckEnvelope_NilViolationsSerializesAsArray(t *testing.T) {
	env := output.CheckEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Source:        "source/ch1.md",
		Target:        "target/ch1.md",
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	violations, ok := got["violations"].([]any)
	if !ok {
		t.Fatalf("violations should be an array, got %T", got["violations"])
	}
	if len(violations) != 0 {
		t.Errorf("violations length = %d, want 0", len(violations))
	}

	warnings, ok := got["warnings"].([]any)
	if !ok {
		t.Fatalf("warnings should be an array, got %T", got["warnings"])
	}
	if len(warnings) != 0 {
		t.Errorf("warnings length = %d, want 0", len(warnings))
	}
}

func TestCheckEnvelope_EmptySlicesSerializeAsArrays(t *testing.T) {
	env := output.CheckEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Source:        "source/ch1.md",
		Target:        "target/ch1.md",
		Violations:    []output.CheckViolation{},
		Warnings:      []output.CheckWarning{},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	if !contains(raw, `"violations":[]`) {
		t.Errorf("expected empty violations array, got: %s", raw)
	}
	if !contains(raw, `"warnings":[]`) {
		t.Errorf("expected empty warnings array, got: %s", raw)
	}
}

func TestCheckWarning_JSONShape(t *testing.T) {
	w := output.CheckWarning{
		Type:      "admitted_variant",
		ConceptID: "c001",
		Message:   "admitted variant 'alt' used instead of preferred 'pref'",
	}

	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["type"] != "admitted_variant" {
		t.Errorf("type = %v, want admitted_variant", got["type"])
	}
	if got["concept_id"] != "c001" {
		t.Errorf("concept_id = %v, want c001", got["concept_id"])
	}
	if got["message"] != "admitted variant 'alt' used instead of preferred 'pref'" {
		t.Errorf("message = %v, want expected string", got["message"])
	}
}

func TestWriteEnvelope_JSONShape(t *testing.T) {
	env := output.WriteEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Result: output.WriteResult{
			ConceptID:    "tzimtzum",
			SubjectField: "kabbalah",
			Languages: map[string]output.WriteTermGroup{
				"he": {
					Preferred: &output.WriteTerm{
						Term:                 "צמצום",
						AdministrativeStatus: "preferredTerm-admn-sts",
						PartOfSpeech:         "noun",
						GrammaticalGender:    "masculine",
					},
				},
				"es": {
					Preferred: &output.WriteTerm{
						Term:                 "tzimtzum",
						AdministrativeStatus: "preferredTerm-admn-sts",
					},
					Deprecated: []output.WriteTerm{
						{Term: "contracción", AdministrativeStatus: "deprecatedTerm-admn-sts"},
					},
				},
			},
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["schema_version"] != float64(1) {
		t.Errorf("schema_version = %v, want 1", got["schema_version"])
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}

	result, ok := got["result"].(map[string]any)
	if !ok {
		t.Fatalf("result is not an object: %T", got["result"])
	}
	if result["concept_id"] != "tzimtzum" {
		t.Errorf("concept_id = %v, want tzimtzum", result["concept_id"])
	}
	if result["subject_field"] != "kabbalah" {
		t.Errorf("subject_field = %v, want kabbalah", result["subject_field"])
	}

	langs, ok := result["languages"].(map[string]any)
	if !ok {
		t.Fatalf("languages is not an object: %T", result["languages"])
	}

	he, ok := langs["he"].(map[string]any)
	if !ok {
		t.Fatalf("he is not an object: %T", langs["he"])
	}
	pref, ok := he["preferred"].(map[string]any)
	if !ok {
		t.Fatalf("preferred is not an object: %T", he["preferred"])
	}
	if pref["term"] != "צמצום" {
		t.Errorf("he preferred term = %v, want צמצום", pref["term"])
	}
	if pref["administrative_status"] != "preferredTerm-admn-sts" {
		t.Errorf("administrative_status = %v, want preferredTerm-admn-sts", pref["administrative_status"])
	}
	if pref["part_of_speech"] != "noun" {
		t.Errorf("part_of_speech = %v, want noun", pref["part_of_speech"])
	}
	if pref["grammatical_gender"] != "masculine" {
		t.Errorf("grammatical_gender = %v, want masculine", pref["grammatical_gender"])
	}

	es, ok := langs["es"].(map[string]any)
	if !ok {
		t.Fatalf("es is not an object: %T", langs["es"])
	}
	deprecated, ok := es["deprecated"].([]any)
	if !ok {
		t.Fatalf("deprecated is not an array: %T", es["deprecated"])
	}
	if len(deprecated) != 1 {
		t.Fatalf("deprecated length = %d, want 1", len(deprecated))
	}
	dep := deprecated[0].(map[string]any)
	if dep["term"] != "contracción" {
		t.Errorf("deprecated term = %v, want contracción", dep["term"])
	}
}

func TestWriteTerm_OmitemptyOptionalFields(t *testing.T) {
	term := output.WriteTerm{
		Term: "tzimtzum",
	}

	data, err := json.Marshal(term)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	for _, absent := range []string{
		`"administrative_status"`, `"part_of_speech"`, `"grammatical_gender"`,
		`"grammatical_number"`, `"register"`, `"term_type"`, `"term_location"`,
		`"geographical_usage"`, `"transfer_comment"`, `"sources"`,
		`"customer_subset"`, `"project_subset"`, `"external_refs"`, `"notes"`,
	} {
		if contains(raw, absent) {
			t.Errorf("JSON contains %s but should omit zero-valued optional field", absent)
		}
	}
	if !contains(raw, `"term"`) {
		t.Errorf("JSON missing required field term")
	}
}

func TestWriteTermGroup_OmitemptySlices(t *testing.T) {
	g := output.WriteTermGroup{
		Preferred: &output.WriteTerm{Term: "test"},
	}

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	for _, absent := range []string{`"admitted"`, `"deprecated"`, `"superseded"`} {
		if contains(raw, absent) {
			t.Errorf("JSON contains %s but should omit empty slice", absent)
		}
	}
}

func TestWriteResult_OmitemptySubjectField(t *testing.T) {
	r := output.WriteResult{
		ConceptID: "test",
		Languages: map[string]output.WriteTermGroup{},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	raw := string(data)
	if contains(raw, `"subject_field"`) {
		t.Errorf("JSON contains subject_field but should omit empty: %s", raw)
	}
}

func TestWriteResult_FullConceptShape(t *testing.T) {
	r := output.WriteResult{
		ConceptID:    "tzimtzum",
		SubjectField: "kabbalah",
		Definitions:  []string{"The divine self-contraction"},
		Notes:        []string{"Central to Lurianic Kabbalah"},
		CrossRefs:    []output.WriteCrossRef{{Target: "sefirot", Label: "related"}},
		Sources:      []string{"Etz Chaim"},
		Languages: map[string]output.WriteTermGroup{
			"he": {
				Preferred: &output.WriteTerm{Term: "צמצום"},
			},
		},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	defs, ok := got["definitions"].([]any)
	if !ok {
		t.Fatalf("definitions is not an array: %T", got["definitions"])
	}
	if len(defs) != 1 || defs[0] != "The divine self-contraction" {
		t.Errorf("definitions = %v, want [The divine self-contraction]", defs)
	}

	notes, ok := got["notes"].([]any)
	if !ok {
		t.Fatalf("notes is not an array: %T", got["notes"])
	}
	if len(notes) != 1 {
		t.Errorf("notes length = %d, want 1", len(notes))
	}

	xrefs, ok := got["cross_refs"].([]any)
	if !ok {
		t.Fatalf("cross_refs is not an array: %T", got["cross_refs"])
	}
	if len(xrefs) != 1 {
		t.Errorf("cross_refs length = %d, want 1", len(xrefs))
	}
	xref := xrefs[0].(map[string]any)
	if xref["target"] != "sefirot" {
		t.Errorf("cross_ref target = %v, want sefirot", xref["target"])
	}

	sources, ok := got["sources"].([]any)
	if !ok {
		t.Fatalf("sources is not an array: %T", got["sources"])
	}
	if len(sources) != 1 {
		t.Errorf("sources length = %d, want 1", len(sources))
	}
}

func TestWriteEnvelope_NilLanguagesSerializesAsEmptyObject(t *testing.T) {
	env := output.WriteEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Result: output.WriteResult{
			ConceptID: "test",
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	result := got["result"].(map[string]any)
	langs, ok := result["languages"].(map[string]any)
	if !ok {
		t.Fatalf("languages should be an object, got %T", result["languages"])
	}
	if len(langs) != 0 {
		t.Errorf("languages length = %d, want 0", len(langs))
	}
}

func TestWriteEnvelope_ConceptRemoveShowsLastState(t *testing.T) {
	env := output.WriteEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Result: output.WriteResult{
			ConceptID:    "obsolete-concept",
			SubjectField: "theology",
			Languages: map[string]output.WriteTermGroup{
				"en": {
					Preferred: &output.WriteTerm{
						Term:                 "obsolete term",
						AdministrativeStatus: "preferredTerm-admn-sts",
					},
				},
			},
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	result := got["result"].(map[string]any)
	if result["concept_id"] != "obsolete-concept" {
		t.Errorf("removed concept_id = %v, want obsolete-concept", result["concept_id"])
	}
	langs := result["languages"].(map[string]any)
	en := langs["en"].(map[string]any)
	pref := en["preferred"].(map[string]any)
	if pref["term"] != "obsolete term" {
		t.Errorf("removed concept preferred term = %v, want 'obsolete term'", pref["term"])
	}
}

func TestWriteTerm_AllDataCategories(t *testing.T) {
	term := output.WriteTerm{
		Term:                 "צמצום",
		AdministrativeStatus: "preferredTerm-admn-sts",
		PartOfSpeech:         "noun",
		GrammaticalGender:    "masculine",
		GrammaticalNumber:    "singular",
		Register:             "technicalRegister",
		TermType:             "fullForm",
		TermLocation:         "checkBox",
		GeographicalUsage:    "IL",
		TransferComment:      "transliterate as tzimtzum",
		Reading:              "tsimtsum",
		ReadingNote:          "Ashkenazi pronunciation",
		Sources:              []string{"Etz Chaim"},
		CustomerSubset:       "academic",
		ProjectSubset:        "translation-2024",
		ExternalRefs:         []string{"https://example.com/tzimtzum"},
		CrossRefs:            []output.WriteCrossRef{{Target: "sefirot"}},
		Contexts:             []string{"The concept of tzimtzum is central..."},
		Notes:                []string{"Lurianic term"},
	}

	data, err := json.Marshal(term)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	checks := map[string]string{
		"term":                  "צמצום",
		"administrative_status": "preferredTerm-admn-sts",
		"part_of_speech":        "noun",
		"grammatical_gender":    "masculine",
		"grammatical_number":    "singular",
		"register":              "technicalRegister",
		"term_type":             "fullForm",
		"term_location":         "checkBox",
		"geographical_usage":    "IL",
		"transfer_comment":      "transliterate as tzimtzum",
		"reading":               "tsimtsum",
		"reading_note":          "Ashkenazi pronunciation",
		"customer_subset":       "academic",
		"project_subset":        "translation-2024",
	}
	for k, want := range checks {
		if got[k] != want {
			t.Errorf("%s = %v, want %v", k, got[k], want)
		}
	}

	sliceChecks := map[string]int{
		"sources":       1,
		"external_refs": 1,
		"cross_refs":    1,
		"contexts":      1,
		"notes":         1,
	}
	for k, wantLen := range sliceChecks {
		arr, ok := got[k].([]any)
		if !ok {
			t.Errorf("%s is not an array: %T", k, got[k])
			continue
		}
		if len(arr) != wantLen {
			t.Errorf("%s length = %d, want %d", k, len(arr), wantLen)
		}
	}
}

func TestApplyEnvelope_JSONShape(t *testing.T) {
	env := output.ApplyEnvelope{
		SchemaVersion: 1,
		OK:            true,
		Applied: output.ApplyResult{
			Added:     []string{"tzimtzum"},
			Updated:   []string{"razon-historica"},
			Removed:   []string{},
			Unchanged: []string{"binah", "malkhut"},
		},
		Warnings: []string{},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["schema_version"] != float64(1) {
		t.Errorf("schema_version = %v, want 1", got["schema_version"])
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}

	applied, ok := got["applied"].(map[string]any)
	if !ok {
		t.Fatalf("applied is not an object: %T", got["applied"])
	}

	added, ok := applied["added"].([]any)
	if !ok {
		t.Fatalf("added is not an array: %T", applied["added"])
	}
	if len(added) != 1 || added[0] != "tzimtzum" {
		t.Errorf("added = %v, want [tzimtzum]", added)
	}

	updated, ok := applied["updated"].([]any)
	if !ok {
		t.Fatalf("updated is not an array: %T", applied["updated"])
	}
	if len(updated) != 1 || updated[0] != "razon-historica" {
		t.Errorf("updated = %v, want [razon-historica]", updated)
	}

	removed, ok := applied["removed"].([]any)
	if !ok {
		t.Fatalf("removed is not an array: %T", applied["removed"])
	}
	if len(removed) != 0 {
		t.Errorf("removed length = %d, want 0", len(removed))
	}

	unchanged, ok := applied["unchanged"].([]any)
	if !ok {
		t.Fatalf("unchanged is not an array: %T", applied["unchanged"])
	}
	if len(unchanged) != 2 {
		t.Errorf("unchanged length = %d, want 2", len(unchanged))
	}
}

func TestApplyEnvelope_NilSlicesSerializeAsArrays(t *testing.T) {
	env := output.ApplyEnvelope{
		SchemaVersion: 1,
		OK:            true,
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	applied, ok := got["applied"].(map[string]any)
	if !ok {
		t.Fatalf("applied is not an object: %T", got["applied"])
	}

	for _, field := range []string{"added", "updated", "removed", "unchanged"} {
		arr, ok := applied[field].([]any)
		if !ok {
			t.Errorf("applied.%s should be an array, got %T", field, applied[field])
			continue
		}
		if len(arr) != 0 {
			t.Errorf("applied.%s length = %d, want 0", field, len(arr))
		}
	}

	warnings, ok := got["warnings"].([]any)
	if !ok {
		t.Fatalf("warnings should be an array, got %T", got["warnings"])
	}
	if len(warnings) != 0 {
		t.Errorf("warnings length = %d, want 0", len(warnings))
	}
}

func TestApplyEnvelope_Registered(t *testing.T) {
	v, ok := output.EnvelopeFor("apply")
	if !ok {
		t.Fatal("apply envelope not registered")
	}
	if _, ok := v.(output.ApplyEnvelope); !ok {
		t.Errorf("apply envelope type = %T, want output.ApplyEnvelope", v)
	}
}

func TestApplyExitCodes_Registered(t *testing.T) {
	codes, ok := output.ExitCodesFor("apply")
	if !ok {
		t.Fatal("apply exit codes not registered")
	}
	want := map[int]bool{0: true, 1: true, 2: true, 3: true, 65: true}
	got := make(map[int]bool)
	for _, c := range codes {
		got[c] = true
	}
	for code := range want {
		if !got[code] {
			t.Errorf("missing exit code %d", code)
		}
	}
	if len(codes) != len(want) {
		t.Errorf("exit codes = %v, want exactly {0, 1, 2, 3, 65}", codes)
	}
}

func TestApplyFailure_JSONShape(t *testing.T) {
	f := output.ApplyFailure{
		ConceptID: "razon-historica",
		Code:      "dangling_crossref",
		Message:   "unresolved reference to 'tzimtzum'",
	}

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["concept_id"] != "razon-historica" {
		t.Errorf("concept_id = %v, want razon-historica", got["concept_id"])
	}
	if got["code"] != "dangling_crossref" {
		t.Errorf("code = %v, want dangling_crossref", got["code"])
	}
	if got["message"] != "unresolved reference to 'tzimtzum'" {
		t.Errorf("message = %v, want expected string", got["message"])
	}
}

func TestWriteEnvelope_Registered(t *testing.T) {
	for _, cmd := range []string{"concept add", "concept update", "concept remove", "term add", "term deprecate"} {
		v, ok := output.EnvelopeFor(cmd)
		if !ok {
			t.Errorf("%s envelope not registered", cmd)
			continue
		}
		if _, ok := v.(output.WriteEnvelope); !ok {
			t.Errorf("%s envelope type = %T, want output.WriteEnvelope", cmd, v)
		}
	}
}
