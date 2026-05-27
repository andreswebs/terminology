package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreswebs/terminology/internal/clock"
)

type fakeClock struct{ T time.Time }

func (f fakeClock) Now() time.Time { return f.T }

var fixedTime = time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC)

func writeCtx() context.Context {
	return clock.With(context.Background(), fakeClock{T: fixedTime})
}

func copyFixture(t *testing.T, name string) string {
	t.Helper()
	src := filepath.Join("testdata", "fixtures", name)
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read fixture %s: %v", src, err)
	}
	dst := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write fixture copy: %v", err)
	}
	return dst
}

func pipeStdin(t *testing.T, data string) func() {
	t.Helper()
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	_, _ = w.WriteString(data)
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}
	os.Stdin = r
	return func() { os.Stdin = origStdin }
}

// --- concept add ---

func TestConceptAdd_Flags_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "concept_add/flags", []string{
		"terminology", "--tbx", tbx, "concept", "add",
		"--lang", "es", "--term", "sefirot",
		"--status", "preferredTerm-admn-sts",
		"--subject-field", "kabbalah",
	})
}

func TestConceptAdd_DryRun_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "concept_add/dry_run", []string{
		"terminology", "--tbx", tbx, "concept", "add",
		"--lang", "en", "--term", "sefirot",
		"--dry-run",
	})
}

func TestConceptAdd_DuplicateID_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "concept_add/duplicate_id", []string{
		"terminology", "--tbx", tbx, "concept", "add",
		"--id", "tzimtzum",
		"--lang", "en", "--term", "tzimtzum",
	})
}

func TestConceptAdd_Transaction_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGoldenCtx(t, "concept_add/transaction", []string{
		"terminology", "--tbx", tbx, "concept", "add",
		"--lang", "en", "--term", "sefirot",
		"--transaction", "--author", "test-author",
	}, writeCtx())
}

func TestConceptAdd_JSONStdin_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	restore := pipeStdin(t, `{
  "concept_id": "binah",
  "subject_field": "kabbalah",
  "languages": {
    "en": {
      "preferred": { "term": "binah", "part_of_speech": "noun" }
    },
    "he": {
      "preferred": { "term": "בינה", "part_of_speech": "noun" }
    }
  }
}`)
	defer restore()
	runGolden(t, "concept_add/json_stdin", []string{
		"terminology", "--tbx", tbx, "concept", "add",
	})
}

func TestConceptAdd_TBXFragment_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	restore := pipeStdin(t, `<conceptEntry id="binah">
  <min:subjectField>kabbalah</min:subjectField>
  <langSec xml:lang="en">
    <termSec>
      <term>binah</term>
      <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
    </termSec>
  </langSec>
</conceptEntry>`)
	defer restore()
	runGolden(t, "concept_add/tbx_fragment", []string{
		"terminology", "--tbx", tbx, "concept", "add",
	})
}

func TestConceptAdd_IDDerivation_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "concept_add/id_derivation", []string{
		"terminology", "--tbx", tbx, "concept", "add",
		"--lang", "en", "--term", "Razón Histórica",
	})
}

// --- concept update ---

func TestConceptUpdate_Merge_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "concept_update/merge_happy", []string{
		"terminology", "--tbx", tbx, "concept", "update", "tzimtzum",
		"--merge",
		"--lang", "es", "--term", "tzimtzum",
		"--status", "preferredTerm-admn-sts",
	})
}

func TestConceptUpdate_Replace_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "concept_update/replace_happy", []string{
		"terminology", "--tbx", tbx, "concept", "update", "tzimtzum",
		"--replace",
		"--lang", "en", "--term", "contraction",
		"--subject-field", "mysticism",
	})
}

func TestConceptUpdate_NotFound_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "concept_update/not_found", []string{
		"terminology", "--tbx", tbx, "concept", "update", "nonexistent",
		"--merge",
		"--lang", "en", "--term", "test",
	})
}

func TestConceptUpdate_IDStability_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "concept_update/id_stability", []string{
		"terminology", "--tbx", tbx, "concept", "update", "tzimtzum",
		"--replace",
		"--lang", "en", "--term", "completely-different-term",
	})
}

// --- concept remove ---

func TestConceptRemove_Clean_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "concept_remove/clean", []string{
		"terminology", "--tbx", tbx, "concept", "remove", "tzimtzum",
	})
}

func TestConceptRemove_NotFound_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "concept_remove/not_found", []string{
		"terminology", "--tbx", tbx, "concept", "remove", "nonexistent",
	})
}

func TestConceptRemove_DanglingCrossref_Golden(t *testing.T) {
	tbx := copyFixture(t, "crossref-dct.tbx")
	runGolden(t, "concept_remove/dangling_crossref", []string{
		"terminology", "--tbx", tbx, "concept", "remove", "tzimtzum",
	})
}

func TestConceptRemove_Force_Golden(t *testing.T) {
	tbx := copyFixture(t, "crossref-dct.tbx")
	runGolden(t, "concept_remove/force", []string{
		"terminology", "--tbx", tbx, "concept", "remove", "tzimtzum", "--force",
	})
}

// --- term add ---

func TestTermAdd_ExistingLangSec_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "term_add/existing_langsec", []string{
		"terminology", "--tbx", tbx, "term", "add", "tzimtzum",
		"--lang", "en", "--term", "contraction",
		"--status", "admittedTerm-admn-sts",
	})
}

func TestTermAdd_NewLangSec_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "term_add/new_langsec", []string{
		"terminology", "--tbx", tbx, "term", "add", "tzimtzum",
		"--lang", "es", "--term", "tzimtzum",
		"--status", "preferredTerm-admn-sts",
		"--part-of-speech", "noun",
	})
}

func TestTermAdd_NotFound_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "term_add/not_found", []string{
		"terminology", "--tbx", tbx, "term", "add", "nonexistent",
		"--lang", "en", "--term", "test",
	})
}

// --- term deprecate ---

func TestTermDeprecate_Happy_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "term_deprecate/happy", []string{
		"terminology", "--tbx", tbx, "term", "deprecate", "tzimtzum",
		"--lang", "en", "--term", "tzimtzum",
	})
}

func TestTermDeprecate_NotFound_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "term_deprecate/not_found", []string{
		"terminology", "--tbx", tbx, "term", "deprecate", "nonexistent",
		"--lang", "en", "--term", "test",
	})
}

// --- sanitizer rejection tests ---

func TestConceptAdd_InvalidID_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "sanitize/concept_add_invalid_id", []string{
		"terminology", "--tbx", tbx, "concept", "add",
		"--id", "../traversal", "--lang", "en", "--term", "test",
	})
}

func TestConceptAdd_InvalidLang_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "sanitize/concept_add_invalid_lang", []string{
		"terminology", "--tbx", tbx, "concept", "add",
		"--lang", "not valid!", "--term", "test",
	})
}

func TestConceptAdd_InvalidTerm_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "sanitize/concept_add_invalid_term", []string{
		"terminology", "--tbx", tbx, "concept", "add",
		"--lang", "en", "--term", "bad\x00term",
	})
}

func TestConceptUpdate_InvalidID_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "sanitize/concept_update_invalid_id", []string{
		"terminology", "--tbx", tbx, "concept", "update", "id%2Fencoded",
		"--merge", "--lang", "en", "--term", "test",
	})
}

func TestConceptRemove_InvalidID_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "sanitize/concept_remove_invalid_id", []string{
		"terminology", "--tbx", tbx, "concept", "remove", "id?query",
	})
}

func TestTermAdd_InvalidID_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "sanitize/term_add_invalid_id", []string{
		"terminology", "--tbx", tbx, "term", "add", "../escape",
		"--lang", "en", "--term", "test",
	})
}

func TestTermAdd_InvalidLang_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "sanitize/term_add_invalid_lang", []string{
		"terminology", "--tbx", tbx, "term", "add", "tzimtzum",
		"--lang", "x", "--term", "test",
	})
}

func TestTermDeprecate_InvalidID_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "sanitize/term_deprecate_invalid_id", []string{
		"terminology", "--tbx", tbx, "term", "deprecate", "../traversal",
		"--lang", "en", "--term", "test",
	})
}

func TestTBXPath_Traversal_Golden(t *testing.T) {
	runGolden(t, "sanitize/tbx_path_traversal", []string{
		"terminology", "--tbx", "../../../etc/passwd", "validate",
	})
}
