---
id: ter-8gyy
status: closed
deps: [ter-uqyn, ter-told, ter-lyrx]
links: []
created: 2026-05-22T19:19:20Z
type: epic
priority: 3
assignee: Andre Silva
tags: [epic, write]
---
# E7 — Write commands: concept & term

Spec: docs/specs/007-write-commands.md

Five granular mutations on top of E2's writer:
- concept add / update [--merge] / remove [--force]
- term add / term deprecate

Shared internal/write/:
- id.go — concept-ID derivation (cli-design.md §Concept IDs); stable across preferred-term rename
- transaction.go — <transacGrp> emission scoped to mutation node (conceptEntry/langSec/termSec)
- dryrun.go — final-state preview without rename

Input layering: flags → JSON stdin → TBX fragment stdin (--format tbx, singular <conceptEntry> or <conceptEntryList> wrapper; full <tbx> rejected). JSON payload struct = lookup output struct (round-trip trivial). --merge: term natural key (Surface, AdministrativeStatus); definitions/notes replace-if-present.

Author resolution via urfave Sources (--author > TERMINOLOGY_AUTHOR); missing-author when --transaction set emits WARN-level slog (no envelope pollution).

Pre-write validation runs the full E3 pipeline on the in-memory result before rename; failure aborts before rename, file untouched. Concurrency: ${TBX}.lock held across read→modify→write window.

New sentinels: duplicate_id, dangling_crossref, not_found, invalid_id, invalid_picklist, invalid_input.

## Tasks

| ID       | Task                                          | Deps                    |
| -------- | --------------------------------------------- | ----------------------- |
| ter-pe7f | R1 — pickFlag wiring (retroactive, closed)    | —                       |
| ter-ttwj | T1 — Clock injection (internal/clock)         | entry gate              |
| ter-8i7w | T2 — Write error sentinels                    | entry gate              |
| ter-kybc | T3 — Write envelope types (output/types.go)   | entry gate              |
| ter-bmce | T4 — Concept-ID derivation (write/id.go)      | entry gate              |
| ter-5b07 | T5 — Transaction record builder               | T1                      |
| ter-w01i | T6 — Write pipeline (write/write.go)          | T2, T3, T5              |
| ter-x3kh | T7 — concept add command action               | T4, T6                  |
| ter-smzv | T8 — concept update command action             | T6                      |
| ter-2cqg | T9 — concept remove command action             | T6                      |
| ter-iu37 | T10 — term add command action                  | T6                      |
| ter-ppld | T11 — term deprecate command action            | T6                      |
| ter-7ik8 | T12 — Golden CLI tests for write commands      | T7, T8, T9, T10, T11   |
| ter-sxkb | T13 — Manual QA plan for E7                    | T12                     |

## Gates

- Entry gate: ter-st7u
- Exit gate: ter-lyrx
