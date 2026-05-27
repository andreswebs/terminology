---
id: ter-mf43
status: closed
deps: [ter-bedf]
links: []
created: 2026-05-24T01:05:11Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, output, foundation]
---
# E3.T2 — Validate envelope types (output/types.go)

## Goal

Define the Go struct types for the `terminology validate` JSON envelope in `internal/output/types.go`. Per the [schema-source-of-truth ADR](docs/adr/schema-source-of-truth.md), Go struct types in `internal/output/types.go` are the canonical contract for envelope shapes — no separately maintained JSON Schema.

The validate envelope has two shapes:
1. **Success envelope** — `ok: true` with concept count, language list, and warnings array.
2. **Error envelope** — `ok: false` with error code/message/hint (already handled by the existing error emitter).

This ticket covers only the success envelope types. The error envelope is already implemented in `internal/output/errors.go`.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Output"
- Schema ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md) §"Go code is the canonical contract"
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md) §"JSON output"

## Files to create / modify

- `src/internal/output/types.go` (create — new file)

## Type definitions

```go
package output

type ValidateEnvelope struct {
    SchemaVersion int               `json:"schema_version"`
    OK            bool              `json:"ok"`
    Concepts      int               `json:"concepts"`
    Languages     []string          `json:"languages"`
    Warnings      []ValidateWarning `json:"warnings"`
}

type ValidateWarning struct {
    Code      string `json:"code"`
    Message   string `json:"message"`
    ConceptID string `json:"concept_id,omitempty"`
    Line      int    `json:"line,omitempty"`
    Col       int    `json:"column,omitempty"`
}
```

Design notes:
- **Types are exported** — the validate command and tests both import them.
- **`json:"column"` not `json:"col"`** — spec shows `"column"` in the warning shape.
- **`omitempty` on optional fields** — concept_id, line, column are optional per spec.
- **`Languages` must serialize as `[]` not `null`** — the command action must ensure non-nil slice before marshaling.
- **`Warnings` must serialize as `[]` not `null`** — same treatment.

## TDD cycles

### Cycle 1 — Envelope JSON shape
RED: Marshal a `ValidateEnvelope{SchemaVersion: 1, OK: true, Concepts: 2, Languages: []string{"en", "he"}, Warnings: []ValidateWarning{}}` and assert the JSON matches the expected shape from the spec: `{"schema_version":1,"ok":true,"concepts":2,"languages":["en","he"],"warnings":[]}`.
GREEN: Define the struct types with correct json tags.

### Cycle 2 — Warning omitempty
RED: Marshal a `ValidateWarning{Code: "duplicate_id", Message: "..."}` with zero Line/Col and empty ConceptID. Assert the JSON does NOT contain `"concept_id"`, `"line"`, or `"column"` keys.
GREEN: Already passing — `omitempty` handles this.

### Cycle 3 — Warning with all fields
RED: Marshal a warning with all fields populated. Assert JSON contains `"concept_id"`, `"line"`, and `"column"` keys with correct values.
GREEN: Already passing.

## Deviation note

The current implementation defines `validateEnvelope` and `validateWarning` as unexported types local to `src/internal/app/commands/validate.go`. The spec requires these types to live in `internal/output/types.go` as the canonical contract. This ticket creates the types in the correct location. The validate command (T12) will be updated to import from `output` instead of using local types.

## Out of scope

- Error envelope types (already in `output/errors.go`)
- The validate command action itself (T12)
- `EmitJSON` helper (already in `output/emit.go`)

## Acceptance

- `make build` passes
- `ValidateEnvelope` and `ValidateWarning` are exported types in `internal/output/types.go`
- JSON serialization matches the spec's envelope shape exactly
- Optional fields use `omitempty`
- `"column"` not `"col"` in the JSON tag


## Notes

**2026-05-25T18:41:13Z**

Created src/internal/output/types.go with exported ValidateEnvelope and ValidateWarning types. JSON tags match the spec exactly (schema_version, ok, concepts, languages, warnings; warning fields: code, message, concept_id omitempty, line omitempty, column omitempty). Updated commands/validate.go to import and use the exported types instead of local unexported duplicates. Three TDD test cycles in types_test.go: envelope JSON shape, omitempty on zero-valued optional fields, and all-fields-populated warning. make build passes clean.
