---
id: ter-xby1
status: closed
deps: [ter-6dsf]
links: []
created: 2026-05-27T00:35:19Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, task, output, envelope]
---
# E8.T1 — Apply envelope types and error sentinel

Define the apply-specific output types in internal/output/types.go and the ErrApplyValidationFailed sentinel.

## Spec refs

- [008-apply.md §Output](docs/specs/008-apply.md)
- [error-handling.md](docs/adr/error-handling.md)

## Scope

### ApplyEnvelope (output/types.go)

New types: ApplyEnvelope (schema_version, ok, applied, warnings), ApplyResult (added, updated, removed, unchanged — all []string), ApplyFailure (concept_id, code, message — for error details.failures[]).

MarshalJSON: nil slices serialize as [], not null.

### Error sentinel (write/errors.go)

ErrApplyValidationFailed — code: "apply_validation_failed", exit: 1 (recoverable per spec), hint: "fix per-concept errors in failures[] and retry".

### Registrations

- RegisterEnvelope("apply", ApplyEnvelope{})
- Update RegisterExitCodes("apply", ...) to {0, 1, 2, 3, 65} (add exit 1)

## Acceptance Criteria

- make build passes
- ApplyEnvelope, ApplyResult, ApplyFailure types in output/types.go
- ErrApplyValidationFailed sentinel in write/errors.go with exit code 1
- Apply exit codes registered as {0, 1, 2, 3, 65}
- MarshalJSON handles nil slices as empty arrays
- Schema golden files updated if needed


## Notes

**2026-05-27T12:00:19Z**

Implemented ApplyEnvelope, ApplyResult, ApplyFailure types in output/types.go. Added ErrApplyValidationFailed sentinel in write/errors.go (code: apply_validation_failed, exit: 1). Registered apply envelope and updated exit codes to {0, 1, 2, 3, 65}. MarshalJSON coerces nil slices in Applied (added/updated/removed/unchanged) and Warnings to empty arrays. Schema golden files regenerated. All tests pass, make build clean.
