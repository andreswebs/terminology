package write

import (
	"testing"

	"github.com/andreswebs/terminology/internal/terr"
)

func TestErrorSentinels(t *testing.T) {
	tests := []struct {
		name     string
		sentinel *terr.E
		code     string
		exit     int
		wantHint bool
	}{
		{"ErrInvalidID", ErrInvalidID, "invalid_id", 65, true},
		{"ErrDuplicateID", ErrDuplicateID, "duplicate_id", 65, true},
		{"ErrNotFound", ErrNotFound, "not_found", 65, true},
		{"ErrDanglingCrossref", ErrDanglingCrossref, "dangling_crossref", 65, true},
		{"ErrInvalidInput", ErrInvalidInput, "invalid_input", 65, true},
		{"ErrApplyValidationFailed", ErrApplyValidationFailed, "apply_validation_failed", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var coded terr.Coded = tt.sentinel
			if coded.Code() != tt.code {
				t.Errorf("Code() = %q, want %q", coded.Code(), tt.code)
			}
			if coded.ExitCode() != tt.exit {
				t.Errorf("ExitCode() = %d, want %d", coded.ExitCode(), tt.exit)
			}
			if tt.wantHint && coded.Hint() == "" {
				t.Error("Hint() is empty, want non-empty")
			}
			if coded.Error() == "" {
				t.Error("Error() is empty, want non-empty")
			}
		})
	}
}

func TestErrorSentinels_RegisteredInRegistry(t *testing.T) {
	all := terr.All()
	expected := map[string]bool{
		"invalid_id":              false,
		"duplicate_id":            false,
		"not_found":               false,
		"dangling_crossref":       false,
		"invalid_input":           false,
		"apply_validation_failed": false,
	}
	for _, e := range all {
		if _, ok := expected[e.Code()]; ok {
			expected[e.Code()] = true
		}
	}
	for code, found := range expected {
		if !found {
			t.Errorf("sentinel %q not found in terr.All() registry", code)
		}
	}
}
