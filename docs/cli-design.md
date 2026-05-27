# `terminology` — CLI Specification (Go)

A command-line tool for agent-driven, terminology-focused academic translation. Reads markdown source, enforces consistent terminology against a TBX glossary, exposes a small set of deterministic operations as subcommands.

## Scope

This tool does one thing well: it makes terminology capture and usage deterministic for a translation project. It does not manage translation memory, perform XLIFF round-trips, track per-chapter status, or do segmentation. Those belong to other tools or other tools' problem.

In scope:

- Reading and validating a TBX-Linguist glossary file (Core + Min + Basic + Linguist modules, DCT and DCA styles on input)
- Writing TBX-Linguist files in canonical DCT style, including granular concept/term edits and bulk declarative `apply`
- Scanning markdown source files for terminology occurrences
- Verifying translated markdown against source and glossary
- Bootstrapping candidate terms from a corpus

Out of scope:

- Translation memory
- XLIFF / TMX / SRX
- Markdown structure validation (separate tool)
- Project state management
- Web UI, daemon mode, multi-user concerns

## Design principles

The tool is **agent-first, human-tolerant**. The primary caller is an AI agent
driving a translation or curation loop; humans use the same surface for
interactive lookup, debugging, and curation. Every interface decision is
driven by the following properties:

1. **Machine-readable output.** All commands emit JSON to stdout by default. Human-readable output is opt-in via `--format=text` (indented hierarchical blocks with `✓`/`✗` status symbols — no color, no tables, no Unicode dependency beyond status glyphs).
2. **Meaningful exit codes.** `0` indicates success. `1` indicates a logical "not found" or "violations exist" — not an error, but actionable. `2` indicates usage error (argv parse failure). `3` indicates I/O or storage error. `65` (the BSD `EX_DATAERR` convention) indicates a validation error — input data was rejected and the operation could not complete; state is unchanged. Emitted by `validate` (unparseable file), `apply` (payload validation failed), and write commands for runtime input rejection (`duplicate_id`, `not_found`, `dangling_crossref`, `invalid_id`, `invalid_picklist`). `75` (the BSD `EX_TEMPFAIL` convention) was used during development for the transient "under construction" state emitted by stub commands. All commands are now implemented; exit code 75 is no longer emitted.
3. **Single static binary.** No runtime dependencies. Cross-compiles for Linux, macOS, Windows. Distributed as a single executable.
4. **Structured error envelope.** Errors emit JSON to **stderr** when `--format=json`, human-readable text to stderr when `--format=text`. The JSON shape is `{ "ok": false, "error": { "code": "...", "message": "...", "hint": "..." } }`. Stdout stays clean for piping.
5. **Field masks on reads.** Every read command accepts `--fields a,b,c.d` to trim the output. Defaults are lean (just enough to identify what was found); callers opt into more detail. This protects the agent's context window as the data model grows.
6. **Dry-run on writes.** Every mutating command accepts `--dry-run`. The output is the **final-state preview** of the touched concepts (not a diff). Lets an agent validate a write before committing it.
7. **Schema introspection at runtime.** `terminology schema` dumps a `schema.json` embedded in the binary, describing every command, flag, and output shape. Generated at build time from the CLI metadata so it stays in sync.
8. **Input hardening.** The agent is not a trusted operator. Concept IDs, language tags, and file paths reject control characters, path traversals (`..`, `%2e`), percent-encoded segments, and embedded query params (`?`, `#`). Output paths are sandboxed to the CWD subtree.
9. **Explicit configuration.** No magic discovery. The TBX file path is resolved in order: `--tbx` flag, then `TERMINOLOGY_TBX` environment variable, then error. No CWD search.

Go conventions apply throughout: stdlib first, minimal external dependencies, idiomatic error handling, table-driven tests.

### `--strict` semantics

The `--strict` flag appears on more than one command with intuitive but
distinct meanings; document and test each independently:

| Command    | Effect of `--strict`                                                                           |
| ---------- | ---------------------------------------------------------------------------------------------- |
| `validate` | Warns on elements outside the supported set; promotes unresolved IDREFs from warning to error. |
| `check`    | Admitted (`admittedTerm-admn-sts`) variants raise violations instead of warnings.              |

## Project layout

```
terminology/
├── cmd/
│   └── terminology/
│       └── main.go              # CLI entry, subcommand dispatch
├── internal/
│   ├── tbx/                     # TBX parsing (DCT + DCA on read), DCT emission
│   │   ├── tbx.go
│   │   ├── emit.go              # canonical DCT writer
│   │   ├── id.go                # concept-ID derivation
│   │   ├── tbx_test.go
│   │   └── testdata/
│   ├── scan/                    # term scanning in markdown
│   │   ├── scan.go
│   │   └── scan_test.go
│   ├── check/                   # verification logic
│   │   ├── check.go
│   │   └── check_test.go
│   ├── extract/                 # term candidate extraction
│   │   ├── extract.go
│   │   └── extract_test.go
│   ├── write/                   # concept/term/apply mutation logic
│   │   ├── write.go
│   │   ├── apply.go
│   │   └── write_test.go
│   ├── app/commands/sanitize.go  # input validation (IDs, lang tags, paths)
│   │   └── sanitize_test.go
│   ├── output/                  # JSON / text formatting, field masks, error envelope
│   │   ├── output.go
│   │   ├── fields.go
│   │   └── errors.go
│   └── schema/                  # reflective `terminology schema` impl
│       └── schema.go            # walks urfave tree + output structs + terr registry
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

Standard Go layout. `cmd/terminology/main.go` is the only entry point. All logic lives under `internal/` so it isn't importable as a public library — keeping the API surface zero is deliberate.

## TBX dialect

The tool reads TBX-Linguist (ISO 30042:2019), which telescopes Core + Min +
Basic + Linguist modules. Both encoding styles are supported on input:

- **DCT (Data Category as Tag)** — preferred, more readable.
- **DCA (Data Category as Attribute)** — accepted for interoperability.

The TBX default namespace is `urn:iso:std:iso:30042:ed-2` (the dialect
schemas use the `ed-2` URN even though prose sometimes mentions `ed:3.0`; the
schemas are authoritative). In DCT mode the module namespaces are:

| Prefix   | URI                                  |
| -------- | ------------------------------------ |
| `min:`   | `http://www.tbxinfo.net/ns/min`      |
| `basic:` | `http://www.tbxinfo.net/ns/basic`    |
| `ling:`  | `http://www.tbxinfo.net/ns/linguist` |

See [`.local/research/tbx-linguist.md`](../.local/research/tbx-linguist.md)
for the full dialect specification, including the rationale that Linguist is
a small delta over Basic.

### Required structural elements

- `<tbx>` with `type="TBX-Linguist"` and `style="dct"` or `style="dca"`.
- `<conceptEntry>` with required `id` attribute.
- `<langSec>` with required `xml:lang` (BCP 47).
- `<termSec>` with required `<term>` child.

`<tbxHeader>` is parsed but not used for terminology semantics. Unknown
elements outside the supported set are tolerated and ignored (or reported as
warnings under `--strict`).

### Supported data categories

The tool understands the categories below at the indicated levels. Each row
applies to both encoding styles; the DCA carrier shows the equivalent
`<element type="name">` form.

#### Concept level (`<conceptEntry>`)

| Category                 | Module | DCT element                    | DCA carrier                            |
| ------------------------ | ------ | ------------------------------ | -------------------------------------- |
| `subjectField`           | Min    | `<min:subjectField>`           | `<descrip type="subjectField">`        |
| `definition`             | Basic  | `<basic:definition>`           | `<descrip type="definition">`          |
| `crossReference`         | Basic  | `<basic:crossReference>`       | `<ref type="crossReference">`          |
| `customerSubset`         | Min    | `<min:customerSubset>`         | `<admin type="customerSubset">`        |
| `projectSubset`          | Basic  | `<basic:projectSubset>`        | `<admin type="projectSubset">`         |
| `source`                 | Basic  | `<basic:source>`               | `<admin type="source">`                |
| `xGraphic`               | Basic  | `<basic:xGraphic>`             | `<xref type="xGraphic">`               |
| `externalCrossReference` | Min    | `<min:externalCrossReference>` | `<xref type="externalCrossReference">` |
| `note`                   | Core   | `<note>`                       | `<note>`                               |

#### Language-section level (`<langSec>`)

| Category     | Module | DCT element          | DCA carrier                   |
| ------------ | ------ | -------------------- | ----------------------------- |
| `definition` | Basic  | `<basic:definition>` | `<descrip type="definition">` |
| `source`     | Basic  | `<basic:source>`     | `<admin type="source">`       |

#### Term-section level (`<termSec>`)

| Category                 | Module   | Value type | DCT element                            | DCA carrier                                        |
| ------------------------ | -------- | ---------- | -------------------------------------- | -------------------------------------------------- |
| `administrativeStatus`   | Min      | picklist   | `<min:administrativeStatus>`           | `<termNote type="administrativeStatus">`           |
| `partOfSpeech`           | Min      | picklist   | `<min:partOfSpeech>`                   | `<termNote type="partOfSpeech">`                   |
| `grammaticalGender`      | Basic    | picklist   | `<basic:grammaticalGender>`            | `<termNote type="grammaticalGender">`              |
| `grammaticalNumber`      | Linguist | picklist   | `<ling:grammaticalNumber>`             | `<termNote type="grammaticalNumber">`              |
| `register`               | Linguist | picklist   | `<ling:register>`                      | `<termNote type="register">`                       |
| `termType`               | Basic    | picklist   | `<basic:termType>`                     | `<termNote type="termType">`                       |
| `termLocation`           | Basic    | picklist   | `<basic:termLocation>`                 | `<termNote type="termLocation">`                   |
| `geographicalUsage`      | Basic    | string     | `<basic:geographicalUsage>`            | `<termNote type="geographicalUsage">`              |
| `context`                | Basic    | noteText   | `<basic:context>`                      | `<descrip type="context">`                         |
| `transferComment`        | Linguist | string     | `<ling:transferComment>`               | `<termNote type="transferComment">`                |
| `reading`                | Linguist | string     | `<ling:reading>` (in `<adminGrp>`)     | `<admin type="reading">` (in `<adminGrp>`)         |
| `readingNote`            | Linguist | string     | `<ling:readingNote>` (in `<adminGrp>`) | `<adminNote type="readingNote">` (in `<adminGrp>`) |
| `source`                 | Basic    | string     | `<basic:source>`                       | `<admin type="source">`                            |
| `customerSubset`         | Min      | string     | `<min:customerSubset>`                 | `<admin type="customerSubset">`                    |
| `projectSubset`          | Basic    | string     | `<basic:projectSubset>`                | `<admin type="projectSubset">`                     |
| `externalCrossReference` | Min      | URL        | `<min:externalCrossReference>`         | `<xref type="externalCrossReference">`             |
| `crossReference`         | Basic    | IDREF      | `<basic:crossReference>`               | `<ref type="crossReference">`                      |

#### Transactions

`<transacGrp>` is recognized at concept, langSec, and termSec levels and
carries `<basic:transactionType>` (picklist: `origination`, `modification`),
`<date>`, and `<basic:responsibility>`.

#### Inline markup (within noteText content)

The XLIFF-2-aligned Core inline elements are preserved on read but treated as
plain text for matching purposes: `<hi>`, `<sc>`, `<ec>`, `<ph>`, `<foreign>`.

### Picklist values

| Category               | Accepted values                                                                                                                                                                                                                         |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `administrativeStatus` | `preferredTerm-admn-sts`, `admittedTerm-admn-sts`, `deprecatedTerm-admn-sts`, `supersededTerm-admn-sts` (legacy bare forms `preferredTerm` / `admittedTerm` / `deprecatedTerm` / `supersededTerm` accepted on read, normalized on emit) |
| `partOfSpeech`         | `noun`, `verb`, `adjective`, `adverb`, `other`                                                                                                                                                                                          |
| `grammaticalGender`    | `masculine`, `feminine`, `neuter`, `other`                                                                                                                                                                                              |
| `grammaticalNumber`    | `singular`, `plural`, `dual`, `mass`, `otherNumber`                                                                                                                                                                                     |
| `register`             | `colloquialRegister`, `neutralRegister`, `technicalRegister`, `in-houseRegister`, `bench-levelRegister`, `slangRegister`, `vulgarRegister` (legacy `usageRegister` accepted on read as alias for `register`)                            |
| `termType`             | `fullForm`, `acronym`, `abbreviation`, `shortForm`, `variant`, `phrase`                                                                                                                                                                 |
| `termLocation`         | 18 UI-element values per the Basic module                                                                                                                                                                                               |
| `transactionType`      | `origination`, `modification`                                                                                                                                                                                                           |

`geographicalUsage` is a free-form string per spec; the tool does not enforce
ISO 3166.

### Terminology semantics

For each `<langSec>` within a concept:

- The **preferred term** is the `<term>` whose enclosing `<termSec>` has
  `administrativeStatus = preferredTerm-admn-sts`. If none is marked, the
  first `<termSec>` is treated as preferred.
- **Forbidden variants** are `<term>` elements whose enclosing `<termSec>` has
  `administrativeStatus = deprecatedTerm-admn-sts` or `supersededTerm-admn-sts`.
- **Admitted variants** (`admittedTerm-admn-sts`) are tolerated but not
  preferred — they satisfy the check but generate a warning. Under `--strict`,
  admitted variants raise violations.

These are standard TBX semantics — no custom fields needed.

## CLI surface

Commands are flat (no resource-prefix namespacing). The surface splits into
**read commands** that emit terminology data and **write commands** that
mutate the TBX file. A small set of utility commands rounds out the surface.

Read commands:

- `terminology validate` — structural and semantic validation
- `terminology lookup` — term lookup across languages
- `terminology scan` — find glossary terms in markdown
- `terminology check` — verify a translated file against a source
- `terminology extract` — bootstrap candidate terms from a corpus

Write commands (see "Write commands" section below):

- `terminology concept add | update | remove`
- `terminology term add | deprecate`
- `terminology apply` — bulk declarative writes from a payload

Utility commands:

- `terminology schema` — dump the embedded `schema.json` describing the full CLI

Every command accepts the global flags `--tbx PATH`, `--format json|text`
(default `json`), and `--fields a,b,c.d` (on read commands). Write commands
additionally accept `--dry-run` and, where transactions are opt-in,
`--transaction` plus `--author NAME` (or `TERMINOLOGY_AUTHOR` env). The TBX
path is resolved from `--tbx`, then `TERMINOLOGY_TBX`, then errors (`code:
no_tbx_path`).

---

### `terminology validate`

Validates a TBX file against the supported subset.

**Synopsis**

```
terminology validate [--tbx PATH] [--strict]
```

**Behavior**

Parses the file, checks structural validity (required elements present, BCP 47 tags well-formed, IDs unique, `crossReference` IDREFs resolve). With `--strict`, warns on elements outside the supported set and promotes unresolved IDREFs from warning to error.

**Output**

```json
{
  "schema_version": 1,
  "ok": true,
  "concepts": 47,
  "languages": ["en", "es", "he"],
  "warnings": []
}
```

`languages` is sorted by `xml:lang` ASCII byte order (per
[determinism](adr/determinism.md)).

Exit 0 if valid, 1 if validation warnings present, 65 if file unparseable or required structure missing.

---

### `terminology lookup TERM`

Look up a term across all languages in the TBX file.

**Synopsis**

```
terminology lookup TERM [--lang LANG] [--tbx PATH] [--fields LIST]
```

**Flags**

- `--lang LANG` — restrict search to a specific language section
- `--tbx PATH` — path to TBX file (overrides `TERMINOLOGY_TBX`)
- `--fields LIST` — comma-separated dotted paths to include beyond the default lean set (e.g. `--fields definitions,languages.*.preferred.grammatical_gender`)

**Behavior**

Searches for `TERM` as a `<term>` value in any `<termSec>`. Returns matching concept(s).

**Output (default — lean)**

By default `lookup` emits a minimal projection sufficient to identify the
concept and recognize its preferred terms per language:

```json
{
  "ok": true,
  "results": [
    {
      "concept_id": "tzimtzum",
      "subject_field": "kabbalah",
      "languages": {
        "he": { "preferred": { "term": "צמצום" } },
        "es": { "preferred": { "term": "tzimtzum" } },
        "en": { "preferred": { "term": "tzimtzum" } }
      }
    }
  ]
}
```

Additional fields are opt-in via `--fields`. Examples:

- `--fields definitions,notes` — adds concept-level definitions and notes
- `--fields languages.*.preferred.part_of_speech,languages.*.preferred.grammatical_gender` — adds morphology
- `--fields languages.*.deprecated,languages.*.admitted` — adds variants

When opted in, per-term objects carry the supported data categories that are
populated for that term — keys absent from the input TBX are omitted rather
than emitted with empty values.

**Output (not found)**: `results` is empty array. Exit code 1.

---

### `terminology scan FILE`

Find all glossary term occurrences in a markdown file.

**Synopsis**

```
terminology scan FILE [--lang LANG] [--tbx PATH] [--fields LIST]
```

**Flags**

- `--lang LANG` — only scan for terms in the given language section (default: scan all languages)
- `--tbx PATH` — path to TBX file
- `--fields LIST` — comma-separated dotted paths to include beyond the default lean projection

**Behavior**

For each term in the TBX (across all language sections), scans `FILE` for whole-word occurrences. Whole-word boundary uses Unicode-aware matching: `\p{L}` and `\p{N}` define word characters, so Hebrew, Spanish, and Latin scripts all work correctly.

Returns matches sorted by line number.

**Output**

```json
{
  "ok": true,
  "file": "source/chapter-01.md",
  "matches": [
    {
      "concept_id": "c001",
      "term": "tzimtzum",
      "lang": "es",
      "line": 14,
      "column": 23,
      "context": "...El concepto de tzimtzum es central..."
    },
    {
      "concept_id": "c001",
      "term": "צמצום",
      "lang": "he",
      "line": 42,
      "column": 8,
      "context": "...la noción de צמצום aparece..."
    }
  ],
  "summary": {
    "total_matches": 2,
    "unique_concepts": 1
  }
}
```

Exit 0 always (scanning is informational, not a verification).

---

### `terminology check SRC TGT`

Verify that a translated target file respects the glossary given the source file.

**Synopsis**

```
terminology check SRC TGT [--source-lang LANG] [--target-lang LANG] [--tbx PATH]
                   [--strict] [--fields LIST]
```

**Flags**

- `--source-lang LANG` — source language. No default; resolved as
  YAML-frontmatter `lang:` in `SRC` → `--source-lang` flag → fail with
  `language_required` (exit `2`).
- `--target-lang LANG` — target language. Same precedence against `TGT`.
- `--tbx PATH` — path to TBX file
- `--strict` — admitted variants raise violations instead of warnings
- `--fields LIST` — comma-separated dotted paths to include beyond the default projection

Violations embed enough context for an agent to fix in one pass without a
follow-up `lookup` call: `concept_id` (stable across runs), `expected_target`,
and a context window for each `forbidden_variant`.

**Behavior**

For each concept that has terms in both source language and target language sections:

1. Scan `SRC` for occurrences of source-language terms (preferred or admitted).
2. For each concept with source occurrences:
   - Scan `TGT` for the preferred target term. If count is zero, emit a `missing` violation.
   - Scan `TGT` for any deprecated target variants. Each occurrence is a `forbidden_variant` violation.
3. Concepts not present in source are ignored (no over-eager enforcement).

The check is approximate by design: it doesn't enforce 1:1 occurrence counts (pronouns and elision legitimately reduce target occurrences).

**Output**

```json
{
  "ok": false,
  "source": "source/ch1.md",
  "target": "target/ch1.md",
  "violations": [
    {
      "type": "missing",
      "concept_id": "c001",
      "source_term": "tzimtzum",
      "source_occurrences": 3,
      "expected_target": "tzimtzum",
      "target_occurrences": 0
    },
    {
      "type": "forbidden_variant",
      "concept_id": "c001",
      "variant": "contraction",
      "line": 17,
      "column": 4
    }
  ],
  "warnings": [],
  "summary": {
    "violations": 2,
    "warnings": 0,
    "concepts_checked": 12
  }
}
```

Exit 0 if no violations, 1 if any violations present.

---

### `terminology extract FILE...`

Extract candidate terminology from one or more markdown files.

**Synopsis**

```
terminology extract FILE... [--min-freq N] [--lang LANG] [--exclude PATH]
                     [--script SCRIPT]
```

**Flags**

- `--min-freq N` — minimum occurrence count to include (default 3)
- `--lang LANG` — assume this language for the corpus (default: auto-detect per file)
- `--exclude PATH` — path to TBX file; exclude terms already defined
- `--script SCRIPT` — filter results to a specific script: `latin`, `hebrew`, `cyrillic`, `arabic`, `any` (default `any`)

**Behavior**

Extracts candidate terms using three heuristics:

1. **Capitalized phrases.** Sequences of capitalized words not at sentence start.
2. **Foreign script tokens.** Tokens whose script differs from the surrounding text — particularly useful for Hebrew terms in Spanish source.
3. **High-frequency noun-like tokens.** Frequency-based with a basic stoplist per language.

Results are aggregated across all input files. NER-grade quality is not the goal; this is a triage list for human curation.

**Output**

```json
{
  "ok": true,
  "candidates": [
    {
      "term": "צמצום",
      "script": "hebrew",
      "frequency": 47,
      "sample_contexts": [
        "...la noción de צמצום aparece...",
        "...el concepto de צמצום..."
      ]
    },
    {
      "term": "Razón Histórica",
      "script": "latin",
      "frequency": 12,
      "sample_contexts": ["..."]
    }
  ]
}
```

Exit 0.

---

## Write commands

Write commands mutate the TBX file in place. All writes are atomic at the
file level (write-temp + rename); a write either fully succeeds or leaves the
file untouched. All writes always emit canonical DCT regardless of the input
file's style.

### Common write affordances

Every write command accepts:

- `--dry-run` — perform validation and produce the final-state preview without modifying the file. Output shows the affected concept(s) as they would appear after the write.
- `--transaction` — append a `<transacGrp>` with `transactionType=modification` and `date=<now in ISO 8601>` to the touched element(s).
- `--author NAME` (or `TERMINOLOGY_AUTHOR` env) — supplies the `responsibility` value for the transaction record. Ignored if `--transaction` is not set. If `--transaction` is set without an author, the transaction is written without `responsibility`.
- `--format json|text` — same semantics as on read commands.

Writes accept input in three layered forms:

1. **Flags** — for simple cases (e.g. `--id`, `--lang`, `--term`, `--status`).
2. **`--format json` payload on stdin** — for any case where flags don't suffice (e.g. setting `definitions`, `notes`, multiple terms in one call). The payload schema matches the JSON shape emitted by `lookup --fields` for the same concept, so a read-modify-write round-trip is straightforward.
3. **`--format tbx` fragment on stdin** — accepts a raw `<conceptEntry>` (or list thereof) in either DCT or DCA style. Always normalized to DCT on disk.

Write commands fail with structured errors (stderr JSON) on:

- `code: duplicate_id` — `concept add` for an ID that already exists. Hint suggests `concept update`.
- `code: dangling_crossref` — a write introduces an IDREF that does not resolve. The error lists the unresolved references.
- `code: invalid_picklist` — a picklist value not in the accepted set.
- `code: invalid_id` — the supplied ID contains rejected characters (see Input hardening).
- `code: not_found` — `concept update` / `remove` for a missing ID.

### `terminology concept add`

Create a new concept.

**Synopsis**

```
terminology concept add [--id ID] [--subject-field FIELD] [--lang LANG --term TERM ...]
                        [--dry-run] [--transaction] [--author NAME]
```

If `--id` is omitted, an ID is derived deterministically from the canonical
language's preferred term (see "Concept IDs"). If the supplied or derived ID
already exists, the command fails with `duplicate_id`.

For complex concepts (multiple terms per language, definitions, notes, cross-references), pipe a JSON payload to stdin:

```bash
terminology concept add < concept.json
```

**Output**: the created concept in `lookup --full` shape.

### `terminology concept update ID`

Modify an existing concept. The `--id` is positional and the underlying
concept ID never changes — this is intentional, so renaming a preferred term
does not silently churn IDs.

**Synopsis**

```
terminology concept update ID [--subject-field FIELD] [--lang LANG --term TERM ...]
                              (--merge | --replace)
                              [--dry-run] [--transaction] [--author NAME]
```

`--merge` and `--replace` are mutually exclusive; exactly one must be
supplied (passing both, or neither, is a usage error and exits `2`).

- `--merge` — supplied fields/languages are merged with the existing concept; unspecified data is preserved.
- `--replace` — supplied data replaces the entire concept content (except its ID).

### `terminology concept remove ID`

Delete a concept.

**Synopsis**

```
terminology concept remove ID [--dry-run] [--transaction] [--author NAME]
                              [--force]
```

Fails with `dangling_crossref` if any other concept's `crossReference`
targets `ID`, unless `--force` is set (which then leaves the dangling
references in the file — `validate` will surface them).

### `terminology term add ID`

Add a `<termSec>` to an existing concept's `<langSec>` (creating the
`<langSec>` if needed).

**Synopsis**

```
terminology term add ID --lang LANG --term TERM
                        [--status preferredTerm-admn-sts|admittedTerm-admn-sts|deprecatedTerm-admn-sts|supersededTerm-admn-sts]
                        [--part-of-speech POS] [--register REG] ...
                        [--dry-run] [--transaction] [--author NAME]
```

Picklist values are validated against the dialect's accepted set; legacy bare
forms (`preferredTerm`, etc.) are accepted and normalized to the suffixed
form on emit.

### `terminology term deprecate ID --lang LANG --term TERM`

Convenience operation: set an existing term's `administrativeStatus` to
`deprecatedTerm-admn-sts`. Errors with `not_found` if the term doesn't exist
in the given language section.

### `terminology apply --file PAYLOAD`

Bulk declarative write. The payload describes desired state for one or more
concepts; `apply` reconciles by computing the minimal set of add/update
operations.

**Synopsis**

```
terminology apply --file PAYLOAD [--prune] [--dry-run] [--transaction] [--author NAME]
```

- `--file PAYLOAD` (short alias `-f`) — path to a JSON file, or `-` for stdin. The file may also be `--format tbx` for a raw TBX fragment.
- `--prune` — concepts in the existing TBX that are _not_ present in the payload are removed. Without `--prune`, `apply` only adds and updates.

**Output**: an envelope summarizing the reconciliation.

```json
{
  "ok": true,
  "applied": {
    "added": ["tzimtzum"],
    "updated": ["razon-historica"],
    "removed": [],
    "unchanged": ["malkhut", "binah"]
  },
  "warnings": []
}
```

Exit 0 on success, 65 if any concept failed validation (none applied —
the file is left untouched), 2 on usage error, 3 on I/O error.

---

## Concept IDs

Concept IDs are the stable handle agents (and humans) use to refer to a
concept across runs. The rules:

1. **Default derivation.** When `--id` is omitted, the ID is derived from the canonical-language preferred term:
   - NFKD-normalize the term, drop combining marks.
   - Lowercase under the Unicode default casefold.
   - Replace any sequence of non-`[a-z0-9]` characters with a single `-`.
   - Trim leading/trailing `-`.
   - Truncate to 64 codepoints (cut on a `-` boundary if possible).
   - If empty after the above (e.g. all non-Latin script with no romanization), fail with `invalid_id` and require an explicit `--id`.

2. **Canonical language.** Resolved in order: `--canonical-lang` flag, `TERMINOLOGY_CANONICAL_LANG` env, then `en`, then the first language section in document order.

3. **Override.** `--id` is accepted on `concept add` (and inside JSON payloads via the `concept_id` field). Validated against the input-hardening rules.

4. **Stability.** Once written, an ID never changes implicitly. Renaming the preferred term via `concept update` or `term add` does **not** re-derive the ID. To rename an ID, the caller must `concept remove` + `concept add` deliberately (and update any cross-references).

5. **Predictability.** The derivation is deterministic and exposed via the schema, so an agent can compute the expected ID before calling `concept add`.

---

## `terminology schema`

Emit a JSON description of the complete CLI surface — every command,
flag, envelope shape, and error code. The output is **computed
reflectively at runtime** from the live `urfave/cli/v3` command tree,
the output struct types in `internal/output`, and the `terr` sentinel
registry. There is no embedded `schema.json` file; Go code is the
canonical source of truth (see
[`schema-source-of-truth.md`](adr/schema-source-of-truth.md)).

**Synopsis**

```
terminology schema [--command NAME]
```

- `--command NAME` — restrict the output to one command's entry.

The schema describes every command, its flags (with types, defaults, picklist
values, hardening rules), its expected output shape, and the structured error
codes it can emit. Agents are expected to call this once at session start in
lieu of pre-stuffed documentation.

The output is generated **reflectively** at runtime by walking the live
urfave command tree, reflecting over the Go output struct types in
`internal/output`, and enumerating the `terr` sentinel registry. Go
code is the source of truth (see
[`schema-source-of-truth.md`](adr/schema-source-of-truth.md));
[`docs/specs/target.schema.json`](specs/target.schema.json) exists
transiently as a scaffolding artifact and is deleted once the Go
scaffolding is in place. Two conventions enforced by the code are worth
surfacing here because they shape every command's flag list:

- **Long-form-canonical flags.** Every flag is identified by its long form
  (`--file`, `--lang`, `--term`, ...). Short forms exist only as aliases on a
  long-form flag (e.g. `apply` exposes `-f` as the alias for `--file`). The
  schema rejects short-only flag declarations.
- **Typed flag values.** Flags carry a semantic type (`path`, `lang`, `id`,
  `term`, `enum`, ...) rather than just a Go primitive. The type drives both
  the hardening rules applied and, for `enum`, the picklist reference. The
  recognized types are listed in the schema.

### `internal/tbx`

```go
package tbx

type Glossary struct {
    Style    Style       // dct or dca, as read
    Concepts []Concept
}

type Style int

const (
    StyleDCT Style = iota
    StyleDCA
)

type Concept struct {
    ID             string
    SubjectField   string                  // Min
    Definitions    []NoteText              // Basic, concept-level
    CrossRefs      []CrossRef              // Basic, IDREF + free-text
    ExternalRefs   []string                // Min, URLs
    Graphics       []string                // Basic, xGraphic URLs
    Sources        []string                // Basic
    CustomerSubset string                  // Min
    ProjectSubset  string                  // Basic
    Transactions   []Transaction
    Notes          []string
    Languages      map[string]LangSection  // key: BCP 47 tag
}

type LangSection struct {
    Lang        string
    Definitions []NoteText                 // Basic, langSec-level
    Sources     []string
    Terms       []Term
}

type Term struct {
    Surface              string
    AdministrativeStatus Status            // Min picklist
    PartOfSpeech         string            // Min picklist
    GrammaticalGender    string            // Basic picklist
    GrammaticalNumber    string            // Linguist picklist
    Register             string            // Linguist picklist; legacy usageRegister normalized
    TermType             string            // Basic picklist
    TermLocation         string            // Basic picklist
    GeographicalUsage    string            // Basic, free-form
    Contexts             []NoteText        // Basic
    TransferComment      string            // Linguist
    Reading              string            // Linguist
    ReadingNote          string            // Linguist
    Sources              []string          // Basic
    CustomerSubset       string            // Min
    ProjectSubset        string            // Basic
    ExternalRefs         []string          // Min, URLs
    CrossRefs            []CrossRef        // Basic
    Transactions         []Transaction
    Notes                []string
}

type Status int

const (
    StatusUnspecified Status = iota
    StatusPreferred                        // preferredTerm-admn-sts
    StatusAdmitted                         // admittedTerm-admn-sts
    StatusDeprecated                       // deprecatedTerm-admn-sts
    StatusSuperseded                       // supersededTerm-admn-sts
)

type CrossRef struct {
    Target string                          // IDREF or free-text
    Label  string
}

type Transaction struct {
    Type           string                  // origination | modification
    Date           string                  // ISO 8601
    Responsibility string
}

// NoteText is plain text with the inline runs preserved as raw XML for
// faithful round-trip; matching treats inline content as plain text.
type NoteText struct {
    Plain string
    Raw   string
}

func Load(path string) (*Glossary, error)
func (g *Glossary) Validate(strict bool) []Warning
func (g *Glossary) Lookup(term string, lang string) []Concept
func (g *Glossary) Preferred(conceptID, lang string) (Term, bool)
func (g *Glossary) Deprecated(conceptID, lang string) []Term
```

Parses with stdlib `encoding/xml`. The parser branches on `<tbx>` `@style`:
DCT reads namespaced elements directly; DCA reads generic `<descrip>`,
`<termNote>`, `<admin>`, `<ref>`, `<xref>` and dispatches on `@type`.
`administrativeStatus` values are normalized on read: legacy bare forms
(`preferredTerm`, `admittedTerm`, `deprecatedTerm`, `supersededTerm`) are
accepted and mapped to the suffixed forms. Legacy `usageRegister` is mapped
to `register`. Unknown elements are tolerated; under `--strict` they are
reported as warnings.

### `internal/scan`

```go
package scan

type Match struct {
    ConceptID string
    Term      string
    Lang      string
    Line      int
    Column    int
    Context   string
}

func Scan(text string, glossary *tbx.Glossary, lang string) []Match
```

Uses **Aho-Corasick multi-pattern matching** via `github.com/cloudflare/ahocorasick` over a pre-normalized canonical form of the corpus (NFC, case-folded, niqqud-stripped, whitespace-collapsed). One scan matches every glossary term in a single pass instead of `O(N·M)` regex-per-term. Matches are validated against the original text for word-boundary correctness (`\p{L}`, `\p{N}` as word characters) and translated back through a position map for line/column reporting.

Critical detail: word boundaries across scripts. The boundary contract `(^|[^\p{L}\p{N}])TERM([^\p{L}\p{N}]|$)` works for Spanish, English, and Hebrew alike — and crucially handles the case where a Hebrew term appears immediately adjacent to Spanish punctuation. See [`docs/specs/005-matcher.md`](specs/005-matcher.md) for the full pipeline (canonical normalization, position mapping, boundary post-filter, longest-match-at-same-start, status tagging).

### `internal/check`

```go
package check

type Violation struct {
    Type           string  // "missing" or "forbidden_variant"
    ConceptID      string
    SourceTerm     string  // for "missing"
    ExpectedTarget string  // for "missing"
    Variant        string  // for "forbidden_variant"
    Line           int     // for "forbidden_variant"
    Column         int     // for "forbidden_variant"
    SourceOccurrences int
    TargetOccurrences int
}

func Check(src, tgt string, glossary *tbx.Glossary,
    srcLang, tgtLang string, strict bool) ([]Violation, []Warning)
```

Composes `tbx.Glossary` and `scan` to do the actual verification logic.

### `internal/extract`

```go
package extract

type Candidate struct {
    Term            string
    Script          string  // "latin", "hebrew", etc.
    Frequency       int
    SampleContexts  []string
}

func Extract(files []string, opts ExtractOptions) []Candidate
```

Uses `golang.org/x/text/unicode/runenames` or direct Unicode block matching to identify script per token.

### `internal/output`

Helpers for emitting JSON or text format consistently across commands.

- **Success envelope (stdout):** `{ "ok": true, ...payload }`. Payload keys are command-specific.
- **Error envelope (stderr):** `{ "ok": false, "error": { "code": "...", "message": "...", "hint": "..." } }` when `--format=json`; human-readable text when `--format=text`.
- **Field masks:** `fields.go` parses the flat `a,b,c.d` syntax (with `*` for map wildcards) and applies it to any output struct before serialization. Unknown paths return `code: invalid_field`.
- **Text mode:** `output.go` renders indented hierarchical blocks with `✓` / `✗` status glyphs paired with text labels so a terminal that strips Unicode still produces parseable output.

## Dependencies

Minimize aggressively:

| Package                                       | Purpose                                                            |
| --------------------------------------------- | ------------------------------------------------------------------ |
| stdlib `encoding/xml`                         | TBX parsing (hand-rolled emitter on top — see [E2](specs/002-domain-and-io.md)) |
| stdlib `encoding/json`                        | Envelope serialization                                             |
| `github.com/urfave/cli/v3`                    | CLI parsing, subcommand dispatch, help                             |
| `golang.org/x/text/unicode/...`               | Script detection, NFKD normalization, BCP 47 well-formedness       |
| `golang.org/x/term`                           | TTY detection for log handler selection                            |
| `github.com/rogpeppe/go-internal/lockedfile`  | Cross-process advisory lock around TBX writes ([E2](specs/002-domain-and-io.md)) |
| `github.com/yuin/goldmark`                    | Markdown parsing — code-region skipping for `scan`/`check`/`extract` |
| `github.com/cloudflare/ahocorasick`           | Multi-pattern matching for the term matcher ([E5](specs/005-matcher.md)) |

That's the full list. The dependency posture is stdlib + Go-team `x/*` packages plus four single-purpose third-party libraries with stable APIs and permissive licenses. See [urfave-cli-reference.md](./urfave-cli-reference.md) for the CLI API surface this project draws on.

## What this gives the agent

### Translation loop (per chapter)

```bash
export TERMINOLOGY_TBX=glossary/terms.tbx

# Before translation
terminology scan source/ch1.md > scan.json
# Agent reads scan.json, looks up unfamiliar terms with `terminology lookup`.
# Agent translates source/ch1.md → target/ch1.md.

# After translation
terminology check source/ch1.md target/ch1.md
# Exit 0 → done. Exit 1 → read violations.

# Each violation embeds the expected target term and a context window,
# so the agent fixes in one pass without a follow-up lookup. Concept IDs
# are stable across runs, letting the agent dedupe and learn.
```

### Curation loop (corpus → glossary)

```bash
export TERMINOLOGY_TBX=glossary/terms.tbx
export TERMINOLOGY_AUTHOR="andre"

# Triage candidates from a corpus
terminology extract source/*.md --exclude "$TERMINOLOGY_TBX" > candidates.json

# Agent (or human) reviews candidates and emits a desired-state payload
terminology apply -f new-concepts.json --dry-run     # preview
terminology apply -f new-concepts.json --transaction # commit

# Confirm the file still parses against the supported subset
terminology validate
```

For granular edits — adding a single term, deprecating a variant — use the
`concept` and `term` verbs instead of `apply`. All writes accept `--dry-run`
to preview the final state without touching the file.

## Out of scope

Explicitly not part of v1:

- **Translation memory.** Academic prose has too little segment-level repetition to justify the storage. If needed later, a separate `terminology tm` subcommand can be added without disturbing the existing surface.
- **Markdown structure verification.** Useful but a different concern. Belongs in a separate tool or a future `terminology structure` subcommand.
- **Status tracking.** The filesystem tells you which chapters are done. No need for a status file.
- **DCA on write.** The tool reads DCA for interoperability but always emits canonical DCT. Round-tripping DCA is a future concern.
- **NDJSON streaming.** All commands emit a single JSON document. Stream-friendly output for very large corpora is a v2 consideration.
- **Per-command skill files.** `terminology schema` is the agent's contract for v1; narrative skill files (`docs/skills/<command>.md`) are deferred until real friction shows up.
- **Multi-target-language projects.** One source language, one target language at a time, passed via flags.

These are all incremental additions if needed. The architecture leaves room for them without committing to them up front.
