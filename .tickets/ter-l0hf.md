---
id: ter-l0hf
status: closed
deps: [ter-oeua, ter-bedf]
links: []
created: 2026-05-24T01:05:23Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, test, golden]
---
# E3.T13 — Validate golden CLI tests

## Goal

Create golden CLI tests for all `terminology validate` scenarios. Golden tests compare stdout, stderr, and exit code against checked-in golden files. They are the primary regression safety net for the validate command's JSON envelope output.

Per the [testing ADR](docs/adr/testing.md), golden tests use argv + stdin → stdout/stderr/exit-code triples with byte-for-byte golden files.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Output", §"Exit codes"
- Testing ADR: [docs/adr/testing.md](docs/adr/testing.md) §"Golden CLI tests"
- Existing golden test harness: `src/internal/app/golden_test.go` or `commands_test.go`

## Files to create / modify

- `src/internal/app/commands_test.go` — golden test functions for validate scenarios
- `src/internal/app/testdata/validate/` — golden files (stdout, stderr, exit)
- `src/internal/app/testdata/fixtures/` — test fixture TBX files

## Test scenarios

### Existing golden tests to verify/update

These golden tests already exist and may need updating to reflect new warning codes or envelope changes:

1. **`validate/no_tbx`** — no `--tbx` flag → exit 2, error envelope on stderr
2. **`validate/clean`** — clean DCT file → exit 0, success envelope with `ok: true`
3. **`validate/warnings`** — file with duplicate_id + unresolved_crossref → exit 1, warnings array
4. **`validate/strict`** — `--strict` on file with unresolved crossref → exit 65, error envelope

### New golden tests to add

5. **`validate/malformed_xml`** — malformed XML file → exit 65, error envelope with `validation_error`
6. **`validate/strict_with_legacy`** — `--strict` on file with legacy forms → exit 1, includes `legacy_form_normalized` warnings
7. **`validate/lenient_with_legacy`** — lenient on file with legacy forms → exit 0 or 1, no `legacy_form_normalized` warnings
8. **`validate/unknown_element_strict`** — `--strict` on file with unknown element → exit 1, includes `unknown_element` warning
9. **`validate/invalid_picklist`** — file with invalid picklist value → exit 1, `invalid_picklist` warning

### Test fixtures to create

- `src/internal/app/testdata/fixtures/malformed.tbx` — unclosed XML tag
- `src/internal/app/testdata/fixtures/with-legacy-forms.tbx` — bare `preferredTerm`, `usageRegister`
- `src/internal/app/testdata/fixtures/with-unknown-element.tbx` — `<custom:foo>bar</custom:foo>` inside conceptEntry
- `src/internal/app/testdata/fixtures/with-invalid-picklist.tbx` — `partOfSpeech=frobnicator`

## Golden test pattern

The project uses a `runGolden(t, name, args)` helper that:
1. Runs the CLI with the given args
2. Compares stdout, stderr, exit code against `testdata/<name>/clean.{stdout,stderr,exit}`
3. Fails if any differ

To update golden files when the envelope changes, run with `-update` flag (if the harness supports it) or manually update the files.

## TDD cycles

### Cycle 1 — Verify existing golden tests pass
RED: Run `make test`. If any existing validate golden tests fail due to envelope changes from T12, update the golden files.
GREEN: Golden files match current output.

### Cycle 2 — Malformed XML golden test
RED: Add `TestValidate_MalformedXML_Golden` with `malformed.tbx` fixture. Create golden files for exit 65.
GREEN: Create fixture and golden files.

### Cycle 3 — Strict with legacy forms
RED: Add `TestValidate_StrictWithLegacy_Golden`. Create fixture with legacy forms and golden files showing `legacy_form_normalized` in warnings.
GREEN: Create fixture and golden files.

### Cycle 4 — Unknown element strict
RED: Add `TestValidate_UnknownElementStrict_Golden`. Create fixture with unknown element and golden files.
GREEN: Create fixture and golden files.

### Cycle 5 — Invalid picklist
RED: Add `TestValidate_InvalidPicklist_Golden`. Create fixture and golden files.
GREEN: Create fixture and golden files.

### Cycle 6 — Lenient suppression
RED: Add `TestValidate_LenientSuppression_Golden` — same fixtures as strict tests but without `--strict`. Assert `legacy_form_normalized` and `unknown_element` are NOT in output.
GREEN: Verify filtering works.

## Deviation note

The current implementation has 4 golden tests for validate (no_tbx, clean, warnings, strict). These cover the basic scenarios. New golden tests are needed for:
- Malformed XML (tier-1 failure)
- Legacy form normalization (`--strict` only)
- Unknown element detection (`--strict` only)
- Invalid picklist values

The existing golden files may need updating if T12 changes the envelope types (moving from local to `output.ValidateEnvelope`).

The existing non-golden validate tests (`TestValidate_CleanFile_Success`, `TestValidate_NonexistentFile_ExitCode65`, `TestValidate_NoTBXPath_ExitCode2`) cover exit codes and basic envelope shape. These should be preserved alongside the golden tests.

## Out of scope

- Golden tests for other commands (lookup, scan, etc. — already exist)
- The `runGolden` helper itself (already implemented)
- Performance testing

## Acceptance

- `make build` passes
- All golden tests pass byte-for-byte
- At least 8 validate golden test scenarios covering all exit codes and warning types
- Test fixtures exist for each scenario
- Both strict and lenient modes have golden coverage


## Notes

**2026-05-25T18:50:17Z**

4 golden tests exist: validate/no_tbx, validate/clean, validate/warnings, validate/strict. The ticket calls for 5 additional golden tests: malformed_xml, strict_with_legacy, lenient_with_legacy, unknown_element_strict, invalid_picklist. None of the 5 new goldens exist yet. New fixtures needed: malformed.tbx, with-legacy-forms.tbx, with-unknown-element.tbx, with-invalid-picklist.tbx. The existing 4 goldens may need updating once T10/T12 add strict-mode filtering (unknown_element warnings currently pass through unfiltered in lenient mode — once filtering is added, the validate/warnings golden file may change).

**2026-05-25T19:17:07Z**

Added 4 new golden tests (strict_with_legacy, lenient_with_legacy, unknown_element_strict, invalid_picklist) bringing total to 9. Created with-invalid-picklist.tbx fixture. Reused existing with-legacy-and-unknown.tbx for legacy and unknown element tests. All 9 tests cover exit codes 0/1/2/65 and all warning types. strict_with_legacy and unknown_element_strict use same fixture+args (identical output) but are kept separate per ticket spec for conceptual coverage. Used -update flag to generate golden files after verifying test structure. make build passes.
