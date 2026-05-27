---
id: ter-2tcq
status: closed
deps: [ter-ab56]
links: []
created: 2026-05-25T19:37:19Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, lookup, output]
---
# E4.T5 — Lookup envelope type

## Goal

Define the `LookupEnvelope` and `LookupResult` types in `internal/output/types.go` with the JSON shape specified in the E4 spec.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"lookup — match policy", lean output shape

## Files to create / modify

- `src/internal/output/types.go` — add `LookupEnvelope`, `LookupResult`, `LookupLanguages`, `LookupTermGroup`
- `src/internal/output/types_test.go` — JSON shape tests

## Behavior

Envelope shape per spec:

```json
{
  "schema_version": 1,
  "ok": true,
  "results": [{
    "concept_id": "tzimtzum",
    "subject_field": "kabbalah",
    "languages": {
      "he": {"preferred": {"term": "צמצום"}},
      "es": {"preferred": {"term": "tzimtzum"}}
    }
  }]
}
```

Go types:

```go
type LookupEnvelope struct {
    SchemaVersion int            `json:"schema_version"`
    OK            bool           `json:"ok"`
    Results       []LookupResult `json:"results"`
}

type LookupResult struct {
    ConceptID    string                       `json:"concept_id"`
    SubjectField string                       `json:"subject_field,omitempty"`
    Languages    map[string]LookupTermGroup   `json:"languages"`
}

type LookupTermGroup struct {
    Preferred *LookupTerm `json:"preferred,omitempty"`
    Admitted  []LookupTerm `json:"admitted,omitempty"`
}

type LookupTerm struct {
    Term string `json:"term"`
}
```

Not found → `results: []` (empty array, never null).

## TDD cycles

### Cycle 1 — JSON shape matches spec
RED: Marshal a `LookupEnvelope` with one result. Assert JSON keys match spec exactly (`concept_id`, `subject_field`, `languages`).
GREEN: Define the types with correct `json` tags.

### Cycle 2 — Empty results serializes as empty array
RED: `LookupEnvelope{Results: []LookupResult{}}` → JSON has `"results":[]`, not `null`.
GREEN: Initialize with empty slice.

### Cycle 3 — omitempty on optional fields
RED: `LookupResult` with empty `SubjectField` — assert `subject_field` is absent from JSON.
GREEN: Correct `omitempty` tag.

## Acceptance

- `make build` passes
- JSON shape matches E4 spec exactly
- Empty results → `[]` not `null`

## Notes

**2026-05-25T20:00:49Z**

Implemented LookupEnvelope, LookupResult, LookupTermGroup, and LookupTerm types in internal/output/types.go. Added MarshalJSON on LookupEnvelope to guarantee results serializes as [] (not null) when nil. Tests cover: JSON shape matching spec, empty/nil results as array, omitempty on subject_field, omitempty on admitted, admitted present with values. All tests pass, make build clean.
