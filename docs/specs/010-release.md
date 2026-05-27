# E10 — Release: versioning, signing, distribution

> **Status**: APPROVED. Takes the project from "green tests on `main`"
> to "users can install and trust the binary".

## Scope

- Versioning policy (semver for the binary, separate `schema_version`).
- Compatibility statement.
- Build & release tooling (already partially in `Makefile`: `make dist`).
- Distribution channels.
- Checksums (signing deferred to v2).
- Reproducible-build guarantees.

## Versioning

Two version numbers, deliberately decoupled:

1. **Binary version** — semver. Embedded via `runtime/debug.BuildInfo`
   and `-ldflags` (decided in [001 Q9](001-cli-surface-stub.md)).
   Exposed via `--version`.
2. **Schema version** — a single integer (`schema_version: 1`) in
   every envelope. Lives in `internal/output/version.go` as a
   constant; reflected by `terminology schema`. Not semver — schema
   history is linear and never rolled back, so a single bumped
   integer captures everything semver would.

### Binary semver policy

- **PATCH** — bug fixes, no surface change.
- **MINOR** — new commands, new flags, new envelope fields. Existing
  callers unaffected.
- **MAJOR** — removed/renamed flags, removed/renamed envelope fields,
  exit-code changes, behavior changes that break a documented contract.

### Schema-version policy

- Adding envelope fields → no bump (callers ignore unknown keys).
- Removing or renaming a field → bump.
- Tightening a type or picklist value set → bump.

Bumped on **any** breaking change. Linear; no "schema minor".

### Version-string source precedence

```
ldflags-injected (-X main.version=v0.x.y)  >  runtime/debug.BuildInfo  >  "dev"
```

- A `make release VERSION=v0.x.y` build injects via `-ldflags`.
- A `go install ...@vX.Y.Z` recovers from `BuildInfo`.
- A bare `go build` produces `"dev"`.

The `Makefile`'s release target is the single source of the
`-ldflags` invocation.

## Leaving v0

Tag `v1.0.0` after [E8 `apply`](008-apply.md) ships **and** one
subsequent minor release goes by without an envelope or surface
break. Real-world usage validates the stability claim before locking
it; pre-`apply` is too early to commit to a stable agent surface.

Pre-release versions: `v0.x.y-rc.N` for the final 1–2 weeks before
each minor. No `-alpha`/`-beta` — narrow the semantics to one
pre-release channel.

## Build & release flow

Existing `Makefile` targets (per `CLAUDE.md`):

- `make build` — validate + compile current host.
- `make build-all` — cross-compile matrix.
- `make dist` — archives + `SHA256SUMS.txt`.

Added by this epic:

- **`make release VERSION=v0.x.y`** — tags the commit, regenerates
  dist, uploads to GitHub Releases via `gh release create`. Injects
  the version string via `-ldflags`.
- **CI gate**: `make validate && make build-all` on every PR.
- **CI gate**: nightly fuzz schedule (per [E9](009-hardening.md)).

## Reproducible builds

The release `go build` flags are frozen and documented in the
`Makefile`'s `release` target:

```
go build -trimpath -ldflags "-s -w -buildid= -X main.version=${VERSION}"
```

- `-trimpath` — strips local path prefixes.
- `-buildid=` — empties the build ID for byte-stable output.
- `-X main.version=...` — injects the released version string.

Combined with the deterministic-emit guarantees from
[determinism](../adr/determinism.md), a user
can rebuild a tagged version and verify the result against
`SHA256SUMS.txt`. The exact build environment (Go version, toolchain)
is captured in the release notes of each tag for full reproducibility.

## Distribution channels

| Channel                                    | v0 / v1 | v2+        |
| ------------------------------------------ | ------- | ---------- |
| GitHub Releases (raw binaries + tarballs)  | yes     | yes        |
| `SHA256SUMS.txt`                           | yes     | yes        |
| `cosign` / `gpg` signature                 | no      | candidate  |
| Homebrew tap (separate repo)               | yes     | yes        |
| `go install ...@latest`                    | yes     | yes        |
| asdf plugin                                | no      | candidate  |
| `scoop` / `winget`                         | no      | candidate  |
| Docker image                               | no      | candidate  |

### Homebrew tap

Lives in a separate repo (e.g. `andreswebs/homebrew-tap`) with
`Formula/terminology.rb`. **Manual** updates during v0 — hand-edit the
formula on each release. Once v1 cadence stabilizes, add a GitHub
Actions workflow triggered by tag push that rewrites the formula.
Avoids upfront CI plumbing while the release cadence is still being
discovered.

### Signing

`SHA256SUMS.txt` is the integrity floor for v1. Code signing
(`cosign`/Sigstore or GPG) is **deferred to v2**: single-developer
project, key-management and user-education overhead disproportionate
to the v1 surface. Revisit when adoption justifies the operational
cost — by v2 the candidate is Sigstore keyless signing via CI OIDC
rather than long-lived GPG keys.

## Docs / CI sync

Three CI-level assertions, inherited from [E9](009-hardening.md) scope
(deferred here because they are release-readiness concerns, not runtime
hardening):

1. **Scenarios parse.** Every fenced `terminology …` invocation in
   [docs/examples/scenarios.md](../examples/scenarios.md) parses via
   urfave's argv parser (not full action invocation). Catches doc rot
   when commands or flags rename.
2. **Errors reference is fresh.** `docs/reference/errors.md` is
   generated from the `terr` sentinel registry. CI re-runs the
   generator and fails on diff. Ensures documented codes match
   compiled-in codes.
3. **Help-text health.** Every flag declares a non-empty `Usage`;
   every command declares a non-empty `Description`. Runs in
   `make lint`.

## Changelog

No formal changelog policy in v0. Git history + release tags are the
record. A `CHANGELOG.md` (Keep-a-changelog style) or
`gh release --generate-notes` workflow can be adopted later if the
release cadence justifies the maintenance — neither is in scope now.

## Hand-offs

- Version embedding mechanics:
  [001 §Q9 (cli-surface-stub)](001-cli-surface-stub.md).
- Schema-versioning machinery:
  [schema-source-of-truth](../adr/schema-source-of-truth.md).
- CI gates and the perf-budget schedule:
  [testing](../adr/testing.md).
- Reproducible-build flags carried by the writer's determinism
  guarantees: [determinism](../adr/determinism.md).
- Build & validation entry points: `Makefile` + `CLAUDE.md`.
