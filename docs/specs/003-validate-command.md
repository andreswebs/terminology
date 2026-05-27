# E3 — `terminology validate`

> **Status**: APPROVED. First real command end-to-end; consumes the
> domain model from [E2](002-domain-and-io.md) and produces the
> documented envelope on stdout plus a structured warnings list.

## Scope

Three validation tiers, each independently testable:

1. **Well-formedness** — XML parses; required structural elements
   (`<tbx>`, `<conceptEntry>`, `<langSec>`, `<termSec>`) present with
   their required attributes.
2. **Schema (dialect) tier** — every used element belongs to the
   TBX-Linguist supported set (Core + Min + Basic + Linguist modules);
   attributes well-typed; picklist values from the accepted set.
3. **Semantic tier** — BCP 47 language tags well-formed; concept IDs
   unique; `crossReference` IDREFs resolve; dialect-specific consistency
   rules (e.g. a `<langSec>` has at least one `<termSec>` with a
   `<term>`).

### `--strict`

`--strict` is **opt-in**. Lenient default matches cli-design.md
§"Required structural elements" ("tolerate unknown elements"), so a
default `validate` invocation against a file with foreign extensions
still produces a useful envelope.

When set, `--strict`:

- Promotes "out-of-set element" from silent to **warning**.
- Promotes "unresolved `crossReference` IDREF" from warning to **error**.

## Output

```json
{
  "schema_version": 1,
  "ok": true,
  "concepts": 47,
  "languages": ["en", "es", "he"],
  "warnings": []
}
```

`concepts` and `languages` are **as-found, raw counts** — `concepts:
47` means "the file declared 47 `<conceptEntry>` elements". Warnings
tell you that 2 of them collide on `id`; the consumer can compute the
deduplicated count from the warning list if needed. Reporting the
file's shape unambiguously is more useful than masking duplicates
behind a single "meaningful" number.

`languages` is sorted by `xml:lang` ASCII byte order, per
[determinism](../adr/determinism.md).

### Exit codes

- `0` — clean (no warnings, no errors).
- `1` — warnings present, no errors (recoverable).
- `65` — file unparseable or required structure missing. Routes through
  the error envelope with code `validation_error`
  (`ErrValidationError` sentinel in `internal/tbx/errors.go`).
- `3` — I/O error reading the file.

`65 = validation_error` (BSD `EX_DATAERR`) is unified across `validate`,
`apply`, and write-side runtime-input-rejection codes (`duplicate_id`,
`not_found`, `dangling_crossref`, `invalid_id`, `invalid_picklist`).
See [error-handling](../adr/error-handling.md).

## Tier sequencing

If tier 1 (well-formedness) fails, the command returns immediately —
the domain model can't be built, so tiers 2 and 3 have nothing to
inspect. The error envelope carries `validation_error` with the parse
error chained.

If tier 1 succeeds, the domain model is built. Tiers 2 and 3 then run
**together**, aggregating warnings; a tier 2 failure does **not**
short-circuit tier 3 because the model is already in memory and the
checks are independent. The agent gets a single envelope with the full
picture instead of paying for a second `validate` round-trip.

## Warnings shape

```json
{
  "code": "unresolved_crossref",
  "message": "concept 'razon-historica' references unknown ID 'kabbalah'",
  "concept_id": "razon-historica",
  "line": 142,
  "column": 12
}
```

Every warning carries: `code`, `message`, optional `concept_id`, optional
`line` + `column`.

### Warning codes (initial set)

- `unknown_element` — element outside the supported set (`--strict` only).
- `unresolved_crossref` — IDREF doesn't resolve.
- `duplicate_id` — two `<conceptEntry>` share an `id`.
- `invalid_lang_tag` — BCP 47 well-formedness fails.
- `invalid_picklist` — value not in the accepted set for that picklist.
- `legacy_form_normalized` — `usageRegister` or bare admin-status
  normalized on read (info-only; `--strict` only).
- `missing_term` — `<langSec>` with no `<term>`.

## BCP 47 validation

Language-tag checking is **well-formedness only**, via
`golang.org/x/text/language.Parse`. The parse accepts any
syntactically-valid tag — it does not verify that the subtags exist in
the IANA language-subtag-registry, and it does not canonicalize.

Reasons:

- The cli-design.md scoping declares "agent is not a trusted operator"
  for input but does not require IANA-registry semantic validation
  here.
- Bundling the IANA registry (~2 MB) or pulling a third-party lib is
  disproportionate for the value (catching typos like `enn` instead of
  `en`) — agents rarely emit malformed tags.
- Canonicalization is a write-side concern handed off to E7. Doing it
  on read would conflict with the "normalize on write only" rule from
  [determinism](../adr/determinism.md).

Malformed tags become `invalid_lang_tag` warnings, surfaced with the
`concept_id` and line/column of the offending `<langSec>`.

## Line/column tracking

The reader wraps the input `io.Reader` with a streaming newline-index:

```go
// internal/tbx/lineindex.go
type LineIndex struct {
    newlines []int64  // append-only byte offsets of '\n'
}

func (li *LineIndex) Wrap(r io.Reader) io.Reader { /* counts newlines as bytes pass */ }

func (li *LineIndex) Position(offset int64) (line, col int) {
    // binary-search newlines for the largest offset <= `offset`
    // line = index + 1, col = offset - prev-newline
}
```

File content never sits in memory — `encoding/xml.Decoder` streams the
wrapped reader; the index only stores 8 bytes per newline. A 1 MB / 30k-
line file costs ~240 KB of index. Pathological 1M-line inputs cost
~8 MB. Bounded and acceptable.

Warnings call `LineIndex.Position(decoder.InputOffset())` to attach
line/column. Offsets always reference bytes the decoder has already
streamed past, so the index is up-to-date at lookup time.

## Picklist values

The accepted values for every TBX picklist (part-of-speech,
admin-status, register, …) live in `internal/tbx/picklist.go` as a
single source:

```go
// internal/tbx/picklist.go
package tbx

func PartOfSpeech() []string { return []string{"noun", "verb", "adjective", ...} }
func AdminStatus() []string  { return []string{"preferred-admn-sts", "admitted-admn-sts", ...} }
// ...
```

Consumed by two layers:

- **`internal/app`** — urfave flag enums for write commands (per
  [001 Q3](001-cli-surface-stub.md)).
- **`internal/tbx`** — file-side validation here and in every
  reader-consuming command (unknown picklist values map to
  `StatusUnspecified` etc. with an `invalid_picklist` warning).

Both sides import the same constants; drift is structurally impossible
without changing the shared file.

## Hand-offs

- Picklist values shared with [E7](007-write-commands.md) via
  `internal/tbx/picklist.go`.
- Warning struct shape declared in `internal/output/types.go` (per
  [schema-source-of-truth](../adr/schema-source-of-truth.md)).
- `ErrValidationError` sentinel registered per
  [error-handling](../adr/error-handling.md).
- Round-trip and golden tests:
  [testing](../adr/testing.md).
