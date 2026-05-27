---
id: ter-yfoe
status: closed
deps: [ter-qlw3, ter-anta, ter-2398, ter-6z5g]
links: []
created: 2026-05-24T00:29:51Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, io, foundation]
---
# E2.T12 — Load + detectDialect (io.go)

## Goal

Implement \`tbx.Load(path)\` — the top-level entry point for reading a TBX file. Load opens the file, detects the dialect from the \`<tbx>\` element's \`@type\` attribute, dispatches to the appropriate registered Reader, and returns the parsed glossary.

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Reader / Writer interfaces" — "Top-level \`tbx.Load(path)\` opens the file, inspects \`<tbx>\` \`@type\`, picks a registered Reader, and dispatches"
- E2 spec: §"Cross-process write safety" — "Read commands do NOT acquire the lock"

## Files to create/modify

- \`src/internal/tbx/io.go\` — Load function and detectDialect helper
- \`src/internal/tbx/io_test.go\`

## API

```go
package tbx

func Load(path string) (*Glossary, []Warning, error)
```

Internal:
```go
func detectDialect(r io.Reader) (Dialect, error)
```

Design notes:
- **detectDialect** streams the XML looking for the first \`<tbx>\` start element, reads \`@type\` attribute, maps it to a Dialect constant. Returns ErrUnsupportedDialect for unknown types or missing attribute.
- **Load** opens the file, calls detectDialect, seeks back to start, calls the registered Reader's Decode.
- **No lock acquired** — reads don't need exclusive access. \`os.Open\` gives consistency for the read duration via the kernel's open-file table.
- **Error wrapping**: OS errors (file not found, permission denied) are wrapped with context (\`"opening TBX: %w"\`). Dialect errors propagate as-is (they're already terr.Coded).

## TDD cycles

### Cycle 1 — Load minimal DCT
RED: Load("testdata/minimal-dct.tbx") returns a Glossary with dialect==DialectLinguist, len(Concepts)==1. (Requires linguist blank import for registration.)
GREEN: Implement Load with detectDialect.

### Cycle 2 — Unsupported dialect
RED: Create a temp file with \`<tbx type="TBX-Basic">\`. Load it. Assert error is ErrUnsupportedDialect.
GREEN: detectDialect returns ErrUnsupportedDialect for unknown types.

### Cycle 3 — Missing type attribute
RED: Create a temp file with \`<tbx>\` (no type attr). Load it. Assert ErrUnsupportedDialect.
GREEN: detectDialect handles missing attribute.

### Cycle 4 — File not found
RED: Load("nonexistent.tbx"). Assert error is non-nil and wraps an os.PathError.
GREEN: Handle os.Open error.

### Cycle 5 — No \`<tbx>\` element (malformed XML)
RED: Create a temp file with \`<root></root>\`. Load it. Assert error.
GREEN: detectDialect returns error when it reaches EOF without finding \`<tbx>\`.

## Deviation note

The current implementation puts Load, Save, detectDialect, acquireLock, and all registry functions in a single \`io.go\`. Per the spec's architecture, the file layout separates concerns: \`reader.go\`/\`writer.go\` for interfaces, \`io.go\` for Load/Save, \`lock.go\` for locking. The registry portion was moved to \`registry.go\` in T4. This ticket covers only Load and detectDialect. Save + locking are in T13.

## Out of scope

- Save / atomic write (T13)
- Lock acquisition (T13)
- Registry (T4)
- Reader implementation (T6-T8)

## Acceptance

- \`make build\` passes
- Load returns parsed Glossary for valid TBX-Linguist files
- Load returns ErrUnsupportedDialect for non-Linguist TBX files
- Load returns appropriate errors for missing/malformed files
- No lock is acquired during Load


## Notes

**2026-05-25T13:42:19Z**

Load + detectDialect implementation was already in place from prior work. Added missing test coverage for 3 of the 5 TDD cycles: TestLoad_MissingTypeAttribute (cycle 3), TestLoad_FileNotFound (cycle 4), TestLoad_NoTBXElement (cycle 5). Also strengthened TestLoad_UnsupportedDialect with errors.Is assertion. All tests pass, make build clean.
