package tbx_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/terr"
)

func TestErrUnsupportedDialect(t *testing.T) {
	var coded terr.Coded = tbx.ErrUnsupportedDialect
	if got := coded.Code(); got != "unsupported_dialect" {
		t.Errorf("Code() = %q, want %q", got, "unsupported_dialect")
	}
	if got := coded.ExitCode(); got != 65 {
		t.Errorf("ExitCode() = %d, want %d", got, 65)
	}
	if got := coded.Hint(); got != "supported: TBX-Linguist" {
		t.Errorf("Hint() = %q, want %q", got, "supported: TBX-Linguist")
	}
}

func TestErrTBXLocked(t *testing.T) {
	var coded terr.Coded = tbx.ErrTBXLocked
	if got := coded.Code(); got != "tbx_locked" {
		t.Errorf("Code() = %q, want %q", got, "tbx_locked")
	}
	if got := coded.ExitCode(); got != 3 {
		t.Errorf("ExitCode() = %d, want %d", got, 3)
	}
	if got := coded.Hint(); got != "another terminology process is writing; retry" {
		t.Errorf("Hint() = %q, want %q", got, "another terminology process is writing; retry")
	}
}

func TestErrTBXLocked_Wrap_PreservesCode(t *testing.T) {
	osErr := fmt.Errorf("file locked by pid 1234")
	wrapped := tbx.ErrTBXLocked.Wrap(osErr)

	if got := wrapped.Code(); got != "tbx_locked" {
		t.Errorf("Code() = %q, want %q", got, "tbx_locked")
	}
	if got := wrapped.ExitCode(); got != 3 {
		t.Errorf("ExitCode() = %d, want %d", got, 3)
	}
	if !errors.Is(wrapped, osErr) {
		t.Error("errors.Is(wrapped, osErr) = false, want true")
	}
}
