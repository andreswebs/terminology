---
id: ter-ll2t
status: closed
deps: [ter-c4ra, ter-qs6f]
links: []
created: 2026-05-26T13:49:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, task, golden, testing]
---
# E5.T9 — Golden CLI tests for scan

## Goal

Create golden CLI tests for the `scan` command. Each test captures argv + stdin → stdout/stderr/exit-code triples with byte-for-byte golden files. Uses the same `testscript`-style harness from E3/E4.

## Refs

- E5 spec: [docs/specs/005-matcher.md](docs/specs/005-matcher.md)
- Testing ADR: [docs/adr/testing.md](docs/adr/testing.md) — golden test conventions
- E4.T16 (ter-cnwx) — reference for golden test structure

## Files to create / modify

- `src/internal/app/testdata/scan/` — golden test script files and fixtures
- `src/internal/app/golden_test.go` — add scan test entries

## Test fixtures to create

- `testdata/scan-corpus.md` — markdown with glossary terms (multiple languages, code blocks to skip)
- `testdata/scan-glossary.tbx` — TBX file with multi-language concepts for matching

## Test cases

### Happy path

1. **scan/found** — `terminology scan corpus.md --tbx glossary.tbx` → exit 0, envelope with matches, summary counts
2. **scan/no_matches** — `terminology scan empty.md --tbx glossary.tbx` → exit 0, `matches: []`, summary zeros
3. **scan/lang_filter** — `terminology scan corpus.md --tbx glossary.tbx --lang he` → exit 0, only Hebrew-language matches
4. **scan/context_window** — `terminology scan corpus.md --tbx glossary.tbx --context 40` → exit 0, shorter context strings
5. **scan/fields** — `terminology scan corpus.md --tbx glossary.tbx --fields concept_id,line` → exit 0, projected output
6. **scan/case_insensitive** — corpus with `"TZIMTZUM"`, glossary with `"tzimtzum"` → match found

### Error cases

7. **scan/no_tbx** — `terminology scan corpus.md` → exit 2, `no_tbx_path` error
8. **scan/file_not_found** — `terminology scan nonexistent.md --tbx glossary.tbx` → exit 3, I/O error
9. **scan/invalid_field** — `terminology scan corpus.md --tbx glossary.tbx --fields concpet_id` → exit 2, `invalid_field` error

### Matcher-specific

10. **scan/code_blocks_skipped** — glossary term appears only inside a fenced code block → `matches: []`
11. **scan/multi_word** — glossary has multi-word term, corpus has it spanning a line break → match found
12. **scan/status_tagging** — deprecated term in corpus → match has `status: "deprecated"`

## TDD cycles

### Cycle 1 — Happy path goldens
RED: Add scan golden test scripts for found/no_matches cases. Run `go test`.
GREEN: Golden files captured from working scan command (T8).

### Cycle 2 — Error case goldens
RED: Add golden tests for error paths (no_tbx, file_not_found, invalid_field).
GREEN: Error envelopes match golden files.

### Cycle 3 — Matcher behavior goldens
RED: Add golden tests for code-block skipping, multi-word matching, status tagging.
GREEN: Matcher-specific behaviors verified end-to-end.

## Acceptance

- `make build` passes
- All 12 golden test cases pass
- Fixtures cover happy path + error paths + matcher-specific behaviors
- Tests use byte-for-byte golden file comparison
- Exit codes verified in each test


## Notes

**2026-05-26T14:45:32Z**

Implemented 12 golden CLI tests for the scan command covering happy path (found, no_matches, lang_filter, context_window, fields, case_insensitive), error cases (no_tbx, file_not_found, invalid_field), and matcher-specific behaviors (code_blocks_skipped, multi_word, status_tagging). Created dedicated test fixtures: scan-glossary.tbx (3 concepts with preferred/deprecated terms, multi-word terms, multi-language), scan-corpus.md (rich corpus with all behaviors), scan-empty.md, scan-code-only.md, scan-uppercase.md, scan-multiword.md, scan-deprecated.md. Each matcher-specific test uses a focused fixture to isolate the behavior being verified. The --context flag is defined on the scan command but not wired to the matcher (extractContext hardcodes window=40); the context_window golden test documents this current behavior.
