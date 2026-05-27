---
id: ter-smt4
status: closed
deps: []
links: [ter-qxrg, ter-rb0i, ter-vvm0]
created: 2026-05-25T14:19:20Z
type: bug
priority: 0
assignee: Andre Silva
tags: [e1, bug, blocker, contract-violation]
---
# E1.BUG — urfave-origin errors exit 1 instead of exit 2

urfave-generated errors (unknown flags, missing required flags, invalid enum
values, int parse failures) all exit code 1 with error.code "internal_error"
instead of exit code 2 with an appropriate usage-error code.

The E1 contract requires exit 2 for usage errors. The error interceptor in
main.go or the app-level error handler is not mapping urfave errors to the
correct exit code.

## Scope

~15 individual QA test case failures trace to this single root cause.

## Reproduction

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"

$TT --format yaml validate 2>/dev/null; echo "exit=$?"       # want 2, got 1
$TT apply 2>/dev/null; echo "exit=$?"                         # want 2, got 1
$TT concept add --status klingon 2>/dev/null; echo "exit=$?"  # want 2, got 1
$TT validate --bogus 2>/dev/null; echo "exit=$?"              # want 2, got 1
$TT term add tzimtzum --lang es 2>/dev/null; echo "exit=$?"   # want 2, got 1
$TT scan a.md --context bogus 2>/dev/null; echo "exit=$?"     # want 2, got 1
```

## Affected test cases

TC-ROOT-007, TC-ROOT-FMT-003, TC-VALIDATE-005, TC-SCAN-006, TC-CHECK-002,
TC-CHECK-003, TC-EXTRACT-003, TC-EXTRACT-006, TC-EXTRACT-009, TC-APPLY-004,
TC-SCHEMA-003, TC-CONCEPT-ADD-PICK-003/005/007/009, TC-TERM-ADD-004/005/006,
TC-TERM-DEP-004/005

## Fix direction

Intercept urfave errors in the app-level error handler and classify them as
usage errors (exit 2) instead of internal errors (exit 1). The classification
should inspect the error type or message to assign appropriate error codes
(e.g. unknown_flag, missing_required_flag, invalid_enum_value, invalid_type).


## Notes

**2026-05-25T14:30:16Z**

Fixed by adding classifyUsageError() in internal/output/errors.go that pattern-matches known urfave error messages and maps them to appropriate error codes (unknown_flag, invalid_value, missing_required_flag, duplicate_flag, mutually_exclusive_flags) with exit code 2. Root cause: urfave wraps validator errors with fmt.Errorf using %v (not %w), losing the ExitCoder interface, so ExitCodeFor() fell back to exit 1. Added unit tests in errors_test.go (TestExitCodeFor_UrfaveErrors, TestEmitError_Urfave*) and integration tests in commands_test.go (TestUrfaveErrors_ExitCode2, TestUrfaveErrors_ErrorCodes) covering all reproduction cases from the ticket.
