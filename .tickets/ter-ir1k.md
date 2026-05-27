---
id: ter-ir1k
status: closed
deps: [ter-ppn9]
links: []
created: 2026-05-26T17:24:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-7fyo
tags: [e6, task, output, types]
---
# E6.T3 — Check envelope + violation types

## Goal

Define the output envelope types for the `check` command in `internal/output/types.go` and register the envelope so `terminology schema` reflects it.

## Refs

- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md) §"Output"
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"terminology check SRC TGT"
- Schema ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md)
- Existing types: `src/internal/output/types.go` (ScanEnvelope as reference)

## Files to create / modify

- `src/internal/output/types.go` — add `CheckEnvelope`, `CheckViolation`, `CheckWarning`, `CheckSummary`
- `src/internal/output/registry.go` — register `CheckEnvelope` for the `check` command (exit codes already registered)

## Types

```go
type CheckEnvelope struct {
    SchemaVersion int              `json:"schema_version"`
    OK            bool             `json:"ok"`
    Source        string           `json:"source"`
    Target        string           `json:"target"`
    Violations    []CheckViolation `json:"violations"`
    Warnings      []CheckWarning   `json:"warnings"`
    Summary       CheckSummary     `json:"summary"`
}

type CheckViolation struct {
    Type              string `json:"type"`
    ConceptID         string `json:"concept_id"`
    SourceTerm        string `json:"source_term,omitempty"`
    ExpectedTarget    string `json:"expected_target,omitempty"`
    SourceOccurrences int    `json:"source_occurrences,omitempty"`
    Variant           string `json:"variant,omitempty"`
    Line              int    `json:"line,omitempty"`
    Column            int    `json:"column,omitempty"`
    Context           string `json:"context,omitempty"`
}

type CheckWarning struct {
    Type      string `json:"type"`
    ConceptID string `json:"concept_id"`
    Message   string `json:"message"`
}

type CheckSummary struct {
    Violations      int `json:"violations"`
    Warnings        int `json:"warnings"`
    ConceptsChecked int `json:"concepts_checked"`
}
```

### Violation types

- `"missing"` — concept in SRC but preferred target absent from TGT. Fields: `source_term`, `expected_target`, `source_occurrences`.
- `"forbidden_variant"` — deprecated/superseded variant in TGT. Fields: `variant`, `line`, `column`, `context`.
- `"admitted_variant"` — (`--strict` only) admitted variant in TGT. Fields: `variant`, `line`, `column`, `context`.

### Warning types

- `"admitted_variant"` — admitted variant in TGT when not `--strict`.

## TDD cycles

### Cycle 1 — Types compile
RED: Import `CheckEnvelope` in a test → doesn't exist.
GREEN: Add the types.

### Cycle 2 — JSON serialization shape
RED: Marshal a `CheckEnvelope` with one missing + one forbidden violation → verify JSON shape matches spec.
GREEN: JSON tags produce correct output with `omitempty` on type-specific fields.

### Cycle 3 — Registry
RED: `output.EnvelopeFor("check")` returns nil.
GREEN: Register in `registry.go`.

## Acceptance

- `make build` passes
- All check output types defined with correct JSON tags
- `omitempty` on violation fields that are type-specific
- Envelope registered for `check` command
- `terminology schema --command check` includes the envelope shape

## Notes

**2026-05-26T18:15:26Z**

Implemented CheckEnvelope, CheckViolation, CheckWarning, and CheckSummary types in internal/output/types.go. Registered the check envelope in init(). Added MarshalJSON nil→[] coercion for Violations and Warnings slices (same pattern as Lookup/Extract/Scan envelopes). Violation fields use omitempty so type-specific fields (source_term/expected_target/source_occurrences for missing; variant/line/column/context for forbidden_variant) are omitted when zero. Regenerated schema full golden test file since new envelope registration changes schema output. All tests pass, make build clean.
