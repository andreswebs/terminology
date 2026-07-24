---
id: ter-ve3i
status: closed
deps: [ter-lot8]
links: [ter-lot8]
created: 2026-07-22T17:54:26Z
type: task
priority: 2
assignee: Andre Silva
---
# Migrate exit codes to the fleet taxonomy (ADR 0001)

## Context

`docs/adr/0001-exit-code-taxonomy.md` (already in this repo) adopts a
family-wide taxonomy: 0 success; 1 recoverable result; 2-63 optional
result sub-codes; 64 EX_USAGE; 65 EX_DATAERR; 70 EX_SOFTWARE;
74 EX_IOERR; 78 EX_CONFIG; other sysexits on exact fit; 130/143
signals. This tool is the closest to conformant already (0, 1, and 65
keep their meanings); the migration is: usage 2 becomes 64, I/O 3
becomes 74, lock contention becomes 75, oversize input moves from 2 to
65, and the unclassified fallback becomes 70. Breaking change (v0.x);
include release notes. The tool does not catch signals; nothing to do
there.

The older `docs/adr/error-handling.md` "Exit code families" section
(around lines 155-175) documents the superseded 2/3 scheme and must be
updated to match ADR 0001.

## Old to new mapping (decided; implement exactly this)

| Old | Class | New |
| --- | ----- | --- |
| 0 | success | 0 (unchanged) |
| 1 | violations, partial, not-found, apply validation failed | 1 (unchanged) |
| 2 | usage errors | 64 |
| 2 | `input_too_large` (`ErrInputTooLarge`) | 65 (oversize payload is data rejected, not CLI misuse) |
| 3 | `io_error` sites and I/O sentinels | 74 |
| 3 | `tbx_locked` (`ErrTBXLocked`) | 75 (EX_TEMPFAIL: transient, the hint already says "retry") |
| 65 | data errors | 65 (unchanged) |
| 1 (fallback) | unclassified errors in `ExitCodeFor` | 70 |

## Production sites to change (verified 2026-07-22)

All paths under `src/`.

Sentinels currently exit 2, change to 64 (all in `terr.New`/`terr.Newf`
second int argument):

- `internal/tbx/errors.go:20` `no_tbx_path`
- `internal/output/errors.go:14` `invalid_field`
- `internal/app/errors.go:9` `conflicting_verbosity`
- `internal/app/root.go:116` `no_subcommand`, `root.go:124`
  `unknown_subcommand`
- `internal/app/commands/check.go:18` `language_required`
- `internal/app/commands/concept.go:10` and `term.go:10`
  `no_subcommand`
- `internal/app/commands/argcheck.go:19` `missing_argument`,
  `argcheck.go:30` `excess_arguments`
- `internal/app/commands/schema.go:65` `unknown_command`
- `internal/app/commands/concept_update.go:14` `merge_replace_mutex`,
  `concept_update.go:20` `merge_replace_required`

Sentinel currently exit 2, change to 65:

- `internal/tbx/errors.go:49` `input_too_large` (`ErrInputTooLarge`)

Sites currently exit 3, change to 74 (all `terr.Newf("io_error", 3, ...)`):

- `internal/write/write.go:37`
- `internal/app/commands/check.go:89,97`
- `internal/app/commands/write_helpers.go:17`
- `internal/app/commands/scan.go:61`
- `internal/app/commands/apply.go:67`
- `internal/app/commands/init.go:61,65,79`
- `internal/app/commands/extract.go:94,107,119`

Sentinel currently exit 3, change to 75:

- `internal/tbx/errors.go:13` `tbx_locked` (`ErrTBXLocked`, produced
  only in `internal/tbx/lock.go:18`; reachable from any command that
  calls `tbx.Save`, which acquires the lock at `internal/tbx/io.go:44`).

Boundary changes in `internal/output/errors.go`:

- `ExitCodeFor` (around lines 84-93): the usage-classification branch
  returns 64 instead of 2; the final default returns 70 instead of 1.
- `classifyUsageError` output codes stay the same strings; only the
  exit mapping changes via the above.

Unchanged (verify, do not touch): `violationsError` in
`internal/app/commands/check.go` (exit 1), `ErrApplyValidationFailed`
in `internal/write/errors.go:44` (exit 1), all 65 sentinels in
`internal/tbx/errors.go`, `internal/write/errors.go`, and
`internal/app/commands/sanitize.go`.

## Registry updates (`internal/output/types.go`, around lines 23-40)

Every `RegisterExitCodes` entry changes. Rules: replace 2 with 64;
replace 3 with 74; add 70 to every command's set (the `ExitCodeFor`
fallback is reachable from any command); add 75 to every command whose
path can call `tbx.Save` (concept add/update/remove, term add, term
deprecate, apply, init; verify by tracing `tbx.Save` callers). Example:
`check` goes from `[0, 1, 2, 3, 65]` to `[0, 1, 64, 65, 70, 74]`.

Add a registry conformance test (new ADR requirement): a test that
walks `internal/app/testdata/`, derives the command from each golden
directory path, and asserts every `clean.exit` value is a member of
the registered set for that command. This ties the registry to real
behavior mechanically.

## Tests and goldens

- Golden triples: 126 `clean.exit` files exist (currently 60 hold `0`,
  17 hold `1`, 15 hold `2`, 3 hold `3`, 31 hold `65`). The 2s and 3s
  change. Regenerate with the harness update flag
  (`go test ./internal/app/ -update`, flag defined in
  `internal/app/golden_test.go`) and review the diff: only `clean.exit`
  values 2 to 64 and 3 to 74 (or 75), plus schema stdout goldens,
  should change.
- Schema goldens embed `exit_codes` arrays and the full error registry
  with per-error `exit_code` fields
  (`internal/app/testdata/schema/full/clean.stdout` and
  `schema/command_filter/clean.stdout`); they regenerate with the same
  flag.
- Inline assertions: `internal/app/commands_test.go` and neighbors
  assert exit ints in places; grep `internal/app -rn "exit"` in test
  files and update expected values per the mapping.
- `internal/output/types_test.go` and `registry_test.go` assert
  registry contents; update expected sets.

## Docs to update

- `docs/adr/error-handling.md` "Exit code families" table (lines
  155-175): rows for 2 and 3 change to 64 and 74, add 70 and 75 rows,
  and reference ADR 0001 as the authority.
- `docs/cli-design.md`: design principle 2 (line 33, full scheme
  narrative including the stale exit-75 "under construction" note,
  which now collides with 75 = tbx_locked and must be rewritten), and
  per-command mentions at lines 624 (`language_required` exit 2) and
  any other literal 2/3 mentions (grep). Lines citing 0/1/65 stay.
- `docs/skills/terminology/SKILL.md`: "Exit codes" section (line 50)
  and per-command mentions (line 74 "exit 3, io_error", others; grep
  for "exit").
- `AGENTS.md`: mentions meaningful exit codes generically (lines 8-20);
  update any literal codes if present.
- `docs/specs/001-cli-surface-stub.md` and `004-read-commands.md`
  contain historical exit-code snapshots (`exit_codes: [0, 1, 65, 3]`
  at 004:95); specs are historical records, add a one-line note at the
  top of each affected spec pointing to ADR 0001 rather than rewriting
  history.
- Note for the repo owner: the user-level terminology skill outside
  this repo mirrors SKILL.md and needs the same update when this
  ships.

## Acceptance

- All sentinels and `io_error` sites carry the new exit ints; the
  `ExitCodeFor` fallback returns 70; usage classification returns 64.
- Registry sets updated per the rules above; the new conformance test
  (goldens versus registry) passes.
- All goldens regenerated; the diff contains only expected value
  changes (2 to 64, 3 to 74/75, schema stdout).
- `docs/adr/error-handling.md`, `docs/cli-design.md`, and
  `docs/skills/terminology/SKILL.md` describe only the new scheme.
- `make build` passes from the repo root.
- Release notes drafted (breaking change: usage 2 to 64, I/O 3 to
  74/75, oversize input 2 to 65, unclassified fallback 1 to 70).
- Leave committing to the repo owner; do not create git commits.

## Notes

**2026-07-24T20:37:30Z**

Depends on and linked to ter-lot8 (Migrate Go module from src/ to repo root). After that migration lands, the Go module lives at the REPO ROOT, not under src/. Every path in this ticket that begins with src/ (the entire 'Production sites to change' section, the 'Registry updates' file paths, the goldens under src/internal/app/testdata/, and the doc paths) must be re-checked and the src/ prefix dropped before implementing. Sequence ter-lot8 first, then re-verify line numbers here, since the move may shift them.

**2026-07-24T21:29:58Z**

Migrated exit codes to ADR 0001 taxonomy. Production sites: usage sentinels 2->64 (no_tbx_path, invalid_field, conflicting_verbosity, no_subcommand x3, unknown_subcommand, language_required, missing/excess_argument, unknown_command, merge_replace_mutex/required); input_too_large 2->65; io_error Newf sites 3->74 (write.go, check.go, write_helpers.go, scan.go, apply.go, extract.go x3, init.go x3); tbx_locked 3->75. output/errors.go ExitCodeFor: usage branch ->64, fallback ->70. Registry (output/types.go) rewritten per rules: +70 all, +75 to write-path cmds (concept add/update/remove, term add/deprecate, apply, init). Added TestExitCodeRegistryConformance (internal/app/exit_conformance_test.go) walking testdata goldens, asserting each clean.exit is in the registered set; skips non-command dirs (fixtures, sanitize, bare concept/term parents) with a logged list. Regenerated 130 goldens via 'go test ./internal/app -update': 17 files 2->64, 1 file 2->65 (validate/input_too_large), 3 files 3->74. NOTE: -update also surfaced a pre-existing stale golden validate/nesting_too_deep/clean.stderr (validation_error -> invalid_input) caused by uncommitted in-flight changes in the working tree (reader.go returns tbx.ErrNestingTooDeep; wrapLoadError passes coded errors through). Regenerated it; unrelated to exit codes but required for green build. Docs updated: adr/error-handling.md (Exit code families table), cli-design.md (principle 2 + apply + language_required), skills/terminology/SKILL.md (Exit codes table + per-cmd mentions), skills/terminology/references/error-codes.md (catalog + per-cmd tables; dropped legacy under_construction row that collided with 75). Historical notes added to specs 001 and 004 pointing to ADR 0001. Release notes drafted at docs/release-notes/exit-code-taxonomy.md. AGENTS.md needed no change (no literal codes). Repo owner: the user-level terminology skill outside this repo mirrors SKILL.md and needs the same update when this ships. make build passes.
