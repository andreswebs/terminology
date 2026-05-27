---
id: ter-g656
status: closed
deps: [ter-jfqg, ter-nd3x, ter-de52]
links: []
created: 2026-05-22T19:19:20Z
type: epic
priority: 4
assignee: Andre Silva
tags: [epic, release]
---
# E10 — Release: versioning, signing, distribution

Spec: docs/specs/010-release.md

Takes the project from 'green tests on main' to 'users can install and trust the binary'.

Versioning (two decoupled numbers):
- Binary semver — embedded via runtime/debug.BuildInfo and -ldflags -X main.version
- Schema version — integer constant in internal/output/version.go, bumped on any breaking envelope change

Precedence: ldflags > BuildInfo > 'dev'.

Pre-releases: v0.x.y-rc.N only (no -alpha/-beta).

v1.0.0 gate: after E8 apply ships AND one subsequent minor goes by without envelope or surface break.

Build & release:
- make release VERSION=v0.x.y — tags, regenerates dist, gh release create, ldflags-injects version
- CI gate: make validate && make build-all on every PR
- Nightly fuzz schedule (per E9)

Reproducible builds: go build -trimpath -ldflags '-s -w -buildid= -X main.version=${VERSION}'. User can rebuild a tagged version and verify against SHA256SUMS.txt.

Distribution (v0/v1): GitHub Releases (raw + tarballs) + SHA256SUMS.txt + Homebrew tap (manual at v0, automated at v1) + go install. Deferred to v2: cosign/GPG signing (Sigstore keyless via CI OIDC), asdf, scoop/winget, Docker.

Changelog: no formal policy in v0; git history + release tags are the record.


## Notes

**2026-05-27T19:39:57Z**

## E10 Task tickets created (2026-05-27)

| ID        | Task                                                              | Deps            |
| --------- | ----------------------------------------------------------------- | --------------- |
| ter-rho9  | T1 — Reproducible-build flags in Makefile LDFLAGS                 | entry gate      |
| ter-snic  | T2 — make release target (tag + dist + gh release create)         | T1              |
| ter-m8hc  | T3 — Scenarios-parse test                                         | entry gate      |
| ter-9ixx  | T4 — Error reference generator (go:generate + docs/reference)     | entry gate      |
| ter-kohg  | T5 — Error reference freshness check (make target)                | T4              |
| ter-z4wv  | T6 — Help-text health lint (non-empty Usage/Description)          | entry gate      |
| ter-86sa  | T7 — Version embedding tests                                     | entry gate      |
| ter-tg3m  | T8 — Golden CLI tests for release + --version                     | T1, T7          |

Out of scope: Homebrew tap (separate repo), GitHub Actions workflows (local make targets only for now).
