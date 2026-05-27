---
id: ter-qs6f
status: closed
deps: [ter-c4ra, ter-bmuk, ter-c7vk]
links: []
created: 2026-05-26T13:49:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, task, app, scan]
---
# E5.T8 — Scan command action

## Goal

Replace the `underConstruction` stub in `scan.go` with the real scan command action. Wire the matcher (T6) to the command surface: load TBX, parse markdown spans, run the matcher, build the scan envelope, apply `--fields` projection, emit JSON, and return the correct exit code.

## Refs

- E5 spec: [docs/specs/005-matcher.md](docs/specs/005-matcher.md)
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"terminology scan"
- Error handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md)

## Files to create / modify

- `src/internal/app/commands/scan.go` — implement `scanAction`, replace `underConstruction`
- `src/internal/app/commands/scan_test.go` — unit tests (if needed beyond golden tests)

## Behavior

### Command flow

1. Resolve TBX path (`--tbx` → `TERMINOLOGY_TBX` → error).
2. Load glossary via `tbx.Load(path)`.
3. Read the markdown file (positional arg `FILE`).
4. Parse markdown into spans via `markdown.Spans(src)`.
5. Create matcher: `match.New(glossary, lang)`.
6. Run `matcher.Scan(src, spans)`.
7. Build `ScanEnvelope` with file path, matches, and summary.
8. Apply `--fields` projection if specified.
9. Emit JSON to stdout.
10. Exit 0 always (scan is informational, per spec).

### Flags

Already defined in the stub:
- `--lang` — restrict to a language section (passed to `match.New`)
- `--context N` — context window around each match (chars, default 80)
- `--fields` — field projection

### Exit codes

- `0` — always (scan is informational)
- `2` — usage error (no TBX path, invalid field)
- `3` — I/O error (file not readable)

### Context window

The `--context` flag controls how many characters of surrounding text appear in each match's `context` field. Passed through to the matcher or extracted at the command level.

## TDD cycles

### Cycle 1 — Basic scan
RED: `terminology scan FILE --tbx fixture.tbx` → exit 0, envelope with matches.
GREEN: Wire full pipeline.

### Cycle 2 — No TBX path
RED: `terminology scan FILE` (no --tbx, no env) → exit 2, `no_tbx_path` error.
GREEN: Reuse existing TBX resolution logic.

### Cycle 3 — Language filter
RED: `terminology scan FILE --tbx fixture.tbx --lang he` → only Hebrew-language matches.
GREEN: Pass lang to `match.New`.

### Cycle 4 — Fields projection
RED: `terminology scan FILE --tbx fixture.tbx --fields concept_id,line` → projected output.
GREEN: Apply field projection to envelope.

### Cycle 5 — File not found
RED: `terminology scan nonexistent.md --tbx fixture.tbx` → exit 3, I/O error.
GREEN: Handle file read error.

## Acceptance

- `make build` passes
- Scan command replaces `underConstruction` stub
- Full pipeline: TBX load → markdown parse → matcher scan → envelope → JSON
- Exit 0 always for successful scans (informational)
- `--lang`, `--context`, `--fields` flags work
- Error cases (no TBX, file not found) emit correct error envelopes


## Notes

**2026-05-26T14:40:01Z**

Replaced underConstruction stub with full scanAction implementation. Pipeline: resolve TBX path → tbx.Load → os.ReadFile → markdown.Spans → match.New(g, lang, PolicyFor(lang)) → matcher.Scan → build ScanEnvelope → field projection → emit JSON. Exit 0 always for successful scans. Exit 2 for no TBX path. Exit 3 for I/O errors (file not found). Updated stub_test.go to use 'check' command instead of 'scan' since scan is no longer a stub. Removed stale golden files from testdata/scan/. Added scan-sample.md test fixture with English and Hebrew term occurrences. All tests and make build pass cleanly.
