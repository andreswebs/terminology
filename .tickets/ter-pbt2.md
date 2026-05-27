---
id: ter-pbt2
status: closed
deps: [ter-6dsf, ter-xby1, ter-pwd0, ter-slnb, ter-u2b3, ter-nur0, ter-dgow, ter-nhmw, ter-kqsv, ter-xjpk, ter-827j, ter-l5v3]
links: []
created: 2026-05-25T12:49:28Z
type: chore
priority: 1
assignee: Andre Silva
tags: [qa, exit-gate, manual, e8]
---
# Exit Gate — E8: terminology apply


## Purpose

Post-completion QA gate for the E8 epic (ter-jfqg). This ticket
blocks the epic from being closed until a human has reviewed and approved
the completed work.

This ticket depends on all E8 task tickets. It becomes "ready"
only after every task is closed.

**DO NOT close this ticket programmatically.** Only close it after manual
review and sign-off.

## Checklist

Before closing this gate, verify:

- [ ] All E8 task tickets are closed
- [ ] `make build` passes with all E8 changes
- [ ] Implementation aligns with the spec: [docs/specs/008-apply.md](docs/specs/008-apply.md)
- [ ] No regressions in existing tests
- [ ] Code reviewed for quality, security, and spec conformance
- [ ] Deviation notes in task tickets have been addressed or accepted

## How to use

1. All E8 tasks must be closed first (this ticket auto-unblocks)
2. Review the implementation against the spec
3. Run `make build` and verify everything passes
4. Close this ticket: `tk close ter-pbt2`
5. Close the epic: `tk close ter-jfqg`

## Related

- Entry gate: ter-6dsf
- Epic: ter-jfqg
