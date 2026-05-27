---
id: ter-nlnx
status: closed
deps: [ter-0l8i]
links: []
created: 2026-05-22T19:48:25Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, command, lookup]
---
# E1.T6 — Command stub: lookup

## Goal

Wire `terminology lookup TERM` as a stub. Returns `under_construction` on every valid invocation; urfave rejects bad shapes before the stub fires.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → lookup", §"Short-alias map" (rows: --tbx, --format, --fields, --lang), §"TDD plan → incremental cycles"
- cli-design.md: [§\`terminology lookup TERM\`](docs/cli-design.md) for Usage strings only — flag surface is authoritative in the spec above

## Depends on

- T5

## Files to create / edit

- `src/internal/app/commands/lookup.go` — `Lookup() *cli.Command`
- `src/internal/app/testdata/lookup/found.{argv,stdout,stderr,exit}` — golden for stub fires
- `src/internal/app/commands_test.go` — add lookup table row(s)
- Edit `src/internal/app/root.go` Commands slice: register `commands.Lookup()`

## Surface (declarative)

```go
&cli.Command{
    Name:      "lookup",
    Usage:     "look up a term across all languages in the TBX file",
    ArgsUsage: "TERM",
    Arguments: []cli.Argument{
        &cli.StringArg{Name: "term", Min: 1, Max: 1},
    },
    Flags: []cli.Flag{
        &cli.StringFlag{Name: "lang",   Aliases: []string{"l"}, Usage: "restrict to a language section"},
        &cli.StringFlag{Name: "fields", Aliases: []string{"F"}, Usage: "comma-separated dotted paths to include"},
    },
    Action: underConstruction,  // shared helper from T5
}
```

Inherited globals (`--tbx`/`-T`, `--format`, `--verbose`, `--debug`, `--quiet`) come from Root().

## TDD plan (vertical slices; one cycle per test)

1. **RED** `lookup/found` golden row: argv `[\"terminology\", \"lookup\", \"tzimtzum\"]`; assert exit 75, JSON envelope contains `under_construction` + `lookup`, stdout empty. **GREEN** add Lookup() and register in Root().
2. **RED** `lookup/missing-term` (in-test, no golden — assert urfave error): argv `[\"terminology\", \"lookup\"]`; assert non-zero exit, stderr mentions the missing positional. **GREEN** trivially passes if Min: 1 is set.
3. **RED** `lookup/with-lang-and-fields`: argv `[\"terminology\", \"lookup\", \"tzimtzum\", \"--lang\", \"es\", \"--fields\", \"definitions\"]`; same stub envelope (proves flag declarations parse). **GREEN** covered by Lookup().

Reuse the `runGolden` harness from T5.

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes
- All three test cases above green

## Out of scope

- Actual lookup logic (E4 replaces stub body).
- `--fields` path validation (E4 introduces `internal/output/fields.go`).


## Notes

**2026-05-22T20:57:55Z**

Implemented lookup command stub. Created src/internal/app/commands/lookup.go with StringArg positional, --lang/-l and --fields/-F flags, Action: underConstruction. Registered in root.go Commands slice. Added 3 tests: golden (exit 75, proper envelope), missing-term (urfave rejects), with-lang-and-fields (flags parse, stub fires). All tests green, make build clean.
