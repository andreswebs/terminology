---
id: ter-lobd
status: closed
deps: [ter-ab56]
links: []
created: 2026-05-25T19:37:19Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, lookup, matching]
---
# E4.T6 — Lookup: case-fold + NFC matching

## Goal

Implement the core lookup matching logic as a method on `tbx.Glossary`. Given a query term and optional language filter, return matching concepts. Matching uses Unicode default case-fold and NFC normalization on both the query and glossary terms.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"lookup — match policy"
- `golang.org/x/text` — `cases.Fold()` for Unicode case-folding, `norm.NFC` for normalization

## Files to create / modify

- `src/internal/tbx/lookup.go` — `Glossary.Lookup(term, lang string) []LookupMatch`
- `src/internal/tbx/lookup_test.go` — tests

## Behavior

```go
type LookupMatch struct {
    Concept  Concept
    TermLang string
    TermSurface string
}

func (g *Glossary) Lookup(term, lang string) []LookupMatch
```

### Matching rules

1. **NFC normalize** both query and each `Term.Surface` in the glossary.
2. **Case-fold** both using `cases.Fold()` from `golang.org/x/text/cases`.
3. **Exact match** — no fuzzy, no lemmatization, no suggestions.
4. If `lang` is non-empty, restrict to that language section only.
5. Return all matching concepts (a term may appear in multiple concepts).

### Edge cases

- Query `"Tzimtzum"` matches term `"tzimtzum"` (case-fold).
- Query `"café"` (NFC) matches term `"café"` (NFD with combining accent).
- Empty query → no matches.
- No match → empty slice (not nil).

## TDD cycles

### Cycle 1 — Exact match (same case)
RED: Glossary with one concept containing term `"tzimtzum"`. `Lookup("tzimtzum", "")` returns 1 match.
GREEN: Implement linear scan with NFC + fold comparison.

### Cycle 2 — Case-insensitive match
RED: `Lookup("Tzimtzum", "")` matches `"tzimtzum"`.
GREEN: Apply `cases.Fold()` before comparison.

### Cycle 3 — NFC normalization
RED: Term stored as NFD (`"café"`), query as NFC (`"café"`). Match succeeds.
GREEN: Apply `norm.NFC.String()` before comparison.

### Cycle 4 — Language filter
RED: Concept has terms in `"he"` and `"en"`. `Lookup("tzimtzum", "he")` returns match; `Lookup("tzimtzum", "fr")` returns empty.
GREEN: Filter by `lang` key in `LangSection` map.

### Cycle 5 — No match returns empty slice
RED: `Lookup("nonexistent", "")` returns `[]LookupMatch{}` (not nil).
GREEN: Initialize result slice.

### Cycle 6 — Multiple concepts match
RED: Two concepts both contain term `"term"`. `Lookup("term", "")` returns 2 matches.
GREEN: Continue scanning all concepts.

## Acceptance

- `make build` passes
- Unicode case-fold via `cases.Fold()`
- NFC normalization via `norm.NFC`
- `--lang` filtering works
- No match → empty slice

## Notes

**2026-05-26T00:29:05Z**

Implemented Glossary.Lookup(term, lang string) []LookupMatch in src/internal/tbx/lookup.go. Uses golang.org/x/text/cases.Fold() for Unicode case-folding and golang.org/x/text/unicode/norm.NFC for normalization. Returns one match per concept (deduplicated across language sections). Empty query returns empty slice. All 10 tests pass, make build clean.
