---
id: ter-at09
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
# E3.T6 — Tier-3: invalid_lang_tag + missing_term

## Goal

Implement two tier-3 semantic checks in `Glossary.Validate()`:

1. **`invalid_lang_tag`** — BCP 47 well-formedness check on every `xml:lang` value. Uses `golang.org/x/text/language.Parse` for syntactic validation only (no IANA registry lookup, no canonicalization).
2. **`missing_term`** — every `<langSec>` must have at least one `<termSec>` with a `<term>`. A `LangSection` with an empty `Terms` slice triggers this warning.

Both checks are part of the `Validate()` method's concept iteration loop. They run alongside other tier-2/3 checks without short-circuiting.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Warning codes" (`invalid_lang_tag`, `missing_term`), §"BCP 47 validation"
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md) §"Sort orders summary" — `Languages` sorted ASCII byte order

## Files to modify

- `src/internal/tbx/validate.go` — add `invalid_lang_tag` and `missing_term` checks
- `src/internal/tbx/validate_test.go` — tests for both checks

## Implementation

Inside the `Validate()` concept loop, for each `(lang, langSection)` in `c.Languages`:

```go
_, err := language.Parse(lang)
if err != nil {
    res.Warnings = append(res.Warnings, Warning{
        Code:      "invalid_lang_tag",
        Message:   fmt.Sprintf("concept %q: malformed BCP 47 tag %q", c.ID, lang),
        ConceptID: c.ID,
    })
}

if len(ls.Terms) == 0 {
    res.Warnings = append(res.Warnings, Warning{
        Code:      "missing_term",
        Message:   fmt.Sprintf("concept %q lang %q: langSec has no term", c.ID, lang),
        ConceptID: c.ID,
    })
}
```

Additionally, the method collects all unique language tags across all concepts into a sorted `Languages` slice in the result (ASCII byte order via `sort.Strings`).

## TDD cycles

### Cycle 1 — Valid language tags, no warnings
RED: Glossary with concept having `Languages: {"en": ..., "he": ...}`. Assert no `invalid_lang_tag` warnings. Assert `res.Languages == ["en", "he"]` (sorted).
GREEN: Add language collection loop with `language.Parse` check.

### Cycle 2 — Malformed language tag
RED: Glossary with `Languages: {"not a lang!!!": ...}`. Assert 1 `invalid_lang_tag` warning with concept_id.
GREEN: Already passing from cycle 1.

### Cycle 3 — Missing term
RED: Glossary with `Languages: {"en": {Terms: []}}`. Assert 1 `missing_term` warning.
GREEN: Add empty terms check.

### Cycle 4 — Languages sorted across multiple concepts
RED: Glossary with 2 concepts: first has `{"he": ...}`, second has `{"en": ..., "es": ...}`. Assert `res.Languages == ["en", "es", "he"]` (merged + sorted).
GREEN: Use `map[string]struct{}` to collect unique langs, then sort.

## Deviation note

The current implementation in `validate.go` already implements both checks exactly as described. Language collection, sorting, and the `invalid_lang_tag` / `missing_term` warning emission match the spec. No changes needed.

## Out of scope

- IANA registry semantic validation (spec explicitly excludes this)
- Canonicalization of language tags (write-side concern, E7)
- Line/col tracking (T11)

## Acceptance

- `make build` passes
- Malformed BCP 47 tags produce `invalid_lang_tag` warnings via `x/text/language.Parse`
- Empty `Terms` slices produce `missing_term` warnings
- `Languages` in result is sorted ASCII byte order
- Languages are collected across all concepts, deduplicated

