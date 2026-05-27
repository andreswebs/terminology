---
id: ter-v057
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
# E1.T13 — Command stub: concept update (with --merge|--replace mutex)

## Goal

Wire `terminology concept update ID` as a stub. `--merge` and `--replace` are mutually exclusive AND exactly-one-required (passing both, or neither, is a usage error → exit 2). This is the first command with a flag-mutex check; the pattern is reused by future epics.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → concept.update", §"Short-alias map" (rows: --merge, --replace, --subject-field, --lang/-l, --term/-t, --dry-run/-n, --transaction, --author/-a)
- cli-design.md: [§\`terminology concept update ID\`](docs/cli-design.md) (updated: --merge|--replace mutex, exactly-one-required)
- E7 spec: [docs/specs/007-write-commands.md](docs/specs/007-write-commands.md) §"concept update --merge semantics"

## Depends on

- T5
- T12 (parent must exist; this ticket appends to its Commands slice)

## Files to create / edit

- `src/internal/app/commands/concept_update.go` — `ConceptUpdate() *cli.Command`
- `src/internal/app/commands/concept.go` — append `ConceptUpdate()` to the Commands slice
- `src/internal/app/testdata/concept_update/{merge,replace,both,neither,missing-id}.{argv,stdout,stderr,exit}`
- `src/internal/app/commands_test.go` — append concept-update rows

## Surface

```go
func ConceptUpdate() *cli.Command {
    return &cli.Command{
        Name:      "update",
        Usage:     "modify an existing concept",
        ArgsUsage: "ID",
        Arguments: []cli.Argument{
            &cli.StringArg{Name: "id", Min: 1, Max: 1},
        },
        Flags: []cli.Flag{
            &cli.BoolFlag  {Name: "merge",         Usage: "merge supplied fields with existing (preserve unspecified)"},
            &cli.BoolFlag  {Name: "replace",       Usage: "replace entire concept content (except id)"},
            &cli.StringFlag{Name: "subject-field", Usage: "concept-level subjectField"},
            &cli.StringFlag{Name: "lang",          Aliases: []string{"l"}, Usage: "language tag for the term being added/updated"},
            &cli.StringFlag{Name: "term",          Aliases: []string{"t"}, Usage: "surface form of the term"},
            &cli.BoolFlag  {Name: "dry-run",       Aliases: []string{"n"}, Usage: "validate and preview without writing"},
            &cli.BoolFlag  {Name: "transaction",   Usage: "append a <transacGrp> record"},
            &cli.StringFlag{Name: "author",        Aliases: []string{"a"}, Usage: "responsibility value", Sources: cli.EnvVars("TERMINOLOGY_AUTHOR")},
        },
        Before: requireMergeXorReplace,  // exactly-one-of mutex check
        Action: underConstruction,
    }
}

// requireMergeXorReplace enforces the cli-design.md rule:
// exactly one of --merge or --replace must be supplied.
func requireMergeXorReplace(ctx context.Context, cmd *cli.Command) (context.Context, error) {
    merge   := cmd.Bool("merge")
    replace := cmd.Bool("replace")
    switch {
    case merge && replace:
        return ctx, ErrMergeReplaceMutex
    case !merge && !replace:
        return ctx, ErrMergeReplaceRequired
    }
    return ctx, nil
}
```

### New sentinels (`internal/app/errors.go`)

```go
var ErrMergeReplaceMutex = terr.New(
    "merge_replace_mutex", 2,
    "pass exactly one of --merge or --replace",
    "--merge and --replace are mutually exclusive",
)

var ErrMergeReplaceRequired = terr.New(
    "merge_replace_required", 2,
    "pass exactly one of --merge or --replace",
    "either --merge or --replace must be supplied",
)
```

## TDD plan

1. **RED** `concept_update/merge` golden: argv `[\"terminology\", \"concept\", \"update\", \"tzimtzum\", \"--merge\"]`; exit 75 + stub envelope.
2. **RED** `concept_update/replace` golden: argv `[..., \"--replace\"]`; exit 75.
3. **RED** `concept_update/both` golden: argv `[..., \"--merge\", \"--replace\"]`; exit 2; envelope `merge_replace_mutex`.
4. **RED** `concept_update/neither` golden: argv `[..., \"tzimtzum\"]` (no merge/replace); exit 2; envelope `merge_replace_required`.
5. **RED** `concept_update/missing-id`: argv without ID; non-zero exit (urfave Min: 1).

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes
- Golden files for both error envelopes confirm exact byte content (proves emit-and-exit path for Before-hook errors)

## Out of scope

- Actual merge/replace semantics (E7).
- ID stability under preferred-term rename (E7).


## Notes

**2026-05-22T21:27:05Z**

Implemented concept update stub with --merge/--replace mutex. Created concept_update.go with requireMergeXorReplace Before hook and two terr sentinels (errMergeReplaceMutex, errMergeReplaceRequired). Registered in concept.go Commands slice. Added golden tests for merge, replace, both (exit 2), neither (exit 2), plus behavioral tests for missing-id, all-flags, and author-via-env. StringArg (singular) has no Min/Max — uses Name+UsageText only (same learning as T6). make build clean.
