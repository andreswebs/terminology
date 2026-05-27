---
id: ter-ph7i
status: closed
deps: [ter-ab56]
links: [ter-vri7]
created: 2026-05-25T19:37:26Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, extract, output]
---
# E4.T14 — Extract envelope type + language detection

## Goal

Define the `ExtractEnvelope` type in `internal/output/types.go` and implement the language detection precedence logic (frontmatter → `--lang` → default `en`) for the extract command.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"extract", §"Language detection"

## Files to create / modify

- `src/internal/output/types.go` — add `ExtractEnvelope`, `ExtractCandidate`
- `src/internal/output/types_test.go` — JSON shape tests
- `src/internal/extract/lang.go` — language detection from frontmatter
- `src/internal/extract/lang_test.go` — tests

## Behavior

### Envelope shape

```json
{
  "schema_version": 1,
  "ok": true,
  "candidates": [{
    "term": "Holy Temple",
    "frequency": 5,
    "heuristic": "capitalized_phrase",
    "locations": [{"file": "ch1.md", "line": 12, "col": 5}]
  }]
}
```

Go types:

```go
type ExtractEnvelope struct {
    SchemaVersion int                `json:"schema_version"`
    OK            bool               `json:"ok"`
    Candidates    []ExtractCandidate `json:"candidates"`
}

type ExtractCandidate struct {
    Term      string            `json:"term"`
    Frequency int               `json:"frequency"`
    Heuristic string            `json:"heuristic"`
    Locations []ExtractLocation `json:"locations,omitempty"`
}

type ExtractLocation struct {
    File string `json:"file"`
    Line int    `json:"line,omitempty"`
    Col  int    `json:"col,omitempty"`
}
```

### Language detection precedence

1. **Markdown frontmatter**: If the file's YAML frontmatter contains `lang: LANG`, use that.
2. **`--lang` flag**: If passed, applies to files lacking frontmatter.
3. **Default `en`**: When neither is present.

```go
func DetectLang(src []byte, flagLang string) string
```

Parse YAML frontmatter (between `---` delimiters at file start), extract `lang` key. No external YAML library needed — a simple scanner suffices for this single-key extraction.

## TDD cycles

### Cycle 1 — Envelope JSON shape
RED: Marshal `ExtractEnvelope` with candidates. Assert JSON keys match spec.
GREEN: Define types with correct `json` tags.

### Cycle 2 — Empty candidates → empty array
RED: `ExtractEnvelope{Candidates: []ExtractCandidate{}}` → `"candidates":[]"`.
GREEN: Initialize with empty slice.

### Cycle 3 — Frontmatter lang detection
RED: Input `"---\nlang: es\n---\nHello"` → `DetectLang(src, "")` returns `"es"`.
GREEN: Implement frontmatter scanner.

### Cycle 4 — Flag fallback
RED: Input without frontmatter, `flagLang = "he"` → returns `"he"`.
GREEN: Fallback to flag.

### Cycle 5 — Default to "en"
RED: No frontmatter, empty flag → returns `"en"`.
GREEN: Final default.

## Acceptance

- `make build` passes
- Envelope shape matches spec
- Language detection precedence: frontmatter → flag → "en"
- No external YAML library

## Notes

**2026-05-26T00:50:02Z**

Implemented ExtractEnvelope, ExtractCandidate, ExtractLocation types in internal/output/types.go following existing patterns (MarshalJSON nil→[] coercion, omitempty on optional fields, RegisterEnvelope in init). Implemented DetectLang in internal/extract/lang.go with frontmatter→flag→en precedence. Simple frontmatter scanner handles both Unix and Windows line endings, ignores unclosed frontmatter blocks. No external YAML library needed.
