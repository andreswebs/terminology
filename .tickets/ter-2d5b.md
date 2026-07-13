---
id: ter-2d5b
status: closed
deps: [ter-nfsv]
links: []
created: 2026-07-12T11:51:38Z
type: feature
priority: 2
assignee: Andre Silva
tags: [field-feedback, read, search, feat-5]
---
# FEAT — search command: reading-aware normalized-substring finder (FEAT-5)

# FEAT (C) — search command: reading-aware, normalized-substring finder (FEAT-5)

## Why this ticket exists

Field feedback (`.local/tmp/issues.md`, FEAT-5): the most natural query for a
Japanese martial-arts glossary — the romaji, often without the hyphen
(`katatedori`) — returns nothing. `lookup` matches only `<term>` surface forms,
exact whole-string (NFC + casefold), never `ling:reading` / `ling:readingNote`,
and treats hyphen and space as distinct. A user cannot find an entry by the
very string they remember. A working reference implementation already exists:
`.local/tmp/terminology-search.py`.

Covers: **FEAT-5** (search across terms + readings, substring, diacritic/
separator folding).

## Design decisions (agreed)

- **Distinct from `lookup`.** `lookup` stays the strict exact finder; `search`
  is the discovery finder. Do not overload `lookup` with flags.
- **Match model: normalized substring.** NFKD → strip combining marks
  (macron-fold, `ō`→`o`) → drop non-alphanumerics (hyphens/spaces) → casefold;
  CJK survives. A hit = the normalized query is a substring of a normalized
  field. **No edit-distance.** Deterministic (determinism ADR); results ordered
  by `concept_id`; no scoring.
- **Default haystack:** `concept_id` + every language's term surfaces + per-term
  `reading` and `reading_note`. Opt-in flag widens to
  definitions/notes/contexts/subject_field.
- **Output:** the canonical concept shape (from the foundation ticket).
- **Exit codes mirror `lookup`:** zero hits → exit 1 `not_found`.
- **Search normalization is its own thing**, separate from the prose-scanning
  matcher in `internal/match` (different job: aggressive folding for discovery
  vs. boundary-aware prose matching).

## Reference oracle: `terminology-search.py`

The wrapper in `.local/tmp/terminology-search.py` is the behavioral spec. Port
its cases:

- `norm(s)`: `unicodedata.normalize("NFKD", s)` → drop combining marks → keep
  `isalnum()` chars lowercased. CJK/kana are alnum, so they survive; romaji is
  matched regardless of hyphens, spaces, or macrons.
- Haystack (its `haystack()`): id, subject_field, definitions, and per-language
  `term`/`reading`/`reading_note`. **This ticket's default excludes
  definitions/subject_field** (opt-in via `--include`); everything else matches.
- `matches(c, q)`: `norm(q) in norm(field)` for any field — substring.

Once `search` ships, note in docs that the Python wrapper can be retired.

## Current state (verified against source)

- Only finder is [`Glossary.Lookup`](src/internal/tbx/lookup.go#L14-L50):
  exact, NFC + casefold, term surfaces only.
- The prose matcher's normalization lives in
  `src/internal/match/normalize.go` — a *different* normalization (do not
  reuse; contrast in the ticket).
- No `search` command exists. Registration points: `app/root.go`,
  `output/types.go` `init()` (envelope + exit codes).
- Read-command plumbing to reuse: `tbxPathFromRoot`, `wrapLoadError`,
  `readFieldsFlag()`, `langFlag(...)`, `output.EmitProjected`, and the
  positional-arg pattern in
  [`lookup.go:11-26`](src/internal/app/commands/lookup.go#L11-L26).
- The canonical serializer `write.ConceptToWriteResult` comes from the
  foundation ticket.

## Scope of work

### 1. Search core (`src/internal/tbx/search.go`)

Mirror `Glossary.Lookup` for symmetry (same package, same file style):

```go
// SearchOptions configures Glossary.Search.
type SearchOptions struct {
    Lang    string   // restrict the haystack to this language tag ("" = all)
    Include []string // widen haystack: "definitions","notes","contexts","subject_field"
}

// Search returns concepts whose normalized fields contain the normalized query
// as a substring, ordered by concept id.
func (g *Glossary) Search(query string, opts SearchOptions) []Concept
```

- Normalization helper `foldForSearch(s string) string`: NFKD
  (`golang.org/x/text/unicode/norm`), drop combining marks
  (`unicode.Is(unicode.Mn, r)` or `runes.Remove(runes.In(unicode.Mn))`), keep
  `unicode.IsLetter || unicode.IsNumber`, lowercase (`cases.Fold()` or
  `strings.ToLower` on the folded runes). Match the Python `norm` exactly for
  the ported cases.
- Default haystack per concept: `concept_id`, and for each langSec (all, or the
  one in `opts.Lang`) each term's `Surface`, `Reading`, `ReadingNote`.
  `opts.Include` adds `Definitions` (concept + langSec), `Notes`, `Contexts`,
  `subject_field` as requested.
- A concept matches if the normalized query is a substring of any haystack
  field. Return matched concepts sorted by `ID` (determinism).
- Empty query → empty result (mirror `Lookup`'s guard at
  [`lookup.go:17-19`](src/internal/tbx/lookup.go#L17-L19)).

### 2. `search` command (`src/internal/app/commands/search.go`)

- Positional `query` arg; `Before: argBounds(1,1)`.
- Flags: `readFieldsFlag()`, `langFlag(false, "restrict the search to this
  language's fields")`, and `--include` (comma-separated;
  validate against the allowed set `definitions,notes,contexts,subject_field`).
- Loads glossary (`tbxPathFromRoot` + `wrapLoadError`), calls `g.Search`,
  serializes matches via `write.ConceptToWriteResult`.
- Emits `SearchEnvelope{schema_version, ok, results:[]output.WriteResult}` via
  `output.EmitProjected(cmd.Root().Writer, env, cmd.String("fields"))`.
- Zero hits → exit 1 with error code `not_found` (reuse the
  `lookupNotFoundError` pattern / shared not-found error).

### 3. Registration + schema

- Register the command in [`app/root.go`](src/internal/app/root.go#L26-L37).
- `RegisterEnvelope("search", SearchEnvelope{})` and
  `RegisterExitCodes("search", []int{0,1,2,3,65})` in
  [`output/types.go` `init()`](src/internal/output/types.go#L5-L32); add a
  `MarshalJSON` normalizing nil `Results` → `[]`.

### 4. Documentation

- Add `search` to [`docs/cli-design.md`](docs/cli-design.md) (usage, flags,
  match semantics, exit codes) and the terminology skill docs; state the
  normalization rules and the terms+readings+id default haystack with the
  `--include` widening.
- Note that `.local/tmp/terminology-search.py` is superseded.
- `markdownlint-cli2 --config ~/.markdownlint.yaml --fix` on edited markdown.

### Go conventions

- Exported `Search`, `SearchOptions`, `SearchEnvelope`, and the command
  constructor `Search()` get doc comments beginning with the identifier name;
  `MixedCaps`; `gofmt` clean.
- Keep `foldForSearch` unexported; pure function; no shared mutable state.
- Early-return; happy path at minimal indentation. Reuse
  `golang.org/x/text` (already a dependency per `internal/tbx/lookup.go`).

## TDD plan (vertical slices — one test, one change, repeat)

Public interfaces under test: `tbx.Glossary.Search` (unit) and the `search`
command (CLI/golden). Seed glossaries with the Aikido bundle model (kanji as
term, hiragana as `reading`, romaji as `reading_note`). Port cases from
`terminology-search.py`. Go RED→GREEN per cycle.

### Cycle 1 (tracer) — romaji-without-hyphen finds a reading

RED: a concept with a `ja` preferred term `片手取り`, `reading` `かたてどり`,
`reading_note` `katate-dori`. `g.Search("katatedori", SearchOptions{})` returns
that concept. Fails today (no `Search`).
GREEN: implement `foldForSearch` + substring match over the default haystack.

### Cycle 2 — diacritic and separator folding, substring

RED: `Search("kokyu")` matches a `reading_note` `kokyū` (macron fold);
`Search("grab")` matches an `en` term `single hand grab` (substring + space
fold). 
GREEN: covered by `foldForSearch` + substring; add cases.

### Cycle 3 — CJK and kana queries

RED: `Search("片手取り")` matches the kanji term; `Search("かたてどり")` matches
the hiragana reading.
GREEN: confirm CJK/kana survive normalization (alnum-preserving); add cases.

### Cycle 4 — default haystack excludes descriptive text

RED: a query that appears only in a `definition` returns no hit under default
options, but DOES hit with `SearchOptions{Include: []string{"definitions"}}`.
GREEN: gate definitions/notes/contexts/subject_field behind `Include`.

### Cycle 5 — determinism + `--lang`

RED: multiple matches are returned sorted by `concept_id`;
`SearchOptions{Lang:"ja"}` restricts the haystack to the `ja` langSec (an
`en`-only match is excluded).
GREEN: sort by id; honor `Lang`.

### Cycle 6 — command end-to-end

RED (CLI/golden): `terminology search katatedori` emits canonical `results`
sorted by id; `terminology search zzzz` exits 1 `not_found`; `search <q>
--fields results.concept_id` projects; `search <q> --include definitions`
widens.
GREEN: implement the command; register envelope + exit codes.

### Refactor

- If `search` and `lookup` share a not-found error or a "matches →
  []WriteResult" helper, extract it. Keep `foldForSearch` cohesive.
- Run tests after each step. Never refactor while RED.

## Acceptance criteria

- `terminology search katatedori` finds the entry stored as reading
  `かたてどり` / `katate-dori`; `かたてどり`, `片手取り`, `grab`, `kokyu`→`kokyū`
  all resolve.
- Matching is normalized substring (folds macrons, ignores hyphens/spaces,
  casefolds; CJK preserved); results sorted by `concept_id`; deterministic.
- Default haystack = terms + readings + id; `--include` widens to
  definitions/notes/contexts/subject_field; `--lang` restricts the haystack.
- Output is the canonical concept shape; `--fields` works; zero hits → exit 1
  `not_found`.
- Discoverable via `terminology schema --command search`.
- Docs updated; the Python wrapper is noted as superseded.
- `make build` passes from the project root.

## Files to touch

- `src/internal/tbx/search.go` (+ `search_test.go`) — `Search`,
  `SearchOptions`, `foldForSearch`.
- `src/internal/app/commands/search.go` (+ test) — command.
- `src/internal/output/types.go` — `SearchEnvelope` + `init()` registration.
- `src/internal/app/root.go` — register command.
- `src/internal/app/*_golden_test.go` — search goldens.
- `docs/cli-design.md`, terminology skill docs.

## Validation

`make build` from the project root; tighter loop:
`go test ./src/internal/tbx/... ./src/internal/app/... ./src/internal/output/...`.
Do not silence lint with `_ =`.

## Dependencies

Depends on the **foundation ticket (FEAT-A)** for
`write.ConceptToWriteResult` (search results emit the canonical concept shape).
Independent of the read-commands ticket (FEAT-B).


## Notes

**2026-07-13T17:21:38Z**

Implemented the search command (FEAT-5). New: internal/tbx/search.go (Search, SearchOptions, foldForSearch) + search_test.go; internal/app/commands/search.go (command, --include flag validated against definitions,notes,contexts,subject_field, --lang, --fields); output.SearchEnvelope + registration; registered in app/root.go. Emits canonical write.ConceptToWriteResult shape, results sorted by concept_id, zero hits -> exit 1 not_found (reuses lookupNotFound). foldForSearch mirrors terminology-search.py norm(): NFKD, drop combining marks (macron fold), keep letters/numbers (CJK/kana survive, hyphens/spaces dropped), casefold; matching is substring. Docs: cli-design.md + SKILL.md updated, python wrapper noted superseded. Added fixture testdata/fixtures/aikido-dct.tbx + 10 golden tests. Regenerated both schema golden files. make build green. NOTE for next dev: ter-lutz (FEAT-2/3 export/show/list + lookup canonical-shape upgrade) is the remaining ready ticket; it will make lookup emit the same canonical shape search already uses.
