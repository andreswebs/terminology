# Cross-cutting — Error handling

> **Status**: APPROVED. Architecture and idioms for how errors flow from
> origin → envelope → exit code. Touches every epic.

## Goal

A single, consistent path for errors to leave the binary. Every error a
caller (agent or human) sees emerges via the documented envelope shape
with the documented exit code. No early `os.Exit`, no `fmt.Println` to
stderr as a side channel, no "I forgot to handle that error".

## Layers

### Layer 1 — Origin (typed, coded errors)

Errors originate at package boundaries (`internal/tbx`,
`internal/match`, `internal/write`, `internal/harden`, ...). Each
originator implements a shared interface defined in `internal/terr`:

```go
package terr  // internal/terr

type Coded interface {
    error
    Code() string         // envelope code (e.g. "no_tbx_path")
    ExitCode() int        // process exit code
    Hint() string         // optional, may be ""
}

type E struct {
    code, msg, hint string
    exit            int
    cause           error
}

func New(code string, exit int, hint, format string, args ...any) *E
func (e *E) Error() string  { return e.msg }
func (e *E) Code() string   { return e.code }
func (e *E) Hint() string   { return e.hint }
func (e *E) ExitCode() int  { return e.exit }
func (e *E) Unwrap() error  { return e.cause }
```

Every error code has exactly one exported **package-level sentinel** in
its producing package, named `Err` + CamelCase of the code. The set of
sentinels IS the error registry — there is no separate declarative
source (see
[schema-source-of-truth](schema-source-of-truth.md)):

```go
// internal/tbx/errors.go
package tbx

var ErrUnsupportedDialect = terr.New(
    "unsupported_dialect", 65,
    "supported: TBX-Linguist",
    "unsupported TBX dialect",
)

var ErrNoTBXPath = terr.New(
    "no_tbx_path", 2,
    "set --tbx or TERMINOLOGY_TBX",
    "no TBX file path provided",
)
```

Sentinels live with the producing package (not in `terr`) to avoid circular
imports and to keep the constructor co-located with the call sites that
return it. `internal/terr` collects them at package init time so the
full registry is available for the reflective
[`terminology schema`](schema-source-of-truth.md#agent-facing-introspection)
output and for the generated error reference doc.

### Layer 2 — Propagation (`fmt.Errorf` + `%w`)

Intermediate callers wrap with context. Wrapping a sentinel preserves
both `errors.Is` identity and `errors.As` type extraction:

```go
g, err := tbx.Load(path)
if err != nil {
    return fmt.Errorf("loading %s: %w", path, err)
}
```

Discrimination uses whichever mechanism fits the call site:

- `errors.Is(err, tbx.ErrUnsupportedDialect)` — branch on identity.
- `errors.As(err, &coded)` — extract code/exit/hint at the boundary.

Both coexist; both work on the same error value.

### Layer 3 — Envelope emission (single exit point)

`main.go` is the **only** place that writes the envelope and calls
`os.Exit`. Everything else returns errors.

```go
func main() {
    cmd := app.Root()
    if err := cmd.Run(context.Background(), os.Args); err != nil {
        output.EmitError(cmd.ErrWriter, err)
        os.Exit(output.ExitCodeFor(err))
    }
}
```

`output.EmitError`:

1. `errors.As` into `terr.Coded` — extract `Code`, `Hint`, `Message`.
2. If not coded: fall through to `{code: "internal_error", exit: 1}`
   and log the raw error chain at WARN level (see
   [logging](logging.md)).
3. Under `--debug`, attach `slog.Any("stack", debug.Stack())` to that log
   line. The stack is never computed in normal runs (zero overhead) and
   never enters the envelope — the envelope stays contract-only.
4. Marshal `{ok: false, error: {code, message, hint}}` to JSON or text
   per `--format`.
5. Write to stderr.

`output.ExitCodeFor`:

1. `errors.As` into `terr.Coded` — return `e.ExitCode()`.
2. Else: return `1`.

### Layer 4 — Envelope conformance

Every emitted envelope (success or error) is exercised by golden-file
tests per command. Envelope shapes are defined by Go struct types in
`internal/output/types.go` (see
[schema-source-of-truth](schema-source-of-truth.md));
deterministic serialization plus golden files together cover the
contract. There is no embedded JSON Schema to validate against at
runtime.

See [testing](testing.md) §"Envelope
conformance".

## Recoverable vs fatal

The split is a contract, not a per-command judgment call.
Schema-conformance tests enforce it; every command spec classifies each
failure mode into one of these two buckets.

- **Recoverable** — envelope is `{ok: true, warnings: [...]}` (clean or
  warning-bearing), or `{ok: false, ...}` carrying **partial results** in
  the success-path payload (e.g. `scan` with a partially unreadable file).
  Exit code is **0** (clean) or **1** (warnings/partial). No `error`
  field is present.
- **Fatal** — envelope is `{ok: false, error: {code, message, hint}}`
  with **no result fields** outside the `error` object. Exit code is
  **≥ 2**.

## Exit code families

Exit codes follow BSD `sysexits.h` where applicable. The mapping below
is authoritative; sentinels in `internal/<pkg>/errors.go` carry the
exit code as the second argument to `terr.New(...)`. See
[`docs/cli-design.md`](cli-design.md) §Design principles #2 for the
narrative rationale.

| Code | Class                | Meaning                                                              |
| ---- | -------------------- | -------------------------------------------------------------------- |
| 0    | success              | Clean run.                                                           |
| 1    | warnings / partial   | Recoverable: warnings present or partial results returned.           |
| 2    | usage error          | CLI surface itself was misused (`no_tbx_path`, `invalid_field`).     |
| 3    | I/O error            | Filesystem failure independent of payload validity.                  |
| 65   | validation_error     | `EX_DATAERR` — payload rejected. `validate`, `apply`, write-side input rejection (`duplicate_id`, `not_found`, `dangling_crossref`, `invalid_id`, `invalid_picklist`), `schema --command UNKNOWN`. |

`no_tbx_path` and `invalid_field` remain `exit_code: 2` (usage-error
class — argv was syntactically fine but referenced something invalid in
the CLI surface itself, not in the data). Everything that rejects
payload data — file contents, write-command inputs, apply manifests —
maps to **65**.

## `urfave/cli/v3` integration

urfave has its own `cli.ExitCoder` interface:

```go
type ExitCoder interface {
    error
    ExitCode() int
}
```

Our `terr.Coded` extends it with `Code` and `Hint`. urfave's
`HandleExitCoder` is a wrapper for "call `os.Exit(err.ExitCode())`" — we
**do not** use it directly because we need to emit our envelope *before*
exiting. `main.go` handles the error path itself (layer 3 above).

## Multi-error aggregation

`validate` warnings are **not errors** — they're results. The envelope is
`{ok: true, warnings: [...]}` (recoverable path).

Cases where many failures aggregate fatally (e.g. `apply` validation
catching 5 broken concepts):

- Use `errors.Join` to collect; expose as `terr.Coded` with code
  `apply_validation_failed`.
- The envelope's `error.details.failures` array carries per-concept
  errors.

The aggregated shape is declared by the Go struct types in
`internal/output/types.go` and surfaced via
[`terminology schema --command apply`](schema-source-of-truth.md#agent-facing-introspection).

## Format-specific error rendering

Both `--format json` and `--format text` are produced from the same
`*terr.E` by the renderer in `internal/output/errors.go`.

- **JSON** — `{ok: false, error: {code, message, hint}}`. The `hint` key
  is omitted (not empty-stringed) when `Hint()` returns `""`.
- **Text** — one line per field, with the hint on a two-space-indented
  continuation line:

  ```
  ✗ command: message
    hint: ...
  ```

  If `Hint()` returns `""`, the `hint:` line is omitted. The continuation
  prefix is exactly two spaces (tabs are not used; golden files compare
  byte-for-byte).

## Logging interaction

The error envelope is the **result**. Logs are diagnostic. The two never
duplicate:

- The envelope says `{code: "no_tbx_path", message: "no TBX file path
  provided"}`.
- The log says `slog.Error("command failed", "code", "no_tbx_path",
  "command", "validate", "args", os.Args[1:])`.

The log carries operational metadata (command, args, run-id, timing);
the envelope carries the contract. Stack traces (under `--debug`) attach
to the log, never to the envelope.

See [logging](logging.md).

## Error codes registry

The authoritative registry is the set of `terr.New(...)` sentinels
across all `internal/<pkg>/errors.go` files. `internal/terr` collects
them at init time; the reflective `terminology schema` walks the
collection and the generated
[`docs/reference/errors.md`](reference/errors.md) is regenerated from
it (see
[schema-source-of-truth](schema-source-of-truth.md)).

Adding a code:

1. Add a sentinel `var ErrXxx = terr.New(code, exit, hint, message)` in
   the originating `internal/<pkg>/errors.go`.
2. Reference it from the consuming command(s).
3. Regenerate `docs/reference/errors.md` (`make docs`) and commit it.

There is no separate declarative file to update; the sentinel is the
declaration. Removing or renaming a code is a **breaking change** —
bumps `schema_version` per
[schema-source-of-truth](schema-source-of-truth.md).
