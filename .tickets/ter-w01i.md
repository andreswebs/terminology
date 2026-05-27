---
id: ter-w01i
status: closed
deps: [ter-8i7w, ter-kybc, ter-5b07]
links: []
created: 2026-05-26T19:30:34Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, write, pipeline]
---
# E7.T6 — Write pipeline (internal/write/write.go)

Shared read-modify-write pipeline for all write commands. Pipeline: (1) Load glossary via tbx.Load, (2) dispatch to per-command mutator, (3) run full E3 validation on in-memory result, (4) if --dry-run emit final-state preview without rename, (5) else tbx.Save (acquires lock, atomic rename). Validation failure aborts before rename — file untouched, error envelope emitted. Mutator interface: type Mutator func(*tbx.Glossary) (*tbx.Concept, error) returns the affected concept for envelope output. Also handles stdin input parsing: JSON payload (same shape as lookup output) and TBX fragment (bare conceptEntry or conceptEntryList wrapper; full tbx document rejected with invalid_input).

## Acceptance Criteria

- make build passes
- Pipeline function exported from internal/write
- Mutator interface defined
- Pre-write validation runs full E3 pipeline
- Dry-run skips Save, still runs validation
- Validation failure returns appropriate error sentinel
- JSON stdin parsing with unknown-field rejection (invalid_input)
- TBX fragment parsing: bare conceptEntry and conceptEntryList accepted, full tbx rejected (invalid_input)
- Unit tests for pipeline with mock mutator


## Notes

**2026-05-26T22:49:51Z**

Implemented the shared write pipeline in internal/write/write.go with:

- Mutator type: func(*tbx.Glossary) (*tbx.Concept, error) — per-command functions implement this
- Execute(path, mutator, dryRun) pipeline: load → mutate → validate → save (or skip save on dry-run)
- Pre-write validation via validateForWrite(): promotes semantic warnings (duplicate_id, unresolved_crossref, missing_term, invalid_lang_tag) to fatal errors that abort save
- ParseJSONInput(): parses stdin JSON matching WriteResult shape with unknown-field rejection (DisallowUnknownFields)
- ParseTBXFragment(): accepts bare <conceptEntry> and <conceptEntryList>, rejects full <tbx> document. Uses wrapInTBXShell() + existing LinguistReader.Decode() to reuse all parsing logic
- Added tbx.ReaderForDialect() export to registry.go to allow fragment parsing from write package

12 unit tests covering happy path, dry-run, validation failure, mutator error propagation, dangling crossref, JSON parsing (valid/unknown-field/malformed), TBX fragments (bare/list/full-doc-rejected/malformed).
