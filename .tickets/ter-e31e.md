---
id: ter-e31e
status: closed
deps: [ter-nlnx, ter-5ykh, ter-s3uj, ter-ob6m, ter-4pgc, ter-5ijf, ter-48g9, ter-v057, ter-u8kx, ter-ubfi, ter-so12]
links: []
created: 2026-05-22T19:52:50Z
type: task
priority: 4
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, refactor]
---
# E1.T17 — Refactor: extract shared flag groups

## Goal

After T6–T16 land with their command-local flag declarations, extract the repeated patterns into shared helper constructors. Behavior must not change — every existing golden file and every existing test stays byte-identical green. Pure surgical refactor on green.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Refactor candidates (post-green)"
- TDD skill: never refactor while RED — only after a clean full-suite green pass.

## Depends on

- T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16 (all command tickets green)

## Files to create / edit

- `src/internal/app/commands/flags.go` (new) — shared flag constructors
- `src/internal/app/commands/flags_test.go` (new) — unit tests asserting the constructors emit the right shapes
- Edit each `src/internal/app/commands/<cmd>.go` to consume the new helpers
- All testdata directories unchanged — no golden file should require updating; if a golden changes, you regressed the surface

## Helpers to introduce

```go
package commands

// writeFlags returns the dry-run / transaction / author triple shared
// by every write command (concept add/update/remove, term add/deprecate,
// apply).
func writeFlags() []cli.Flag

// readFieldsFlag returns the --fields/-F flag shared by every read
// command (validate, lookup, scan, check, extract).
func readFieldsFlag() cli.Flag

// langFlag returns a --lang/-l string flag. required toggles Required: true.
func langFlag(required bool) cli.Flag

// termFlag returns a --term/-t string flag. required toggles Required: true.
func termFlag(required bool) cli.Flag

// pickFlag returns a string flag with a closed-enum validator over the
// values returned by valuesFn. Used for --status, --part-of-speech,
// --register, --grammatical-gender, --script.
func pickFlag(name, alias, usage string, valuesFn func() []string) cli.Flag

// authorFlag returns the --author/-a flag with Sources bound to
// TERMINOLOGY_AUTHOR.
func authorFlag() cli.Flag

// dryRunFlag returns --dry-run/-n.
func dryRunFlag() cli.Flag

// transactionFlag returns --transaction.
func transactionFlag() cli.Flag
```

## Process

1. Run `make build` — confirm green baseline before any edit.
2. Land helpers in `commands/flags.go` + unit tests.
3. Migrate **one command at a time**, running `make build` after each migration. If a golden diff appears, revert the migration and investigate — the helper introduced a surface drift.
4. Order of migration (smallest blast radius first): `validate` (just --fields), `lookup`, `scan`, `extract` (introduces pickFlag), `check`, `apply` (introduces authorFlag + dryRunFlag + transactionFlag), `concept_remove`, `term_deprecate`, `term_add`, `concept_update`, `concept_add`.
5. After all migrations: `make build`; goldens unchanged; `git diff src/internal/app/testdata/` returns empty.
6. Remove now-dead inline flag literals from each command file.

## Acceptance

- `make build` clean
- `git diff src/internal/app/testdata/` is empty (no golden file changed)
- `cd src && go test ./internal/app/...` passes
- LOC reduction in `internal/app/commands/` is positive (helpers replace duplication)

## Out of scope

- Behavior changes — this ticket is purely structural.
- Any new tests beyond `flags_test.go` for the helpers themselves.
- Moving the helpers out of `internal/app/commands` — they live where they're consumed.


## Notes

**2026-05-22T21:37:51Z**

Extracted shared flag constructors into src/internal/app/commands/flags.go: writeFlags (dry-run+transaction+author triple), readFieldsFlag, langFlag, termFlag, pickFlag, dryRunFlag, transactionFlag, authorFlag. Removed duplicate conceptEnumValidator/termEnumValidator; scriptPickFlag kept as local helper in extract.go because it has a default Value. All golden files unchanged, make build clean, flags_test.go covers every helper.
