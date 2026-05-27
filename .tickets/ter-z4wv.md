---
id: ter-z4wv
status: closed
deps: [ter-39xj]
links: []
created: 2026-05-27T19:36:05Z
type: task
priority: 3
assignee: Andre Silva
tags: [e10, lint, cli]
---
# E10.T6 — Help-text health lint (non-empty Usage/Description)

Add a test or lint check that verifies every CLI command has a non-empty Description and every flag has a non-empty Usage string.

## Spec reference

docs/specs/010-release.md §Docs / CI sync:

    Help-text health. Every flag declares a non-empty Usage;
    every command declares a non-empty Description. Runs in make lint.

## Design notes

- This should walk the urfave/cli app structure and assert on each command/flag
- Could be a Go test or a custom lint check
- The spec says it should run in make lint, but a Go test in make test is acceptable
- Walk app.Commands recursively (subcommands), check Description != ""
- Walk all Flags in each command, check Usage != ""

## Acceptance criteria

- A test/lint verifies every command has a non-empty Description
- A test/lint verifies every flag has a non-empty Usage
- Runs as part of make test or make lint
- Deliberately emptying a command's Description causes the check to fail

