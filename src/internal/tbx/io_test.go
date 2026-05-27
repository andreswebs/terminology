package tbx_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
	_ "github.com/andreswebs/terminology/internal/tbx/linguist"
	"github.com/andreswebs/terminology/internal/terr"
)

func TestLoad_MinimalDCT(t *testing.T) {
	g, _, err := tbx.Load("testdata/minimal-dct.tbx")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if g.Dialect != tbx.DialectLinguist {
		t.Errorf("dialect = %q, want %q", g.Dialect, tbx.DialectLinguist)
	}
	if len(g.Concepts) != 1 {
		t.Errorf("concepts = %d, want 1", len(g.Concepts))
	}
}

func TestLoad_UnsupportedDialect(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.tbx")
	if err := os.WriteFile(path, []byte(`<?xml version="1.0"?><tbx type="TBX-Basic" style="dct"></tbx>`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := tbx.Load(path)
	if err == nil {
		t.Fatal("expected error for unsupported dialect")
	}
	if !errors.Is(err, tbx.ErrUnsupportedDialect) {
		t.Errorf("error = %v, want ErrUnsupportedDialect", err)
	}
}

func TestLoad_MissingTypeAttribute(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "notype.tbx")
	if err := os.WriteFile(path, []byte(`<?xml version="1.0"?><tbx style="dct"></tbx>`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := tbx.Load(path)
	if err == nil {
		t.Fatal("expected error for missing type attribute")
	}
	if !errors.Is(err, tbx.ErrUnsupportedDialect) {
		t.Errorf("error = %v, want ErrUnsupportedDialect", err)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, _, err := tbx.Load("nonexistent.tbx")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Errorf("error should wrap os.PathError, got %T: %v", err, err)
	}
}

func TestLoad_NoTBXElement(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "notbx.tbx")
	if err := os.WriteFile(path, []byte(`<?xml version="1.0"?><root></root>`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := tbx.Load(path)
	if err == nil {
		t.Fatal("expected error for missing <tbx> element")
	}
	if !errors.Is(err, tbx.ErrUnsupportedDialect) {
		t.Errorf("error = %v, want ErrUnsupportedDialect", err)
	}
}

func TestLoad_MalformedXML(t *testing.T) {
	_, _, err := tbx.Load("linguist/testdata/malformed/bad-xml.tbx")
	if err == nil {
		t.Fatal("expected error for malformed XML")
	}
}

func TestLoad_EmptyBody(t *testing.T) {
	g, warnings, err := tbx.Load("linguist/testdata/malformed/empty-body.tbx")
	if err != nil {
		t.Fatalf("expected no error for empty body, got: %v", err)
	}
	if len(g.Concepts) != 0 {
		t.Errorf("concepts = %d, want 0", len(g.Concepts))
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %d, want 0", len(warnings))
	}
}

func TestSave_OriginalPreservedOnWriteError(t *testing.T) {
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "output.tbx")

	original := []byte(`<?xml version="1.0"?>` + "\n" + `<tbx type="TBX-Linguist" style="dct"><text><body></body></text></tbx>`)
	if err := os.WriteFile(dst, original, 0o644); err != nil {
		t.Fatal(err)
	}

	g := &tbx.Glossary{Dialect: "BOGUS-DIALECT"}
	err := tbx.Save(dst, g)
	if err == nil {
		t.Fatal("expected error for unsupported dialect")
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading original: %v", err)
	}
	if string(got) != string(original) {
		t.Errorf("original file was modified on failed Save:\n--- original ---\n%s\n--- got ---\n%s",
			string(original), string(got))
	}

	matches, _ := filepath.Glob(filepath.Join(tmp, ".terminology-*.tmp"))
	if len(matches) > 0 {
		t.Errorf("temp files should be cleaned up after error, found: %v", matches)
	}
}

func TestSave_ErrTBXLocked_NonExistentDir(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "no-such-dir", "output.tbx")

	g := &tbx.Glossary{
		Dialect:    tbx.DialectLinguist,
		SourceDesc: "test",
	}

	err := tbx.Save(dst, g)
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
	coded, ok := err.(terr.Coded)
	if !ok {
		t.Fatalf("error %T does not implement terr.Coded", err)
	}
	if got := coded.Code(); got != "tbx_locked" {
		t.Errorf("Code() = %q, want %q", got, "tbx_locked")
	}
}

func TestSave_CleanupOnError(t *testing.T) {
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "output.tbx")

	g := &tbx.Glossary{
		Dialect:    tbx.DialectLinguist,
		SourceDesc: "test",
	}

	if err := tbx.Save(dst, g); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := os.Stat(dst); err != nil {
		t.Errorf("output file should exist: %v", err)
	}

	matches, _ := filepath.Glob(filepath.Join(tmp, ".terminology-*.tmp"))
	if len(matches) > 0 {
		t.Errorf("temp files should be cleaned up, found: %v", matches)
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	src := "testdata/minimal-dct.tbx"
	g, _, err := tbx.Load(src)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	tmp := t.TempDir()
	dst := filepath.Join(tmp, "output.tbx")

	if err := tbx.Save(dst, g); err != nil {
		t.Fatalf("Save: %v", err)
	}

	original, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if string(original) != string(written) {
		t.Errorf("Save output differs from original:\n--- original ---\n%s\n--- written ---\n%s",
			string(original), string(written))
	}

	lockPath := dst + ".lock"
	if _, err := os.Stat(lockPath); err == nil {
		t.Errorf("lock file should not persist after successful write")
	}
}
