# Cross-cutting — Logging

> **Status**: APPROVED. `slog` configuration, level discipline, the
> stdout/stderr separation rule.

## Rules of engagement

1. **Logs go to stderr only.** Stdout is reserved for JSON envelopes.
   Mixing the two breaks the agent contract.
2. **Logs never duplicate envelope content.** The envelope is the
   result; logs are operational. If you find yourself logging the same
   thing the envelope says, delete the log.
3. **Default level: WARN.** Quiet by default. The envelope says
   everything a user needs.
4. **Levels:** ERROR / WARN / INFO / DEBUG. `slog`'s defaults.
5. **Handlers:** text on TTY (stderr is a TTY), JSON otherwise. JSON
   when piped means agents can parse our diagnostics too.
6. **Every record carries attributes**: `command`, `run_id`, `version`,
   optional `author`.

## Verbosity controls

Verbosity is exposed as **explicit boolean flags** with no short forms.
This sidesteps the `-v`/`--version` collision in `urfave/cli/v3` and
keeps agent scripts and golden-file assertions unambiguous.

| Flag        | Effect                | Default |
| ----------- | --------------------- | ------- |
| (default)   | WARN                  | yes     |
| `--verbose` | INFO                  |         |
| `--debug`   | DEBUG                 |         |
| `--quiet`   | ERROR (no WARN noise) |         |

The flags are mutually exclusive — passing more than one is a usage
error (`exit 2`).

## Wiring

In `internal/app/root.go`:

```go
func bootstrapLogger(cmd *cli.Command) *slog.Logger {
    level := slog.LevelWarn
    if cmd.Bool("debug") {
        level = slog.LevelDebug
    } else if cmd.Bool("verbose") {
        level = slog.LevelInfo
    } else if cmd.Bool("quiet") {
        level = slog.LevelError
    }

    var h slog.Handler
    opts := &slog.HandlerOptions{Level: level}
    if term.IsTerminal(int(os.Stderr.Fd())) {
        h = slog.NewTextHandler(os.Stderr, opts)
    } else {
        h = slog.NewJSONHandler(os.Stderr, opts)
    }

    return slog.New(h).With(
        "command", cmd.FullName(),
        "run_id",  newRunID(),
        "version", version.Current(),
    )
}
```

TTY detection uses `golang.org/x/term` — same Go-team source as
`golang.org/x/text`, which the project already depends on.

The logger is stashed on the request `context.Context` (see
[Per-package logger access](#per-package-logger-access)) so downstream
packages can retrieve it without threading `*slog.Logger` through every
signature.

## Run ID

Every record carries a `run_id` attribute correlating all log lines
emitted by a single invocation. It is **8 bytes of `crypto/rand`,
hex-encoded** (16 ASCII chars, e.g. `8f7a1e2c4b9d0a3f`).

```go
// internal/logctx/runid.go
package logctx

import (
    "crypto/rand"
    "encoding/hex"
)

func newRunID() string {
    var b [8]byte
    _, _ = rand.Read(b[:])
    return hex.EncodeToString(b[:])
}
```

64 bits of entropy is ample for correlating one-shot CLI runs; nothing
parses the value as a UUID, so RFC 4122 conformance buys nothing. No
third-party dependency.

## Per-package logger access

Code in `internal/tbx`, `internal/match`, `internal/write`, … retrieves
the logger from `context.Context` via a small `internal/logctx` helper:

```go
package logctx

import (
    "context"
    "log/slog"
)

type ctxKey struct{}

func With(ctx context.Context, l *slog.Logger) context.Context {
    return context.WithValue(ctx, ctxKey{}, l)
}

func From(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}
```

`ctx` is already threaded for cancellation; reusing it for the logger
avoids polluting every function signature. `From` falls back to
`slog.Default()` so unit tests that don't construct a context still
work.

Package-level `SetLogger`-style globals are explicitly **not used** —
global mutable state is hostile to test parallelism.

## What to log at each level

### ERROR

- Just before the binary exits non-zero with a fatal error.
- Include the error chain (`slog.Any("error", err)`).
- Under `--debug`: include `slog.Any("stack", debug.Stack())`. The stack
  attaches to the log line, never to the envelope (see
  [error-handling](error-handling.md)).

### WARN

- Recoverable issues: legacy form normalized on read, unknown element
  tolerated, dangling crossref under `--force`.
- `--transaction` set with no `--author` (per
  [E7 Q7](specs/007-write-commands.md#q7--author-resolution-precedence)).

### INFO

- Lifecycle events: TBX loaded (`concepts`, `languages`).
- Effects: dry-run reconciliation summary, files scanned.

### DEBUG

- Internal branches: which TBX dialect-style decoder fired, how many
  patterns the matcher compiled, per-file scan timings.
- The full error chain even on non-fatal paths.

## Attributes

| Key           | Value                  | When                                  |
| ------------- | ---------------------- | ------------------------------------- |
| `command`     | `terminology validate` | always                                |
| `run_id`      | 16-char hex            | always                                |
| `version`     | `v0.1.2`               | always                                |
| `author`      | `--author` value       | when write command + author set       |
| `tbx`         | resolved path          | after `--tbx` resolution              |
| `dialect`     | `TBX-Linguist`         | after load                            |
| `duration_ms` | int                    | on completion                         |

## Redaction

**No redaction in v1.** The tool runs locally; logs go to the invoking
user's own stderr. `--author`, resolved file paths, and other operational
metadata are emitted verbatim. Users who pipe stderr to a shared sink
are responsible for that choice — same posture as `git` and other local
developer tooling.

This is documented in [cli-design.md](cli-design.md) §"Logging and
diagnostics" so the assumption is explicit.

## Log destinations

Stderr is the only sink in v1. Users who want diagnostics in a file
redirect: `terminology validate --debug 2>diag.log`. A first-class
`--log-file PATH` flag is deferred to v2 — adding it requires picking a
rotation/truncation policy, deciding handler format independently of
`--format`, and specifying interaction with `--quiet`/`--debug`. Not
worth the surface area at v1.

## Sampling

Not relevant — the binary is a one-shot CLI, not a long-running server.
Every log record is emitted.
