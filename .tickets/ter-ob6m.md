---
id: ter-ob6m
status: closed
deps: [ter-0l8i]
links: []
created: 2026-05-22T19:48:25Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, command, extract]
---
# E1.T9 — Command stub: extract

## Goal

Wire `terminology extract FILE...` as a stub with a closed enum on `--script` (Q3 enforces picklist at urfave layer).

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → extract", §"Short-alias map" (rows: --exclude/-x, --script, --stopwords, --min-freq, --lang/-l, --fields/-F)
- cli-design.md: [§\`terminology extract FILE...\`](docs/cli-design.md)

## Depends on

- T5

## Files to create / edit

- `src/internal/app/commands/extract.go`
- `src/internal/app/testdata/extract/basic.{argv,stdout,stderr,exit}`
- `src/internal/app/commands_test.go` — append extract rows
- Edit `internal/app/root.go` Commands slice

## Surface

```go
&cli.Command{
    Name:      "extract",
    Usage:     "surface candidate terms from a markdown corpus",
    ArgsUsage: "FILE...",
    Arguments: []cli.Argument{
        &cli.StringArgs{Name: "files", Min: 1, Max: -1},  // variadic
    },
    Flags: []cli.Flag{
        &cli.StringFlag{Name: "exclude", Aliases: []string{"x"}, Usage: "exclude terms already in this TBX"},
        &cli.StringFlag{
            Name:      "script",
            Usage:     "filter by script: latin, hebrew, cyrillic, arabic, any",
            Value:     "any",
            Validator: enumValidator(tbx.Script()),
        },
        &cli.StringFlag{Name: "lang",      Aliases: []string{"l"}, Usage: "language of the corpus"},
        &cli.StringFlag{Name: "stopwords", Usage: "path to a newline-separated stopwords file"},
        &cli.IntFlag   {Name: "min-freq",  Value: 3,                Usage: "minimum frequency for high-frequency heuristic"},
        &cli.StringFlag{Name: "fields",    Aliases: []string{"F"}, Usage: "comma-separated dotted paths to include"},
    },
    Action: underConstruction,
}
```

## TDD plan

1. **RED** `extract/basic` golden: argv `[\"terminology\", \"extract\", \"a.md\", \"b.md\"]`; assert exit 75, stub envelope. **GREEN** add Extract() + register.
2. **RED** `extract/missing-files`: argv `[\"terminology\", \"extract\"]`; non-zero exit.
3. **RED** `extract/invalid-script`: argv `[..., \"--script\", \"klingon\"]`; non-zero exit before stub fires (proves the closed-enum validator works). **GREEN** ensured by enumValidator(tbx.Script()).
4. **RED** `extract/valid-script`: argv `[..., \"--script\", \"hebrew\"]`; stub fires.

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes

## Out of scope

- Heuristic engine (E4).
- Stoplist + script detection (E4).


## Notes

**2026-05-22T21:07:58Z**

Implemented extract command stub with: variadic StringArgs (Min:1, Max:-1), --script enum validator using tbx.Script() picklist, --exclude/-x, --lang/-l, --stopwords, --min-freq (default 3), --fields/-F. Tests: ExitCode75, MissingFiles, InvalidScript, ValidScript, Golden. The scriptValidator was created locally in commands/extract.go rather than reusing root.go's enumValidator (different package). The urfave validator Exit code is not extractable by output.ExitCodeFor (it only handles terr.Coded), consistent with existing TestRoot_InvalidFormat_Errors pattern.
