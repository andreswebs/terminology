---
id: ter-vg07
status: closed
deps: [ter-y57h]
links: []
created: 2026-05-26T17:24:34Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-7fyo
tags: [e6, task, check, ordering]
---
# E6.T6 — Violation ordering

## Goal

Implement the violation and warning sort order specified by the E6 spec. Violations in the check output must be deterministically ordered for golden-file stability and for the agent's "fix top-to-bottom" workflow.

## Refs

- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md) §"Violation ordering"
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md) §"Sort orders summary"

## Files to create / modify

- `src/internal/check/check.go` — add sort step after violation collection
- `src/internal/check/check_test.go` — ordering-specific tests

## Sort rules

### Violations

1. **Primary**: by `(line, column)` in TGT — positional violations appear in reading order.
2. **`missing` violations**: have no line/column (the concept is absent). These sort to the **end** of the list as a tail group.
3. **Within the missing tail group**: ordered by `concept_id` ASCII byte sort.
4. **Tiebreak for same (line, column)**: `concept_id` ASCII byte sort.

### Warnings

Same rules as violations: positional warnings by (line, column), non-positional to tail, within groups by concept_id.

## TDD cycles

### Cycle 1 — Positional violations sorted
RED: Two `forbidden_variant` violations at (10, 5) and (3, 12) → output order is (3, 12), (10, 5).
GREEN: Sort by (line, column).

### Cycle 2 — Missing violations at tail
RED: One `missing` violation + one `forbidden_variant` at (5, 1) → forbidden first, missing last.
GREEN: Missing violations sort after all positional violations.

### Cycle 3 — Missing tail group sorted by concept_id
RED: Two `missing` violations for concepts "tzimtzum" and "sefirah" → "sefirah" before "tzimtzum".
GREEN: ASCII sort within tail group.

### Cycle 4 — Same position tiebreak
RED: Two violations at same (line, column) for different concepts → sorted by concept_id.
GREEN: Tiebreak on concept_id.

## Acceptance

- `make build` passes
- Positional violations sorted by (line, column)
- Missing violations always at end
- Missing tail group sorted by concept_id ASCII
- Deterministic output across runs (golden-file stable)

## Notes

**2026-05-26T18:24:49Z**

Implemented violation and warning sort ordering in internal/check/check.go. Changes: (1) Added concept_id tiebreak to sortViolations for same (line, column) positions. (2) Added sortWarnings function with same rules: positional warnings by (line, column), non-positional to tail, concept_id tiebreak and tail group sort. (3) Called sortWarnings in Check before returning. (4) Added 7 unit tests covering all four TDD cycles for both violations and warnings: positional sorting, missing/non-positional at tail, tail group by concept_id, same-position tiebreak.
