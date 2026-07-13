// Package logctx carries a structured logger through a context and generates
// identifiers for correlating log entries within a single run.
package logctx

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

// With returns a copy of ctx carrying the given logger.
func With(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// From returns the logger stored in ctx, or slog.Default if none is present.
func From(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
