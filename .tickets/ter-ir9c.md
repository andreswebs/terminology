---
id: ter-ir9c
status: closed
deps: [ter-ab56, ter-2tcq, ter-lobd, ter-255x]
links: []
created: 2026-05-25T19:37:19Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, lookup, command]
---
# E4.T7 — Lookup command action + exit codes

## Goal

Replace the `underConstruction` stub in `commands/lookup.go` with the real action. Wire the lookup matching (T6), envelope type (T5), and --fields projection (T2). Handle exit codes per spec.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"lookup — match policy"
- Error-handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md) — exit code semantics

## Files to create / modify

- `src/internal/app/commands/lookup.go` — implement `lookupAction`
- `src/internal/app/commands_test.go` — integration tests

## Behavior

1. Read `--tbx` path from root flag. If empty → `ErrNoTBXPath` (exit 2).
2. `tbx.Load(path)` → glossary. If error → wrap as appropriate.
3. Get `TERM` positional arg and `--lang` flag.
4. `glossary.Lookup(term, lang)` → matches.
5. Convert matches to `output.LookupEnvelope`.
6. If `--fields` is set, validate and project (T2).
7. Emit JSON via `output.EmitJSON`.
8. If `len(results) == 0` → exit 1 (not found, recoverable).

### Exit codes

| Code | Condition |
|------|-----------|
| 0    | Found results |
| 1    | No results (not found) |
| 2    | Usage error (no TBX path, invalid field) |

### Not-found envelope

```json
{"schema_version":1,"ok":true,"results":[]}
```

`ok` is `true` even when no results — the operation succeeded, it just found nothing. Exit code 1 signals "no results" to agents.

## TDD cycles

### Cycle 1 — Successful lookup
RED: Load test fixture, run lookup with known term. Assert exit 0, envelope has results.
GREEN: Implement `lookupAction`.

### Cycle 2 — Not found → exit 1
RED: Lookup with nonexistent term. Assert exit 1, envelope has `results: []`.
GREEN: Return a recoverable exit-1 error when no results.

### Cycle 3 — No TBX path → exit 2
RED: Run lookup without `--tbx`. Assert exit 2 and error envelope with `no_tbx_path`.
GREEN: Check path before proceeding.

### Cycle 4 — --lang filter
RED: Lookup with `--lang he`. Assert only Hebrew results returned.
GREEN: Pass lang to `Glossary.Lookup`.

### Cycle 5 — --fields projection
RED: Lookup with `--fields concept_id`. Assert output contains only `concept_id` in results.
GREEN: Wire `ValidateFields` + `ProjectFields`.

## Acceptance

- `make build` passes
- Stub replaced with real action
- Exit codes: 0 (found), 1 (not found), 2 (usage error)
- `--lang` filtering works
- `--fields` projection works
- `ok: true` in all non-error envelopes

## Notes

**2026-05-26T00:36:17Z**

Implemented lookupAction in commands/lookup.go. Key findings: (1) urfave v3 StringArg values must be accessed via cmd.StringArg(name), not cmd.Args().First() — Args() returns unmatched args only. (2) Used the same exit-code pattern as validate: a lookupNotFoundError type implementing terr.Coded with ExitCode()=1. (3) buildLookupResult converts tbx.Concept to output.LookupResult, mapping StatusPreferred/StatusUnspecified to Preferred and StatusAdmitted to Admitted (deprecated terms omitted from lean output). (4) Field projection uses ValidateFields+ProjectFields on the marshaled JSON. (5) Updated TestUnderConstruction_ReturnsCoded to use scan instead of lookup since lookup is no longer a stub. (6) Copied rich-dct.tbx fixture to app testdata. 10 integration tests cover: success, not-found exit 1, no-tbx exit 2, lang filter, lang filter no match, fields projection, invalid field exit 2, admitted terms in result, case-fold matching, missing term.
