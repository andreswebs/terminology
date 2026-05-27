---
id: ter-8zuv
status: closed
deps: [ter-bedf]
links: []
created: 2026-05-24T01:05:15Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, validation, tier3]
---
# E3.T5 — Tier-3: duplicate concept ID detection

## Goal

Implement `duplicate_id` detection in `Glossary.Validate()`. When two or more `<conceptEntry>` elements share the same `id` attribute, emit a warning with code `"duplicate_id"`. The `concepts` count in the envelope reports the **as-found raw count** (not deduplicated), so consumers can compute the collision from the warning list.

This is a tier-3 (semantic) check. It runs after tier-1 succeeds and alongside other tier-2/3 checks without short-circuiting.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Output" (as-found counts), §"Warning codes" (`duplicate_id`)
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md) §"Sort orders summary"

## Files to create / modify

- `src/internal/tbx/validate.go` — `Glossary.Validate()` method, `ValidateResult` struct
- `src/internal/tbx/validate_test.go` — tests for duplicate_id detection

## API

```go
package tbx

type ValidateResult struct {
    Concepts  int
    Languages []string
    Warnings  []Warning
    Errors    []Warning
}

func (g *Glossary) Validate(strict bool) ValidateResult
```

The method iterates over `g.Concepts`, tracking seen IDs in a `map[string]int` (value = index of first occurrence). When a duplicate is found:

```go
res.Warnings = append(res.Warnings, Warning{
    Code:      "duplicate_id",
    Message:   fmt.Sprintf("concept ID %q already declared at index %d", c.ID, prev),
    ConceptID: c.ID,
})
```

`res.Concepts = len(g.Concepts)` — always the raw count, not deduplicated.

## TDD cycles

### Cycle 1 — No duplicates, clean result
RED: Create a Glossary with 1 concept. Call `Validate(false)`. Assert `Concepts == 1`, `len(Warnings) == 0`.
GREEN: Implement `Validate()` skeleton that counts concepts and returns.

### Cycle 2 — Duplicate ID produces warning
RED: Create a Glossary with 2 concepts sharing the same ID `"c1"`. Call `Validate(false)`. Assert `Concepts == 2` (raw count), `len(Warnings) == 1`, warning code is `"duplicate_id"`, warning ConceptID is `"c1"`.
GREEN: Add ID tracking map and duplicate detection.

### Cycle 3 — Multiple duplicates
RED: Glossary with 3 concepts: IDs `"a"`, `"b"`, `"a"`. Assert exactly 1 `duplicate_id` warning for `"a"`, no warning for `"b"`.
GREEN: Already passing from cycle 2.

## Deviation note

The current implementation in `validate.go` already implements duplicate_id detection exactly as described above. The `Validate(strict bool)` method exists with the `ValidateResult` struct (Concepts int, Languages []string, Warnings/Errors []Warning). No changes needed to align with spec.

## Out of scope

- Language collection and sorting (T6)
- Cross-reference validation (T7)
- Strict mode promotions (T10)
- Line/col tracking in warnings (T11)

## Acceptance

- `make build` passes
- `Glossary.Validate()` returns `ValidateResult` with raw concept count
- Duplicate IDs produce `duplicate_id` warnings with concept_id
- No deduplication of concept count — reports file shape as-is

