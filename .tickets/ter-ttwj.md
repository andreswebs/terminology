---
id: ter-ttwj
status: closed
deps: [ter-st7u]
links: []
created: 2026-05-26T19:29:47Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, shared, clock]
---
# E7.T1 — Clock injection (internal/clock)

Create internal/clock package for injectable time. Transaction timestamps need deterministic time for golden tests. Pattern defined in docs/adr/determinism.md §5.

## Acceptance Criteria

- make build passes
- internal/clock package with Clock interface, Real clock (UTC), context-based With/From/Now
- Unit tests for With/From round-trip and Real.Now() returns UTC
- No time.Now() calls outside this package in write code


## Notes

**2026-05-26T19:42:44Z**

Implemented internal/clock package with Clock interface, realClock (UTC), context-based With/From/Now helpers. 7 unit tests covering: round-trip, fallback to Real, wrong-type fallback, UTC guarantee, recent time bounds, context clock usage, and bare-context fallback. Pattern mirrors internal/logctx. No time.Now() calls in write package.
