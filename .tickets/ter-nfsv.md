---
id: ter-nfsv
status: closed
deps: [ter-po3t, ter-2mia]
links: []
created: 2026-07-12T11:51:37Z
type: feature
priority: 1
assignee: Andre Silva
tags: [field-feedback, read, serializer, feat-1, foundation]
---
# FEAT — Foundation: canonical concept read serializer + per-language definitions (FEAT-1)

# FEAT (A) — Foundation: canonical concept read serializer + per-language definitions (FEAT-1)

## Why this ticket exists

Field feedback (`.local/tmp/issues.md`) shows the tool can *write* a rich
concept model but cannot *read* it back through the CLI, and cannot record a
per-language definition. The agreed design (see "Design decisions" below) makes
one canonical concept JSON shape — the existing `output.WriteResult`, which
`concept add` / `apply` already accept and echo — the single shape all read
commands emit. This ticket builds that shared foundation and closes FEAT-1. It
is the dependency for the read-commands ticket (export/show/list + lookup
upgrade) and the search ticket.

Covers: **FEAT-1** (per-language definitions) in full; establishes the shared
serializer; documents the **FEAT-4** workaround (deferred feature).

## Design decisions (agreed)

- **One canonical shape.** Read output == write input (`output.WriteResult`).
  `export` will emit `{"concepts":[...]}` that `apply` consumes byte-for-byte.
- **FEAT-1 shape.** Per-language definitions are an additive sibling key
  `languages.<lang>.definitions: [string]`, coexisting with the existing
  concept-level (language-neutral) `definitions`. Non-breaking, additive,
  TBX-correct (maps to the langSec-scoped `<basic:definition>`).
- **FEAT-4 deferred.** Do not change the scalar `reading`/`reading_note` model.
  Document the workaround only.

## Current state (verified against source)

- The domain model already supports per-language definitions:
  [`LangSection.Definitions []NoteText`](src/internal/tbx/model.go#L41).
- The reader already parses langSec-level definitions:
  [`decodeLangSecFields`](src/internal/tbx/linguist/reader.go#L530-L553).
- The writer already emits them:
  [`writeLangSec`](src/internal/tbx/linguist/writer.go#L142-L144).
- The **only** gap is the JSON I/O struct: `output.WriteTermGroup`
  ([`types.go:200-205`](src/internal/output/types.go#L200-L205)) has no
  `definitions` key, so `languages.<lang>.definitions` is rejected as an unknown
  field, and there is no path to write or read a language-scoped definition.
- The Concept→JSON serializer today is the unexported
  [`buildWriteResult`](src/internal/app/commands/write_helpers.go#L23-L65) in
  package `commands`; its inverse
  [`WriteResultToConcept`](src/internal/write/apply.go#L36-L71) lives in package
  `write`. The two directions are split across packages and `buildWriteResult`
  is not reusable by future read commands without importing `commands`.

## Scope of work

### 1. Add per-language definitions to the canonical shape (FEAT-1)

- Add `Definitions []string` with tag `json:"definitions,omitempty"` to
  [`output.WriteTermGroup`](src/internal/output/types.go#L200-L205).
- Input wiring: in
  [`WriteResultToConcept`](src/internal/write/apply.go#L36-L71), map
  `grp.Definitions` → `ls.Definitions` (`[]tbx.NoteText{ {Plain: d} }`),
  mirroring how concept-level `wr.Definitions` is mapped at
  [`apply.go:43-45`](src/internal/write/apply.go#L43-L45).
- Output wiring: in the serializer (below), map `ls.Definitions` →
  `grp.Definitions`.
- Reader/writer need no change (already handled).

### 2. Centralize the canonical serializer (shared foundation)

- Introduce exported `write.ConceptToWriteResult(c tbx.Concept)
  output.WriteResult` (new `src/internal/write/serialize.go`, or beside
  `WriteResultToConcept` in `apply.go`), moving the logic currently in
  `commands.buildWriteResult` + `commands.tbxTermToWriteTerm`. Include the new
  per-language `Definitions` output.
- Update
  [`commands.buildWriteResult`](src/internal/app/commands/write_helpers.go#L23-L65)
  to delegate to `write.ConceptToWriteResult` (keep the existing call sites in
  concept_add/update/remove/term commands working unchanged).
- Rationale (deep module): both mapping directions now live together in
  package `write`, and read commands (next tickets) import one canonical
  serializer instead of reaching into `commands`.

### 3. Round-trip fidelity check

- The round-trip contract is "what `apply` consumes." Ensure the fields the
  feedback cares about survive `ConceptToWriteResult` → JSON →
  `WriteResultToConcept`: `subject_field`, concept-level + per-language
  `definitions`, per-term `reading`/`reading_note`/`contexts`/`notes`,
  administrative statuses (preferred/admitted/deprecated/superseded),
  `cross_refs`.
- Concept-level fields NOT represented in `WriteResult` today (`Graphics`,
  `CustomerSubset`, `ProjectSubset`) will be dropped on round-trip. Either add
  them to `WriteResult` (additive, small) or document the omission explicitly in
  the shape docs. Pick one and note it in the ticket on close.

### 4. Documentation

- In
  [`docs/skills/terminology/references/write-details.md`](docs/skills/terminology/references/write-details.md):
  - Document `languages.<lang>.definitions` under "Language section keys"
    (§~98-108) and remove the "No per-language definitions" limitation bullet
    (§~155-165).
  - Add a short "Multiple readings (workaround)" note for **FEAT-4**: represent
    an alternate reading as an *admitted* term (it will be surfaced by
    search/show/export), or stash it in `reading_note`/`notes`.
- If [`docs/cli-design.md`](docs/cli-design.md) documents the concept JSON
  shape, add the per-language `definitions` key there too.
- Run `markdownlint-cli2 --config ~/.markdownlint.yaml --fix` on edited
  markdown.

### Go conventions

- Exported `ConceptToWriteResult` needs a doc comment beginning with the
  identifier name; `MixedCaps`; `gofmt` clean.
- Copy slices at the boundary if the serializer must not alias model slices
  (see the `defensive` guidance) — the current `buildWriteResult` assigns
  `r.ExternalRefs = c.ExternalRefs` directly; preserve or tighten as
  appropriate, but do not introduce aliasing bugs.
- No behavior change for existing write-result echoes (concept add/update/etc.).

## TDD plan (vertical slices — one test, one change, repeat)

Public interfaces under test: `write.ConceptToWriteResult`,
`write.WriteResultToConcept`, `write.ParseJSONInput`, `write.ParseApplyJSON`.
Add to [`src/internal/write/apply_test.go`](src/internal/write/apply_test.go)
(or a new `serialize_test.go`). Go RED→GREEN per cycle.

### Cycle 1 (tracer) — exported serializer exists and carries rich fields

RED: `write.ConceptToWriteResult` on a concept with a concept-level definition
and one `en` langSec (preferred term with `reading`/`reading_note`, a context,
a note) returns a `WriteResult` with those populated. (Symbol does not exist
yet.)
GREEN: create `ConceptToWriteResult` by moving `buildWriteResult`'s logic into
`write`; have `commands.buildWriteResult` delegate.

### Cycle 2 — per-language definition is accepted on input (FEAT-1)

RED: `ParseJSONInput` on a payload with `languages.en.definitions:["EN def"]`
yields a concept whose `Languages["en"].Definitions[0].Plain == "EN def"`. Fails
today (unknown field → `invalid_input`; this is exactly the field-level error
GAP-2 will now produce, which is why this ticket depends on GAP-2).
GREEN: add `Definitions` to `WriteTermGroup`; wire it in
`WriteResultToConcept`.

### Cycle 3 — per-language definition is emitted on output

RED: `ConceptToWriteResult` on a concept with `Languages["en"].Definitions`
returns `WriteResult.Languages["en"].Definitions == ["EN def"]`.
GREEN: wire `ls.Definitions` → `grp.Definitions` in the serializer.

### Cycle 4 — full round-trip is lossless for feedback-relevant fields

RED: build a concept exercising subject_field, concept-level + per-language
definitions, preferred+admitted terms with reading/reading_note/contexts/notes,
and a cross_ref; assert
`WriteResultToConcept(ConceptToWriteResult(c))` equals `c` (compare via the
existing canonical equality helper used by apply, e.g.
[`internal/write/equality.go`](src/internal/write/equality.go)).
GREEN: fix any field that fails to round-trip; decide the
Graphics/CustomerSubset/ProjectSubset omission (add or document).

### Cycle 5 — apply consumes the serializer output (bilingual definitions)

RED: marshal `{"concepts":[ConceptToWriteResult(c)]}` for a concept with an EN
and a PT per-language definition; `ParseApplyJSON` → `WriteResultToConcept`
reproduces both langSec definitions. Then a second apply of the same payload
classifies `unchanged` (idempotence) — reuse the apply golden/idempotence tests
in [`src/internal/app/apply_golden_test.go`](src/internal/app/apply_golden_test.go).
GREEN: covered by Cycles 2-4; wire the test.

### Refactor

- Ensure `commands.buildWriteResult`/`tbxTermToWriteTerm` are thin delegators or
  removed in favor of the `write` functions; dedupe.
- Run tests after each step. Never refactor while RED.

## Acceptance criteria

- `languages.<lang>.definitions: [string]` is accepted on `concept add`, `apply`
  input and echoed on their result envelopes; it emits a langSec-scoped
  `<basic:definition>` and round-trips through `validate`.
- Concept-level `definitions` still works (language-neutral) and coexists.
- `write.ConceptToWriteResult` is exported and is the single Concept→WriteResult
  serializer; `commands` delegates to it.
- Round-trip is lossless for the feedback-relevant fields (Cycle 4); any
  concept-level field not represented in `WriteResult` is either added or
  documented.
- `write-details.md` documents per-language definitions and the FEAT-4
  workaround; the "No per-language definitions" limitation is removed.
- `make build` passes from the project root.

## Files to touch

- `src/internal/output/types.go` — `WriteTermGroup.Definitions`.
- `src/internal/write/serialize.go` (new) or `apply.go` —
  `ConceptToWriteResult`; per-language definitions in both directions.
- `src/internal/app/commands/write_helpers.go` — delegate `buildWriteResult`.
- `src/internal/write/apply_test.go` / new `serialize_test.go` — Cycles 1-5.
- `docs/skills/terminology/references/write-details.md`,
  `docs/cli-design.md` (if it documents the shape).

## Validation

`make build` from the project root; tighter loop:
`go test ./src/internal/write/... ./src/internal/output/... ./src/internal/app/...`.
Do not silence lint with `_ =`.

## Dependencies

Depends on **BUG-1** (`ter-po3t`) and **GAP-2** (`ter-2mia`): this ticket
extends the same JSON I/O path (`ParseJSONInput` / `WriteResultToConcept`) and
the canonical shape. Landing the fragment fix and the field-level validation
errors first avoids conflicts and means a mistyped `languages.<lang>.definitions`
surfaces through GAP-2's field-level error machinery rather than the old generic
one.


## Notes

**2026-07-13T17:13:52Z**

FEAT-1 done. Added output.WriteTermGroup.Definitions ([]string, langSec-scoped basic:definition), wired both directions (WriteResultToConcept input; new write.ConceptToWriteResult output). Centralized the canonical Concept->WriteResult serializer in package write (serialize.go: ConceptToWriteResult + TermToWriteTerm), moved from commands; commands.buildWriteResult is now a delegator and commands.tbxTermToWriteTerm removed. Round-trip lossless for feedback-relevant fields (Cycle 4 via ConceptsEqual). DECISION: concept-level Graphics/CustomerSubset/ProjectSubset and langSec Sources are NOT in WriteResult -> documented as omitted (edit via --format tbx), not added. Regenerated TestSchema_Full_Golden (field added to WriteTermGroup schema). Docs: write-details.md documents languages.<lang>.definitions + FEAT-4 workaround; removed the 'No per-language definitions' limitation. cli-design.md needed no change (no formal write-payload field table). E2E verified: apply -> emits per-langSec basic:definition -> validate ok -> idempotent re-apply=unchanged. make build passes. Note for next: no project markdownlint config exists (~/.markdownlint.yaml and repo both missing); make build does not lint markdown. Unblocks ter-lutz (read commands) and ter-2d5b (search).
