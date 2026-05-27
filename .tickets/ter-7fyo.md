---
id: ter-7fyo
status: closed
deps: [ter-2sqs, ter-b3ug]
links: []
created: 2026-05-22T19:19:20Z
type: epic
priority: 3
assignee: Andre Silva
tags: [epic, scan, read]
---
# E6 — terminology scan & check

Spec: docs/specs/006-scan-check.md

The two matcher-driven commands on top of E5.

- scan FILE — informational; exit 0 regardless. Matches sorted by (line, column). --context default 80 (40 each side, line-clamped). --lang restricts.
- check SRC TGT — at-least-one counting. Source occurrence count doesn't enter the comparison. Violations: missing (source≥1, target 0), forbidden_variant (target uses deprecated/superseded), admitted_variant (--strict only). Code regions stripped symmetrically.

Language resolution: frontmatter → --lang/--source-lang/--target-lang → fail with ErrLanguageRequired (no default — silent misclassification is the failure mode this prevents).

Violation ordering by (line, column) in TGT; missing tail group by concept_id ASCII. Exit 1 on any violation. One new dep allowed: `gopkg.in/yaml.v3` for frontmatter parsing (replaces hand-rolled YAML subset parser).

## Tasks

| ID       | Task                                          | Deps                    |
| -------- | --------------------------------------------- | ----------------------- |
| ter-s7xa | T1 — Frontmatter language extraction (shared) | entry gate              |
| ter-qxpp | T2 — ErrLanguageRequired sentinel             | entry gate              |
| ter-ir1k | T3 — Check envelope + violation types         | entry gate              |
| ter-y57h | T4 — Check algorithm: missing + forbidden     | T1, T2, T3              |
| ter-vv38 | T5 — Check --strict + admitted_variant        | T4                      |
| ter-vg07 | T6 — Violation ordering                       | T4                      |
| ter-hole | T7 — Check command action                     | T4, T5, T6              |
| ter-wx2k | T8 — Scan frontmatter language resolution     | T1, T2                  |
| ter-0db2 | T9 — Golden CLI tests for check               | T7, T8                  |
| ter-heji | T10 — Manual QA plan for E6                   | T9                      |

## Gates

- Entry gate: ter-ppn9
- Exit gate: ter-b3ug
