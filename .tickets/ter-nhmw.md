---
id: ter-nhmw
status: closed
deps: [ter-dgow]
links: []
created: 2026-05-27T00:36:49Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, task, qa, testing]
---
# E8.T7 — Manual QA plan for apply

Create a manual QA test plan for the apply command, following the format of E6 and E7 QA plans.

## Refs

- E8 spec: docs/specs/008-apply.md
- E7 QA plan (reference format): qa/E7-manual-qa.md
- E6 QA plan (reference format): qa/E6-manual-qa.md

## Scope

Manual QA plan at qa/E8-manual-qa.md covering:

1. Basic apply (JSON payload) — add, update, unchanged
2. Idempotency — same payload twice yields all unchanged
3. --prune — removes absent concepts
4. --prune + dangling crossref — refuses with error
5. TBX fragment payload
6. Format auto-detection (extension, content sniffing, stdin)
7. --dry-run — file unchanged
8. --transaction — transacGrp records on added/updated concepts
9. Wholesale replace semantics on update (payload-omitted fields dropped)
10. Concept equality (transacGrp stripping)
11. Concurrency (advisory lock)
12. Error cases: invalid_input, apply_validation_failed, no_tbx_path, I/O errors
13. Output envelope shape
14. Stream routing (stdout/stderr)
15. Regression (existing commands still work)

Follow E7 QA format: work() helper, base glossary fixture, test cases with P0-P2 priorities.

## Acceptance Criteria

- QA plan covers all E8 behaviors per spec
- Test cases have clear steps + expected results
- Priority assigned (P0-P2)
- Plan follows E7 QA plan format


## Notes

**2026-05-27T12:35:38Z**

E8-manual-qa.md was already substantially complete. Verified all 15 scope items from the ticket are covered by 47 test cases (P0-P2). Fixed minor count discrepancy in summary table (was 46, corrected to 47). Plan follows E7 format: work() helper, base glossary fixture, clear steps + expected results, priorities, entry/exit criteria, risk areas, and conventions section. All acceptance criteria met.
