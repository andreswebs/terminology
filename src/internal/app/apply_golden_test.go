package app_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/andreswebs/terminology/internal/app"
)

func writePayloadFile(t *testing.T, name, content string) string {
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

// --- apply: add ---

func TestApply_Add_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[{"concept_id":"binah","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"binah","administrative_status":"preferredTerm-admn-sts","part_of_speech":"noun"}}}}]}`)
	runGolden(t, "apply/add", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload,
	})
}

// --- apply: update ---

func TestApply_Update_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[{"concept_id":"tzimtzum","subject_field":"mysticism","languages":{"en":{"preferred":{"term":"tzimtzum","administrative_status":"preferredTerm-admn-sts","part_of_speech":"noun"}}}}]}`)
	runGolden(t, "apply/update", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload,
	})
}

// --- apply: unchanged ---

func TestApply_Unchanged_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[{"concept_id":"tzimtzum","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"tzimtzum","administrative_status":"preferredTerm-admn-sts","part_of_speech":"noun"}},"he":{"preferred":{"term":"צמצום","administrative_status":"preferredTerm-admn-sts","part_of_speech":"noun"}}}}]}`)
	runGolden(t, "apply/unchanged", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload,
	})
}

// --- apply: mixed (add + update + unchanged) ---

func TestApply_Mixed_Golden(t *testing.T) {
	tbx := copyFixture(t, "crossref-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[{"concept_id":"tzimtzum","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"tzimtzum","administrative_status":"preferredTerm-admn-sts"}}}},{"concept_id":"sefirot","subject_field":"mysticism","cross_refs":[{"target":"tzimtzum","label":"related concept"}],"languages":{"en":{"preferred":{"term":"sefirot","administrative_status":"preferredTerm-admn-sts"}}}},{"concept_id":"binah","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"binah"}}}}]}`)
	runGolden(t, "apply/mixed", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload,
	})
}

// --- apply: idempotent ---

func TestApply_Idempotent_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[{"concept_id":"tzimtzum","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"tzimtzum","administrative_status":"preferredTerm-admn-sts","part_of_speech":"noun"}},"he":{"preferred":{"term":"צמצום","administrative_status":"preferredTerm-admn-sts","part_of_speech":"noun"}}}}]}`)

	{
		var stdout, stderr bytes.Buffer
		cmd := app.Root()
		cmd.Writer = &stdout
		cmd.ErrWriter = &stderr
		if err := cmd.Run(context.Background(), []string{
			"terminology", "--tbx", tbx, "apply", "--file", payload,
		}); err != nil {
			t.Fatalf("first apply: %v", err)
		}
	}

	runGolden(t, "apply/idempotent", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload,
	})
}

// --- apply: prune ---

func TestApply_Prune_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[{"concept_id":"binah","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"binah"}}}}]}`)
	runGolden(t, "apply/prune", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload, "--prune",
	})
}

// --- apply: prune with dangling crossref ---

func TestApply_PruneCrossref_Golden(t *testing.T) {
	tbx := copyFixture(t, "crossref-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[{"concept_id":"sefirot","subject_field":"kabbalah","cross_refs":[{"target":"tzimtzum","label":"related concept"}],"languages":{"en":{"preferred":{"term":"sefirot","administrative_status":"preferredTerm-admn-sts"}}}}]}`)
	runGolden(t, "apply/prune_crossref", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload, "--prune",
	})
}

// --- apply: TBX fragment ---

func TestApply_TBXFragment_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	fragment := `<conceptEntry id="binah">
  <min:subjectField>kabbalah</min:subjectField>
  <langSec xml:lang="en">
    <termSec>
      <term>binah</term>
      <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
    </termSec>
  </langSec>
</conceptEntry>`
	payload := writePayloadFile(t, "payload.tbx", fragment)
	runGolden(t, "apply/tbx_fragment", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload,
	})
}

// --- apply: dry-run ---

func TestApply_DryRun_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[{"concept_id":"binah","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"binah"}}}}]}`)
	runGolden(t, "apply/dry_run", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload, "--dry-run",
	})
}

// --- apply: invalid JSON ---

func TestApply_InvalidJSON_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{this is not valid json}`)
	runGolden(t, "apply/invalid_json", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload,
	})
}

// --- apply: no TBX path ---

func TestApply_NoTBX_Golden(t *testing.T) {
	payload := writePayloadFile(t, "payload.json", `{"concepts":[]}`)
	runGolden(t, "apply/no_tbx", []string{
		"terminology", "apply", "--file", payload,
	})
}

// --- apply: transaction ---

func TestApply_Transaction_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[{"concept_id":"binah","subject_field":"kabbalah","languages":{"en":{"preferred":{"term":"binah"}}}}]}`)
	runGoldenCtx(t, "apply/transaction", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload,
		"--transaction", "--author", "test-author",
	}, writeCtx())
}

// --- apply: file not found ---

func TestApply_FileNotFound_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	runGolden(t, "apply/file_not_found", []string{
		"terminology", "--tbx", tbx, "apply", "--file", "testdata/nonexistent-payload.json",
	})
}

// --- apply: validation error (dangling crossref in payload) ---

func TestApply_ValidationError_Golden(t *testing.T) {
	tbx := copyFixture(t, "minimal-dct.tbx")
	payload := writePayloadFile(t, "payload.json", `{"concepts":[{"concept_id":"new-concept","cross_refs":[{"target":"nonexistent","label":"bad ref"}],"languages":{"en":{"preferred":{"term":"test"}}}}]}`)
	runGolden(t, "apply/validation_error", []string{
		"terminology", "--tbx", tbx, "apply", "--file", payload,
	})
}
