---
id: ter-by7v
status: closed
deps: [ter-ab56]
links: []
created: 2026-05-25T19:37:19Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, markdown, foundation]
---
# E4.T3 — Markdown spans package (goldmark)

## Goal

Create `internal/markdown` package with a `Spans` function that walks markdown source and yields plain-text spans (with byte offsets) for every node that is not a code block, inline code, or HTML block. Uses `github.com/yuin/goldmark` as the parser. This package is shared with E6 (`scan`, `check`).

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"Markdown awareness"
- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md) — reuses this package

## Files to create / modify

- `src/internal/markdown/text.go` — `Spans` function
- `src/internal/markdown/text_test.go` — tests
- `src/go.mod` — add `github.com/yuin/goldmark` dependency

## Behavior

```go
package markdown

import "iter"

type Span struct {
    Text   string
    Offset int
    Line   int
    Col    int
}

func Spans(src []byte) iter.Seq[Span]
```

`Spans` parses `src` as CommonMark, walks the AST, and yields `Span` values for every text node that is not inside:
- Fenced code blocks (` ``` `)
- Inline code (`` ` ``)
- HTML blocks

Each `Span` carries:
- `Text` — the plain text content
- `Offset` — byte offset in the original `src`
- `Line`, `Col` — 1-based line/column in the original source

Line/col preservation is critical for E6 (`scan`/`check`) which reports match locations in the original markdown's coordinates.

## TDD cycles

### Cycle 1 — Plain text yields single span
RED: `Spans([]byte("hello world"))` yields one span with Text="hello world", Offset=0, Line=1.
GREEN: Implement basic goldmark parse + AST walk.

### Cycle 2 — Fenced code blocks excluded
RED: Input with a fenced code block — assert no span contains code block content.
GREEN: Skip `ast.FencedCodeBlock` nodes in walk.

### Cycle 3 — Inline code excluded
RED: Input `"use the \`getUserById\` function"` — spans should not contain `getUserById`.
GREEN: Skip `ast.CodeSpan` nodes.

### Cycle 4 — Multiple paragraphs yield multiple spans
RED: Two-paragraph input — yields spans from both paragraphs with correct offsets.
GREEN: Walk continues across block boundaries.

### Cycle 5 — Line/col accuracy
RED: Multi-line input — assert spans carry correct 1-based line and column numbers.
GREEN: Use goldmark's `text.Segment` byte offsets + pre-computed line index over `src`.

### Cycle 6 — HTML blocks excluded
RED: Input with `<div>...</div>` — no span contains HTML content.
GREEN: Skip `ast.HTMLBlock` nodes.

## Acceptance

- `make build` passes
- `goldmark` added to `go.mod`
- Code blocks, inline code, and HTML blocks are excluded from spans
- Spans carry correct byte offsets and line/col positions
- Iterator-based API using `iter.Seq[Span]`

## Notes

**2026-05-26T00:25:14Z**

Implemented internal/markdown package with Spans(src []byte) iter.Seq[Span] function using goldmark. The function parses CommonMark and yields plain-text spans with byte offsets and 1-based line/col positions, excluding fenced code blocks, indented code blocks, inline code, and HTML blocks. Uses recursive AST walk with goldmark's text.Segment for offset preservation. 10 tests covering all exclusion types, emphasis/link preservation, heading inclusion, multi-paragraph offsets, line/col accuracy, and empty input. goldmark v1.8.2 added to go.mod.
