# Write command details

Detailed semantics for write operations: concept add/update/remove, term add/deprecate, and apply.

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

### 2. JSON stdin

```json
{
  "concept_id": "algorithm",
  "subject_field": "computer science",
  "cross_refs": [{ "target": "data-structure", "label": "related" }],
  "languages": {
    "en": {
      "preferred": {
        "term": "algorithm",
        "administrative_status": "preferredTerm-admn-sts"
      }
    },
    "he": {
      "preferred": {
        "term": "אלגוריתם",
        "administrative_status": "preferredTerm-admn-sts"
      }
    }
  }
}
```

Pipe to the command: `echo '...' | terminology concept add --tbx glossary.tbx`

### 3. TBX fragment stdin

Accepted forms:

- Bare `<conceptEntry>...</conceptEntry>`
- `<conceptEntryList><conceptEntry>...</conceptEntry>...</conceptEntryList>`

Rejected forms:

- Full `<tbx>` document → exit 65, `invalid_input`

Pipe to the command: `cat fragment.xml | terminology concept add --tbx glossary.tbx`

## Concept ID derivation

When `--id` is not provided, the ID is derived from the preferred term:

1.  NFKD normalization
2.  Case folding
3.  Keep only `[a-z0-9]`
4.  Hyphen-join remaining segments
5.  Truncate to 64 characters

`--canonical-lang` selects which language's preferred term derives the ID (default: first provided language).

Edge case: Hebrew-only input with no Latin fallback → `invalid_id` (exit 65).

Override: `--id` flag bypasses derivation entirely.

**Concept ID never changes after creation.** Updating the preferred term or replacing concept content does not alter the ID.

## Merge vs Replace (concept update)

Exactly one of `--merge` or `--replace` is required (mutex violation → exit 2).

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

Transaction records are stripped during concept equality comparison for apply, so prior writes do not cause false "updated" classifications.

## Pre-write validation

Before any file modification (even without `--dry-run`), the full validation pipeline runs on the in-memory result:

- `duplicate_id` — concept add with existing ID
- `dangling_crossref` — remove would break inbound cross-reference
- `invalid_id` — derived ID is empty
- `invalid_input` — malformed JSON or full TBX document

**The file is never written if validation fails.** The atomic write pattern: validate in memory → write to temp file → rename over original.

## Dry-run

`--dry-run` / `-n` shows the result envelope as if the write happened:

```json
{"schema_version":1, "ok":true, "result":{...}}
```

The glossary file is not modified — checksum verifiable before/after.

Pre-write validation still runs, so `--dry-run` can surface errors that would prevent the actual write.

## Dangling crossref protection

`concept remove ID` refuses if other concepts reference the target:

```json
{ "ok": false, "error": { "code": "dangling_crossref", "message": "..." } }
```

`--force` overrides: concept removed, and the orphaned cross-reference becomes a warning in the file.

Apply with `--prune` has the same protection: absent concepts that are cross-referenced trigger `dangling_crossref` (exit 65), file untouched.

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

Idempotent: running the same payload twice produces all `unchanged` on the second run.

`--update` is wholesale replace — the payload is authoritative. Omitted fields and language sections are dropped, not merged. Use `concept update --merge` for additive updates.

## JSON payload format (apply)

```json
{
  "concepts": [
    {
      "concept_id": "algorithm",
      "subject_field": "computer science",
      "cross_refs": [{ "target": "data-structure", "label": "related" }],
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

Field ordering is irrelevant — the canonical writer normalizes output.

## TBX fragment payload format (apply)

```xml
<conceptEntryList>
  <conceptEntry id="algorithm">
    <langSec xml:lang="en">
      <termSec>
        <term>algorithm</term>
        <termNote type="administrativeStatus">preferredTerm-admn-sts</termNote>
      </termSec>
    </langSec>
  </conceptEntry>
</conceptEntryList>
```

Same fragment format accepted by concept add/update.

## File locking

Writes acquire an advisory lock at `${TBX}.lock` (fcntl-based):

- Non-blocking: fails fast if lock is held
- `tbx_locked` error (exit 3)
- Lock is released on process exit

Not reliably testable on macOS (lacks `flock(1)` from util-linux).
