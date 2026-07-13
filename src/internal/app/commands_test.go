package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/app"
	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	_ "github.com/andreswebs/terminology/internal/tbx/linguist"
)

func TestValidate_NoTBXPath_ExitCode2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "validate"})
	if err == nil {
		t.Fatal("expected error for missing --tbx, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}

	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty, got %q", stdout.String())
	}
}

func TestRoot_InvalidFormat_ExitCode2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "--format", "yaml", "validate"})
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestRoot_BareInvocation_ExitsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology"})
	if err == nil {
		t.Fatal("expected error for bare invocation, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestRoot_VerbosityMutex(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"verbose+debug", []string{"terminology", "--verbose", "--debug", "validate"}},
		{"verbose+quiet", []string{"terminology", "--verbose", "--quiet", "validate"}},
		{"debug+quiet", []string{"terminology", "--debug", "--quiet", "validate"}},
		{"all_three", []string{"terminology", "--verbose", "--debug", "--quiet", "validate"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := app.Root()
			cmd.Writer = &stdout
			cmd.ErrWriter = &stderr

			err := cmd.Run(context.Background(), tc.args)
			if err == nil {
				t.Fatal("expected error for conflicting verbosity flags, got nil")
			}

			exitCode := output.ExitCodeFor(err)
			if exitCode != 2 {
				t.Errorf("exit code = %d, want 2", exitCode)
			}
		})
	}
}

func TestRoot_Version(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "--version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if got == "" {
		t.Fatal("--version produced no output")
	}
}

func TestValidate_NoTBXPath_EnvelopeShape(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "validate"})
	if err == nil {
		t.Fatal("expected error for missing --tbx, got nil")
	}

	output.EmitError(&stderr, "json", err)

	var env map[string]any
	if jsonErr := json.Unmarshal(stderr.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr: %q", jsonErr, stderr.String())
	}

	errObj, ok := env["error"].(map[string]any)
	if !ok {
		t.Fatalf("envelope missing error object, got: %s", stderr.String())
	}

	if code, _ := errObj["code"].(string); code != "no_tbx_path" {
		t.Errorf("error code = %q, want %q", code, "no_tbx_path")
	}

	msg, _ := errObj["message"].(string)
	if msg == "" {
		t.Error("error message is empty")
	}
}

func TestValidate_CleanFile_Success(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-dct.tbx", "validate",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	if ok, _ := env["ok"].(bool); !ok {
		t.Errorf("ok = %v, want true", env["ok"])
	}
	if concepts, _ := env["concepts"].(float64); int(concepts) != 1 {
		t.Errorf("concepts = %v, want 1", env["concepts"])
	}

	if sv, _ := env["schema_version"].(float64); int(sv) != output.SchemaVersion {
		t.Errorf("schema_version = %v, want %d", sv, output.SchemaVersion)
	}
}

func TestValidate_MalformedXML_ExitCode65(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/bad-xml.tbx", "validate",
	})
	if err == nil {
		t.Fatal("expected error for malformed XML, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 65 {
		t.Errorf("exit code = %d, want 65", exitCode)
	}

	var buf bytes.Buffer
	output.EmitError(&buf, "json", err)

	var env map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr: %q", jsonErr, buf.String())
	}

	errObj, ok := env["error"].(map[string]any)
	if !ok {
		t.Fatalf("envelope missing error object, got: %s", buf.String())
	}

	if code, _ := errObj["code"].(string); code != "validation_error" {
		t.Errorf("error code = %q, want %q", code, "validation_error")
	}
}

func TestValidate_NonexistentFile_ExitCode65(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/nonexistent.tbx", "validate",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 65 {
		t.Errorf("exit code = %d, want 65", exitCode)
	}
}

func TestValidate_NoTBX_Golden(t *testing.T) {
	runGolden(t, "validate/no_tbx", []string{"terminology", "validate"})
}

func TestValidate_Clean_Golden(t *testing.T) {
	runGolden(t, "validate/clean", []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-dct.tbx", "validate",
	})
}

func TestValidate_Warnings_Golden(t *testing.T) {
	runGolden(t, "validate/warnings", []string{
		"terminology", "--tbx", "testdata/fixtures/with-warnings.tbx", "validate",
	})
}

func TestValidate_Strict_Golden(t *testing.T) {
	runGolden(t, "validate/strict", []string{
		"terminology", "--tbx", "testdata/fixtures/with-warnings.tbx", "validate", "--strict",
	})
}

func TestValidate_MalformedXML_Golden(t *testing.T) {
	runGolden(t, "validate/malformed_xml", []string{
		"terminology", "--tbx", "testdata/fixtures/bad-xml.tbx", "validate",
	})
}

func TestValidate_StrictWithLegacy_Golden(t *testing.T) {
	runGolden(t, "validate/strict_with_legacy", []string{
		"terminology", "--tbx", "testdata/fixtures/with-legacy-and-unknown.tbx", "validate", "--strict",
	})
}

func TestValidate_LenientWithLegacy_Golden(t *testing.T) {
	runGolden(t, "validate/lenient_with_legacy", []string{
		"terminology", "--tbx", "testdata/fixtures/with-legacy-and-unknown.tbx", "validate",
	})
}

func TestValidate_UnknownElementStrict_Golden(t *testing.T) {
	runGolden(t, "validate/unknown_element_strict", []string{
		"terminology", "--tbx", "testdata/fixtures/with-legacy-and-unknown.tbx", "validate", "--strict",
	})
}

func TestValidate_InvalidPicklist_Golden(t *testing.T) {
	runGolden(t, "validate/invalid_picklist", []string{
		"terminology", "--tbx", "testdata/fixtures/with-invalid-picklist.tbx", "validate",
	})
}

func TestValidate_MissingBody_Golden(t *testing.T) {
	runGolden(t, "validate/missing_body", []string{
		"terminology", "--tbx", "testdata/fixtures/missing-body.tbx", "validate",
	})
}

func TestValidate_DoctypeEntity_Golden(t *testing.T) {
	runGolden(t, "validate/doctype_entity", []string{
		"terminology", "--tbx", "testdata/fixtures/doctype-entity.tbx", "validate",
	})
}

func TestValidate_NestingTooDeep_Golden(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min" xmlns:basic="http://www.tbxinfo.net/ns/basic" xmlns:ling="http://www.tbxinfo.net/ns/linguist">`)
	sb.WriteString(`<tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>`)
	sb.WriteString(`<text><body><conceptEntry id="c1"><langSec xml:lang="en"><termSec><term>`)
	for range 260 {
		sb.WriteString("<x>")
	}
	for range 260 {
		sb.WriteString("</x>")
	}
	sb.WriteString(`</term></termSec></langSec></conceptEntry></body></text></tbx>`)

	dir := t.TempDir()
	tbxPath := filepath.Join(dir, "deep.tbx")
	if err := os.WriteFile(tbxPath, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	runGolden(t, "validate/nesting_too_deep", []string{
		"terminology", "--tbx", tbxPath, "validate",
	})
}

func TestLookup_Success(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "tzimtzum",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	if ok, _ := env["ok"].(bool); !ok {
		t.Errorf("ok = %v, want true", env["ok"])
	}
	if sv, _ := env["schema_version"].(float64); int(sv) != output.SchemaVersion {
		t.Errorf("schema_version = %v, want %d", sv, output.SchemaVersion)
	}

	results, _ := env["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results length = %d, want 1", len(results))
	}

	r := results[0].(map[string]any)
	if cid, _ := r["concept_id"].(string); cid != "tzimtzum" {
		t.Errorf("concept_id = %q, want %q", cid, "tzimtzum")
	}
}

func TestLookup_NotFound_ExitCode1(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for not-found term, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	if ok, _ := env["ok"].(bool); !ok {
		t.Errorf("ok = %v, want true", env["ok"])
	}

	results, _ := env["results"].([]any)
	if len(results) != 0 {
		t.Errorf("results length = %d, want 0", len(results))
	}
}

func TestLookup_NoTBXPath_ExitCode2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "lookup", "tzimtzum",
	})
	if err == nil {
		t.Fatal("expected error for missing --tbx, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}

	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty, got %q", stdout.String())
	}
}

func TestLookup_LangFilter(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "צמצום", "--lang", "he",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	results, _ := env["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results length = %d, want 1", len(results))
	}
}

func TestLookup_LangFilter_NoMatch(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "צמצום", "--lang", "en",
	})
	if err == nil {
		t.Fatal("expected error for not-found term in lang, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}
}

func TestLookup_FieldsProjection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "tzimtzum",
		"--fields", "results.concept_id",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	results, _ := env["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results length = %d, want 1", len(results))
	}

	r := results[0].(map[string]any)
	if _, ok := r["concept_id"]; !ok {
		t.Error("projected output missing concept_id")
	}
	if _, ok := r["languages"]; ok {
		t.Error("projected output should not contain languages")
	}
}

func TestLookup_InvalidField_ExitCode2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "tzimtzum",
		"--fields", "bogus_field",
	})
	if err == nil {
		t.Fatal("expected error for invalid field, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestLookup_AdmittedTermsInResult(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "malkuth",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	results, _ := env["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results length = %d, want 1", len(results))
	}

	r := results[0].(map[string]any)
	if cid, _ := r["concept_id"].(string); cid != "malkhut" {
		t.Errorf("concept_id = %q, want %q", cid, "malkhut")
	}

	langs, _ := r["languages"].(map[string]any)
	en, _ := langs["en"].(map[string]any)
	preferred, _ := en["preferred"].(map[string]any)
	if pt, _ := preferred["term"].(string); pt != "malkhut" {
		t.Errorf("preferred term = %q, want %q", pt, "malkhut")
	}

	admitted, _ := en["admitted"].([]any)
	if len(admitted) != 1 {
		t.Fatalf("admitted length = %d, want 1", len(admitted))
	}
	at := admitted[0].(map[string]any)
	if term, _ := at["term"].(string); term != "malkuth" {
		t.Errorf("admitted term = %q, want %q", term, "malkuth")
	}
}

func TestLookup_MissingTerm_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "lookup"})
	if err == nil {
		t.Fatal("expected error for missing positional TERM, got nil")
	}
}

func TestScan_BasicScan_Exit0WithMatches(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/minimal-dct.tbx",
		"testdata/fixtures/scan-sample.md",
	})

	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0; err = %v", exitCode, err)
	}

	var env output.ScanEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal stdout: %v\nraw: %s", err, stdout.String())
	}

	if !env.OK {
		t.Error("envelope ok = false, want true")
	}
	if env.File != "testdata/fixtures/scan-sample.md" {
		t.Errorf("file = %q, want testdata/fixtures/scan-sample.md", env.File)
	}
	if len(env.Matches) == 0 {
		t.Fatal("expected matches, got none")
	}
	if env.Summary.TotalMatches != len(env.Matches) {
		t.Errorf("summary total_matches = %d, want %d", env.Summary.TotalMatches, len(env.Matches))
	}
	if env.Summary.UniqueConcepts < 1 {
		t.Error("expected at least 1 unique concept")
	}
}

func TestScan_NoTBXPath_Exit2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "scan", "testdata/fixtures/scan-sample.md",
	})
	if err == nil {
		t.Fatal("expected error for missing --tbx, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestScan_MissingFile_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "scan"})
	if err == nil {
		t.Fatal("expected error for missing positional FILE, got nil")
	}
}

func TestScan_FileNotFound_Exit3(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/minimal-dct.tbx",
		"nonexistent.md",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 3 {
		t.Errorf("exit code = %d, want 3", exitCode)
	}
}

func TestScan_LangFilter_OnlyMatchesLanguage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/minimal-dct.tbx",
		"--lang", "en",
		"testdata/fixtures/scan-sample.md",
	})

	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0; err = %v", exitCode, err)
	}

	var env output.ScanEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal stdout: %v", err)
	}

	for _, m := range env.Matches {
		if m.Lang != "en" {
			t.Errorf("match lang = %q, want en (lang filter not applied)", m.Lang)
		}
	}
}

func TestScan_FieldsProjection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/minimal-dct.tbx",
		"--fields", "matches.concept_id,matches.line",
		"testdata/fixtures/scan-sample.md",
	})

	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0; err = %v", exitCode, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	matches, ok := raw["matches"].([]any)
	if !ok || len(matches) == 0 {
		t.Fatal("expected matches in projected output")
	}
	m := matches[0].(map[string]any)
	if _, ok := m["concept_id"]; !ok {
		t.Error("expected concept_id in projected match")
	}
	if _, ok := m["term"]; ok {
		t.Error("term should be excluded by projection")
	}
}

func TestScan_FrontmatterLang(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"testdata/fixtures/scan-frontmatter-he.md",
	})

	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0; err = %v", exitCode, err)
	}

	var env output.ScanEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal stdout: %v\nraw: %s", err, stdout.String())
	}

	for _, m := range env.Matches {
		if m.Lang != "he" {
			t.Errorf("match lang = %q, want he (frontmatter should restrict to Hebrew)", m.Lang)
		}
	}
	if len(env.Matches) == 0 {
		t.Error("expected at least one Hebrew match")
	}
}

func TestScan_FrontmatterOverridesFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"--lang", "en",
		"testdata/fixtures/scan-frontmatter-he.md",
	})

	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0; err = %v", exitCode, err)
	}

	var env output.ScanEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal stdout: %v\nraw: %s", err, stdout.String())
	}

	for _, m := range env.Matches {
		if m.Lang != "he" {
			t.Errorf("match lang = %q, want he (frontmatter should override --lang en)", m.Lang)
		}
	}
	if len(env.Matches) == 0 {
		t.Error("expected at least one Hebrew match")
	}
}

func TestCheck_Clean_Exit0(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-clean.md",
	})

	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0; err = %v", exitCode, err)
	}

	var env output.CheckEnvelope
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("unmarshal stdout: %v\nraw: %s", jsonErr, stdout.String())
	}

	if !env.OK {
		t.Error("envelope ok = false, want true")
	}
	if len(env.Violations) != 0 {
		t.Errorf("violations = %d, want 0", len(env.Violations))
	}
	if env.Summary.Violations != 0 {
		t.Errorf("summary.violations = %d, want 0", env.Summary.Violations)
	}
	if env.Summary.ConceptsChecked < 1 {
		t.Error("expected at least 1 concept checked")
	}

	_ = stderr.String()
}

func TestCheck_Violations_Exit1(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-missing.md",
	})
	if err == nil {
		t.Fatal("expected error for violations, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}

	var env output.CheckEnvelope
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("unmarshal stdout: %v\nraw: %s", jsonErr, stdout.String())
	}

	if env.OK {
		t.Error("envelope ok = true, want false")
	}
	if len(env.Violations) == 0 {
		t.Fatal("expected violations, got none")
	}

	hasMissing := false
	for _, v := range env.Violations {
		if v.Type == "missing" {
			hasMissing = true
		}
	}
	if !hasMissing {
		t.Error("expected a 'missing' violation")
	}

	_ = stderr.String()
}

func TestCheck_ForbiddenVariant(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"--source-lang", "en", "--target-lang", "en",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-forbidden.md",
	})

	var env output.CheckEnvelope
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("unmarshal stdout: %v\nraw: %s", jsonErr, stdout.String())
	}

	hasForbidden := false
	for _, v := range env.Violations {
		if v.Type == "forbidden_variant" {
			hasForbidden = true
		}
	}
	if !hasForbidden {
		t.Error("expected a 'forbidden_variant' violation")
	}

	_ = err
	_ = stderr.String()
}

func TestCheck_FrontmatterLang(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-clean.md",
	})

	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0 (frontmatter lang should be auto-detected)", exitCode)
	}
}

func TestCheck_LanguageRequired_Exit2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source-nolang.md",
		"testdata/fixtures/check-target-clean.md",
	})
	if err == nil {
		t.Fatal("expected error for missing language, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestCheck_NoTBXPath_Exit2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "check", "src.md", "tgt.md",
	})
	if err == nil {
		t.Fatal("expected error for missing --tbx, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestCheck_FileNotFound_Exit3(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"nonexistent.md",
		"testdata/fixtures/check-target-clean.md",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 3 {
		t.Errorf("exit code = %d, want 3", exitCode)
	}
}

func TestCheck_FieldsProjection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"--fields", "violations,summary.violations",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-missing.md",
	})

	if err != nil {
		_ = output.ExitCodeFor(err)
	}

	var raw map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &raw); jsonErr != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", jsonErr, stdout.String())
	}

	if _, ok := raw["violations"]; !ok {
		t.Error("projected output missing 'violations'")
	}
	if _, ok := raw["source"]; ok {
		t.Error("projected output should not contain 'source'")
	}
}

func TestCheck_MissingTarget_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "check", "src.md"})
	if err == nil {
		t.Fatal("expected error for missing positional TGT, got nil")
	}
}

func TestCheck_MissingBothArgs_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "check"})
	if err == nil {
		t.Fatal("expected error for missing positionals SRC TGT, got nil")
	}
}

func TestCheck_Clean_Golden(t *testing.T) {
	runGolden(t, "check/clean", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-clean.md",
	})
}

func TestCheck_CleanFrontmatter_Golden(t *testing.T) {
	runGolden(t, "check/clean_frontmatter", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-clean.md",
	})
}

func TestCheck_Missing_Golden(t *testing.T) {
	runGolden(t, "check/missing", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-missing.md",
	})
}

func TestCheck_ForbiddenVariant_Golden(t *testing.T) {
	runGolden(t, "check/forbidden_variant", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-forbidden.md",
	})
}

func TestCheck_StrictAdmitted_Golden(t *testing.T) {
	runGolden(t, "check/strict_admitted", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"--strict",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-admitted.md",
	})
}

func TestCheck_AdmittedWarning_Golden(t *testing.T) {
	runGolden(t, "check/admitted_warning", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-admitted.md",
	})
}

func TestCheck_LangFromFlags_Golden(t *testing.T) {
	runGolden(t, "check/lang_from_flags", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"--source-lang", "en", "--target-lang", "en",
		"testdata/fixtures/check-source-nolang.md",
		"testdata/fixtures/check-target-nofm.md",
	})
}

func TestCheck_LangRequired_Golden(t *testing.T) {
	runGolden(t, "check/lang_required", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source-nolang.md",
		"testdata/fixtures/check-target-clean.md",
	})
}

func TestCheck_ContextWindow_Golden(t *testing.T) {
	runGolden(t, "check/context_window", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"--context", "40",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-forbidden.md",
	})
}

func TestCheck_Fields_Golden(t *testing.T) {
	runGolden(t, "check/fields", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"--fields", "violations,summary.violations",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-missing.md",
	})
}

func TestCheck_NoTBX_Golden(t *testing.T) {
	runGolden(t, "check/no_tbx", []string{
		"terminology", "check",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-clean.md",
	})
}

func TestCheck_FileNotFound_Golden(t *testing.T) {
	runGolden(t, "check/file_not_found", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"nonexistent.md",
		"testdata/fixtures/check-target-clean.md",
	})
}

func TestCheck_InvalidField_Golden(t *testing.T) {
	runGolden(t, "check/invalid_field", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"--fields", "concpet_id",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-clean.md",
	})
}

func TestCheck_ViolationOrdering_Golden(t *testing.T) {
	runGolden(t, "check/violation_ordering", []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/rich-dct.tbx",
		"testdata/fixtures/check-source.md",
		"testdata/fixtures/check-target-multi-violations.md",
	})
}

func TestExtract_NonexistentInput_ExitCode3(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "extract", "a.md", "b.md"})
	if err == nil {
		t.Fatal("expected error for nonexistent files, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 3 {
		t.Errorf("exit code = %d, want 3", exitCode)
	}
}

func TestExtract_MissingFiles_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "extract"})
	if err == nil {
		t.Fatal("expected error for missing positional FILE..., got nil")
	}
	if got := output.ExitCodeFor(err); got != 2 {
		t.Errorf("ExitCodeFor = %d, want 2", got)
	}
}

func TestExtract_InvalidScript_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "a.md", "--script", "klingon",
	})
	if err == nil {
		t.Fatal("expected error for invalid --script value, got nil")
	}
}

func TestExtract_ValidScript_NonexistentFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "a.md", "--script", "hebrew",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 3 {
		t.Errorf("exit code = %d, want 3", exitCode)
	}
}

func TestExtract_Golden(t *testing.T) {
	runGolden(t, "extract", []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
	})
}

func writePayloadInCWD(t *testing.T, name, content string) string {
	t.Helper()
	dir, err := os.MkdirTemp("testdata", "payload-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write payload %s: %v", name, err)
	}
	return p
}

func TestApply_JSONPayload_AddsConcept(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	payload := `{"concepts":[{"concept_id":"binah","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"binah"}}}}]}`
	payloadFile := writePayloadInCWD(t, "payload.json", payload)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "apply", "--file", payloadFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout not valid JSON: %v\nstdout: %s", jsonErr, stdout.String())
	}

	applied, _ := env["applied"].(map[string]any)
	added, _ := applied["added"].([]any)
	unchanged, _ := applied["unchanged"].([]any)

	if len(added) != 1 || added[0] != "binah" {
		t.Errorf("added = %v, want [binah]", added)
	}
	if len(unchanged) != 0 {
		t.Errorf("unchanged = %v, want []", unchanged)
	}
}

func TestApply_DryRun_DoesNotModifyFile(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")
	before, _ := os.ReadFile(tbxPath)

	payload := `{"concepts":[{"concept_id":"binah","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"binah"}}}}]}`
	payloadFile := writePayloadInCWD(t, "payload.json", payload)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "apply", "--file", payloadFile, "--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	after, _ := os.ReadFile(tbxPath)
	if !bytes.Equal(before, after) {
		t.Error("file was modified during --dry-run")
	}
}

func TestApply_Prune_RemovesAbsentConcepts(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	payload := `{"concepts":[{"concept_id":"binah","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"binah"}}}}]}`
	payloadFile := writePayloadInCWD(t, "payload.json", payload)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "apply", "--file", payloadFile, "--prune",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout not valid JSON: %v", jsonErr)
	}

	applied, _ := env["applied"].(map[string]any)
	added, _ := applied["added"].([]any)
	removed, _ := applied["removed"].([]any)

	if len(added) != 1 || added[0] != "binah" {
		t.Errorf("added = %v, want [binah]", added)
	}
	if len(removed) != 1 || removed[0] != "tzimtzum" {
		t.Errorf("removed = %v, want [tzimtzum]", removed)
	}
}

func TestApply_Idempotent_Unchanged(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	payload := `{"concepts":[{"concept_id":"tzimtzum","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"tzimtzum","administrative_status":"preferredTerm-admn-sts","part_of_speech":"noun"}},"he":{"preferred":{"term":"צמצום","administrative_status":"preferredTerm-admn-sts","part_of_speech":"noun"}}}}]}`
	payloadFile := writePayloadInCWD(t, "payload.json", payload)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "apply", "--file", payloadFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout not valid JSON: %v", jsonErr)
	}

	applied, _ := env["applied"].(map[string]any)
	unchanged, _ := applied["unchanged"].([]any)
	added, _ := applied["added"].([]any)
	updated, _ := applied["updated"].([]any)

	if len(unchanged) != 1 || unchanged[0] != "tzimtzum" {
		t.Errorf("unchanged = %v, want [tzimtzum]", unchanged)
	}
	if len(added) != 0 {
		t.Errorf("added = %v, want []", added)
	}
	if len(updated) != 0 {
		t.Errorf("updated = %v, want []", updated)
	}
}

func TestApply_Update_DetectsModifiedConcept(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	payload := `{"concepts":[{"concept_id":"tzimtzum","subject_field":"mysticism","languages":{"en":{"preferred":{"term":"tzimtzum","administrative_status":"preferredTerm-admn-sts","part_of_speech":"noun"}}}}]}`
	payloadFile := writePayloadInCWD(t, "payload.json", payload)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "apply", "--file", payloadFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout not valid JSON: %v", jsonErr)
	}

	applied, _ := env["applied"].(map[string]any)
	updated, _ := applied["updated"].([]any)

	if len(updated) != 1 || updated[0] != "tzimtzum" {
		t.Errorf("updated = %v, want [tzimtzum]", updated)
	}
}

func TestApply_Transaction_ExitsZero(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	payload := `{"concepts":[{"concept_id":"binah","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"binah"}}}}]}`
	payloadFile := writePayloadInCWD(t, "payload.json", payload)

	t.Setenv("TERMINOLOGY_AUTHOR", "Andre")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "apply", "--file", payloadFile,
		"--transaction",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout not valid JSON: %v", jsonErr)
	}

	ok, _ := env["ok"].(bool)
	if !ok {
		t.Error("ok = false, want true")
	}
}

func TestApply_TBXFragment_AddsConcept(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	fragment := `<conceptEntry id="binah">
  <min:subjectField>kabbalah</min:subjectField>
  <langSec xml:lang="en">
    <termSec>
      <term>binah</term>
      <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
    </termSec>
  </langSec>
</conceptEntry>`

	payloadFile := writePayloadInCWD(t, "payload.tbx", fragment)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "apply", "--file", payloadFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout not valid JSON: %v", jsonErr)
	}

	applied, _ := env["applied"].(map[string]any)
	added, _ := applied["added"].([]any)

	if len(added) != 1 || added[0] != "binah" {
		t.Errorf("added = %v, want [binah]", added)
	}
}

func TestApply_Prune_DanglingCrossref_Fails(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/crossref-dct.tbx")

	payload := `{"concepts":[{"concept_id":"sefirot","subject_field":"kabbalah","cross_refs":[{"target":"tzimtzum","label":"related concept"}],"languages":{"en":{"preferred":{"term":"sefirot"}}}}]}`
	payloadFile := writePayloadInCWD(t, "payload.json", payload)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "apply", "--file", payloadFile, "--prune",
	})
	if err == nil {
		t.Fatal("expected error for dangling crossref, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 65 {
		t.Errorf("exit code = %d, want 65", exitCode)
	}
}

func TestApply_NoTBXPath_ExitsTwo(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	payload := `{"concepts":[]}`
	payloadFile := writePayloadInCWD(t, "payload.json", payload)

	err := cmd.Run(context.Background(), []string{
		"terminology", "apply", "--file", payloadFile,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if exitCode := output.ExitCodeFor(err); exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestApply_MissingFile_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "apply"})
	if err == nil {
		t.Fatal("expected error for missing required --file flag, got nil")
	}
}

func TestApply_NonexistentFile_ExitCode3(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	tbxFile := filepath.Join(t.TempDir(), "glossary.tbx")
	if err := os.WriteFile(tbxFile, []byte(`<?xml version="1.0"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2">
<tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
<text><body></body></text></tbx>`), 0o644); err != nil {
		t.Fatal(err)
	}

	err := cmd.Run(context.Background(), []string{
		"terminology", "apply",
		"--tbx", tbxFile,
		"--file", "testdata/nonexistent-payload.json",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent payload file, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 3 {
		t.Errorf("exit code = %d, want 3", exitCode)
	}
}

func TestSchema_FullOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "schema"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	if sv, _ := env["schema_version"].(float64); int(sv) != output.SchemaVersion {
		t.Errorf("schema_version = %v, want %d", sv, output.SchemaVersion)
	}

	for _, key := range []string{"commands", "envelopes", "error_codes"} {
		if _, ok := env[key]; !ok {
			t.Errorf("missing top-level key %q", key)
		}
	}

	cmds, _ := env["commands"].([]any)
	if len(cmds) == 0 {
		t.Error("commands array is empty")
	}

	errCodes, _ := env["error_codes"].([]any)
	if len(errCodes) == 0 {
		t.Error("error_codes array is empty")
	}

	envs, _ := env["envelopes"].(map[string]any)
	if len(envs) == 0 {
		t.Error("envelopes object is empty")
	}
}

func TestConceptAdd_Stub_Golden(t *testing.T) {
	runGolden(t, "concept_add", []string{"terminology", "concept", "add"})
}

func TestConceptAdd_WithFlags_HappyPath(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "add",
		"--lang", "es", "--term", "sefirot", "--status", "preferredTerm-admn-sts",
		"--subject-field", "kabbalah",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if jErr := json.Unmarshal(stdout.Bytes(), &env); jErr != nil {
		t.Fatalf("unmarshal: %v\nstdout: %s", jErr, stdout.String())
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	if env.Result.ConceptID != "sefirot" {
		t.Errorf("concept_id = %q, want %q", env.Result.ConceptID, "sefirot")
	}
	if env.Result.SubjectField != "kabbalah" {
		t.Errorf("subject_field = %q, want %q", env.Result.SubjectField, "kabbalah")
	}
	grp, ok := env.Result.Languages["es"]
	if !ok || grp.Preferred == nil {
		t.Fatal("expected es language with preferred term")
	}
	if grp.Preferred.Term != "sefirot" {
		t.Errorf("preferred term = %q, want %q", grp.Preferred.Term, "sefirot")
	}
}

func TestConceptAdd_DuplicateID_Errors(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "add",
		"--id", "tzimtzum",
		"--lang", "en", "--term", "tzimtzum",
	})
	if err == nil {
		t.Fatal("expected duplicate_id error, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected coded error, got %T: %v", err, err)
	}
	if coded.Code() != "duplicate_id" {
		t.Errorf("code = %q, want %q", coded.Code(), "duplicate_id")
	}
}

func TestConceptAdd_DryRun_DoesNotSave(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "add",
		"--lang", "en", "--term", "newterm", "--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env output.WriteEnvelope
	if jErr := json.Unmarshal(stdout.Bytes(), &env); jErr != nil {
		t.Fatalf("unmarshal: %v", jErr)
	}
	if env.Result.ConceptID != "newterm" {
		t.Errorf("concept_id = %q, want %q", env.Result.ConceptID, "newterm")
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	for _, c := range g.Concepts {
		if c.ID == "newterm" {
			t.Error("dry-run should not persist the new concept")
		}
	}
}

func TestConceptAdd_ExplicitID(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "add",
		"--id", "my-custom-id",
		"--lang", "en", "--term", "Custom Term",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if jErr := json.Unmarshal(stdout.Bytes(), &env); jErr != nil {
		t.Fatalf("unmarshal: %v", jErr)
	}
	if env.Result.ConceptID != "my-custom-id" {
		t.Errorf("concept_id = %q, want %q", env.Result.ConceptID, "my-custom-id")
	}
}

func TestConceptAdd_MissingLangTerm_Errors(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "add",
	})
	if err == nil {
		t.Fatal("expected error for missing --lang/--term, got nil")
	}
}

func TestConceptAdd_Transaction(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "add",
		"--lang", "en", "--term", "sefirot",
		"--transaction", "--author", "test-author",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}

	var found bool
	for _, c := range g.Concepts {
		if c.ID == "sefirot" {
			found = true
			if len(c.Transactions) == 0 {
				t.Error("expected transaction record on concept")
			} else {
				txn := c.Transactions[0]
				if txn.Type != "modification" {
					t.Errorf("transaction type = %q, want %q", txn.Type, "modification")
				}
				if txn.Responsibility != "test-author" {
					t.Errorf("responsibility = %q, want %q", txn.Responsibility, "test-author")
				}
			}
		}
	}
	if !found {
		t.Error("concept 'sefirot' not found after add")
	}
}

func TestConceptAdd_PersistedToFile(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "add",
		"--lang", "en", "--term", "sefirot",
		"--part-of-speech", "noun",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}

	if len(g.Concepts) != 2 {
		t.Fatalf("got %d concepts, want 2", len(g.Concepts))
	}

	var found *tbx.Concept
	for i := range g.Concepts {
		if g.Concepts[i].ID == "sefirot" {
			found = &g.Concepts[i]
			break
		}
	}
	if found == nil {
		t.Fatal("concept 'sefirot' not found in saved file")
	}

	en, ok := found.Languages["en"]
	if !ok {
		t.Fatal("expected en language section")
	}
	if len(en.Terms) != 1 {
		t.Fatalf("expected 1 term, got %d", len(en.Terms))
	}
	if en.Terms[0].Surface != "sefirot" {
		t.Errorf("term = %q, want %q", en.Terms[0].Surface, "sefirot")
	}
	if en.Terms[0].PartOfSpeech != "noun" {
		t.Errorf("part_of_speech = %q, want %q", en.Terms[0].PartOfSpeech, "noun")
	}
}

func copyTBXFixture(t *testing.T, src string) string {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read fixture %s: %v", src, err)
	}
	dst := filepath.Join(t.TempDir(), "test.tbx")
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write fixture copy: %v", err)
	}
	return dst
}

func TestConceptAdd_InvalidStatus_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "concept", "add", "--status", "nope",
	})
	if err == nil {
		t.Fatal("expected error for invalid --status value, got nil")
	}
}

func TestConceptAdd_InvalidPartOfSpeech_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "concept", "add", "--part-of-speech", "frobnicator",
	})
	if err == nil {
		t.Fatal("expected error for invalid --part-of-speech value, got nil")
	}
}

func TestConcept_BareInvocation_ExitsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "concept"})
	if err == nil {
		t.Fatal("expected error for bare concept invocation, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestConcept_BareInvocation_Golden(t *testing.T) {
	runGolden(t, "concept", []string{"terminology", "concept"})
}

func TestTerm_BareInvocation_ExitsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "term"})
	if err == nil {
		t.Fatal("expected error for bare term invocation, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestTerm_BareInvocation_Golden(t *testing.T) {
	runGolden(t, "term", []string{"terminology", "term"})
}

func TestTermAdd_HappyPath_ExistingLangSec(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "add", "tzimtzum",
		"--lang", "en", "--term", "contraction",
		"--status", "admittedTerm-admn-sts",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	if env.Result.ConceptID != "tzimtzum" {
		t.Errorf("concept_id = %q, want %q", env.Result.ConceptID, "tzimtzum")
	}
	enGrp, ok := env.Result.Languages["en"]
	if !ok {
		t.Fatal("missing en language group")
	}
	if len(enGrp.Admitted) != 1 {
		t.Fatalf("admitted count = %d, want 1", len(enGrp.Admitted))
	}
	if enGrp.Admitted[0].Term != "contraction" {
		t.Errorf("admitted term = %q, want %q", enGrp.Admitted[0].Term, "contraction")
	}
}

func TestTermAdd_CreatesLangSec(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "add", "tzimtzum",
		"--lang", "es", "--term", "tzimtzum",
		"--status", "preferredTerm-admn-sts",
		"--part-of-speech", "noun",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	esGrp, ok := env.Result.Languages["es"]
	if !ok {
		t.Fatal("missing es language group — langSec was not created")
	}
	if esGrp.Preferred == nil {
		t.Fatal("expected preferred term in es")
	}
	if esGrp.Preferred.Term != "tzimtzum" {
		t.Errorf("preferred term = %q, want %q", esGrp.Preferred.Term, "tzimtzum")
	}
}

func TestTermAdd_NotFound(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "add", "nonexistent",
		"--lang", "en", "--term", "test",
	})
	if err == nil {
		t.Fatal("expected not_found error, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 65 {
		t.Errorf("exit code = %d, want 65", exitCode)
	}
}

func TestTermAdd_DryRun(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "add", "tzimtzum",
		"--lang", "es", "--term", "tzimtzum",
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env output.WriteEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	if _, ok := env.Result.Languages["es"]; !ok {
		t.Error("preview should include es langSec")
	}

	g, _, err := tbx.Load(tbxPath)
	if err != nil {
		t.Fatalf("reload TBX: %v", err)
	}
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			if _, ok := c.Languages["es"]; ok {
				t.Error("dry-run should not modify the file")
			}
		}
	}
}

func TestTermAdd_WithTransaction(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "add", "tzimtzum",
		"--lang", "es", "--term", "tzimtzum",
		"--transaction", "--author", "Andre",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g, _, err := tbx.Load(tbxPath)
	if err != nil {
		t.Fatalf("reload TBX: %v", err)
	}
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			ls, ok := c.Languages["es"]
			if !ok {
				t.Fatal("missing es langSec")
			}
			if len(ls.Terms) == 0 {
				t.Fatal("no terms in es")
			}
			lastTerm := ls.Terms[len(ls.Terms)-1]
			if len(lastTerm.Transactions) == 0 {
				t.Fatal("expected transaction on new term")
			}
			txn := lastTerm.Transactions[0]
			if txn.Responsibility != "Andre" {
				t.Errorf("responsibility = %q, want %q", txn.Responsibility, "Andre")
			}
			if txn.Type != "modification" {
				t.Errorf("transaction type = %q, want %q", txn.Type, "modification")
			}
			return
		}
	}
	t.Fatal("concept tzimtzum not found after write")
}

func TestTermAdd_WithAllFlags(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "add", "tzimtzum",
		"--lang", "es", "--term", "tzimtzum",
		"--status", "preferredTerm-admn-sts",
		"--part-of-speech", "noun",
		"--register", "technicalRegister",
		"--grammatical-gender", "masculine",
		"-n", "--transaction", "-a", "Andre",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	esGrp, ok := env.Result.Languages["es"]
	if !ok {
		t.Fatal("missing es language group")
	}
	if esGrp.Preferred == nil {
		t.Fatal("expected preferred term")
	}
	if esGrp.Preferred.PartOfSpeech != "noun" {
		t.Errorf("part_of_speech = %q, want %q", esGrp.Preferred.PartOfSpeech, "noun")
	}
	if esGrp.Preferred.Register != "technicalRegister" {
		t.Errorf("register = %q, want %q", esGrp.Preferred.Register, "technicalRegister")
	}
	if esGrp.Preferred.GrammaticalGender != "masculine" {
		t.Errorf("grammatical_gender = %q, want %q", esGrp.Preferred.GrammaticalGender, "masculine")
	}
}

func TestTermAdd_MissingLang_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "term", "add", "tzimtzum", "--term", "tzimtzum",
	})
	if err == nil {
		t.Fatal("expected error for missing --lang, got nil")
	}
}

func TestTermAdd_MissingTerm_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "term", "add", "tzimtzum", "--lang", "es",
	})
	if err == nil {
		t.Fatal("expected error for missing --term, got nil")
	}
}

func TestTermAdd_MissingID_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "term", "add", "--lang", "es", "--term", "tzimtzum",
	})
	if err == nil {
		t.Fatal("expected error for missing positional ID, got nil")
	}
}

func TestTermAdd_InvalidStatus_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "term", "add", "tzimtzum",
		"--lang", "es", "--term", "tzimtzum", "--status", "nope",
	})
	if err == nil {
		t.Fatal("expected error for invalid --status value, got nil")
	}
}

func TestTermAdd_NoTBXPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "term", "add", "tzimtzum",
		"--lang", "es", "--term", "tzimtzum",
	})
	if err == nil {
		t.Fatal("expected error for missing TBX path, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestTermAdd_IDStability(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "add", "tzimtzum",
		"--lang", "en", "--term", "new-preferred",
		"--status", "preferredTerm-admn-sts",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env output.WriteEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if env.Result.ConceptID != "tzimtzum" {
		t.Errorf("concept ID changed to %q, want %q", env.Result.ConceptID, "tzimtzum")
	}
}

func TestConceptUpdate_Replace_HappyPath(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--replace",
		"--lang", "es", "--term", "contracción", "--status", "preferredTerm-admn-sts",
		"--subject-field", "mysticism",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if jErr := json.Unmarshal(stdout.Bytes(), &env); jErr != nil {
		t.Fatalf("unmarshal: %v\nstdout: %s", jErr, stdout.String())
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	if env.Result.ConceptID != "tzimtzum" {
		t.Errorf("concept_id = %q, want %q", env.Result.ConceptID, "tzimtzum")
	}
	if env.Result.SubjectField != "mysticism" {
		t.Errorf("subject_field = %q, want %q", env.Result.SubjectField, "mysticism")
	}
	if _, ok := env.Result.Languages["en"]; ok {
		t.Error("replace should have removed the en langSec")
	}
	if _, ok := env.Result.Languages["he"]; ok {
		t.Error("replace should have removed the he langSec")
	}
	grp, ok := env.Result.Languages["es"]
	if !ok || grp.Preferred == nil {
		t.Fatal("expected es language with preferred term")
	}
	if grp.Preferred.Term != "contracción" {
		t.Errorf("preferred term = %q, want %q", grp.Preferred.Term, "contracción")
	}
}

func TestConceptUpdate_Replace_PersistedToFile(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--replace",
		"--lang", "es", "--term", "contracción",
		"--subject-field", "mysticism",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	var found bool
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			found = true
			if c.SubjectField != "mysticism" {
				t.Errorf("subject_field = %q, want %q", c.SubjectField, "mysticism")
			}
			if _, ok := c.Languages["en"]; ok {
				t.Error("en langSec should have been replaced away")
			}
			ls, ok := c.Languages["es"]
			if !ok {
				t.Fatal("expected es langSec")
			}
			if len(ls.Terms) != 1 || ls.Terms[0].Surface != "contracción" {
				t.Errorf("expected single term 'contracción', got %+v", ls.Terms)
			}
		}
	}
	if !found {
		t.Error("concept 'tzimtzum' not found after update")
	}
}

func TestConceptUpdate_Merge_HappyPath(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--merge",
		"--lang", "es", "--term", "tzimtzum", "--status", "preferredTerm-admn-sts",
		"--subject-field", "mysticism",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if jErr := json.Unmarshal(stdout.Bytes(), &env); jErr != nil {
		t.Fatalf("unmarshal: %v\nstdout: %s", jErr, stdout.String())
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	if env.Result.ConceptID != "tzimtzum" {
		t.Errorf("concept_id = %q, want %q", env.Result.ConceptID, "tzimtzum")
	}
	if env.Result.SubjectField != "mysticism" {
		t.Errorf("subject_field = %q, want %q", env.Result.SubjectField, "mysticism")
	}
	if _, ok := env.Result.Languages["en"]; !ok {
		t.Error("merge should preserve the existing en langSec")
	}
	if _, ok := env.Result.Languages["he"]; !ok {
		t.Error("merge should preserve the existing he langSec")
	}
	grp, ok := env.Result.Languages["es"]
	if !ok || grp.Preferred == nil {
		t.Fatal("expected es language with preferred term after merge")
	}
	if grp.Preferred.Term != "tzimtzum" {
		t.Errorf("preferred term = %q, want %q", grp.Preferred.Term, "tzimtzum")
	}
}

func TestConceptUpdate_Merge_PreservesUnspecifiedFields(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--merge",
		"--lang", "es", "--term", "tzimtzum",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			if c.SubjectField != "kabbalah" {
				t.Errorf("merge without --subject-field should preserve original, got %q", c.SubjectField)
			}
			if _, ok := c.Languages["en"]; !ok {
				t.Error("merge should preserve existing en langSec")
			}
			if _, ok := c.Languages["he"]; !ok {
				t.Error("merge should preserve existing he langSec")
			}
			return
		}
	}
	t.Error("concept 'tzimtzum' not found after update")
}

func TestConceptUpdate_Merge_TermNaturalKey(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--merge",
		"--lang", "en", "--term", "tzimtzum", "--status", "preferredTerm-admn-sts",
		"--part-of-speech", "verb",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			ls := c.Languages["en"]
			if len(ls.Terms) != 1 {
				t.Fatalf("expected 1 term (merged by natural key), got %d", len(ls.Terms))
			}
			if ls.Terms[0].PartOfSpeech != "verb" {
				t.Errorf("part_of_speech = %q, want %q", ls.Terms[0].PartOfSpeech, "verb")
			}
			return
		}
	}
	t.Error("concept 'tzimtzum' not found")
}

func TestConceptUpdate_Merge_TermAppended(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--merge",
		"--lang", "en", "--term", "contraction", "--status", "deprecatedTerm-admn-sts",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			ls := c.Languages["en"]
			if len(ls.Terms) != 2 {
				t.Fatalf("expected 2 terms (original + appended), got %d", len(ls.Terms))
			}
			return
		}
	}
	t.Error("concept 'tzimtzum' not found")
}

func TestConceptUpdate_NotFound_Errors(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "nonexistent",
		"--replace",
		"--lang", "en", "--term", "foo",
	})
	if err == nil {
		t.Fatal("expected not_found error, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected coded error, got %T: %v", err, err)
	}
	if coded.Code() != "not_found" {
		t.Errorf("code = %q, want %q", coded.Code(), "not_found")
	}
}

func TestConceptUpdate_IDStability(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--replace",
		"--lang", "en", "--term", "completely-different-term",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if jErr := json.Unmarshal(stdout.Bytes(), &env); jErr != nil {
		t.Fatalf("unmarshal: %v", jErr)
	}
	if env.Result.ConceptID != "tzimtzum" {
		t.Errorf("ID changed to %q, want %q (ID stability)", env.Result.ConceptID, "tzimtzum")
	}
}

func TestConceptUpdate_DryRun_DoesNotSave(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--replace", "--dry-run",
		"--lang", "es", "--term", "contracción",
		"--subject-field", "mysticism",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env output.WriteEnvelope
	if jErr := json.Unmarshal(stdout.Bytes(), &env); jErr != nil {
		t.Fatalf("unmarshal: %v", jErr)
	}
	if env.Result.SubjectField != "mysticism" {
		t.Errorf("dry-run preview should show new subject_field, got %q", env.Result.SubjectField)
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			if c.SubjectField != "kabbalah" {
				t.Error("dry-run should not persist changes to file")
			}
			return
		}
	}
	t.Error("concept 'tzimtzum' not found")
}

func TestConceptUpdate_Transaction(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--replace",
		"--lang", "en", "--term", "tzimtzum",
		"--transaction", "--author", "test-author",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			if len(c.Transactions) == 0 {
				t.Error("expected transaction record on concept")
			} else {
				txn := c.Transactions[0]
				if txn.Type != "modification" {
					t.Errorf("transaction type = %q, want %q", txn.Type, "modification")
				}
				if txn.Responsibility != "test-author" {
					t.Errorf("responsibility = %q, want %q", txn.Responsibility, "test-author")
				}
			}
			return
		}
	}
	t.Error("concept 'tzimtzum' not found")
}

func TestConceptUpdate_JSONStdin(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	jsonPayload := `{
		"concept_id": "tzimtzum",
		"subject_field": "mysticism",
		"languages": {
			"es": {
				"preferred": {"term": "contracción"}
			}
		}
	}`

	origStdin := os.Stdin
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(jsonPayload)
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--replace",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if jErr := json.Unmarshal(stdout.Bytes(), &env); jErr != nil {
		t.Fatalf("unmarshal: %v", jErr)
	}
	if env.Result.SubjectField != "mysticism" {
		t.Errorf("subject_field = %q, want %q", env.Result.SubjectField, "mysticism")
	}
	grp, ok := env.Result.Languages["es"]
	if !ok || grp.Preferred == nil {
		t.Fatal("expected es language with preferred term")
	}
	if grp.Preferred.Term != "contracción" {
		t.Errorf("preferred term = %q, want %q", grp.Preferred.Term, "contracción")
	}
}

func TestConceptUpdate_Both_Golden(t *testing.T) {
	runGolden(t, "concept_update/both", []string{"terminology", "concept", "update", "tzimtzum", "--merge", "--replace"})
}

func TestConceptUpdate_Neither_Golden(t *testing.T) {
	runGolden(t, "concept_update/neither", []string{"terminology", "concept", "update", "tzimtzum"})
}

func TestConceptUpdate_MissingID_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "concept", "update", "--merge"})
	if err == nil {
		t.Fatal("expected error for missing positional ID, got nil")
	}
}

func TestConceptUpdate_BothFlags_ExitsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "concept", "update", "tzimtzum", "--merge", "--replace",
	})
	if err == nil {
		t.Fatal("expected error for both --merge and --replace, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestConceptUpdate_NeitherFlag_ExitsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "concept", "update", "tzimtzum",
	})
	if err == nil {
		t.Fatal("expected error when neither --merge nor --replace supplied, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestConceptUpdate_MissingLangTerm_Errors(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--replace",
	})
	if err == nil {
		t.Fatal("expected error for missing --lang/--term, got nil")
	}
}

func TestConceptRemove_HappyPath(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "remove", "tzimtzum",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty, got %q", stderr.String())
	}

	var env output.WriteEnvelope
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("unmarshal stdout: %v\nraw: %s", jsonErr, stdout.String())
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	if env.Result.ConceptID != "tzimtzum" {
		t.Errorf("concept_id = %q, want %q", env.Result.ConceptID, "tzimtzum")
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	if len(g.Concepts) != 0 {
		t.Errorf("expected 0 concepts after remove, got %d", len(g.Concepts))
	}
}

func TestConceptRemove_NotFound(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "remove", "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 65 {
		t.Errorf("exit code = %d, want 65", exitCode)
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "not_found" {
		t.Errorf("code = %q, want %q", coded.Code(), "not_found")
	}
}

func TestConceptRemove_DanglingCrossref(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/crossref-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "remove", "tzimtzum",
	})
	if err == nil {
		t.Fatal("expected dangling_crossref error, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "dangling_crossref" {
		t.Errorf("code = %q, want %q", coded.Code(), "dangling_crossref")
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	if len(g.Concepts) != 2 {
		t.Errorf("expected 2 concepts (unchanged), got %d", len(g.Concepts))
	}
}

func TestConceptRemove_Force(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/crossref-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "remove", "tzimtzum", "--force",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	if len(g.Concepts) != 1 {
		t.Errorf("expected 1 concept after force remove, got %d", len(g.Concepts))
	}
	if g.Concepts[0].ID != "sefirot" {
		t.Errorf("remaining concept = %q, want %q", g.Concepts[0].ID, "sefirot")
	}
}

func TestConceptRemove_Force_ThenValidate_ShowsUnresolvedCrossref(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/crossref-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "remove", "tzimtzum", "--force",
	})
	if err != nil {
		t.Fatalf("remove: %v\nstderr: %s", err, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	cmd2 := app.Root()
	cmd2.Writer = &stdout
	cmd2.ErrWriter = &stderr

	err = cmd2.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "validate",
	})

	exitCode := output.ExitCodeFor(err)
	if exitCode != 1 {
		t.Errorf("validate exit code = %d, want 1 (warnings present)", exitCode)
	}

	out := stdout.String()
	if !strings.Contains(out, "unresolved_crossref") {
		t.Errorf("expected unresolved_crossref warning in output:\n%s", out)
	}
}

func TestConceptRemove_DryRun(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "remove", "tzimtzum", "--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("unmarshal stdout: %v", jsonErr)
	}
	if env.Result.ConceptID != "tzimtzum" {
		t.Errorf("concept_id = %q, want %q", env.Result.ConceptID, "tzimtzum")
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	if len(g.Concepts) != 1 {
		t.Errorf("dry-run should not save; got %d concepts, want 1", len(g.Concepts))
	}
}

func TestConceptRemove_MissingID_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "concept", "remove"})
	if err == nil {
		t.Fatal("expected error for missing positional ID, got nil")
	}
}

func TestConceptRemove_Transaction(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "remove", "tzimtzum",
		"--dry-run", "--transaction", "--author", "Andre",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "tzimtzum") {
		t.Errorf("expected concept in output:\n%s", out)
	}
}

func TestConceptRemove_AuthorViaEnv(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")
	t.Setenv("TERMINOLOGY_AUTHOR", "Andre")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "remove", "tzimtzum",
		"--transaction",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	if !strings.Contains(stdout.String(), "tzimtzum") {
		t.Error("expected concept in output")
	}
}

func TestConceptUpdate_AuthorViaEnv(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")
	t.Setenv("TERMINOLOGY_AUTHOR", "Andre")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "concept", "update", "tzimtzum",
		"--replace", "--transaction",
		"--lang", "en", "--term", "tzimtzum",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	g, _, loadErr := tbx.Load(tbxPath)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			if len(c.Transactions) == 0 {
				t.Fatal("expected transaction record")
			}
			if c.Transactions[0].Responsibility != "Andre" {
				t.Errorf("responsibility = %q, want %q", c.Transactions[0].Responsibility, "Andre")
			}
			return
		}
	}
	t.Error("concept 'tzimtzum' not found")
}

func TestSchema_CommandsPopulated(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "schema"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v", jsonErr)
	}

	cmds, _ := env["commands"].([]any)
	names := make(map[string]bool, len(cmds))
	for _, c := range cmds {
		cm := c.(map[string]any)
		names[cm["name"].(string)] = true
	}

	for _, want := range []string{"validate", "lookup", "schema", "extract", "scan", "check"} {
		if !names[want] {
			t.Errorf("missing command %q in schema output", want)
		}
	}
}

func TestSchema_FilteredCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "schema", "--command", "validate",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	if name, _ := env["name"].(string); name != "validate" {
		t.Errorf("name = %q, want %q", name, "validate")
	}

	if sv, _ := env["schema_version"].(float64); int(sv) != output.SchemaVersion {
		t.Errorf("schema_version = %v, want %d", sv, output.SchemaVersion)
	}

	if _, ok := env["flags"]; !ok {
		t.Error("missing 'flags' key in filtered output")
	}

	if _, ok := env["envelope"]; !ok {
		t.Error("missing 'envelope' key in filtered output")
	}

	exitCodesRaw, ok := env["exit_codes"]
	if !ok {
		t.Fatal("missing 'exit_codes' key in filtered output")
	}
	exitCodesArr, ok := exitCodesRaw.([]any)
	if !ok {
		t.Fatalf("exit_codes is not an array: %T", exitCodesRaw)
	}
	codes := make(map[int]bool, len(exitCodesArr))
	for _, v := range exitCodesArr {
		codes[int(v.(float64))] = true
	}
	for _, want := range []int{0, 1, 2, 65} {
		if !codes[want] {
			t.Errorf("exit_codes missing %d", want)
		}
	}
}

func TestSchema_FilteredCommand_ExitCodesPresent(t *testing.T) {
	for _, name := range []string{"validate", "lookup", "extract", "schema"} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := app.Root()
			cmd.Writer = &stdout
			cmd.ErrWriter = &stderr

			err := cmd.Run(context.Background(), []string{
				"terminology", "schema", "--command", name,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var env map[string]any
			if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
				t.Fatalf("stdout is not valid JSON: %v", jsonErr)
			}

			exitCodesRaw, ok := env["exit_codes"]
			if !ok {
				t.Fatal("missing 'exit_codes' key")
			}
			exitCodesArr, ok := exitCodesRaw.([]any)
			if !ok || len(exitCodesArr) == 0 {
				t.Fatal("exit_codes is empty or not an array")
			}
		})
	}
}

func TestSchema_UnknownCommand_ExitCode2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "schema", "--command", "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for unknown --command, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}

	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty, got %q", stdout.String())
	}
}

func TestSchema_ErrorCodesEnumerated(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "schema"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v", jsonErr)
	}

	errCodes, _ := env["error_codes"].([]any)
	codes := make(map[string]bool, len(errCodes))
	for _, ec := range errCodes {
		ecm := ec.(map[string]any)
		codes[ecm["code"].(string)] = true
	}

	for _, want := range []string{"validation_error", "no_tbx_path", "invalid_field"} {
		if !codes[want] {
			t.Errorf("missing error code %q in schema output", want)
		}
	}
}

func TestSchema_EnvelopesPopulated(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "schema"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v", jsonErr)
	}

	envs, _ := env["envelopes"].(map[string]any)
	for _, want := range []string{"validate", "lookup", "extract"} {
		if _, ok := envs[want]; !ok {
			t.Errorf("missing envelope %q in schema output", want)
		}
	}
}

func TestSchema_NoTBXRequired(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "schema"})
	if err != nil {
		t.Fatalf("schema should not require --tbx, got error: %v", err)
	}
}

func TestTermDeprecate_HappyPath(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "deprecate", "tzimtzum",
		"--lang", "en", "--term", "tzimtzum",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	var env output.WriteEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	if env.Result.ConceptID != "tzimtzum" {
		t.Errorf("concept_id = %q, want %q", env.Result.ConceptID, "tzimtzum")
	}
	enGrp, ok := env.Result.Languages["en"]
	if !ok {
		t.Fatal("missing en language group")
	}
	if len(enGrp.Deprecated) != 1 {
		t.Fatalf("deprecated count = %d, want 1", len(enGrp.Deprecated))
	}
	if enGrp.Deprecated[0].Term != "tzimtzum" {
		t.Errorf("deprecated term = %q, want %q", enGrp.Deprecated[0].Term, "tzimtzum")
	}
	if enGrp.Deprecated[0].AdministrativeStatus != "deprecatedTerm-admn-sts" {
		t.Errorf("status = %q, want %q", enGrp.Deprecated[0].AdministrativeStatus, "deprecatedTerm-admn-sts")
	}
}

func TestTermDeprecate_MissingLang_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "term", "deprecate", "tzimtzum", "--term", "contraction",
	})
	if err == nil {
		t.Fatal("expected error for missing --lang, got nil")
	}
}

func TestTermDeprecate_MissingTerm_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "term", "deprecate", "tzimtzum", "--lang", "es",
	})
	if err == nil {
		t.Fatal("expected error for missing --term, got nil")
	}
}

func TestTermDeprecate_MissingID_Errors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "term", "deprecate", "--lang", "es", "--term", "contraction",
	})
	if err == nil {
		t.Fatal("expected error for missing positional ID, got nil")
	}
}

func TestTermDeprecate_NotFound_Concept(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "deprecate", "nonexistent",
		"--lang", "en", "--term", "test",
	})
	if err == nil {
		t.Fatal("expected not_found error, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 65 {
		t.Errorf("exit code = %d, want 65", exitCode)
	}
}

func TestTermDeprecate_NotFound_Lang(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "deprecate", "tzimtzum",
		"--lang", "fr", "--term", "test",
	})
	if err == nil {
		t.Fatal("expected not_found error, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 65 {
		t.Errorf("exit code = %d, want 65", exitCode)
	}
}

func TestTermDeprecate_NotFound_Term(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "deprecate", "tzimtzum",
		"--lang", "en", "--term", "nonexistent",
	})
	if err == nil {
		t.Fatal("expected not_found error, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 65 {
		t.Errorf("exit code = %d, want 65", exitCode)
	}
}

func TestTermDeprecate_DryRun(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "deprecate", "tzimtzum",
		"--lang", "en", "--term", "tzimtzum",
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env output.WriteEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	enGrp, ok := env.Result.Languages["en"]
	if !ok {
		t.Fatal("missing en language group in preview")
	}
	if len(enGrp.Deprecated) != 1 {
		t.Fatalf("preview deprecated count = %d, want 1", len(enGrp.Deprecated))
	}

	g, _, err := tbx.Load(tbxPath)
	if err != nil {
		t.Fatalf("reload TBX: %v", err)
	}
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			ls := c.Languages["en"]
			for _, term := range ls.Terms {
				if term.Surface == "tzimtzum" && term.AdministrativeStatus == tbx.StatusDeprecated {
					t.Error("dry-run should not modify the file")
				}
			}
		}
	}
}

func TestTermDeprecate_WithTransaction(t *testing.T) {
	tbxPath := copyTBXFixture(t, "testdata/fixtures/minimal-dct.tbx")

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", tbxPath, "term", "deprecate", "tzimtzum",
		"--lang", "en", "--term", "tzimtzum",
		"--transaction", "--author", "Andre",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	g, _, err := tbx.Load(tbxPath)
	if err != nil {
		t.Fatalf("reload TBX: %v", err)
	}
	for _, c := range g.Concepts {
		if c.ID == "tzimtzum" {
			ls := c.Languages["en"]
			for _, term := range ls.Terms {
				if term.Surface == "tzimtzum" {
					if term.AdministrativeStatus != tbx.StatusDeprecated {
						t.Errorf("status = %v, want StatusDeprecated", term.AdministrativeStatus)
					}
					if len(term.Transactions) == 0 {
						t.Fatal("expected transaction on deprecated term")
					}
					txn := term.Transactions[len(term.Transactions)-1]
					if txn.Type != "modification" {
						t.Errorf("transaction type = %q, want %q", txn.Type, "modification")
					}
					if txn.Responsibility != "Andre" {
						t.Errorf("responsibility = %q, want %q", txn.Responsibility, "Andre")
					}
					return
				}
			}
			t.Fatal("term 'tzimtzum' not found after deprecation")
		}
	}
	t.Fatal("concept 'tzimtzum' not found")
}

func TestArgBounds_ExitCode2(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"lookup_missing", []string{"terminology", "lookup"}},
		{"lookup_excess", []string{"terminology", "lookup", "tzimtzum", "extra"}},
		{"scan_missing", []string{"terminology", "scan"}},
		{"scan_excess", []string{"terminology", "scan", "a.md", "b.md"}},
		{"check_excess", []string{"terminology", "check", "a.md", "b.md", "c.md"}},
		{"concept_update_missing", []string{"terminology", "concept", "update", "--merge"}},
		{"concept_remove_missing", []string{"terminology", "concept", "remove"}},
		{"term_add_missing", []string{"terminology", "term", "add", "--lang", "es", "--term", "tzim"}},
		{"term_deprecate_missing", []string{"terminology", "term", "deprecate", "--lang", "es", "--term", "tzim"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := app.Root()
			cmd.Writer = &stdout
			cmd.ErrWriter = &stderr

			err := cmd.Run(context.Background(), tc.args)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			exitCode := output.ExitCodeFor(err)
			if exitCode != 2 {
				t.Errorf("exit code = %d, want 2", exitCode)
			}
		})
	}
}

func TestArgBounds_ErrorCodes(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantCode string
	}{
		{"missing_arg", []string{"terminology", "lookup"}, "missing_argument"},
		{"excess_arg", []string{"terminology", "lookup", "tzimtzum", "extra"}, "excess_arguments"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := app.Root()
			cmd.Writer = &stdout
			cmd.ErrWriter = &stderr

			err := cmd.Run(context.Background(), tc.args)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var buf bytes.Buffer
			output.EmitError(&buf, "json", err)

			var env map[string]any
			if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
				t.Fatalf("not valid JSON: %v", jsonErr)
			}
			errObj := env["error"].(map[string]any)
			if code := errObj["code"].(string); code != tc.wantCode {
				t.Errorf("error code = %q, want %q", code, tc.wantCode)
			}
		})
	}
}

func TestExtract_BasicMarkdown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	if ok, _ := env["ok"].(bool); !ok {
		t.Errorf("ok = %v, want true", env["ok"])
	}
	if sv, _ := env["schema_version"].(float64); int(sv) != output.SchemaVersion {
		t.Errorf("schema_version = %v, want %d", sv, output.SchemaVersion)
	}

	candidates, _ := env["candidates"].([]any)
	if len(candidates) == 0 {
		t.Fatal("expected at least one candidate, got 0")
	}

	terms := make(map[string]bool)
	for _, c := range candidates {
		cm := c.(map[string]any)
		terms[cm["term"].(string)] = true
	}
	if !terms["Razón Histórica"] {
		t.Error("expected capitalized phrase 'Razón Histórica' in candidates")
	}
}

func TestExtract_CodeBlocksExcluded(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	candidates, _ := env["candidates"].([]any)
	for _, c := range candidates {
		cm := c.(map[string]any)
		term := cm["term"].(string)
		if term == "TzimtzumProcess" {
			t.Error("code block identifier 'TzimtzumProcess' should not appear in candidates")
		}
	}
}

func TestExtract_ExcludeGlossaryTerms(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
		"--exclude", "testdata/fixtures/minimal-dct.tbx",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	candidates, _ := env["candidates"].([]any)
	for _, c := range candidates {
		cm := c.(map[string]any)
		term := cm["term"].(string)
		if term == "tzimtzum" || term == "צמצום" {
			t.Errorf("excluded glossary term %q should not appear in candidates", term)
		}
	}
}

func TestExtract_StopwordsFiltering(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
		"--stopwords", "testdata/fixtures/extract-stopwords.txt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	candidates, _ := env["candidates"].([]any)
	for _, c := range candidates {
		cm := c.(map[string]any)
		term := cm["term"].(string)
		if term == "el" || term == "la" || term == "de" {
			t.Errorf("stopword %q should not appear in candidates", term)
		}
	}
}

func TestExtract_MinFreqThreshold(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
		"--min-freq", "100",
	})

	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	candidates, _ := env["candidates"].([]any)
	for _, c := range candidates {
		cm := c.(map[string]any)
		heuristic := cm["heuristic"].(string)
		if heuristic == "high_frequency" {
			t.Errorf("no high_frequency candidates should pass min-freq 100")
		}
	}

	_ = exitCode
}

func TestExtract_ScriptFilter(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
		"--script", "hebrew",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	candidates, _ := env["candidates"].([]any)
	for _, c := range candidates {
		cm := c.(map[string]any)
		heuristic := cm["heuristic"].(string)
		if heuristic == "capitalized_phrase" {
			t.Errorf("script=hebrew should exclude capitalized_phrase candidates (Latin script)")
		}
	}
}

func TestExtract_FieldsProjection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
		"--fields", "candidates.term,candidates.frequency",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	candidates, _ := env["candidates"].([]any)
	if len(candidates) == 0 {
		t.Fatal("expected candidates in projected output")
	}

	c := candidates[0].(map[string]any)
	if _, ok := c["term"]; !ok {
		t.Error("projected output missing 'term'")
	}
	if _, ok := c["heuristic"]; ok {
		t.Error("projected output should not contain 'heuristic'")
	}
}

func TestExtract_NoCandidates_ExitCode1(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "testdata/fixtures/extract-empty.md",
		"--min-freq", "100",
	})
	if err == nil {
		t.Fatal("expected error for no candidates, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	if ok, _ := env["ok"].(bool); !ok {
		t.Errorf("ok = %v, want true", env["ok"])
	}

	candidates, _ := env["candidates"].([]any)
	if len(candidates) != 0 {
		t.Errorf("candidates length = %d, want 0", len(candidates))
	}
}

func TestExtract_NonexistentFile_ExitCode3(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "testdata/fixtures/nonexistent.md",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 3 {
		t.Errorf("exit code = %d, want 3", exitCode)
	}
}

func TestExtract_FrontmatterLangAffectsForeignScript(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract", "testdata/fixtures/extract-hebrew-doc.md",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	candidates, _ := env["candidates"].([]any)

	foreignTerms := make(map[string]bool)
	allTerms := make(map[string]bool)
	for _, c := range candidates {
		cm := c.(map[string]any)
		term := cm["term"].(string)
		allTerms[term] = true
		if cm["heuristic"].(string) == "foreign_script" {
			foreignTerms[term] = true
		}
	}

	// Latin tokens should appear as candidates (foreign_script or merged with another heuristic)
	for _, latin := range []string{"concept", "The", "Kabbalah"} {
		if !allTerms[latin] {
			t.Errorf("expected Latin token %q to appear as a candidate in Hebrew doc", latin)
		}
	}

	// At least one Latin-only token (not caught by other heuristics) should be foreign_script
	if !foreignTerms["concept"] {
		t.Error("expected lowercase Latin token 'concept' to be flagged as foreign_script in Hebrew doc")
	}

	// Hebrew tokens must NOT be flagged as foreign_script
	for _, hebrew := range []string{"הצמצום", "הוא", "מרכזי"} {
		if foreignTerms[hebrew] {
			t.Errorf("Hebrew token %q should not be flagged as foreign_script in Hebrew doc", hebrew)
		}
	}

	_ = stderr.String()
}

func TestTermDeprecate_NoTBXPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "term", "deprecate", "tzimtzum",
		"--lang", "en", "--term", "tzimtzum",
	})
	if err == nil {
		t.Fatal("expected no_tbx_path error, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}
}

func TestUrfaveErrors_ExitCode2(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"unknown_global_flag", []string{"terminology", "--bogus", "validate"}},
		{"invalid_format_enum", []string{"terminology", "--format", "yaml", "validate"}},
		{"invalid_status_enum", []string{"terminology", "concept", "add", "--status", "klingon"}},
		{"missing_required_file", []string{"terminology", "apply"}},
		{"missing_required_lang", []string{"terminology", "term", "add", "tzimtzum", "--term", "tzim"}},
		{"missing_required_term", []string{"terminology", "term", "add", "tzimtzum", "--lang", "es"}},
		{"invalid_script_enum", []string{"terminology", "extract", "a.md", "--script", "klingon"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := app.Root()
			cmd.Writer = &stdout
			cmd.ErrWriter = &stderr

			err := cmd.Run(context.Background(), tc.args)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			exitCode := output.ExitCodeFor(err)
			if exitCode != 2 {
				t.Errorf("exit code = %d, want 2", exitCode)
			}
		})
	}
}

func TestUrfaveErrors_ErrorCodes(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantCode string
	}{
		{"unknown_flag", []string{"terminology", "--bogus", "validate"}, "unknown_flag"},
		{"invalid_enum", []string{"terminology", "--format", "yaml", "validate"}, "invalid_value"},
		{"missing_required", []string{"terminology", "apply"}, "missing_required_flag"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := app.Root()
			cmd.Writer = &stdout
			cmd.ErrWriter = &stderr

			err := cmd.Run(context.Background(), tc.args)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var buf bytes.Buffer
			output.EmitError(&buf, "json", err)

			var env map[string]any
			if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
				t.Fatalf("not valid JSON: %v", jsonErr)
			}
			errObj := env["error"].(map[string]any)
			if code := errObj["code"].(string); code != tc.wantCode {
				t.Errorf("error code = %q, want %q", code, tc.wantCode)
			}
		})
	}
}

func TestUnknownSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown subcommand, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}

	var buf bytes.Buffer
	output.EmitError(&buf, "json", err)

	var env map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
		t.Fatalf("not valid JSON: %v", jsonErr)
	}
	errObj := env["error"].(map[string]any)

	if code := errObj["code"].(string); code != "unknown_subcommand" {
		t.Errorf("error code = %q, want %q", code, "unknown_subcommand")
	}
	msg := errObj["message"].(string)
	if !strings.Contains(msg, "bogus") {
		t.Errorf("error message %q should mention the unknown subcommand 'bogus'", msg)
	}
}

func TestNoSubcommand_StillWorks(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology"})
	if err == nil {
		t.Fatal("expected error for no subcommand, got nil")
	}

	var buf bytes.Buffer
	output.EmitError(&buf, "json", err)

	var env map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
		t.Fatalf("not valid JSON: %v", jsonErr)
	}
	errObj := env["error"].(map[string]any)

	if code := errObj["code"].(string); code != "no_subcommand" {
		t.Errorf("error code = %q, want %q", code, "no_subcommand")
	}
}

func TestValidate_Lenient_SuppressesStrictOnlyWarnings(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/with-legacy-and-unknown.tbx", "validate",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	warnings, _ := env["warnings"].([]any)
	for _, w := range warnings {
		wm := w.(map[string]any)
		code := wm["code"].(string)
		if code == "legacy_form_normalized" || code == "unknown_element" {
			t.Errorf("lenient mode should suppress %q warnings, but found one", code)
		}
	}
}

func TestValidate_Strict_IncludesStrictOnlyWarnings(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/with-legacy-and-unknown.tbx", "validate", "--strict",
	})

	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
	}
	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1 (warnings present)", exitCode)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	warnings, _ := env["warnings"].([]any)
	var foundLegacy, foundUnknown bool
	for _, w := range warnings {
		wm := w.(map[string]any)
		code := wm["code"].(string)
		if code == "legacy_form_normalized" {
			foundLegacy = true
		}
		if code == "unknown_element" {
			foundUnknown = true
		}
	}
	if !foundLegacy {
		t.Error("strict mode should include legacy_form_normalized warnings")
	}
	if !foundUnknown {
		t.Error("strict mode should include unknown_element warnings")
	}
}

func TestValidate_Strict_PromotesUnresolvedCrossrefToError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/with-warnings.tbx", "validate", "--strict",
	})
	if err == nil {
		t.Fatal("expected error for strict mode with unresolved crossref")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 65 {
		t.Errorf("exit code = %d, want 65 (unresolved_crossref promoted to error)", exitCode)
	}
}

// Golden CLI tests — Lookup

func TestLookup_Found_Golden(t *testing.T) {
	runGolden(t, "lookup/found", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "tzimtzum",
	})
}

func TestLookup_NotFound_Golden(t *testing.T) {
	runGolden(t, "lookup/not_found", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "nonexistent",
	})
}

func TestLookup_NoTBX_Golden(t *testing.T) {
	runGolden(t, "lookup/no_tbx", []string{"terminology", "lookup", "tzimtzum"})
}

func TestLookup_LangFilter_Golden(t *testing.T) {
	runGolden(t, "lookup/lang_filter", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "צמצום", "--lang", "he",
	})
}

func TestLookup_CaseInsensitive_Golden(t *testing.T) {
	runGolden(t, "lookup/case_insensitive", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "TZIMTZUM",
	})
}

func TestLookup_Fields_Golden(t *testing.T) {
	runGolden(t, "lookup/fields", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "tzimtzum",
		"--fields", "results.concept_id",
	})
}

func TestLookup_InvalidField_Golden(t *testing.T) {
	runGolden(t, "lookup/invalid_field", []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "tzimtzum",
		"--fields", "concpet_id",
	})
}

// Golden CLI tests — Search

func TestSearch_RomajiNoHyphen_Golden(t *testing.T) {
	runGolden(t, "search/romaji", []string{
		"terminology", "--tbx", "testdata/fixtures/aikido-dct.tbx", "search", "katatedori",
	})
}

func TestSearch_HiraganaReading_Golden(t *testing.T) {
	runGolden(t, "search/hiragana", []string{
		"terminology", "--tbx", "testdata/fixtures/aikido-dct.tbx", "search", "かたてどり",
	})
}

func TestSearch_MacronFold_Golden(t *testing.T) {
	runGolden(t, "search/macron", []string{
		"terminology", "--tbx", "testdata/fixtures/aikido-dct.tbx", "search", "kokyu",
	})
}

func TestSearch_SubstringEn_Golden(t *testing.T) {
	runGolden(t, "search/substring", []string{
		"terminology", "--tbx", "testdata/fixtures/aikido-dct.tbx", "search", "grab",
	})
}

func TestSearch_NotFound_Golden(t *testing.T) {
	runGolden(t, "search/not_found", []string{
		"terminology", "--tbx", "testdata/fixtures/aikido-dct.tbx", "search", "zzzz",
	})
}

func TestSearch_Fields_Golden(t *testing.T) {
	runGolden(t, "search/fields", []string{
		"terminology", "--tbx", "testdata/fixtures/aikido-dct.tbx", "search", "katatedori",
		"--fields", "results.concept_id",
	})
}

func TestSearch_IncludeDefinitions_Golden(t *testing.T) {
	runGolden(t, "search/include_definitions", []string{
		"terminology", "--tbx", "testdata/fixtures/aikido-dct.tbx", "search", "wrist",
		"--include", "definitions",
	})
}

func TestSearch_LangFilter_Golden(t *testing.T) {
	runGolden(t, "search/lang_filter", []string{
		"terminology", "--tbx", "testdata/fixtures/aikido-dct.tbx", "search", "grab", "--lang", "ja",
	})
}

func TestSearch_InvalidInclude_Golden(t *testing.T) {
	runGolden(t, "search/invalid_include", []string{
		"terminology", "--tbx", "testdata/fixtures/aikido-dct.tbx", "search", "katatedori",
		"--include", "bogus",
	})
}

func TestSearch_NoTBX_Golden(t *testing.T) {
	runGolden(t, "search/no_tbx", []string{"terminology", "search", "katatedori"})
}

// Golden CLI tests — Schema

func TestSchema_Full_Golden(t *testing.T) {
	runGolden(t, "schema/full", []string{"terminology", "schema"})
}

func TestSchema_CommandFilter_Golden(t *testing.T) {
	runGolden(t, "schema/command_filter", []string{
		"terminology", "schema", "--command", "validate",
	})
}

func TestSchema_UnknownCommand_Golden(t *testing.T) {
	runGolden(t, "schema/unknown_command", []string{
		"terminology", "schema", "--command", "nonexistent",
	})
}

// Golden CLI tests — Extract

func TestExtract_Basic_Golden(t *testing.T) {
	runGolden(t, "extract/basic", []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
	})
}

func TestExtract_NoCandidates_Golden(t *testing.T) {
	runGolden(t, "extract/no_candidates", []string{
		"terminology", "extract", "testdata/fixtures/extract-empty.md", "--min-freq", "100",
	})
}

func TestExtract_Exclude_Golden(t *testing.T) {
	runGolden(t, "extract/exclude", []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
		"--exclude", "testdata/fixtures/minimal-dct.tbx",
	})
}

func TestExtract_ScriptFilter_Golden(t *testing.T) {
	runGolden(t, "extract/script_filter", []string{
		"terminology", "extract", "testdata/fixtures/extract-mixed.md",
		"--script", "hebrew",
	})
}

func TestExtract_MinFreq_Golden(t *testing.T) {
	runGolden(t, "extract/min_freq", []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
		"--min-freq", "5",
	})
}

func TestExtract_Fields_Golden(t *testing.T) {
	runGolden(t, "extract/fields", []string{
		"terminology", "extract", "testdata/fixtures/extract-sample.md",
		"--fields", "candidates.term,candidates.frequency",
	})
}

// Golden CLI tests — Scan

func TestScan_Found_Golden(t *testing.T) {
	runGolden(t, "scan/found", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"testdata/fixtures/scan-corpus.md",
	})
}

func TestScan_NoMatches_Golden(t *testing.T) {
	runGolden(t, "scan/no_matches", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"testdata/fixtures/scan-empty.md",
	})
}

func TestScan_LangFilter_Golden(t *testing.T) {
	runGolden(t, "scan/lang_filter", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"--lang", "he",
		"testdata/fixtures/scan-corpus.md",
	})
}

func TestScan_ContextWindow_Golden(t *testing.T) {
	runGolden(t, "scan/context_window", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"--context", "40",
		"testdata/fixtures/scan-corpus.md",
	})
}

func TestScan_Fields_Golden(t *testing.T) {
	runGolden(t, "scan/fields", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"--fields", "matches.concept_id,matches.line",
		"testdata/fixtures/scan-corpus.md",
	})
}

func TestScan_CaseInsensitive_Golden(t *testing.T) {
	runGolden(t, "scan/case_insensitive", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"testdata/fixtures/scan-uppercase.md",
	})
}

func TestScan_NoTBX_Golden(t *testing.T) {
	runGolden(t, "scan/no_tbx", []string{
		"terminology", "scan", "testdata/fixtures/scan-corpus.md",
	})
}

func TestScan_FileNotFound_Golden(t *testing.T) {
	runGolden(t, "scan/file_not_found", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"nonexistent.md",
	})
}

func TestScan_InvalidField_Golden(t *testing.T) {
	runGolden(t, "scan/invalid_field", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"--fields", "concpet_id",
		"testdata/fixtures/scan-corpus.md",
	})
}

func TestScan_CodeBlocksSkipped_Golden(t *testing.T) {
	runGolden(t, "scan/code_blocks_skipped", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"testdata/fixtures/scan-code-only.md",
	})
}

func TestScan_MultiWord_Golden(t *testing.T) {
	runGolden(t, "scan/multi_word", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"testdata/fixtures/scan-multiword.md",
	})
}

func TestScan_StatusTagging_Golden(t *testing.T) {
	runGolden(t, "scan/status_tagging", []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		"testdata/fixtures/scan-deprecated.md",
	})
}

func TestLookup_CaseFoldMatch(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", "testdata/fixtures/rich-dct.tbx", "lookup", "Tzimtzum",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &env); jsonErr != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", jsonErr, stdout.String())
	}

	results, _ := env["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results length = %d, want 1 (case-fold should match)", len(results))
	}
}

// --- sanitizer rejection golden tests ---

func TestScan_PathTraversal_Golden(t *testing.T) {
	runGolden(t, "sanitize/scan_path_traversal", []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-dct.tbx",
		"scan", "../../../etc/passwd",
	})
}

func TestCheck_PathTraversal_Golden(t *testing.T) {
	runGolden(t, "sanitize/check_path_traversal", []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-dct.tbx",
		"check", "../../../etc/passwd", "testdata/fixtures/check-target-clean.md",
	})
}

func TestCheck_InvalidLang_Golden(t *testing.T) {
	runGolden(t, "sanitize/check_invalid_lang", []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-dct.tbx",
		"check", "testdata/fixtures/check-source.md", "testdata/fixtures/check-target-clean.md",
		"--source-lang", "not valid!",
	})
}

func TestExtract_PathTraversal_Golden(t *testing.T) {
	runGolden(t, "sanitize/extract_path_traversal", []string{
		"terminology", "extract", "../../../etc/passwd",
	})
}

func TestApply_PathTraversal_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "sanitize/apply_path_traversal", []string{
		"terminology", "--tbx", tbx, "apply", "--file", "../../../etc/payload.json",
	})
}

func TestLookup_InvalidLang_Golden(t *testing.T) {
	runGolden(t, "sanitize/lookup_invalid_lang", []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-dct.tbx",
		"lookup", "tzimtzum", "--lang", "not valid!",
	})
}

func TestScan_AbsolutePathOutsideCWD_Golden(t *testing.T) {
	runGolden(t, "sanitize/scan_absolute_outside", []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-dct.tbx",
		"scan", "/etc/passwd",
	})
}

func TestCheck_AbsolutePathOutsideCWD_Golden(t *testing.T) {
	runGolden(t, "sanitize/check_absolute_outside", []string{
		"terminology", "--tbx", "testdata/fixtures/minimal-dct.tbx",
		"check", "/etc/passwd", "testdata/fixtures/check-target-clean.md",
	})
}

func TestExtract_AbsolutePathOutsideCWD_Golden(t *testing.T) {
	runGolden(t, "sanitize/extract_absolute_outside", []string{
		"terminology", "extract", "/etc/passwd",
	})
}

// --- bounded reads tests ---

func TestValidate_InputTooLarge(t *testing.T) {
	dir := t.TempDir()
	hugeTBX := filepath.Join(dir, "huge.tbx")
	if err := os.WriteFile(hugeTBX, make([]byte, tbx.MaxTBXSize+1), 0o644); err != nil {
		t.Fatalf("creating oversized TBX: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{"terminology", "validate", "--tbx", hugeTBX})
	if err == nil {
		t.Fatal("expected error for oversized TBX, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}

	output.EmitError(&stderr, "json", err)
	assertErrorCode(t, stderr.Bytes(), "input_too_large")
}

func TestInputTooLarge_Golden(t *testing.T) {
	dir := t.TempDir()
	hugeTBX := filepath.Join(dir, "huge.tbx")
	if err := os.WriteFile(hugeTBX, make([]byte, tbx.MaxTBXSize+1), 0o644); err != nil {
		t.Fatalf("creating oversized TBX: %v", err)
	}

	runGolden(t, "validate/input_too_large", []string{
		"terminology", "validate", "--tbx", hugeTBX,
	})
}

func writeLargeFileInCWD(t *testing.T, name string, size int64) string {
	t.Helper()
	dir, err := os.MkdirTemp("testdata", "bounded-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatalf("creating oversized file: %v", err)
	}
	return path
}

func TestScan_InputTooLarge(t *testing.T) {
	hugeMD := writeLargeFileInCWD(t, "huge.md", tbx.MaxMarkdownSize+1)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "scan",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		hugeMD,
	})
	if err == nil {
		t.Fatal("expected error for oversized markdown, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}

	output.EmitError(&stderr, "json", err)
	assertErrorCode(t, stderr.Bytes(), "input_too_large")
}

func TestCheck_InputTooLarge(t *testing.T) {
	hugeMD := writeLargeFileInCWD(t, "huge.md", tbx.MaxMarkdownSize+1)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "check",
		"--tbx", "testdata/fixtures/scan-glossary.tbx",
		hugeMD, "testdata/fixtures/check-target-clean.md",
	})
	if err == nil {
		t.Fatal("expected error for oversized source, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}

	output.EmitError(&stderr, "json", err)
	assertErrorCode(t, stderr.Bytes(), "input_too_large")
}

func TestExtract_InputTooLarge(t *testing.T) {
	hugeMD := writeLargeFileInCWD(t, "huge.md", tbx.MaxMarkdownSize+1)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "extract",
		hugeMD,
	})
	if err == nil {
		t.Fatal("expected error for oversized extract input, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}

	output.EmitError(&stderr, "json", err)
	assertErrorCode(t, stderr.Bytes(), "input_too_large")
}

func TestApply_InputTooLarge(t *testing.T) {
	tbxFile := copyFixture(t, "minimal-dct.tbx")
	hugePayload := writeLargeFileInCWD(t, "huge.json", tbx.MaxPayloadSize+1)

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), []string{
		"terminology", "apply",
		"--tbx", tbxFile,
		"--file", hugePayload,
	})
	if err == nil {
		t.Fatal("expected error for oversized payload, got nil")
	}

	exitCode := output.ExitCodeFor(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}

	output.EmitError(&stderr, "json", err)
	assertErrorCode(t, stderr.Bytes(), "input_too_large")
}

func assertErrorCode(t *testing.T, stderrJSON []byte, wantCode string) {
	t.Helper()
	var env map[string]any
	if err := json.Unmarshal(stderrJSON, &env); err != nil {
		t.Fatalf("unmarshaling stderr: %v (raw: %s)", err, stderrJSON)
	}
	errObj, ok := env["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object in stderr; got %v", env)
	}
	code, _ := errObj["code"].(string)
	if code != wantCode {
		t.Errorf("error code = %q, want %q", code, wantCode)
	}
}
