# E7 — Write commands: `concept` & `term`

> **Status**: APPROVED. The five granular mutation commands on top of
> [E2 the writer](002-domain-and-io.md) (atomic rename, canonical DCT,
> deterministic serialization). Bulk declarative writes (`apply`) are
> [E8](008-apply.md).

## Scope

- `concept add` — create a new `<conceptEntry>`.
- `concept update ID` — modify an existing concept (merge or replace).
- `concept remove ID` — delete a concept.
- `term add ID` — append a `<termSec>` to a concept's `<langSec>`
  (creating the `<langSec>` if needed).
- `term deprecate ID --lang LANG --term TERM` — convenience: set
  `administrativeStatus = deprecatedTerm-admn-sts`.

## Shared write infrastructure

Lives in `internal/write/`:

- `write.go` — read → modify → write pipeline (load glossary, dispatch
  to per-command mutator, save).
- `id.go` — concept-ID derivation (cli-design.md §"Concept IDs").
- `transaction.go` — `<transacGrp>` emission when `--transaction` is set.
- `dryrun.go` — final-state preview without rename.

Per-command files: `concept_add.go`, `concept_update.go`, …, each a
single `Mutate(*tbx.Glossary, Args) error`-style function.

## Common affordances

(Per cli-design.md §"Common write affordances")

- `--dry-run` — performs validation and prints the final-state preview;
  does not call rename.
- `--transaction` — appends a `<transacGrp>` with
  `transactionType=modification` + `date=<now ISO 8601 UTC>` + optional
  `responsibility=<author>`.
- `--author NAME` (or `TERMINOLOGY_AUTHOR`) — supplies `responsibility`;
  silently ignored if `--transaction` is not set.

### Author resolution

`--author` and `TERMINOLOGY_AUTHOR` are wired via urfave's `Sources`
(flag wins over env). When `--transaction` is set but neither resolves
to a value:

- The `<transacGrp>` is emitted **without** `responsibility`. Matches
  cli-design.md.
- A **WARN-level** slog record is emitted to stderr noting the
  attribution gap. Operators see the signal without it polluting the
  JSON envelope. See [logging](../adr/logging.md).

## Input layering

Per cli-design.md:

1. **Flags** for simple cases.
2. **JSON payload on stdin** when flags don't suffice.
3. **TBX fragment on stdin** with `--format tbx`.

### JSON payload shape

The payload structurally matches the same Go types used by
`lookup --fields` output. There is no separate JSON Schema doc — the
struct in `internal/output/types.go` *is* the schema, validated on
stdin by the same reflective machinery that powers `terminology schema`
(per [schema-source-of-truth](../adr/schema-source-of-truth.md)).

Discoverable: `terminology schema --command 'concept add'` returns the
expected envelope shape derived from the live struct. Unknown JSON
fields produce an `invalid_field` error envelope (exit `2`).

Consequence: round-trip read → modify → write is structurally trivial,
because the read-side output type and the write-side input type are
the same Go struct.

### TBX fragment input boundaries

Accepted shapes for `--format tbx`:

- **Singular** — a bare `<conceptEntry>`.
- **Plural** — `<conceptEntryList><conceptEntry/>...</conceptEntryList>`,
  a non-TBX wrapper element. No `<tbx>` or `<tbxHeader>` needed; we're
  not loading a full file.

A full `<tbx>` document on stdin is **rejected** with an
`invalid_input` error envelope. For full-file ingest, use
`apply --file`.

## `concept update --merge` semantics

When `--merge` is set, the payload overlays the existing concept:

- **`Languages`** (map) — payload keys overlay existing keys; absent
  keys preserved.
- **`Terms`** (slice within each `LangSection`) — matched by
  `(Surface, AdministrativeStatus)`. Equal pair → merge fields of the
  matched term; otherwise the payload entry is appended.
- **`Definitions`, `CrossRefs`, `Sources`, `Notes`** — replace if the
  payload supplies the field; preserve if absent.

The replace-if-present rule for free-form lists avoids the "append
forever" pitfall where repeated merges accumulate stale definitions.
Term-level granularity is preserved via the `(Surface, status)` key,
because terms have a stable natural key while definitions and notes
don't.

When `--merge` is not set, the payload replaces the concept wholesale
(after which validation runs — see below).

## Error envelope codes (write-only)

Sentinels live in `internal/write/errors.go` and
`internal/tbx/errors.go`; each registered per
[error-handling](../adr/error-handling.md).

- `duplicate_id` — `concept add` for an existing ID.
- `dangling_crossref` — write introduces an unresolved IDREF.
- `invalid_picklist` — picklist value not in the accepted set.
- `invalid_id` — ID fails hardening (control chars, traversal, …).
- `not_found` — `concept update`/`remove`/`term deprecate` for a missing ID.
- `invalid_input` — stdin payload malformed (JSON parse error, unknown
  field, unsupported fragment shape).

All emit on stderr per the standard envelope shape; exit code per
[error-handling](../adr/error-handling.md).

## `concept remove --force` and dangling crossrefs

Default behavior: refuse with `dangling_crossref` if any other concept's
`<crossReference>` targets `ID`.

`--force` overrides the refusal and leaves the references in place;
`validate` will subsequently surface them as `unresolved_crossref`
warnings against each referencing concept. The force behavior is
exercised by an integration test: `remove --force` → `validate` →
expect N `unresolved_crossref` warnings.

## Concept-ID stability

Per cli-design.md §"Concept IDs" rule 4: renaming the preferred term
does **not** re-derive the ID. Both pathways must be covered by tests:

- `concept update ID --preferred new-term` keeps `ID` unchanged.
- `term add ID --status preferred ...` (when replacing the existing
  preferred) keeps `ID` unchanged.

## Picklist normalization on write

Flag input is normalized at the urfave validator layer (decided in
[001 Q3](001-cli-surface-stub.md)). Accepted enums list both modern
and legacy forms; the validator returns the modern form to downstream
code. Reader-side normalization handles file input. The in-memory
model only ever holds the normalized form, so the writer emits
deterministically without re-checking.

Picklist values are sourced from `internal/tbx/picklist.go` —
single source shared with E3 validation.

## Transaction record placement

`<transacGrp>` is appended to the touched node at the appropriate
scope:

- `concept add` / `concept update` (whole-concept) / `concept remove` →
  `<conceptEntry>/<transacGrp>`.
- `concept update` touching a single langSec → `<langSec>/<transacGrp>`.
- `term add` / `term deprecate` → `<termSec>/<transacGrp>`.

Each command knows its mutation scope and attaches the record there.
Removed concepts get the transac record on the wrapping concept just
before serialization — the entry is then dropped, so in practice the
record only surfaces for `add`/`update`/`term *`. Removal records are
not persisted (cli-design.md: transaction records live alongside the
entity they describe; a removed entity has no persistent home).

## Pre-write validation

Every write runs the full E3 validation pipeline on the **in-memory
result** before rename:

1. Well-formedness — by construction (we're emitting fresh XML), but
   we re-parse the serialized bytes to integration-test the emitter.
2. Schema (dialect) tier — every emitted element/attribute is in the
   TBX-Linguist supported set.
3. Semantic tier — concept IDs unique, IDREFs resolve, BCP 47 tags
   well-formed.

A failure aborts before rename: the on-disk file is untouched and the
corresponding error envelope is emitted (`duplicate_id`,
`dangling_crossref`, `validation_error`, etc.). The dry-run preview
also runs full validation, so `--dry-run` is a reliable preflight.

## Concurrency

Writes hold the sibling `${TBX}.lock` advisory lock for the read →
modify → write window, per [E2](002-domain-and-io.md). The lock is
non-blocking — a concurrent write fails fast with `tbx_locked`.

## Hand-offs

- Writer & atomic rename: [E2](002-domain-and-io.md).
- Pre-write validation pipeline: [E3](003-validate-command.md).
- Concept-ID derivation: cli-design.md §"Concept IDs" — lives in
  `internal/write/id.go`.
- Picklist values: `internal/tbx/picklist.go` (E3).
- Reflective payload-schema discovery:
  [schema-source-of-truth](../adr/schema-source-of-truth.md).
- Bulk variant (`apply`): [E8](008-apply.md).
- Input hardening (path, ID, lang): [E9](009-hardening.md).
- Determinism on the emit side:
  [determinism](../adr/determinism.md).
- Telemetry for missing-author case:
  [logging](../adr/logging.md).
