---
id: ter-qxrg
status: closed
deps: []
links: [ter-smt4, ter-rb0i, ter-v3t5, ter-vvm0]
created: 2026-05-22T19:19:19Z
type: epic
priority: 1
assignee: Andre Silva
tags: [epic, foundation, cli]
---
# E1 — CLI surface stub

Spec: docs/specs/001-cli-surface-stub.md

Stand up the urfave/cli v3 command tree with every subcommand returning exit 75 (EX_TEMPFAIL) via terr.UnderConstruction. Wire foundational packages:

- internal/terr — sentinel registry, error envelope construction, boundary emission in main.go
- internal/logctx — context.Context slog injection with 8-byte hex run_id
- internal/clock — injectable clock for determinism
- internal/output — envelope shell (schema_version, ok, error) shared by all commands
- internal/version — ldflags-injected via -X main.version

Establishes the agent-first contract: JSON on stdout by default, structured error envelopes on stderr, meaningful exit codes (0/1/2/3/65/75), global flags (--json/--text, --verbose/--debug/--quiet, --tbx). Picklists wired from internal/tbx/picklist.go (created here, populated by E3).

Acceptance: every command surface exists and returns under_construction; ldflags injection works (make build sets main.version); --json/--text switch the renderer; logctx propagates run_id into every slog record.


## Notes

**2026-05-22T21:39:16Z**

All 17 child tickets closed. Verified acceptance criteria: every command surface returns under_construction (exit 75), ldflags injection sets version, --format json/text switches renderer, logctx propagates run_id. Epic complete — E2 is now unblocked.
