# E9 — Hardening: input validation, fuzz, perf, security

> **Status**: APPROVED. Cross-cutting safety and regression
> infrastructure that becomes its own epic once the surface and the
> commands exist. Some rules are referenced inline by E2/E3/E5/E7;
> E9's job is the omnibus authoritative statement plus the CI
> machinery.

## Scope

- `internal/app/commands/sanitize.go` — input validation for concept
  IDs, language tags, file paths, terms (unexported helpers in the
  commands package, not a separate package).
- XML parser hardening (XXE, bombs, DOCTYPE policy).
- Bounded reads — refuse pathologically large files.
- Path sandbox — refuse outputs outside CWD subtree; constrained
  policy for `--tbx`.
- Fuzz testing — matcher, TBX parser, ID derivation.
- Perf budget — fixed numbers; CI fails the build if exceeded.
- Docs/CI sync — assertions that survive the Go-code-as-source-of-truth
  pivot.

## Input hardening rules

Per cli-design.md §"Input hardening":

- Reject control characters in IDs / lang tags / paths.
- Reject path traversal (`..`) — both literal and percent-encoded.
- Reject percent-encoded segments.
- Reject embedded query params (`?`, `#`) in path args.
- Output paths sandboxed to CWD subtree.

API (unexported, in `internal/app/commands`):

```go
func sanitizeConceptID(s string) error
func sanitizeLangTag(s string) error
func sanitizePath(s string, baseDir string) (string, error)  // cleaned absolute path within baseDir
func sanitizeTerm(s string) error
```

Error envelope codes: `invalid_id`, `invalid_lang_tag`,
`invalid_path`, `invalid_term`.

### Where hardening runs

Validation lives at the **command-action boundary** — the urfave action
handler calls `sanitize*` before any I/O or domain call. Inner packages
(`internal/tbx`, `internal/match`, `internal/write`) assume their
inputs are clean. Single point of trust per request.

The test table in `sanitize_test.go` is the authoritative spec of
accepted vs rejected patterns. cli-design.md §"Input hardening"
references it rather than restating; drift between docs and behavior
is structurally impossible because the tests are the spec.

## XML parser hardening

Go's `encoding/xml` does **not** resolve external entities by default —
the XXE class is closed by the standard library. Additional measures:

### DOCTYPE policy

Accept the innocuous form (`<!DOCTYPE tbx>` with no internal subset
and no external ID); reject anything carrying entity declarations or
`SYSTEM`/`PUBLIC` URLs:

```xml
<!DOCTYPE tbx>                              -- accepted
<!DOCTYPE tbx [<!ENTITY x "...">]>          -- rejected (entities)
<!DOCTYPE tbx SYSTEM "http://...">          -- rejected (external ID)
<!DOCTYPE tbx PUBLIC "..." "...">           -- rejected (external ID)
```

Rationale: TBX files in the wild sometimes carry the bare
`<!DOCTYPE tbx>` header. Rejecting all DOCTYPEs would break those
files for no security gain. Internal subsets and external IDs are
the actual XXE-class smells, and those we reject hard.

Detection happens via a small streaming scan before handing the byte
stream to `encoding/xml.Decoder` — the standard decoder doesn't
surface DOCTYPE shape, so a wrapping check is required.

### Nesting depth + entity expansion

`xml.Decoder.Strict = true` plus a depth-counter wrapper. Hard cap on
element nesting depth (256 levels — TBX-Linguist files rarely exceed
8). Above the cap → `invalid_input` with `nesting_too_deep` hint.

## Bounded reads

| Input                          | Cap   | Rationale                                                          |
| ------------------------------ | ----- | ------------------------------------------------------------------ |
| TBX file (`--tbx`)             | 50 MB | ~100k+ concepts. Above this, batch is the answer.                  |
| Markdown file (`scan`/`check`) | 10 MB | Realistic chapters are <1 MB; 10 MB is a giant margin.             |
| Stdin payload (`apply`, write) | 10 MB | Same.                                                              |
| `extract` per-file             | 10 MB | Same.                                                              |
| Glossary in memory             | —     | Bounded transitively by the TBX cap.                               |

`io.LimitedReader` wraps every external input. Hitting the limit
returns `input_too_large` with the cap echoed in the message and
hint.

```go
var ErrInputTooLarge = terr.New(
    "input_too_large", 2,
    "split the input into smaller files or batches",
    "input exceeds maximum size",
)
```

## Path sandbox

Output paths and `--file` are sandboxed to the CWD subtree:

1. `filepath.Clean`.
2. `filepath.Abs`.
3. Assert the result is under CWD.
4. Resolve symlinks first; reapply the prefix check (refuses symlinks
   that escape).

### `--tbx` exemption

`--tbx` is **exempt from the CWD sandbox** — agents legitimately pass
absolute paths to shared glossaries outside CWD. The path is still
sanitized: literal `..` segments and percent-encoded segments are
rejected (typo/injection guardrail), but the resulting absolute path
is not pinned to CWD.

Documented asymmetry: input glossary may live anywhere; output
artifacts (and `--file` write payloads) stay inside CWD.

## Fuzz tests

Targets, in priority order:

1. **`internal/tbx/linguist.Reader.Decode`** — fuzz on arbitrary bytes;
   must never panic. Crashers go to `testdata/fuzz/`.
2. **`internal/match.Matcher.Scan`** — fuzz on arbitrary text +
   small in-memory glossaries; must never panic, must terminate.
3. **`internal/write.DeriveID`** — fuzz on arbitrary strings; must
   produce a syntactically valid ID or return `invalid_id`.

Run as `go test -fuzz=. -fuzztime=30s` in CI on a **nightly**
schedule, not on every PR — fuzz time is incompatible with PR latency.
Linux-only in CI; macOS and Windows runners execute the regular tests.
Shared fuzz corpus committed under `testdata/fuzz/`.

## Perf budget

Targets are set *before* optimizing; the build fails when exceeded.

| Op         | Input                                       | Target  |
| ---------- | ------------------------------------------- | ------- |
| `validate` | 10k-concept TBX                             | <500 ms |
| `lookup`   | 10k-concept TBX                             | <50 ms  |
| `scan`     | 200-concept TBX × 100 KB markdown           | <100 ms |
| `scan`     | 5000-concept TBX × 5 MB markdown × 1 file   | <5 s    |
| `check`    | same as `scan` × 2                          | <10 s   |
| `extract`  | 1 MB markdown corpus                        | <2 s    |

Numbers are hardcoded constants in `_test.go` files; ratchet via PR
review. A config file (`testdata/budget.json` etc.) would invite
silent loosening; PR diff visibility is the regression mechanism.

Tests live in `internal/<pkg>/perf_test.go` behind the `perf` build
tag, per [testing](../adr/testing.md). Run by
`make perf`.

## Docs / CI sync

> **Deferred to [E10](010-release.md).** These are release-readiness
> concerns, not runtime hardening. The assertions are specified here
> for completeness but implemented as E10 tasks.

Three CI-level assertions, post-pivot:

1. **Scenarios parse.** Every invocation in
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

The "schema ↔ JSON declaration" check that the original draft
described is **obsolete** — Go code is the source of truth (per
[schema-source-of-truth](../adr/schema-source-of-truth.md)),
so there is no second declaration to drift against.

## Hand-offs

- Used by every command that takes a path, ID, or term — boundary
  calls to `sanitize*` helpers in `internal/app/commands`.
- Perf budget infrastructure:
  [testing](../adr/testing.md).
- Generated errors reference:
  [schema-source-of-truth](../adr/schema-source-of-truth.md).
- Error envelope codes:
  [error-handling](../adr/error-handling.md).
- DOCTYPE policy applies to readers in [E2](002-domain-and-io.md) and
  the streaming validator pipeline in [E3](003-validate-command.md).
