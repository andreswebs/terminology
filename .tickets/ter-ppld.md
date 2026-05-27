---
id: ter-ppld
status: closed
deps: [ter-w01i]
links: []
created: 2026-05-26T19:31:16Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, write, term-deprecate]
---
# E7.T11 — term deprecate command action

Implement term deprecate command action. Convenience: set existing term's administrativeStatus to deprecatedTerm-admn-sts. Requires positional ID arg + --lang + --term to identify the target termSec. not_found error if concept, langSec, or term doesn't exist. Transaction record at termSec level.

## Acceptance Criteria

- make build passes
- term deprecate sets status to deprecatedTerm-admn-sts
- not_found error when concept/lang/term not found
- --transaction attaches transacGrp to termSec
- --dry-run previews status change
- Output shows full concept in write envelope
- Existing tests unaffected


## Notes

**2026-05-26T23:22:38Z**

Implemented termDeprecateAction in term_deprecate.go. The mutator finds the concept by ID, then the langSec by lang, then the term by surface form — returning not_found (exit 65) at each level if missing. Sets AdministrativeStatus to StatusDeprecated. Transaction record attached at termSec level when --transaction is set. Dry-run previews the change without writing. Replaced 4 stub tests with 7 real tests (happy path, 3x not_found for concept/lang/term, dry-run, transaction, no-tbx-path). Updated stub_test.go to use 'apply' (last remaining stub). Removed stale term_deprecate golden files.
