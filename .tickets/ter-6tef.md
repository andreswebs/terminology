---
id: ter-6tef
status: closed
deps: [ter-c4ra, ter-2y2w]
links: []
created: 2026-05-26T13:49:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, task, match, boundary]
---
# E5.T4 — Word-boundary check

## Goal

Implement the post-filter that validates raw Aho-Corasick matches against word boundaries in the **original** (pre-normalization) text. The boundary contract is `(^|[^\p{L}\p{N}])TERM([^\p{L}\p{N}]|$)` — a match is valid only if the characters immediately before and after it in the original text are not letters or numbers.

## Refs

- E5 spec: [docs/specs/005-matcher.md](docs/specs/005-matcher.md) §"Boundary check (post-filter)"
- Unicode categories: `\p{L}` (Letter), `\p{N}` (Number)

## Files to create / modify

- `src/internal/match/boundary.go` — `validBoundary` function
- `src/internal/match/boundary_test.go` — tests

## Behavior

```go
func validBoundary(orig []byte, start, end int) bool
```

Checks whether the byte positions `start` and `end` (exclusive) in the original text constitute word boundaries:

1. If `start > 0`, the rune ending at `start-1` must **not** be `\p{L}` or `\p{N}`.
2. If `end < len(orig)`, the rune starting at `end` must **not** be `\p{L}` or `\p{N}`.
3. Start-of-text and end-of-text are valid boundaries.

### Why original text?

Boundary validation must happen against the **original** text, not the canonical form. Diacritics and niqqud are word characters — stripping them in canonical form would produce spurious boundary failures. The `Canonical.Map` from T2 translates AC match positions back to original-text positions for this check.

## TDD cycles

### Cycle 1 — Valid boundary (space-delimited)
RED: `validBoundary([]byte("the tzimtzum concept"), 4, 12)` → true.
GREEN: Check rune before start and after end.

### Cycle 2 — Invalid boundary (embedded in word)
RED: `validBoundary([]byte("pretzimtzumx"), 3, 11)` → false.
GREEN: `\p{L}` at boundary positions.

### Cycle 3 — Start of text
RED: `validBoundary([]byte("tzimtzum concept"), 0, 8)` → true.
GREEN: Start == 0 is a valid boundary.

### Cycle 4 — End of text
RED: `validBoundary([]byte("the tzimtzum"), 4, 12)` → true.
GREEN: End == len(orig) is a valid boundary.

### Cycle 5 — Punctuation boundary
RED: `validBoundary([]byte("(tzimtzum)"), 1, 9)` → true.
GREEN: Parentheses are not `\p{L}` or `\p{N}`.

### Cycle 6 — Hebrew adjacent to Spanish punctuation
RED: `validBoundary([]byte("el צמצום,"), 3, 9)` → true (comma is not \p{L}).
GREEN: Cross-script boundary works.

### Cycle 7 — Number boundary
RED: `validBoundary([]byte("3tzimtzum"), 1, 9)` → false (digit before).
GREEN: `\p{N}` is a word character.

## Acceptance

- `make build` passes
- Boundary check uses `\p{L}` and `\p{N}` as word characters
- Validates against original text, not canonical
- Handles start/end of text, punctuation, cross-script adjacency


## Notes

**2026-05-26T14:24:14Z**

Implemented validBoundary(orig []byte, start, end int) bool in src/internal/match/boundary.go. Uses utf8.DecodeLastRune to check the rune ending before start and utf8.DecodeRune to check the rune starting at end. Word characters are unicode.Letter and unicode.Number (matching the \p{L} and \p{N} contract). 9 table-driven tests covering: space-delimited, embedded-in-word, start/end of text, punctuation boundary, Hebrew adjacent to Spanish punctuation, number boundary, both boundaries invalid, and entire-text match.
