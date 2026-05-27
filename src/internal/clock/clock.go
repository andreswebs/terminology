package clock

import (
	"context"
	"time"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

var Real Clock = realClock{}

type ctxKey struct{}

func With(ctx context.Context, c Clock) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

func From(ctx context.Context) Clock {
	if c, ok := ctx.Value(ctxKey{}).(Clock); ok {
		return c
	}
	return Real
}

func Now(ctx context.Context) time.Time { return From(ctx).Now() }
