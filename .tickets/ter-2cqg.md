---
id: ter-2cqg
status: closed
deps: [ter-w01i]
links: []
created: 2026-05-26T19:31:02Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, write, concept-remove]
---
# E7.T9 — concept remove command action

Implement concept remove command action. Default: refuse with dangling_crossref if any other concept's crossReference targets ID. --force: override refusal, leave dangling refs (validate will surface them as unresolved_crossref warnings). not_found error if ID missing. Output: the removed concept's last state in write envelope. Transaction records are NOT persisted for removals per spec (removed entity has no persistent home).

## Acceptance Criteria

- make build passes
- concept remove deletes concept by ID
- dangling_crossref error when other concepts reference this ID
- --force removes despite dangling refs
- not_found error when ID missing
- Output shows removed concept's last state
- --dry-run previews without writing
- Integration test: remove --force then validate shows unresolved_crossref warnings
- Existing tests unaffected


## Notes

**2026-05-26T23:13:25Z**

Implemented concept remove command action. Key decisions: (1) Manual pipeline (load/check/remove/save) instead of write.Execute because --force needs to bypass unresolved_crossref fatal warning. Removing a concept cannot introduce duplicate_id/missing_term/invalid_lang_tag, so post-mutation validation is unnecessary. (2) findCrossRefsTo checks both concept-level and term-level CrossRefs across all other concepts. (3) The removed concept's last state is emitted in the WriteEnvelope output. (4) Transaction records are attached to the removed concept copy (for output only — not persisted since the concept is deleted). (5) stub_test.go updated to use 'term add' as the next stub command. (6) Stale golden files for concept_remove removed since old golden test was replaced with real integration tests. (7) New crossref-dct.tbx fixture created for dangling crossref tests. (8) Integration test: remove --force then validate shows unresolved_crossref warnings. 9 tests total covering: happy path, not found, dangling crossref, force, force+validate, dry-run, missing ID, transaction, author via env.
