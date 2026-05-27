---
id: ter-48g9
status: closed
deps: [ter-0l8i]
links: []
created: 2026-05-22T19:51:57Z
type: task
priority: 3
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, command, concept, write]
---
# E1.T12 — Command stub: concept parent + concept add

## Goal

Wire the `concept` parent (help-only, no Action — invoking it bare prints subcommand help and exits 2) and its first subcommand `concept add`. Bundles the parent here so subsequent tickets (T13 concept update, T14 concept remove) only add subcommands.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → concept parent + concept.add", §"Short-alias map" (rows: --id/-i, --subject-field, --canonical-lang, --lang/-l, --term/-t, --status/-s, --part-of-speech/-p, --register/-r, --grammatical-gender, --dry-run/-n, --transaction, --author/-a)
- cli-design.md: [§\`terminology concept add\`](docs/cli-design.md)

## Depends on

- T5

## Files to create / edit

- `src/internal/app/commands/concept.go` — parent (help-only, no Action)
- `src/internal/app/commands/concept_add.go` — `ConceptAdd() *cli.Command` (sub-command struct returned for assembly in concept.go)
- `src/internal/app/testdata/concept_add/basic.{argv,stdout,stderr,exit}`
- `src/internal/app/testdata/concept/help.{argv,stdout,stderr,exit}`
- `src/internal/app/commands_test.go` — append concept rows
- Edit `internal/app/root.go` Commands slice — register `commands.Concept()`

## Surface

### Parent

```go
func Concept() *cli.Command {
    return &cli.Command{
        Name:  "concept",
        Usage: "create, update, or remove a concept entry",
        // No Action: urfave will print subcommand help and exit non-zero
        // when no subcommand is matched. Confirm urfave v3's exact exit
        // code in this case; if it isn't 2, override via a tiny Action
        // that calls cli.ShowSubcommandHelp + returns cli.Exit(\"\", 2).
        Commands: []*cli.Command{
            ConceptAdd(),
            // T13 appends ConceptUpdate(); T14 appends ConceptRemove().
        },
    }
}
```

### `concept add`

```go
func ConceptAdd() *cli.Command {
    return &cli.Command{
        Name:  "add",
        Usage: "create a new concept entry",
        Flags: []cli.Flag{
            &cli.StringFlag{Name: "id",                Aliases: []string{"i"}, Usage: "explicit concept id (otherwise derived from canonical preferred term)"},
            &cli.StringFlag{Name: "subject-field",                              Usage: "concept-level subjectField"},
            &cli.StringFlag{Name: "canonical-lang",                             Usage: "language used for id derivation when --id is omitted"},
            &cli.StringFlag{Name: "lang",              Aliases: []string{"l"}, Usage: "language tag for the term being added"},
            &cli.StringFlag{Name: "term",              Aliases: []string{"t"}, Usage: "surface form of the term"},
            &cli.StringFlag{Name: "status",            Aliases: []string{"s"}, Usage: "administrative status", Validator: enumValidator(tbx.AdminStatus())},
            &cli.StringFlag{Name: "part-of-speech",    Aliases: []string{"p"}, Usage: "part of speech",          Validator: enumValidator(tbx.PartOfSpeech())},
            &cli.StringFlag{Name: "register",          Aliases: []string{"r"}, Usage: "register",               Validator: enumValidator(tbx.Register())},
            &cli.StringFlag{Name: "grammatical-gender",                         Usage: "grammatical gender",     Validator: enumValidator(tbx.GrammaticalGender())},
            &cli.BoolFlag  {Name: "dry-run",           Aliases: []string{"n"}, Usage: "validate and preview without writing"},
            &cli.BoolFlag  {Name: "transaction",                                Usage: "append a <transacGrp> record"},
            &cli.StringFlag{Name: "author",            Aliases: []string{"a"}, Usage: "responsibility value for the transaction record", Sources: cli.EnvVars("TERMINOLOGY_AUTHOR")},
        },
        Action: underConstruction,
    }
}
```

No positionals on `concept add` (id comes from `--id` or stdin per the spec).

## TDD plan

1. **RED** `concept_add/basic` golden: argv `[\"terminology\", \"concept\", \"add\"]`; exit 75; envelope contains `concept add` (FullName).
2. **RED** `concept/help` (parent alone): argv `[\"terminology\", \"concept\"]`; assert exit 2 + help text on stderr. (urfave v3's exact behavior — confirm; override via a tiny Action if needed.)
3. **RED** `concept_add/with-flags`: argv `[..., \"--lang\", \"es\", \"--term\", \"tzimtzum\", \"--status\", \"preferredTerm-admn-sts\"]`; stub fires.
4. **RED** `concept_add/invalid-status`: argv `[..., \"--status\", \"nope\"]`; urfave rejects before stub. (Proves Q3 enum validation.)
5. **RED** `concept_add/invalid-pos`: argv `[..., \"--part-of-speech\", \"frobnicator\"]`; urfave rejects.

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes

## Out of scope

- ID derivation (E7).
- JSON / TBX-fragment stdin payload (E7).
- Validation pipeline (E3 + E7).


## Notes

**2026-05-22T21:19:05Z**

Implemented concept parent and concept add stub. concept.go has a help-only parent with terr.New sentinel (not urfcli.Exit) for exit code 2, matching the rootAction pattern from T5. concept_add.go wires all flags per spec (--id/-i, --subject-field, --canonical-lang, --lang/-l, --term/-t, --status/-s, --part-of-speech/-p, --register/-r, --grammatical-gender, --dry-run/-n, --transaction, --author/-a with TERMINOLOGY_AUTHOR env). Enum validators use a local conceptEnumValidator function (same per-command pattern as scriptValidator in extract.go). 6 new tests added, golden files generated. make build clean.
