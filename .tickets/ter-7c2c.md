---
id: ter-7c2c
status: closed
deps: [ter-ab56, ter-vzki]
links: [ter-vri7]
created: 2026-05-25T19:37:26Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, extract, heuristic]
---
# E4.T12 — Extract: foreign-script tokens heuristic

## Goal

Implement the second extract heuristic: foreign-script token detection. Tokens whose dominant Unicode script differs from the surrounding paragraph's script are flagged as candidates. The motivating example is Hebrew words in a Spanish text.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"extract — heuristic engine" heuristic 2
- `golang.org/x/text/unicode` — script detection

## Files to create / modify

- `src/internal/extract/foreign.go` — foreign-script heuristic
- `src/internal/extract/foreign_test.go` — tests

## Behavior

```go
func ForeignScriptTokens(spans []Span, opts Options) []Candidate
```

### Algorithm

1. For each span, determine the **dominant script** by counting runes per Unicode script.
2. Tokenize the span into words.
3. For each token, determine its dominant script.
4. If the token's script differs from the span's dominant script, it's a foreign-script candidate.

### Script detection

Use `unicode.Scripts` (standard library) or `golang.org/x/text/unicode` to classify runes. The dominant script of a text segment is the script with the most runes (excluding Common/Inherited).

### `--script` filtering

When `opts.Script` is set (e.g. `"hebrew"`), only return candidates whose script matches. `"any"` returns all foreign-script tokens.

### Aggregation

Same as T11 — aggregate by NFC-normalized surface form, track frequency and locations.

## TDD cycles

### Cycle 1 — Hebrew in Latin text
RED: Input `"the concept of צמצום in Kabbalah"` → candidate `"צמצום"` with heuristic `"foreign_script"`.
GREEN: Implement script detection + comparison.

### Cycle 2 — Latin in Hebrew text
RED: Input in Hebrew with an English word → English word detected as foreign.
GREEN: Symmetric detection.

### Cycle 3 — --script filter
RED: With `opts.Script = "hebrew"`, only Hebrew-script candidates returned.
GREEN: Filter by target script.

### Cycle 4 — Common/Inherited characters ignored
RED: Punctuation and digits don't affect script detection.
GREEN: Skip Common/Inherited scripts in dominant-script calculation.

### Cycle 5 — Frequency aggregation
RED: Same foreign token appears twice → frequency 2.
GREEN: Aggregate by normalized form.

## Acceptance

- `make build` passes
- Foreign-script tokens detected by comparing token script to surrounding dominant script
- `--script` filter works
- Common/Inherited Unicode categories excluded from script detection
- Aggregation consistent with T11

## Notes

**2026-05-26T01:00:09Z**

Implemented ForeignScriptTokens in src/internal/extract/foreign.go with Options type (Script field for --script filtering). Uses unicode.RangeTable for script detection via dominantScript() helper that counts runes per script (excluding Common/Inherited). scriptByName() maps CLI picklist values (latin, hebrew, cyrillic, arabic) to unicode range tables. Reuses tokenize() and stripTrailingPunct() from capitalized.go. Aggregation follows the same NFC-normalized map pattern as CapitalizedPhrases. 9 tests in foreign_test.go covering: Hebrew-in-Latin, Latin-in-Hebrew, --script filter, --script any, Common/Inherited ignored, frequency aggregation, empty input, all-same-script, location tracking.
