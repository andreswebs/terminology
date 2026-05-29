package app_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/andreswebs/terminology/internal/app"
	"github.com/andreswebs/terminology/internal/output"
)

func runMissingTBX(t *testing.T, argv []string) error {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr
	return cmd.Run(context.Background(), argv)
}

func assertIOError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error for missing TBX, got nil")
	}
	if code := output.ExitCodeFor(err); code != 3 {
		t.Errorf("exit code = %d, want 3", code)
	}
	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected coded error, got %T", err)
	}
	if coded.Code() != "io_error" {
		t.Errorf("code = %q, want %q", coded.Code(), "io_error")
	}
}

func TestApply_MissingTBX_IOError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nonexistent.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[]}`)
	err := runMissingTBX(t, []string{
		"terminology", "--tbx", missing, "apply", "--file", payload,
	})
	assertIOError(t, err)
}

func TestConceptAdd_MissingTBX_IOError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nonexistent.tbx")
	err := runMissingTBX(t, []string{
		"terminology", "--tbx", missing, "concept", "add",
		"--lang", "en", "--term", "tzimtzum",
	})
	assertIOError(t, err)
}

func TestConceptUpdate_MissingTBX_IOError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nonexistent.tbx")
	err := runMissingTBX(t, []string{
		"terminology", "--tbx", missing, "concept", "update", "tzimtzum",
		"--merge", "--lang", "en", "--term", "tzimtzum",
	})
	assertIOError(t, err)
}

func TestConceptRemove_MissingTBX_IOError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nonexistent.tbx")
	err := runMissingTBX(t, []string{
		"terminology", "--tbx", missing, "concept", "remove", "tzimtzum",
	})
	assertIOError(t, err)
}

func TestTermAdd_MissingTBX_IOError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nonexistent.tbx")
	err := runMissingTBX(t, []string{
		"terminology", "--tbx", missing, "term", "add", "tzimtzum",
		"--lang", "en", "--term", "tzimtzum",
	})
	assertIOError(t, err)
}

func TestTermDeprecate_MissingTBX_IOError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nonexistent.tbx")
	err := runMissingTBX(t, []string{
		"terminology", "--tbx", missing, "term", "deprecate", "tzimtzum",
		"--lang", "en", "--term", "tzimtzum",
	})
	assertIOError(t, err)
}
