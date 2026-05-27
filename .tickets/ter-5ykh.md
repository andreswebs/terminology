---
id: ter-5ykh
status: closed
deps: [ter-0l8i]
links: []
created: 2026-05-22T19:48:25Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, command, scan]
---
# E1.T7 — Command stub: scan

## Goal

Wire `terminology scan FILE` as a stub.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → scan", §"Short-alias map" (rows: --lang, --fields, --context)
- cli-design.md: [§\`terminology scan FILE\`](docs/cli-design.md)

## Depends on

- T5

## Files to create / edit

- `src/internal/app/commands/scan.go`
- `src/internal/app/testdata/scan/basic.{argv,stdout,stderr,exit}`
- `src/internal/app/commands_test.go` — append scan rows
- Edit `internal/app/root.go` Commands slice

## Surface

```go
&cli.Command{
    Name:      "scan",
    Usage:     "find all glossary term occurrences in a markdown file",
    ArgsUsage: "FILE",
    Arguments: []cli.Argument{
        &cli.StringArg{Name: "file", Min: 1, Max: 1},
    },
    Flags: []cli.Flag{
        &cli.StringFlag{Name: "lang",    Aliases: []string{"l"}, Usage: "restrict to a language section"},
        &cli.IntFlag   {Name: "context", Value:   80,            Usage: "context window around each match (chars)"},
        &cli.StringFlag{Name: "fields",  Aliases: []string{"F"}, Usage: "comma-separated dotted paths to include"},
    },
    Action: underConstruction,
}
```

## TDD plan

1. **RED** `scan/basic` golden: argv `[\"terminology\", \"scan\", \"source/ch1.md\"]`; assert exit 75, stub envelope. **GREEN** add Scan() + register.
2. **RED** `scan/missing-file`: argv `[\"terminology\", \"scan\"]`; assert non-zero exit (urfave rejects). **GREEN** ensured by Min: 1.
3. **RED** `scan/with-context`: argv `[..., \"--context\", \"120\"]`; stub fires. **GREEN** covered.

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes

## Out of scope

- Real matcher (E5/E6).
- File I/O (the stub never touches the filesystem per E1 spec).


## Notes

**2026-05-22T21:00:47Z**

Implemented scan command stub following existing patterns from validate/lookup. Created src/internal/app/commands/scan.go with StringArg positional (singular form, no Min/Max per E1.T6 learning), --lang/-l, --context (default 80), --fields/-F flags. Registered in root.go. Added 3 tests: golden test, missing-file rejection, and flags-with-stub test. All pass, make build clean.
