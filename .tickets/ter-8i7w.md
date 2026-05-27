---
id: ter-8i7w
status: closed
deps: [ter-st7u]
links: []
created: 2026-05-26T19:29:55Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, errors, sentinels]
---
# E7.T2 — Write error sentinels

Add error sentinels for write commands in internal/write/errors.go per E7 spec §Error envelope codes. Sentinels: duplicate_id (exit 65), not_found (exit 65), dangling_crossref (exit 65), invalid_id (exit 65), invalid_input (exit 65). invalid_picklist already handled by pickFlag urfave validator (exit 2) — confirm and document. All sentinels registered via terr.New() per docs/adr/error-handling.md.

## Acceptance Criteria

- make build passes
- All 5 sentinels defined in internal/write/errors.go
- Each sentinel has correct code, exit code, hint, and message
- Sentinels auto-register in terr.All() for schema introspection
- Tests verify code, exit code, and hint for each sentinel


## Notes

**2026-05-26T19:45:18Z**

Added 4 new error sentinels to internal/write/errors.go: ErrDuplicateID, ErrNotFound, ErrDanglingCrossref, ErrInvalidInput. All use exit code 65 per spec. ErrInvalidID was already present from E7.T4. invalid_picklist is confirmed handled by urfave pickFlag validator (exit 2), not a write sentinel. Tests in errors_test.go verify code, exit code, hint, error message, and registry presence for all 5 sentinels. Schema golden file unchanged because internal/write is not yet imported transitively by internal/app (expected per E7.T4 learnings — schema will update when command actions import the write package).
