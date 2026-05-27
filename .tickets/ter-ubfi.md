---
id: ter-ubfi
status: closed
deps: [ter-0l8i]
links: []
created: 2026-05-22T19:51:57Z
type: task
priority: 3
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, command, term, write]
---
# E1.T15 — Command stub: term parent + term add

## Goal

Wire the `term` parent (help-only, no Action) and its first subcommand `term add ID --lang LANG --term TERM`. Bundles the parent here so T16 only adds `term deprecate`.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → term parent + term.add", §"Short-alias map" (rows: --lang/-l Required, --term/-t Required, --status/-s, --part-of-speech/-p, --register/-r, --grammatical-gender, --dry-run/-n, --transaction, --author/-a)
- cli-design.md: [§\`terminology term add ID\`](docs/cli-design.md)

## Depends on

- T5

## Files to create / edit

- `src/internal/app/commands/term.go` — parent (help-only)
- `src/internal/app/commands/term_add.go` — `TermAdd() *cli.Command`
- `src/internal/app/testdata/term_add/{basic,missing-id,missing-lang,missing-term,invalid-status}.{argv,stdout,stderr,exit}`
- `src/internal/app/testdata/term/help.{argv,stdout,stderr,exit}`
- `src/internal/app/commands_test.go` — append term rows
- Edit `internal/app/root.go` Commands slice — register `commands.Term()`

## Surface

### Parent

```go
func Term() *cli.Command {
    return &cli.Command{
        Name:  "term",
        Usage: "add or deprecate a term within an existing concept",
        Commands: []*cli.Command{
            TermAdd(),
            // T16 appends TermDeprecate().
        },
    }
}
```

### `term add`

```go
func TermAdd() *cli.Command {
    return &cli.Command{
        Name:      "add",
        Usage:     "add a term to an existing concept's langSec",
        ArgsUsage: "ID",
        Arguments: []cli.Argument{
            &cli.StringArg{Name: "id", Min: 1, Max: 1},
        },
        Flags: []cli.Flag{
            &cli.StringFlag{Name: "lang",              Aliases: []string{"l"}, Usage: "language tag",       Required: true},
            &cli.StringFlag{Name: "term",              Aliases: []string{"t"}, Usage: "surface form",      Required: true},
            &cli.StringFlag{Name: "status",            Aliases: []string{"s"}, Usage: "administrative status", Validator: enumValidator(tbx.AdminStatus())},
            &cli.StringFlag{Name: "part-of-speech",    Aliases: []string{"p"}, Usage: "part of speech",        Validator: enumValidator(tbx.PartOfSpeech())},
            &cli.StringFlag{Name: "register",          Aliases: []string{"r"}, Usage: "register",             Validator: enumValidator(tbx.Register())},
            &cli.StringFlag{Name: "grammatical-gender",                         Usage: "grammatical gender",   Validator: enumValidator(tbx.GrammaticalGender())},
            &cli.BoolFlag  {Name: "dry-run",           Aliases: []string{"n"}, Usage: "validate and preview without writing"},
            &cli.BoolFlag  {Name: "transaction",                                Usage: "append a <transacGrp> record"},
            &cli.StringFlag{Name: "author",            Aliases: []string{"a"}, Usage: "responsibility value", Sources: cli.EnvVars("TERMINOLOGY_AUTHOR")},
        },
        Action: underConstruction,
    }
}
```

## TDD plan

1. **RED** `term_add/basic` golden: argv `[\"terminology\", \"term\", \"add\", \"tzimtzum\", \"--lang\", \"es\", \"--term\", \"tzimtzum\"]`; exit 75 + stub envelope.
2. **RED** `term_add/missing-id`: argv without ID; non-zero exit.
3. **RED** `term_add/missing-lang`: argv without `--lang`; non-zero exit (urfave Required).
4. **RED** `term_add/missing-term`: argv without `--term`; non-zero exit.
5. **RED** `term_add/invalid-status`: `--status nope`; urfave rejects.
6. **RED** `term/help` (parent alone): argv `[\"terminology\", \"term\"]`; non-zero exit with help.

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes

## Out of scope

- Lang-sec auto-creation logic (E7).
- Real term insertion (E7).


## Notes

**2026-05-22T21:23:03Z**

Implemented term parent command and term add subcommand stub. Created term.go (parent with errNoTermSubcommand sentinel, exit 2 on bare invocation) and term_add.go (stub with all flags per spec: --lang/-l Required, --term/-t Required, --status/-s, --part-of-speech/-p, --register/-r, --grammatical-gender, --dry-run/-n, --transaction, --author/-a with TERMINOLOGY_AUTHOR env). Registered Term() in root.go. Uses per-command termEnumValidator following T9/T12 pattern. StringArg (singular) for positional ID per T6 learning. Golden tests for both term (bare) and term_add (basic). Additional tests: missing-lang, missing-term, missing-id, invalid-status, all-flags, author-via-env. make build clean.
