---
id: ter-eedk
status: closed
deps: [ter-bedf]
links: [ter-9dho]
created: 2026-05-24T01:05:14Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, validation, tier1]
---
# E3.T4 — Tier-1 well-formedness validation

## Goal

Implement tier-1 validation: verify the TBX file is well-formed XML with the required structural elements present (`<tbx>`, `<conceptEntry>`, `<langSec>`, `<termSec>`) and their required attributes. Tier 1 is a gate — if it fails, the command returns immediately with `ErrValidationError` (exit 65) because the domain model can't be built. Tiers 2+3 never run.

In practice, the E2 reader (`linguist/reader.go`) already rejects malformed XML by returning an error from `Decode()`. This ticket formalizes that behavior as tier-1 and ensures structured error reporting.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Scope" tier 1, §"Tier sequencing"
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"Required structural elements"

## Files to create / modify

- `src/internal/tbx/linguist/reader.go` — ensure structural checks return clear errors
- `src/internal/tbx/linguist/reader_test.go` — add tier-1 failure test cases

## Behavior

Tier-1 checks are implicit in the reader's `Decode()` method. When `encoding/xml.Decoder` encounters malformed XML or the reader can't find required elements, it returns an error. The validate command wraps this in `ErrValidationError.Wrap(err)`.

Required structural elements (presence required for a valid TBX file):
1. `<tbx>` root element with `type` attribute
2. At least one `<conceptEntry>` with `id` attribute
3. Each `<conceptEntry>` must contain at least one `<langSec>` with `xml:lang` attribute
4. Each `<langSec>` must contain at least one `<termSec>` with a `<term>` child

Note: the "required structure" checks overlap with tier-3's `missing_term` warning. The distinction is: tier-1 catches XML that can't be parsed at all (e.g. missing closing tags, invalid XML), while tier-3 catches structurally valid XML with semantic issues (e.g. a `<langSec>` that parses fine but has no `<termSec>`).

## Test fixtures to create

`src/internal/tbx/linguist/testdata/malformed/`:

1. `bad-xml.tbx` — invalid XML (unclosed tag)
2. `missing-tbx-type.tbx` — `<tbx>` without `type` attribute
3. `empty-body.tbx` — valid XML structure but `<body>` has no `<conceptEntry>`

## TDD cycles

### Cycle 1 — Malformed XML returns error
RED: Pass `bad-xml.tbx` (contains unclosed tag) to `Decode()`. Assert error is non-nil.
GREEN: Already passing — `encoding/xml.Decoder` rejects malformed XML.

### Cycle 2 — Missing type attribute
RED: Pass `missing-tbx-type.tbx` to `Decode()`. Assert error is non-nil (reader can't detect dialect without `type`).
GREEN: Ensure `detectStyle` or the top-level decode returns an error when `type` attribute is missing from `<tbx>`.

### Cycle 3 — Empty body produces empty Glossary (not error)
RED: Pass `empty-body.tbx` to `Decode()`. Assert it returns a Glossary with `len(Concepts) == 0` and no error. An empty-but-valid file is not a tier-1 failure — it's just empty.
GREEN: Ensure the reader handles an empty `<body>` gracefully.

### Cycle 4 — Validate command wraps reader error as ErrValidationError
RED: Integration test: run validate with a malformed fixture file. Assert exit code is 65 and error envelope has `code: "validation_error"`.
GREEN: Already passing — `validateAction` wraps `tbx.Load` errors with `ErrValidationError.Wrap(err)`.

## Deviation note

The current implementation handles tier-1 implicitly: `tbx.Load()` calls `Decode()`, and if the reader errors, `validateAction` wraps it with `ErrValidationError.Wrap(err)`. This matches the spec's intent but isn't explicitly structured as a "tier-1 check." The test in `commands_test.go` (`TestValidate_NonexistentFile_ExitCode65`) covers the I/O error case (file not found → exit 65), but there are no tests for malformed XML specifically. This ticket adds those tests and the malformed fixtures.

No structural changes to the reader are expected — the implicit tier-1 behavior is correct. The main deliverable is test coverage and fixtures proving the behavior.

## Out of scope

- Tier-2 schema validation (T8, T9)
- Tier-3 semantic validation (T5, T6, T7)
- Error envelope formatting (already in `output/errors.go`)

## Acceptance

- `make build` passes
- Malformed XML → `Decode()` returns error
- Validate command returns exit 65 for unparseable files
- Test fixtures exist for common malformation scenarios
- Empty-but-valid file → no error, empty Glossary


## Notes

**2026-05-25T18:49:46Z**

Tier-1 well-formedness: the reader implicitly handles malformed XML (encoding/xml.Decoder rejects it) and missing dialect type (reader returns unsupported_dialect error). The validate command wraps these as ErrValidationError (exit 65). This behavior is correct and spec-aligned. Gap: the ticket calls for test fixtures in testdata/malformed/ (bad-xml.tbx, missing-tbx-type.tbx, empty-body.tbx) — only invalid-picklist.tbx exists currently in that directory. The malformed fixtures and dedicated reader-level tests for tier-1 failure modes still need to be created. The validate command integration tests (commands_test.go TestValidate_NonexistentFile_ExitCode65) cover the exit code path but not malformed XML specifically.

**2026-05-25T19:04:53Z**

Completed tier-1 well-formedness validation. Created three test fixtures in testdata/malformed/: bad-xml.tbx (unclosed tag), missing-tbx-type.tbx (no type attribute), empty-body.tbx (valid but empty). Added reader-level tests (TestDecode_MalformedXML_ReturnsError, TestDecode_EmptyBody_ReturnsEmptyGlossary), tbx.Load-level tests (TestLoad_MalformedXML, TestLoad_EmptyBody), and integration tests (TestValidate_MalformedXML_ExitCode65 with error code verification, golden test). All tier-1 behavior was already implicit in the reader (encoding/xml rejects malformed XML) and validate command (wraps Load errors as ErrValidationError exit 65). No structural code changes needed — the deliverable was test coverage and fixtures proving the behavior.
