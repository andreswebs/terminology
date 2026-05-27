# E6 — `terminology scan` and `terminology check`

> **Status**: APPROVED. The two matcher-driven commands sitting on top
> of [E5 the matcher](005-matcher.md) and [E2 the domain
> model](002-domain-and-io.md). E6 is the wiring layer: load glossary,
> build matcher, run it, format the envelope.

## Scope

- `scan FILE` — list every glossary term occurrence in `FILE`.
  Informational; exit `0` regardless of findings.
- `check SRC TGT` — verify translated `TGT` against `SRC` given the
  glossary. Exit `1` on any violation.

Both commands feed markdown through the shared `internal/markdown`
package (goldmark-backed, from [E4](004-read-commands.md)) so code
regions are skipped consistently. The matcher's
canonical-normalization + position-mapping pipeline (E5) carries
line/column reporting back to the original source.

## `scan` details

```
terminology scan FILE [--lang LANG] [--tbx PATH]
                      [--context CHARS] [--fields LIST]
```

- Single-file in v1. Variadic `scan FILE...` is a non-breaking
  extension if a use case emerges; agents loop trivially today.
- Default scans all languages. `--lang` restricts.
- Output: matches sorted by `(line, column)`.
- Each match carries: `concept_id`, `term`, `lang`, `status`, `line`,
  `column`, `context`.

`scan` always exits `0` — it's informational. A file with no matches
returns `matches: []`, still exit `0`.

### Context window

`--context CHARS` controls how much surrounding text accompanies each
match. **Default `80`** (40 characters on each side of the match,
clamped at line boundaries). Balances usefulness (enough prose to
disambiguate the sentence) against agent-context cost when a file has
many matches.

Larger values are available for human-driven inspection; `--context 0`
omits the field for the leanest envelope.

## `check` details

```
terminology check SRC TGT [--source-lang LANG] [--target-lang LANG]
                          [--tbx PATH] [--strict]
                          [--context CHARS] [--fields LIST]
```

### Algorithm

1. Scan `SRC` for source-language preferred + admitted terms.
2. For each concept with source occurrences:
   - Scan `TGT` for the preferred target term. **Zero** occurrences →
     `missing` violation.
   - Scan `TGT` for any deprecated/superseded target variants. Each
     occurrence → `forbidden_variant` violation.
3. Concepts absent from `SRC` are ignored.

### Counting policy

"At least one" semantics. Source occurrence count does not enter the
comparison:

- Source ≥1, target ≥1 → OK.
- Source ≥1, target 0 → `missing` violation.

Under-translation warnings (e.g. source=5, target=1) are not emitted.
Pronouns, elision, and stylistic omission legitimately reduce target
occurrences; we don't have a way to distinguish those from real
problems, and any threshold would be arbitrary. The bar is what
cli-design.md documented: at least one match.

### `--strict`

Promotes `admittedTerm` matches in `TGT` to violations of type
`admitted_variant`. Distinct from `--strict` on `validate` (which
controls dialect-strictness on read). The two flags share a name but
no implementation — different commands, different `*cli.BoolFlag`
declarations.

### Exit code

`0` if no violations, `1` if any (recoverable per
[error-handling](../adr/error-handling.md)).

## Language resolution

Both commands need to know which language a file is in. The precedence
is:

1. **Markdown frontmatter.** YAML `lang: LANG` in `SRC` / `TGT` wins.
2. **CLI flag.** `--lang` (for `scan`) or `--source-lang` /
   `--target-lang` (for `check`).
3. **Fail with usage error.** If neither is present, return
   `ErrLanguageRequired` (`language_required`, exit `2`). No hardcoded
   default — silently assuming `es`/`en` bites multilingual projects
   and is exactly the failure mode the precedence is here to prevent.

The hint enumerates both ways to supply the language:

```
✗ check: language not specified for SRC
  hint: pass --source-lang LANG or add 'lang: LANG' to the file's frontmatter
```

A new sentinel lands with this epic:

```go
// internal/match/errors.go (or wherever scan/check live)
var ErrLanguageRequired = terr.New(
    "language_required", 2,
    "pass --lang/--source-lang/--target-lang or add 'lang: LANG' to frontmatter",
    "language not specified",
)
```

Mirrors the precedence in [E4 `extract`](004-read-commands.md) which
defaults to `en` only because `extract` works on a corpus where one
language is typically dominant. `scan`/`check` have no such fallback
because misclassifying source/target language silently produces
wrong-but-plausible violations.

## Output

Per [cli-design.md](../cli-design.md). Notable: every violation
carries enough context to fix in one pass — no follow-up `lookup`
round-trip needed.

```json
{
  "schema_version": 1,
  "ok": false,
  "violations": [{
    "type": "missing",
    "concept_id": "tzimtzum",
    "source_term": "tzimtzum",
    "expected_target": "צמצום",
    "source_occurrences": 5
  }, {
    "type": "forbidden_variant",
    "concept_id": "razon-historica",
    "variant": "razón histórica",
    "line": 142,
    "column": 12,
    "context": "...la razón histórica de su pensamiento se manifiesta..."
  }]
}
```

### Violation ordering

By `(line, column)` in `TGT`, primary. Matches the agent's
"fix top-to-bottom" flow — violations appear in the order they would
be encountered while editing the target file.

`missing` violations (which have no line/column in `TGT` — the
concept is *absent* by definition) sort to the end of the list,
grouped together. Within that tail group, `concept_id` ASCII order.

A future `--group-by concept` is a non-breaking opt-in if grouping
becomes a common request.

## Code regions

Both `scan` and `check` strip code regions (fenced blocks and inline
code) before matching, via `internal/markdown` (E4). This applies
**symmetrically** in `check`: code regions are skipped in both `SRC`
and `TGT`. Asymmetric skipping would produce false-positive `missing`
violations when a target uses a term in code that the source uses in
prose, or vice versa.

Scanning literal code for glossary terms is out of v1 scope. A
documented limitation in the package doc.

## Dependencies introduced

None new this epic. `internal/markdown` (goldmark), the matcher
(`cloudflare/ahocorasick`), and `internal/output` are all consumed
from prior epics.

## Hand-offs

- Matcher: [E5](005-matcher.md).
- Domain model: [E2](002-domain-and-io.md).
- Output envelope: `internal/output` from [E4](004-read-commands.md).
- Markdown plain-text spans: `internal/markdown` from
  [E4](004-read-commands.md).
- `ErrLanguageRequired` registered per
  [error-handling](../adr/error-handling.md).
- Perf budget: [testing](../adr/testing.md) —
  `check` is the most expensive read command (two files × full
  variant-tagged matcher run).
