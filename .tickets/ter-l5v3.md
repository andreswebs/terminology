---
id: ter-l5v3
status: closed
deps: []
links: []
created: 2026-05-27T13:46:20Z
type: bug
priority: 2
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, bug]
---
# BUG: Nonexistent payload file returns exit 1 internal_error instead of exit 3

## Summary

When `--file` points to a nonexistent path, `apply` returns exit 1 with
`internal_error` instead of exit 3 (I/O error) per the exit code table
in `docs/cli-design.md`.

## Reproduce

```sh
$TT apply --tbx "${W}" --file /nonexistent/path.json
# Returns: exit 1, {"error":{"code":"internal_error","message":"reading payload file: open /nonexistent/path.json: no such file or directory"}}
# Expected: exit 3, I/O error
```

## Root cause

The file-not-found error from `os.Open` in the payload loading path is
not classified as an I/O error. It propagates as a generic internal error.

## Fix

Detect `os.IsNotExist` (or `errors.Is(err, fs.ErrNotExist)`) in the
payload loading path and return the appropriate I/O error sentinel
(exit 3).

## Affected QA tests

- TC-ERR-004

## Refs

- QA report: `qa/E8-manual-qa.report.md` (BUG-003)
- Exit code table: `docs/cli-design.md`


## Notes

**2026-05-27T13:59:59Z**

Fixed: readPayloadFile wraps os.ReadFile errors with fmt.Errorf, preserving the fs.ErrNotExist cause. In applyAction, errors.Is(err, fs.ErrNotExist) now detects this and returns terr.Newf("io_error", 3, ...) for exit 3. Added TestApply_NonexistentFile_ExitCode3 unit test and TestApply_FileNotFound_Golden golden test. Consistent with scan/check/extract I/O error handling pattern.
