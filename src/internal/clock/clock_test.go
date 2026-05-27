package clock

import (
	"context"
	"testing"
	"time"
)

type fakeClock struct {
	T time.Time
}

func (f fakeClock) Now() time.Time { return f.T }

func TestWith_From_Roundtrip(t *testing.T) {
	t.Parallel()
	fixed := time.Date(2026, 5, 21, 12, 34, 56, 0, time.UTC)
	fc := fakeClock{T: fixed}
	ctx := With(context.Background(), fc)
	got := From(ctx)
	if got != fc {
		t.Errorf("From(With(ctx, c)) returned different clock")
	}
}

func TestFrom_FallbackToReal(t *testing.T) {
	t.Parallel()
	got := From(context.Background())
	if got != Real {
		t.Errorf("From(bare ctx) = %v, want Real", got)
	}
}

func TestFrom_WrongTypeFallback(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), ctxKey{}, 42)
	got := From(ctx)
	if got != Real {
		t.Errorf("From(ctx with wrong type) = %v, want Real", got)
	}
}

func TestReal_Now_ReturnsUTC(t *testing.T) {
	t.Parallel()
	now := Real.Now()
	if now.Location() != time.UTC {
		t.Errorf("Real.Now().Location() = %v, want UTC", now.Location())
	}
}

func TestReal_Now_RecentTime(t *testing.T) {
	t.Parallel()
	before := time.Now().UTC()
	now := Real.Now()
	after := time.Now().UTC()
	if now.Before(before) || now.After(after) {
		t.Errorf("Real.Now() = %v, not between %v and %v", now, before, after)
	}
}

func TestNow_UsesContextClock(t *testing.T) {
	t.Parallel()
	fixed := time.Date(2026, 5, 21, 12, 34, 56, 0, time.UTC)
	ctx := With(context.Background(), fakeClock{T: fixed})
	got := Now(ctx)
	if !got.Equal(fixed) {
		t.Errorf("Now(ctx) = %v, want %v", got, fixed)
	}
}

func TestNow_FallbackToReal(t *testing.T) {
	t.Parallel()
	before := time.Now().UTC()
	got := Now(context.Background())
	after := time.Now().UTC()
	if got.Before(before) || got.After(after) {
		t.Errorf("Now(bare ctx) = %v, not between %v and %v", got, before, after)
	}
}
