---
id: ter-s7xa
status: closed
deps: [ter-ppn9]
links: []
created: 2026-05-26T17:24:11Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-7fyo
tags: [e6, task, markdown, shared]
---
# E6.T1 — Frontmatter language extraction (shared)

## Goal

Move the YAML frontmatter `lang:` parser from `internal/extract` to `internal/markdown` as an exported function. Both `extract` and the new `check`/`scan` language resolution need frontmatter parsing — having it in `internal/markdown` (where `Spans` already lives) avoids duplication.

## Refs

- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md) §"Language resolution"
- Existing impl: `src/internal/extract/lang.go` — `frontmatterLang()` (unexported, hand-rolled YAML subset)

## New dependency

Add `gopkg.in/yaml.v3` — replaces the hand-rolled frontmatter parser with proper YAML unmarshaling. Handles quoted values, comments, and edge cases the current parser silently breaks on.

## Files to create / modify

- `src/go.mod` — add `gopkg.in/yaml.v3`
- `src/internal/markdown/frontmatter.go` — new file, export `FrontmatterLang(src []byte) string`
- `src/internal/markdown/frontmatter_test.go` — tests
- `src/internal/extract/lang.go` — update `DetectLang` to call `markdown.FrontmatterLang`, delete private `frontmatterLang`
- `src/internal/extract/lang_test.go` — keep tests that exercise `DetectLang` behavior; frontmatter-specific tests move to markdown package

## Behavior

```go
package markdown

func FrontmatterLang(src []byte) string
```

Implementation:
1. Check for `---\n` prefix.
2. Find closing `\n---`.
3. Extract the block between delimiters.
4. `yaml.Unmarshal` into `struct{ Lang string \x60yaml:"lang"\x60 }`.
5. Return `Lang` (empty string if no frontmatter, no lang key, or parse error).

## TDD cycles

### Cycle 1 — Basic extraction
RED: `FrontmatterLang([]byte("---\nlang: es\n---\nHello"))` → `"es"`.
GREEN: Implement with yaml.v3 unmarshal.

### Cycle 2 — No frontmatter
RED: `FrontmatterLang([]byte("Hello world"))` → `""`.
GREEN: Guard on prefix check.

### Cycle 3 — Unclosed frontmatter
RED: `FrontmatterLang([]byte("---\nlang: de\nHello"))` → `""`.
GREEN: Guard on closing delimiter.

### Cycle 4 — Quoted value
RED: `FrontmatterLang([]byte("---\nlang: \"pt-BR\"\n---\n"))` → `"pt-BR"`.
GREEN: Already handled by yaml.v3.

### Cycle 5 — Extract still works
RED: Existing `extract.DetectLang` tests pass after refactoring.
GREEN: Call `markdown.FrontmatterLang` from `DetectLang`.

## Acceptance

- `make build` passes
- `gopkg.in/yaml.v3` added to `go.mod`
- `FrontmatterLang` exported from `internal/markdown`
- `extract.DetectLang` delegates to `markdown.FrontmatterLang`
- No duplicate frontmatter parsing logic
- All existing extract tests pass unchanged
- Handles quoted values, inline comments, and other YAML edge cases

## Notes

**2026-05-26T18:09:44Z**

Implemented FrontmatterLang in internal/markdown/frontmatter.go using gopkg.in/yaml.v3. Refactored extract.DetectLang to delegate to markdown.FrontmatterLang, removing the hand-rolled parser. All 10 existing extract tests pass unchanged. 12 new frontmatter-specific tests in the markdown package cover: basic extraction, no frontmatter, unclosed frontmatter, quoted/single-quoted values, inline YAML comments, other keys, empty value, empty/nil input, Windows line endings.
