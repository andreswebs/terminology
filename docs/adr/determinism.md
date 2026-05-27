# Cross-cutting — Determinism

> **Status**: APPROVED. Reproducibility rules for XML and JSON output,
> the time and order disciplines, and the reproducible-build contract.

## The rule

Same input → same bytes out. Always. Across runs, across hosts.

Without this, golden tests are impossible, diffs are noisy, and
write-command output isn't reviewable. Determinism is upstream of every
test layer.

## Domains

### 1. XML output (TBX writes)

The writer ([E2](specs/002-domain-and-io.md)) must produce byte-identical
output for byte-identical input. Rules:

- **Fixed namespace prefixes.** `min:`, `basic:`, `ling:`. Even if the
  input used different prefixes, the output uses the fixed set.
- **Element order matches the TBX-Linguist schema.** Parent-to-child
  ordering is dialect-defined. For sibling elements within a parent
  (e.g. multiple `<termSec>` in a `<langSec>`), apply secondary keys:
  - `<termSec>`: ordered by status (preferred → admitted → deprecated →
    superseded → unspecified), secondary key = declaration order on read.
  - `<langSec>`: ordered by `xml:lang` (ASCII byte sort).
  - `<conceptEntry>`: ordered by `id` (ASCII byte sort).
- **Attribute order canonicalized.** Within an element, attributes in
  fixed order: namespace declarations, then `id`, then `xml:lang`, then
  domain attributes.
- **Whitespace.** 2-space indent, LF line endings, single trailing newline.
- **No UTF-8 BOM.**
- **Self-closing for empty elements.**
- **Comments are stripped on write.** The dialect doesn't require XML
  comments; preserving them creates "is this comment authoritative?"
  ambiguity and complicates every domain type with comment-attachment
  metadata. Users who want commentary use the TBX `<note>` element —
  the dialect's first-class place for it.

`encoding/xml`'s emitter doesn't do most of this — implementation will
likely hand-roll the writer atop the decoder. See E2 Q2.

### 2. JSON output (envelopes)

- **Top-level key order.** Fixed per envelope shape: `schema_version`
  first, then `ok`, then payload keys in spec-declared order, then
  `warnings`. Error envelopes: `schema_version`, `ok`, `error`.
- **Inner objects.** Struct-based — Go's `encoding/json` emits fields in
  declared order. Maps are sorted by key.
- **Arrays.** Preserve declared/computed order — never sort silently
  unless the spec says so (e.g. `validate.languages` alphabetical,
  `scan.matches` by `(line, column)`).
- **Numbers.** Integers as integers, no scientific notation. Floats
  unused in the schema.
- **No trailing whitespace, no trailing newline.** (`encoding/json` does
  this naturally.)
- **Indentation.** Compact by default (no indent). `--format=json` is
  the agent contract; pretty-print is a v2 affordance via `--pretty`.

### 3. Concept-ID derivation

Per cli-design.md §"Concept IDs", derivation is deterministic:

- NFKD normalize.
- Drop combining marks.
- Lowercase under Unicode default casefold.
- Non-`[a-z0-9]` runs → single `-`.
- Trim leading/trailing `-`.
- Truncate to 64 codepoints on a `-` boundary.

No host-, locale-, or time-dependence. Property test: `DeriveID(s)` is
pure.

### 4. Transaction timestamps

Every `<transacGrp>` written by a write command carries a `<date>` in
**RFC3339 seconds-precision UTC**:

```xml
<date>2026-05-21T12:34:56Z</date>
```

Produced via `time.Format(time.RFC3339)` on a UTC `time.Time`. Matches
the `<basic:date>` examples in TBX-Linguist samples; sub-second
precision adds noise without use case (write commands aren't sub-second
events).

Two determinism risks the format addresses:

- **Wall clock.** Real `time.Now()` changes every run.
- **Timezone.** Local time leaks host context.

Both are mitigated by the **injectable clock** pattern below.

### 5. Clock injection

Write commands read the current time via `internal/clock`, which sits
on `context.Context` exactly like `internal/logctx`:

```go
package clock

import (
    "context"
    "time"
)

type Clock interface {
    Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

var Real Clock = realClock{}

type ctxKey struct{}

func With(ctx context.Context, c Clock) context.Context {
    return context.WithValue(ctx, ctxKey{}, c)
}

func From(ctx context.Context) Clock {
    if c, ok := ctx.Value(ctxKey{}).(Clock); ok {
        return c
    }
    return Real
}

func Now(ctx context.Context) time.Time { return From(ctx).Now() }
```

`Real.Now()` returns UTC, so callers cannot accidentally emit
host-local time. Tests inject a fixed clock:

```go
ctx := clock.With(context.Background(), fakeClock{T: testTime})
```

Package-level globals (`var now = time.Now`) are explicitly **not used**
— same anti-pattern we rejected for the logger.

### 6. Run-ids / UUIDs

The `run_id` attribute on log records (see
[logging](logging.md)) is non-deterministic
— that's intentional, it's an observability hook. Logs are not part of
the deterministic-output contract.

If a write command ever needs an identifier embedded in output (it
shouldn't), generate it from content-hash, not from a UUID.

### 7. Sort orders summary

All sorts are **ASCII byte sort** unless noted. BCP 47 tags are an ASCII
subset, so byte sort produces stable, locale-independent ordering
across hosts. The cost is that combining-script edge cases (none in
practice for BCP 47) would not collate "linguistically" — documented
limitation.

| Field                                            | Sort                                                                       | Tiebreak                  |
| ------------------------------------------------ | -------------------------------------------------------------------------- | ------------------------- |
| `<conceptEntry>` siblings                        | by `id` (ASCII)                                                            | n/a                       |
| `<langSec>` siblings                             | by `xml:lang` (ASCII)                                                      | n/a                       |
| `<termSec>` siblings                             | by status (preferred → admitted → deprecated → superseded → unspecified)   | declaration order on read |
| `validate.languages` array                       | by tag (ASCII)                                                             | n/a                       |
| `lookup.results` array                           | match-order — see [E5 Q5](specs/005-matcher.md#q5--match-precedence-on-overlap) | n/a                  |
| `scan.matches` array                             | `(line, column)`                                                           | n/a                       |
| `extract.candidates` array                       | by `frequency` descending                                                  | by surface (ASCII)        |
| `apply.applied.{added,updated,removed,unchanged}`| by `concept_id` (ASCII)                                                    | n/a                       |

## Read/write canonicalization model

**Canonicalization happens on write only.** The reader preserves input
order in the domain model; the writer applies the canonical sort rules
above on emit. Reasons:

- A hand-curated TBX may have meaningful term order within a `<langSec>`
  (e.g. preferred-then-rare ordering not reducible to status alone).
  Keeping that in memory until the writer chooses to re-emit preserves
  the option to inspect "what the file actually said".
- Sort-on-write is sufficient for byte-determinism — there is no test
  that requires the in-memory model to be sorted.
- Sorting at both layers is belt-and-suspenders with no payoff.

`validate` and `lookup` read from the model directly and therefore see
the input's original order. `apply` and the write commands re-emit
through the canonical writer.

## What is intentionally non-deterministic

- **Logs** (`run_id`, timestamps, durations). Diagnostic.
- **Process exit timing.** Obviously.
- **Network anything.** There is no network.

## Reproducible-build guarantee

The binary itself is reproducible: same source tree + same Go toolchain
version → same binary bytes, across hosts.

`Makefile` build flags:

```make
GOFLAGS := -trimpath
LDFLAGS := -buildid=
GO_BUILD := go build $(GOFLAGS) -ldflags '$(LDFLAGS) -X $(VERSION_PKG).version=$(VERSION)'
```

- `-trimpath` removes absolute build paths from the binary.
- `-buildid=` zeroes the build-id (which otherwise varies by linker run).
- Version metadata is injected explicitly via `-X` from `$(VERSION)`
  rather than read from `time.Now()` or `hostname` at build time.
- `-buildvcs=true` (Go's default) embeds VCS revision deterministically
  from the git checkout state.

Verified per release: `make build` on two hosts produces matching
`sha256sum bin/terminology-*`. Tracked alongside
[E10 Q8](specs/010-release.md#q8--reproducible-builds).
