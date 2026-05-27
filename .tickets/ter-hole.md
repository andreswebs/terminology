---
id: ter-hole
status: closed
deps: [ter-y57h, ter-vv38, ter-vg07]
links: []
created: 2026-05-26T17:24:39Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-7fyo
tags: [e6, task, app, check]
---
# E6.T7 — Check command action

## Goal

Replace the `underConstruction` stub in `check.go` with the real `checkAction`. Wire the full pipeline: TBX load → language resolution → file reads → check algorithm → envelope build → field projection → JSON emit → exit code.

## Refs

- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md)
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"terminology check SRC TGT"
- Scan action (reference): `src/internal/app/commands/scan.go`
- Error handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md)

## Files to create / modify

- `src/internal/app/commands/check.go` — replace `underConstruction` with `checkAction`
- `src/internal/app/stub_test.go` — update (remove check from stub expectations, similar to what E5 did for scan)

## Command flow

1. Resolve TBX path (`--tbx` → `TERMINOLOGY_TBX` → `ErrNoTBXPath`).
2. Load glossary via `tbx.Load(path)`.
3. Read SRC file (first positional arg).
4. Read TGT file (second positional arg).
5. Resolve source language: `markdown.FrontmatterLang(srcData)` → `--source-lang` flag → `ErrLanguageRequired`.
6. Resolve target language: `markdown.FrontmatterLang(tgtData)` → `--target-lang` flag → `ErrLanguageRequired`.
7. Run `check.Check(g, srcData, tgtData, srcLang, tgtLang, contextSize, strict)`.
8. Build `CheckEnvelope` from result.
9. Set `OK` = `len(violations) == 0`.
10. Apply `--fields` projection if specified.
11. Emit JSON to stdout.
12. Return `nil` if no violations; return a coded error (exit 1) if violations present.

### Language resolution

Per spec, the precedence for each file is:
1. YAML frontmatter `lang:` in the file itself → wins.
2. CLI flag (`--source-lang` for SRC, `--target-lang` for TGT) → used if no frontmatter.
3. Neither → return `ErrLanguageRequired` with a hint naming the file.

The hint should be file-specific: `"language not specified for SRC"`.

### Exit codes

- `0` — no violations.
- `1` — violations present (recoverable). Use a non-`terr.Coded` approach or a dedicated coded error.
- `2` — usage error (`no_tbx_path`, `language_required`, `invalid_field`).
- `3` — I/O error (file not readable).

### Error for exit 1

The check command needs to return exit 1 when violations exist, but this isn't an error in the `terr.Coded` sense — the envelope is the result, not an error envelope. Follow the pattern from `validate` which uses a `warningsError` type.

## TDD cycles

### Cycle 1 — Basic clean check
RED: `terminology check src.md tgt.md --tbx glossary.tbx --source-lang es --target-lang he` → exit 0, ok true, violations empty.
GREEN: Wire full pipeline.

### Cycle 2 — Violations present
RED: TGT missing a preferred term → exit 1, ok false, violations non-empty.
GREEN: Return exit 1 on violations.

### Cycle 3 — Language from frontmatter
RED: SRC has `lang: es` in frontmatter, no `--source-lang` flag → language resolved from frontmatter.
GREEN: Wire `markdown.FrontmatterLang`.

### Cycle 4 — Language required error
RED: SRC has no frontmatter and no `--source-lang` → exit 2, `language_required`.
GREEN: Return `ErrLanguageRequired`.

### Cycle 5 — Field projection
RED: `--fields violations.concept_id,violations.type` → projected output.
GREEN: Wire field projection.

### Cycle 6 — File not found
RED: `terminology check nonexistent.md tgt.md --tbx glossary.tbx` → exit 3, io_error.
GREEN: Handle file read error.

## Acceptance

- `make build` passes
- Check command replaces `underConstruction` stub
- Full pipeline: TBX load → language resolution → check → envelope → JSON
- Exit 0 for clean, exit 1 for violations
- Language resolution: frontmatter > flag > ErrLanguageRequired
- `--strict`, `--context`, `--fields` flags work
- Error cases emit correct envelopes and exit codes
- Stub test updated (check removed from stub list)

## Notes

**2026-05-26T18:31:53Z**

Replaced underConstruction stub with real checkAction. Full pipeline: TBX load → frontmatter lang resolution → file reads → check algorithm → CheckEnvelope build → field projection → JSON emit → exit code. ErrLanguageRequired moved from app/errors.go to commands/check.go to break the app→commands→app import cycle. Tests cover: clean check (exit 0), violations present (exit 1), forbidden_variant detection, frontmatter language auto-detection, language_required error (exit 2), no_tbx_path (exit 2), file not found (exit 3), and --fields projection. Schema golden files regenerated for the hint text change. stub_test.go updated to use concept add instead of check.
