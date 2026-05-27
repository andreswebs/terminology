---
id: ter-c7vk
status: closed
deps: [ter-c4ra]
links: []
created: 2026-05-26T13:49:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, task, output, scan]
---
# E5.T7 — Scan envelope type

## Goal

Define the `ScanEnvelope` output type for the `scan` command and register it in the output type registry. The envelope shape matches the spec's scan output format.

## Refs

- E5 spec: [docs/specs/005-matcher.md](docs/specs/005-matcher.md) §"API" → Match type
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"terminology scan"
- Schema-source-of-truth ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md)

## Files to create / modify

- `src/internal/output/types.go` — add `ScanEnvelope`, `ScanMatch`, `ScanSummary` types; register envelope
- `src/internal/output/types_test.go` — tests (if needed)

## Behavior

```go
type ScanEnvelope struct {
    SchemaVersion int          `json:"schema_version"`
    OK            bool         `json:"ok"`
    File          string       `json:"file"`
    Matches       []ScanMatch  `json:"matches"`
    Summary       ScanSummary  `json:"summary"`
}

type ScanMatch struct {
    ConceptID string `json:"concept_id"`
    Term      string `json:"term"`
    Lang      string `json:"lang"`
    Status    string `json:"status"`
    Line      int    `json:"line"`
    Column    int    `json:"column"`
    Context   string `json:"context"`
}

type ScanSummary struct {
    TotalMatches   int `json:"total_matches"`
    UniqueConcepts int `json:"unique_concepts"`
}
```

### Registration

```go
func init() {
    RegisterEnvelope("scan", ScanEnvelope{})
}
```

### MarshalJSON

Like `LookupEnvelope` and `ExtractEnvelope`, `ScanEnvelope` needs a custom `MarshalJSON` to ensure `matches` serializes as `[]` (not `null`) when empty.

## TDD cycles

### Cycle 1 — Envelope registration
RED: `EnvelopeFor("scan")` returns a `ScanEnvelope` zero value.
GREEN: Add init registration.

### Cycle 2 — JSON serialization
RED: Marshal a `ScanEnvelope` with matches → JSON has correct keys in spec order.
GREEN: Struct with JSON tags.

### Cycle 3 — Empty matches serialize as array
RED: Marshal `ScanEnvelope{Matches: nil}` → `"matches": []` (not null).
GREEN: Custom `MarshalJSON`.

## Acceptance

- `make build` passes
- `ScanEnvelope` registered in output type registry
- JSON output matches spec shape (file, matches, summary)
- Empty matches serialize as `[]`
- Exit codes for scan already registered: `[0, 2, 3, 65]`


## Notes

**2026-05-26T14:34:40Z**

Implemented ScanEnvelope, ScanMatch, and ScanSummary types in internal/output/types.go. Registered scan envelope in init(). Added MarshalJSON nil→[] coercion following the same alias pattern as LookupEnvelope and ExtractEnvelope. Added 4 tests: registration, JSON shape, nil matches → array, empty matches → array. Updated schema/full golden file to include the new scan envelope fields.
