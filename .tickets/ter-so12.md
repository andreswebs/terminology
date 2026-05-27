---
id: ter-so12
status: closed
deps: [ter-0l8i, ter-ubfi]
links: []
created: 2026-05-22T19:51:57Z
type: task
priority: 3
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, command, term, write]
---
# E1.T16 — Command stub: term deprecate

## Goal

Wire `terminology term deprecate ID --lang LANG --term TERM` as a stub.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → term.deprecate", §"Short-alias map" (rows: --lang/-l Required, --term/-t Required, --dry-run/-n, --transaction, --author/-a)
- cli-design.md: [§\`terminology term deprecate\`](docs/cli-design.md)

## Depends on

- T5
- T15 (appends to term parent's Commands slice)

## Files to create / edit

- `src/internal/app/commands/term_deprecate.go`
- `src/internal/app/commands/term.go` — append `TermDeprecate()` to the slice
- `src/internal/app/testdata/term_deprecate/{basic,missing-lang,missing-term,missing-id}.{argv,stdout,stderr,exit}`
- `src/internal/app/commands_test.go` — append rows

## Surface

```go
func TermDeprecate() *cli.Command {
    return &cli.Command{
        Name:      "deprecate",
        Usage:     "set an existing term's administrativeStatus to deprecatedTerm-admn-sts",
        ArgsUsage: "ID",
        Arguments: []cli.Argument{
            &cli.StringArg{Name: "id", Min: 1, Max: 1},
        },
        Flags: []cli.Flag{
            &cli.StringFlag{Name: "lang",        Aliases: []string{"l"}, Usage: "language tag", Required: true},
            &cli.StringFlag{Name: "term",        Aliases: []string{"t"}, Usage: "surface form", Required: true},
            &cli.BoolFlag  {Name: "dry-run",     Aliases: []string{"n"}, Usage: "validate and preview without writing"},
            &cli.BoolFlag  {Name: "transaction",                          Usage: "append a <transacGrp> record"},
            &cli.StringFlag{Name: "author",      Aliases: []string{"a"}, Usage: "responsibility value", Sources: cli.EnvVars("TERMINOLOGY_AUTHOR")},
        },
        Action: underConstruction,
    }
}
```

## TDD plan

1. **RED** `term_deprecate/basic` golden: argv `[\"terminology\", \"term\", \"deprecate\", \"tzimtzum\", \"--lang\", \"es\", \"--term\", \"contraction\"]`; exit 75 + stub envelope.
2. **RED** `term_deprecate/missing-lang`: argv without `--lang`; urfave rejects.
3. **RED** `term_deprecate/missing-term`: argv without `--term`; urfave rejects.
4. **RED** `term_deprecate/missing-id`: argv without ID; urfave rejects.

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes

## Out of scope

- not_found semantics when the term is absent (E7).
- Setting administrativeStatus on the actual file (E7).


## Notes

**2026-05-22T21:31:56Z**

Implemented term deprecate command stub. Created term_deprecate.go with the urfave command definition (ID positional, --lang/-l, --term/-t required flags, --dry-run/-n, --transaction, --author/-a with TERMINOLOGY_AUTHOR env source). Registered TermDeprecate() in term.go Commands slice. Added 6 tests: golden (basic), missing-lang, missing-term, missing-id, all-flags, author-via-env. All tests pass, make build clean.
