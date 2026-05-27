---
id: ter-4pgc
status: closed
deps: [ter-0l8i]
links: []
created: 2026-05-22T19:48:25Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, command, apply]
---
# E1.T10 — Command stub: apply

## Goal

Wire `terminology apply --file PAYLOAD` as a stub. `--file` is Required + has short `-f`.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → apply", §"Short-alias map" (rows: --file/-f, --prune, --dry-run/-n, --transaction, --author/-a)
- cli-design.md: [§\`terminology apply --file PAYLOAD\`](docs/cli-design.md)

## Depends on

- T5

## Files to create / edit

- `src/internal/app/commands/apply.go`
- `src/internal/app/testdata/apply/basic.{argv,stdout,stderr,exit}`
- `src/internal/app/commands_test.go` — append apply rows
- Edit `internal/app/root.go` Commands slice

## Surface

```go
&cli.Command{
    Name:  "apply",
    Usage: "reconcile a declarative payload against the glossary",
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:     "file",
            Aliases:  []string{"f"},
            Usage:    "path to JSON or TBX payload; '-' for stdin",
            Required: true,
            TakesFile: true,
        },
        &cli.BoolFlag  {Name: "prune",       Usage: "remove concepts absent from payload"},
        &cli.BoolFlag  {Name: "dry-run",     Aliases: []string{"n"}, Usage: "preview without modifying file"},
        &cli.BoolFlag  {Name: "transaction", Usage: "append a <transacGrp> record"},
        &cli.StringFlag{
            Name:    "author",
            Aliases: []string{"a"},
            Usage:   "responsibility value for the transaction record",
            Sources: cli.EnvVars("TERMINOLOGY_AUTHOR"),
        },
    },
    Action: underConstruction,
}
```

## TDD plan

1. **RED** `apply/basic` golden: argv `[\"terminology\", \"apply\", \"--file\", \"-\"]`; exit 75, stub envelope.
2. **RED** `apply/missing-file`: argv `[\"terminology\", \"apply\"]`; non-zero exit (urfave rejects Required: true).
3. **RED** `apply/with-prune-dryrun-transaction-author`: argv `[..., \"--prune\", \"-n\", \"--transaction\", \"-a\", \"Andre\"]`; stub fires (proves all flags parse).
4. **RED** `apply/author-via-env`: argv `[..., \"--file\", \"-\"]` with env `TERMINOLOGY_AUTHOR=Andre`; stub fires (proves Sources wiring).

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes

## Out of scope

- Patch model + payload validation (E8).
- TBX I/O (E2).


## Notes

**2026-05-22T21:11:16Z**

Implemented apply command stub with all flags: --file/-f (Required, TakesFile), --prune, --dry-run/-n, --transaction, --author/-a (with TERMINOLOGY_AUTHOR env source). Registered in root.go. Golden test + unit tests for missing-file, all-flags-together, and author-via-env. make build clean.
