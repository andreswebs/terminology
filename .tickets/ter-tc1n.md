---
id: ter-tc1n
status: closed
deps: [ter-c4ra]
links: []
created: 2026-05-26T13:49:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, task, match, ahocorasick]
---
# E5.T3 — Aho-Corasick dependency + automaton build

## Goal

Add the `github.com/cloudflare/ahocorasick` dependency and implement the automaton-building layer. Given a set of canonical pattern byte strings, build an Aho-Corasick automaton that can scan canonical text and return raw match positions + pattern IDs.

## Refs

- E5 spec: [docs/specs/005-matcher.md](docs/specs/005-matcher.md) §"Architecture"
- `github.com/cloudflare/ahocorasick` — BSD-3, byte-oriented multi-pattern matching

## Files to create / modify

- `src/go.mod` / `src/go.sum` — add `github.com/cloudflare/ahocorasick`
- `src/internal/match/automaton.go` — automaton builder, `rawMatch` type
- `src/internal/match/automaton_test.go` — tests

## Behavior

```go
type rawMatch struct {
    PatternID int
    Start     int // byte offset in canonical text
    End       int // byte offset (exclusive) in canonical text
}

type automaton struct {
    machine  *ahocorasick.Matcher
    patterns [][]byte
}

func buildAutomaton(patterns [][]byte) *automaton
func (a *automaton) Search(canonical []byte) []rawMatch
```

### Pattern building

Each pattern is a canonical byte string (already normalized via T2). The automaton is constructed once and reused across multiple `Search` calls.

### Raw match output

`Search` returns all substring matches found by the AC automaton. These are raw — no boundary checking, no deduplication. Those are handled by T4 and T5.

## TDD cycles

### Cycle 1 — Single pattern match
RED: Build automaton with `["hello"]`, search `"say hello world"` → one rawMatch at correct offset.
GREEN: Wrap `ahocorasick.NewMatcher` and `Search`.

### Cycle 2 — Multiple pattern match
RED: Build with `["hello", "world"]`, search `"hello world"` → two rawMatches.
GREEN: Map pattern IDs back from AC results.

### Cycle 3 — Overlapping patterns
RED: Build with `["abc", "abcdef"]`, search `"xabcdefy"` → both matches reported (filtering is T5's job).
GREEN: AC reports all substring matches.

### Cycle 4 — No matches
RED: Build with `["xyz"]`, search `"hello world"` → empty slice.
GREEN: Return nil/empty.

### Cycle 5 — Empty patterns list
RED: Build with no patterns → Search returns empty on any input.
GREEN: Handle empty automaton.

## Acceptance

- `make build` passes
- `go get github.com/cloudflare/ahocorasick` succeeds, go.mod updated
- Automaton builds from byte patterns
- `Search` returns all substring matches with correct pattern IDs and offsets
- Reusable across multiple Search calls


## Notes

**2026-05-26T14:22:05Z**

Added github.com/cloudflare/ahocorasick dependency. Implemented automaton.go with buildAutomaton() and Search() in internal/match. The cloudflare AC library's Match() only returns which pattern IDs matched (not positions), so Search() uses AC for pattern detection then bytes.Index loops for position finding. Results are sorted by start position (ascending), with longest match first at same start. 8 tests cover: single pattern, multiple patterns, overlapping patterns, no matches, empty patterns, multiple occurrences of same pattern, reuse across searches, and deterministic sort order.
