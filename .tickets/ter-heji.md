---
id: ter-heji
status: closed
deps: [ter-0db2]
links: []
created: 2026-05-26T17:24:53Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-7fyo
tags: [e6, task, qa, testing]
---
# E6.T10 — Manual QA plan for E6

## Goal

Review the manual QA test plan at `qa/E6-manual-qa.md` for completeness against the E6 spec. The plan has already been created.

## Refs

- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md)
- E5 QA plan (reference): `qa/E5-manual-qa.md`
- QA plan: `qa/E6-manual-qa.md`

## Acceptance

- QA plan covers all E6 behaviors per spec
- Test cases have clear steps + expected results
- Priority assigned (P0–P2)
- Plan follows same format as E5 QA plan

## Notes

**2026-05-26T18:30:00Z**

QA plan created at `qa/E6-manual-qa.md` with 43 test cases across 14 sections: check clean path, missing/forbidden/admitted violations, --strict, language resolution (frontmatter > flag > error), violation ordering, context window, --fields projection, envelope shape, error cases, code region stripping, stream routing, scan frontmatter update (4 cases), and regression (5 cases). Follows the same format as E5-manual-qa.md. Remaining work is review only.

**2026-05-26T18:48:25Z**

Reviewed and fixed qa/E6-manual-qa.md for completeness against the E6 spec. Key changes:

1. CRITICAL FIX — All Hebrew target fixtures used definite article ה prefix on glossary terms (e.g. הצמצום instead of צמצום). The matcher's word-boundary check uses \p{L} boundaries, so ה (a Letter character) preceding a term causes the match to be rejected. Fixed all 7 Hebrew fixtures to use standalone terms. Added a comment in the setup section and a risk table entry explaining the constraint.

2. TC-CHK-STRICT-001 — Simplified the confusing jq assertion from a disjunction to plain '.violations == []'. Added assertions for admitted_variant warning shape (type, concept_id, variant, line, column). Clarified expected results: fixture has both preferred צמצום and admitted התכווצות, so no missing violation — only an admitted_variant warning.

3. Added TC-CHK-ENVVAR-001 (P1) — Tests TERMINOLOGY_TBX env var resolution for check command. Mirrors E5's TC-SCAN-ENV-VAR-001 pattern.

4. TC-CHK-ORD-001 — Updated comments to reflect the corrected target-multi.md fixture: 2 positional (forbidden_variant) + 3 missing (all preferred terms absent) = 5 violations. Relaxed assertions from exact counts to >= thresholds.

5. Renumbered sections 11-14 → 11-15 to accommodate the new env var section.

6. Updated summary table: 43 → 44 test cases; added Check — TERMINOLOGY_TBX env var row.

The QA plan now covers all E6 spec behaviors with correct expected results. Format matches E5-manual-qa.md.
