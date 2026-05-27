---
id: ter-0db2
status: closed
deps: [ter-hole, ter-wx2k]
links: []
created: 2026-05-26T17:24:49Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-7fyo
tags: [e6, task, golden, testing]
---
# E6.T9 — Golden CLI tests for check

## Goal

Create golden CLI tests for the `check` command. Each test captures argv + stdin → stdout/stderr/exit-code triples with byte-for-byte golden files. Uses the same `runGolden` harness from E3/E4/E5.

## Refs

- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md)
- Testing ADR: [docs/adr/testing.md](docs/adr/testing.md) — golden test conventions
- E5.T9 (ter-ll2t) — reference for golden test structure
- Golden harness: `src/internal/app/golden_test.go`

## Files to create / modify

- `src/internal/app/testdata/check/` — golden test script files
- `src/internal/app/testdata/fixtures/` — test fixtures (TBX + markdown files)
- `src/internal/app/commands_test.go` — add check golden test entries

## Test fixtures to create

- `check-glossary.tbx` — TBX with concepts having preferred, admitted, deprecated terms in multiple languages (es, he, en)
- `check-source.md` — Spanish source with glossary terms (frontmatter `lang: es`)
- `check-target-clean.md` — Hebrew target with all preferred terms present
- `check-target-missing.md` — target missing a preferred term
- `check-target-forbidden.md` — target with a deprecated variant
- `check-target-admitted.md` — target with an admitted variant
- `check-source-nofm.md` — source without frontmatter (for language_required test)
- `check-target-nofm.md` — target without frontmatter

## Test cases

### Happy path

1. **check/clean** — `terminology check src.md tgt.md --tbx glossary.tbx` → exit 0, ok true, violations empty
2. **check/clean_frontmatter** — source and target have `lang:` in frontmatter, no flags needed → exit 0

### Violation cases

3. **check/missing** — preferred target absent → exit 1, `missing` violation with source_term, expected_target, source_occurrences
4. **check/forbidden_variant** — deprecated variant in target → exit 1, `forbidden_variant` with line, column, context
5. **check/strict_admitted** — `--strict` + admitted variant → exit 1, `admitted_variant` violation
6. **check/admitted_warning** — admitted variant without `--strict` → exit 0, warning in warnings array

### Language resolution

7. **check/lang_from_flags** — no frontmatter, languages via `--source-lang` and `--target-lang` → works
8. **check/lang_required** — no frontmatter, no flag → exit 2, `language_required` error

### Flags

9. **check/context_window** — `--context 40` → shorter context strings in violations
10. **check/fields** — `--fields violations.concept_id,violations.type` → projected output

### Error cases

11. **check/no_tbx** — `terminology check src.md tgt.md` → exit 2, `no_tbx_path`
12. **check/file_not_found** — nonexistent source file → exit 3, `io_error`
13. **check/invalid_field** — `--fields concpet_id` → exit 2, `invalid_field`

### Ordering

14. **check/violation_ordering** — multiple violations → sorted by (line, column), missing at tail

## TDD cycles

### Cycle 1 — Happy path goldens
RED: Add check golden tests for clean/missing. Run `go test`.
GREEN: Golden files captured from working check command (T7).

### Cycle 2 — Violation goldens
RED: Add golden tests for forbidden_variant, strict_admitted, admitted_warning.
GREEN: Violation envelopes match golden files.

### Cycle 3 — Error case goldens
RED: Add golden tests for no_tbx, file_not_found, invalid_field, lang_required.
GREEN: Error envelopes match golden files.

### Cycle 4 — Ordering + flags
RED: Add violation_ordering, context_window, fields goldens.
GREEN: All golden files match.

## Acceptance

- `make build` passes
- All 14 golden test cases pass
- Fixtures cover happy path + violations + errors + language resolution + flags
- Tests use byte-for-byte golden file comparison
- Exit codes verified in each test
- Existing check stub test replaced

## Notes

**2026-05-26T18:39:21Z**

Implemented 14 golden CLI tests for the check command covering: happy path (clean, clean_frontmatter), violation cases (missing, forbidden_variant, strict_admitted, admitted_warning), language resolution (lang_from_flags, lang_required), flags (context_window, fields), error cases (no_tbx, file_not_found, invalid_field), and ordering (violation_ordering). Created 3 new test fixtures: check-target-admitted.md (admitted variant), check-target-nofm.md (no frontmatter), check-target-multi-violations.md (multiple violations for ordering test). All tests use the existing runGolden harness with byte-for-byte golden file comparison. Exit codes verified: 0 (clean/frontmatter/lang_from_flags), 1 (missing/forbidden/strict/admitted/ordering), 2 (no_tbx/lang_required/invalid_field), 3 (file_not_found).
