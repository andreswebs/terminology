package app_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/app/commands"
	"github.com/andreswebs/terminology/internal/terr"
)

func TestErrLanguageRequired_Coded(t *testing.T) {
	var coded terr.Coded = commands.ErrLanguageRequired
	if coded.Code() != "language_required" {
		t.Errorf("Code() = %q, want %q", coded.Code(), "language_required")
	}
	if coded.ExitCode() != 2 {
		t.Errorf("ExitCode() = %d, want 2", coded.ExitCode())
	}
	if coded.Hint() == "" {
		t.Error("Hint() is empty")
	}
}

func TestErrLanguageRequired_HintMentionsBothPaths(t *testing.T) {
	hint := commands.ErrLanguageRequired.Hint()
	for _, want := range []string{"--source-lang", "--target-lang", "frontmatter"} {
		if !strings.Contains(hint, want) {
			t.Errorf("Hint() = %q, missing %q", hint, want)
		}
	}
}

func TestErrLanguageRequired_Wrap_PreservesCodeAndExit(t *testing.T) {
	wrapped := commands.ErrLanguageRequired.Wrap(fmt.Errorf("for src.md"))
	if wrapped.Code() != "language_required" {
		t.Errorf("Code() = %q, want %q", wrapped.Code(), "language_required")
	}
	if wrapped.ExitCode() != 2 {
		t.Errorf("ExitCode() = %d, want 2", wrapped.ExitCode())
	}
	var asErr error = wrapped
	cause := errors.Unwrap(asErr)
	if cause == nil || cause.Error() != "for src.md" {
		t.Errorf("Unwrap() = %v, want 'for src.md'", cause)
	}
}
