---
id: ter-0l8i
status: closed
deps: [ter-zwxt, ter-usuw, ter-qxff, ter-bsat]
links: []
created: 2026-05-22T19:44:19Z
type: task
priority: 1
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, tracer, validate]
---
# E1.T5 — TRACER: validate stub end-to-end (Root + stub + main + harness)

## Goal (tracer bullet)

Prove the entire wiring end-to-end with the smallest possible command (`validate`). Lands every piece of infrastructure that every subsequent command ticket (T6–T16) inherits:

- `internal/tbx/picklist.go` — closed picklist enum values (used by urfave validators)
- `internal/app/root.go` — `Root()` constructor, global flags, Before-hook logger bootstrap, verbosity-mutex check
- `internal/app/errors.go` — `ErrConflictingVerbosity` sentinel
- `internal/app/stub.go` — `underConstruction` shared action handler
- `internal/app/commands/validate.go` — the tracer bullet's command constructor
- `src/cmd/terminology/main.go` — rewrites the placeholder into the spec's emit-and-exit shape
- Golden-file test harness (`runGolden`) + `testdata/validate/clean.{stdout,stderr,exit}`

After this ticket, T6–T16 each add a single `commands/<name>.go` + a single golden directory.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"TDD plan → Tracer bullet (cycle 0)", §"Logger bootstrap", §"Package layout", §"Global flags"
- E1 short-alias table: [001-cli-surface-stub.md §Short-alias map](docs/specs/001-cli-surface-stub.md) — authoritative for every flag/alias
- Error-handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md) §"Layer 3 — Envelope emission"
- Logging ADR: [docs/adr/logging.md](docs/adr/logging.md) §"Verbosity controls" + §"Wiring"
- Testing ADR: [docs/adr/testing.md](docs/adr/testing.md) §"Golden CLI tests" + §"Envelope conformance"

## Depends on

- T1 (`internal/terr`)
- T2 (`internal/output`)
- T3 (`internal/version`)
- T4 (`internal/logctx`)

## Files to create / edit

| Path | Purpose |
| --- | --- |
| `src/internal/tbx/picklist.go` | Closed picklist enum values |
| `src/internal/tbx/picklist_test.go` | Smoke tests for the constants |
| `src/internal/app/root.go` | `Root()` + global flags + Before hook |
| `src/internal/app/root_test.go` | Root-level tests (verbosity mutex, --version) |
| `src/internal/app/errors.go` | `ErrConflictingVerbosity` sentinel |
| `src/internal/app/stub.go` | `underConstruction` helper |
| `src/internal/app/stub_test.go` | Stub-helper isolation tests |
| `src/internal/app/commands/validate.go` | Validate() command constructor |
| `src/internal/app/golden_test.go` | `runGolden` shared test harness |
| `src/internal/app/commands_test.go` | Table-driven per-command tests (validate row only here) |
| `src/internal/app/testdata/validate/clean.argv` | Golden input: argv as one-arg-per-line |
| `src/internal/app/testdata/validate/clean.stdout` | Golden stdout (empty) |
| `src/internal/app/testdata/validate/clean.stderr` | Golden stderr (JSON envelope) |
| `src/internal/app/testdata/validate/clean.exit` | Golden exit code ("75\n") |
| `src/cmd/terminology/main.go` | Rewrites placeholder to spec shape |

## Picklist seed (`internal/tbx/picklist.go`)

Source values per [docs/cli-design.md §Picklist values](docs/cli-design.md):

```go
package tbx

func Format() []string { return []string{"json", "text"} }

// AdminStatus accepts modern and legacy bare forms. The urfave layer
// accepts either; the writer (E7) normalizes to the modern -admn-sts form.
func AdminStatus() []string {
    return []string{
        "preferredTerm-admn-sts", "admittedTerm-admn-sts",
        "deprecatedTerm-admn-sts", "supersededTerm-admn-sts",
        // legacy bare forms (accepted on read; normalized on write per E3/E7)
        "preferredTerm", "admittedTerm", "deprecatedTerm", "supersededTerm",
    }
}

func PartOfSpeech() []string {
    return []string{"noun", "verb", "adjective", "adverb", "other"}
}

func GrammaticalGender() []string {
    return []string{"masculine", "feminine", "neuter", "other"}
}

func Register() []string {
    return []string{
        "colloquialRegister", "neutralRegister", "technicalRegister",
        "in-houseRegister", "bench-levelRegister", "slangRegister",
        "vulgarRegister",
        // legacy alias accepted on read per cli-design.md
        "usageRegister",
    }
}

func Script() []string {
    return []string{"latin", "hebrew", "cyrillic", "arabic", "any"}
}
```

Picklist test (`picklist_test.go`) asserts each function returns a non-empty unique-element slice and that known canonical values are present (e.g. `\"json\"` in Format, `\"preferredTerm-admn-sts\"` in AdminStatus). Drift-impossible claim is intrinsic — there is no second declaration to compare against.

## `ErrConflictingVerbosity` (`internal/app/errors.go`)

```go
package app

import "github.com/andreswebs/terminology/internal/terr"

var ErrConflictingVerbosity = terr.New(
    "conflicting_verbosity", 2,
    "pass at most one of --verbose, --debug, --quiet",
    "--verbose, --debug, and --quiet are mutually exclusive",
)
```

This is the first non-stub sentinel introduced anywhere in the codebase.

## `Root()` (`internal/app/root.go`)

```go
package app

import (
    urfcli "github.com/urfave/cli/v3"
    "github.com/andreswebs/terminology/internal/app/commands"
    "github.com/andreswebs/terminology/internal/tbx"
    "github.com/andreswebs/terminology/internal/version"
)

func Root() *urfcli.Command {
    cmd := &urfcli.Command{
        Name:    "terminology",
        Usage:   "agent-driven, terminology-focused academic translation",
        Version: version.Current(),
        Flags:   globalFlags(),
        Before:  beforeHook,         // logger bootstrap + verbosity mutex
        Action:  rootHelpAction,     // prints subcommand help, exits 2 (Q10)
        Commands: []*urfcli.Command{
            commands.Validate(),
            // T6–T16 each append their own command here
        },
    }
    return cmd
}

func globalFlags() []urfcli.Flag {
    return []urfcli.Flag{
        &urfcli.StringFlag{
            Name:      "tbx",
            Aliases:   []string{"T"},   // per Q-table short-alias map
            Usage:     "path to TBX glossary file",
            Sources:   urfcli.EnvVars("TERMINOLOGY_TBX"),
            TakesFile: true,
        },
        &urfcli.StringFlag{
            Name:  "format",
            Usage: "output format: json or text",
            Value: "json",
            Validator: enumValidator(tbx.Format()),  // see picklist helper
        },
        &urfcli.BoolFlag{Name: "verbose", Usage: "INFO-level diagnostics"},
        &urfcli.BoolFlag{Name: "debug",   Usage: "DEBUG-level diagnostics"},
        &urfcli.BoolFlag{Name: "quiet",   Usage: "ERROR-only diagnostics"},
    }
}
```

### Before hook

```go
func beforeHook(ctx context.Context, cmd *urfcli.Command) (context.Context, error) {
    // Verbosity mutex (per docs/adr/logging.md §"Verbosity controls").
    if bools(cmd.Bool("verbose"), cmd.Bool("debug"), cmd.Bool("quiet")) > 1 {
        return ctx, ErrConflictingVerbosity
    }
    // Logger bootstrap (per docs/adr/logging.md §"Wiring").
    level := slog.LevelWarn
    switch {
    case cmd.Bool("debug"):   level = slog.LevelDebug
    case cmd.Bool("verbose"): level = slog.LevelInfo
    case cmd.Bool("quiet"):   level = slog.LevelError
    }
    var h slog.Handler
    opts := &slog.HandlerOptions{Level: level}
    if term.IsTerminal(int(os.Stderr.Fd())) {
        h = slog.NewTextHandler(os.Stderr, opts)
    } else {
        h = slog.NewJSONHandler(os.Stderr, opts)
    }
    logger := slog.New(h).With(
        "command", cmd.FullName(),
        "run_id",  logctx.NewRunID(),
        "version", version.Current(),
    )
    return logctx.With(ctx, logger), nil
}
```

`bools` is a tiny helper that counts true values. `enumValidator` returns an urfave Validator func that errors when the value is not in the allowed slice.

### Root action (Q10)

```go
func rootHelpAction(ctx context.Context, cmd *urfcli.Command) error {
    if cmd.Args().Len() == 0 && cmd.NumFlags() == 0 {
        return urfcli.ShowSubcommandHelp(cmd)  // exits via urfave; we set exit code on return
    }
    // urfave routes to this when no recognized subcommand matched.
    return urfcli.Exit("unknown subcommand; run 'terminology --help'", 2)
}
```

Note: urfave v3 emits subcommand help via `ShowSubcommandHelpAndExit(cmd, 2)` or by returning a coded error after printing help. Pick whichever matches urfave v3's current API and document the choice in the file.

## `underConstruction` (`internal/app/stub.go`)

```go
package app

import (
    "github.com/andreswebs/terminology/internal/output"
    "github.com/andreswebs/terminology/internal/terr"
    urfcli "github.com/urfave/cli/v3"
)

// underConstruction is the shared action handler used by every stub command.
// It writes the structured error envelope to cmd.ErrWriter (so tests can
// capture it via bytes.Buffer) and returns the typed error so main.go's
// exit handler picks up the exit code.
func underConstruction(ctx context.Context, cmd *urfcli.Command) error {
    err := terr.UnderConstruction(cmd.FullName())
    format := cmd.Root().String("format")
    output.EmitError(cmd.Root().ErrWriter, format, err)
    return err
}
```

Note: `cmd.Root()` is required to reach the global `--format` flag (per urfave v3's flag-inheritance model). If the actual API differs, follow whichever urfave v3 idiom yields the global flag value.

## `commands.Validate()` (`internal/app/commands/validate.go`)

```go
package commands

import urfcli "github.com/urfave/cli/v3"

func Validate() *urfcli.Command {
    return &urfcli.Command{
        Name:  "validate",
        Usage: "validate a TBX file against the supported subset",
        Flags: []urfcli.Flag{
            &urfcli.BoolFlag{Name: "strict", Usage: "promote unknown elements and unresolved IDREFs to errors"},
            // --fields (-F) is a read-set flag per Q4; the shared helper lands in T17.
            // For T5, declare it inline:
            &urfcli.StringFlag{Name: "fields", Aliases: []string{"F"}, Usage: "comma-separated dotted paths to include in output"},
        },
        Action: app.UnderConstruction,  // exported re-export of the unexported helper, or move helper to commands package
    }
}
```

(Final placement of `underConstruction` is a judgement call: either exported from `internal/app` or moved into `internal/app/commands` as a package-private helper. Pick one; document the choice; T6–T16 reuse it.)

## `main.go` (`src/cmd/terminology/main.go`)

Rewrite the placeholder to the spec's shape:

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
        format := cmd.String("format")
        // Stub paths already wrote the envelope from underConstruction.
        // This branch covers errors that bypass that path
        // (e.g. ErrConflictingVerbosity from the Before hook).
        if !alreadyEmitted(err) {
            output.EmitError(cmd.ErrWriter, format, err)
        }
        os.Exit(output.ExitCodeFor(err))
    }
}
```

The `alreadyEmitted` guard avoids double-printing the envelope for stub errors. Implementation choice: either a sentinel-wrapper interface, or have `underConstruction` return a special wrapper that satisfies `Coded` but signals \"already emitted\". Simplest: keep stub path emission **only** in main.go — `underConstruction` just returns the typed error, main.go emits. That removes the need for any \"already emitted\" flag. **Adopt this simpler design**: the stub action returns `terr.UnderConstruction(cmd.FullName())`; `main.go` is the single emit-and-exit point per ADR. The envelope is written exactly once.

**Revised `underConstruction` (simpler):**

```go
func underConstruction(ctx context.Context, cmd *urfcli.Command) error {
    return terr.UnderConstruction(cmd.FullName())
}
```

This makes main.go the single envelope sink — matching docs/adr/error-handling.md §"Layer 3". Update the test harness expectations accordingly.

## Golden harness (`internal/app/golden_test.go`)

```go
// runGolden runs argv against a fresh Root() with stdout/stderr/stdin
// captured in buffers, then compares against testdata/<name>/clean.{stdout,stderr,exit}.
// The -update flag rewrites goldens.
//
// Per docs/adr/testing.md:
//   - byte-for-byte comparison (no JSON structural compare)
//   - CRLF normalized to LF on fixture read
//   - envelope conformance enforced via output.AssertEnvelopeShape
//     for any non-empty stderr starting with '{'
func runGolden(t *testing.T, name string, argv []string, stdin string)
```

Layout: `testdata/<cmd>/<scenario>.argv` (one arg per line, comments with `#` allowed), `.stdout`, `.stderr`, `.exit` (integer + trailing newline). For T5 only `testdata/validate/clean.*` exists.

A package-level `var update = flag.Bool(\"update\", false, \"rewrite goldens\")` controls regeneration.

## TDD plan (vertical slices — strict red→green→next)

Cycle 0: pick `validate` and prove the path. **Do not** write multiple tests before any implementation.

1. **RED** `TestValidate_Stub_Golden` — empty testdata; run `terminology validate`; harness fails because no golden files exist. **GREEN** create harness + commit goldens generated via `-update` after a passing run. Assert: exit code 75; stderr is the JSON envelope with `under_construction` + `validate`; stdout empty.
2. **RED** `TestRoot_BareInvocation_ExitsUsage` — `terminology` (no args, no flags); assert exit 2 + subcommand help on stderr. **GREEN** wire `rootHelpAction`.
3. **RED** `TestRoot_UnknownSubcommand_ExitsUsage` — `terminology nope`; assert urfave-default usage error (exit 2 path; envelope is urfave's, not ours — fine for now, documented).
4. **RED** `TestRoot_Version` — `terminology --version`; assert stdout contains the value returned by `version.Current()`, exit 0. **GREEN** wire Root().Version.
5. **RED** `TestRoot_VerbosityMutex` — table-driven for pairs (verbose+debug, verbose+quiet, debug+quiet, all three); assert envelope `conflicting_verbosity`, exit 2. **GREEN** wire mutex check in Before hook.
6. **RED** `TestStub_Helper_Direct` — call `underConstruction` directly with a constructed *cli.Command (set `Name`, `Parent`); assert returned *terr.E has code/exit/message containing the command's FullName. **GREEN** trivially passes if the helper just calls terr.UnderConstruction.
7. **RED** `TestPicklist_HasCanonicalValues` — assert `slices.Contains(tbx.Format(), \"json\")`, `tbx.AdminStatus()` contains `preferredTerm-admn-sts`, etc. **GREEN** seed picklist.go.

After each cycle: `make build` clean.

## Acceptance

- `make build` clean from project root
- `cd src && go test ./...` passes
- Running `./bin/terminology-$(go env GOOS)-$(go env GOARCH) validate` returns exit 75 with the JSON envelope on stderr (manual smoke)
- Running `./bin/terminology-* --version` prints the git-describe value (proves T3 + T5 wiring)
- Running `./bin/terminology-* --verbose --debug` exits 2 with the `conflicting_verbosity` envelope
- New dependency added to go.mod: `github.com/urfave/cli/v3` and `golang.org/x/term` (per spec dep posture)

## Out of scope

- Any command other than `validate` (T6–T16 each add one).
- Any business logic in validate (E3 replaces the stub body).
- Reflective `terminology schema` (E4).
- `--fields` projection validation (E4).
- Hardening (E9).
- Refactor of shared flag groups (T17).


## Notes

**2026-05-22T20:53:45Z**

Completed E1.T5 tracer bullet. Created: internal/tbx/picklist.go (closed picklist enums), internal/app/root.go (Root() constructor, global flags, Before-hook logger bootstrap, verbosity mutex), internal/app/errors.go (ErrConflictingVerbosity sentinel), internal/app/commands/validate.go (validate command + underConstruction helper in commands pkg), internal/app/golden_test.go (runGolden harness), internal/app/commands_test.go (table-driven tests), testdata/validate/clean.{stdout,stderr,exit}. Rewrote cmd/terminology/main.go to spec shape. Design choice: underConstruction lives in commands package (not app) to avoid import cycle. ExitErrHandler set to no-op on Root so urfave doesn't print errors — main.go is the single emit point per ADR. urfave v3 FullName() includes root name so stubPath() strips it to avoid 'terminology terminology validate' in messages. Added deps: urfave/cli/v3 v3.9.0, golang.org/x/term v0.43.0, golang.org/x/sys v0.44.0.
