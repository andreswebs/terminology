---
id: ter-anta
status: closed
deps: [ter-ydak, ter-6z5g]
links: []
created: 2026-05-24T00:22:36Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, interfaces, foundation]
---
# E2.T4 — Reader/Writer interfaces + dialect registry

## Goal

Define the Reader and Writer interfaces that abstract over TBX dialect implementations, and the dialect registry that dispatches to concrete readers/writers based on the dialect identifier.

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Reader / Writer interfaces", §"Architecture" (reader.go, writer.go)
- Schema-source-of-truth ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md) — interfaces are part of the contract

## Files to create

- `src/internal/tbx/reader.go`
- `src/internal/tbx/writer.go`
- `src/internal/tbx/registry.go` (dialect dispatch — registry portion currently in io.go)
- `src/internal/tbx/registry_test.go`

## Interface definitions

### reader.go

```go
package tbx

import "io"

type Reader interface {
    Decode(r io.Reader) (*Glossary, []Warning, error)
}
```

### writer.go

```go
package tbx

import "io"

type Writer interface {
    Encode(w io.Writer, g *Glossary) error
}
```

### registry.go

```go
package tbx

var (
    readers = map[Dialect]func() Reader{}
    writers = map[Dialect]func() Writer{}
)

func RegisterDialect(d Dialect, rf func() Reader, wf func() Writer) {
    readers[d] = rf
    writers[d] = wf
}

func readerFor(d Dialect) (Reader, error) {
    rf, ok := readers[d]
    if !ok {
        return nil, ErrUnsupportedDialect
    }
    return rf(), nil
}

func writerFor(d Dialect) (Writer, error) {
    wf, ok := writers[d]
    if !ok {
        return nil, ErrUnsupportedDialect
    }
    return wf(), nil
}
```

Design notes:
- **Reader.Decode returns `(*Glossary, []Warning, error)`** — warnings are data, not errors. A file can parse successfully but produce warnings (legacy forms, unknown elements).
- **Writer.Encode takes `(io.Writer, *Glossary)`** — writes canonical output. No warnings; the writer is deterministic.
- **Registry uses factory functions** (`func() Reader`) not singletons, so each call gets a fresh instance (safe for concurrent use).
- **`readerFor`/`writerFor` are unexported** — they're consumed only by Load/Save in T12/T13.
- **`RegisterDialect` is exported** for use by dialect packages in their `init()`.

## Deviation note

The current implementation puts the registry maps and functions inside `io.go` rather than a separate `registry.go`. The spec's architecture diagram shows `reader.go`, `writer.go` as separate files for the interfaces. Separating registry logic into its own file improves clarity — `io.go` handles file I/O (Load/Save), `registry.go` handles dispatch.

## TDD cycles

### Cycle 1 — RegisterDialect + readerFor
RED: Register a mock dialect with a stub Reader factory. Call readerFor with that dialect. Assert the returned Reader is non-nil.
GREEN: Implement RegisterDialect and readerFor.

### Cycle 2 — Unregistered dialect returns ErrUnsupportedDialect
RED: Call readerFor with an unregistered dialect. Assert errors.Is(err, ErrUnsupportedDialect).
GREEN: Implement the not-found branch.

### Cycle 3 — writerFor
RED: Same pattern as cycles 1-2 for writerFor.
GREEN: Implement writerFor.

## Out of scope

- Load/Save file I/O (T12, T13)
- detectDialect (T12)
- Concrete reader/writer implementations (T6–T10)

## Acceptance

- `make build` passes
- Reader and Writer interfaces are exported
- RegisterDialect registers factories
- readerFor/writerFor dispatch correctly and return ErrUnsupportedDialect for unknown dialects


## Notes

**2026-05-25T13:12:06Z**

Extracted registry logic (RegisterDialect, readerFor, writerFor, maps) from io.go into registry.go. Added unregisterDialect helper (unexported, test-only cleanup). Created registry_test.go with 5 tests: registered reader/writer dispatch, unregistered dialect returns ErrUnsupportedDialect, and factory-fresh-instances verification. Reader/Writer interfaces (reader.go, writer.go) were already in their own files. io.go now contains only Load/Save/detectDialect/acquireLock. make build passes cleanly.
