# E4 — Read commands: `lookup`, `schema`, `extract`

> **Status**: APPROVED. The read commands that don't depend on the
> matcher, plus the shared `internal/output` envelope helpers consumed
> by every read and write command.

## Scope

- `lookup` — direct surface-form search against the glossary; uses E2
  domain model.
- `schema` — **reflective** introspection over the live urfave command
  tree + `internal/output` struct types + `terr` sentinel registry;
  optionally filtered to one command via `--command NAME`.
- `extract` — corpus → candidate terms via heuristics (capitalized
  phrases, foreign-script tokens, high-frequency tokens). Does **not**
  use the matcher because it isn't glossary-driven.

`scan` and `check` ride on the matcher and live in
[E6](006-scan-check.md).

In scope this epic:

- `internal/output` — JSON / text envelope helpers, `--fields`
  projection, error envelope. Shared across every read command.
- `internal/extract` — heuristic engine.
- `internal/markdown` — goldmark-backed wrapper that yields plain-text
  spans excluding code regions. Shared with E6.
- Reflective `terminology schema` implementation.

## `lookup` — match policy

Per cli-design.md, `lookup TERM` searches for `TERM` as a `<term>` value
in any `<termSec>`. Specifics:

- **Case sensitivity** — case-fold under Unicode default casefold.
- **Normalization** — NFC the query and the glossary term before
  compare.
- **Match type** — exact surface form (not lemmatized, not fuzzy). No
  "did you mean" suggestions in v1 — that requires per-script
  edit-distance tuning and a ranking policy that aren't worth the
  surface for an exact-lookup tool. Agents that need suggestions build
  them on top of `extract` or multiple `lookup` calls.
- **`--lang LANG`** — restricts to that language section.

Lean output (default):

```json
{
  "schema_version": 1,
  "ok": true,
  "results": [{
    "concept_id": "tzimtzum",
    "subject_field": "kabbalah",
    "languages": {
      "he": {"preferred": {"term": "צמצום"}},
      "es": {"preferred": {"term": "tzimtzum"}},
      "en": {"preferred": {"term": "tzimtzum"}}
    }
  }]
}
```

Not found → `results: []`, exit `1` (recoverable per
[error-handling](../adr/error-handling.md)).

## `schema` — reflective introspection

No embedded JSON file. At runtime `schema` walks three sources and
composes a single JSON description:

1. **Command tree** — recursive walk of `app.Root()` collecting
   `*cli.Command` and `cli.Flag` declarations: names, types, defaults,
   short aliases, picklist values, required flags.
2. **Output struct types** — reflection over the registry of envelope
   types in `internal/output/types.go` to produce per-command envelope
   shapes.
3. **Error registry** — enumeration of the `terr` sentinel collection
   built at init time across all `internal/<pkg>/errors.go` files.

```bash
$ terminology schema
{
  "schema_version": 1,
  "commands":    [ ... ],
  "envelopes":   { ... },
  "error_codes": [ ... ]
}

$ terminology schema --command validate
{
  "schema_version": 1,
  "name": "validate",
  "flags":      [ ... ],
  "envelope":   { ... },
  "exit_codes": [0, 1, 65, 3]
}
```

Because the output is computed from the live binary, drift between
"what `schema` says" and "what the binary does" is structurally
impossible. See
[schema-source-of-truth](../adr/schema-source-of-truth.md).

## `extract` — heuristic engine

Three heuristics, applied in order over the corpus:

1. **Capitalized phrases.** Sequences of capitalized words not at
   sentence start. Tunable via per-language pattern hints (e.g.
   Spanish title-case rules differ from English).
2. **Foreign-script tokens.** Tokens whose dominant Unicode script
   (via `golang.org/x/text/unicode`) differs from the surrounding
   paragraph's script. The Hebrew-in-Spanish case is the motivating
   example.
3. **High-frequency tokens.** Frequency-based; subject to `--min-freq`
   (default 3). Subject to the stoplist when `--stopwords` is provided
   (see below).

Aggregated across all input files. Triage list, not NER.

### Flags

- `--exclude PATH` — exclude terms already defined in the given TBX
  file.
- `--script SCRIPT` — filter results to a specific script: `latin`,
  `hebrew`, `cyrillic`, `arabic`, `any`. Defaults to `any`.
- `--lang LANG` — see "Language detection" below.
- `--stopwords PATH` — see "Stoplist policy" below.
- `--min-freq N` — minimum frequency for the high-frequency heuristic.
  Default `3`.

### Stoplist policy

No stoplists ship with the binary. The high-frequency heuristic runs
without filler-word suppression by default — meaning `--min-freq`
alone gates the noise.

Users who want filler suppression pass `--stopwords PATH` pointing at
a newline-separated file of words to exclude. When provided, the
stoplist applies only to the high-frequency heuristic (capitalized-
phrases and foreign-script results are unaffected — those heuristics
don't need filler exclusion).

Rationale: bundling per-language lists means picking sources, picking
sizes, and explaining provenance for every shipped language.
Forcing the user to supply (or omit) the list moves that decision to
where it belongs.

#### Desired future feature: bundled NLP-grade stoplists

A v2 candidate worth researching: ship curated stoplists derived from
**NLTK** or **SpaCy** per supported language, embedded via
`//go:embed` and selected by `--lang`. Open questions to resolve
before adopting:

- **License compatibility.** NLTK data is Apache-2.0 friendly; SpaCy's
  language data carries per-language licenses (mostly MIT/CC-BY but
  not uniform). Attribution requirements need to be captured in
  `NOTICE`.
- **Binary size.** A representative ~200-word list per language is
  trivial (~2 KB); larger "professional" lists (1k–5k words) per
  language could add hundreds of KB across the matrix. Budget against
  the release size goal.
- **Coverage matrix.** Decide the supported set up front (`en`, `es`,
  `he`, then?) so the bundled list and the language-detection logic
  don't drift.
- **Override precedence.** `--stopwords PATH` continues to override
  the bundle; `--no-stopwords` may be needed to suppress the bundle
  when a user wants pure-frequency output.

Tracked as a v2 candidate, not v1 scope.

### Language detection

`--lang` defaults follow this precedence:

1. **Markdown frontmatter.** If the file's YAML frontmatter contains
   `lang: LANG`, that wins.
2. **`--lang` flag.** If passed, applies to all files in the corpus
   that lack frontmatter.
3. **Default to `en`.** Only when neither frontmatter nor `--lang` is
   present.

If the corpus is mixed-language and lacks frontmatter, the user is
expected to pass `--lang` explicitly per invocation (typically running
`extract` once per language). No statistical n-gram detector — that
introduces probabilistic failure modes on short documents and pulls in
another dependency for a triage tool.

### Markdown awareness

`extract` parses markdown rather than treating it as plain text. Code
regions (fenced blocks ``` ``` ``` ``` and `inline code`) are skipped
to eliminate false positives from identifiers like `getUserById`,
filenames, or library names that appear in code samples.

Parsing uses **`github.com/yuin/goldmark`** — CommonMark-compliant,
well-maintained, single dependency. The same parser is reused by E6
(`scan`, `check`) for consistency across the binary's markdown
handling. The shared wrapper lives in `internal/markdown/`:

```go
// internal/markdown/text.go
package markdown

// Spans walks src and yields plain-text spans (with their byte
// offsets in the original input) for every node that is not a
// code block, inline code, or HTML block.
func Spans(src []byte) iter.Seq[Span]
```

`extract` consumes the spans; `scan`/`check` (E6) consume them with
their line/column positions preserved so match locations are reported
in the original markdown's coordinates, not the de-coded text's.

Streaming output is **not** in v1. A single JSON envelope holds the
full result list; agents that hit memory pressure paginate by passing
fewer files per invocation. NDJSON streaming is a v2 candidate per
cli-design.md §"Out of scope".

## `--fields` projection

Shared across all read commands. Path syntax: `a,b,c.d`, with `*` for
map wildcards. Lives in `internal/output/fields.go`.

Paths are **validated** against the output struct's `json` tags via
reflection. Unknown paths produce an `invalid_field` error envelope
(exit `2` — usage error, per
[error-handling](../adr/error-handling.md)):

```bash
$ terminology lookup tzimtzum --fields concpet_id
{"schema_version":1,"ok":false,"error":{
  "code":"invalid_field",
  "message":"unknown field path: concpet_id",
  "hint":"valid paths for lookup: concept_id, subject_field, languages, languages.*.preferred.term, ..."
}}
```

Agents get deterministic feedback instead of silently-empty results
from typos. The hint enumerates valid paths for the current command,
derived from the same reflection that powers `terminology schema`.

A new sentinel lands with this epic:

```go
// internal/output/errors.go
var ErrInvalidField = terr.New(
    "invalid_field", 2,
    "see `terminology schema --command CMD` for valid paths",
    "unknown field path",
)
```

## Dependencies introduced

- **`github.com/yuin/goldmark`** — markdown parser, reused across E4
  and E6.

That brings the project's third-party Go list to:

- `urfave/cli/v3`
- `golang.org/x/text`
- `golang.org/x/term` (from E2/logging)
- `github.com/rogpeppe/go-internal/lockedfile` (from E2)
- `github.com/yuin/goldmark` (this epic)

All Go-team-adjacent or single-purpose libraries; the dependency
posture in cli-design.md still holds.

## Hand-offs

- `internal/output` is consumed by every read command and every write
  command's envelope.
- `internal/markdown` is shared with [E6](006-scan-check.md) — defined
  once here.
- Reflective `terminology schema` mechanics are formalized in
  [schema-source-of-truth](../adr/schema-source-of-truth.md).
- `ErrInvalidField` registered per
  [error-handling](../adr/error-handling.md).
