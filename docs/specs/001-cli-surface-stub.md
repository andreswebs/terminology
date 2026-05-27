# E1 — CLI surface stub

> **Status**: APPROVED. Decisions locked in the table below. No
> business logic in this epic; that lands in E2–E10 via TDD vertical
> slices, one command at a time.

## Decisions

| Tag | Decision                                                                                                                                                                              |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Q1  | **Stub exit code `75`** (`EX_TEMPFAIL`) + new error code `under_construction`. Registered as a `terr` sentinel.                                                                       |
| Q2  | **Conventional short aliases** added per the locked map below.                                                                                                                        |
| Q3  | **Enforce closed picklists at the urfave layer** now (`--format`, `--status`, `--part-of-speech`, `--register`, `--grammatical-gender`, `--script`). Free-form hardening deferred to E9. |
| Q4  | `--fields` declared as a **shared `[]cli.Flag` slice**, included per read command (`validate`, `lookup`, `scan`, `check`, `extract`). Absent from write/utility commands.              |
| Q5  | `--author` and `TERMINOLOGY_AUTHOR` are **per-write-command**, mirroring the rest of the write-command surface.                                                                       |
| Q6  | Package name: **`internal/app`**. Inside it, the import alias `urfcli "github.com/urfave/cli/v3"` is used when naming collisions hurt readability.                                    |
| Q7  | File grouping: **flat with prefix** under `internal/app/commands/` (e.g. `concept_add.go`, `term_deprecate.go`). Re-evaluate per parent if any file exceeds ~200 lines.                |
| Q8  | Module path is `github.com/andreswebs/terminology`, root at `src/go.mod`. New code goes under `src/internal/app/`.                                                                    |
| Q9  | **Wire `--version` on root** now, populated from `runtime/debug.BuildInfo` (with `-ldflags` override available).                                                                      |
| Q10 | Bare `terminology` invocation: **exit `2`** with subcommand help printed to stderr. Achieved via a tiny root `Action` calling `cli.ShowSubcommandHelpAndExit(cmd, 2)`.                 |
| Q11 | **No top-level aliases** (no `terminology ls`, no `terminology verify`). Can be added later without breaking changes.                                                                  |

### Short-alias map

Reserved by urfave defaults: **`-h`** (help), **`-v`** (version).

| Flag                  | Short | Scope                                                                        |
| --------------------- | ----- | ---------------------------------------------------------------------------- |
| `--tbx` (global)      | `-T`  | every command                                                                |
| `--format` (global)   | —     | every command                                                                |
| `--fields` (read set) | `-F`  | `validate`, `lookup`, `scan`, `check`, `extract`                             |
| `--lang`              | `-l`  | `lookup`, `scan`, `extract`, `concept add`, `term add`, `term deprecate`     |
| `--term`              | `-t`  | `concept add`, `term add`, `term deprecate`                                  |
| `--status`            | `-s`  | `term add`                                                                   |
| `--part-of-speech`    | `-p`  | `term add`                                                                   |
| `--register`          | `-r`  | `term add`                                                                   |
| `--grammatical-gender`| —     | `term add`                                                                   |
| `--source-lang`       | `-S`  | `check` (uppercase to stay distinct from `--status`)                         |
| `--target-lang`       | —     | `check` (`-t`/`-T` already taken)                                            |
| `--strict`            | —     | `validate`, `check`                                                          |
| `--min-freq`          | —     | `extract`                                                                    |
| `--exclude`           | `-x`  | `extract`                                                                    |
| `--stopwords`         | —     | `extract`                                                                    |
| `--script`            | —     | `extract` (`-s` reserved for `--status` for consistency)                     |
| `--id`                | `-i`  | `concept add`                                                                |
| `--subject-field`     | —     | `concept add`                                                                |
| `--canonical-lang`    | —     | `concept add`                                                                |
| `--context`           | —     | `scan`, `check`                                                              |
| `--dry-run`           | `-n`  | every write command                                                          |
| `--transaction`       | —     | every write command                                                          |
| `--author`            | `-a`  | every write command                                                          |
| `--merge`             | —     | `concept update` (mutex with `--replace`)                                    |
| `--replace`           | —     | `concept update` (mutex with `--merge`)                                      |
| `--force`             | —     | `concept remove` (destructive — verbose by design)                           |
| `--prune`             | —     | `apply` (destructive — verbose by design)                                    |
| `--file`              | `-f`  | `apply`                                                                      |
| `--command`           | —     | `schema`                                                                     |
| `--verbose`           | —     | global (boolean; no short to avoid `-v` collision with `--version`)          |
| `--debug`             | —     | global (boolean)                                                             |
| `--quiet`             | —     | global (boolean)                                                             |

Per-command flag-set sanity check passes: no two visible flags collide
on the same letter within any command's scope (including inherited
globals).

## Purpose

Stand up the **complete** `terminology` CLI command surface — every
command, subcommand, positional, flag, alias, env-var binding, and
help string defined in [`docs/cli-design.md`](../cli-design.md) —
wired up with `github.com/urfave/cli/v3`, but with **no business
logic**. Every action handler returns a structured "under
construction" response and a distinct exit code.

The output of this work is a binary you can already `--help` against,
that already validates flag presence/types at the urfave layer, and
that already produces the documented exit-code shape for the
not-yet-implemented portion. Subsequent tickets replace `TODO` action
bodies one by one without touching the surface.

The urfave declarations themselves are the canonical source of truth
for the command/flag surface (per
[schema-source-of-truth](../adr/schema-source-of-truth.md));
`terminology schema` reflects over them at runtime.

## Scope

In scope (this epic + the implementation tickets it spawns):

- One root `*cli.Command` named `terminology`.
- All top-level commands listed in
  [`docs/cli-design.md`](../cli-design.md): `validate`, `lookup`,
  `scan`, `check`, `extract`, `apply`, `schema`.
- All parent commands with subcommands: `concept` (`add`, `update`,
  `remove`), `term` (`add`, `deprecate`).
- Every flag, alias, env-var source, default, picklist enum,
  `Required` marker, and `--help` string per cli-design.md.
- Every positional argument with the right `Min`/`Max`/variadic shape.
- A single shared "under construction" action handler used by every
  command.
- A unit test per command verifying: parses without error on a
  representative argument vector, emits the documented stub envelope
  on stderr, exits with `75`.
- A unit test per command verifying that a deliberately invalid
  invocation (missing required flag, unknown subcommand, bad enum
  value) is **rejected by urfave** with a usage-error exit code.

Explicitly out of scope:

- TBX parsing, scanning, checking, extraction, apply reconciliation,
  ID derivation, hardening, field-mask projection, schema generation.
- Any reading or writing of files. The stub never touches the
  filesystem.
- `terminology schema` reflective output. The reflective walker is
  scaffolded in E4; in E1 `schema` stubs like the rest.
- Shell completion. Defer until the surface is real.
- The `internal/tbx`, `internal/match`, `internal/extract`,
  `internal/write`, `internal/harden`, `internal/markdown` packages.
  None of those exist yet and the stub does not pull them in.

## Under-construction contract

Every stub handler returns the **same** response shape, parameterized
by the command name. Agents calling the binary against an unfinished
commit get a parseable error envelope, not silent success.

### Response shape (JSON, default)

`--format json` (default) — written to **stderr**:

```json
{
  "schema_version": 1,
  "ok": false,
  "error": {
    "code": "under_construction",
    "message": "terminology <command.dotted.path> is not implemented yet",
    "hint": "track progress in .tickets/ or rebuild from a newer commit"
  }
}
```

Stdout is **empty** in this mode.

### Response shape (text)

`--format text` — written to **stderr**, single line:

```
✗ terminology <command.dotted.path>: not implemented yet
```

Stdout is empty.

### Exit code

`75` (`EX_TEMPFAIL` from `sysexits.h`) — registered alongside
`under_construction` in the `terr` sentinel registry. The
not-implemented state is a first-class observable, not aliased on top
of a real failure mode.

### Flag and argument parsing posture

urfave parses **everything** before the stub handler runs. That means:

- `--lang ""` → urfave accepts (string flag, no validator yet); stub
  fires.
- `--status not-a-real-value` on `term add` → urfave **rejects**
  because the flag is declared as a closed enum (Q3); process exits
  with urfave's usage-error code without the stub firing.
- Missing `--lang`/`--term` on `term add` → urfave rejects (flags
  marked `Required: true`); stub does not fire.
- Missing positional `ID` on `concept update` → urfave rejects
  (argument `Min: 1`); stub does not fire.
- Unknown flag → urfave rejects (no `AllowExtFlags`); stub does not
  fire.
- `--help` / `-h` → urfave handles, prints help, exits `0`.
- `--version` / `-v` on root → urfave handles, prints
  `version.Current()`, exits `0`.

Invalid invocations are rejected **before** any business logic exists.
The stub fires only on shape-valid invocations.

## Surface inventory

### Top-level

| Command       | urfave name | Parent | Stubs to emit                   |
| ------------- | ----------- | ------ | ------------------------------- |
| `validate`    | `validate`  | root   | `under_construction(validate)`  |
| `lookup`      | `lookup`    | root   | `under_construction(lookup)`    |
| `scan`        | `scan`      | root   | `under_construction(scan)`      |
| `check`       | `check`     | root   | `under_construction(check)`     |
| `extract`     | `extract`   | root   | `under_construction(extract)`   |
| `apply`       | `apply`     | root   | `under_construction(apply)`     |
| `schema`      | `schema`    | root   | `under_construction(schema)`    |
| `concept`     | `concept`   | root   | help-only parent (no `Action`)  |
| `term`        | `term`      | root   | help-only parent (no `Action`)  |

### Parent → subcommand

| Name               | urfave path           | Stub                                |
| ------------------ | --------------------- | ----------------------------------- |
| `concept.add`      | `concept add`         | `under_construction(concept.add)`   |
| `concept.update`   | `concept update`      | `under_construction(concept.update)`|
| `concept.remove`   | `concept remove`      | `under_construction(concept.remove)`|
| `term.add`         | `term add`            | `under_construction(term.add)`      |
| `term.deprecate`   | `term deprecate`      | `under_construction(term.deprecate)`|

A `concept` or `term` invocation with no subcommand prints subcommand
help and exits `2` (urfave default for "missing required subcommand";
relies on not setting an `Action` on the parent — Q10).

## Global flags

These live on the **root** `*cli.Command` and are inherited by every
subcommand. urfave v3 inherits root-level flags by default unless
`Local: true` is set.

| Flag        | Type         | Default | Env source           | Notes                                                                                  |
| ----------- | ------------ | ------- | -------------------- | -------------------------------------------------------------------------------------- |
| `--tbx`     | `StringFlag` | (none)  | `TERMINOLOGY_TBX`    | `TakesFile: true`. Not `Required` at urfave layer — `no_tbx_path` is a runtime concern. |
| `--format`  | `StringFlag` | `json`  | —                    | Validator restricts to `{json, text}` (Q3).                                            |
| `--verbose` | `BoolFlag`   | `false` | —                    | Verbosity (per [logging](../adr/logging.md)).                  |
| `--debug`   | `BoolFlag`   | `false` | —                    | Maximum verbosity.                                                                     |
| `--quiet`   | `BoolFlag`   | `false` | —                    | Suppress non-error output.                                                             |

`--fields` is **not** a global flag (Q4); see the per-command surface
below.

## Per-command surface

Each command declaration in code mirrors the contract documented in
[`docs/cli-design.md`](../cli-design.md): same `Name`, same `Aliases`,
same `Usage` string, same `ArgsUsage`, same `Arguments` shape, same
`Flags` list (with `Required`, `Sources`, `Value`, picklist enums for
enum flags). Action body in every case is the shared
`underConstruction(cmd)` helper. **TODO** markers in the code
reference the corresponding section of `cli-design.md`.

The urfave declarations themselves are the canonical surface; CI
asserts that `docs/examples/scenarios.md` invocations all parse and
that the generated `docs/reference/errors.md` matches the live `terr`
sentinel registry (per [E9](009-hardening.md)).

Nuances worth surfacing here:

- `lookup TERM` — single required positional, `StringArg{Name: "term", Min: 1, Max: 1}`.
- `scan FILE` — single required positional, path.
- `check SRC TGT` — two required positionals.
- `extract FILE...` — variadic positional, `StringArgs{Name: "files", Min: 1, Max: -1}`.
- `concept update ID`, `concept remove ID`, `term add ID`, `term deprecate ID`
  — each takes a single required positional ID.
- `concept add` — no positionals; ID comes from `--id` or stdin payload.
- `apply` — no positionals; payload comes from `--file` / `-f`.
- `schema` — no positionals.

### Picklist values

The closed picklists (`--format`, `--status`, `--part-of-speech`,
`--register`, `--grammatical-gender`, `--script`) source their
accepted values from `internal/tbx/picklist.go` (per
[E3](003-validate-command.md)). The urfave flag declaration imports
the same constants used by the reader-side validator — single source,
drift impossible.

## Package layout

```
src/
├── cmd/terminology/
│   └── main.go                  # thin entry: build root, Run, emit-and-exit
└── internal/
    ├── app/
    │   ├── root.go              # Root() — assembles tree, global flags, slog Before hook
    │   ├── stub.go              # underConstruction(cmd) + envelope emission
    │   ├── stub_test.go         # tests the stub helper in isolation
    │   ├── commands/
    │   │   ├── validate.go
    │   │   ├── lookup.go
    │   │   ├── scan.go
    │   │   ├── check.go
    │   │   ├── extract.go
    │   │   ├── apply.go
    │   │   ├── schema.go
    │   │   ├── concept.go       # parent with subcommands
    │   │   ├── concept_add.go
    │   │   ├── concept_update.go
    │   │   ├── concept_remove.go
    │   │   ├── term.go          # parent with subcommands
    │   │   ├── term_add.go
    │   │   └── term_deprecate.go
    │   └── commands_test.go     # one test per command; table-driven
    ├── output/
    │   ├── errors.go            # EmitError, ExitCodeFor — single-point emission
    │   └── version.go           # SchemaVersion constant (int)
    ├── terr/
    │   └── terr.go              # type E + New(code, exit, hint, message) + UnderConstruction()
    ├── logctx/
    │   └── logctx.go            # From(ctx), With(ctx, logger)
    └── version/
        └── version.go           # Current() — BuildInfo + -ldflags override
```

Inside `internal/app`, the import alias
`urfcli "github.com/urfave/cli/v3"` is used when readability suffers
(Q6).

Rationale (Q7):

- One file per command, named after its dotted path. Easy to grep,
  easy to replace stubs one ticket at a time.
- `internal/app/root.go` owns the root tree + global flags + the
  slog-bootstrap `Before` hook; commands are pure constructors
  returning `*cli.Command`.
- Tests live in `internal/app/commands_test.go` (one big table-driven
  test per command makes scanning the matrix trivial) and
  `stub_test.go` (covers the shared helper).
- `cmd/terminology/main.go` shrinks to ~15 lines: build root, `Run`,
  delegate error/exit handling to `internal/output`.

## `main.go` shape

```go
package main

import (
    "context"
    "os"

    "github.com/andreswebs/terminology/internal/app"
    "github.com/andreswebs/terminology/internal/output"
)

func main() {
    cmd := app.Root()
    if err := cmd.Run(context.Background(), os.Args); err != nil {
        output.EmitError(cmd.ErrWriter, err)
        os.Exit(output.ExitCodeFor(err))
    }
}
```

The **single emit-and-exit point** referenced by
[error-handling](../adr/error-handling.md).
No other file calls `os.Exit` and no other file writes the error
envelope.

`output.EmitError` and `output.ExitCodeFor`, scaffolded in E1 with
minimum behavior:

- `EmitError` extracts a `*terr.E` via `errors.As`; falls back to
  `{code: "internal_error"}` for unknown errors; writes the envelope
  in `--format json` or `--format text` shape.
- `ExitCodeFor` extracts the typed exit code; falls back to `1`.

`internal/terr/terr.go` lands `type E struct {…}`, `func New(...)`,
and `func UnderConstruction(cmdPath string) *E` (code
`under_construction`, exit `75`, hint
`"track progress in .tickets/ or rebuild from a newer commit"`).
Subsequent epics add new sentinels alongside their originator
packages (per
[error-handling](../adr/error-handling.md));
E1 does not introduce any non-stub errors.

## Version wiring (Q9)

E1 lands `internal/version`:

```go
package version

var Override = ""  // overridable via -ldflags "-X .../version.Override=v0.x.y"

func Current() string {
    if Override != "" {
        return Override
    }
    if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "(devel)" {
        return bi.Main.Version
    }
    return "dev"
}
```

`internal/app/root.go` sets `Root().Version = version.Current()`, so
urfave populates `--version` / `-v` from a single source.

`Makefile` carries a `RELEASE_VERSION ?= ` variable threaded via
`-ldflags "-X .../version.Override=$(RELEASE_VERSION)"` in
`make release` (per [E10](010-release.md)).

## Logger bootstrap

E1 wires the logger per
[logging](../adr/logging.md):

- `internal/logctx/logctx.go` — `From(ctx) *slog.Logger`,
  `With(ctx, l)`.
- `internal/app/root.go`'s `Before` hook constructs a logger
  (level + handler from `--verbose`/`--debug`/`--quiet`, attributes
  `command` + `run_id` + `version`), stashes it on `ctx` via
  `logctx.With`, and passes the new context through.

The stub action does not log — it emits the envelope and returns the
typed error. Subsequent epics consume `logctx.From(ctx)`.

Verbosity flags are booleans (`--verbose`, `--debug`, `--quiet`) with
no short aliases — sidesteps the `-v` collision with urfave's
`--version` short.

## TDD plan

> Follows the vertical-slice tracer-bullet approach. **Do not** write
> all tests then all implementation. Each cycle is **red → green →
> next**.

### Tracer bullet (cycle 0)

**Goal**: prove the wiring end-to-end. Pick the smallest command
(`validate`) and prove that the binary builds, parses
`terminology validate`, runs the stub, and exits `75`.

1. **RED** — `commands_test.go` declares `TestValidateStub` that:
   - constructs `Root()`,
   - sets its `Writer` and `ErrWriter` to `bytes.Buffer`,
   - calls `cmd.Run(ctx, []string{"terminology", "validate"})`,
   - asserts the returned error implements `cli.ExitCoder` with exit
     `75`,
   - asserts stderr contains `"under_construction"` and `"validate"`,
   - asserts stdout is empty.
2. **GREEN** — implement `Root()` with a single `validate` subcommand
   whose `Action` is `underConstruction`. Implement `underConstruction`
   to write the JSON envelope to `cmd.ErrWriter` and return
   `cli.Exit(...)` with exit `75`.
3. Run `make build`. Both `make test` and the lint gate must pass.

The bullet proves: package layout, `Root()` constructor,
`underConstruction` helper, exit-code plumbing,
`cli.Command.Writer/ErrWriter` wiring for tests.

### Incremental cycles (1..N)

For **each remaining command in this order** (cheapest to richest):

1. `lookup` — single positional arg.
2. `scan` — single positional path.
3. `check` — two positionals + two enum-defaulted flags.
4. `extract` — variadic positional + `--script` enum.
5. `apply` — `Required` flag with short alias `-f`.
6. `schema` — utility, no args, `--command` only.
7. `concept add` — first parent+subcommand; tests `terminology
   concept` alone prints help and exits non-zero; tests the
   subcommand fires the stub.
8. `concept update` — single positional ID + mutex flags
   `--merge`/`--replace`.
9. `concept remove` — positional ID + `--force`.
10. `term add` — required `--lang`/`--term`, picklist `--status`.
11. `term deprecate` — last command.

Each cycle:

- **RED**: add a test entry (table row) covering: stub fires on valid
  invocation; missing required positional/flag rejected by urfave;
  (for picklist-bearing commands) invalid enum rejected.
- **GREEN**: add the `<command>.go` constructor wired into the tree.

After cycle 11: full surface lit up, full test matrix green,
`make build` clean. **Stop**. Refactor pass extracts common flag
groups on green.

### Refactor candidates (post-green)

- Shared `writeFlags()` returning the `--dry-run`/`--transaction`/
  `--author` triple, included by every write command.
- Shared `langFlag()` / `termFlag()` constructors with
  `Required: true` variants.
- Shared `readFieldsFlag()` for the `--fields` projection.
- Inline picklist-enum-flag helper consuming
  `internal/tbx/picklist.go` constants.

## Verification (`make build`)

After every cycle:

```sh
make build
```

The quality gate (`fmt-check` + `vet` + `lint` + `test`) must pass
before proceeding. If lint complains about unused parameters in
`underConstruction`, handle them properly — the stub uses
`cmd.Name()` and `cmd.FullName()`, so unused-parameter warnings
should not trigger.

## Out of scope (reminder)

This epic lands the **shell only**. It does not:

- Read, parse, or write any TBX file.
- Read any markdown.
- Implement field-mask projection, hardening rules beyond enum
  validators, the `output` envelope helpers' richer modes, or the
  reflective `terminology schema` walker.
- Touch `internal/tbx`, `internal/match`, `internal/markdown`,
  `internal/extract`, `internal/write`, `internal/harden`.

Each is a follow-up epic that **replaces a stub's body**, leaving the
surface (and the tests for the surface) untouched.

## Hand-offs to other epics

- TBX read/write & domain model → [E2](002-domain-and-io.md)
- `validate` real body → [E3](003-validate-command.md)
- `lookup` / `schema` (reflective) / `extract` real bodies → [E4](004-read-commands.md)
- Matcher policy & engine → [E5](005-matcher.md)
- `scan` / `check` real bodies → [E6](006-scan-check.md)
- `concept` / `term` real bodies → [E7](007-write-commands.md)
- `apply` real body & payload format → [E8](008-apply.md)
- Input hardening, fuzz, perf, security → [E9](009-hardening.md)
- Versioning policy, distribution → [E10](010-release.md)

Cross-cutting concerns established at E1 and respected by every epic:

- [error-handling](../adr/error-handling.md)
- [logging](../adr/logging.md)
- [testing](../adr/testing.md)
- [schema-source-of-truth](../adr/schema-source-of-truth.md)
- [determinism](../adr/determinism.md)
