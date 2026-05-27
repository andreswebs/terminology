---
id: ter-cnwx
status: closed
deps: [ter-ab56, ter-ir9c, ter-uw6e, ter-t8lr]
links: []
created: 2026-05-25T19:37:27Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, golden, testing]
---
# E4.T16 — Golden CLI tests for lookup, schema, extract

## Goal

Create golden CLI tests for all three E4 commands: `lookup`, `schema`, `extract`. Each test captures argv + stdin → stdout/stderr/exit-code triples with byte-for-byte golden files. Uses the same `testscript`-style harness from E3.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md)
- Testing ADR: [docs/adr/testing.md](docs/adr/testing.md) — golden test conventions

## Files to create / modify

- `src/internal/app/testdata/` — golden test script files and fixtures
- `src/internal/app/golden_test.go` — add test entries (or extend existing)

## Test cases

### Lookup

1. **lookup/found** — `terminology --tbx fixture.tbx lookup tzimtzum` → exit 0, envelope with results
2. **lookup/not_found** — `terminology --tbx fixture.tbx lookup nonexistent` → exit 1, `results: []`
3. **lookup/no_tbx** — `terminology lookup tzimtzum` → exit 2, error envelope `no_tbx_path`
4. **lookup/lang_filter** — `terminology --tbx fixture.tbx lookup tzimtzum --lang he` → exit 0, only Hebrew results
5. **lookup/case_insensitive** — `terminology --tbx fixture.tbx lookup TZIMTZUM` → exit 0, matches case-insensitively
6. **lookup/fields** — `terminology --tbx fixture.tbx lookup tzimtzum --fields concept_id` → projected output
7. **lookup/invalid_field** — `terminology --tbx fixture.tbx lookup tzimtzum --fields concpet_id` → exit 2, `invalid_field` error

### Schema

8. **schema/full** — `terminology schema` → exit 0, has `commands`, `envelopes`, `error_codes`
9. **schema/command_filter** — `terminology schema --command validate` → exit 0, single command entry
10. **schema/unknown_command** — `terminology schema --command nonexistent` → exit 2, error envelope

### Extract

11. **extract/basic** — `terminology extract testdata/corpus.md` → exit 0, candidates in envelope
12. **extract/no_candidates** — `terminology extract testdata/empty.md` → exit 1, `candidates: []`
13. **extract/exclude** — `terminology extract --exclude fixture.tbx testdata/corpus.md` → glossary terms excluded
14. **extract/script_filter** — `terminology extract --script hebrew testdata/mixed.md` → only Hebrew candidates
15. **extract/min_freq** — `terminology extract --min-freq 5 testdata/corpus.md` → frequency-gated
16. **extract/fields** — `terminology extract --fields term,frequency testdata/corpus.md` → projected output

## Test fixtures to create

- `testdata/corpus.md` — markdown with capitalized phrases, foreign-script tokens, repeated terms, and code blocks
- `testdata/empty.md` — markdown with only code blocks or common words below threshold
- `testdata/mixed.md` — markdown mixing Latin and Hebrew scripts
- `testdata/stopwords.txt` — sample stopwords file

## TDD cycles

### Cycle 1 — Lookup golden tests
RED: Add lookup golden test scripts. Run `go test`. Assert golden files created.
GREEN: Golden files captured from working lookup command (T7).

### Cycle 2 — Schema golden tests
RED: Add schema golden test scripts. Run `go test`.
GREEN: Golden files captured from working schema command (T10).

### Cycle 3 — Extract golden tests
RED: Add extract golden test scripts.
GREEN: Golden files captured from working extract command (T15).

### Cycle 4 — Error case goldens
RED: Add golden tests for error paths (no_tbx, invalid_field, unknown_command, no_candidates).
GREEN: Error envelopes match golden files.

## Acceptance

- `make build` passes
- All 16 golden test cases pass
- Fixtures cover happy path + error paths + flag combinations
- Tests use byte-for-byte golden file comparison
- Exit codes verified in each test

## Out of scope

- Performance benchmarks
- Fuzz testing (E9)

## Notes

**2026-05-26T01:21:14Z**

Implemented all 16 golden CLI tests for lookup (7), schema (3), and extract (6). Created extract-mixed.md fixture for mixed-script testing. Removed stale schema stub golden files (testdata/schema/clean.*) that referenced exit code 75. The old TestExtract_Golden (testdata/extract/clean.*) was kept alongside the new extract/basic subdirectory — both coexist since golden_test.go maps names to directories. Schema golden output is large (~100+ KB for full) since it serializes the entire CLI surface reflectively; this is expected and deterministic.
