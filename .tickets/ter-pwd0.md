---
id: ter-pwd0
status: closed
deps: [ter-6dsf]
links: []
created: 2026-05-27T00:35:30Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, task, write, equality]
---
# E8.T2 — Concept equality (transacGrp-stripping canonical comparison)

Implement concept equality for the apply reconciliation algorithm. Two concepts compare equal if their canonicalized XML forms are byte-identical after stripping <transacGrp> elements.

## Spec refs

- [008-apply.md §Equality](docs/specs/008-apply.md)

## Scope

New function in internal/write/ (or internal/tbx/):

func ConceptsEqual(a, b *tbx.Concept) (bool, error)

Algorithm:
1. Encode each concept to canonical XML via the linguist writer (single-concept TBX body).
2. Strip all <transacGrp>...</transacGrp> elements from both outputs.
3. Byte-compare the results.

Rationale from spec: ignoring transacGrp means a payload without transactions matches a glossary concept with transactions as "unchanged". Apply preserves existing transactions and only appends new ones when content actually changed.

Edge cases:
- Concepts with identical content but different transaction histories → equal
- Concepts that differ only in field ordering → equal (canonical writer sorts)
- Concepts that differ in any non-transaction field → not equal

## Acceptance Criteria

- make build passes
- ConceptsEqual function with unit tests
- Stripping covers transacGrp at concept, langSec, and termSec levels
- Byte-identical canonical output for semantically equal concepts
- Tests cover: identical concepts, differing transactions only, differing content, empty concepts


## Notes

**2026-05-27T12:04:51Z**

Implemented ConceptsEqual(a, b *tbx.Concept) (bool, error) in internal/write/equality.go. Algorithm: wraps each concept in a minimal Glossary, encodes via the linguist writer to canonical DCT XML, strips all <transacGrp> blocks via regex, and byte-compares. Added WriterForDialect() to tbx/registry.go (parallel to ReaderForDialect). 8 unit tests cover: identical concepts, differing transactions at concept and term levels, differing content, differing terms, empty concepts, field ordering irrelevance, and transactions at all levels simultaneously.
