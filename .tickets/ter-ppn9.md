---
id: ter-ppn9
status: closed
deps: []
links: []
created: 2026-05-24T01:24:05Z
type: chore
priority: 1
assignee: Andre Silva
tags: [qa, entry-gate, manual, e6]
---
# Entry Gate — E6: scan & check

## Purpose

Pre-work approval gate for the E6 epic (ter-7fyo). All E6
task tickets depend on this gate. Closing it unlocks the tasks for work.

This ticket is blocked by the sentinel (`ter-go3r`). A human must remove
that dependency and close this ticket before any agent can start working
on E6 tasks.

**DO NOT close this ticket programmatically.**

## Checklist

Before unlocking this epic for work, verify:

- [ ] All prerequisite epics are closed (check `tk dep tree ter-7fyo`)
- [ ] The spec is reviewed and understood: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md)
- [ ] Task tickets are complete and well-defined
- [ ] Ready to commit resources to this epic

## How to use

```bash
tk undep ter-ppn9 ter-go3r   # remove sentinel dep
tk close ter-ppn9            # tasks become ready
```

## Related

- Exit gate: ter-b3ug
- Epic: ter-7fyo
- Sentinel: ter-go3r
