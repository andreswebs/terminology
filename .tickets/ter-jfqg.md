---
id: ter-jfqg
status: closed
deps: [ter-8gyy, ter-pbt2]
links: []
created: 2026-05-22T19:19:20Z
type: epic
priority: 3
assignee: Andre Silva
tags: [epic, write, apply]
---
# E8 — terminology apply

Spec: docs/specs/008-apply.md

Bulk declarative write. Full-state patch model: payload describes target shape of listed concepts; apply diffs against current glossary and computes minimal {added, updated, removed, unchanged} ops. Idempotent by construction.

Update rule: wholesale replace (payload IS authoritative). For surgical merges, use concept update --merge (E7).
Equality: canonicalized XML byte-identical AFTER stripping <transacGrp>. So a payload without transactions matches a glossary with transactions as 'unchanged'; apply preserves existing transactions.
--prune: glossary concepts absent from payload → remove. If any remaining concept crossReferences a pruned concept, refuse with dangling_crossref (file untouched). No --drop-refs in v1.

Payload formats:
- JSON (default) — {concepts: [...]} list of concept records (same struct as lookup output)
- --format tbx — <conceptEntryList> wrapper
- Auto-detect from extension; stdin without --format and no hint → invalid_input

Lists in output sorted ASCII byte order by concept_id. New sentinel: ErrApplyValidationFailed (distinct code apply_validation_failed with failures[] details). All-or-nothing on disk. Holds ${TBX}.lock across read→modify→validate→write.


## Notes

**2026-05-27T00:34:52Z**

2026-05-26T23:50:00Z

Pre-implementation review. Studied spec, ADRs, and full codebase (internal/write, internal/app/commands, internal/output, internal/tbx). Key decisions and deviations from spec:

1. **Input format detection**: Spec describes --format for input format selection, but --format is the global output format flag (json|text). Resolution: use extension-based detection for file paths (.json → JSON, .tbx/.xml → TBX) and content sniffing (looksLikeXML) for stdin. No new flag. This aligns with E7's existing stdin auto-detection pattern.

2. **Exit codes**: Spec says exit 1 for apply_validation_failed (recoverable with failures[] detail). Exit 65 still applies for input-level errors (invalid_input, invalid_field). Need to add exit 1 to RegisterExitCodes for apply.

3. **Write infrastructure reuse**: apply cannot use write.Execute() directly (single-concept mutator pattern). Will implement a separate bulk execution path reusing tbx.Load, tbx.Save, and validateForWrite.

4. **Concept equality**: New utility needed — canonicalized XML byte-identical after stripping transacGrp. Will encode each concept to canonical XML, strip transactions, then compare bytes.

5. **Payload shape**: JSON payload is {"concepts": [...]} where each concept matches the WriteResult struct shape (same as lookup output). TBX payload is conceptEntryList fragment.

Task breakdown: 7 tasks (T1-T7) covering envelope types, equality, payload parsing, reconciliation, command action, golden tests, and QA plan.
