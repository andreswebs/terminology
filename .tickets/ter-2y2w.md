---
id: ter-2y2w
status: closed
deps: [ter-c4ra, ter-i5cp]
links: []
created: 2026-05-26T13:49:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, task, match, normalize]
---
# E5.T2 — Canonical normalization + offset map

## Goal

Implement the normalization layer that transforms input text into a canonical byte stream for Aho-Corasick scanning. The `Normalize` function applies NFC normalization, Unicode case-folding, niqqud stripping, and whitespace collapsing — all driven by the `Policy` from T1. It emits a per-byte offset map so that any position in the canonical form can be translated back to the original text.

## Refs

- E5 spec: [docs/specs/005-matcher.md](docs/specs/005-matcher.md) §"Normalization & position mapping"
- `golang.org/x/text/unicode/norm` — NFC normalization
- `golang.org/x/text/cases` — Unicode case-folding

## Files to create / modify

- `src/internal/match/normalize.go` — `Normalize` function, `Canonical` type
- `src/internal/match/normalize_test.go` — tests

## Behavior

```go
type Canonical struct {
    Bytes []byte // canonical form
    Map   []int  // Map[i] = original-text byte offset of canonical byte i
}

func Normalize(src []byte, p Policy) Canonical
```

### Properties

- `len(Map) == len(Bytes)` — per-byte mapping.
- `Map` is monotonically non-decreasing.
- When `\s+` collapses to one space, all original whitespace bytes map back to the **first** whitespace byte of the run.
- NFC normalization applied first, then case-fold, then niqqud strip, then whitespace collapse.

### Niqqud stripping

Hebrew niqqud (vowel points, U+0591–U+05C7 combining marks) are stripped when `Policy.StripNiqqud` is true. Both the input text and pattern text go through the same normalization.

### Whitespace collapsing

Any run of `\s+` (Unicode whitespace) collapses to a single ASCII space (`0x20`). This enables multi-word term matching across line breaks.

## TDD cycles

### Cycle 1 — NFC normalization
RED: Input `"café"` (decomposed é) → canonical bytes are NFC form.
GREEN: Apply `norm.NFC.Bytes()`.

### Cycle 2 — Case-folding
RED: Input `"Tzimtzum"` with `CaseFold: true` → `"tzimtzum"`.
GREEN: Apply `cases.Fold()`.

### Cycle 3 — Offset map correctness
RED: Input `"ABC"` → Map `[0,1,2]` for canonical `"abc"`.
GREEN: Build per-byte map during transformation.

### Cycle 4 — Whitespace collapse
RED: Input `"hello  \n  world"` → canonical `"hello world"`, collapsed whitespace maps back to first space byte.
GREEN: Collapse whitespace runs, map to first byte of run.

### Cycle 5 — Niqqud stripping
RED: Input `"שָׁלוֹם"` (with niqqud) with `StripNiqqud: true` → canonical without combining marks.
GREEN: Strip Unicode range U+0591–U+05C7.

### Cycle 6 — Combined pipeline
RED: Hebrew text with niqqud + mixed case + whitespace → all transformations applied, map still valid.
GREEN: Chain all transformations with correct offset tracking.

### Cycle 7 — Empty input
RED: Empty input → empty Canonical with empty Map.
GREEN: Handle zero-length input.

## Acceptance

- `make build` passes
- Canonical form is NFC + case-folded + niqqud-stripped + whitespace-collapsed
- Offset map has per-byte granularity, monotonically non-decreasing
- Whitespace collapse maps to first byte of the run
- Pipeline order: NFC → case-fold → niqqud strip → whitespace collapse


## Notes

**2026-05-26T14:17:19Z**

Implemented Normalize() in internal/match/normalize.go with per-byte offset map. Pipeline: applyNormForm (NFC/NFKD via segment-based approach using norm.Properties.BoundaryBefore) → case-fold → niqqud strip → diacritics strip → whitespace collapse. Key decisions: (1) applyNormForm walks source by normalization segments (starter + combining marks) using BoundaryBefore(), maps all output bytes to segment start offset — correct and precise for single-rune segments (common case). (2) FoldDiacritics forces NFD to decompose composed chars before stripping Mn category marks. (3) cases.Fold().Bytes() applied per-rune for correct offset tracking (ß→ss maps both output bytes to ß source offset). 15 tests covering all TDD cycles plus edge cases (ß expansion, NFKD ligature decomposition, monotonicity property, FoldDiacritics).
