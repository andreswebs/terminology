---
id: ter-y57h
status: closed
deps: [ter-s7xa, ter-qxpp, ter-ir1k]
links: []
created: 2026-05-26T17:24:26Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-7fyo
tags: [e6, task, check, algorithm]
---
# E6.T4 — Check algorithm: missing + forbidden_variant

## Goal

Implement the core check verification logic. Given a source file, target file, glossary, and language pair, produce a list of violations (`missing` and `forbidden_variant`). This is the algorithmic core that T7 wires into the command surface.

## Refs

- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md) §"Algorithm", §"Counting policy"
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"terminology check SRC TGT" — behavior section
- Matcher API: `src/internal/match/match.go` — `New`, `Scan`
- Markdown spans: `src/internal/markdown/text.go` — `Spans`
- TBX model: `src/internal/tbx/model.go` — `Concept`, `LangSection`, `Term`, `Status`

## Files to create / modify

- `src/internal/check/check.go` — new package with `Check` function
- `src/internal/check/check_test.go` — unit tests

## Algorithm

```go
package check

type Result struct {
    Violations      []output.CheckViolation
    Warnings        []output.CheckWarning
    ConceptsChecked int
}

func Check(g *tbx.Glossary, srcText, tgtText []byte,
    srcLang, tgtLang string, contextSize int, strict bool) (*Result, error)
```

### Steps

1. Build source matcher: `match.New(g, srcLang, match.PolicyFor(srcLang))`.
2. Parse SRC markdown spans; scan SRC for source-lang matches.
3. Group source matches by `concept_id`. Concepts with zero source matches are ignored.
4. Build target matcher: `match.New(g, tgtLang, match.PolicyFor(tgtLang))`.
5. Parse TGT markdown spans; scan TGT for target-lang matches.
6. For each concept with source occurrences:
   a. Find the preferred target term for this concept in `tgtLang` (from glossary).
   b. Count preferred target matches in TGT scan results. Zero → `missing` violation.
   c. Find deprecated/superseded target terms for this concept. Each TGT match → `forbidden_variant` violation.
   d. (Admitted variants handled in T5.)
7. Each violation carries enough context for one-pass fixes: `concept_id`, `expected_target`, `source_term`, `source_occurrences`, and for positional violations: `line`, `column`, `context`.

### Counting policy

"At least one" semantics:
- Source ≥1, target ≥1 → OK
- Source ≥1, target 0 → `missing`
- Source 0 → concept ignored entirely

### Code region skipping

Both SRC and TGT go through `markdown.Spans()` which strips code blocks. Symmetric application per spec §"Code regions".

## TDD cycles

### Cycle 1 — No violations (clean check)
RED: SRC has "tzimtzum", TGT has "צמצום" (the preferred target) → zero violations.
GREEN: Implement basic pipeline.

### Cycle 2 — Missing violation
RED: SRC has "tzimtzum" 3 times, TGT has zero target matches → one `missing` violation with `source_occurrences: 3`.
GREEN: Detect absent preferred term.

### Cycle 3 — Forbidden variant
RED: TGT contains deprecated variant "contraction" at line 5, col 10 → `forbidden_variant` violation with position.
GREEN: Scan TGT for deprecated terms, emit violations.

### Cycle 4 — Concept absent from SRC
RED: Glossary has concept X, SRC doesn't mention it → no violations for concept X.
GREEN: Skip concepts with zero source matches.

### Cycle 5 — Multiple concepts
RED: SRC has 3 concepts; 1 missing, 1 with forbidden variant, 1 clean → exactly 2 violations.
GREEN: Iterate all concepts.

### Cycle 6 — Code blocks excluded
RED: Preferred target term appears only in a fenced code block in TGT → `missing` violation (code region skipped).
GREEN: Use `markdown.Spans()` for both files.

### Cycle 7 — Context window
RED: Forbidden variant violation carries context string with surrounding text.
GREEN: Pass `contextSize` through to matcher.

## Acceptance

- `make build` passes
- `check.Check` returns correct violations for missing + forbidden_variant cases
- "At least one" counting policy: source count doesn't matter beyond ≥1
- Code blocks excluded symmetrically from SRC and TGT
- Concepts absent from SRC are ignored
- Each violation carries concept_id, type, and position/context where applicable

## Notes

**2026-05-26T18:19:20Z**

Implemented internal/check/check.go with Check() function and 11 unit tests covering: no violations, missing violation, forbidden_variant, concept absent from source, multiple concepts, code blocks excluded, context window, unspecified status treated as preferred, superseded variants, no target language in glossary, admitted source terms triggering check. The algorithm: (1) builds source matcher, scans SRC for preferred+admitted terms, (2) groups by concept_id, (3) builds target matcher, scans TGT, (4) emits missing violations when preferred target has 0 occurrences, (5) emits forbidden_variant for each deprecated/superseded target match. Violations sorted per spec: positional (line,col) first, missing at end grouped by concept_id. make build passes.
