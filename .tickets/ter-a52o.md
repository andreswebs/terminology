---
id: ter-a52o
status: closed
deps: [ter-bedf]
links: []
created: 2026-05-24T01:05:16Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, validation, tier3]
---
# E3.T7 — Tier-3: unresolved cross-reference detection

## Goal

Implement `unresolved_crossref` detection in `Glossary.Validate()`. Cross-references (`CrossRef.Target`) at both concept level and term level must resolve to a known concept ID. When a target ID doesn't exist in the glossary, the check emits a warning.

In **lenient mode** (default): unresolved cross-refs are warnings.
In **strict mode**: unresolved cross-refs are promoted to errors (handled in T10).

This ticket implements the lenient-mode behavior only. The strict promotion is wired in T10.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Warning codes" (`unresolved_crossref`), §"--strict"
- Domain model: `Concept.CrossRefs []CrossRef` and `Term.CrossRefs []CrossRef`

## Files to modify

- `src/internal/tbx/validate.go` — add cross-reference resolution checks
- `src/internal/tbx/validate_test.go` — tests for resolved + unresolved refs

## Implementation

After tracking seen IDs (from T5's duplicate detection), check cross-refs:

```go
// Concept-level cross-refs
for _, cr := range c.CrossRefs {
    if _, found := idSeen[cr.Target]; !found {
        w := Warning{
            Code:      "unresolved_crossref",
            Message:   fmt.Sprintf("concept %q references unknown ID %q", c.ID, cr.Target),
            ConceptID: c.ID,
        }
        if strict {
            res.Errors = append(res.Errors, w)
        } else {
            res.Warnings = append(res.Warnings, w)
        }
    }
}

// Term-level cross-refs
for _, ls := range c.Languages {
    for _, t := range ls.Terms {
        for _, cr := range t.CrossRefs {
            if _, found := idSeen[cr.Target]; !found {
                w := Warning{
                    Code:      "unresolved_crossref",
                    Message:   fmt.Sprintf("concept %q term %q references unknown ID %q", c.ID, t.Surface, cr.Target),
                    ConceptID: c.ID,
                }
                if strict {
                    res.Errors = append(res.Errors, w)
                } else {
                    res.Warnings = append(res.Warnings, w)
                }
            }
        }
    }
}
```

Note: cross-ref resolution is order-dependent — a concept can only reference IDs that appear **before** it in the `Concepts` slice (since `idSeen` is built as we iterate). Forward references are detected as unresolved. This is by design: the reader preserves input order, and cross-refs are validated against what's been seen so far.

## TDD cycles

### Cycle 1 — Resolved cross-ref, no warning
RED: Glossary with 2 concepts. Second concept has `CrossRefs: [{Target: firstConceptID}]`. Assert no `unresolved_crossref` warnings.
GREEN: Add cross-ref resolution logic.

### Cycle 2 — Unresolved concept-level cross-ref (lenient)
RED: Concept with `CrossRefs: [{Target: "nonexistent"}]`. `Validate(false)`. Assert 1 `unresolved_crossref` warning (not error).
GREEN: Already passing from cycle 1.

### Cycle 3 — Unresolved term-level cross-ref (lenient)
RED: Term with `CrossRefs: [{Target: "nonexistent"}]`. `Validate(false)`. Assert 1 `unresolved_crossref` warning.
GREEN: Add term-level cross-ref checking.

### Cycle 4 — Unresolved cross-ref in strict mode
RED: `Validate(true)` with unresolved cross-ref. Assert the issue appears in `res.Errors` (not `res.Warnings`).
GREEN: Add strict-mode conditional routing.

## Deviation note

The current implementation in `validate.go` already handles both concept-level and term-level cross-reference resolution with correct lenient/strict routing. The behavior matches the spec exactly. No changes needed.

## Out of scope

- The `--strict` flag wiring in the validate command (T10/T12)
- Line/col tracking (T11)
- Forward-reference resolution (by design, cross-refs are validated in order)

## Acceptance

- `make build` passes
- Concept-level and term-level cross-refs are checked
- Unresolved refs → warning in lenient mode, error in strict mode
- Warning carries `concept_id` and target ID in the message
- Resolved refs produce no warnings

