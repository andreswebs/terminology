---
id: ter-vvm0
status: closed
deps: []
links: [ter-qxrg, ter-smt4]
created: 2026-05-25T14:46:51Z
type: bug
priority: 0
assignee: Andre Silva
tags: [e1, bug, blocker]
---
# E1.BUG — extract no-args urfave error not classified as usage error

The `extract` command with no positional arguments exits 1 with
`internal_error` instead of exit 2 with a usage-error code. This is the same
class of bug as ter-smt4 — urfave's argument-count error is not caught by
the error classifier.

## Reproduction

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"
$TT extract 2>&1; echo "exit=$?"
```

Actual: exit 1, `{"schema_version":1,"ok":false,"error":{"code":"internal_error","message":"sufficient count of arg files not provided, given 0 expected 1"}}`

Expected: exit 2, `error.code` = `missing_argument` (or similar usage-error code).

## Affected test case

TC-EXTRACT-003

## Root cause

The urfave error message "sufficient count of arg files not provided, given 0
expected 1" doesn't match the patterns used by the error classifier added in
the ter-smt4 fix. The classifier needs to handle this additional urfave
argument-count error format.


## Notes

**2026-05-25T14:49:19Z**

Fix: added pattern match for urfave's argument-count error ("sufficient count of arg") in classifyUsageError(), mapping it to code=missing_argument, exit=2. Added unit tests in errors_test.go (table-driven entry + EmitError test) and strengthened TestExtract_MissingFiles_Errors to assert exit code 2.
