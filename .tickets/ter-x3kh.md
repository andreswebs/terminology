---
id: ter-x3kh
status: closed
deps: [ter-bmce, ter-w01i]
links: []
created: 2026-05-26T19:30:43Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, write, concept-add]
---
# E7.T7 — concept add command action

Implement concept add command action replacing the underConstruction stub. Input: flags (--id, --lang, --term, --status, etc.), JSON stdin, or TBX fragment. Behavior: (1) parse input from flags or stdin, (2) derive ID if --id omitted (via T4 DeriveID + canonical-lang resolution), (3) check for duplicate_id, (4) build Concept, (5) append to glossary, (6) optionally attach transacGrp (T5), (7) run write pipeline (T6). Output: full concept in write envelope (T3).

## Acceptance Criteria

- make build passes
- concept add creates a concept from flags (happy path)
- concept add creates a concept from JSON stdin
- concept add creates a concept from TBX fragment stdin
- ID auto-derivation works when --id omitted
- duplicate_id error when ID exists
- --dry-run previews without writing
- --transaction appends transacGrp to conceptEntry
- Pre-write validation catches invalid data
- Existing tests unaffected


## Notes

**2026-05-26T22:59:11Z**

Implemented concept add command action replacing the underConstruction stub. Key changes: (1) concept_add.go: full action with 3 input modes - flags, JSON stdin, TBX fragment stdin. Flag mode requires --lang + --term; stdin mode auto-detects JSON vs XML. ID auto-derivation from canonical-lang preferred term when --id omitted. Duplicate ID check in mutator. --dry-run and --transaction supported. (2) write_helpers.go: shared buildWriteResult (tbx.Concept → output.WriteResult) and tbxTermToWriteTerm helpers for all write commands. (3) tbx/model.go: added ParseStatus(string) Status and Status.String() methods for string↔enum conversion. (4) Updated stub_test.go to use concept update (still a stub). (5) Updated golden files for concept_add (now tests no-tbx-path error), schema full, and schema command_filter. (6) Added tests: happy path with flags, duplicate_id, dry-run, explicit ID, missing lang/term, transaction with author, persisted-to-file verification. (7) Added ParseStatus and StatusString unit tests in tbx/model_test.go.
