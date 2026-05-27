---
id: ter-pwwn
status: closed
deps: [ter-ab56]
links: []
created: 2026-05-25T19:37:19Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, terr, foundation]
---
# E4.T4 — Terr sentinel registry

## Goal

Extend `internal/terr` with an auto-collection registry that captures every `terr.New(...)` sentinel at init time. The `schema` command needs to enumerate all error codes without hardcoding a list. The sentinel-as-registry pattern from the schema-source-of-truth ADR requires that every sentinel self-registers.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"schema — reflective introspection" source 3
- Schema-source-of-truth ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md) §"Error registry & generated reference"

## Files to create / modify

- `src/internal/terr/terr.go` — add registry (package-level slice + `Register`/`All` functions)
- `src/internal/terr/terr_test.go` — test registry collection

## Behavior

### Registry API

```go
func Register(e *E) *E   // called by New(), returns e for var declaration chaining
func All() []*E           // returns a copy of all registered sentinels
```

### Approach

Modify `terr.New()` to auto-register each sentinel it creates. Every `var ErrFoo = terr.New(...)` across the codebase automatically appears in `All()`. No manual registration needed.

```go
var registry []*E

func New(code string, exit int, hint, format string, args ...any) *E {
    e := &E{
        code: code,
        exit: exit,
        hint: hint,
        msg:  fmt.Sprintf(format, args...),
    }
    registry = append(registry, e)
    return e
}

func All() []*E {
    cp := make([]*E, len(registry))
    copy(cp, registry)
    return cp
}
```

This is safe because `New()` is only called during `var` initialization (package init time), which is single-threaded in Go.

### What gets registered

All existing sentinels across the codebase:
- `tbx.ErrUnsupportedDialect` (exit 65)
- `tbx.ErrTBXLocked` (exit 3)
- `tbx.ErrNoTBXPath` (exit 2)
- `tbx.ErrValidationError` (exit 65)
- `output.ErrInvalidField` (exit 2) — from T1

`terr.UnderConstruction()` is NOT registered because it's a function call, not a `var` sentinel.

## TDD cycles

### Cycle 1 — New() registers sentinel
RED: Call `terr.New(...)`, then `terr.All()`. Assert the returned slice contains the new sentinel.
GREEN: Add registry slice and append in `New()`.

### Cycle 2 — All() returns a copy
RED: Modify the slice returned by `All()`, call `All()` again. Assert original is unchanged.
GREEN: Return `copy` in `All()`.

### Cycle 3 — Existing sentinels appear in registry
RED: Import `tbx` and `output` packages, call `terr.All()`. Assert slice contains entries with codes `"unsupported_dialect"`, `"tbx_locked"`, `"no_tbx_path"`, `"validation_error"`.
GREEN: Already passing — existing `var ErrX = terr.New(...)` calls auto-register.

## Acceptance

- `make build` passes
- `terr.All()` returns all sentinels declared via `terr.New()`
- `All()` returns a fresh copy each call
- No manual registration needed — `New()` auto-registers

## Notes

**2026-05-25T19:50:44Z**

Implemented terr sentinel registry. New() auto-registers; added Newf() as non-registering constructor for runtime error creation. Changed UnderConstruction() to construct E directly instead of calling New(). Migrated runtime terr.New() calls in argcheck.go and root.go:116 to Newf(). Integration test in app/registry_test.go verifies well-known sentinels (unsupported_dialect, tbx_locked, no_tbx_path, validation_error, conflicting_verbosity, no_subcommand) appear in All(). Unit tests verify: New registers, All returns copy, Newf and UnderConstruction do not register.
