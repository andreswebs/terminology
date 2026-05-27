---
id: ter-qxff
status: closed
deps: []
links: []
created: 2026-05-22T19:41:34Z
type: task
priority: 1
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, foundation, version]
---
# E1.T3 — Foundation: internal/version + Makefile ldflags wiring

## Goal

Stand up `internal/version` with a deterministic precedence: ldflags-injected `Override` → `runtime/debug.BuildInfo.Main.Version` → `\"dev\"`. Wire the Makefile to inject `Override` via `-ldflags -X` at release-build time.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Version wiring (Q9)" — the `Override` sketch is authoritative
- E10 spec: [docs/specs/010-release.md](docs/specs/010-release.md) §"Version-string source precedence"

## Files to create / edit

- `src/internal/version/version.go` — package source
- `src/internal/version/version_test.go` — unit tests
- Edit `Makefile` LDFLAGS line: add `-X github.com/andreswebs/terminology/internal/version.Override=$(VERSION)`

## API

```go
package version

import "runtime/debug"

// Override is set at build time via:
//   go build -ldflags \"-X github.com/andreswebs/terminology/internal/version.Override=v0.1.2\"
// When non-empty, it wins over BuildInfo.
var Override = ""

// Current resolves the version string: Override > BuildInfo.Main.Version (excluding \"(devel)\") > \"dev\".
func Current() string
```

## TDD plan

Tests in `src/internal/version/version_test.go`.

1. **RED** `TestCurrent_Override` — set `version.Override = \"v9.9.9\"` via t.Cleanup-restored value; assert Current()==\"v9.9.9\". **GREEN** implement Override branch.
2. **RED** `TestCurrent_DevFallback` — clear Override; under `go test` BuildInfo.Main.Version is typically \"(devel)\"; assert Current()==\"dev\". **GREEN** implement fallback branch with the \"(devel)\" exclusion.
3. **RED** `TestCurrent_OverrideEmptyStringFallsThrough` — Override=\"\" explicitly; same result as #2. **GREEN** covered by branch order.

Note: the BuildInfo-with-real-version branch is hard to unit-test (requires `go install`-built binary). The integration coverage lives in E10's release tests. Document this in a comment above the BuildInfo branch.

## Makefile change

Current:
```make
LDFLAGS     := -s -w
```

After:
```make
LDFLAGS     := -s -w -X github.com/andreswebs/terminology/internal/version.Override=$(VERSION)
```

`VERSION` is already defined upstream (`VERSION ?= $(shell git describe …)`).

Verify by hand after edit: `make build && ./bin/terminology-$(go env GOOS)-$(go env GOARCH) --version` should print the git-describe value (after T5 wires Root().Version). This ticket only verifies the Makefile parses + injects; `--version` end-to-end coverage lives in T5.

## Acceptance

- `make build` clean
- `cd src && go test ./internal/version/...` passes
- `make build` produces a binary that includes the Override symbol (smoke check: `go tool nm bin/terminology-* | grep version.Override` shows the symbol; not required as an automated test, just a manual confirmation in the ticket close note)

## Out of scope

- Wiring `Root().Version` to `Current()` — T5 does that.
- `make release VERSION=v0.x.y` workflow + tag automation — E10.


## Notes

**2026-05-22T20:44:07Z**

Implemented internal/version with Override > BuildInfo > dev precedence. Three tests cover: Override wins, empty Override falls through to dev (BuildInfo returns (devel) under go test), and empty string never leaks. Makefile LDFLAGS now injects -X .../version.Override=$(VERSION). Smoke-checked: strings on the cross-compiled binary confirms the git-describe value is embedded. nm unavailable for cross-arch binaries (stripped + cross-compiled), used strings instead.
