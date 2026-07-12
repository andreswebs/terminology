---
id: ter-lutz
status: open
deps: [ter-nfsv]
links: []
created: 2026-07-12T11:51:37Z
type: feature
priority: 2
assignee: Andre Silva
tags: [field-feedback, read, export, show, list, lookup, feat-2, feat-3]
---
# FEAT — Round-trip read commands: export, show, list + lookup upgrade (FEAT-2/FEAT-3)

# FEAT (B) — Round-trip read commands: export, show, list + lookup upgrade (FEAT-2 + FEAT-3)

## Why this ticket exists

Field feedback (`.local/tmp/issues.md`, FEAT-2 / FEAT-3): everything the
enrichment keys let you *write* (definitions, readings, notes, contexts,
non-preferred langSecs) is unreadable through the CLI, and there is no way to
enumerate or export the glossary. Writes cannot be verified except by opening
the raw TBX. This ticket adds the read surface, all emitting the single
canonical concept shape established in the foundation ticket.

Covers: **FEAT-2** (read path for the rich fields) and **FEAT-3** (list /
export).

## Design decisions (agreed)

- **Canonical shape.** Every read command emits the `output.WriteResult`
  concept shape produced by `write.ConceptToWriteResult` (foundation ticket).
- **Four surfaces, all explicit:** `export`, `show <id>`, `list`, and an
  upgraded `lookup`.
- **`export` is apply-consumable:** `{"concepts":[...]}` == what `apply`
  ingests, enabling read-modify-write.
- **`list` is a projection, not a new schema:** the same canonical concept
  element, populated only with `concept_id` + per-language preferred term
  (equivalent to `export --fields concept_id,languages.*.preferred.term`).
- **`lookup` stays a strict finder** (exact whole-string, NFC + casefold, term
  surfaces only) but its *output* is upgraded to the canonical concept shape.
  The fuzzy/substring/reading-aware finder is a separate `search` command
  (its own ticket).
- **Exit codes mirror `lookup`:** `show <missing>` and (in the search ticket)
  zero-hit → exit 1 `not_found`; empty `export`/`list` → exit 0 with `[]`.
- **Conventions:** results sorted by `concept_id` (determinism ADR); `--fields`
  projection supported like the other read commands.

## Current state (verified against source)

- `lookup` emits a reduced shape: `LookupResult` / `LookupTermGroup` /
  `LookupTerm` carry only `term` per language
  ([`output/types.go:50-69`](src/internal/output/types.go#L50-L69)); the
  builder [`buildLookupResult`](src/internal/app/commands/lookup.go#L78-L101)
  drops definitions/readings/contexts/notes and non-preferred/admitted terms.
  This is why `lookup "entering throw" --fields results.definitions` returns
  `invalid_field` — the path does not exist in the envelope.
- No `export` / `show` / `list` commands exist. Commands are registered in
  [`app/root.go:26-37`](src/internal/app/root.go#L26-L37).
- Envelopes + allowed exit codes are registered in
  [`output/types.go` `init()`](src/internal/output/types.go#L5-L32).
- Read-command plumbing to reuse: `tbxPathFromRoot(cmd)`,
  `wrapLoadError(err)`, `readFieldsFlag()`
  ([`flags.go:17-19`](src/internal/app/commands/flags.go#L17-L19)),
  `langFlag(...)`, `output.EmitProjected(w, env, fieldsStr)` and
  `output.EmitJSON(w, env)` (see
  [`lookup.go:28-68`](src/internal/app/commands/lookup.go#L28-L68) and
  [`concept_add.go:88-99`](src/internal/app/commands/concept_add.go#L88-L99)).
- The `--fields` engine (`output.ValidateFields` / `output.ProjectFields`) lives
  in `src/internal/output/fields.go`.
- The determinism contract is in
  [`docs/adr/determinism.md`](docs/adr/determinism.md); the writer already sorts
  concepts by id
  ([`writer.go:245-252`](src/internal/tbx/linguist/writer.go#L245-L252)).

## Scope of work

Depends on the foundation ticket's `write.ConceptToWriteResult`. All new
serialization goes through it.

### 1. `export` command (`src/internal/app/commands/export.go`)

- Loads the glossary, serializes every concept via `ConceptToWriteResult`,
  emits `ExportEnvelope{schema_version, ok, concepts:[]WriteResult}` sorted by
  `concept_id`.
- Flags: `readFieldsFlag()`, `langFlag(false, "restrict emitted language
  sections to this tag")`. When `--lang` is set, restrict each concept's
  `languages` map to that tag (concepts with no such langSec still appear with
  an empty `languages`, or are omitted — pick and document; recommend: include
  with empty map for a stable set).
- Empty glossary → exit 0 with `concepts:[]` (envelope `MarshalJSON` normalizes
  nil → `[]`, mirroring existing envelopes).
- Emit via `output.EmitProjected(cmd.Root().Writer, env, cmd.String("fields"))`.

### 2. `show <concept-id>` command (`show.go`)

- Positional `concept-id` arg (see the `ArgsUsage`/`Arguments`/`argBounds`
  pattern in [`lookup.go:11-26`](src/internal/app/commands/lookup.go#L11-L26)).
- Finds the concept by id
  ([`write.ConceptIndex`](src/internal/write/concept.go#L11-L18) or a direct
  scan); serializes via `ConceptToWriteResult`.
- Present → `ShowEnvelope{schema_version, ok, concept: WriteResult}`, exit 0.
- Absent → exit 1 with error code `not_found` (reuse the
  `lookupNotFoundError` pattern at
  [`lookup.go:103-112`](src/internal/app/commands/lookup.go#L103-L112) or a
  shared not-found error).
- Flag: `readFieldsFlag()`.

### 3. `list` command (`list.go`)

- Emits the canonical concept element but populated only with `concept_id`
  (+ `subject_field`) and, per language, the preferred term. Implement as a
  projection: build the full `[]WriteResult`, then reduce each element to id +
  preferred terms (or reuse `output.ProjectFields` with a fixed field set).
- `ListEnvelope{schema_version, ok, concepts:[]WriteResult}` sorted by id.
  Empty → exit 0, `concepts:[]`.
- Document that `list` == `export --fields concept_id,subject_field,
  languages.*.preferred.term`.

### 4. `lookup` upgrade (`lookup.go` + `output/types.go`)

- Change `LookupEnvelope.Results` to `[]output.WriteResult` (or make
  `buildLookupResults` return canonical concepts). Serialize matched concepts
  via `ConceptToWriteResult`. Remove/retire `LookupResult`/`LookupTermGroup`/
  `LookupTerm` if unused.
- Keep the matching semantics of
  [`Glossary.Lookup`](src/internal/tbx/lookup.go#L14-L31) unchanged (exact
  whole-string, NFC + casefold). Only the output shape changes.
- This is a **breaking change to the lookup envelope**; update lookup golden
  tests and the `--fields` paths accordingly. Note it in the ticket and docs.

### 5. Registration + schema

- Register the three new commands in
  [`app/root.go`](src/internal/app/root.go#L26-L37).
- Add `RegisterEnvelope` + `RegisterExitCodes` entries for `export`, `show`,
  `list` in [`output/types.go` `init()`](src/internal/output/types.go#L5-L32):
  export/list `[]int{0,2,3,65}`, show `[]int{0,1,2,3,65}`. This makes them
  discoverable via `terminology schema --command CMD`.
- Give each new envelope a `MarshalJSON` that normalizes nil slices/maps to
  `[]`/`{}` like the existing envelopes
  ([`types.go:99-106,229-236`](src/internal/output/types.go#L99-L106)).

### 6. Documentation

- Add `export`, `show`, `list` and the upgraded `lookup` to
  [`docs/cli-design.md`](docs/cli-design.md) (command surface, usage, exit
  codes) and to the terminology skill docs. Show the read-modify-write loop:
  `terminology export | jq ... | terminology apply --file -`.
- Update
  [`docs/skills/terminology/references/write-details.md`](docs/skills/terminology/references/write-details.md)
  "Dry-run" note (§~277-280) and any FEAT-2/FEAT-3 mentions that claimed the
  data is unreadable.
- `markdownlint-cli2 --config ~/.markdownlint.yaml --fix` on edited markdown.

### Go conventions

- Each command mirrors the existing read-command structure (`Before:
  argBounds(...)`, `Action`, `tbxPathFromRoot`, `wrapLoadError`); early-return
  on error; happy path at minimal indentation.
- Exported command constructors (`Export`, `Show`, `List`) get doc comments
  beginning with the name; `MixedCaps`; `gofmt` clean.
- Reuse `output.EmitProjected` so `--fields` and format handling stay uniform.

## TDD plan (vertical slices — one test, one change, repeat)

Public interfaces under test: the command actions end-to-end (drive via the CLI
harness used in `src/internal/app/*_test.go` / golden tests) and the envelope
builders. Go RED→GREEN per cycle.

### Cycle 1 (tracer) — `export` round-trips

RED: seed a temp glossary with 2 concepts (one with definitions + readings);
run `export`; assert `concepts` is sorted by id and each element carries the
rich fields. Then feed the exact output to `apply --file -` and assert every
concept classifies `unchanged` (read-modify-write is a no-op).
GREEN: implement `export` on top of `ConceptToWriteResult`.

### Cycle 2 — `show` by id, present and absent

RED: `show <id>` returns the full canonical concept (exit 0); `show <missing>`
exits 1 with `error.code == "not_found"`.
GREEN: implement `show`.

### Cycle 3 — `list` is the projected view

RED: `list` returns each concept as id + preferred term(s) only, sorted by id;
no definitions/readings present.
GREEN: implement `list` as a projection.

### Cycle 4 — `lookup` now exposes the rich fields (FEAT-2 repro)

RED: reproduce `lookup "entering throw" --fields results.definitions` — assert
it is NOT `invalid_field` and that definitions/readings/contexts appear in the
result for a matched concept.
GREEN: upgrade the lookup envelope to the canonical shape; update goldens.

### Cycle 5 — conventions: `--fields`, `--lang`, empty glossary

RED: `export --fields concept_id` yields only id (+ envelope boilerplate);
`export --lang ja` restricts emitted language sections; `export`/`list` on an
empty glossary exit 0 with `concepts:[]`.
GREEN: wire `EmitProjected` + `--lang` filtering + nil-normalizing MarshalJSON.

### Cycle 6 — schema discoverability

RED: `terminology schema --command export` (and show/list) returns the
registered envelope + exit codes.
GREEN: add the `RegisterEnvelope`/`RegisterExitCodes` entries.

### Refactor

- Extract any shared "load glossary → []WriteResult" helper used by
  export/list/lookup. Extract a shared not-found error if show and lookup both
  need one.
- Run tests after each step. Never refactor while RED.

## Acceptance criteria

- `terminology export` emits `{"concepts":[...]}` in the canonical shape, sorted
  by id, and `export | apply --file -` is a no-op (all `unchanged`).
- `terminology show <id>` returns the full concept; `show <missing>` exits 1
  `not_found`.
- `terminology list` enumerates id + preferred terms per language, sorted.
- `terminology lookup <term>` now exposes definitions/readings/contexts/notes
  and non-preferred terms; `--fields results.definitions` is valid.
- `--fields` and `--lang` work on the new commands; empty glossary → exit 0.
- All four are discoverable via `terminology schema --command CMD`.
- Docs updated (cli-design + skill); the FEAT-2/FEAT-3 "unreadable" claims are
  gone.
- `make build` passes from the project root.

## Files to touch

- `src/internal/app/commands/export.go`, `show.go`, `list.go` (new) + tests.
- `src/internal/app/commands/lookup.go` — upgrade output.
- `src/internal/output/types.go` — new envelopes + `init()` registration;
  lookup envelope change.
- `src/internal/app/root.go` — register commands.
- `src/internal/app/*_golden_test.go` — export/show/list goldens; update lookup
  goldens.
- `docs/cli-design.md`, `docs/skills/terminology/references/write-details.md`,
  terminology skill docs.

## Validation

`make build` from the project root; tighter loop:
`go test ./src/internal/app/... ./src/internal/output/...`. Do not silence lint
with `_ =`.

## Dependencies

Depends on the **foundation ticket (FEAT-A)** for
`write.ConceptToWriteResult` and the finalized canonical shape (including
per-language definitions).

