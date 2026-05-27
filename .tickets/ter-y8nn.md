---
id: ter-y8nn
status: closed
deps: [ter-ab56, ter-vzki]
links: []
created: 2026-05-25T19:37:26Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, extract, heuristic]
---
# E4.T13 — Extract: high-frequency tokens + stoplist

## Goal

Implement the third extract heuristic: high-frequency token detection with optional stoplist filtering. Tokens that appear at least `--min-freq` times (default 3) across the corpus are flagged as candidates. When `--stopwords PATH` is provided, those words are excluded.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"extract — heuristic engine" heuristic 3, §"Stoplist policy"

## Files to create / modify

- `src/internal/extract/frequency.go` — high-frequency heuristic
- `src/internal/extract/frequency_test.go` — tests
- `src/internal/extract/stopwords.go` — stoplist loader
- `src/internal/extract/stopwords_test.go` — tests

## Behavior

```go
func HighFrequencyTokens(spans []Span, opts Options) []Candidate
func LoadStopwords(path string) (map[string]bool, error)
```

### High-frequency heuristic

1. Tokenize all spans into words.
2. NFC-normalize and case-fold each word.
3. Count frequency per unique form.
4. If `opts.Stopwords` is set, exclude matching words.
5. Return candidates with frequency >= `opts.MinFreq`.

### Stoplist

- File format: newline-separated, one word per line.
- Lines starting with `#` are comments.
- Empty lines ignored.
- Words are NFC-normalized and case-folded before comparison.
- Stoplist applies ONLY to this heuristic (not capitalized phrases or foreign-script).

### No bundled stoplists

Per spec, no stoplists ship with the binary. Users supply via `--stopwords PATH`.

## TDD cycles

### Cycle 1 — Frequency counting
RED: Input with word "temple" appearing 5 times, `MinFreq: 3`. Candidate "temple" returned with frequency 5.
GREEN: Implement tokenize + count + threshold.

### Cycle 2 — Below threshold excluded
RED: Word appearing 2 times with `MinFreq: 3` → not returned.
GREEN: Filter by threshold.

### Cycle 3 — Case-fold aggregation
RED: "Temple" and "temple" aggregate to frequency 2.
GREEN: Case-fold before counting.

### Cycle 4 — Stopwords exclusion
RED: "the" appears 10 times but is in stopwords → not returned.
GREEN: Check stopwords set before including.

### Cycle 5 — Load stopwords file
RED: File with `"the\na\nan\n"` → `LoadStopwords` returns map with 3 entries.
GREEN: Implement line reader with comment/empty-line handling.

### Cycle 6 — Stopwords case-folded
RED: Stopword file has "The", input has "the" → excluded.
GREEN: Normalize stopwords on load.

## Acceptance

- `make build` passes
- Frequency threshold gates output
- Stopwords file loaded and applied
- Only affects high-frequency heuristic (not T11/T12)
- No bundled stoplists

## Notes

**2026-05-26T01:05:21Z**

Implemented HighFrequencyTokens in frequency.go and LoadStopwords in stopwords.go. Extended the shared Options struct (in foreign.go) with MinFreq and Stopwords fields as planned in T12 learnings. HighFrequencyTokens tokenizes spans, NFC-normalizes + case-folds each word, counts frequency, filters by stopwords and min-freq threshold (default 3). LoadStopwords reads a newline-separated file, skipping comments (#) and empty lines, NFC-normalizing and case-folding each entry. Tests cover: basic frequency counting, below-threshold exclusion, case-fold aggregation, NFC normalization, multi-span aggregation, default MinFreq, stopwords exclusion, stopwords file parsing with comments/empty lines, stopwords case-folding, file-not-found error. The defer f.Close() pattern uses the project convention: defer func() { _ = f.Close() }().
