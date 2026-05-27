---
id: ter-wx2k
status: closed
deps: [ter-s7xa, ter-qxpp]
links: []
created: 2026-05-26T17:24:44Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-7fyo
tags: [e6, task, app, scan]
---
# E6.T8 — Scan frontmatter language resolution

## Goal

Update the `scan` command to detect the file's language from YAML frontmatter before falling back to the `--lang` flag. Unlike `check`, scan does not error when no language is specified — it scans all languages as before.

## Refs

- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md) §"Language resolution"
- Current scan: `src/internal/app/commands/scan.go`
- Frontmatter extraction: `internal/markdown` (T1)

## Files to create / modify

- `src/internal/app/commands/scan.go` — add frontmatter lang detection before `--lang` flag
- Possibly update scan golden test files if behavior changes (e.g. a fixture with frontmatter)

## Behavior

Current precedence: `--lang` flag → scan all languages.

New precedence: `markdown.FrontmatterLang(data)` → `--lang` flag → scan all languages (no error).

```go
lang := markdown.FrontmatterLang(data)
if lang == "" {
    lang = cmd.String("lang")
}
// lang="" means scan all languages (existing behavior)
```

### Impact on existing tests

If existing scan test fixtures don't have frontmatter, behavior is unchanged. If any fixture does have `lang:` in frontmatter, its matches would be restricted to that language — check and update golden files if needed.

### Edge case: frontmatter vs --lang conflict

Frontmatter wins per spec. If a file has `lang: es` and the user passes `--lang he`, the scan uses `es`. This is documented behavior — frontmatter is the strongest signal.

## TDD cycles

### Cycle 1 — Frontmatter detected
RED: Scan a file with `lang: he` frontmatter, no `--lang` flag → only Hebrew matches.
GREEN: Read frontmatter before flag.

### Cycle 2 — No frontmatter, flag used
RED: File without frontmatter, `--lang he` → only Hebrew matches (existing behavior).
GREEN: Fall through to flag.

### Cycle 3 — No frontmatter, no flag
RED: File without frontmatter, no flag → all-language scan (existing behavior).
GREEN: Empty lang passes through.

### Cycle 4 — Frontmatter overrides flag
RED: File has `lang: es`, `--lang he` passed → Spanish matches only.
GREEN: Frontmatter checked first.

## Acceptance

- `make build` passes
- Scan uses frontmatter `lang:` when present
- Falls back to `--lang` flag when no frontmatter
- Scans all languages when neither is specified (no error)
- Existing scan golden tests still pass (or updated if fixture had frontmatter)

## Notes

**2026-05-26T18:34:18Z**

Added frontmatter language resolution to scan command. Precedence: markdown.FrontmatterLang(data) → --lang flag → scan all languages. Follows the same pattern as check.go (E6.T7). Two new tests: TestScan_FrontmatterLang (frontmatter detected, restricts to that language) and TestScan_FrontmatterOverridesFlag (frontmatter wins over --lang flag). New fixture: scan-frontmatter-he.md. No existing scan tests affected since no prior fixtures had frontmatter. make build passes clean.
