---
id: ter-5sps
status: closed
deps: [ter-ab56, ter-pwwn, ter-971n]
links: [ter-ix1i]
created: 2026-05-25T19:37:20Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, schema, reflection]
---
# E4.T9 — Schema: reflective walkers

## Goal

Implement the three reflective walkers that power `terminology schema`: command tree walker, output type reflector, and terr sentinel enumerator. These produce the data that the schema command action (T10) assembles into the final JSON output.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"schema — reflective introspection"
- Schema-source-of-truth ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md)

## Files to create / modify

- `src/internal/app/schema.go` (or `src/internal/app/introspect.go`) — walker functions
- `src/internal/app/schema_test.go` — tests

## Behavior

### Source 1: Command tree walker

Walk `*urfcli.Command` tree recursively, collecting for each command:
- `name` — command name
- `usage` — usage string
- `flags` — array of `{name, type, default, aliases, required, usage}`
- `arguments` — array of `{name, min, max}`

```go
type CommandDesc struct {
    Name      string     `json:"name"`
    Usage     string     `json:"usage"`
    Flags     []FlagDesc `json:"flags"`
    Arguments []ArgDesc  `json:"arguments,omitempty"`
}

func WalkCommands(root *urfcli.Command) []CommandDesc
```

### Source 2: Output type reflector

Uses the output type registry (T8) to reflect over each envelope struct and produce a JSON-schema-like shape:

```go
type EnvelopeDesc struct {
    Fields []FieldDesc `json:"fields"`
}

type FieldDesc struct {
    Name     string      `json:"name"`
    Type     string      `json:"type"`
    Children []FieldDesc `json:"children,omitempty"`
}

func ReflectEnvelope(zero any) EnvelopeDesc
```

Walk struct fields via `reflect`, use `json` tags for field names, map Go types to JSON types (`string`, `number`, `boolean`, `array`, `object`).

### Source 3: Terr sentinel enumerator

Uses `terr.All()` (T4) to list all error codes:

```go
type ErrorCodeDesc struct {
    Code     string `json:"code"`
    ExitCode int    `json:"exit_code"`
    Message  string `json:"message"`
    Hint     string `json:"hint,omitempty"`
}

func EnumerateErrors() []ErrorCodeDesc
```

## TDD cycles

### Cycle 1 — Walk command tree
RED: Build a minimal urfave command tree with 2 commands and flags. `WalkCommands` returns descriptions with correct names and flag details.
GREEN: Implement recursive command walk.

### Cycle 2 — Flag types and defaults
RED: Flags with different types (string, bool, int) — assert correct type strings and defaults.
GREEN: Switch on flag type in walker.

### Cycle 3 — Reflect envelope fields
RED: `ReflectEnvelope(ValidateEnvelope{})` returns fields including `schema_version` (number), `ok` (boolean), `warnings` (array).
GREEN: Implement struct reflection with json tag parsing.

### Cycle 4 — Nested struct reflection
RED: `ReflectEnvelope(LookupEnvelope{})` — nested `LookupResult` fields appear as children.
GREEN: Recursive field reflection.

### Cycle 5 — Enumerate errors
RED: `EnumerateErrors()` returns entries for all known sentinels.
GREEN: Map `terr.All()` to `ErrorCodeDesc` values.

## Acceptance

- `make build` passes
- Command walker captures all flags with types, defaults, aliases
- Envelope reflector handles nested structs, maps, slices
- Error enumerator lists all sentinels from registry

## Notes

**2026-05-26T00:56:12Z**

Implemented three reflective walkers in src/internal/schema/schema.go:

1. WalkCommands(root *urfcli.Command) []CommandDesc — recursive command tree walk extracting names, usage, flags (with type/default/aliases/required), arguments (with min/max bounds), and nested subcommands.

2. ReflectEnvelope(zero any) EnvelopeDesc — struct reflection using json tags to produce field descriptors with JSON types (string/number/boolean/array/object) and recursive children for nested structs, slices, maps, and pointer types.

3. EnumerateErrors() []ErrorCodeDesc — walks terr.All() sentinel registry, maps to code/exit_code/message/hint, sorted by code.

Package placed in internal/schema/ (not internal/app/) to avoid circular dependency: commands/ needs to import these walkers for the T10 schema command action, but commands/ cannot import app/ (app imports commands).

Unit tests in schema/schema_test.go (15 tests). Integration tests against real command tree, real envelopes, and real sentinels in app/schema_integration_test.go (6 tests).

flagUsage() helper uses reflect to read the Usage field from urfave flag types because the urfave Flag interface doesn't expose Usage directly.
