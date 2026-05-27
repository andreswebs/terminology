---
id: ter-cplu
status: closed
deps: [ter-6z5g]
links: []
created: 2026-05-24T00:22:37Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, lineindex, foundation]
---
# E2.T5 — Line index for reader position tracking

## Goal

Implement a streaming line-index package that maps byte offsets to (line, column) positions. This is consumed by the TBX reader to populate Warning.Line and Warning.Col fields, giving agents and users precise source locations for reader diagnostics.

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Architecture" — `internal/tbx/lineindex — streaming newline-index for line/col tracking`
- E2 spec: §"Reader warnings" — `Warning` has `Line, Col int` fields
- Testing ADR: [docs/adr/testing.md](docs/adr/testing.md) §"Matcher" — line/col tracking is also used by E5 matcher

## Files to create

- `src/internal/tbx/lineindex/lineindex.go`
- `src/internal/tbx/lineindex/lineindex_test.go`

## API

```go
package lineindex

import "io"

type Index struct {
    offsets []int
}

func New(r io.Reader) (*Index, error)

func (idx *Index) Position(offset int) (line, col int)
```

Design notes:
- **`New(r io.Reader)`** reads the entire content once, building a slice of newline byte offsets. This is done before XML parsing begins (the reader can call `New` on the raw bytes, then seek back for XML decoding, or buffer the content).
- **`Position(offset int)`** uses binary search over the offset slice to return 1-based (line, col) for a given byte offset.
- The index is **read-only after construction** — safe for concurrent access without locks.
- Column is **byte offset within the line**, not rune offset. This matches encoding/xml's offset reporting and avoids the cost of UTF-8 decoding in the index.
- For the E2 reader integration: `encoding/xml.Decoder` exposes `InputOffset() int64` which gives the byte offset of the current token. The reader can capture this before decoding each element and pass it to `Position()` to populate Warning fields.

## Deviation note

**This package does not exist in the current codebase.** The existing implementation has Warning.Line and Warning.Col fields on the struct but they are never populated by the reader — all warnings have Line=0, Col=0. This ticket implements the missing piece. The reader integration (wiring InputOffset into Position calls) happens in T6 when the DCT reader is built.

## TDD cycles

### Cycle 1 — Empty input
RED: New(strings.NewReader("")) returns an index. Position(0) returns (1, 1).
GREEN: Handle empty input edge case.

### Cycle 2 — Single line
RED: New for "hello" (no newline). Position(0)=(1,1), Position(4)=(1,5).
GREEN: Build offset slice from reader.

### Cycle 3 — Multiple lines
RED: New for "line1\nline2\nline3". Position(0)=(1,1), Position(5)=(1,6), Position(6)=(2,1), Position(11)=(2,6), Position(12)=(3,1).
GREEN: Implement binary search in Position.

### Cycle 4 — Offset past end
RED: Position(999) for a short input returns the last valid line with clamped column.
GREEN: Handle bounds checking.

### Cycle 5 — Large input
RED: Build index from a 10,000-line input. Verify Position at various offsets.
GREEN: Confirm performance is acceptable (linear build, log-n lookup).

## Out of scope

- XML parsing (T6)
- Reader integration (T6 wires InputOffset → Position)
- Matcher line/col tracking (E5)

## Acceptance

- `make build` passes
- New() builds index from any io.Reader
- Position() returns correct 1-based (line, col) for all byte offsets
- Handles edge cases: empty input, offset at newline, offset past end


## Notes

**2026-05-25T13:14:26Z**

Implemented lineindex package at src/internal/tbx/lineindex/. New(io.Reader) builds a slice of line-start byte offsets in a single pass. Position(offset) uses sort.SearchInts (binary search) to return 1-based (line, col). Column is byte offset within the line, matching encoding/xml's InputOffset reporting. Handles edge cases: empty input, negative offset, offset past end, trailing newline, CRLF line endings. 8 test functions covering all TDD cycles from the ticket.
