---
id: ter-htl6
status: closed
deps: []
links: []
created: 2026-05-29T01:13:40Z
type: bug
priority: 1
assignee: Andre Silva
parent: ter-8gyy
tags: [beta-feedback, write, errors]
---
# BETA — write against missing TBX returns internal_error, not io_error

Beta feedback: apply/concept add against a missing target TBX return internal_error ("opening TBX: no such file", exit 1) instead of a clean io_error. Root cause: tbx.Load() in src/internal/tbx/io.go:12-16 wraps os.Open with a plain fmt.Errorf, so output/errors.go:36 falls through to internal_error. Map fs.ErrNotExist at the write boundary to io_error (exit 3), mirroring the existing payload-file handling in src/internal/app/commands/apply.go:62-65. Decision (Andre): reuse io_error, do NOT introduce a new no_tbx_path code. Applies to all write commands that open the target: apply, concept add/update/remove, term add/deprecate, reconcile.

## Acceptance Criteria

make build passes
Missing target TBX on every write command returns io_error (exit 3), not internal_error (exit 1)
Error message remains informative ("opening TBX: ...")
Reuses io_error sentinel — no new error code added
Read commands' missing-file behaviour unchanged
Test covers missing target TBX for at least apply and concept add


## Notes

**2026-05-29T01:19:45Z**

Added loadTBXForWrite helper in internal/app/commands/write_helpers.go that maps fs.ErrNotExist to io_error (exit 3). Used in apply.go and concept_remove.go. Mirrored mapping inside write.Execute (internal/write/write.go) to cover concept add/update and term add/deprecate which load TBX via that path. Reuses existing io_error sentinel - no new code introduced. New tests in internal/app/write_missing_tbx_test.go cover all six write commands. Error message preserved (e.g. 'opening TBX: open <path>: no such file or directory'). Read commands' missing-file behaviour unchanged (validate/lookup/scan/check still return internal_error for missing TBX, matching existing golden files).
