---
id: ter-c4ra
status: closed
deps: []
links: []
created: 2026-05-24T01:24:05Z
type: chore
priority: 1
assignee: Andre Silva
tags: [qa, entry-gate, manual, e5]
---
# Entry Gate — E5: Matcher

## Purpose

Pre-work approval gate for the E5 epic (ter-2sqs). All E5
task tickets depend on this gate. Closing it unlocks the tasks for work.

This ticket is blocked by the sentinel (`ter-go3r`). A human must remove
that dependency and close this ticket before any agent can start working
on E5 tasks.

**DO NOT close this ticket programmatically.**

## Checklist

Before unlocking this epic for work, verify:

- [ ] All prerequisite epics are closed (check `tk dep tree ter-2sqs`)
- [ ] The spec is reviewed and understood: [docs/specs/005-matcher.md](docs/specs/005-matcher.md)
- [ ] Task tickets are complete and well-defined
- [ ] Ready to commit resources to this epic

## How to use

```bash
tk undep ter-c4ra ter-go3r   # remove sentinel dep
tk close ter-c4ra            # tasks become ready
```

## Related

- Exit gate: ter-cfeu
- Epic: ter-2sqs
- Sentinel: ter-go3r
