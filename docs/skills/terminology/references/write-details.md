# Write command details

Detailed semantics for write operations: concept add/update/remove, term
add/deprecate, and apply.

> Corrected/expanded revision. The upstream skill copy
> (`~/.claude/skills/terminology/references/write-details.md`) documents only the
> four term-group keys in the JSON payload and omits every concept-level and
> term-level enrichment key (`definitions`, `notes`, `reading`, `reading_note`,
> `contexts`, `part_of_speech`, `register`, `grammatical_gender`,
> `cross_refs`). All additions below were verified against CLI build `b675e4d`
> by writing payloads and inspecting the resulting TBX.

## Input layering

Write commands accept three input modes, auto-detected:

### 1. Flags

```sh
terminology concept add --tbx glossary.tbx \
    --lang en --term "algorithm" \
    --subject-field "computer science" \
    --status preferredTerm-admn-sts \
    --part-of-speech noun
```

Available term-level flags:

| Flag                   | Values                                                                                                                                                      |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--term`               | Surface form (free text)                                                                                                                                    |
| `--status`             | `preferredTerm-admn-sts`, `admittedTerm-admn-sts`, `deprecatedTerm-admn-sts`, `supersededTerm-admn-sts` (+ legacy bare forms)                               |
| `--part-of-speech`     | `noun`, `verb`, `adjective`, `adverb`, `other`                                                                                                              |
| `--register`           | `colloquialRegister`, `neutralRegister`, `technicalRegister`, `in-houseRegister`, `bench-levelRegister`, `slangRegister`, `vulgarRegister`, `usageRegister` |
| `--grammatical-gender` | `masculine`, `feminine`, `neuter`, `other`                                                                                                                  |

The flag interface cannot set definitions, notes, readings, contexts, or
cross-references. Use JSON stdin (below) for any of those.

### 2. JSON stdin

The JSON payload is far richer than the upstream doc shows. The complete,
verified shape:

```json
{
  "concept_id": "algorithm",
  "subject_field": "computer science",
  "definitions": ["A finite sequence of well-defined instructions."],
  "notes": ["Concept-level free-text note."],
  "cross_refs": [{ "target": "data-structure", "label": "related" }],
  "languages": {
    "en": {
      "preferred": {
        "term": "algorithm",
        "administrative_status": "preferredTerm-admn-sts",
        "part_of_speech": "noun",
        "register": "technicalRegister",
        "grammatical_gender": "neuter",
        "reading": "ˈæl.ɡə.rɪð.əm",
        "reading_note": "IPA",
        "contexts": ["The sorting algorithm ran in linearithmic time."],
        "notes": ["Term-level free-text note."]
      },
      "admitted": [
        {
          "term": "algo",
          "administrative_status": "admittedTerm-admn-sts",
          "register": "colloquialRegister"
        }
      ],
      "deprecated": [
        { "term": "recipe", "administrative_status": "deprecatedTerm-admn-sts" }
      ],
      "superseded": [
        { "term": "procedure", "administrative_status": "supersededTerm-admn-sts" }
      ]
    }
  }
}
```

Pipe to the command: `echo '...' | terminology concept add --tbx glossary.tbx`

#### Concept-level keys

| Key             | Shape                                | Emitted element                     | Notes                                                                         |
| --------------- | ------------------------------------ | ----------------------------------- | ----------------------------------------------------------------------------- |
| `concept_id`    | string                               | `id` attribute                      | Overrides ID derivation (see below). Required by `apply`.                     |
| `subject_field` | string                               | `min:subjectField`                  | Single value.                                                                 |
| `definitions`   | array of **strings**                 | one `basic:definition` per string   | **Language-neutral.** Use `languages.<lang>.definitions` for per-language.    |
| `notes`         | array of **strings**                 | one `note` per string               | Free text.                                                                    |
| `cross_refs`    | array of `{ "target", "label" }`     | cross-reference                     | `target` must resolve to an existing concept ID, else `unresolved_crossref`.  |

#### Language section keys

Each language section uses one optional definitions key plus four optional
term-group keys:

| Key           | Shape                | Meaning                                                                                             |
| ------------- | -------------------- | -------------------------------------------------------------------------------------------------- |
| `definitions` | array of **strings** | Language-scoped definitions; one `basic:definition` per string, emitted inside this `langSec`.      |
| `preferred`   | object or null       | The single preferred term (`preferredTerm-admn-sts`). At most one per language section.             |
| `admitted`    | array of object      | Tolerated variants (`admittedTerm-admn-sts`); satisfy `check` but warn (or fail under `--strict`).  |
| `deprecated`  | array of object      | Forbidden variants (`deprecatedTerm-admn-sts`); flagged as `forbidden_variant` by `check`.          |
| `superseded`  | array of object      | Historical variants (`supersededTerm-admn-sts`); treated like `deprecated` for verification.        |

All keys are optional and omitted (not null) when empty. Per-language
`definitions` coexist with the concept-level (language-neutral) `definitions`:
carry an English and a Portuguese definition as
`languages.en.definitions: ["..."]` and `languages.pt.definitions: ["..."]`,
each emitted as a `basic:definition` scoped to its `langSec`.

#### Term object keys

Every term object (in any of the four groups) accepts:

| Key                     | Shape                | Emitted element                              |
| ----------------------- | -------------------- | -------------------------------------------- |
| `term`                  | string               | `term`                                       |
| `administrative_status` | string               | `min:administrativeStatus`                   |
| `part_of_speech`        | string               | `min:partOfSpeech`                           |
| `register`              | string               | `ling:register`                              |
| `grammatical_gender`    | string               | grammatical-gender element                   |
| `reading`               | string               | `ling:reading` (inside `adminGrp`)           |
| `reading_note`          | string               | `ling:readingNote` (inside `adminGrp`)       |
| `contexts`              | array of **strings** | one `basic:context` per string               |
| `notes`                 | array of **strings** | one `note` per string                        |

`reading` / `reading_note` are the natural home for phonetic or transliterated
forms. For CJK glossaries this cleanly holds three representations in one
language section: the term is the ideographic form, `reading` the phonetic
reading, and `reading_note` a romanization.

Verified element output for a fully populated term:

```xml
<conceptEntry id="probe">
  <min:subjectField>cs</min:subjectField>
  <basic:definition>def one</basic:definition>
  <basic:definition>def two</basic:definition>
  <note>concept note</note>
  <langSec xml:lang="en">
    <termSec>
      <term>probe</term>
      <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
      <min:partOfSpeech>noun</min:partOfSpeech>
      <ling:register>technicalRegister</ling:register>
      <basic:context>used in a sentence here</basic:context>
      <adminGrp>
        <ling:reading>proub</ling:reading>
        <ling:readingNote>IPA</ling:readingNote>
      </adminGrp>
      <note>term note</note>
    </termSec>
  </langSec>
</conceptEntry>
```

#### JSON payload limitations (verified)

- **`definitions` and `notes` are arrays of strings, not objects.** The
  intuitive `[{ "text": "...", "lang": "en" }]` form is rejected with an
  `invalid_input` error that names the offending field and the expected vs
  actual type (e.g. `definitions: expected string, got object`). The error
  envelope's `error.details` carries the structured `path`, `expected`,
  `actual`, and `kind` (`type_mismatch` | `unknown_field` | `syntax`) so a
  malformed payload can be fixed in one pass. For `apply`, the path includes
  the array context (e.g. `concepts.definitions`).
- **Concept-level fields not in the JSON shape.** Concept-level `graphics`,
  `customer_subset`, and `project_subset`, and langSec-level `sources`, are not
  represented in the canonical write/read shape and do not survive a
  round-trip through it. Edit those with a `--format tbx` fragment instead.
- **Empty arrays are fine but pointless** - omit rather than send `[]`.

#### Multiple readings (workaround)

There is no repeated `reading` / `reading_note` on a single term: both are
scalar strings holding one reading each. To record an alternate reading for the
same surface form, add it as a separate **admitted** term in the same language
section (it will be surfaced by `lookup`, and by the forthcoming search/show/
export read commands), or stash the extra reading in `reading_note` or a
`notes` entry. This keeps the canonical shape simple while still capturing the
alternate reading for retrieval.

### 3. TBX fragment stdin

Accepted forms:

- Bare `<conceptEntry>...</conceptEntry>`
- `<conceptEntryList><conceptEntry>...</conceptEntry>...</conceptEntryList>`

Rejected forms:

- Full `<tbx>` document -> exit 65, `invalid_input`

Pipe to the command: `cat fragment.xml | terminology concept add --tbx glossary.tbx`

Both encoding styles are accepted and round-trip in full:

- **DCT (namespaced)** - `<basic:definition>`, `<min:administrativeStatus>`,
  `<ling:reading>`, and the other namespaced elements.
- **DCA (generic carriers)** - `<descrip type="definition">`,
  `<termNote type="administrativeStatus">`, `<admin type="...">`, and so on.

The style is auto-detected from the fragment's elements; definitions,
administrative statuses, readings, and notes are all preserved regardless of
style. Input is always normalized to canonical DCT on disk.

> **Fails closed on unsupported input.** Any element the reader cannot map is
> never silently dropped. The command fails with exit 65 and an
> `invalid_input` error whose message names the offending element(s), leaving
> the glossary file untouched.

## Concept ID derivation

When `--id` (flag) or `concept_id` (JSON) is not provided, the ID is derived
from the preferred term:

1. NFKD normalization
2. Case folding
3. Keep only `[a-z0-9]`
4. Hyphen-join remaining segments
5. Truncate to 64 characters

`--canonical-lang` selects which language's preferred term derives the ID
(default: first provided language).

Edge case: a preferred term with no Latin/numeric characters (for example a
purely CJK or Hebrew term) derives to an empty string -> `invalid_id`
(exit 65). Supply an explicit `concept_id` / `--id` in that case. This is the
common situation for non-Latin glossaries and is the main reason to always set
`concept_id` explicitly in `apply` payloads.

Override: `--id` / `concept_id` bypasses derivation entirely.

**Concept ID never changes after creation.** Updating the preferred term or
replacing concept content does not alter the ID.

## Merge vs Replace (concept update)

Exactly one of `--merge` or `--replace` is required (mutex violation -> exit 2).

### --merge

- Adds new language sections not present in existing concept
- Preserves language sections absent from the update
- Overlays existing terms matched by (surface form, status)
- Use for: adding a translation to an existing concept

### --replace

- Replaces entire concept content except concept ID
- Language sections omitted from the update are **dropped**
- Use for: full correction of a concept

## Transaction records

`--transaction --author NAME` adds a `<transacGrp>` element:

```xml
<transacGrp>
  <transac type="modification">...</transac>
  <Date>2026-05-27</Date>
  <basic:responsibility>Author Name</basic:responsibility>
</transacGrp>
```

Placement:

- Concept commands: at concept level
- Term commands: at termSec level
- Apply: on added and updated concepts only (unchanged concepts do NOT get records)

Behavior:

- `--transaction` without `--author`: omits `<basic:responsibility>`, WARN on stderr
- `--author` without `--transaction`: silently ignored
- `TERMINOLOGY_AUTHOR` env var as fallback for `--author`

Transaction records are stripped during concept equality comparison for apply,
so prior writes do not cause false "updated" classifications.

## Pre-write validation

Before any file modification (even without `--dry-run`), the full validation
pipeline runs on the in-memory result:

- `duplicate_id` - concept add with existing ID
- `dangling_crossref` - remove would break inbound cross-reference
- `unresolved_crossref` - a `cross_refs` target does not match any concept ID
- `invalid_id` - derived ID is empty
- `invalid_input` - malformed JSON or full TBX document

**The file is never written if validation fails.** The atomic write pattern:
validate in memory -> write to temp file -> rename over original.

## Dry-run

`--dry-run` / `-n` shows the result envelope as if the write happened. The
glossary file is not modified (checksum verifiable before/after). Pre-write
validation still runs, so `--dry-run` surfaces errors that would block the
real write.

Note: the result envelope echoed by a write reflects the parsed input, not the
file. To read a write back from disk, use `show <id>` (single concept),
`export` (whole glossary), or `lookup <term>` -- all three emit the canonical
concept shape, so definitions, readings, notes, contexts, and non-preferred
variants are fully readable.

## Apply patch model

Apply computes the minimal set of operations:

| Classification | Condition                                                                  |
| -------------- | -------------------------------------------------------------------------- |
| `added`        | Concept in payload but not in glossary                                     |
| `updated`      | Concept in both but different after canonicalization (transacGrp stripped) |
| `unchanged`    | Concept in both and byte-identical after canonicalization                  |
| `removed`      | Concept in glossary but not in payload (only with `--prune`)               |

Result envelope:

```json
{
  "schema_version": 1,
  "ok": true,
  "applied": {
    "added": ["sorted-ids"],
    "updated": ["sorted-ids"],
    "removed": ["sorted-ids"],
    "unchanged": ["sorted-ids"]
  }
}
```

All four arrays are always present (never null), sorted ASCII byte order.

Idempotent: running the same payload twice produces all `unchanged` on the
second run. The `apply` payload accepts the same rich concept shape documented
under JSON stdin above (`definitions`, `notes`, `reading`, `reading_note`,
`contexts`, etc.), wrapped in a top-level `{ "concepts": [ ... ] }`.

`--update` is wholesale replace - the payload is authoritative. Omitted fields
and language sections are dropped, not merged. Use `concept update --merge` for
additive updates.

## JSON payload format (apply)

```json
{
  "concepts": [
    {
      "concept_id": "algorithm",
      "subject_field": "computer science",
      "definitions": ["A finite sequence of well-defined instructions."],
      "languages": {
        "en": {
          "preferred": {
            "term": "algorithm",
            "administrative_status": "preferredTerm-admn-sts"
          }
        }
      }
    }
  ]
}
```

Field ordering is irrelevant - the canonical writer normalizes output.

## File locking

Writes acquire an advisory lock at `${TBX}.lock` (fcntl-based):

- Non-blocking: fails fast if lock is held
- `tbx_locked` error (exit 3)
- Lock is released on process exit
