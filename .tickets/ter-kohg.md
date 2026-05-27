---
id: ter-kohg
status: closed
deps: [ter-9ixx, ter-39xj]
links: []
created: 2026-05-27T19:35:53Z
type: task
priority: 3
assignee: Andre Silva
tags: [e10, ci, lint]
---
# E10.T5 — Error reference freshness check (make target)

Add a make target that re-runs the error reference generator and fails if the output differs from the committed file.

## Spec reference

docs/specs/010-release.md §Docs / CI sync:

    CI re-runs the generator and fails on diff.

## Design notes

- Depends on the generator from E10.T4
- The target should: run go generate, then git diff --exit-code docs/reference/errors.md
- Wire into make validate (or a new make check-generated) so it runs as part of the quality gate
- This ensures no one adds a new sentinel without regenerating the reference doc

## Acceptance criteria

- A make target re-runs go generate and fails if docs/reference/errors.md has a diff
- The target is wired into the validation pipeline (make validate or make build)
- Deliberately adding a sentinel without regenerating causes the check to fail

