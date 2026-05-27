---
id: ter-smzv
status: closed
deps: [ter-w01i]
links: []
created: 2026-05-26T19:30:53Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, write, concept-update]
---
# E7.T8 — concept update command action

Implement concept update command action. Requires positional ID arg + exactly one of --merge/--replace. --merge semantics per spec: Languages map overlays (payload keys overlay, absent preserved), Terms matched by (Surface, AdministrativeStatus) natural key (equal pair merges fields, else appends), Definitions/CrossRefs/Sources/Notes replace-if-present. --replace: wholesale replacement (except ID). ID stability: renaming preferred term does NOT re-derive ID. Input: flags, JSON stdin, or TBX fragment.

## Acceptance Criteria

- make build passes
- concept update --merge merges languages, terms by natural key, replaces definitions/notes
- concept update --replace replaces entire concept content except ID
- not_found error when ID missing
- ID never changes on update (stability contract)
- --dry-run previews merge/replace result
- --transaction appends transacGrp at correct scope
- JSON stdin and TBX fragment input work
- Existing tests unaffected


## Notes

**2026-05-26T23:06:29Z**

Implemented concept update command action with --merge and --replace modes. Key details: (1) --replace replaces entire concept content except ID (ID stability). (2) --merge overlays: Languages map merges (absent keys preserved), Terms matched by (Surface, AdministrativeStatus) natural key (match merges fields, no match appends), Definitions/CrossRefs/Sources/Notes replace-if-present. (3) Added picklist flags (--status, --part-of-speech, --register, --grammatical-gender) matching concept add. (4) JSON stdin and TBX fragment input work via shared parseConceptFromStdin. (5) --dry-run, --transaction, --author all work. (6) Updated stub_test.go to use concept remove. (7) Removed stale stub golden files for merge/replace. (8) Regenerated schema golden files due to new flags. (9) 17 passing tests covering replace, merge, natural key matching, term append, not_found, ID stability, dry-run, transaction, JSON stdin, author via env.
