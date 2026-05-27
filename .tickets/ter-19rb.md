---
id: ter-19rb
status: closed
deps: [ter-12nr, ter-43jt, ter-a52o, ter-bbbs, ter-8zuv, ter-mf43, ter-mule, ter-at09, ter-oeua, ter-eedk, ter-l0hf, ter-w6kr, ter-wsy4, ter-bedf, ter-9dho, ter-97c1]
links: []
created: 2026-05-25T12:49:28Z
type: chore
priority: 1
assignee: Andre Silva
tags: [qa, exit-gate, manual, e3]
---
# Exit Gate — E3: terminology validate


## Purpose

Post-completion QA gate for the E3 epic (ter-told). This ticket
blocks the epic from being closed until a human has reviewed and approved
the completed work.

This ticket depends on all E3 task tickets. It becomes "ready"
only after every task is closed.

**DO NOT close this ticket programmatically.** Only close it after manual
review and sign-off.

## Checklist

Before closing this gate, verify:

- [ ] All E3 task tickets are closed
- [ ] `make build` passes with all E3 changes
- [ ] Implementation aligns with the spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md)
- [ ] No regressions in existing tests
- [ ] Code reviewed for quality, security, and spec conformance
- [ ] Deviation notes in task tickets have been addressed or accepted

## How to use

1. All E3 tasks must be closed first (this ticket auto-unblocks)
2. Review the implementation against the spec
3. Run `make build` and verify everything passes
4. Close this ticket: `tk close ter-19rb`
5. Close the epic: `tk close ter-told`

## Related

- Entry gate: ter-bedf
- Epic: ter-told
