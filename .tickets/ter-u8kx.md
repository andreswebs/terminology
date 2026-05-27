---
id: ter-u8kx
status: closed
deps: [ter-0l8i, ter-48g9]
links: []
created: 2026-05-22T19:51:57Z
type: task
priority: 3
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, command, concept, write]
---
# E1.T14 — Command stub: concept remove

## Goal

Wire `terminology concept remove ID` as a stub.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → concept.remove", §"Short-alias map" (rows: --force, --dry-run/-n, --transaction, --author/-a)
- cli-design.md: [§\`terminology concept remove ID\`](docs/cli-design.md)

## Depends on

- T5
- T12 (appends to concept parent's Commands slice)

## Files to create / edit

- `src/internal/app/commands/concept_remove.go`
- `src/internal/app/commands/concept.go` — append `ConceptRemove()` to the slice
- `src/internal/app/testdata/concept_remove/{basic,with-force,missing-id}.{argv,stdout,stderr,exit}`
- `src/internal/app/commands_test.go` — append rows

## Surface

```go
func ConceptRemove() *cli.Command {
    return &cli.Command{
        Name:      "remove",
        Usage:     "delete a concept entry",
        ArgsUsage: "ID",
        Arguments: []cli.Argument{
            &cli.StringArg{Name: "id", Min: 1, Max: 1},
        },
        Flags: []cli.Flag{
            &cli.BoolFlag  {Name: "force",       Usage: "remove even if other concepts cross-reference this ID"},
            &cli.BoolFlag  {Name: "dry-run",     Aliases: []string{"n"}, Usage: "validate and preview without writing"},
            &cli.BoolFlag  {Name: "transaction", Usage: "append a <transacGrp> record (not persisted for removals; emitted only in the dry-run preview)"},
            &cli.StringFlag{Name: "author",      Aliases: []string{"a"}, Usage: "responsibility value", Sources: cli.EnvVars("TERMINOLOGY_AUTHOR")},
        },
        Action: underConstruction,
    }
}
```

## TDD plan

1. **RED** `concept_remove/basic` golden: argv `[\"terminology\", \"concept\", \"remove\", \"tzimtzum\"]`; exit 75 + stub envelope.
2. **RED** `concept_remove/with-force` golden: argv `[..., \"--force\"]`; stub fires.
3. **RED** `concept_remove/missing-id`: argv without ID; non-zero exit (urfave Min: 1).

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes

## Out of scope

- Dangling crossref policy (E7).
- Real removal semantics (E7).


## Notes

**2026-05-22T21:30:18Z**

Implemented concept remove stub. Created concept_remove.go with --force, --dry-run/-n, --transaction, --author/-a flags and positional ID arg. Registered in concept.go Commands slice. Added 5 tests: golden (basic), with-force, with-all-flags, missing-id, author-via-env. All tests pass, make build clean. Pattern follows concept_update.go (positional StringArg for ID, underConstruction action).
