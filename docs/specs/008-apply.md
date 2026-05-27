# E8 — `terminology apply`

> **Status**: APPROVED. Bulk declarative write. Reconciles a desired-state
> payload against the current glossary, computes the minimal set of
> add/update/remove operations, applies them all-or-nothing.
>
> Sits on top of [E7](007-write-commands.md) infrastructure.

## Synopsis

```
terminology apply --file PAYLOAD [--prune] [--dry-run] [--transaction] [--author NAME] [--format FMT]
```

`--file -` reads from stdin.

## Patch model

Full-state. The payload describes the **target** shape of the listed
concepts; `apply` diffs against the current glossary and computes the
minimal mutation set. Idempotent by construction (a clean run on an
already-converged file yields zero ops).

JSON Patch–style ordered operations are not supported in v1 — full
state strict-supersets them (any operation list can be reproduced by
a full-state payload). Adding ops later remains a non-breaking
extension behind `--format ops` if a use case emerges.

## Algorithm

1. Load current glossary (E2).
2. Load payload from `--file` (path or `-`).
3. For each concept in payload:
   - Not in glossary → **add**.
   - In glossary, content differs → **update** (wholesale replace; see
     "Update rule").
   - In glossary, content matches (per equality below) → **unchanged**.
4. If `--prune`, every glossary concept absent from payload → **remove**
   (subject to the crossref policy below).
5. Validate the **resulting** in-memory glossary (E3 full pipeline).
6. If validation fails → abort, emit the distinct error envelope,
   leave the file untouched.
7. Otherwise → atomic rename (E2 writer).

### Update rule

When a payload concept exists in the glossary but content differs, the
concept is **replaced wholesale**. The payload is authoritative;
payload-omitted fields are dropped from the resulting concept.

This matches the full-state philosophy: the payload IS the target
state. For surgical, additive merges use `concept update --merge`
([E7](007-write-commands.md)). The two commands deliberately diverge:
`apply` converges to a stated shape; `concept update --merge` overlays.

### Equality (for "unchanged")

Two concepts compare equal if their **canonicalized** XML forms are
byte-identical **after stripping `<transacGrp>` elements**. Rationale:

- Canonicalization handles field ordering, whitespace, and attribute
  ordering — payload order doesn't matter.
- Ignoring `<transacGrp>` means a payload without transactions matches
  a glossary with transactions as "unchanged". `apply` preserves
  existing transactions; it does not append a new transaction record
  unless the concept content actually changed.

### `--prune` and dangling cross-references

If `--prune` would remove concept `C`, and another concept (whether
payload-present or payload-absent-but-preserved) `<crossReference>`s
`C`, the run **refuses** with `dangling_crossref` and leaves the file
untouched. Matches `concept remove` semantics without `--force`.

The agent's recourse: drop the offending ref in the payload, or keep
`C` in the payload. v1 deliberately omits a `--drop-refs` flag — the
explicit-payload fix is the right pattern for a declarative tool.

## Output

### Success envelope

```json
{
  "schema_version": 1,
  "ok": true,
  "applied": {
    "added": ["tzimtzum"],
    "updated": ["razon-historica"],
    "removed": [],
    "unchanged": ["binah", "malkhut"]
  },
  "warnings": []
}
```

Lists are sorted **ASCII byte order by `concept_id`**, per
[determinism](../adr/determinism.md). Stable
across runs regardless of payload order.

Exit `0` on success.

### Failure envelope

When validation of the resulting glossary fails (or any other
recoverable apply-level error), emit the standard error envelope with
a dedicated code and a `failures` detail array:

```json
{
  "schema_version": 1,
  "ok": false,
  "error": {
    "code": "apply_validation_failed",
    "message": "5 concepts failed validation; no changes written",
    "details": {
      "failures": [
        {"concept_id": "razon-historica", "code": "dangling_crossref", "message": "..."},
        {"concept_id": "tzimtzum",        "code": "invalid_picklist",  "message": "..."}
      ]
    }
  }
}
```

The dedicated `apply_validation_failed` code lets agents pattern-match
on `error.code` instead of inferring failure from envelope shape.
The on-disk file is **never** partially written — all-or-nothing.

Exit `1` on validation failure (recoverable). `2` on usage error.
`3` on I/O.

New sentinel:

```go
// internal/apply/errors.go (or wherever apply lives)
var ErrApplyValidationFailed = terr.New(
    "apply_validation_failed", 1,
    "fix per-concept errors in failures[] and retry",
    "apply payload failed validation; no changes written",
)
```

## Payload format

The default format is **JSON**, structurally a list of concept
records:

```json
{
  "concepts": [
    {
      "concept_id": "tzimtzum",
      "subject_field": "kabbalah",
      "languages": {
        "he": {"preferred": {"term": "צמצום"}},
        "es": {"preferred": {"term": "tzimtzum"}}
      }
    }
  ]
}
```

The concept-record shape is the same Go struct used by `lookup` output
(per [E7](007-write-commands.md) §"JSON payload shape"). The
`internal/output` reflective machinery validates the payload; unknown
fields produce an `invalid_field` error envelope.

Alternative: `--format tbx` — payload is a `<conceptEntryList>` wrapper
with one or more `<conceptEntry>` children. Same fragment rules as
[E7](007-write-commands.md) §"TBX fragment input boundaries".

### Format selection precedence

1. **Explicit `--format`** wins when set.
2. **Auto-detect from extension** when `--format` is absent:
   `.json` → JSON, `.tbx`/`.xml` → TBX fragment.
3. **`--file -` (stdin)** with no `--format` and no extension hint →
   error envelope `invalid_input` with a hint to set `--format`.

## Idempotency

Running the same payload twice in a row must yield zero ops on the
second invocation:

```
apply payload.json   # → applied: {added:[X], ...}
apply payload.json   # → applied: {unchanged:[X, ...]}
```

The transac-strip rule in the equality compare guarantees that a
transaction record added by the first run does not flip the second
run's concept to "update". Tested explicitly.

## Concurrency

`apply` holds the sibling `${TBX}.lock` advisory lock for the entire
read → modify → validate → write window (E2). Non-blocking; a
concurrent write fails fast with `tbx_locked`. Bulk operations are
disproportionately costly to retry, so failing fast is the right
default — the operator decides whether to wait and retry.

## Hand-offs

- Write path: [E7](007-write-commands.md).
- Domain model: [E2](002-domain-and-io.md).
- Pre-write validation pipeline (full three-tier): [E3](003-validate-command.md).
- Determinism (ASCII byte sort, RFC3339 timestamps):
  [determinism](../adr/determinism.md).
- Payload schema discoverability:
  [schema-source-of-truth](../adr/schema-source-of-truth.md).
- `ErrApplyValidationFailed` registered per
  [error-handling](../adr/error-handling.md).
