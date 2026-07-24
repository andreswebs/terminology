package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/andreswebs/terminology/internal/app"
	"github.com/andreswebs/terminology/internal/output"
)

func runReadCLI(t *testing.T, argv []string) (map[string]any, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr
	err := cmd.Run(context.Background(), argv)
	var env map[string]any
	if stdout.Len() > 0 {
		if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
			t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
		}
	}
	return env, err
}

// Cycle 1 — export round-trips.

func TestExport_SortedRichFields(t *testing.T) {
	env, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "export",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	concepts, _ := env["concepts"].([]any)
	if len(concepts) != 2 {
		t.Fatalf("concepts length = %d, want 2", len(concepts))
	}
	c0 := concepts[0].(map[string]any)
	c1 := concepts[1].(map[string]any)
	if c0["concept_id"] != "malkhut" || c1["concept_id"] != "tzimtzum" {
		t.Errorf("concepts not sorted by id: %v, %v", c0["concept_id"], c1["concept_id"])
	}
	// rich fields: concept-level definition on malkhut
	defs, _ := c0["definitions"].([]any)
	if len(defs) != 1 || defs[0] != "The tenth sefirah" {
		t.Errorf("malkhut definitions = %v, want [The tenth sefirah]", c0["definitions"])
	}
	// term-level administrative_status present (canonical shape)
	langs := c0["languages"].(map[string]any)
	en := langs["en"].(map[string]any)
	pref := en["preferred"].(map[string]any)
	if pref["administrative_status"] != "preferredTerm-admn-sts" {
		t.Errorf("preferred.administrative_status = %v", pref["administrative_status"])
	}
}

func TestExport_ApplyRoundTripNoOp(t *testing.T) {
	env, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "export",
	})
	if err != nil {
		t.Fatalf("export error: %v", err)
	}
	raw, _ := json.Marshal(env)
	payloadFile := writePayloadInCWD(t, "export-roundtrip.json", string(raw))
	tbxPath := copyTBXFixture(t, "testdata/fixtures/rich-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr
	if err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "apply", "--file", payloadFile,
	}); err != nil {
		t.Fatalf("apply error: %v\nstderr: %s", err, stderr.String())
	}
	var applyEnv output.ApplyEnvelope
	if jsonErr := json.Unmarshal(stdout.Bytes(), &applyEnv); jsonErr != nil {
		t.Fatalf("apply stdout not JSON: %v", jsonErr)
	}
	if len(applyEnv.Applied.Added) != 0 || len(applyEnv.Applied.Updated) != 0 || len(applyEnv.Applied.Removed) != 0 {
		t.Errorf("expected no-op, got added=%v updated=%v removed=%v",
			applyEnv.Applied.Added, applyEnv.Applied.Updated, applyEnv.Applied.Removed)
	}
	if len(applyEnv.Applied.Unchanged) != 2 {
		t.Errorf("unchanged = %v, want 2 concepts", applyEnv.Applied.Unchanged)
	}
}

// Cycle 2 — show by id, present and absent.

func TestShow_Present(t *testing.T) {
	env, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "show", "malkhut",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := env["concept"].(map[string]any)
	if !ok {
		t.Fatalf("concept missing or wrong type: %T", env["concept"])
	}
	if c["concept_id"] != "malkhut" {
		t.Errorf("concept_id = %v, want malkhut", c["concept_id"])
	}
	langs := c["languages"].(map[string]any)
	en := langs["en"].(map[string]any)
	if _, ok := en["admitted"].([]any); !ok {
		t.Errorf("expected admitted terms in show output, got %v", en)
	}
}

func TestShow_Absent_NotFound(t *testing.T) {
	env, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "show", "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for missing concept")
	}
	if code := output.ExitCodeFor(err); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if coded, ok := err.(interface{ Code() string }); !ok || coded.Code() != "not_found" {
		t.Errorf("error code = %v, want not_found", err)
	}
	if env != nil {
		t.Errorf("expected empty stdout for absent concept, got %v", env)
	}
}

// Cycle 3 — list is the projected view.

func TestList_PreferredOnly(t *testing.T) {
	env, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "list",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	concepts, _ := env["concepts"].([]any)
	if len(concepts) != 2 {
		t.Fatalf("concepts length = %d, want 2", len(concepts))
	}
	c0 := concepts[0].(map[string]any)
	if c0["concept_id"] != "malkhut" {
		t.Errorf("first concept_id = %v, want malkhut (sorted)", c0["concept_id"])
	}
	if _, ok := c0["definitions"]; ok {
		t.Errorf("list should not include definitions: %v", c0)
	}
	langs := c0["languages"].(map[string]any)
	en := langs["en"].(map[string]any)
	pref := en["preferred"].(map[string]any)
	if pref["term"] != "malkhut" {
		t.Errorf("preferred term = %v, want malkhut", pref["term"])
	}
	if _, ok := pref["administrative_status"]; ok {
		t.Errorf("list preferred should carry term only, got %v", pref)
	}
	if _, ok := en["admitted"]; ok {
		t.Errorf("list should not include admitted variants: %v", en)
	}
}

// Cycle 4 — lookup now exposes the rich fields (FEAT-2 repro).

func TestLookup_RichFields(t *testing.T) {
	env, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "malkhut",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	results, _ := env["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results length = %d, want 1", len(results))
	}
	r := results[0].(map[string]any)
	defs, _ := r["definitions"].([]any)
	if len(defs) != 1 {
		t.Errorf("expected concept definitions in lookup output, got %v", r["definitions"])
	}
	langs := r["languages"].(map[string]any)
	en := langs["en"].(map[string]any)
	pref := en["preferred"].(map[string]any)
	if pref["administrative_status"] != "preferredTerm-admn-sts" {
		t.Errorf("preferred.administrative_status = %v", pref["administrative_status"])
	}
	if _, ok := en["deprecated"].([]any); !ok {
		t.Errorf("expected deprecated variants in lookup output, got %v", en)
	}
}

func TestLookup_FieldsDefinitionsValid(t *testing.T) {
	_, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "malkhut",
		"--fields", "results.definitions",
	})
	if err != nil {
		t.Fatalf("results.definitions should be a valid field path, got: %v", err)
	}
}

// Cycle 5 — conventions: --fields, --lang, empty glossary.

func TestExport_FieldsProjection(t *testing.T) {
	env, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "export",
		"--fields", "concepts.concept_id",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	concepts, _ := env["concepts"].([]any)
	c0 := concepts[0].(map[string]any)
	if _, ok := c0["languages"]; ok {
		t.Errorf("projected export should not contain languages: %v", c0)
	}
	if c0["concept_id"] != "malkhut" {
		t.Errorf("concept_id = %v", c0["concept_id"])
	}
}

func TestExport_LangFilter(t *testing.T) {
	env, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "export", "--lang", "he",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	concepts, _ := env["concepts"].([]any)
	for _, ci := range concepts {
		c := ci.(map[string]any)
		langs := c["languages"].(map[string]any)
		for tag := range langs {
			if tag != "he" {
				t.Errorf("concept %v has non-he language %q after --lang he", c["concept_id"], tag)
			}
		}
	}
}

func TestExport_EmptyGlossary(t *testing.T) {
	env, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-empty.tbx", "export",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	concepts, ok := env["concepts"].([]any)
	if !ok {
		t.Fatalf("concepts is not an array: %T", env["concepts"])
	}
	if len(concepts) != 0 {
		t.Errorf("concepts length = %d, want 0", len(concepts))
	}
}

func TestList_EmptyGlossary(t *testing.T) {
	env, err := runReadCLI(t, []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-empty.tbx", "list",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if concepts, ok := env["concepts"].([]any); !ok || len(concepts) != 0 {
		t.Errorf("concepts = %v, want empty array", env["concepts"])
	}
}

// Cycle 6 — schema discoverability.

func TestSchema_NewReadCommands(t *testing.T) {
	for _, name := range []string{"export", "show", "list"} {
		env, err := runReadCLI(t, []string{"terminology", "schema", "--command", name})
		if err != nil {
			t.Fatalf("schema --command %s: %v", name, err)
		}
		if env["name"] != name {
			t.Errorf("schema name = %v, want %s", env["name"], name)
		}
		if _, ok := env["envelope"]; !ok {
			t.Errorf("schema for %s missing envelope", name)
		}
		if _, ok := env["exit_codes"]; !ok {
			t.Errorf("schema for %s missing exit_codes", name)
		}
	}
}

// Golden CLI tests — export, show, list.

func TestExport_Golden(t *testing.T) {
	runGolden(t, "export/all", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "export",
	})
}

func TestExport_LangFilter_Golden(t *testing.T) {
	runGolden(t, "export/lang_he", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "export", "--lang", "he",
	})
}

func TestExport_Empty_Golden(t *testing.T) {
	runGolden(t, "export/empty", []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-empty.tbx", "export",
	})
}

func TestShow_Present_Golden(t *testing.T) {
	runGolden(t, "show/present", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "show", "malkhut",
	})
}

func TestShow_Absent_Golden(t *testing.T) {
	runGolden(t, "show/absent", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "show", "nonexistent",
	})
}

func TestList_Golden(t *testing.T) {
	runGolden(t, "list/all", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "list",
	})
}
