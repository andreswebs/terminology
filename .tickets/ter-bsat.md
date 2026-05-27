---
id: ter-bsat
status: closed
deps: []
links: []
created: 2026-05-22T19:41:34Z
type: task
priority: 1
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, foundation, logctx]
---
# E1.T4 — Foundation: internal/logctx (slog on context + run-id)

## Goal

Stand up `internal/logctx` — a context-threaded `*slog.Logger` accessor plus the run-id generator that uniquely tags every CLI invocation.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Logger bootstrap"
- Logging ADR: [docs/adr/logging.md](docs/adr/logging.md) §"Run ID" + §"Per-package logger access"

## Files to create

- `src/internal/logctx/logctx.go` — `With`, `From`, `ctxKey`
- `src/internal/logctx/runid.go` — `NewRunID` (exported for use by T5's Before hook)
- `src/internal/logctx/logctx_test.go` — unit tests

## API

```go
package logctx

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "log/slog"
)

type ctxKey struct{}

// With returns a child context carrying logger l. Idiomatic logger
// propagation: parent code constructs a logger once and threads it
// through downstream packages via ctx.
func With(ctx context.Context, l *slog.Logger) context.Context

// From returns the logger stashed in ctx, or slog.Default() when none
// is present. The fallback keeps unit tests that don't construct a
// context functional.
func From(ctx context.Context) *slog.Logger

// NewRunID returns 8 bytes of crypto/rand encoded as 16-char hex.
// 64 bits of entropy is ample for correlating one-shot CLI runs.
func NewRunID() string
```

## TDD plan

Tests in `src/internal/logctx/logctx_test.go`.

1. **RED** `TestWith_From_Roundtrip` — construct a non-default *slog.Logger; `ctx := logctx.With(context.Background(), l)`; assert `logctx.From(ctx) == l` (pointer identity). **GREEN** implement With + From + ctxKey.
2. **RED** `TestFrom_FallbackToDefault` — pass `context.Background()` directly; assert `logctx.From(ctx) == slog.Default()` (pointer identity is OK here — slog.Default returns the same pointer every call within a process). **GREEN** the type-assertion miss path.
3. **RED** `TestFrom_WrongTypeFallbackToDefault` — manually stash a non-*slog.Logger value under the key (need to expose a test helper that returns the key, or use `context.WithValue(ctx, logctx.Key(), 42)` — pick whichever doesn't leak the key publicly; simplest is to define an unexported helper test-file in the same package). **GREEN** robust type assertion.
4. **RED** `TestNewRunID_FormatAndLength` — call once; assert `len(id) == 16` and `regexp.MustCompile(\`^[0-9a-f]{16}\$\`).MatchString(id)`. **GREEN** implement NewRunID.
5. **RED** `TestNewRunID_Uniqueness` — generate 1000 IDs; assert all unique (map[string]bool). **GREEN** covered by entropy.

## Acceptance

- `make build` clean
- `cd src && go test ./internal/logctx/...` passes
- Stdlib only (`context`, `crypto/rand`, `encoding/hex`, `log/slog`)

## Out of scope

- Logger construction (level + handler selection + attributes) — T5 owns the Before-hook bootstrap.
- TTY detection via `golang.org/x/term` — T5 (it's a Before-hook concern, not a logctx concern).
- The verbosity-mutex check — T5 (uses Before hook).
- Sampling, redaction, `--log-file` — explicit non-goals per logging ADR.


## Notes

**2026-05-22T20:46:26Z**

Implemented internal/logctx with three files: logctx.go (With/From context-threaded logger), runid.go (NewRunID 8-byte crypto/rand hex), logctx_test.go (5 tests: roundtrip, default fallback, wrong-type fallback, format/length, uniqueness). Stdlib only — no external deps. All tests parallel. make build clean.
