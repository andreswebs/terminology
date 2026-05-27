---
id: ter-m8hc
status: closed
deps: [ter-39xj]
links: []
created: 2026-05-27T19:35:29Z
type: task
priority: 3
assignee: Andre Silva
tags: [e10, ci, docs]
---
# E10.T3 — Scenarios-parse test

Add a Go test (or make target) that parses every fenced terminology invocation in docs/examples/scenarios.md and verifies each one parses via urfave's argv parser (not full action invocation).

## Spec reference

docs/specs/010-release.md §Docs / CI sync:

    Scenarios parse. Every fenced terminology … invocation in
    docs/examples/scenarios.md parses via urfave's argv parser
    (not full action invocation). Catches doc rot when commands
    or flags rename.

## Design notes

- The test should extract fenced code blocks containing terminology invocations from scenarios.md
- Each invocation should be parsed through the CLI app's argument parser without executing the action
- This catches renamed/removed commands or flags in documentation
- Should be runnable via make test (part of the standard test suite)
- scenarios.md exists and has 592 lines

## Acceptance criteria

- A Go test extracts all terminology invocations from docs/examples/scenarios.md
- Each invocation is parsed through the urfave CLI parser
- Test fails if any invocation references a nonexistent command or flag
- Test passes with current scenarios.md content
- Runs as part of make test

