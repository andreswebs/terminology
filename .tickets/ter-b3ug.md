---
id: ter-b3ug
status: closed
deps: [ter-ppn9, ter-s7xa, ter-qxpp, ter-ir1k, ter-y57h, ter-vv38, ter-vg07, ter-hole, ter-wx2k, ter-0db2, ter-heji]
links: []
created: 2026-05-25T12:49:28Z
type: chore
priority: 1
assignee: Andre Silva
tags: [qa, exit-gate, manual, e6]
---
# Exit Gate — E6: scan & check


## Purpose

Post-completion QA gate for the E6 epic (ter-7fyo). This ticket
blocks the epic from being closed until a human has reviewed and approved
the completed work.

This ticket depends on all E6 task tickets. It becomes "ready"
only after every task is closed.

**DO NOT close this ticket programmatically.** Only close it after manual
review and sign-off.

## Checklist

Before closing this gate, verify:

- [ ] All E6 task tickets are closed
- [ ] `make build` passes with all E6 changes
- [ ] Implementation aligns with the spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md)
- [ ] No regressions in existing tests
- [ ] Code reviewed for quality, security, and spec conformance
- [ ] Deviation notes in task tickets have been addressed or accepted

## How to use

1. All E6 tasks must be closed first (this ticket auto-unblocks)
2. Review the implementation against the spec
3. Run `make build` and verify everything passes
4. Close this ticket: `tk close ter-b3ug`
5. Close the epic: `tk close ter-7fyo`

## Related

- Entry gate: ter-ppn9
- Epic: ter-7fyo
