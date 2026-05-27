---
id: ter-st7u
status: closed
deps: []
links: []
created: 2026-05-24T01:24:05Z
type: chore
priority: 1
assignee: Andre Silva
tags: [qa, entry-gate, manual, e7]
---
# Entry Gate — E7: Write commands

## Purpose

Pre-work approval gate for the E7 epic (ter-8gyy). All E7
task tickets depend on this gate. Closing it unlocks the tasks for work.

This ticket is blocked by the sentinel (`ter-go3r`). A human must remove
that dependency and close this ticket before any agent can start working
on E7 tasks.

**DO NOT close this ticket programmatically.**

## Checklist

Before unlocking this epic for work, verify:

- [ ] All prerequisite epics are closed (check `tk dep tree ter-8gyy`)
- [ ] The spec is reviewed and understood: [docs/specs/007-write-commands.md](docs/specs/007-write-commands.md)
- [ ] Task tickets are complete and well-defined
- [ ] Ready to commit resources to this epic

## How to use

```bash
tk undep ter-st7u ter-go3r   # remove sentinel dep
tk close ter-st7u            # tasks become ready
```

## Related

- Exit gate: ter-lyrx
- Epic: ter-8gyy
- Sentinel: ter-go3r
