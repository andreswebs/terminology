---
id: ter-oeua
status: closed
deps: [ter-mf43, ter-wsy4, ter-eedk, ter-8zuv, ter-at09, ter-a52o, ter-mule, ter-bedf]
links: []
created: 2026-05-24T01:05:22Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, command, validate]
---
# E3.T12 — Validate command action + exit codes

## Goal

Implement the `terminology validate` command action that wires together all validation tiers, produces the JSON envelope on stdout, and returns the correct exit code. This is the integration point — it consumes the reader (tier 1), `Glossary.Validate()` (tiers 2+3), and the envelope types (T2), and handles `--strict` filtering (T10).

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Output", §"Exit codes", §"Tier sequencing", §"--strict"
- Schema ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md) — envelope types from `output/types.go`
- Error-handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md) — exit code mapping

## Files to create / modify

- `src/internal/app/commands/validate.go` — validate command + action

## Command definition

```go
func Validate() *urfcli.Command {
    return &urfcli.Command{
        Name:  "validate",
        Usage: "validate a TBX file against the supported subset",
        Flags: []urfcli.Flag{
            &urfcli.BoolFlag{Name: "strict", Usage: "promote unknown elements and unresolved IDREFs to errors"},
            readFieldsFlag(),
        },
        Action: validateAction,
    }
}
```

## Action flow

```
validateAction:
  1. Get --tbx path (return ErrNoTBXPath if empty)
  2. Call tbx.Load(path) → (glossary, loadWarnings, err)
     - If err → return ErrValidationError.Wrap(err)      [tier-1 failure, exit 65]
  3. Call glossary.Validate(strict) → ValidateResult
     - If res.Errors > 0 → return ErrValidationError.Wrap(...)  [exit 65]
  4. Merge loadWarnings + res.Warnings
  5. Filter by --strict:
     - If NOT strict: remove warnings with code "unknown_element" or "legacy_form_normalized"
  6. Build ValidateEnvelope from output.ValidateEnvelope (T2)
     - SchemaVersion: output.SchemaVersion
     - OK: true
     - Concepts: res.Concepts
     - Languages: res.Languages (ensure non-nil slice)
     - Warnings: merged+filtered list (ensure non-nil slice)
  7. Emit JSON via output.EmitJSON(cmd.Writer, envelope)
  8. If len(warnings) > 0 → return warningsPresent(n)      [exit 1]
  9. Return nil                                              [exit 0]
```

## Exit codes

| Code | Condition |
|------|-----------|
| 0 | Clean — no warnings, no errors |
| 1 | Warnings present, no errors (recoverable) |
| 65 | Tier-1 failure (can't parse) or strict-mode errors |
| 2 | Missing `--tbx` path |
| 3 | I/O error reading the file |

Exit code 1 uses a custom error type `warningsError` that satisfies `terr.Coded` with code `"warnings"` and exit code 1.

## Envelope types (from T2)

The validate command imports `output.ValidateEnvelope` and `output.ValidateWarning` from `internal/output/types.go` instead of defining local types. The conversion from `tbx.Warning` to `output.ValidateWarning` maps fields 1:1.

## TDD cycles

### Cycle 1 — Missing --tbx returns exit 2
RED: Run validate with no `--tbx` flag. Assert exit code 2.
GREEN: Check path, return `ErrNoTBXPath`.

### Cycle 2 — Clean file returns exit 0 with envelope
RED: Run validate with a clean fixture. Assert exit 0, stdout is valid JSON with `ok: true`, `concepts: 1`, non-empty `languages`.
GREEN: Implement full action flow.

### Cycle 3 — Nonexistent file returns exit 65
RED: Run validate with nonexistent file path. Assert exit 65.
GREEN: `tbx.Load` errors, wrapped as `ErrValidationError`.

### Cycle 4 — File with warnings returns exit 1
RED: Run validate with `with-warnings.tbx` fixture. Assert exit 1, `ok: true`, non-empty `warnings` array.
GREEN: Wire warning merging and `warningsPresent` error.

### Cycle 5 — Strict mode with errors returns exit 65
RED: Run validate with `--strict` on file with unresolved cross-refs. Assert exit 65, error envelope with `code: "validation_error"`.
GREEN: Check `res.Errors`, return `ErrValidationError`.

### Cycle 6 — Strict includes legacy_form_normalized
RED: Run validate with `--strict` on file with legacy forms. Assert `legacy_form_normalized` warning in output.
GREEN: Don't filter legacy warnings when strict.

### Cycle 7 — Lenient suppresses unknown_element and legacy_form_normalized
RED: Run validate without `--strict` on file with unknown elements and legacy forms. Assert no `unknown_element` or `legacy_form_normalized` warnings in output.
GREEN: Add filtering logic.

### Cycle 8 — Envelope uses output.ValidateEnvelope types
RED: Assert the output JSON has `schema_version` key (not `schemaVersion` or other casing).
GREEN: Use `output.ValidateEnvelope` with correct json tags.

## Deviation note

The current implementation in `commands/validate.go` is mostly complete but has these deviations from spec:

1. **Local types instead of output/ types** — `validateEnvelope` and `validateWarning` are defined locally as unexported types. They should import from `output/types.go` (T2).
2. **No strict filtering** — the command doesn't filter `unknown_element` or `legacy_form_normalized` warnings by strict mode. It passes all load warnings + validate warnings through without filtering.
3. **Strict error envelope missing warnings** — when strict errors are found, the command returns `ErrValidationError` which produces an error envelope, but the warnings that WERE found (non-promoted) are lost. The spec's behavior here is ambiguous — the error envelope doesn't have a `warnings` field.

Changes needed:
1. Replace local types with `output.ValidateEnvelope` / `output.ValidateWarning`
2. Add warning filtering based on `--strict` mode
3. Decide how to handle the strict error + warnings case (may need spec clarification)

## Out of scope

- The validate command registration in `root.go` (already done in E1)
- Error envelope formatting (already in `output/errors.go`)
- Golden test files (T13)

## Acceptance

- `make build` passes
- Exit codes: 0 (clean), 1 (warnings), 65 (errors/unparseable), 2 (missing path)
- JSON envelope matches spec shape
- Uses `output.ValidateEnvelope` / `output.ValidateWarning` from `output/types.go`
- Strict mode includes `unknown_element` + `legacy_form_normalized`
- Lenient mode filters them out
- Warnings array is always `[]` not `null`
- Languages array is always `[]` not `null`


## Notes

**2026-05-25T18:50:10Z**

Validate command (commands/validate.go) is partially implemented. Already done: uses output.ValidateEnvelope/ValidateWarning from output/types.go (T2 deviation resolved), merges loadWarnings + res.Warnings, exit codes 0/1/2/65 all work, nil-guard on Languages/Warnings arrays. Gap: no --strict filtering — the command passes ALL warnings through regardless of strict mode. It should filter out unknown_element and legacy_form_normalized warnings in lenient mode. This is the same gap noted on T10. Once T10 adds legacy_form_normalized emission in the reader, T12 needs the filtering logic added.

**2026-05-25T19:14:19Z**

Implementation was already complete from prior work. Verified all acceptance criteria: exit codes 0/1/2/65 correct, uses output.ValidateEnvelope/ValidateWarning from output/types.go, strict mode includes unknown_element + legacy_form_normalized warnings, lenient mode filters them, nil guards ensure [] not null for warnings and languages arrays. All 13 validate tests pass, 5 golden test cases match expected output. Full make build passes (vet, lint, test, compile).
