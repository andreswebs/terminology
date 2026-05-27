# E2 — Domain Model & TBX I/O

> **Status**: APPROVED. Scope, architecture, and idioms for the
> dialect-agnostic domain model and the TBX-Linguist reader/writer pair.

## Scope

Stand up the dialect-agnostic **domain model** (the internal types every
other package consumes) and the **first concrete reader/writer pair** for
TBX-Linguist. This is the load-bearing seam: dialect-specific code is
isolated behind interfaces so a second TBX dialect (TBX-Basic, TBX-Min,
future TBX-3.0) can slot in as a sibling package without touching
commands.

In scope:

- Internal domain types (`Glossary`, `Concept`, `LangSection`, `Term`,
  `Transaction`, `NoteText`, …) — independent of any XML shape.
- `tbx.Reader` and `tbx.Writer` interfaces.
- `internal/tbx/linguist/` — concrete reader (DCT + DCA on input) and
  writer (canonical DCT on output).
- Style detection on read (`<tbx>` `@style` attribute).
- Legacy-form normalization on read (`usageRegister` → `register`; bare
  `preferredTerm` → `preferredTerm-admn-sts`; etc.).
- Atomic file write — temp-file + `os.Rename`, gated by an advisory
  cross-process lock.
- Hand-rolled deterministic XML emitter (handed to
  [determinism](../adr/determinism.md)).
- Round-trip property: `read → write → read` produces an equivalent
  model (modulo non-canonical input).

Out of scope (handed off):

- Validation logic → [E3](003-validate-command.md).
- Concept-ID derivation → [E7](007-write-commands.md) (lives in
  `internal/tbx/id.go` per [cli-design.md](../cli-design.md), but the
  rules are write-side concerns).
- Markdown parsing → [E5](005-matcher.md).

## Architecture

```
src/internal/tbx/
├── model.go         — Glossary, Concept, LangSection, Term, …
├── style.go         — Style enum + dialect identifier
├── reader.go        — Reader interface + dispatch on dialect
├── writer.go        — Writer interface + dispatch on dialect
├── io.go            — atomic write helper, Load/Save entry points
├── lock.go          — advisory lock acquisition (lockedfile wrapper)
├── errors.go        — terr sentinels (ErrUnsupportedDialect, ErrTBXLocked, …)
└── linguist/
    ├── reader.go    — DCT + DCA decoder
    ├── writer.go    — canonical DCT emitter (hand-rolled)
    ├── normalize.go — legacy-form normalization on read
    └── testdata/    — fixtures: minimal DCT, minimal DCA, mixed corpus
```

### Domain types

Domain types are sketched in [cli-design.md §`internal/tbx`](../cli-design.md).
Two specifics:

- **Exported fields**, not accessors. `c.ID`, `c.LangSections`, etc.
  Go convention for value types; trivial to inspect in tests and
  templates. Setters arrive only when an invariant needs protecting
  (e.g. when `concept add` enforces ID derivation).
- **Dialect identifier** lives as `tbx.Dialect` rather than as a
  hardcoded branch, so a future TBX-Basic reader/writer is a registered
  implementation.

### Reader / Writer interfaces

```go
type Reader interface {
    Decode(io.Reader) (*Glossary, []Warning, error)
}

type Writer interface {
    Encode(io.Writer, *Glossary) error
}
```

Top-level `tbx.Load(path)` opens the file, inspects `<tbx>` `@type`,
picks a registered `Reader`, and dispatches. Unsupported `@type` values
return `ErrUnsupportedDialect` (`unsupported_dialect`, exit 65) with a
hint pointing to the supported set (`supported: TBX-Linguist`).

### Reader warnings

The reader knows about legacy-form normalization and unknown elements.
Warnings flow back via `[]Warning` from `Decode`:

```go
type Warning struct {
    Code       string // e.g. "legacy_form_normalized", "unknown_element"
    Message    string
    ConceptID  string // optional
    Line, Col  int    // optional
}
```

`validate` surfaces them in its envelope (E3). Read commands (`lookup`,
`scan`, `extract`) discard reader warnings silently unless verbosity is
INFO or higher (see [logging](../adr/logging.md)),
in which case they're emitted via `logctx.From(ctx)`. Warnings as data
keeps the reader free of logging-policy concerns.

### Hand-rolled XML emitter

`encoding/xml` is fine for parsing but unsuitable for canonical output:
it injects namespace prefixes nondeterministically, mishandles
whitespace around nested elements, and cannot produce self-closing
empty elements on demand. `internal/tbx/linguist/writer.go` implements
a focused emitter (~200 lines) that emits exactly the canonical form
defined in [determinism](../adr/determinism.md):

- Fixed namespace prefixes: `min:`, `basic:`, `ling:`.
- Element order matches the TBX-Linguist schema (parent → child
  ordering is dialect-defined).
- Attribute order canonicalized.
- Languages emitted by `xml:lang` ASCII byte order.
- Terms within a `<langSec>`: preferred → admitted → deprecated →
  superseded → unspecified; secondary key = declaration order on read.
- 2-space indent, LF line endings, single trailing newline.
- UTF-8 BOM-free.
- Self-closing for empty elements.
- XML comments stripped on write (the dialect's `<note>` element is
  the first-class place for commentary).

A focused emitter is cheaper than fighting `encoding/xml`'s quirks and
avoids a third-party dependency; matches the stdlib + urfave + x/text
posture in cli-design.md.

### Round-trip fidelity for unknown elements

Elements outside the TBX-Linguist supported set are **dropped on write
in v1**. `validate --strict` emits an `unknown_element` warning so
users see what's being lost. Round-trip fidelity for arbitrary
extensions is a TBX-3.0-class problem we defer; the model types stay
free of opaque `[]xml.Token` fields, equality/diff/JSON serialization
stay simple.

### DCA-on-emit

cli-design.md declares DCA-on-emit out of scope for v1. The `Writer`
interface is dialect-shaped, so a future `linguist.DCAWriter` is an
additive sibling implementation — no breaking change. v1 ships only
`linguist.DCTWriter`.

## Atomic write contract

Every successful write produces a complete, valid TBX or leaves the
target file untouched:

1. Acquire the advisory write lock (see below).
2. Write the new content to `${TARGET}.tmp-XXXX` in the same directory
   as `${TARGET}`.
3. `fsync` the temp file.
4. `os.Rename` over the target — atomic on POSIX, atomic-enough on
   Windows via `os.Rename`.
5. Release the lock.
6. On any error before rename, remove the temp file.

Never leave half-written state on disk; never leave a temp file behind
on the success path.

## Cross-process write safety

Two agents invoking `concept add` against the same TBX simultaneously
would otherwise last-rename-win. The writer takes an **advisory file
lock** before the read-modify-write window:

- **Lock target**: a sibling `.lock` file — `${TBX}.lock` — created
  next to the TBX. Separate file (not the TBX itself) so the lock
  fd's lifecycle is independent of the file replacement done by the
  rename. The lockfile is left in place between runs; it's a marker,
  not state, and is safe to commit-ignore.
- **Acquisition**: **non-blocking**. If another process holds the
  lock, return `ErrTBXLocked` (`tbx_locked`, exit 3 — I/O class) with
  a hint: `another terminology process is writing; retry`. Agents see
  deterministic failure rather than hangs.
- **Implementation**: `github.com/rogpeppe/go-internal/lockedfile` —
  used by the Go toolchain itself for `GOPATH` and module-cache
  locking. Handles POSIX `flock` and Windows `LockFileEx` correctly.
  Single small dependency; the only third-party Go library added by
  this epic.
- **Scope**: the lock is held across read → modify → write → rename.
  Released on function exit (defer).

A new error sentinel lands with this epic:

```go
// internal/tbx/errors.go
var ErrTBXLocked = terr.New(
    "tbx_locked", 3,
    "another terminology process is writing; retry",
    "TBX file is locked by another process",
)
```

Read commands (`validate`, `lookup`, `scan`, `extract`) do **not**
acquire the lock — they only need consistency for the duration of the
read, which `os.Open` already gives via the kernel's open-file table.

## Public API of model types

Domain fields are **exported**:

```go
type Concept struct {
    ID           string
    LangSections []LangSection
    Transaction  *Transaction
    // ...
}
```

Go convention for value types; lets tests, templates, and the writer
inspect freely. Methods land later only when there is an invariant to
defend — for v1, model types are plain data.

## Hand-offs

- Reader is consumed by every read command — first real consumer is
  [E3 `validate`](003-validate-command.md).
- Writer is consumed by every write command — first real consumer is
  [E7](007-write-commands.md).
- Atomic-write helper used by both E7 and [E8 `apply`](008-apply.md).
- `ErrTBXLocked` and `ErrUnsupportedDialect` are registered as
  sentinels per
  [error-handling](../adr/error-handling.md).
- Determinism guarantees:
  [determinism](../adr/determinism.md).
- Round-trip property test:
  [testing](../adr/testing.md).
