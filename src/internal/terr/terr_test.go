package terr_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/andreswebs/terminology/internal/terr"
)

var _ terr.Coded = (*terr.E)(nil)

func TestE_Accessors(t *testing.T) {
	e := terr.New("x", 2, "h", "msg %d", 1)

	if got := e.Error(); got != "msg 1" {
		t.Errorf("Error() = %q, want %q", got, "msg 1")
	}
	if got := e.Code(); got != "x" {
		t.Errorf("Code() = %q, want %q", got, "x")
	}
	if got := e.ExitCode(); got != 2 {
		t.Errorf("ExitCode() = %d, want %d", got, 2)
	}
	if got := e.Hint(); got != "h" {
		t.Errorf("Hint() = %q, want %q", got, "h")
	}
}

func TestE_Wrap_PreservesIdentity(t *testing.T) {
	original := terr.New("test_code", 3, "some hint", "original message")
	wrapped := original.Wrap(io.EOF)

	if !errors.Is(wrapped, io.EOF) {
		t.Error("errors.Is(wrapped, io.EOF) = false, want true")
	}
	if wrapped.Code() != original.Code() {
		t.Errorf("Code() = %q, want %q", wrapped.Code(), original.Code())
	}
	if wrapped.Hint() != original.Hint() {
		t.Errorf("Hint() = %q, want %q", wrapped.Hint(), original.Hint())
	}
	if wrapped.ExitCode() != original.ExitCode() {
		t.Errorf("ExitCode() = %d, want %d", wrapped.ExitCode(), original.ExitCode())
	}
	if wrapped == original {
		t.Error("Wrap returned same pointer, want a copy")
	}
}

func TestE_ErrorsAs_ThroughFmtErrorf(t *testing.T) {
	e := terr.New("wrap_code", 65, "a hint", "inner error")
	outer := fmt.Errorf("context: %w", e)

	var target *terr.E
	if !errors.As(outer, &target) {
		t.Fatal("errors.As did not extract *terr.E from wrapped error")
	}
	if target.Code() != "wrap_code" {
		t.Errorf("Code() = %q, want %q", target.Code(), "wrap_code")
	}
	if target.ExitCode() != 65 {
		t.Errorf("ExitCode() = %d, want %d", target.ExitCode(), 65)
	}
	if target.Hint() != "a hint" {
		t.Errorf("Hint() = %q, want %q", target.Hint(), "a hint")
	}
}

func TestNew_RegistersSentinel(t *testing.T) {
	before := len(terr.All())
	s := terr.New("test_registry_code", 2, "hint", "msg")
	after := terr.All()
	if len(after) != before+1 {
		t.Fatalf("All() length = %d, want %d", len(after), before+1)
	}
	if after[len(after)-1] != s {
		t.Error("last entry in All() is not the sentinel just created")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	a := terr.All()
	if len(a) == 0 {
		t.Fatal("All() returned empty slice")
	}
	a[0] = nil
	b := terr.All()
	if b[0] == nil {
		t.Error("modifying All() result affected subsequent call; want independent copy")
	}
}

func TestNewf_NotRegistered(t *testing.T) {
	before := len(terr.All())
	e := terr.Newf("runtime_code", 2, "hint", "msg %s", "arg")
	after := len(terr.All())
	if after != before {
		t.Errorf("Newf registered a sentinel: All() grew from %d to %d", before, after)
	}
	if got := e.Code(); got != "runtime_code" {
		t.Errorf("Code() = %q, want %q", got, "runtime_code")
	}
	if got := e.Error(); got != "msg arg" {
		t.Errorf("Error() = %q, want %q", got, "msg arg")
	}
}
