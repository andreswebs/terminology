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
)

func runInit(t *testing.T, argv []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(context.Background(), argv)
	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
		output.EmitError(&stderr, cmd.String("format"), err)
	}
	return stdout.String(), stderr.String(), exitCode
}

func TestInit_Golden(t *testing.T) {
	target := filepath.Join(t.TempDir(), "out.tbx")
	runGolden(t, "init/happy", []string{
		"terminology", "--tbx", target, "init",
		"--source-lang", "es",
		"--title", "Glossary",
	})
}

func TestInit_HappyPath_CreatesValidTBX(t *testing.T) {
	target := filepath.Join(t.TempDir(), "new.tbx")

	stdout, _, exit := runInit(t, []string{
		"terminology", "--tbx", target, "init",
		"--source-lang", "es",
		"--title", "My Glossary",
	})
	if exit != 0 {
		t.Fatalf("exit = %d, want 0; stdout=%s", exit, stdout)
	}

	var env output.InitEnvelope
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("unmarshal stdout: %v\n%s", err, stdout)
	}
	if !env.OK {
		t.Errorf("ok = false")
	}
	if env.SourceLang != "es" {
		t.Errorf("source_lang = %q, want es", env.SourceLang)
	}
	if env.Title != "My Glossary" {
		t.Errorf("title = %q", env.Title)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, `xml:lang="es"`) {
		t.Errorf("missing xml:lang=es in file:\n%s", body)
	}
	if !strings.Contains(body, `<title>My Glossary</title>`) {
		t.Errorf("missing title in file:\n%s", body)
	}
	if !strings.Contains(body, `type="TBX-Linguist"`) {
		t.Errorf("missing TBX-Linguist type")
	}
}

func TestInit_ValidateRoundTrip(t *testing.T) {
	target := filepath.Join(t.TempDir(), "new.tbx")

	_, _, exit := runInit(t, []string{
		"terminology", "--tbx", target, "init",
		"--source-lang", "es",
	})
	if exit != 0 {
		t.Fatalf("init exit = %d, want 0", exit)
	}

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr
	err := cmd.Run(context.Background(), []string{
		"terminology", "--tbx", target, "validate",
	})
	if err != nil {
		t.Fatalf("validate failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
}

func TestInit_SourceLangMissing_UsageError(t *testing.T) {
	target := filepath.Join(t.TempDir(), "new.tbx")

	_, _, exit := runInit(t, []string{
		"terminology", "--tbx", target, "init",
	})
	if exit != 2 {
		t.Errorf("exit = %d, want 2", exit)
	}

	if _, err := os.Stat(target); err == nil {
		t.Errorf("file should not have been created")
	}
}

func TestInit_ExistingTarget_IOError(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing.tbx")
	if err := os.WriteFile(target, []byte("<tbx/>"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	_, _, exit := runInit(t, []string{
		"terminology", "--tbx", target, "init",
		"--source-lang", "es",
	})
	if exit != 3 {
		t.Errorf("exit = %d, want 3", exit)
	}

	data, _ := os.ReadFile(target)
	if string(data) != "<tbx/>" {
		t.Errorf("existing file was modified: %s", data)
	}
}

func TestInit_DryRun_NoWrite(t *testing.T) {
	target := filepath.Join(t.TempDir(), "preview.tbx")

	stdout, _, exit := runInit(t, []string{
		"terminology", "--tbx", target, "init",
		"--source-lang", "es",
		"--dry-run",
	})
	if exit != 0 {
		t.Fatalf("exit = %d, want 0", exit)
	}

	var env output.InitEnvelope
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !env.DryRun {
		t.Errorf("dry_run should be true in envelope")
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("dry-run should not create file, stat err = %v", err)
	}
}

func TestInit_InvalidSourceLang(t *testing.T) {
	target := filepath.Join(t.TempDir(), "new.tbx")

	_, _, exit := runInit(t, []string{
		"terminology", "--tbx", target, "init",
		"--source-lang", "not valid!",
	})
	if exit != 65 {
		t.Errorf("exit = %d, want 65", exit)
	}
}

func TestInit_SchemaCommandDescribed(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	if err := cmd.Run(context.Background(), []string{
		"terminology", "schema", "--command", "init",
	}); err != nil {
		t.Fatalf("schema --command init failed: %v\nstderr=%s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, `"name":"init"`) {
		t.Errorf("schema missing init: %s", out)
	}
	if !strings.Contains(out, "source-lang") {
		t.Errorf("schema missing source-lang flag: %s", out)
	}
}
