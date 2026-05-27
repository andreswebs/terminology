---
id: ter-5ijf
status: closed
deps: [ter-0l8i]
links: []
created: 2026-05-22T19:48:25Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, command, schema]
---
# E1.T11 — Command stub: schema

## Goal

Wire `terminology schema` as a stub. Real reflective body lands in E4.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Surface inventory → schema", §"Short-alias map" (rows: --command)
- cli-design.md: [§\`terminology schema\`](docs/cli-design.md) — note: now describes the reflective implementation; E1 still stubs the action

## Depends on

- T5

## Files to create / edit

- `src/internal/app/commands/schema.go`
- `src/internal/app/testdata/schema/basic.{argv,stdout,stderr,exit}`
- `src/internal/app/commands_test.go` — append schema rows
- Edit `internal/app/root.go` Commands slice

## Surface

```go
&cli.Command{
    Name:  "schema",
    Usage: "emit a reflective description of the CLI surface, envelopes, and error codes",
    Flags: []cli.Flag{
        &cli.StringFlag{Name: "command", Usage: "restrict output to one command's entry"},
    },
    Action: underConstruction,
}
```

## TDD plan

1. **RED** `schema/basic` golden: argv `[\"terminology\", \"schema\"]`; exit 75, stub envelope.
2. **RED** `schema/with-command`: argv `[..., \"--command\", \"validate\"]`; stub fires (proves the flag parses).

## Acceptance

- `make build` clean
- `cd src && go test ./internal/app/...` passes

## Out of scope

- Reflective walker (E4).
- The `schema --command UNKNOWN` 65-exit-code path (E4).


## Notes

**2026-05-22T21:14:22Z**

Implemented schema command stub. Created src/internal/app/commands/schema.go with --command flag (no short alias per spec). Wired into root.go Commands slice. Added two tests: TestSchema_Stub_Golden (golden file comparison) and TestSchema_WithCommand_Stub (proves --command flag parses, still exits 75). Golden files generated via -update flag. make build clean.
