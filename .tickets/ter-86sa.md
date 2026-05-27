---
id: ter-86sa
status: closed
deps: [ter-39xj]
links: []
created: 2026-05-27T19:36:20Z
type: task
priority: 2
assignee: Andre Silva
tags: [e10, test, version]
---
# E10.T7 — Version embedding tests

Expand test coverage for the version package to cover all three precedence branches and the --version CLI output.

## Spec reference

docs/specs/010-release.md §Version-string source precedence:

    ldflags-injected (-X main.version=v0.x.y)  >  runtime/debug.BuildInfo  >  "dev"

## Current state

internal/version/version.go has the three-branch logic. version_test.go exists but has a comment noting "integration coverage for the real-version branch lives in E10's release tests."

## Design notes

- Test 1: Override set → returns Override value
- Test 2: Override empty, BuildInfo returns valid version → returns BuildInfo version
- Test 3: Override empty, BuildInfo returns "(devel)" → returns "dev"
- Test 4 (golden CLI): terminology --version outputs "terminology version <version>"
- The BuildInfo branch is hard to test in-process (it reads from the binary itself). Consider testing via a built binary with ldflags or accept the coverage gap and document it.

## Acceptance criteria

- Override branch tested (set Override, call Current(), verify)
- "dev" fallback tested (empty Override, no BuildInfo)
- Golden CLI test for --version output format
- Tests pass in make test

