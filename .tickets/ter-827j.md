---
id: ter-827j
status: closed
deps: []
links: []
created: 2026-05-27T13:46:20Z
type: bug
priority: 1
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, bug]
---
# BUG: ErrApplyValidationFailed sentinel declared but never returned

## Summary

`ErrApplyValidationFailed` is declared in `src/internal/write/errors.go`
(code `"apply_validation_failed"`, exit 1) but no code path in
`reconcile.go` or `apply.go` ever returns it. Per-concept validation
failures during reconciliation return `validation_error` (exit 65) instead.

The spec-defined `details.failures[]` array — which should contain
per-concept error detail (`concept_id`, `code`, `message`) — is never
populated.

## Reproduce

```sh
# Payload with concept that has crossref to nonexistent target.
echo '{"concepts": [{"concept_id": "x", "cross_refs": [{"target": "nonexistent"}], ...}]}' | \
  $TT apply --tbx "${W}" --file -
# Returns: exit 65, validation_error (not exit 1, apply_validation_failed)
```

## Root cause

The sentinel is defined but never wired into the reconciliation path:

```sh
grep -rn "ErrApplyValidationFailed" src/internal/ --include="*.go" | grep -v _test.go
# Only: src/internal/write/errors.go:35:var ErrApplyValidationFailed = terr.New(
```

## Fix

Wire `ErrApplyValidationFailed` into the reconciliation path for
per-concept validation failures, with the `failures[]` detail array as
specified in `docs/specs/008-apply.md`.

## Affected QA tests

- TC-VALFAIL-001 (apply_validation_failed envelope)
- TC-ATOMIC-001 (atomicity — can't verify without this sentinel)

## Refs

- QA report: `qa/E8-manual-qa.report.md` (BUG-002)
- Spec: `docs/specs/008-apply.md`


## Notes

**2026-05-27T13:56:43Z**

Fixed: ErrApplyValidationFailed is now returned by the reconciliation path. Changes: (1) Added Detailed interface to output/errors.go for errors carrying structured details, with optional Details field on errorDetail. (2) Created ApplyValidationError type in write/write.go implementing terr.Coded + output.Detailed, carries []ApplyFailure. (3) Added validateForApply() that collects ALL per-concept fatal warnings into failures[] instead of failing on first (validateForWrite still used by single-concept write commands). (4) reconcile.go now calls validateForApply instead of validateForWrite. (5) Updated golden test for apply/validation_error to expect exit 1 with apply_validation_failed code and details.failures[] array. Error envelope now matches spec 008-apply.md exactly.
