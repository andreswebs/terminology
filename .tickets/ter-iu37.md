---
id: ter-iu37
status: closed
deps: [ter-w01i]
links: []
created: 2026-05-26T19:31:10Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, write, term-add]
---
# E7.T10 — term add command action

Implement term add command action. Appends a termSec to an existing concept's langSec (creating the langSec if it doesn't exist). Requires positional ID arg + --lang + --term. Optional: --status, --part-of-speech, --register, --grammatical-gender. ID stability: adding a new preferred term does NOT re-derive the concept ID. Transaction record placed at termSec level. not_found error if concept ID missing.

## Acceptance Criteria

- make build passes
- term add appends termSec to existing langSec
- term add creates langSec if needed
- Picklist values validated (via existing pickFlag)
- not_found error when concept ID missing
- ID never changes even when adding preferred term
- --transaction attaches transacGrp to termSec
- --dry-run previews
- Output shows full concept in write envelope
- Existing tests unaffected


## Notes

**2026-05-26T23:18:04Z**

Implemented termAddAction in term_add.go. The action follows the same mutator pattern as concept add/update: finds existing concept by ID (ErrNotFound if missing), finds or creates the LangSection for --lang, appends a new Term with all picklist flags, attaches transaction at termSec level. Updated stub_test.go to use term deprecate (still a stub). Replaced 7 stub tests with 12 real tests covering: happy path (existing langSec), langSec creation, not_found error, dry-run preview, transaction attachment, all flags, missing required flags, invalid picklist, no TBX path, and ID stability (adding preferred term does not change concept ID). Removed stale golden files. make build passes clean.
