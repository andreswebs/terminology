---
id: ter-bmuk
status: closed
deps: [ter-c4ra, ter-2y2w, ter-tc1n, ter-6tef, ter-j3y2]
links: []
created: 2026-05-26T13:49:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, task, match, api]
---
# E5.T6 — Matcher API (New + Scan)

## Goal

Implement the top-level `Matcher` that assembles the full pipeline: compile glossary terms into an Aho-Corasick automaton (T3), scan text through canonical normalization (T2) + AC search + boundary check (T4) + longest-match filter (T5), and emit `[]Match` with concept IDs, term surfaces, language, status, line, column, and context.

## Refs

- E5 spec: [docs/specs/005-matcher.md](docs/specs/005-matcher.md) §"API"
- E4 `internal/markdown` — provides `Span` type for line/col

## Files to create / modify

- `src/internal/match/match.go` — `Matcher`, `Match`, `New`, `Scan`
- `src/internal/match/match_test.go` — tests

## Behavior

```go
type Matcher struct {
    lang     string
    policy   Policy
    machine  *automaton
    patterns []termPattern
}

type termPattern struct {
    ConceptID string
    Term      string
    Lang      string
    Status    string // "preferred" | "admitted" | "deprecated" | "superseded" | "unspecified"
}

type Match struct {
    ConceptID string
    Term      string
    Lang      string
    Status    string
    Line      int
    Column    int
    Context   string
}

func New(g *tbx.Glossary, lang string) (*Matcher, error)
func (m *Matcher) Scan(text []byte, spans []markdown.Span) []Match
```

### Pattern compilation (`New`)

1. Iterate all concepts in the glossary.
2. For each concept's language sections (filtered by `lang` if non-empty), iterate all terms.
3. For each term, normalize its surface via `Normalize(surface, PolicyFor(lang))` to get canonical bytes.
4. Build a `termPattern` with concept ID, original surface, language, and admin status string.
5. Build the AC automaton from all canonical pattern bytes.

### Scanning (`Scan`)

1. For each span, normalize the span's text via `Normalize`.
2. Search the canonical bytes with the AC automaton.
3. Map raw match positions back to original text via `Canonical.Map`.
4. Validate each raw match with `validBoundary` on the original text.
5. Apply `longestMatchPerStart` filter.
6. Translate byte offsets to line/column using the span's position metadata.
7. Extract a context window around each match.
8. Return sorted `[]Match` by `(Line, Column)`.

### Status tagging

Every emitted `Match` carries the status of the matched term (`preferred`, `admitted`, `deprecated`, `superseded`, `unspecified`). The matcher is dialect-neutral — consumer logic (E6 `check`) decides what to do with each status.

## TDD cycles

### Cycle 1 — Single term match
RED: Glossary with one concept `"tzimtzum"`, text `"the concept of tzimtzum in Kabbalah"` → one Match with correct line/col.
GREEN: Full pipeline: New → Scan.

### Cycle 2 — Case-insensitive match
RED: Glossary term `"Tzimtzum"`, text `"TZIMTZUM appears here"` → match found.
GREEN: Both sides case-folded via normalization.

### Cycle 3 — Hebrew term with niqqud
RED: Glossary term `"שָׁלוֹם"` (with niqqud), text has `"שלום"` (without) → match found.
GREEN: Niqqud stripped from both sides for Hebrew policy.

### Cycle 4 — Multi-word term
RED: Glossary term `"tzimtzum primordial"`, text `"the tzimtzum primordial concept"` → one match.
GREEN: Whitespace-collapsed matching.

### Cycle 5 — Status tagging
RED: Glossary with deprecated variant `"contraction"` → Match has `Status: "deprecated"`.
GREEN: termPattern carries status through to Match.

### Cycle 6 — Language filter
RED: Glossary with `en` and `he` terms, `lang="he"` → only Hebrew patterns compiled.
GREEN: Filter language sections in New.

### Cycle 7 — Longest match wins
RED: Glossary has `"tzimtzum"` and `"tzimtzum primordial"`, text has `"tzimtzum primordial"` → only the longer match emitted.
GREEN: longestMatchPerStart filter applied.

### Cycle 8 — Context window
RED: Match at column 20 in a 100-char line → Context includes surrounding text.
GREEN: Extract context window from original text.

### Cycle 9 — No matches
RED: Glossary terms not in text → empty `[]Match`.
GREEN: Return nil/empty.

### Cycle 10 — Multiple spans
RED: Text split across two markdown spans → matches found in both, correct line/col.
GREEN: Scan iterates spans, accumulates matches.

## Acceptance

- `make build` passes
- Matcher compiles patterns for all term variants (all statuses)
- Scan produces correctly positioned matches with concept ID, term, lang, status
- Case-folding, niqqud stripping, whitespace collapse all work end-to-end
- Longest-match-at-same-start applied
- Matches sorted by (line, column)
- Context window extracted


## Notes

**2026-05-26T14:31:45Z**

Implemented Matcher API (New + Scan) in src/internal/match/match.go. New(g, lang, policy) follows the spec signature (explicit policy parameter, not the ticket's 2-param version). Pattern compilation uses the caller-supplied policy for both patterns and text normalization — this ensures AC matching works since all patterns and text must use the same normalization. origEnd uses utf8.DecodeRune to correctly advance past multi-byte runes when mapping canonical offsets back to original text. 15 tests cover: single term, case-insensitive, Hebrew niqqud, multi-word, status tagging (all 5 statuses), language filter, longest-match-wins, context window, no matches, multiple spans, empty glossary, boundary rejection, sorted output, and Hebrew term boundaries.
