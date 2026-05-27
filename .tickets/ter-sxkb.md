---
id: ter-sxkb
status: closed
deps: [ter-7ik8]
links: []
created: 2026-05-26T19:31:32Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, qa, testing]
---
# E7.T13 — Manual QA plan for E7

Review the manual QA test plan at `qa/E7-manual-qa.md` for completeness against the E7 spec. The plan has already been created.

## Refs

- E7 spec: [docs/specs/007-write-commands.md](docs/specs/007-write-commands.md)
- E6 QA plan (reference): `qa/E6-manual-qa.md`
- QA plan: `qa/E7-manual-qa.md`

## Acceptance Criteria

- QA plan covers all E7 behaviors per spec
- Test cases have clear steps + expected results
- Priority assigned (P0–P2)
- Plan follows same format as E5/E6 QA plans

## Notes

**2026-05-26T20:00:00Z**

QA plan created at `qa/E7-manual-qa.md` with 61 test cases across 20 sections covering all 5 write commands (concept add/update/remove, term add/deprecate). Sections: flag input, ID derivation, JSON stdin, TBX fragment stdin, merge semantics, replace semantics, merge/replace mutex, basic remove, dangling crossref + --force, term add, term deprecate, dry-run, transaction records, ID stability, pre-write validation, concurrency/file locking, envelope shape, error cases (7 sentinels), stream routing, and regression (6 cases against existing commands). Follows E5/E6 QA plan format with `work()` helper for mutable fixtures, base glossary with cross-references, JSON and TBX fragment payloads. Remaining work is review only.


**2026-05-26T23:35:07Z**

Reviewed QA plan against E7 spec (007-write-commands.md) and cli-design.md. Added 6 missing test cases (61→67 total): TC-UPD-MERGE-003 (term-level merge by Surface+status tuple), TC-UPD-MERGE-004 (TBX fragment stdin for update), TC-TERM-DEP-003 (nonexistent langSec not_found), TC-DRY-003/004 (dry-run on concept update and remove), TC-TXN-004 (TERMINOLOGY_AUTHOR env var). Updated risk areas table with new coverage. Updated summary table counts. Plan follows E5/E6 QA format with priorities P0–P2.
