---
id: ter-vzki
status: closed
deps: [ter-ab56]
links: []
created: 2026-05-25T19:37:26Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, task, extract, heuristic]
---
# E4.T11 — Extract: capitalized phrases heuristic

## Goal

Create `internal/extract` package and implement the first heuristic: capitalized phrase detection. This establishes the package structure, shared types, and aggregation logic that T12 and T13 build on.

## Refs

- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md) §"extract — heuristic engine" heuristic 1

## Files to create / modify

- `src/internal/extract/extract.go` — shared types and aggregation
- `src/internal/extract/capitalized.go` — capitalized phrases heuristic
- `src/internal/extract/capitalized_test.go` — tests

## Behavior

### Shared types

```go
package extract

type Candidate struct {
    Term      string
    Frequency int
    Heuristic string
    Locations []Location
}

type Location struct {
    File   string
    Line   int
    Col    int
    Offset int
}

type Options struct {
    MinFreq   int
    Script    string
    Stopwords map[string]bool
    Exclude   map[string]bool
    Lang      string
}
```

### Capitalized phrases heuristic

Detect sequences of capitalized words that are not at sentence start. A "capitalized phrase" is one or more consecutive words where each word starts with an uppercase letter, appearing mid-sentence (not after `.`, `!`, `?`, or at the start of a paragraph).

Input: plain-text spans (from `internal/markdown` or raw text).

```go
func CapitalizedPhrases(spans []Span, lang string) []Candidate
```

The `lang` parameter allows per-language tuning (e.g., German capitalizes all nouns — may need different rules). For v1, use English-centric rules; document that German/other languages may produce false positives.

### Aggregation

Candidates are aggregated across all input spans:
- Same surface form → increment frequency, append location
- Surface forms are NFC-normalized before aggregation

## TDD cycles

### Cycle 1 — Single capitalized word mid-sentence
RED: Input `"the Holy Temple was destroyed"` → candidate `"Holy Temple"` with frequency 1.
GREEN: Implement tokenizer + capitalization check.

### Cycle 2 — Sentence-start words excluded
RED: Input `"The cat sat. The dog ran."` → no candidates (all "The" are sentence-start).
GREEN: Track sentence boundaries.

### Cycle 3 — Multi-word capitalized phrases
RED: Input `"the Dead Sea Scrolls were found"` → candidate `"Dead Sea Scrolls"`.
GREEN: Aggregate consecutive capitalized tokens.

### Cycle 4 — Frequency aggregation
RED: Input mentioning `"Holy Temple"` three times → frequency 3.
GREEN: Aggregate by normalized surface form.

### Cycle 5 — NFC normalization in aggregation
RED: Same phrase in NFC and NFD forms → aggregated as one candidate.
GREEN: Normalize before map lookup.

## Acceptance

- `make build` passes
- Package `internal/extract` created with shared types
- Capitalized phrases detected mid-sentence
- Sentence-start words excluded
- Frequency aggregation works
- Locations tracked per occurrence

## Notes

**2026-05-26T00:43:30Z**

Implemented internal/extract package with shared types (Candidate, Location, Span) in extract.go and capitalized phrases heuristic in capitalized.go. Key design decisions: (1) Each span is treated as a paragraph start (afterSentenceStart=true at the start of each span), matching goldmark's text node behavior. (2) Sentence boundaries detected by trailing '.', '!', '?' on word tokens — both in separators and words. (3) Trailing punctuation stripped from phrase terms via stripTrailingPunct. (4) NFC normalization applied before aggregation map lookup. (5) Tokenizer splits text into word/separator token pairs with a single-pass loop. 12 tests covering: single/multi-word phrases, sentence-start exclusion, paragraph-start exclusion, frequency aggregation, NFC normalization, location tracking, empty input, all-lowercase, mixed sentences, exclamation-point boundaries, and multiple distinct phrases.
