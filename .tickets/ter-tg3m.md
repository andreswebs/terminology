---
id: ter-tg3m
status: closed
deps: [ter-rho9, ter-86sa, ter-39xj]
links: []
created: 2026-05-27T19:36:32Z
type: task
priority: 3
assignee: Andre Silva
tags: [e10, test, golden]
---
# E10.T8 — Golden CLI tests for release + --version

Add golden CLI tests that verify the release workflow end-to-end:

1. A binary built with make build-local VERSION=v0.99.0 outputs "terminology version v0.99.0" for --version
2. The version string appears in the correct format in JSON envelopes (if applicable)

## Design notes

- Builds a binary with a specific version, runs --version, asserts output
- This is the integration test for the ldflags → version.Override → CLI output path
- May share a test helper with E10.T7 or be a separate golden test file
- Should be a Go test that invokes the built binary (exec.Command)

## Acceptance criteria

- Test builds binary with ldflags-injected version
- Test verifies --version output matches "terminology version vX.Y.Z"
- Test passes in make test

