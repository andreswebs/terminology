---
id: ter-bmce
status: closed
deps: [ter-st7u]
links: []
created: 2026-05-26T19:30:14Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, write, id]
---
# E7.T4 — Concept-ID derivation (internal/write/id.go)

Implement concept-ID derivation per cli-design.md §Concept IDs. Function: DeriveID(term string) (string, error). Steps: NFKD normalize, drop combining marks, Unicode casefold, replace non-[a-z0-9] runs with single hyphen, trim hyphens, truncate to 64 codepoints on hyphen boundary. Returns invalid_id error if result is empty (e.g. all non-Latin with no romanization). Canonical-lang resolution: --canonical-lang flag > TERMINOLOGY_CANONICAL_LANG env > 'en' > first langSec. Property: DeriveID is pure and idempotent (DeriveID(DeriveID(s)) == DeriveID(s) when result is valid).

## Acceptance Criteria

- make build passes
- DeriveID exported from internal/write
- Handles Hebrew, Spanish, Latin, mixed-script input
- Truncation on hyphen boundary at 64 codepoints
- Empty result returns invalid_id error
- Property test: idempotence
- Table-driven tests for edge cases (accented chars, CJK, Hebrew-only, already-slugified)


## Notes

**2026-05-26T19:40:34Z**

Implemented DeriveID in internal/write/id.go with ErrInvalidID sentinel in internal/write/errors.go. Algorithm: NFKD normalize → drop combining marks (unicode.Mn) → Unicode casefold (cases.Fold) → replace non-[a-z0-9] runs with single hyphen → trim hyphens → truncate to 64 codepoints on hyphen boundary. Tests cover: basic Latin, accented chars, non-alnum runs, already-slugified, Hebrew-only (invalid_id), empty string, mixed script, CJK, numbers, leading/trailing special chars, eszett folding, truncation (3 variants), exactly-64 passthrough, error sentinel fields, and idempotence property. The write package is not yet imported transitively by app, so the invalid_id sentinel won't appear in schema output until a command action imports write (expected in T6/T7).
