# Cross-cutting — Testing strategy

> **Status**: APPROVED. The test layers every epic plugs into.

## Goals

- Every command's surface enforced (urfave-level rejection of bad args).
- Every emitted envelope conforms to its declared Go struct shape — see
  [schema-source-of-truth](schema-source-of-truth.md).
- Every `internal/` package covered by unit + (where useful) property tests.
- Round-trip invariants asserted: `read → write → read` is identity (on
  canonical input).
- Matcher and parser fuzzed.
- Perf budget enforced.

## Test layers

### 1. Unit tests

Standard table-driven `_test.go` files per package. The first reach for
every package.

Convention: each table row is named; failures print row name + `diff`-ish
output of expected vs actual.

### 2. Golden CLI tests

End-to-end at the urfave level: argv + stdin → stdout/stderr/exit-code
triples, captured as **byte-for-byte golden files** in `testdata/`.

```go
// internal/app/commands_test.go
func TestValidateGoldens(t *testing.T) {
    for _, tc := range []struct {
        name string
        argv []string
        stdin string
    }{
        {"clean",    []string{"terminology", "validate", "--tbx", "testdata/clean.tbx"}, ""},
        {"warnings", []string{"terminology", "validate", "--tbx", "testdata/dirty.tbx"}, ""},
        {"missing",  []string{"terminology", "validate", "--tbx", "testdata/missing.tbx"}, ""},
    }{
        t.Run(tc.name, func(t *testing.T) { runGolden(t, tc.argv, tc.stdin) })
    }
}
```

`runGolden`:

- Builds `Root()`, sets `Writer`/`ErrWriter`/`Reader` to buffers.
- Injects a fixed clock via `clock.With(ctx, fakeClock{T: testTime})`
  (see [determinism](determinism.md)) so
  transaction timestamps are stable across runs and hosts.
- Runs.
- Compares `stdout`, `stderr`, exit code against `testdata/<name>.stdout`,
  `testdata/<name>.stderr`, `testdata/<name>.exit` via `bytes.Equal`.

Byte-for-byte is the right comparator here because the
[determinism contract](determinism.md) guarantees stable
output. Structural JSON comparison would tolerate exactly the bugs
(key-order, whitespace, sub-second jitter) the determinism layer is
there to prevent.

### 3. Envelope conformance

Every JSON envelope emitted by the binary matches the shape declared
by its Go output-struct type in `internal/output/types.go`. Conformance
is enforced by **construction**, not by runtime schema validation:

- Envelopes are built from typed structs, serialized via `encoding/json`.
- Goldens capture the bytes that serialization produces.
- Drift between the struct definition and the emitted output is caught
  the moment the golden file changes.

There is no JSON Schema validator in the test stack. The Go type system
plus deterministic serialization plus golden files are the contract
(see [schema-source-of-truth](schema-source-of-truth.md)).

A small helper in `internal/output/conformance_test.go` enforces the
invariant that every envelope carries `schema_version` and exactly one
of `ok: true` (success) or `ok: false, error: {...}` (fatal) at the
top level — bookkeeping that's easier to assert in code than in every
golden file. The helper is invoked by `runGolden`.

### 4. Property tests

Round-trip invariants:

```go
// internal/tbx/linguist/roundtrip_test.go
func TestRoundTrip(t *testing.T) {
    for _, path := range listFixtures(t, "testdata/canonical/") {
        original := readFile(t, path)
        g, _, err := Reader{}.Decode(bytes.NewReader(original))
        require.NoError(t, err)

        var buf bytes.Buffer
        require.NoError(t, Writer{}.Encode(&buf, g))

        if !bytes.Equal(original, buf.Bytes()) {
            t.Errorf("round-trip mismatch in %s:\n%s", path, diff(original, buf.Bytes()))
        }
    }
}
```

The `testdata/canonical/` corpus contains only canonically-formatted TBX
files. Non-canonical inputs land under `testdata/normalized/` with a
`.expected` sibling holding the post-normalization form.

Other property tests:

- ID derivation: `DeriveID(s)` is deterministic and idempotent on its
  output (`DeriveID(DeriveID(s)) == DeriveID(s)`).
- Matcher: a match at `(line, col)` always corresponds to a glossary term
  surface, never to an unrelated substring.

### 5. Fuzz

Run nightly in CI (`go test -fuzz=. -fuzztime=5m` per target):

- `internal/tbx/linguist.Reader.Decode` — must never panic.
- `internal/match.Matcher.Scan` — must never panic, must terminate.
- `internal/write.DeriveID` — must produce a valid ID or a typed error.

Crashers stored in `testdata/fuzz/<target>/`. Each crasher promoted to a
unit test in the green phase.

### 6. Perf budget

`*_perf_test.go` files marked with `//go:build perf` build tag. Run via
`make perf` and in CI on a separate workflow. Each asserts on a budget
defined in [E9](specs/009-hardening.md) §"Perf budget".

```go
//go:build perf

func TestValidatePerf_10K(t *testing.T) {
    g := loadFixture(t, "testdata/perf/10k-concepts.tbx")
    start := time.Now()
    _ = g.Validate(false)
    if d := time.Since(start); d > 500*time.Millisecond {
        t.Errorf("validate took %s, budget 500ms", d)
    }
}
```

### 7. Integration (real binary)

A small set of tests under `integration_test/` shells out to the built
binary (`./bin/terminology-<host>-<arch>`) with real argv/stdin and
asserts on real stdout/stderr/exit. Caps the "did we forget to wire the
main loop" risk that golden tests miss.

Run via `make test-integration`, after `make build`.

## Golden update mechanism

Each golden-running test defines a `flag.Bool("update", false, ...)`.
Invoking `go test ./... -update` rewrites every golden file to match
the current output. CI runs without `-update`, so any unintentional
output change fails the build until the developer regenerates and
commits the affected goldens. PR review catches the unintended via
`git diff`.

## Cross-platform discipline

Goldens are a single set; no `linux/` vs `windows/` subdirectories.
Three sources of platform drift are eliminated at the test boundary:

- **Timestamps.** Fake clock injected via `clock.With(ctx, fakeClock)`
  in every golden-running test. No real `time.Now()` reaches the
  output.
- **Line endings.** Test helpers normalize CRLF to LF on file read
  before comparison. The binary itself emits LF (per
  [determinism](determinism.md)); the normalization is a
  safety net for fixtures committed from Windows checkouts.
- **Paths.** Where paths appear in output (rare — most envelopes carry
  IDs, not paths), the test compares via `filepath.ToSlash`.

OS-specific goldens are an anti-pattern here: they hide platform bugs
behind "just update the Windows version" PR comments.

## Coverage policy

Coverage is **tracked but not gated**. `make test` emits a coverage
profile; `make validate` does not fail on a percentage threshold.

Gating by number incentivizes the wrong work — adding tests for the
line, not the behavior. The agent-facing contract (envelope shape,
exit code, error code) is what matters, and golden + conformance tests
cover that directly regardless of what a coverage tool reports. We
revisit if quality slips visibly.

## Test execution time

No proactive build-tagging of slow tests beyond the existing `perf`
suite. `make test` runs the full default suite — if it gets slow,
that's the signal to split. Pre-tagging would be optimization without
a problem.

The matcher's "compile N regexes" test on a 10k-term fixture is the
known watch item; for now it stays in the default suite. If it starts
dominating the wall clock, it migrates to `perf` like the other
fixture-driven heavy tests.

## Test fixtures layout

```
src/internal/tbx/linguist/testdata/
├── canonical/                   # round-trip identity
│   ├── minimal-dct.tbx
│   ├── linguist-with-transactions.tbx
│   └── ...
├── normalized/                  # legacy → canonical
│   ├── usage-register.tbx
│   ├── usage-register.expected
│   └── ...
├── invalid/                     # parse errors
│   ├── missing-langSec.tbx
│   └── ...
└── fuzz/
    ├── decode/
    └── ...

src/internal/app/testdata/
├── validate/
│   ├── clean.stdout
│   ├── clean.stderr
│   ├── clean.exit
│   └── ...
└── ...
```

Naming convention: one fixture per scenario, no monolithic
`big-fixture.tbx`.

## CI matrix

| Job           | Trigger      | Suite                                          |
| ------------- | ------------ | ---------------------------------------------- |
| `validate`    | PR + main    | `make validate` (fmt-check, vet, lint, test)   |
| `build`       | PR + main    | `make build`                                   |
| `cross`       | PR + main    | `make build-all`                               |
| `integration` | main         | `make test-integration`                        |
| `perf`        | nightly + tag| `make perf`                                    |
| `fuzz`        | nightly      | `go test -fuzz=. -fuzztime=5m` per target      |
