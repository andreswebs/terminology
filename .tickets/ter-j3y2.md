---
id: ter-j3y2
status: closed
deps: [ter-c4ra]
links: []
created: 2026-05-26T13:49:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, task, match, filter]
---
# E5.T5 — Longest-match-at-same-start filter

## Goal

Implement the post-filter that, when multiple matches share the same start position, keeps only the longest one. This is a standard tokenizer convention — the longer phrase is more specific (e.g., `"tzimtzum primordial"` wins over `"tzimtzum"` when both start at the same offset).

## Refs

- E5 spec: [docs/specs/005-matcher.md](docs/specs/005-matcher.md) §"Longest-match-at-same-start"

## Files to create / modify

- `src/internal/match/filter.go` — `longestMatchPerStart` function
- `src/internal/match/filter_test.go` — tests

## Behavior

```go
func longestMatchPerStart(matches []rawMatch) []rawMatch
```

1. Group all raw matches by their `Start` position.
2. Within each group, keep only the match with the largest `End - Start` span.
3. Return the filtered list, preserving the original order of surviving matches by start position.

## TDD cycles

### Cycle 1 — No overlaps
RED: Input `[{P:0, S:0, E:5}, {P:1, S:10, E:15}]` → both survive.
GREEN: Different start positions, no filtering needed.

### Cycle 2 — Overlapping at same start
RED: Input `[{P:0, S:0, E:8}, {P:1, S:0, E:20}]` → only the longer one (E:20) survives.
GREEN: Group by start, keep longest.

### Cycle 3 — Multiple groups
RED: Three matches — two at start 0 (lengths 5 and 10), one at start 20 (length 5) → two survivors.
GREEN: Per-group filtering.

### Cycle 4 — Empty input
RED: Empty slice → empty result.
GREEN: Handle nil/empty.

### Cycle 5 — Single match
RED: One match → returned unchanged.
GREEN: Trivial case.

## Acceptance

- `make build` passes
- Only the longest match per start position survives
- Order preserved by start position
- Works with empty and single-element inputs


## Notes

**2026-05-26T14:26:47Z**

Implemented longestMatchPerStart in src/internal/match/filter.go with tests in filter_test.go. The function exploits the pre-sorted order from automaton.Search() (start ascending, longest first) — it simply deduplicates by Start position, keeping the first (longest) match per group. Five tests cover: empty input, single match, no overlaps, same-start filtering, and multiple groups.
