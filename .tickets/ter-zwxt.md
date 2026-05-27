---
id: ter-zwxt
status: closed
deps: []
links: []
created: 2026-05-22T19:41:34Z
type: task
priority: 1
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, foundation, terr]
---
# E1.T1 — Foundation: internal/terr

## Goal

Stand up `internal/terr` — the typed-error foundation every other package builds on. Sentinels live in their producing packages; this ticket lands the type, the constructor, the `Coded` interface, and the well-known `UnderConstruction` sentinel used by every stub action.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"main.go shape" + §"Under-construction contract"
- Error-handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md) §"Layer 1 — Origin (typed, coded errors)"
- Q1 decision: stub exit code 75 (EX_TEMPFAIL) + sentinel code `under_construction`

## Files to create

- `src/internal/terr/terr.go`
- `src/internal/terr/terr_test.go`

## Type & API

```go
package terr

import "fmt"

// Coded is the interface every typed error in this codebase satisfies.
// urfave/cli/v3's ExitCoder is a subset.
type Coded interface {
    error
    Code() string
    ExitCode() int
    Hint() string
}

type E struct {
    code, msg, hint string
    exit            int
    cause           error
}

func New(code string, exit int, hint, format string, args ...any) *E
func (e *E) Error() string
func (e *E) Code() string
func (e *E) Hint() string
func (e *E) ExitCode() int
func (e *E) Unwrap() error
func (e *E) Wrap(cause error) *E  // returns copy with cause set; original Code/Hint/Exit preserved

// UnderConstruction is the well-known stub sentinel used by every
// not-yet-implemented action handler. cmdPath is urfave's
// cli.Command.FullName() (spaces preserved, e.g. "concept add").
func UnderConstruction(cmdPath string) *E
```

`UnderConstruction` returns: code=`under_construction`, exit=`75`, hint=`track progress in .tickets/ or rebuild from a newer commit`, message=`terminology <cmdPath> is not implemented yet`.

## TDD plan (vertical slices — one cycle per test, never write all tests first)

Tests live in `src/internal/terr/terr_test.go`. Table-driven where shape repeats (per [docs/adr/testing.md](docs/adr/testing.md) §"Unit tests").

1. **RED** `TestE_Accessors` — `New(\"x\", 2, \"h\", \"msg %d\", 1)` then assert each accessor returns the expected value. **GREEN** implement constructor + accessors.
2. **RED** `TestE_ImplementsCoded` — `var _ Coded = (*E)(nil)` (compile-time assertion as a test body). **GREEN** ensures alignment.
3. **RED** `TestE_Wrap_PreservesIdentity` — construct *E, call Wrap(io.EOF), assert `errors.Is(wrapped, io.EOF)` AND wrapped.Code()/Hint()/ExitCode() are unchanged AND wrapped != original (different pointer). **GREEN** implement Wrap.
4. **RED** `TestE_ErrorsAs_ThroughFmtErrorf` — wrap via `fmt.Errorf(\"ctx: %w\", e)`, assert `errors.As(outer, &target)` extracts the *E with intact fields. **GREEN** covered by Unwrap.
5. **RED** `TestUnderConstruction` — table-driven for cmdPath in {\"validate\", \"concept add\", \"term deprecate\"}; assert code/exit/hint constants and that message contains the cmdPath verbatim. **GREEN** implement UnderConstruction.

## Acceptance

- `make build` from project root passes (full quality gate)
- `cd src && go test ./internal/terr/...` passes
- No new dependencies in go.mod

## Out of scope

- Sentinel-registry collection for `terminology schema` reflection (deferred to E4).
- Any non-stub error codes; subsequent epics add sentinels alongside their packages.
- Envelope rendering — that lives in `internal/output` (T2).
- Logging integration — T4 owns `logctx`, T5 wires the Before hook.


## Notes

**2026-05-22T20:38:24Z**

Implemented internal/terr package via TDD (5 red-green cycles): E type with New constructor, Code/Hint/ExitCode/Error/Unwrap accessors, Wrap method (returns copy with cause, preserves identity), and UnderConstruction sentinel factory (code=under_construction, exit=75). Compile-time Coded interface assertion. All tests pass, make build passes, no new dependencies.
