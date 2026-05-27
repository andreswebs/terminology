# terminology

A CLI for agent-driven, terminology-focused academic translation. Reads
markdown source, enforces consistent terminology against a TBX-Linguist
glossary, and exposes a small set of deterministic operations as subcommands.

The tool is **agent-first, human-tolerant**: JSON on stdout by default,
meaningful exit codes, structured error envelopes, dry-run on writes, schema
introspection via `terminology schema`.

## Where things live

- [`docs/cli-design.md`](docs/cli-design.md) — full CLI specification, TBX
  dialect, command surface, exit codes. Reference for `Usage` strings and
  dialect details. **Note**: when in conflict with a spec or ADR, the
  spec/ADR wins (cli-design.md is the narrative; specs are the contract).
- [`docs/specs/`](docs/specs/) — epic specs `E1`–`E10`
  ([`001-cli-surface-stub.md`](docs/specs/001-cli-surface-stub.md) …
  [`010-release.md`](docs/specs/010-release.md)). **Authoritative** for the
  surface (flags, aliases, env sources, exit codes, behaviour) the
  corresponding tickets implement. Each ticket body cites the relevant
  spec sections.
- [`docs/adr/`](docs/adr/) — cross-cutting decisions referenced by every
  epic: [`error-handling.md`](docs/adr/error-handling.md),
  [`logging.md`](docs/adr/logging.md),
  [`testing.md`](docs/adr/testing.md),
  [`determinism.md`](docs/adr/determinism.md),
  [`schema-source-of-truth.md`](docs/adr/schema-source-of-truth.md).
- [`docs/specs/target.schema.json`](docs/specs/target.schema.json) —
  **obsolete scaffolding artifact**, slated for deletion. Go code is the
  canonical source of truth; see
  [`docs/adr/schema-source-of-truth.md`](docs/adr/schema-source-of-truth.md).
- [`docs/examples/scenarios.md`](docs/examples/scenarios.md) — end-to-end
  agent usage scenarios.
- [`docs/research/`](docs/research/) — external references (e.g. `urfave-cli`, TBX-Linguist and TBX-standard
  dialect notes).
- Go source: `src/`. Entry point [`src/cmd/terminology/main.go`](src/cmd/terminology/main.go);
  all logic lives under `src/internal/` (deliberately unimportable as a
  public library).

## Ticket Workflow

Tickets are managed with the `tk` CLI. Files live in `.tickets/` (created
lazily on first use).

```bash
tk ls                    # List all tickets
tk ready                 # List tickets with all deps resolved (ready to start)
tk blocked               # List tickets blocked by unresolved deps
tk show <id>             # Show full ticket details
tk dep tree <id>         # Show dependency tree
tk start <id>            # Mark as in_progress
tk close <id>            # Mark as closed
tk add-note <id> "..."   # Append a note
```

### QA gates

Every open epic (E2–E10) has two manual gate tickets:

- **Entry gate** — blocks all task tickets. Must be closed by a human
  before any work can start on the epic.
- **Exit gate** — blocks the epic itself. Depends on all task tickets in
  the epic, so it only becomes "ready" after every task is closed. Must
  be closed by a human after reviewing the completed work.

A **sentinel ticket** (`ter-go3r`) blocks all entry gates. Since the
sentinel is never closed, no epic can start until a human explicitly
releases it.

**Lifecycle of an epic:**

```txt
sentinel blocks entry gate
  → human undeps + closes entry gate
    → tasks become ready
      → agents work on tasks
        → all tasks closed
          → exit gate becomes ready
            → human reviews + closes exit gate
              → epic can be closed
```

**To unlock an epic for work:**

```bash
ENTRY_GATE="ter-6z5g"           # the epic's entry gate ticket ID
SENTINEL_ID="ter-go3r"

tk undep "${ENTRY_GATE}" "${SENTINEL_ID}"  # remove sentinel dep
tk close "${ENTRY_GATE}"                   # tasks become ready
```

**After all tasks are done, review and close:**

```bash
EXIT_GATE="ter-v203"            # the epic's exit gate ticket ID
EPIC_ID="ter-uqyn"

# exit gate is now "ready" (all tasks closed)
# ... do manual QA review ...
tk close "${EXIT_GATE}"                    # approve the work
tk close "${EPIC_ID}"                      # close the epic
```

**Gate ticket IDs:**

| Entry Gate | Exit Gate  | Epic                        |
| ---------- | ---------- | --------------------------- |
| `ter-6z5g` | `ter-v203` | E2 — Domain model & TBX I/O |
| `ter-bedf` | `ter-19rb` | E3 — terminology validate   |
| `ter-ab56` | `ter-m9l4` | E4 — Read commands          |
| `ter-c4ra` | `ter-cfeu` | E5 — Matcher                |
| `ter-ppn9` | `ter-b3ug` | E6 — scan & check           |
| `ter-st7u` | `ter-lyrx` | E7 — Write commands         |
| `ter-6dsf` | `ter-pbt2` | E8 — terminology apply      |
| `ter-hv9n` | `ter-uj1y` | E9 — Hardening              |
| `ter-39xj` | `ter-de52` | E10 — Release               |

**Important:** Do not close the sentinel (`ter-go3r`). Do not close any
gate ticket programmatically — only a human should do so after review.
When creating new task tickets for an epic, add the exit gate as a dep:
`tk dep <exit-gate> <new-task>`.

## Build & Validation

All commands are run from the project root via `make`. Go source lives under
`src/`.

| Command          | Purpose                                                                                                             |
| ---------------- | ------------------------------------------------------------------------------------------------------------------- |
| `make build`     | Full build — runs fmt-check, vet, lint, test, clean-local, then compiles to `bin/terminology-<host_os>-<host_arch>` |
| `make build-all` | Cross-compile for all supported platforms                                                                           |
| `make dist`      | Package cross-platform archives into `dist/` with `SHA256SUMS.txt`                                                  |
| `make run`       | Run the CLI directly with `go run`                                                                                  |
| `make test`      | Run all tests (`go test ./...`)                                                                                     |
| `make test-race` | Run tests with the race detector                                                                                    |
| `make vet`       | Run `go vet ./...`                                                                                                  |
| `make fmt`       | Format all Go source with `gofmt -w`                                                                                |
| `make fmt-check` | Fail if any files are not formatted                                                                                 |
| `make lint`      | Run `golangci-lint run ./...`                                                                                       |
| `make validate`  | `fmt-check` + `vet` + `lint` + `test` (the quality gate run by `build`)                                             |
| `make clean`     | Remove build artifacts from `bin/` and `dist/`                                                                      |

### Validating your work

After any code change, **always run `make build`** from the project root
before considering the task complete. It enforces the full quality gate
(`fmt-check`, `vet`, `lint`, `test`) and then compiles.

If `make build` fails, fix the underlying issue. Do not silence lint errors
with `_ =` — handle them properly (log, return, or report in tests).

For a quicker feedback loop during development, use `make test` or
`make lint` individually, but always finish with a full `make build`.
