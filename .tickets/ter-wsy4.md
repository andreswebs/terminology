---
id: ter-wsy4
status: closed
deps: [ter-bedf]
links: []
created: 2026-05-24T01:05:13Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, error, foundation]
---
# E3.T3 — ErrValidationError sentinel

## Goal

Define the `ErrValidationError` sentinel error in `internal/tbx/errors.go`. This is the error returned when a TBX file fails validation — either tier-1 well-formedness failure (can't parse) or tier-2/3 errors promoted by `--strict`.

Per the [error-handling ADR](docs/adr/error-handling.md), sentinels live with their producing package. `ErrValidationError` is produced by the `tbx` package (validation logic) and consumed by the validate command.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Exit codes" — exit 65 = `EX_DATAERR`
- Error-handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md) §"Sentinel errors", §"Exit code mapping"
- `terr` package: [src/internal/terr/terr.go](src/internal/terr/terr.go) — `terr.New(code, exit, hint, message)`

## Files to modify

- `src/internal/tbx/errors.go` — add `ErrValidationError`

## Definition

```go
var ErrValidationError = terr.New(
    "validation_error", 65,
    "check the TBX file structure and content",
    "TBX validation failed",
)
```

- **Code**: `"validation_error"` — appears in the JSON error envelope `error.code` field.
- **Exit code**: `65` — BSD `EX_DATAERR`, unified across `validate`, `apply`, and write-side runtime-input-rejection.
- **Hint**: user-facing guidance for resolution.
- **Message**: internal description.

The sentinel supports `.Wrap(cause)` to chain the underlying parse or validation error.

## TDD cycles

### Cycle 1 — Sentinel exists with correct code
RED: Test that `tbx.ErrValidationError` is non-nil and satisfies `terr.Coded` interface with code `"validation_error"`.
GREEN: Define the sentinel in errors.go.

### Cycle 2 — Exit code is 65
RED: Test that `tbx.ErrValidationError.ExitCode()` returns 65.
GREEN: Already passing from cycle 1.

### Cycle 3 — Wrap preserves code
RED: Wrap a dummy error with `ErrValidationError.Wrap(fmt.Errorf("parse failed"))`. Assert the wrapped error still satisfies `terr.Coded` with code `"validation_error"` and exit 65.
GREEN: Already passing — `terr.E.Wrap` copies the error and sets the cause.

## Deviation note

The current implementation already has `ErrValidationError` defined in `src/internal/tbx/errors.go` with the exact definition above. It also defines `ErrNoTBXPath` (exit 2) in the same file — that sentinel belongs to E1 (CLI surface) rather than E3, but its presence doesn't conflict. No changes are needed to align with the spec.

## Out of scope

- Other error sentinels (`ErrUnsupportedDialect`, `ErrTBXLocked` — E2)
- `ErrNoTBXPath` — E1 concern, already present
- The `terr` package itself

## Acceptance

- `make build` passes
- `ErrValidationError` exists in `internal/tbx/errors.go`
- Code is `"validation_error"`, exit code is 65
- Satisfies `terr.Coded` interface
- `.Wrap(cause)` preserves code and exit code

