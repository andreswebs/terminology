package app_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	_ "github.com/andreswebs/terminology/internal/app"
	"github.com/andreswebs/terminology/internal/output"
)

// nonCommandGoldenDirs are top-level golden directories that do not correspond
// to a single leaf command with a registered exit-code set: shared fixtures,
// the cross-command sanitize suite, and the bare parent commands (which only
// ever emit the usage-level no_subcommand error).
var nonCommandGoldenDirs = map[string]bool{
	"fixtures": true,
	"sanitize": true,
	"concept":  true,
	"term":     true,
}

// TestExitCodeRegistryConformance ties the per-command exit-code registry to
// real behavior: it walks the golden directories, derives the command from each
// directory path, and asserts every recorded clean.exit value is a member of
// the registered set for that command (ADR 0001).
func TestExitCodeRegistryConformance(t *testing.T) {
	root := "testdata"

	var skipped []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "clean.exit" {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		top := strings.Split(rel, string(filepath.Separator))[0]
		if nonCommandGoldenDirs[top] {
			skipped = append(skipped, rel)
			return nil
		}

		// Golden directories join a multi-word command with an underscore
		// (e.g. concept_add -> "concept add"); the registry is keyed by the
		// space-separated command name.
		command := strings.ReplaceAll(top, "_", " ")

		registered, ok := output.ExitCodesFor(command)
		if !ok {
			t.Errorf("%s: no registered exit codes for command %q", rel, command)
			return nil
		}

		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		got, convErr := strconv.Atoi(strings.TrimSpace(string(raw)))
		if convErr != nil {
			t.Errorf("%s: unparseable exit value %q: %v", rel, raw, convErr)
			return nil
		}

		if !containsInt(registered, got) {
			t.Errorf("%s: exit %d not in registered set %v for command %q",
				rel, got, registered, command)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}

	if len(skipped) > 0 {
		t.Logf("skipped %d non-command golden(s): %s",
			len(skipped), strings.Join(skipped, ", "))
	}
}

func containsInt(xs []int, want int) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
