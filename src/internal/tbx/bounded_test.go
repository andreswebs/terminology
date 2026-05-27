package tbx_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
)

func TestReadBounded_UnderLimit(t *testing.T) {
	data := []byte("hello world")
	r := bytes.NewReader(data)

	got, err := tbx.ReadBounded(r, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestReadBounded_ExactlyAtLimit(t *testing.T) {
	data := []byte("12345")
	r := bytes.NewReader(data)

	got, err := tbx.ReadBounded(r, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestReadBounded_OverLimit(t *testing.T) {
	data := []byte(strings.Repeat("x", 100))
	r := bytes.NewReader(data)

	_, err := tbx.ReadBounded(r, 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected Coded error, got %T", err)
	}
	if coded.Code() != "input_too_large" {
		t.Errorf("code = %q, want %q", coded.Code(), "input_too_large")
	}
}

func TestReadBounded_MessageIncludesCap(t *testing.T) {
	data := []byte(strings.Repeat("x", 100))
	r := bytes.NewReader(data)

	_, err := tbx.ReadBounded(r, 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	unwrapped := errors.Unwrap(err)
	if unwrapped == nil {
		t.Fatal("expected wrapped error with cause")
	}
	if !strings.Contains(unwrapped.Error(), "50") {
		t.Errorf("cause should include the cap value; got %q", unwrapped.Error())
	}
}

func TestErrInputTooLarge_Coded(t *testing.T) {
	var coded interface{ Code() string }
	if !errors.As(tbx.ErrInputTooLarge, &coded) {
		t.Fatal("ErrInputTooLarge should implement Coded")
	}
	if coded.Code() != "input_too_large" {
		t.Errorf("code = %q, want %q", coded.Code(), "input_too_large")
	}
}

func TestErrInputTooLarge_ExitCode(t *testing.T) {
	var exiter interface{ ExitCode() int }
	if !errors.As(tbx.ErrInputTooLarge, &exiter) {
		t.Fatal("ErrInputTooLarge should implement ExitCode")
	}
	if exiter.ExitCode() != 2 {
		t.Errorf("exit code = %d, want 2", exiter.ExitCode())
	}
}

func TestLoad_RejectsOversizedFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/huge.tbx"

	data := make([]byte, tbx.MaxTBXSize+1)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, _, err := tbx.Load(path)
	if err == nil {
		t.Fatal("expected error for oversized TBX, got nil")
	}
	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected Coded error, got %T: %v", err, err)
	}
	if coded.Code() != "input_too_large" {
		t.Errorf("code = %q, want %q", coded.Code(), "input_too_large")
	}
}

func TestReadFileBounded_RejectsOversizedFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/huge.txt"

	data := make([]byte, 100)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err := tbx.ReadFileBounded(path, 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected Coded error, got %T: %v", err, err)
	}
	if coded.Code() != "input_too_large" {
		t.Errorf("code = %q, want %q", coded.Code(), "input_too_large")
	}
}
