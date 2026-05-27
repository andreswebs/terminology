---
id: ter-qxpp
status: closed
deps: [ter-ppn9]
links: []
created: 2026-05-26T17:24:15Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-7fyo
tags: [e6, task, terr, error]
---
# E6.T2 — ErrLanguageRequired sentinel

## Goal

Add the `ErrLanguageRequired` sentinel error for `check` command language resolution. When neither frontmatter nor a CLI flag provides a language, this error is returned with a hint enumerating both ways to supply it.

## Refs

- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md) §"Language resolution"
- Error handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md)
- Existing sentinels: `src/internal/tbx/errors.go`

## Files to create / modify

- `src/internal/app/errors.go` — add `ErrLanguageRequired` sentinel (this file already exists with other app-level errors)

## Behavior

```go
var ErrLanguageRequired = terr.New(
    "language_required", 2,
    "pass --source-lang/--target-lang or add 'lang: LANG' to frontmatter",
    "language not specified",
)
```

- Code: `language_required`
- Exit: `2` (usage error)
- Hint enumerates both resolution paths (flag and frontmatter)

The spec shows a per-file hint: `"check: language not specified for SRC"`. The sentinel provides the generic message; the command action wraps it with the file-specific context.

## TDD cycles

### Cycle 1 — Sentinel exists
RED: Test that `ErrLanguageRequired` satisfies `terr.Coded` with code `"language_required"` and exit `2`.
GREEN: Add the sentinel.

### Cycle 2 — Wrapping with context
RED: `ErrLanguageRequired.Wrap(fmt.Errorf("for SRC"))` preserves code and exit code.
GREEN: Verify `terr.E.Wrap` behavior (already implemented in terr).

## Acceptance

- `make build` passes
- `ErrLanguageRequired` is a `terr.Coded` with code `language_required`, exit `2`
- Sentinel appears in `terminology schema` error registry
- Hint mentions both flag and frontmatter resolution paths

## Notes

**2026-05-26T18:12:40Z**

Added ErrLanguageRequired sentinel in src/internal/app/errors.go with code 'language_required', exit 2, and hint mentioning --lang/--source-lang/--target-lang and frontmatter. Tests in errors_test.go cover: Coded interface satisfaction, hint content, and Wrap preserving code/exit. Registry test updated to include the new sentinel. Schema golden files regenerated (schema/full, schema/command_filter) since the new sentinel appears in the error_codes output.
