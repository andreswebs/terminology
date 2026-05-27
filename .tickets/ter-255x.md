---
id: ter-255x
status: closed
deps: [ter-ab56, ter-4344]
links: []
created: 2026-05-25T19:37:19Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, output, fields]
---
# E4.T2 — --fields projection engine

## Goal

Implement the `--fields` projection engine in `internal/output/fields.go`. This is shared across all read commands. Given a comma-separated list of dotted paths (e.g. `concept_id,languages.*.preferred.term`), it validates paths against the output struct's `json` tags via reflection, then projects the JSON output to include only the requested fields.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"--fields projection"
- Schema-source-of-truth ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md)

## Files to create / modify

- `src/internal/output/fields.go` — projection engine
- `src/internal/output/fields_test.go` — tests

## Behavior

### Path syntax

- Comma-separated dotted paths: `a,b,c.d`
- `*` for map wildcards: `languages.*.preferred.term`
- Paths validated against struct `json` tags via reflection

### Validation

Unknown paths produce `ErrInvalidField` (from T1) with a message like `"unknown field path: concpet_id"`. The hint enumerates valid paths for the current envelope type. This gives agents deterministic feedback on typos.

### Projection

Given a marshaled JSON value and a set of validated paths, produce a new JSON value containing only the requested fields. Top-level envelope fields (`schema_version`, `ok`) are always included.

### API sketch

```go
func ValidateFields(paths string, envelope any) ([]string, error)
func ProjectFields(data []byte, fields []string) ([]byte, error)
```

`ValidateFields` takes the raw `--fields` string and a zero-value of the envelope struct, walks its `json` tags, and returns validated paths or `ErrInvalidField`.

`ProjectFields` takes marshaled JSON and the validated paths, returns projected JSON.

## TDD cycles

### Cycle 1 — Validate known paths
RED: `ValidateFields("concept_id,subject_field", LookupEnvelope{})` returns `["concept_id", "subject_field"]`, nil error.
GREEN: Implement reflection walker over struct `json` tags.

### Cycle 2 — Reject unknown paths
RED: `ValidateFields("concpet_id", LookupEnvelope{})` returns `ErrInvalidField` with message containing `"concpet_id"`.
GREEN: Return wrapped `ErrInvalidField` for unmatched paths.

### Cycle 3 — Wildcard paths
RED: `ValidateFields("languages.*.preferred.term", LookupEnvelope{})` succeeds.
GREEN: Handle `*` as map wildcard in reflection walker.

### Cycle 4 — Projection filters JSON
RED: Given full lookup JSON and fields `["concept_id"]`, `ProjectFields` returns JSON with only `schema_version`, `ok`, and `results[].concept_id`.
GREEN: Implement JSON projection via `map[string]any` traversal.

### Cycle 5 — Projection with wildcards
RED: Fields `["languages.*.preferred.term"]` — projected JSON retains nested structure under each language key.
GREEN: Handle wildcard expansion during projection.

## Acceptance

- `make build` passes
- Unknown paths → `ErrInvalidField` with valid-path hint
- Projection preserves only requested fields + envelope boilerplate
- Wildcard `*` works for map keys

## Notes

**2026-05-25T19:57:57Z**

Implemented --fields projection engine in internal/output/fields.go. Two public functions: ValidateFields(paths string, envelope any) validates comma-separated dotted paths against struct json tags via reflection (returns ErrInvalidField with valid-path hint on unknown paths); ProjectFields(data []byte, fields []string) projects marshaled JSON to include only requested fields plus schema_version/ok boilerplate. Handles map wildcards (*), pointer-to-struct map values, and slice traversal. Empty input returns nil. 12 tests covering known paths, unknown paths, wildcards, array projection, map pointer values, and error hints.
