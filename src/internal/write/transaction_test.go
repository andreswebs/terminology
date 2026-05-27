package write

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/andreswebs/terminology/internal/clock"
	"github.com/andreswebs/terminology/internal/logctx"
)

type fakeClock struct{ T time.Time }

func (fc fakeClock) Now() time.Time { return fc.T }

func TestNewTransaction_WithAuthor(t *testing.T) {
	ts := time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC)
	ctx := clock.With(context.Background(), fakeClock{T: ts})

	tx := NewTransaction(ctx, "andre")

	if tx.Type != "modification" {
		t.Errorf("Type = %q, want %q", tx.Type, "modification")
	}
	if tx.Date != "2025-03-15T10:30:00Z" {
		t.Errorf("Date = %q, want %q", tx.Date, "2025-03-15T10:30:00Z")
	}
	if tx.Responsibility != "andre" {
		t.Errorf("Responsibility = %q, want %q", tx.Responsibility, "andre")
	}
}

func TestNewTransaction_MissingAuthor(t *testing.T) {
	ts := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	ctx := clock.With(context.Background(), fakeClock{T: ts})

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	ctx = logctx.With(ctx, logger)

	tx := NewTransaction(ctx, "")

	if tx.Type != "modification" {
		t.Errorf("Type = %q, want %q", tx.Type, "modification")
	}
	if tx.Date != "2025-06-01T00:00:00Z" {
		t.Errorf("Date = %q, want %q", tx.Date, "2025-06-01T00:00:00Z")
	}
	if tx.Responsibility != "" {
		t.Errorf("Responsibility = %q, want empty", tx.Responsibility)
	}

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("expected WARN log for missing author, got no log output")
	}
	if !bytes.Contains(buf.Bytes(), []byte("missing author")) {
		t.Errorf("log output should mention missing author, got: %s", logOutput)
	}
}

func TestNewTransaction_DeterministicTimestamp(t *testing.T) {
	ts := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	ctx := clock.With(context.Background(), fakeClock{T: ts})

	tx1 := NewTransaction(ctx, "agent")
	tx2 := NewTransaction(ctx, "agent")

	if tx1.Date != tx2.Date {
		t.Errorf("non-deterministic: %q != %q", tx1.Date, tx2.Date)
	}
	if tx1.Date != "2025-12-31T23:59:59Z" {
		t.Errorf("Date = %q, want %q", tx1.Date, "2025-12-31T23:59:59Z")
	}
}
