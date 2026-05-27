package logctx

import (
	"context"
	"log/slog"
	"regexp"
	"testing"
)

func TestWith_From_Roundtrip(t *testing.T) {
	t.Parallel()
	l := slog.New(slog.NewTextHandler(nil, nil))
	ctx := With(context.Background(), l)
	got := From(ctx)
	if got != l {
		t.Errorf("From(With(ctx, l)) returned different logger")
	}
}

func TestFrom_FallbackToDefault(t *testing.T) {
	t.Parallel()
	got := From(context.Background())
	if got != slog.Default() {
		t.Errorf("From(bare ctx) = %p, want slog.Default() = %p", got, slog.Default())
	}
}

func TestFrom_WrongTypeFallback(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), ctxKey{}, 42)
	got := From(ctx)
	if got != slog.Default() {
		t.Errorf("From(ctx with wrong type) = %p, want slog.Default() = %p", got, slog.Default())
	}
}

var hexPattern = regexp.MustCompile(`^[0-9a-f]{16}$`)

func TestNewRunID_FormatAndLength(t *testing.T) {
	t.Parallel()
	id := NewRunID()
	if len(id) != 16 {
		t.Errorf("len(NewRunID()) = %d, want 16", len(id))
	}
	if !hexPattern.MatchString(id) {
		t.Errorf("NewRunID() = %q, want 16-char lowercase hex", id)
	}
}

func TestNewRunID_Uniqueness(t *testing.T) {
	t.Parallel()
	seen := make(map[string]bool, 1000)
	for range 1000 {
		id := NewRunID()
		if seen[id] {
			t.Fatalf("duplicate run ID: %s", id)
		}
		seen[id] = true
	}
}
