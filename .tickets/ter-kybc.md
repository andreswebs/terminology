---
id: ter-kybc
status: closed
deps: [ter-st7u]
links: []
created: 2026-05-26T19:30:03Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, output, envelope]
---
# E7.T3 — Write envelope types (output/types.go)

Define output envelope types for all 5 write commands in internal/output/types.go. All write commands emit the affected concept in full lookup-style shape (concept_id, subject_field, languages with all terms). Register envelopes and exit codes. The write envelope wraps a ConceptResult (same shape as LookupResult but with full term details including status, POS, gender, etc.) since read/write round-trip uses the same struct.

## Acceptance Criteria

- make build passes
- WriteEnvelope type registered for concept add, concept update, concept remove, term add, term deprecate
- Envelope carries schema_version, ok, result (full concept shape)
- concept remove envelope includes the removed concept's last state
- MarshalJSON ensures nil slices serialize as []
- Existing envelope registrations unchanged


## Notes

**2026-05-26T19:51:59Z**

Implemented WriteEnvelope, WriteResult, WriteTermGroup, WriteTerm, and WriteCrossRef types in internal/output/types.go. WriteResult carries the full concept shape (all TBX data categories at concept and term levels) for read→modify→write round-trip. WriteEnvelope registered for all 5 write commands (concept add/update/remove, term add/deprecate). MarshalJSON ensures nil Languages map serializes as {}. WriteTerm includes all data categories from tbx.Term: administrative_status, part_of_speech, grammatical_gender, grammatical_number, register, term_type, term_location, geographical_usage, transfer_comment, reading, reading_note, contexts, sources, customer_subset, project_subset, external_refs, cross_refs, notes. Schema golden files regenerated with -update flag. Exit codes for write commands were already registered.
