---
id: ter-6z5g
status: closed
deps: []
links: []
created: 2026-05-24T01:24:05Z
type: chore
priority: 1
assignee: Andre Silva
tags: [qa, entry-gate, manual, e2]
---
# Entry Gate — E2: Domain model & TBX I/O

## Purpose

Pre-work approval gate for the E2 epic (ter-uqyn). All E2
task tickets depend on this gate. Closing it unlocks the tasks for work.

This ticket is blocked by the sentinel (`ter-go3r`). A human must remove
that dependency and close this ticket before any agent can start working
on E2 tasks.

**DO NOT close this ticket programmatically.**

## Checklist

Before unlocking this epic for work, verify:

- [ ] All prerequisite epics are closed (check `tk dep tree ter-uqyn`)
- [ ] The spec is reviewed and understood: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md)
- [ ] Task tickets are complete and well-defined
- [ ] Ready to commit resources to this epic

## How to use

```bash
tk undep ter-6z5g ter-go3r   # remove sentinel dep
tk close ter-6z5g            # tasks become ready
```

## Related

- Exit gate: ter-v203
- Epic: ter-uqyn
- Sentinel: ter-go3r
