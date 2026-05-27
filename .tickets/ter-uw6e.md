---
id: ter-uw6e
status: closed
deps: [ter-ab56, ter-5sps]
links: []
created: 2026-05-25T19:37:20Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, schema, command]
---
# E4.T10 — Schema command action + --command filter

## Goal

Replace the `underConstruction` stub in `commands/schema.go` with the real action. Assemble output from the three reflective walkers (T9) into the schema JSON envelope. Support `--command NAME` to filter to a single command's entry.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"schema — reflective introspection"
- Schema-source-of-truth ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md)

## Files to create / modify

- `src/internal/app/commands/schema.go` — implement `schemaAction`
- `src/internal/app/commands_test.go` — integration tests

## Behavior

### Full output (no `--command`)

```json
{
  "schema_version": 1,
  "commands": [ ... ],
  "envelopes": { "validate": { ... }, "lookup": { ... }, ... },
  "error_codes": [ ... ]
}
```

### Filtered output (`--command validate`)

```json
{
  "schema_version": 1,
  "name": "validate",
  "flags": [ ... ],
  "envelope": { ... },
  "exit_codes": [0, 1, 65, 3]
}
```

### Exit codes per command

The per-command `exit_codes` array is derived from the error sentinels that the command can produce. This requires either a mapping from commands to their possible errors, or a convention-based approach. For v1, derive from the envelope type's associated sentinels + the universal codes (0, 1, 2).

### Error handling

- `--command` with unknown command name → exit 2 (usage error) with error envelope.
- No `--tbx` required (schema is reflective, not data-dependent).

## TDD cycles

### Cycle 1 — Full schema output
RED: Run `terminology schema`. Assert JSON has `schema_version`, `commands`, `envelopes`, `error_codes` keys.
GREEN: Assemble walker results into envelope.

### Cycle 2 — Commands array populated
RED: Assert `commands` array contains entries for `validate`, `lookup`, `schema`, `extract`.
GREEN: Walk the live command tree.

### Cycle 3 — --command filter
RED: `terminology schema --command validate`. Assert output has `name: "validate"` and `flags` array.
GREEN: Filter command tree, select single envelope.

### Cycle 4 — Unknown command → exit 2
RED: `terminology schema --command nonexistent`. Assert exit 2 and error envelope.
GREEN: Return usage error for unknown command.

### Cycle 5 — Error codes enumerated
RED: `error_codes` array contains `"validation_error"`, `"no_tbx_path"`, `"invalid_field"`.
GREEN: Wire `EnumerateErrors()` output.

## Acceptance

- `make build` passes
- Stub replaced with real action
- Full and filtered outputs match spec shapes
- No `--tbx` required
- Unknown `--command` → exit 2

## Notes

**2026-05-26T01:09:30Z**

Replaced underConstruction stub in commands/schema.go with schemaAction. Implementation: (1) Full output assembles schema.WalkCommands + output.AllEnvelopes (reflected via schema.ReflectEnvelope) + schema.EnumerateErrors into a single JSON envelope with schema_version, commands, envelopes, error_codes. (2) --command NAME filters to a single command using findCommand recursive search; outputs name, usage, flags, arguments, commands (for parent commands), envelope (if registered), and error_codes. Unknown --command returns exit 2 with unknown_command code. (3) No --tbx required since schema is reflective. (4) Two envelope types: schemaFullEnvelope and schemaFilteredEnvelope. (5) Updated existing stub tests (TestSchema_Stub_Golden, TestSchema_WithCommand_Stub) to real behavior tests: TestSchema_FullOutput, TestSchema_CommandsPopulated, TestSchema_FilteredCommand, TestSchema_UnknownCommand_ExitCode2, TestSchema_ErrorCodesEnumerated, TestSchema_EnvelopesPopulated, TestSchema_NoTBXRequired.
