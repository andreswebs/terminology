---
id: ter-6dsf
status: closed
deps: []
links: []
created: 2026-05-24T01:24:05Z
type: chore
priority: 1
assignee: Andre Silva
tags: [qa, entry-gate, manual, e8]
---
# Entry Gate — E8: terminology apply

## Purpose

Pre-work approval gate for the E8 epic (ter-jfqg). All E8
task tickets depend on this gate. Closing it unlocks the tasks for work.

This ticket is blocked by the sentinel (`ter-go3r`). A human must remove
that dependency and close this ticket before any agent can start working
on E8 tasks.

**DO NOT close this ticket programmatically.**

## Checklist

Before unlocking this epic for work, verify:

- [ ] All prerequisite epics are closed (check `tk dep tree ter-jfqg`)
- [ ] The spec is reviewed and understood: [docs/specs/008-apply.md](docs/specs/008-apply.md)
- [ ] Task tickets are complete and well-defined
- [ ] Ready to commit resources to this epic

## How to use

```bash
tk undep ter-6dsf ter-go3r   # remove sentinel dep
tk close ter-6dsf            # tasks become ready
```

## Related

- Exit gate: ter-pbt2
- Epic: ter-jfqg
- Sentinel: ter-go3r
