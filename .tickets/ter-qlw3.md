---
id: ter-qlw3
status: closed
deps: [ter-6z5g]
links: []
created: 2026-05-24T00:21:45Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, errors, foundation]
---
# E2.T3 — Error sentinels (ErrUnsupportedDialect, ErrTBXLocked)

## Goal

Register the E2 error sentinels in `internal/tbx/errors.go` using the `terr` package from E1. Two sentinels are specified by E2: `ErrUnsupportedDialect` and `ErrTBXLocked`.

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Cross-process write safety" (ErrTBXLocked), §"Reader / Writer interfaces" (ErrUnsupportedDialect)
- Error-handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md) §"Layer 1 — Origin"
- E1 terr package: `src/internal/terr/terr.go` — `terr.New(code, exit, hint, format, args...)` constructor

## Files to create

- `src/internal/tbx/errors.go`
- `src/internal/tbx/errors_test.go`

## Sentinel definitions

```go
package tbx

import "github.com/andreswebs/terminology/internal/terr"

var ErrUnsupportedDialect = terr.New(
    "unsupported_dialect", 65,
    "supported: TBX-Linguist",
    "unsupported TBX dialect",
)

var ErrTBXLocked = terr.New(
    "tbx_locked", 3,
    "another terminology process is writing; retry",
    "TBX file is locked by another process",
)
```

Design notes:
- **Exit 65 = EX_DATAERR** for unsupported dialect — the input data is wrong.
- **Exit 3 = I/O class** for locked file — transient operational error.
- Both sentinels use `terr.New` so they satisfy `terr.Coded` and integrate with `output.EmitError` and `output.ExitCodeFor`.
- `ErrTBXLocked` is used via `.Wrap(err)` in the lock-acquisition code (T13) to preserve the underlying OS error while keeping the typed code.

## Deviation note

The current implementation also includes `ErrNoTBXPath` (exit 2) and `ErrValidationError` (exit 65) in this file. Per the spec, `ErrNoTBXPath` is a usage error that belongs in the command layer (urfave flag validation), not in the domain package. `ErrValidationError` belongs to E3. This ticket creates only the two sentinels specified by E2. If an agent finds `ErrNoTBXPath` or `ErrValidationError` already present, leave them — they will be addressed by their respective epics.

## TDD cycles

### Cycle 1 — ErrUnsupportedDialect
RED: Test that ErrUnsupportedDialect satisfies terr.Coded, returns Code()=="unsupported_dialect", ExitCode()==65, Hint()=="supported: TBX-Linguist".
GREEN: Define the sentinel.

### Cycle 2 — ErrTBXLocked
RED: Test that ErrTBXLocked satisfies terr.Coded, returns Code()=="tbx_locked", ExitCode()==3.
GREEN: Define the sentinel.

### Cycle 3 — Wrap preserves code
RED: Test that ErrTBXLocked.Wrap(fmt.Errorf("os error")).Code() == "tbx_locked" and errors.Is(wrapped, ErrTBXLocked) behavior (via Unwrap chain).
GREEN: Already satisfied by terr.E.Wrap from E1 — this cycle just confirms integration.

## Out of scope

- ErrNoTBXPath (command-layer concern)
- ErrValidationError (E3)
- Lock acquisition logic (T13)

## Acceptance

- `make build` passes
- Both sentinels satisfy `terr.Coded`
- Exit codes match spec (65, 3)
- Hint strings match spec


## Notes

**2026-05-25T13:09:49Z**

Added errors_test.go with three test cases covering all acceptance criteria: (1) ErrUnsupportedDialect satisfies terr.Coded with code=unsupported_dialect, exit=65, hint=supported: TBX-Linguist; (2) ErrTBXLocked satisfies terr.Coded with code=tbx_locked, exit=3; (3) ErrTBXLocked.Wrap preserves code and supports errors.Is unwrap chain. The sentinels were already correctly defined in errors.go. ErrNoTBXPath and ErrValidationError left in place per deviation note. make build passes clean.
