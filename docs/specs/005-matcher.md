# E5 — Matcher

> **Status**: APPROVED. The algorithmic core: given a glossary and a
> markdown blob, find every occurrence of a glossary term. Powers
> [E6 `scan`/`check`](006-scan-check.md). Does **not** power `lookup`
> (direct surface-form match) or `extract` (corpus-side heuristics, no
> glossary input).

## Architecture

Multi-pattern matching via **Aho-Corasick**, scanning a canonical
pre-normalized form of the corpus. One pass over the corpus matches
every glossary term simultaneously — orders-of-magnitude faster than
N independent regex scans at the cost of one new dependency and an
explicit position-mapping layer.

The pipeline:

```
corpus markdown
    │
    ▼  internal/markdown (goldmark, from E4)
plain-text spans (code regions stripped, line/col preserved)
    │
    ▼  internal/match/normalize
canonical bytes + offset map  ──────────►  []int : canonical → original
    │
    ▼  internal/match (Aho-Corasick)
[]rawMatch{patternID, canonicalOffset}
    │
    ▼  boundary check on the *original* text at mapped offset
    ▼  longest-match-at-same-start filter
[]Match{ConceptID, Term, Lang, Status, Line, Column, Context}
```

`Matcher` is reusable across files. Patterns are compiled once when
the matcher is constructed; per-file work is `Scan(text) → []Match`.

### Why Aho-Corasick from v1

cli-design.md §`internal/scan` originally committed to stdlib `regexp`
with one pre-compiled regex per term. That scales as `O(N·M)` (terms ×
corpus). For the realistic upper bound — ~5000 terms × ~5 MB
markdown × ~100 files — regex-per-term has a meaningfully worse
constant factor than a single Aho-Corasick scan. Locking the
algorithm now avoids a v1.x migration once the corpus grows.

cli-design.md's "stdlib `regexp` for matching" line is superseded by
this spec.

## v1 match policy

The baseline policy is locked. All flags below describe defaults;
per-language overrides live in `internal/match/policy.go`:

| Axis           | v1 default                                              |
| -------------- | ------------------------------------------------------- |
| Case           | Unicode default case-fold (Hebrew unchanged)            |
| Diacritics     | **Strict** — `razón` matches `razón` only, not `razon`  |
| Niqqud         | **Strip** from both sides before compare                |
| Normalization  | **NFC** on both sides                                   |
| Multi-word     | Collapse `\s+` runs; match across line breaks           |
| Lemmatization  | None                                                    |
| Inflection     | None — glossary declares variants explicitly            |
| Code regions   | Skipped (fenced + inline) via shared `internal/markdown`|
| Block quotes   | Scanned (real prose)                                    |
| Emphasis       | Scanned                                                 |

Rationale for the diacritics-strict default: a glossary that lists
`razón` does so on purpose; silently matching `razon` rewards sloppy
source. Users can later flip via `--fold-diacritics` on `scan`/`check`
if friction emerges (no flag in v1).

### Per-language overrides

```go
// internal/match/policy.go
package match

type Policy struct {
    CaseFold    bool
    FoldDiacritics bool
    StripNiqqud bool
    Normalize   Form // NFC | NFKD
}

var baseline = Policy{CaseFold: true, FoldDiacritics: false, StripNiqqud: false, Normalize: NFC}

var byLanguage = map[string]Policy{
    "he": baseline.with(StripNiqqud(true)),
    // future per-language exceptions land here
}
```

Hardcoded for v1. Config file is a v2 concern; reading policy from
TBX `tbxHeader.fileDesc` is rejected (couples policy to data, every
glossary author would have to learn match semantics).

## Normalization & position mapping

`internal/match/normalize.go` is the load-bearing piece. It transforms
the input text into a canonical byte stream and emits an offset map
so that any byte position in the canonical form can be translated
back to the original.

```go
type Canonical struct {
    Bytes []byte // canonical form: NFC, case-folded, niqqud-stripped, \s+→single space
    Map   []int  // Map[i] = original-text byte offset of canonical byte i
}

func Normalize(src []byte, p Policy) Canonical
```

Properties:

- `len(Map) == len(Bytes)` (per-byte mapping).
- `Map` is monotonically non-decreasing.
- When `\s+` collapses to one space, all original whitespace bytes
  map back to the *first* whitespace byte of the run (so a match
  beginning at the collapsed space is reported at the run's start).

The Aho-Corasick automaton scans `Bytes`; matches carry canonical
offsets; the consumer translates to original offsets via `Map` to
report line/column.

### Boundary check (post-filter)

Aho-Corasick matches substrings. To enforce the word-boundary
contract (`(^|[^\p{L}\p{N}])TERM([^\p{L}\p{N}]|$)`), every raw match
is re-validated against the **original text** at the mapped position:

```go
func validBoundary(orig []byte, start, end int) bool {
    if start > 0 && isLetterOrNumber(rune(orig[start-1])) { return false }
    if end < len(orig) && isLetterOrNumber(rune(orig[end])) { return false }
    return true
}
```

Validating against the original (not the canonical) is important —
diacritics and niqqud are word characters; stripping them in
canonical form would otherwise produce spurious boundary failures.

### Longest-match-at-same-start

Two terms can overlap (`tzimtzum` and `tzimtzum primordial`). When
multiple matches share a start position, the **longest one wins** and
the shorter ones are suppressed. Standard tokenizer convention; the
longer phrase is more specific. Documented in `Match` semantics.

Implementation: collect all raw matches, group by start position,
emit only the longest per group.

## API

```go
package match

type Matcher struct {
    glossary *tbx.Glossary
    lang     string         // "" = all languages
    policy   Policy
    machine  *ahocorasick.Matcher
    patterns []termPattern  // patternID → {ConceptID, Term, Lang, Status}
}

type Match struct {
    ConceptID string
    Term      string
    Lang      string
    Status    string // "preferred" | "admitted" | "deprecated" | "superseded" | "unspecified"
    Line      int
    Column    int
    Context   string
}

func New(g *tbx.Glossary, lang string, policy Policy) (*Matcher, error)
func (m *Matcher) Scan(text []byte) []Match
```

### Variants and status tagging

The matcher **compiles patterns for every term variant** in every
`<termSec>` — preferred, admitted, deprecated, superseded,
unspecified. Each pattern carries its source status as metadata, and
every emitted `Match` is tagged with that status.

Consumer logic (E6 `check`) uses the tag to decide: a `preferred`
match is informational, an `admitted` match is informational, a
`deprecated` match warrants a warning, a `superseded` match in
contemporary prose is a violation. Centralizing match production
in this epic and leaving policy decisions to consumers keeps the
matcher dialect-neutral.

## Dependencies introduced

- **`github.com/cloudflare/ahocorasick`** — BSD-3, byte-oriented,
  production-tested. The matcher operates on canonical UTF-8 bytes
  after normalization, so byte-level matching is exactly the right
  shape.

Project dependency list now:

- `urfave/cli/v3`
- `golang.org/x/text`
- `golang.org/x/term`
- `github.com/rogpeppe/go-internal/lockedfile` (E2)
- `github.com/yuin/goldmark` (E4)
- `github.com/cloudflare/ahocorasick` (this epic)

All single-purpose libraries with stable APIs and permissive licenses.

## Performance

Pre-compile patterns once; per-file cost is one normalize pass + one
Aho-Corasick scan over the canonical bytes + boundary checks
proportional to match count. Expected:

- 200 terms × 100 KB markdown: sub-millisecond per file.
- 5000 terms × 5 MB markdown: tens of milliseconds per file.
- 5000 terms × 5 MB × 100 files: low single-digit seconds total.

A perf-budget test (`//go:build perf`) is committed under
[testing](../adr/testing.md) §"Perf budget"
with a 500 ms target on a representative 1k-term × 1 MB fixture.
Failure fails the `make perf` job.

## CJK / Thai

Out of scope for v1. The `\p{L}\p{N}` boundary contract requires
whitespace-tokenized languages. CJK and Thai need a segmenter
(ICU-based or per-language); deferring keeps the matcher's contract
simple and avoids a heavyweight ICU dependency.

Documented in cli-design.md §"Supported languages" and at the top of
the `internal/match` package doc. A user pointing the binary at a CJK
corpus gets useful results within whitespace-delimited segments
(loanwords, names) and silently misses run-on text — acceptable for
v1 given the project's actual language matrix (English, Spanish,
Hebrew).

## Hand-offs

- Compiled matcher consumed by [E6](006-scan-check.md).
- Shared `internal/markdown` (goldmark-backed plain-text spans) from
  [E4](004-read-commands.md).
- Picklist values for admin-status tags shared with
  [E3](003-validate-command.md) via `internal/tbx/picklist.go`.
- Perf budget: [testing](../adr/testing.md).
- cli-design.md §`internal/scan` (stdlib regexp note) is superseded
  by this spec — needs a follow-up edit.
