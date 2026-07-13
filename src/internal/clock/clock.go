// Package clock provides an injectable time source, allowing the current
// time to be carried through a context and substituted in tests.
package clock

import (
	"context"
	"time"
)

// Clock reports the current time.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// Real is the default Clock, reporting the current wall-clock time in UTC.
var Real Clock = realClock{}

type ctxKey struct{}

// With returns a copy of ctx carrying the given Clock.
func With(ctx context.Context, c Clock) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// From returns the Clock stored in ctx, or Real if none is present.
func From(ctx context.Context) Clock {
	if c, ok := ctx.Value(ctxKey{}).(Clock); ok {
		return c
	}
	return Real
}

// Now returns the current time according to the Clock stored in ctx.
func Now(ctx context.Context) time.Time { return From(ctx).Now() }
