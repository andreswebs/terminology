---
id: ter-s3uj
status: closed
deps: [ter-0l8i]
links: []
created: 2026-05-22T19:48:25Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, command, check]
---
# E1.T8 — Command stub: check

## Goal

Wire `terminology check SRC TGT` as a stub.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → check", §"Short-alias map" (rows: --source-lang/-S, --target-lang, --strict, --context, --fields)
- cli-design.md: [§\`terminology check SRC TGT\`](docs/cli-design.md) — note: defaults for --source-lang/--target-lang were removed (per E6 the language must come from frontmatter or the flag, otherwise `language_required` at runtime; E1 still declares the flags but no default values)

## Depends on

- T5

## Files to create / edit

- `src/internal/app/commands/check.go`
- `src/internal/app/testdata/check/basic.{argv,stdout,stderr,exit}`
- `src/internal/app/commands_test.go` — append check rows
- Edit `internal/app/root.go` Commands slice

## Surface

```go
&cli.Command{
    Name:      "check",
    Usage:     "verify a translated target file against the source given the glossary",
    ArgsUsage: "SRC TGT",
    Arguments: []cli.Argument{
        &cli.StringArgs{Name: "files", Min: 2, Max: 2},
        // OR two separate StringArg entries — pick whichever matches urfave v3's API
    },
    Flags: []cli.Flag{
        &cli.StringFlag{Name: "source-lang", Aliases: []string{"S"}, Usage: "source language"},
        &cli.StringFlag{Name: "target-lang", Usage: "target language"},
        &cli.BoolFlag  {Name: "strict", Usage: "admitted variants raise violations"},
        &cli.IntFlag   {Name: "context", Value: 80, Usage: "context window around each violation"},
        &cli.StringFlag{Name: "fields", Aliases: []string{"F"}, Usage: "comma-separated dotted paths to include"},
    },
    Action: underConstruction,
}
```

**Note**: `--source-lang` short is `-S` (uppercase) to stay distinct from `--status` (lowercase `-s`), per the E1 short-alias map.

## TDD plan

1. **RED** `check/basic` golden: argv `[\"terminology\", \"check\", \"src.md\", \"tgt.md\"]`; assert exit 75, stub envelope. **GREEN** add Check() + register.
2. **RED** `check/missing-target`: argv `[\"terminology\", \"check\", \"src.md\"]`; non-zero exit (urfave rejects). **GREEN** ensured by Min: 2.
3. **RED** `check/with-langs`: argv with `--source-lang es --target-lang en --strict`; stub fires.

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes

## Out of scope

- Frontmatter/flag/error language resolution (E6 — owns `ErrLanguageRequired`).
- Match/violation production (E5/E6).


## Notes

**2026-05-22T21:03:32Z**

Implemented check command stub with: two required positionals (SRC TGT) via StringArgs{Min:2, Max:2}, flags --source-lang/-S, --target-lang, --strict, --context (default 80), --fields/-F. Registered in root.go. Added 4 tests: golden file test, missing-target error, missing-both-args error, and with-langs-and-strict stub test. All tests pass, make build clean.
