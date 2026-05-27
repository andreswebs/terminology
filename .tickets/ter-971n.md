---
id: ter-971n
status: closed
deps: [ter-ab56]
links: []
created: 2026-05-25T19:37:20Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, schema, output]
---
# E4.T8 — Schema: output type registry

## Goal

Create a registry in `internal/output` that maps command names to their envelope types. The `schema` command (T9/T10) uses this to reflect over envelope shapes per command without hardcoding type knowledge.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"schema — reflective introspection" source 2
- Schema-source-of-truth ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md)

## Files to create / modify

- `src/internal/output/registry.go` — envelope type registry
- `src/internal/output/registry_test.go` — tests

## Behavior

```go
func RegisterEnvelope(command string, zero any)
func EnvelopeFor(command string) (any, bool)
func AllEnvelopes() map[string]any
```

Each command registers its zero-value envelope during init or setup:

```go
func init() {
    output.RegisterEnvelope("validate", ValidateEnvelope{})
    output.RegisterEnvelope("lookup", LookupEnvelope{})
}
```

`EnvelopeFor("validate")` returns the zero-value `ValidateEnvelope{}` which the schema walker can reflect over to discover fields and JSON tags.

`AllEnvelopes()` returns all registered mappings (used for the full `terminology schema` output).

## TDD cycles

### Cycle 1 — Register and retrieve
RED: `RegisterEnvelope("test", TestEnvelope{})`, then `EnvelopeFor("test")` returns the value and `true`.
GREEN: Implement map-backed registry.

### Cycle 2 — Unknown command returns false
RED: `EnvelopeFor("nonexistent")` returns `nil, false`.
GREEN: Map lookup with ok check.

### Cycle 3 — AllEnvelopes returns copy
RED: Modify returned map, call again. Assert original unchanged.
GREEN: Return copy of map.

### Cycle 4 — Existing commands registered
RED: After importing relevant packages, `EnvelopeFor("validate")` succeeds.
GREEN: Add init registration calls in `types.go`.

## Acceptance

- `make build` passes
- Commands can register their envelope types
- `schema` can discover envelope shapes via `EnvelopeFor`/`AllEnvelopes`
- Registry returns copies (not mutable references)

## Notes

**2026-05-26T00:46:48Z**

Implemented map-backed envelope type registry in internal/output/registry.go with RegisterEnvelope, EnvelopeFor, AllEnvelopes. Added init() in types.go to register ValidateEnvelope and LookupEnvelope. AllEnvelopes returns a defensive copy. Tests use saveAndResetEnvelopes helper with t.Cleanup to restore init-registered entries after tests that reset the map.
