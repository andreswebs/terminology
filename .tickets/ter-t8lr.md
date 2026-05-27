---
id: ter-t8lr
status: closed
deps: [ter-ab56, ter-vzki, ter-7c2c, ter-y8nn, ter-ph7i, ter-by7v, ter-255x]
links: []
created: 2026-05-25T19:37:27Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, extract, command]
---
# E4.T15 — Extract command action

## Goal

Replace the `underConstruction` stub in `commands/extract.go` with the real action. Wire the three heuristics (T11–T13), markdown spans (T3), language detection (T14), --fields projection (T2), --exclude filtering, and all flags.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"extract — heuristic engine", §"Flags"

## Files to create / modify

- `src/internal/app/commands/extract.go` — implement `extractAction`
- `src/internal/app/commands_test.go` — integration tests

## Behavior

### Flow

1. Read positional `FILE...` arguments.
2. Parse flags: `--exclude`, `--script`, `--lang`, `--stopwords`, `--min-freq`, `--fields`.
3. If `--stopwords` is set, load the file via `extract.LoadStopwords`.
4. If `--exclude` is set, load the TBX file and collect all term surface forms into an exclusion set.
5. For each input file:
   a. Read file contents.
   b. Detect language (frontmatter → `--lang` → default `en`).
   c. Parse markdown via `markdown.Spans()` to get plain-text spans.
   d. Run all three heuristics on the spans.
6. Aggregate candidates across all files (same term → merge frequency + locations).
7. Apply `--exclude` filter (remove candidates matching glossary terms).
8. Apply `--script` filter.
9. Convert to `output.ExtractEnvelope`.
10. If `--fields` is set, validate and project.
11. Emit JSON.

### Exit codes

| Code | Condition |
|------|-----------|
| 0    | Candidates found |
| 1    | No candidates found |
| 2    | Usage error (missing files, invalid flag) |

### No streaming

Per spec, single JSON envelope with full result list. No NDJSON streaming in v1.

## TDD cycles

### Cycle 1 — Basic extraction from markdown
RED: Run extract on a test markdown file. Assert candidates returned in envelope.
GREEN: Wire markdown parsing + heuristics.

### Cycle 2 — Code blocks excluded
RED: Markdown with code blocks — assert no code identifiers in candidates.
GREEN: Markdown spans filter works.

### Cycle 3 — --exclude filters glossary terms
RED: Extract with `--exclude` pointing at a TBX with term "tzimtzum". Assert "tzimtzum" not in candidates.
GREEN: Load TBX, build exclusion set, filter.

### Cycle 4 — --stopwords filtering
RED: Extract with `--stopwords` file containing common words. Assert those words absent from frequency candidates.
GREEN: Wire stopwords loading and filtering.

### Cycle 5 — --min-freq threshold
RED: `--min-freq 5` — only candidates with frequency >= 5 returned.
GREEN: Pass MinFreq to options.

### Cycle 6 — --script filter
RED: `--script hebrew` — only Hebrew-script candidates returned.
GREEN: Pass Script to options.

### Cycle 7 — --fields projection
RED: `--fields term,frequency` — output contains only those fields.
GREEN: Wire ValidateFields + ProjectFields.

### Cycle 8 — No candidates → exit 1
RED: Extract on file with no candidate terms. Assert exit 1.
GREEN: Return recoverable exit-1 error.

## Acceptance

- `make build` passes
- Stub replaced with real action
- All flags wired: `--exclude`, `--script`, `--lang`, `--stopwords`, `--min-freq`, `--fields`
- Code blocks excluded from analysis
- Aggregation across multiple files works
- Exit codes: 0 (found), 1 (none), 2 (usage error)

## Notes

**2026-05-26T01:16:01Z**

Implemented extractAction replacing the underConstruction stub. Wired all three heuristics (CapitalizedPhrases, ForeignScriptTokens, HighFrequencyTokens), markdown spans parsing, language detection, --exclude TBX filtering, --stopwords, --min-freq, --script filtering, and --fields projection. Exit codes: 0 (candidates found), 1 (no candidates), 2 (usage error), 3 (I/O error). Updated old stub tests to reflect the new implementation. Added 9 integration tests covering: basic extraction, code block exclusion, glossary term exclusion, stopwords filtering, min-freq threshold, script filtering, fields projection, no-candidates exit code, and nonexistent file handling. Golden test updated.
