---
id: ter-lot8
status: closed
deps: []
links: [ter-ve3i]
created: 2026-07-24T20:37:13Z
type: task
priority: 1
assignee: Andre Silva
---

# Migrate Go module from src/ to repo root

## Goal

The `terminology` repository keeps its Go module under `src/` (`src/go.mod`),
but `go.mod` declares the module path `github.com/andreswebs/terminology`. The
Go toolchain and module proxy expect the module named by that path to live at
the repository root. As published today the module is unreachable through the
proxy:

- `go install github.com/andreswebs/terminology/cmd/terminology@latest` fails.
- pkg.go.dev, govulncheck, and dependabot cannot resolve the module.

This plan moves the module contents from `src/` up to the repository root.
Import paths do NOT change, because the declared module path already omits
`src`. No `.go` file needs editing. `terminology` is a standard-shape CLI
(only `cmd/` and `internal/`, no Docker, no Nix, no Atlas) and is the first
repo to run this recipe, so the emphasis is a minimal, provable diff.

The module path is confirmed as `github.com/andreswebs/terminology` with
`go 1.26.5` and the main package at `cmd/terminology/main.go`.

## Current layout (relevant parts)

```text
terminology/
  Makefile
  AGENTS.md            (CLAUDE.md is a symlink to AGENTS.md)
  README.md
  .gitignore
  .editorconfig
  .envrc
  .github/
    dependabot.yml
    workflows/ci.yml
    workflows/release.yml
    actions/gh-release/...
  .vscode/settings.json
  docs/
  bin/                 (build output, gitignored)
  dist/                (build output, gitignored)
  src/
    .golangci.yml
    go.mod
    go.sum
    cmd/terminology/main.go
    internal/...
```

## Target layout

```text
terminology/
  .golangci.yml
  go.mod
  go.sum
  cmd/terminology/main.go
  internal/...
  Makefile, README.md, docs/, bin/, dist/, .github/, .vscode/  (unchanged shape)
```

The `src/` directory is removed entirely.

## Step 1: Move module contents to the repo root

Run from the repository root. `src/` contains exactly five entries:
`.golangci.yml`, `go.mod`, `go.sum`, `cmd/`, `internal/`.

```sh
git mv src/go.mod src/go.sum src/.golangci.yml src/cmd src/internal .
rmdir src
```

Notes:

- `git mv` preserves history and stages the moves. Use `git mv` (not a plain
  `mv`) so the change is a rename in the diff.
- After the move, `src/` must be empty; `rmdir src` removes it. If `rmdir`
  fails, run `ls -la src` to find leftover entries (there should be none) and
  move them too.
- No `.go` file changes. Import paths such as
  `github.com/andreswebs/terminology/internal/app` continue to resolve because
  the module path is unchanged and now correctly rooted.

## Step 2: Makefile

File: `Makefile`. The build/test/run targets `cd $(SRC_DIR)` before invoking
the Go toolchain. `SRC_DIR` is defined once at line 2.

Recommended change (minimal diff, chosen deliberately for this first repo so
the recipe is easy to review): repoint `SRC_DIR` at the repo root and leave the
`cd $(SRC_DIR) &&` prefixes in place.

Line 2, before:

```make
SRC_DIR     := $(CURDIR)/src
```

Line 2, after:

```make
SRC_DIR     := $(CURDIR)
```

`CMD_DIR := ./cmd/terminology` (line 5) is already relative to the module root
and is unchanged. `LDFLAGS` (line 7) references
`github.com/andreswebs/terminology/internal/version.Override`; the module path
does not change, so it is unchanged. The `cd $(SRC_DIR) &&` prefixes appear at
lines 29, 33, 69, 72, 75, 78, 81, 84, 87, 90; with `SRC_DIR := $(CURDIR)` they
become `cd $(CURDIR) && ...`, which is a harmless no-op and keeps the diff to a
single line.

Optional cleaner end state (NOT recommended for this first migration, listed
for completeness): delete the `SRC_DIR` variable and strip every
`cd $(SRC_DIR) &&` prefix so the recipes run directly at the repo root. This
touches ~10 lines instead of 1 and is better deferred to a follow-up so the
migration diff stays purely structural.

## Step 3: GitHub workflows (setup-go paths)

Two workflows point setup-go at the module under `src/`.

File: `.github/workflows/ci.yml`, lines 23-24.

Before:

```yaml
go-version-file: src/go.mod
cache-dependency-path: src/go.sum
```

After:

```yaml
go-version-file: go.mod
cache-dependency-path: go.sum
```

File: `.github/workflows/release.yml`, lines 32-33. Identical change:

Before:

```yaml
go-version-file: src/go.mod
cache-dependency-path: src/go.sum
```

After:

```yaml
go-version-file: go.mod
cache-dependency-path: go.sum
```

Neither workflow has `src/` in prose comments, so no comment edits are needed.

## Step 4: Release workflow SBOM step

File: `.github/workflows/release.yml`, the "Generate SBOM" step (lines 56-65).
The step currently scans `path: src`.

Before (line 62):

```yaml
with:
  path: src
```

After (line 62):

```yaml
with:
  path: .
```

`SYFT_SOURCE_NAME: terminology` and `SYFT_SOURCE_VERSION` (lines 59-60) already
name the SBOM root package by the repo/release identity and are unchanged.
Nothing about the cosign signing (lines 68-76) or attestation
(subject-checksums, lines 80-83) steps references `src/`; they are unchanged.

### SBOM scope decision (resolves the draft's OPEN question)

Recommendation: keep `path: .` AND add a Syft exclude config that limits the
scan to the Go module by dropping build outputs and docs.

Rationale: the SBOM step runs AFTER "Build distribution archives"
(`make dist`, line 44-45), so by the time Syft runs, `bin/` holds the compiled
binaries and `dist/` holds the packaged archives and `SHA256SUMS.txt`. Today
`path: src` scans only Go source (`src/` has no build outputs), producing a
clean Go-dependency SBOM. A bare `path: .` at the root would additionally let
Syft's binary and archive catalogers ingest `bin/` and `dist/`, polluting the
SBOM with the released artifacts themselves and any docs assets. Excluding
those directories preserves the exact current semantics (Go module dependencies
only) while satisfying the module-root move.

Implementation: create `.syft.yaml` at the repo root. `anchore/sbom-action`
auto-discovers a `.syft.yaml` in the scan directory.

```yaml
exclude:
  - ./bin/**
  - ./dist/**
  - ./docs/**
  - ./.git/**
```

If explicit wiring is preferred over auto-discovery, pass the file through the
action instead:

```yaml
with:
  path: .
  config: .syft.yaml
  format: spdx-json
  output-file: dist/terminology-${{ github.ref_name }}.spdx.json
  upload-artifact: false
```

The two rejected alternatives from the draft: (a) accept the wider scan, which
produces a misleading SBOM that lists the release binaries as packages; (b)
point Syft directly at `go.mod`, which is brittle across sbom-action versions
and drops the `go.sum`-derived transitive graph. The exclude-config approach is
the robust middle ground and should be applied before the first
post-migration tagged release.

## Step 5: Dependabot

File: `.github/dependabot.yml`. The gomod ecosystem points at `/src` (line 10).

Before (lines 9-12):

```yaml
- package-ecosystem: gomod
  directory: /src
  schedule:
    interval: weekly
```

After:

```yaml
- package-ecosystem: gomod
  directory: /
  schedule:
    interval: weekly
```

The `github-actions` ecosystem entry (directory `/`) is unchanged.

## Step 6: Editor and env config

- `.vscode/settings.json` sets only `gopls.buildFlags: ["-tags=perf"]` and
  contains no `src/` reference. No change required. After migration, confirm
  gopls resolves the module at the repo root (it will, because `go.mod` is now
  at the root; the fleet `go.work` update in Step 8 also covers workspace mode).
- `.editorconfig`, `.envrc`, and `.gitignore` contain no `src/` references
  (`.gitignore` ignores `bin/`, `dist/`, `.envrc`, `.DS_Store`, `*.bak`,
  `*.bak.*`; all still valid at the root). No changes required.

## Step 7: Documentation sweep

`grep -rn 'src/' --exclude-dir=.git --exclude-dir=.tickets .` after Steps 1-6
should leave only documentation references. Update the following. In every case
the fix is to drop the leading `src/` segment (`src/internal/...` becomes
`internal/...`, `src/cmd/...` becomes `cmd/...`), because the paths those docs
point at now live at the repo root.

Primary docs (should be updated to reflect reality):

- `AGENTS.md` line 37-38: rewrite the "Go source: `src/`" bullet. Before:
  "Go source: `src/`. Entry point
  [`src/cmd/terminology/main.go`](src/cmd/terminology/main.go); all logic lives
  under `src/internal/` ...". After: "Go source lives at the repo root. Entry
  point [`cmd/terminology/main.go`](cmd/terminology/main.go); all logic lives
  under `internal/` (deliberately unimportable as a public library)." Note
  `CLAUDE.md` is a symlink to `AGENTS.md`, so editing `AGENTS.md` fixes both.
- `AGENTS.md` line 127-128: the "Build & Validation" paragraph says "Go source
  lives under `src/`." Change to "Go source lives at the repo root." (The
  `make` targets are described generically and need no other change.)
- `docs/learnings.md` line 681: `src/internal/write/jsonerr.go` becomes
  `internal/write/jsonerr.go`.
- `docs/adr/term-field-vocabulary.md` lines 11, 50: the relative links
  `../../src/internal/...` become `../../internal/...` (still two levels up from
  `docs/adr/` to the repo root, then into `internal/`).
- `docs/adr/picklist-validation-at-decode.md` lines 11, 13: same
  `../../src/internal/...` to `../../internal/...` fix.
- `docs/adr/testing.md` lines 210, 226: the fenced tree paths
  `src/internal/tbx/linguist/testdata/` and `src/internal/app/testdata/` become
  `internal/tbx/linguist/testdata/` and `internal/app/testdata/`.
- `docs/qa/*.md` FIXTURES shell exports: `docs/qa/E2-manual-qa.md` (lines 88,
  106), `docs/qa/E3-manual-qa.md` (lines 65, 66), `docs/qa/E4-manual-qa.md`
  (line 61), `docs/qa/E5-manual-qa.md` (line 64), `docs/qa/E6-manual-qa.md`
  (line 60), `docs/qa/E7-manual-qa.md` (line 66), `docs/qa/E8-manual-qa.md`
  (line 70). Each sets `export FIXTURES="src/internal/tbx/linguist/testdata"`
  (and E3 additionally `APP_FIXTURES="src/internal/app/testdata/fixtures"`);
  drop the `src/` prefix in each.
- `docs/qa/E8-manual-qa.report.md` lines 30, 43: prose "root cause" references
  to `src/internal/tbx/validate.go` and `src/internal/write/errors.go`; drop
  the `src/` prefix.

Historical specs (do NOT rewrite the body; these are point-in-time records):

- `docs/specs/001-cli-surface-stub.md` lines 18, 270 and
  `docs/specs/002-domain-and-io.md` line 43 embed `src/`-era paths and layout
  snapshots. Leave the historical text intact and add a single note line near
  the top of each affected spec, for example:
  "Note: the Go module moved from `src/` to the repo root after this spec was
  written; paths below that begin with `src/` now live at the repo root." This
  mirrors how historical specs are treated elsewhere in this repo.

Not a `src/` directory reference (do NOT change): `docs/research/`
`urfave-cli-reference.md` lines 293, 647 match the grep only because of the
package name `github.com/urfave/cli-altsrc/v3` (the substring "altsrc"). These
are unrelated and must be left alone.

Agent skill: `docs/skills/terminology/SKILL.md` invokes the installed
`terminology` CLI by command name, not by path, and has no `src/` reference. It
needs no change. Confirm after rebuilding that the skill still drives the CLI.

## Step 8: Fleet workspace (go.work)

This repo is consumed through the fleet workspace at the go-projects root. The
workspace file lists the module by its `src/` subdirectory.

File: `go.work` at the fleet root. Change the terminology entry:

Before:

```text
 ./terminology/src
```

After:

```text
 ./terminology
```

Leave the other repos' entries untouched; they migrate independently. If the
workspace is not needed for verification, this line can also be updated via
`go work use ./terminology` after removing the old entry. The fleet-root
`AGENTS.md` states each module lives at `<project>/src/`; that is a fleet-wide
statement and should be reconciled when the whole fleet has migrated, not in
this single-repo change.

## Step 9: Verification

Run from the terminology repo root after Steps 1-7:

```sh
# Structural sanity: src/ is gone, go.mod is at the root.
test ! -d src && test -f go.mod && echo "layout OK"

# Full quality gate (fmt-check, vet, lint, test) then compile.
make validate
make build

# Run the freshly built binary.
./bin/terminology-$(go env GOOS)-$(go env GOARCH) --version
./bin/terminology-$(go env GOOS)-$(go env GOARCH) schema

# Confirm the module still resolves within the fleet workspace (Step 8).
go build ./... 2>/dev/null || (cd "$(git rev-parse --show-toplevel)" && go build ./terminology/...)

# No stray src/ references remain except the intentional historical spec notes
# and the urfave-cli-altsrc false positives.
grep -rn 'src/' --exclude-dir=.git --exclude-dir=.tickets .
```

After the change is merged and a new version tag is pushed (tagging is left to
the repo owner), verify external installability from OUTSIDE the workspace, in
a scratch directory, so the module proxy is exercised rather than the local
checkout:

```sh
TMP="$(mktemp -d)"
cd "$TMP"
GOFLAGS=-mod=mod GOPROXY=proxy.golang.org \
  go install github.com/andreswebs/terminology/cmd/terminology@latest
"$(go env GOPATH)/bin/terminology" --version
```

The main package is `cmd/terminology`, so `@latest` on the module root path
alone will not install a binary; the `cmd/terminology` suffix is required.

Optional CI dry run: push the branch and confirm the CI workflow's setup-go
step caches on `go.sum` and `make build` passes. Do a pre-release check of the
SBOM step by confirming `.syft.yaml` excludes `bin/`, `dist/`, and `docs/`.

## Rollback

The change is a pure file move plus config edits and is fully reversible before
tagging. To roll back an unmerged branch, discard it (git checkout of the prior
commit / branch deletion, done by the repo owner). To reverse the moves
manually:

```sh
mkdir -p src
git mv go.mod go.sum .golangci.yml cmd internal src/
```

Then revert the `Makefile` (line 2 back to `$(CURDIR)/src`), the two workflow
files (setup-go paths back to `src/go.mod` and `src/go.sum`, SBOM `path` back to
`src`), `.github/dependabot.yml` (`directory: /src`), the doc edits, the
removed `.syft.yaml`, and the fleet `go.work` entry
(`./terminology` back to `./terminology/src`). Because no import paths or `.go`
files changed and no tag was published, there is nothing to retract from the
module proxy; old tags were never resolvable through the proxy in the first
place.

## Summary of files to change

| File                                                         | Change                                     |
| ------------------------------------------------------------ | ------------------------------------------ |
| `src/*` -> repo root                                         | `git mv` the 5 entries, `rmdir src`        |
| `Makefile` (line 2)                                          | `SRC_DIR := $(CURDIR)`                     |
| `.github/workflows/ci.yml` (23-24)                           | setup-go paths drop `src/`                 |
| `.github/workflows/release.yml` (32-33)                      | setup-go paths drop `src/`                 |
| `.github/workflows/release.yml` (62)                         | SBOM `path: src` -> `path: .`              |
| `.syft.yaml` (new)                                           | exclude `bin/`, `dist/`, `docs/`, `.git/`  |
| `.github/dependabot.yml` (10)                                | gomod `directory: /src` -> `/`             |
| `AGENTS.md` (37-38, 127-128)                                 | drop `src/`, rewrite source-location prose |
| `docs/learnings.md` (681)                                    | drop `src/`                                |
| `docs/adr/term-field-vocabulary.md` (11, 50)                 | `../../src/` -> `../../`                   |
| `docs/adr/picklist-validation-at-decode.md` (11, 13)         | `../../src/` -> `../../`                   |
| `docs/adr/testing.md` (210, 226)                             | drop `src/`                                |
| `docs/qa/E2..E8-manual-qa.md` (FIXTURES)                     | drop `src/`                                |
| `docs/qa/E8-manual-qa.report.md` (30, 43)                    | drop `src/`                                |
| `docs/specs/001-cli-surface-stub.md`, `002-domain-and-io.md` | add historical note, keep body             |
| `go.work` (fleet root)                                       | `./terminology/src` -> `./terminology`     |

No changes: `.vscode/settings.json`, `.editorconfig`, `.envrc`, `.gitignore`,
`README.md`, `docs/skills/terminology/SKILL.md`, `docs/research/`
`urfave-cli-reference.md` (false positive), and every `.go` file.

## Notes

**2026-07-24T21:00:47Z**

FLEET DECISIONS finalized (see .local/planning/go-module-root-migration.md
"Decisions"). All match this plan; no step changes required:

- Makefile: minimal diff — `SRC_DIR := $(CURDIR)`, keep the cd prefixes.
- SBOM: `path: .` plus a root `.syft.yaml` exclude (this plan's rationale
  about the Syft step running after `make dist` is the fleet-wide reason).
- `.github/dependabot.yml`: gomod `directory: /src` -> `/` (already flagged).

**2026-07-24T21:13:26Z**

DONE. Moved Go module from src/ to repo root via 'git mv src/{go.mod,go.sum,.golangci.yml,cmd,internal} .' + rmdir src (recorded as renames; pre-existing unstaged edits to extract.go/sanitize.go/commands_test.go/go.mod/go.sum and testdata golden files preserved as working-tree mods on top of the staged renames). No .go files edited; module path github.com/andreswebs/terminology unchanged so imports still resolve. Config edits: Makefile SRC_DIR := $(CURDIR); ci.yml + release.yml setup-go paths go.mod/go.sum; release.yml SBOM path: . plus new root .syft.yaml excluding bin/,dist/,docs/,.git/; dependabot.yml gomod directory /. Docs swept (AGENTS.md source-location prose, learnings.md, adr/testing.md, adr/picklist-validation-at-decode.md, adr/term-field-vocabulary.md, qa/E2-E8 FIXTURES exports + E2 entry criteria, E8 report root-cause paths). Historical specs 001/002 kept intact with a top note. Step 8 (fleet go.work) N/A: no go.work exists in this environment. Verified: make build (fmt-check/vet/lint/test) green from root, binary --version + schema OK, go build ./... OK, git renames recorded. GOTCHA: golangci-lint reported a phantom 'unused' issue against a stale src/ cache path after the move; 'golangci-lint cache clean' fixed it.
