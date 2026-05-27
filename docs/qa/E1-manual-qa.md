# Test Plan: E1 — CLI surface stub manual QA

> **Status**: ready to execute once E1 is completed
> Run end-to-end in a single sitting.

## Purpose

E1's deliverable is **the complete command surface** — every command,
subcommand, positional, flag, alias, env-var source, and exit-code
shape — with every action returning the `under_construction` envelope.
This QA pass exercises that surface against a built binary to catch
wiring drift that the unit tests might miss (urfave version quirks,
env-var lifecycle, format rendering, log destinations, exit-code
plumbing through `main.go`).

This is **surface QA, not behaviour QA**. Real validation, scanning,
checking, applying, etc. land in E2–E10. Anything that isn't the
documented `under_construction` envelope is a regression against E1's
contract.

## Scope

### In scope

- Every command and subcommand parses its documented argument shape.
- Every flag (long form + short alias) is recognized.
- Every closed picklist enum rejects unknown values **before** the
  stub fires.
- Every required positional / required flag is rejected by urfave
  when missing.
- Global flags (`--tbx`/`-T`, `--format`, `--verbose`, `--debug`,
  `--quiet`) work on every command.
- Env-var sources (`TERMINOLOGY_TBX`, `TERMINOLOGY_AUTHOR`) feed
  their respective flags.
- The verbosity mutex (`--verbose` / `--debug` / `--quiet` are
  mutually exclusive) is enforced.
- The `concept update --merge | --replace` mutex is enforced
  (exactly-one-required).
- `--version` / `-v` prints the version string injected at build time.
- `--help` / `-h` works at root and on every command/subcommand.
- Output contract: stub-paths emit exit 75 + empty stdout + JSON
  envelope on stderr.
- Envelope conformance: every error envelope carries
  `schema_version`, `ok: false`, and an `error: {code, message, hint?}`
  object.

### Out of scope

- TBX parsing / writing / validation (E2–E3).
- Markdown parsing, matching, scanning, checking, extracting (E4–E6).
- Concept-ID derivation, write semantics, transaction records (E7).
- `apply` reconciliation (E8).
- Input hardening beyond closed-enum validation, fuzz, perf (E9).
- Distribution / signing (E10).
- Performance or scalability.

If any of the above areas misbehave **at the surface level** (wrong
exit code, malformed envelope, missing flag), file it as an E1 bug.
If they misbehave at the **logic level** (e.g. `validate` doesn't
actually validate), that's expected — the stub fires.

## Environment & preconditions

- macOS, Linux, or Windows shell that can run the binary directly.
- Go toolchain installed (only required for `make build`).
- `jq` recommended for envelope assertions (some test cases use it).
- Working directory: project root.

### Pre-flight setup

```sh
# 1. Build the binary.
make build

# 2. Bind a convenience variable to the host binary.
export TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"

# 3. Smoke check: it runs.
$TT --version
$TT --help

# 4. Optional: alias for terser commands in this doc.
alias tt="$TT"
```

If any of steps 1–3 fails, **stop**: build is broken, not E1.

## Entry criteria

- All E1 tickets `ter-zwxt … ter-e31e` closed.
- `make build` exits 0 cleanly from the project root.
- `cd src && go test ./...` exits 0.
- `bin/terminology-<host>-<arch>` exists and is executable.

## Exit criteria

- Every **P0** test case passes — no exceptions.
- Every **P1** test case passes — no exceptions.
- Every **P2** test case passes, OR a follow-up ticket is filed against
  E1 with a reproducer.
- Every **P3** test case is run and recorded; failures noted but not
  blocking.

## Risk areas

| Risk                                                      | Mitigation                                                                                 |
| --------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| urfave v3 parent-without-Action exit code quirk           | Explicit cases for `concept` and `term` parents (TC-CONCEPT-PARENT-_, TC-TERM-PARENT-_)    |
| urfave v3 unknown-subcommand exit code                    | Explicit case TC-ROOT-005                                                                  |
| Verbosity mutex semantics on multi-bool flags             | Explicit table TC-ROOT-VERB-001 … TC-ROOT-VERB-007                                         |
| Env-var lifecycle (set in one sub-shell, read by another) | TC-APPLY-005 / TC-WRITE-AUTHOR-001 use a single shell line `VAR=… $TT …` to isolate        |
| `--format` enum enforcement on a global flag              | TC-ROOT-FMT-003 — `--format yaml` must be rejected before any action fires                 |
| Concept-update mutex sentinel codes                       | TC-CONCEPT-UPDATE-MUTEX-001/002 assert exact `error.code` via jq                           |
| Stale env from a previous case bleeding in                | Each case that touches env is `env -i SHELL=$SHELL VAR=… $TT …` to fully reset             |
| Cross-platform path handling for `--tbx`                  | TC-ROOT-TBX-001 uses a POSIX path; on Windows substitute backslash equivalent              |
| Whitespace in JSON envelope                               | jq-based assertions tolerate it; byte-for-byte expectations live in unit goldens, not here |

## Conventions

### Envelope shape (stub path)

```json
{
  "schema_version": 1,
  "ok": false,
  "error": {
    "code": "under_construction",
    "message": "terminology <command> is not implemented yet",
    "hint": "track progress in .tickets/ or rebuild from a newer commit"
  }
}
```

- `schema_version` is always `1`.
- `ok` is always `false` on error paths.
- `error.hint` may be omitted (not empty-string) for some sentinels.

### Exit code map (E1 surface only)

| Code | Meaning            | E1 source                                    |
| ---- | ------------------ | -------------------------------------------- |
| 0    | success            | `--help`, `--version`                        |
| 2    | usage error        | missing required, unknown flag/subcmd, mutex |
| 75   | under_construction | any valid stub invocation                    |

### How to read a test case

Each case has the form:

```
TC-<MODULE>-NNN — <name>                          P0|P1|P2|P3
$ <argv>                                          (paste-runnable)
exit=<n>
stdout: <expectation>
stderr: <expectation>
```

`exit=$?` is the convention used in shell snippets. When a case asserts
on the envelope's `error.code`, the snippet pipes stderr through `jq`.

When a single block exercises multiple sub-cases (e.g. enum rejection
across part-of-speech / register / grammatical-gender / status), they
are grouped in a table.

---

# 1. Pre-flight

## TC-PRE-001 — build cleanly P0

```sh
make build
echo "exit=$?"
```

- **exit**: `0`
- **artefact**: `bin/terminology-<host>-<arch>` exists.

## TC-PRE-002 — binary exists & boots P0

```sh
$TT --version
echo "exit=$?"
```

- **exit**: `0`
- **stdout**: non-empty string (the version) on a single line.

---

# 2. Root cross-cutting

## TC-ROOT-001 — root help (long form) P0

```sh
$TT --help; echo "exit=$?"
```

- **exit**: `0`
- **stdout**: contains `terminology`, lists subcommands
  `validate`, `lookup`, `scan`, `check`, `extract`, `apply`, `schema`,
  `concept`, `term`, and global flags `--tbx`, `--format`, `--verbose`,
  `--debug`, `--quiet`, `--help`, `--version`.

## TC-ROOT-002 — root help (short alias) P0

```sh
$TT -h; echo "exit=$?"
```

Same expectation as TC-ROOT-001.

## TC-ROOT-003 — `--version` (long) P0

```sh
$TT --version; echo "exit=$?"
```

- **exit**: `0`
- **stdout**: the version string (matches output of
  `git describe --tags --dirty --always` from project root, OR `"dev"`
  if no git tags).

## TC-ROOT-004 — `-v` (short) P0

```sh
$TT -v; echo "exit=$?"
```

Same expectation as TC-ROOT-003. `-v` is **not** `--verbose`.

## TC-ROOT-005 — bare invocation prints help and exits non-zero P0

```sh
$TT; echo "exit=$?"
```

- **exit**: `2`
- **stderr** (or **stdout**, depending on urfave): subcommand help.
- **note**: This is Q10. Exact stream may be stdout; the contract is
  exit `2` + the user sees help.

## TC-ROOT-006 — unknown subcommand P0

```sh
$TT bogus 2>&1; echo "exit=$?"
```

- **exit**: non-zero (typically `2`).
- **stderr**: contains a usage-error message mentioning `bogus`.

## TC-ROOT-007 — unknown flag at root P1

```sh
$TT --bogus-flag 2>&1; echo "exit=$?"
```

- **exit**: non-zero (`2`).
- **stderr**: urfave usage error.

## TC-ROOT-008 — `--tbx` flag accepted on every command (sanity) P1

```sh
$TT --tbx /tmp/x.tbx validate 2>&1; echo "exit=$?"
```

- **exit**: `75`
- **stderr**: under_construction envelope (the path is **not** read in
  E1; the test asserts the flag parses).

## TC-ROOT-009 — `-T` short alias for `--tbx` P1

```sh
$TT -T /tmp/x.tbx validate 2>&1; echo "exit=$?"
```

Same expectation as TC-ROOT-008.

## TC-ROOT-010 — `TERMINOLOGY_TBX` env source P1

```sh
env -i SHELL=$SHELL HOME=$HOME PATH=$PATH TERMINOLOGY_TBX=/tmp/x.tbx $TT validate 2>&1; echo "exit=$?"
```

- **exit**: `75`
- **stderr**: under_construction envelope.
- **note**: With no `--tbx` flag and no env, the surface still accepts
  the invocation (E1 does not require `--tbx`); the assertion here is
  that the env source is **wired** — verifiable when E2+ lands.

## TC-ROOT-FMT-001 — `--format json` default P0

```sh
$TT validate 2>&1 >/dev/null | jq -e '.error.code == "under_construction"'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- **stdout** (from $TT): empty.

## TC-ROOT-FMT-002 — `--format text` P1

```sh
$TT --format text validate 2>&1; echo "exit=$?"
```

- **exit**: `75`
- **stderr**: starts with `✗ ` and contains
  `terminology validate is not implemented yet`; followed by a
  continuation line `  hint: ...`.
- **stdout**: empty.

## TC-ROOT-FMT-003 — `--format yaml` rejected (closed enum) P0

```sh
$TT --format yaml validate 2>&1; echo "exit=$?"
```

- **exit**: `2`
- **stderr**: urfave validator error mentioning `format` and the
  allowed values `json`, `text`.

### Verbosity matrix

| Case ID          | argv (after `$TT`)                   | exit | envelope code           | P   |
| ---------------- | ------------------------------------ | ---- | ----------------------- | --- |
| TC-ROOT-VERB-001 | `--verbose validate`                 | 75   | `under_construction`    | P1  |
| TC-ROOT-VERB-002 | `--debug validate`                   | 75   | `under_construction`    | P1  |
| TC-ROOT-VERB-003 | `--quiet validate`                   | 75   | `under_construction`    | P1  |
| TC-ROOT-VERB-004 | `--verbose --debug validate`         | 2    | `conflicting_verbosity` | P0  |
| TC-ROOT-VERB-005 | `--verbose --quiet validate`         | 2    | `conflicting_verbosity` | P0  |
| TC-ROOT-VERB-006 | `--debug --quiet validate`           | 2    | `conflicting_verbosity` | P0  |
| TC-ROOT-VERB-007 | `--verbose --debug --quiet validate` | 2    | `conflicting_verbosity` | P0  |

For each row:

```sh
$TT <argv> 2>err.json; echo "exit=$?"
jq -e --arg c "<envelope-code>" '.error.code == $c' err.json
```

---

# 3. Per-command surface

## 3.1 `validate`

### TC-VALIDATE-001 — stub fires (basic) P0

```sh
$TT validate 2>&1; echo "exit=$?"
```

- **exit**: `75`
- **stderr**: envelope with `error.code == "under_construction"`,
  message contains `validate`.
- **stdout**: empty.

### TC-VALIDATE-002 — `--strict` P1

```sh
$TT validate --strict 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope.

### TC-VALIDATE-003 — `--fields LIST` P1

```sh
$TT validate --fields concepts,languages 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope.

### TC-VALIDATE-004 — `-F` short alias P1

```sh
$TT validate -F concepts 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope.

### TC-VALIDATE-005 — unknown flag rejected P1

```sh
$TT validate --bogus 2>&1; echo "exit=$?"
```

- **exit**: `2` (urfave usage error).

### TC-VALIDATE-006 — `validate --help` P2

```sh
$TT validate --help; echo "exit=$?"
```

- **exit**: `0`, lists `--strict`, `--fields`, `-F`.

## 3.2 `lookup`

### TC-LOOKUP-001 — stub fires with positional P0

```sh
$TT lookup tzimtzum 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope mentioning `lookup`.

### TC-LOOKUP-002 — missing positional rejected P0

```sh
$TT lookup 2>&1; echo "exit=$?"
```

- **exit**: `2` (urfave Min: 1 rejection).
- **stderr**: mentions the missing argument (`term` or similar).

### TC-LOOKUP-003 — too many positionals rejected P1

```sh
$TT lookup tzimtzum extra-arg 2>&1; echo "exit=$?"
```

- **exit**: `2` (urfave Max: 1 rejection).

### TC-LOOKUP-004 — `--lang` long P1

```sh
$TT lookup tzimtzum --lang es 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope.

### TC-LOOKUP-005 — `-l` short P1

```sh
$TT lookup tzimtzum -l es 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope.

### TC-LOOKUP-006 — `--fields` and `-F` P1

```sh
$TT lookup tzimtzum --fields definitions,notes 2>&1
$TT lookup tzimtzum -F definitions 2>&1
```

- Both: exit `75`, stub envelope.

### TC-LOOKUP-007 — `lookup --help` P2

```sh
$TT lookup --help; echo "exit=$?"
```

- **exit**: `0`, lists `TERM` positional + `--lang`/`-l` + `--fields`/`-F`.

## 3.3 `scan`

### TC-SCAN-001 — stub fires P0

```sh
$TT scan source/ch1.md 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope mentioning `scan`.

### TC-SCAN-002 — missing positional P0

```sh
$TT scan 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-SCAN-003 — too many positionals P1

```sh
$TT scan a.md b.md 2>&1; echo "exit=$?"
```

- **exit**: `2` (Max: 1).

### TC-SCAN-004 — `--lang` long & `-l` short P1

```sh
$TT scan a.md --lang es 2>&1
$TT scan a.md -l es 2>&1
```

- Both: exit `75`, stub.

### TC-SCAN-005 — `--context N` (int) P1

```sh
$TT scan a.md --context 120 2>&1; echo "exit=$?"
```

- **exit**: `75`.

### TC-SCAN-006 — `--context bogus` rejected (type) P1

```sh
$TT scan a.md --context bogus 2>&1; echo "exit=$?"
```

- **exit**: `2` (urfave int parser error).

### TC-SCAN-007 — `--fields` / `-F` P1

```sh
$TT scan a.md --fields matches.context 2>&1
$TT scan a.md -F matches.context 2>&1
```

- Both: exit `75`.

### TC-SCAN-008 — `scan --help` P2

```sh
$TT scan --help; echo "exit=$?"
```

- **exit**: `0`, lists `FILE` + `--lang/-l`, `--context`, `--fields/-F`.

## 3.4 `check`

### TC-CHECK-001 — stub fires (both positionals) P0

```sh
$TT check src.md tgt.md 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope mentioning `check`.

### TC-CHECK-002 — only one positional rejected P0

```sh
$TT check src.md 2>&1; echo "exit=$?"
```

- **exit**: `2` (Min: 2).

### TC-CHECK-003 — no positionals rejected P0

```sh
$TT check 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-CHECK-004 — too many positionals rejected P1

```sh
$TT check a.md b.md c.md 2>&1; echo "exit=$?"
```

- **exit**: `2` (Max: 2).

### TC-CHECK-005 — `--source-lang` / `-S` short P1

```sh
$TT check a.md b.md --source-lang es 2>&1
$TT check a.md b.md -S es 2>&1
```

- Both: exit `75`. Note `-S` is **uppercase**, distinct from
  `--status`'s `-s`.

### TC-CHECK-006 — `--target-lang` (no short) P1

```sh
$TT check a.md b.md --target-lang en 2>&1
```

- **exit**: `75`.

### TC-CHECK-007 — `--strict` P1

```sh
$TT check a.md b.md --strict 2>&1
```

- **exit**: `75`.

### TC-CHECK-008 — `--context` int P1

```sh
$TT check a.md b.md --context 100 2>&1
```

- **exit**: `75`.

### TC-CHECK-009 — `--fields` / `-F` P1

```sh
$TT check a.md b.md --fields violations.context 2>&1
$TT check a.md b.md -F violations 2>&1
```

- Both: exit `75`.

### TC-CHECK-010 — `check --help` P2

```sh
$TT check --help; echo "exit=$?"
```

- **exit**: `0`, lists `SRC TGT` + all flags.

## 3.5 `extract`

### TC-EXTRACT-001 — stub fires (one file) P0

```sh
$TT extract a.md 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope mentioning `extract`.

### TC-EXTRACT-002 — stub fires (variadic, many files) P1

```sh
$TT extract a.md b.md c.md d.md 2>&1; echo "exit=$?"
```

- **exit**: `75`.

### TC-EXTRACT-003 — no positionals rejected P0

```sh
$TT extract 2>&1; echo "exit=$?"
```

- **exit**: `2` (Min: 1).

### TC-EXTRACT-004 — `--exclude` / `-x` short P1

```sh
$TT extract a.md --exclude old.tbx 2>&1
$TT extract a.md -x old.tbx 2>&1
```

- Both: exit `75`.

### TC-EXTRACT-005 — `--script` valid values (table) P1

| Value      | Expected |
| ---------- | -------- |
| `latin`    | exit 75  |
| `hebrew`   | exit 75  |
| `cyrillic` | exit 75  |
| `arabic`   | exit 75  |
| `any`      | exit 75  |

```sh
for s in latin hebrew cyrillic arabic any; do
  $TT extract a.md --script "$s" >/dev/null 2>err.json
  echo "$s: exit=$? code=$(jq -r '.error.code' err.json)"
done
```

Every row should print `exit=75 code=under_construction`.

### TC-EXTRACT-006 — `--script` invalid rejected P0

```sh
$TT extract a.md --script klingon 2>&1; echo "exit=$?"
```

- **exit**: `2` (urfave closed-enum validator).

### TC-EXTRACT-007 — `--lang` / `-l` P1

```sh
$TT extract a.md --lang es 2>&1
$TT extract a.md -l es 2>&1
```

- Both: exit `75`.

### TC-EXTRACT-008 — `--stopwords PATH` P1

```sh
$TT extract a.md --stopwords stop.txt 2>&1
```

- **exit**: `75`.

### TC-EXTRACT-009 — `--min-freq` int P1

```sh
$TT extract a.md --min-freq 5 2>&1
$TT extract a.md --min-freq bogus 2>&1
```

- First: exit `75`.
- Second: exit `2` (urfave int parser).

### TC-EXTRACT-010 — `--fields` / `-F` P1

```sh
$TT extract a.md --fields candidates 2>&1
$TT extract a.md -F candidates 2>&1
```

- Both: exit `75`.

### TC-EXTRACT-011 — `extract --help` P2

```sh
$TT extract --help; echo "exit=$?"
```

- **exit**: `0`, lists `FILE...` + all flags including `--script`'s
  allowed values.

## 3.6 `apply`

### TC-APPLY-001 — `--file -` (stdin) P0

```sh
$TT apply --file - 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope mentioning `apply`.

### TC-APPLY-002 — `--file PATH` P1

```sh
$TT apply --file payload.json 2>&1; echo "exit=$?"
```

- **exit**: `75`.

### TC-APPLY-003 — `-f` short P1

```sh
$TT apply -f - 2>&1; echo "exit=$?"
```

- **exit**: `75`.

### TC-APPLY-004 — missing `--file` rejected (Required) P0

```sh
$TT apply 2>&1; echo "exit=$?"
```

- **exit**: `2` (urfave Required: true).
- **stderr**: mentions `file` or `f`.

### TC-APPLY-005 — `--prune` P1

```sh
$TT apply --file - --prune 2>&1
```

- **exit**: `75`.

### TC-APPLY-006 — `--dry-run` / `-n` P1

```sh
$TT apply --file - --dry-run 2>&1
$TT apply --file - -n 2>&1
```

- Both: exit `75`.

### TC-APPLY-007 — `--transaction` P1

```sh
$TT apply --file - --transaction 2>&1
```

- **exit**: `75`.

### TC-APPLY-008 — `--author NAME` / `-a NAME` P1

```sh
$TT apply --file - --transaction --author Andre 2>&1
$TT apply --file - --transaction -a Andre 2>&1
```

- Both: exit `75`.

### TC-APPLY-009 — `TERMINOLOGY_AUTHOR` env P1

```sh
env -i SHELL=$SHELL HOME=$HOME PATH=$PATH TERMINOLOGY_AUTHOR=Andre \
  $TT apply --file - --transaction 2>&1; echo "exit=$?"
```

- **exit**: `75`. Env is wired into urfave's Sources for `--author`.
  E1 cannot fully verify "the value flows through", but the
  invocation must not error out due to the env being set.

### TC-APPLY-010 — `apply --help` P2

```sh
$TT apply --help; echo "exit=$?"
```

- **exit**: `0`, lists `--file`/`-f` (Required), `--prune`,
  `--dry-run`/`-n`, `--transaction`, `--author`/`-a`.

## 3.7 `schema`

### TC-SCHEMA-001 — stub fires P0

```sh
$TT schema 2>&1; echo "exit=$?"
```

- **exit**: `75`, stub envelope mentioning `schema`.

### TC-SCHEMA-002 — `--command NAME` P1

```sh
$TT schema --command validate 2>&1; echo "exit=$?"
```

- **exit**: `75`. (E1 does not validate the command name; E4 does.)

### TC-SCHEMA-003 — unknown flag rejected P1

```sh
$TT schema --bogus 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-SCHEMA-004 — `schema --help` P2

```sh
$TT schema --help; echo "exit=$?"
```

- **exit**: `0`, lists `--command`.

## 3.8 `concept` parent

### TC-CONCEPT-PARENT-001 — bare parent prints help, exits non-zero P0

```sh
$TT concept; echo "exit=$?"
```

- **exit**: non-zero (typically `2` — urfave's default for parent with
  no Action and no matched subcommand).
- **stderr** or **stdout**: lists subcommands `add`, `update`, `remove`.

### TC-CONCEPT-PARENT-002 — unknown subcommand P1

```sh
$TT concept bogus 2>&1; echo "exit=$?"
```

- **exit**: non-zero.

### TC-CONCEPT-PARENT-003 — `concept --help` P2

```sh
$TT concept --help; echo "exit=$?"
```

- **exit**: `0`, lists subcommands.

## 3.9 `concept add`

### TC-CONCEPT-ADD-001 — stub fires (no flags) P0

```sh
$TT concept add 2>&1; echo "exit=$?"
```

- **exit**: `75`, envelope mentions `concept add`.

### TC-CONCEPT-ADD-002 — `--id ID` / `-i ID` P1

```sh
$TT concept add --id tzimtzum 2>&1
$TT concept add -i tzimtzum 2>&1
```

- Both: exit `75`.

### TC-CONCEPT-ADD-003 — `--subject-field` P1

```sh
$TT concept add --subject-field kabbalah 2>&1
```

- **exit**: `75`.

### TC-CONCEPT-ADD-004 — `--canonical-lang` P1

```sh
$TT concept add --canonical-lang en 2>&1
```

- **exit**: `75`.

### TC-CONCEPT-ADD-005 — `--lang/-l` + `--term/-t` P1

```sh
$TT concept add --lang es --term tzimtzum 2>&1
$TT concept add -l es -t tzimtzum 2>&1
```

- Both: exit `75`.

### TC-CONCEPT-ADD-PICK-001 — `--status` valid (modern) P1

```sh
for v in preferredTerm-admn-sts admittedTerm-admn-sts \
         deprecatedTerm-admn-sts supersededTerm-admn-sts; do
  $TT concept add --status "$v" >/dev/null 2>err.json
  echo "$v: exit=$? code=$(jq -r '.error.code' err.json)"
done
```

- All rows: `exit=75 code=under_construction`.

### TC-CONCEPT-ADD-PICK-002 — `--status` valid (legacy bare forms) P2

```sh
for v in preferredTerm admittedTerm deprecatedTerm supersededTerm; do
  $TT concept add --status "$v" >/dev/null 2>err.json
  echo "$v: exit=$? code=$(jq -r '.error.code' err.json)"
done
```

- All rows: `exit=75 code=under_construction`. Legacy bare forms are
  accepted at the urfave layer (E7 normalizes on write).

### TC-CONCEPT-ADD-PICK-003 — `--status` invalid P0

```sh
$TT concept add --status klingon 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-CONCEPT-ADD-PICK-004 — `-s` short alias for `--status` P1

```sh
$TT concept add -s preferredTerm-admn-sts 2>&1
```

- **exit**: `75`.

### TC-CONCEPT-ADD-PICK-005 — `--part-of-speech` matrix P1

| Value         | Expected |
| ------------- | -------- |
| `noun`        | exit 75  |
| `verb`        | exit 75  |
| `adjective`   | exit 75  |
| `adverb`      | exit 75  |
| `other`       | exit 75  |
| `frobnicator` | exit 2   |

```sh
for v in noun verb adjective adverb other; do
  $TT concept add --part-of-speech "$v" >/dev/null 2>&1; echo "$v: $?"
done
$TT concept add --part-of-speech frobnicator 2>&1; echo "frobnicator: $?"
```

### TC-CONCEPT-ADD-PICK-006 — `-p` short for `--part-of-speech` P1

```sh
$TT concept add -p noun 2>&1
```

- **exit**: `75`.

### TC-CONCEPT-ADD-PICK-007 — `--register` matrix P1

| Value                 | Expected               |
| --------------------- | ---------------------- |
| `colloquialRegister`  | exit 75                |
| `neutralRegister`     | exit 75                |
| `technicalRegister`   | exit 75                |
| `in-houseRegister`    | exit 75                |
| `bench-levelRegister` | exit 75                |
| `slangRegister`       | exit 75                |
| `vulgarRegister`      | exit 75                |
| `usageRegister`       | exit 75 (legacy alias) |
| `something-else`      | exit 2                 |

### TC-CONCEPT-ADD-PICK-008 — `-r` short for `--register` P1

```sh
$TT concept add -r technicalRegister 2>&1
```

- **exit**: `75`.

### TC-CONCEPT-ADD-PICK-009 — `--grammatical-gender` matrix P1

| Value        | Expected |
| ------------ | -------- |
| `masculine`  | exit 75  |
| `feminine`   | exit 75  |
| `neuter`     | exit 75  |
| `other`      | exit 75  |
| `non-binary` | exit 2   |

`--grammatical-gender` has no short alias.

### TC-CONCEPT-ADD-010 — write affordances P1

```sh
$TT concept add --dry-run 2>&1
$TT concept add -n 2>&1
$TT concept add --transaction 2>&1
$TT concept add --transaction --author Andre 2>&1
$TT concept add --transaction -a Andre 2>&1
```

- All: exit `75`.

### TC-CONCEPT-ADD-011 — `TERMINOLOGY_AUTHOR` env P1

```sh
env -i SHELL=$SHELL HOME=$HOME PATH=$PATH TERMINOLOGY_AUTHOR=Andre \
  $TT concept add --transaction 2>&1; echo "exit=$?"
```

- **exit**: `75`.

### TC-CONCEPT-ADD-012 — `concept add --help` P2

```sh
$TT concept add --help; echo "exit=$?"
```

- **exit**: `0`, lists every flag declared above.

## 3.10 `concept update`

### TC-CONCEPT-UPDATE-001 — `--merge` stub fires P0

```sh
$TT concept update tzimtzum --merge 2>&1; echo "exit=$?"
```

- **exit**: `75`, envelope mentions `concept update`.

### TC-CONCEPT-UPDATE-002 — `--replace` stub fires P0

```sh
$TT concept update tzimtzum --replace 2>&1
```

- **exit**: `75`.

### TC-CONCEPT-UPDATE-MUTEX-001 — both flags rejected P0

```sh
$TT concept update tzimtzum --merge --replace 2>err.json; echo "exit=$?"
jq -e '.error.code == "merge_replace_mutex"' err.json
```

- **exit**: `2`.
- **envelope.error.code**: `merge_replace_mutex`.

### TC-CONCEPT-UPDATE-MUTEX-002 — neither flag rejected P0

```sh
$TT concept update tzimtzum 2>err.json; echo "exit=$?"
jq -e '.error.code == "merge_replace_required"' err.json
```

- **exit**: `2`.
- **envelope.error.code**: `merge_replace_required`.

### TC-CONCEPT-UPDATE-003 — missing positional rejected P0

```sh
$TT concept update --merge 2>&1; echo "exit=$?"
```

- **exit**: `2` (urfave Min: 1 on the ID positional).

### TC-CONCEPT-UPDATE-004 — `--subject-field` P1

```sh
$TT concept update tzimtzum --merge --subject-field kabbalah 2>&1
```

- **exit**: `75`.

### TC-CONCEPT-UPDATE-005 — `--lang/-l` + `--term/-t` P1

```sh
$TT concept update tzimtzum --merge --lang es --term tzim 2>&1
$TT concept update tzimtzum --merge -l es -t tzim 2>&1
```

- Both: exit `75`.

### TC-CONCEPT-UPDATE-006 — write affordances P1

```sh
$TT concept update tzimtzum --merge --dry-run 2>&1
$TT concept update tzimtzum --merge -n 2>&1
$TT concept update tzimtzum --merge --transaction 2>&1
$TT concept update tzimtzum --merge --transaction --author Andre 2>&1
$TT concept update tzimtzum --merge --transaction -a Andre 2>&1
```

- All: exit `75`.

### TC-CONCEPT-UPDATE-007 — `concept update --help` P2

```sh
$TT concept update --help; echo "exit=$?"
```

- **exit**: `0`, lists `ID` positional, `--merge`, `--replace`,
  `--subject-field`, `--lang/-l`, `--term/-t`, `--dry-run/-n`,
  `--transaction`, `--author/-a`.

## 3.11 `concept remove`

### TC-CONCEPT-REMOVE-001 — stub fires P0

```sh
$TT concept remove tzimtzum 2>&1; echo "exit=$?"
```

- **exit**: `75`, envelope mentions `concept remove`.

### TC-CONCEPT-REMOVE-002 — missing positional rejected P0

```sh
$TT concept remove 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-CONCEPT-REMOVE-003 — `--force` P1

```sh
$TT concept remove tzimtzum --force 2>&1
```

- **exit**: `75`.

### TC-CONCEPT-REMOVE-004 — write affordances P1

```sh
$TT concept remove tzimtzum --dry-run 2>&1
$TT concept remove tzimtzum -n 2>&1
$TT concept remove tzimtzum --transaction --author Andre 2>&1
$TT concept remove tzimtzum --transaction -a Andre 2>&1
```

- All: exit `75`.

### TC-CONCEPT-REMOVE-005 — `concept remove --help` P2

```sh
$TT concept remove --help; echo "exit=$?"
```

- **exit**: `0`, lists `ID`, `--force`, `--dry-run/-n`,
  `--transaction`, `--author/-a`.

## 3.12 `term` parent

### TC-TERM-PARENT-001 — bare parent prints help, exits non-zero P0

```sh
$TT term; echo "exit=$?"
```

- **exit**: non-zero (typically `2`).
- **stderr/stdout**: lists subcommands `add`, `deprecate`.

### TC-TERM-PARENT-002 — unknown subcommand P1

```sh
$TT term bogus 2>&1; echo "exit=$?"
```

- **exit**: non-zero.

### TC-TERM-PARENT-003 — `term --help` P2

```sh
$TT term --help; echo "exit=$?"
```

- **exit**: `0`.

## 3.13 `term add`

### TC-TERM-ADD-001 — stub fires (with required flags) P0

```sh
$TT term add tzimtzum --lang es --term tzimtzum 2>&1; echo "exit=$?"
```

- **exit**: `75`, envelope mentions `term add`.

### TC-TERM-ADD-002 — short aliases for required flags P1

```sh
$TT term add tzimtzum -l es -t tzimtzum 2>&1
```

- **exit**: `75`.

### TC-TERM-ADD-003 — missing positional rejected P0

```sh
$TT term add --lang es --term tzim 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-TERM-ADD-004 — missing `--lang` rejected P0

```sh
$TT term add tzimtzum --term tzim 2>&1; echo "exit=$?"
```

- **exit**: `2` (urfave Required: true).

### TC-TERM-ADD-005 — missing `--term` rejected P0

```sh
$TT term add tzimtzum --lang es 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-TERM-ADD-006 — missing both required flags P1

```sh
$TT term add tzimtzum 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-TERM-ADD-PICK-001 — `--status` matrix P1

Same matrix as TC-CONCEPT-ADD-PICK-001 / 003: every modern + legacy
admin-status value → exit `75`; an unknown value → exit `2`.

### TC-TERM-ADD-PICK-002 — `-s` short alias P1

```sh
$TT term add tzimtzum --lang es --term tzim -s preferredTerm-admn-sts 2>&1
```

- **exit**: `75`.

### TC-TERM-ADD-PICK-003 — `--part-of-speech` matrix P1

Same matrix as TC-CONCEPT-ADD-PICK-005.

### TC-TERM-ADD-PICK-004 — `-p` short alias P1

```sh
$TT term add tzimtzum --lang es --term tzim -p noun 2>&1
```

- **exit**: `75`.

### TC-TERM-ADD-PICK-005 — `--register` matrix P1

Same matrix as TC-CONCEPT-ADD-PICK-007.

### TC-TERM-ADD-PICK-006 — `-r` short alias P1

```sh
$TT term add tzimtzum --lang es --term tzim -r technicalRegister 2>&1
```

- **exit**: `75`.

### TC-TERM-ADD-PICK-007 — `--grammatical-gender` matrix P1

Same matrix as TC-CONCEPT-ADD-PICK-009.

### TC-TERM-ADD-007 — write affordances P1

```sh
$TT term add tzimtzum --lang es --term tzim --dry-run 2>&1
$TT term add tzimtzum --lang es --term tzim -n 2>&1
$TT term add tzimtzum --lang es --term tzim --transaction 2>&1
$TT term add tzimtzum --lang es --term tzim --transaction --author Andre 2>&1
$TT term add tzimtzum --lang es --term tzim --transaction -a Andre 2>&1
```

- All: exit `75`.

### TC-TERM-ADD-008 — `TERMINOLOGY_AUTHOR` env P1

```sh
env -i SHELL=$SHELL HOME=$HOME PATH=$PATH TERMINOLOGY_AUTHOR=Andre \
  $TT term add tzimtzum --lang es --term tzim --transaction 2>&1; echo "exit=$?"
```

- **exit**: `75`.

### TC-TERM-ADD-009 — `term add --help` P2

```sh
$TT term add --help; echo "exit=$?"
```

- **exit**: `0`, lists `ID` + required `--lang/-l`, `--term/-t`,
  optional `--status/-s`, `--part-of-speech/-p`, `--register/-r`,
  `--grammatical-gender`, write affordances.

## 3.14 `term deprecate`

### TC-TERM-DEP-001 — stub fires P0

```sh
$TT term deprecate tzimtzum --lang es --term contraction 2>&1; echo "exit=$?"
```

- **exit**: `75`, envelope mentions `term deprecate`.

### TC-TERM-DEP-002 — short aliases for required flags P1

```sh
$TT term deprecate tzimtzum -l es -t contraction 2>&1
```

- **exit**: `75`.

### TC-TERM-DEP-003 — missing positional rejected P0

```sh
$TT term deprecate --lang es --term contraction 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-TERM-DEP-004 — missing `--lang` rejected P0

```sh
$TT term deprecate tzimtzum --term contraction 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-TERM-DEP-005 — missing `--term` rejected P0

```sh
$TT term deprecate tzimtzum --lang es 2>&1; echo "exit=$?"
```

- **exit**: `2`.

### TC-TERM-DEP-006 — write affordances P1

```sh
$TT term deprecate tzimtzum --lang es --term contraction --dry-run 2>&1
$TT term deprecate tzimtzum --lang es --term contraction -n 2>&1
$TT term deprecate tzimtzum --lang es --term contraction --transaction 2>&1
$TT term deprecate tzimtzum --lang es --term contraction --transaction --author Andre 2>&1
$TT term deprecate tzimtzum --lang es --term contraction --transaction -a Andre 2>&1
```

- All: exit `75`.

### TC-TERM-DEP-007 — `TERMINOLOGY_AUTHOR` env P1

```sh
env -i SHELL=$SHELL HOME=$HOME PATH=$PATH TERMINOLOGY_AUTHOR=Andre \
  $TT term deprecate tzimtzum --lang es --term contraction --transaction 2>&1
echo "exit=$?"
```

- **exit**: `75`.

### TC-TERM-DEP-008 — `term deprecate --help` P2

```sh
$TT term deprecate --help; echo "exit=$?"
```

- **exit**: `0`, lists `ID` + required `--lang/-l`, `--term/-t`, write
  affordances.

---

# 4. Output contract checks

These run **across** the surface to verify the envelope and stream
contract.

## TC-CONTRACT-001 — stub paths produce empty stdout P0

```sh
for cmd in \
  "validate" \
  "lookup tzimtzum" \
  "scan a.md" \
  "check a.md b.md" \
  "extract a.md" \
  "apply --file -" \
  "schema" \
  "concept add" \
  "concept update tzimtzum --merge" \
  "concept remove tzimtzum" \
  "term add tzimtzum --lang es --term tzim" \
  "term deprecate tzimtzum --lang es --term tzim"; do
  out=$(eval "$TT $cmd" 2>/dev/null)
  if [ -n "$out" ]; then
    echo "FAIL: '$cmd' produced stdout: $out"
  else
    echo "ok: $cmd"
  fi
done
```

Every line should print `ok: ...`. Any FAIL is a P0 regression.

## TC-CONTRACT-002 — stub paths produce a valid envelope on stderr P0

```sh
for cmd in \
  "validate" \
  "lookup tzimtzum" \
  "scan a.md" \
  "check a.md b.md" \
  "extract a.md" \
  "apply --file -" \
  "schema" \
  "concept add" \
  "concept update tzimtzum --merge" \
  "concept remove tzimtzum" \
  "term add tzimtzum --lang es --term tzim" \
  "term deprecate tzimtzum --lang es --term tzim"; do
  eval "$TT $cmd" 2>err.json >/dev/null
  ok=$(jq -e '.schema_version == 1 and .ok == false and .error.code == "under_construction" and (.error.message | test($c))' --arg c "$(echo "$cmd" | awk '{print $1}')" err.json)
  if [ "$ok" = "true" ]; then echo "ok: $cmd"; else echo "FAIL: $cmd"; fi
done
```

Every line should print `ok: ...`.

## TC-CONTRACT-003 — exit-code map P0

```sh
$TT validate;                                  echo "exit=$? want=75"
$TT --help >/dev/null;                         echo "exit=$? want=0"
$TT --version >/dev/null;                      echo "exit=$? want=0"
$TT apply 2>/dev/null;                         echo "exit=$? want=2"   # required-flag
$TT concept update tzimtzum 2>/dev/null;       echo "exit=$? want=2"   # mutex
$TT concept update tzimtzum --merge --replace 2>/dev/null;
  echo "exit=$? want=2"   # mutex
$TT --verbose --debug validate 2>/dev/null;    echo "exit=$? want=2"
$TT --format yaml validate 2>/dev/null;        echo "exit=$? want=2"
$TT lookup 2>/dev/null;                        echo "exit=$? want=2"   # missing positional
```

Every `exit=` must match its `want=`.

## TC-CONTRACT-004 — text format renders shape P1

```sh
$TT --format text validate 2>&1 >/dev/null
```

Expected stderr output (exact shape, modulo concrete strings):

```
✗ terminology validate is not implemented yet
  hint: track progress in .tickets/ or rebuild from a newer commit
```

- Single `✗ ` prefix on the message line.
- Two-space indent on the hint line.
- No trailing whitespace.

## TC-CONTRACT-005 — text format omits empty hint P2

Use `concept update tzimtzum` (which surfaces `merge_replace_required`).
If that sentinel has a hint configured, this case verifies it appears;
if it has no hint, this case verifies the continuation line is absent.

```sh
$TT --format text concept update tzimtzum 2>&1 >/dev/null
```

Expected: one line starting with `✗ `, optionally followed by one
`  hint: ...` line and nothing else.

---

# 5. Help system coverage

`--help` and `-h` must work on every command path. The expected exit
is always `0`; the expected output always contains the documented
flags and (where applicable) positionals.

Iterate the matrix below; any deviation is a P2 finding.

| Path                      | argv                    | P   |
| ------------------------- | ----------------------- | --- |
| root                      | `--help`                | P0  |
| root (short)              | `-h`                    | P0  |
| validate                  | `validate --help`       | P2  |
| lookup                    | `lookup --help`         | P2  |
| scan                      | `scan --help`           | P2  |
| check                     | `check --help`          | P2  |
| extract                   | `extract --help`        | P2  |
| apply                     | `apply --help`          | P2  |
| schema                    | `schema --help`         | P2  |
| concept parent            | `concept --help`        | P2  |
| concept add               | `concept add --help`    | P2  |
| concept update            | `concept update --help` | P2  |
| concept remove            | `concept remove --help` | P2  |
| term parent               | `term --help`           | P2  |
| term add                  | `term add --help`       | P2  |
| term deprecate            | `term deprecate --help` | P2  |
| `-h` on a deep subcommand | `concept add -h`        | P3  |

```sh
for path in \
  "validate" "lookup" "scan" "check" "extract" "apply" "schema" \
  "concept" "concept add" "concept update" "concept remove" \
  "term" "term add" "term deprecate"; do
  eval "$TT $path --help" >/dev/null 2>&1; echo "$path --help: exit=$?"
done
```

Every line must end with `exit=0`.

---

# 6. Final coverage matrix

Tick each cell as the case passes. Empty cells indicate the flag is
not available on that command (per the E1 short-alias map).

| Flag / alias            | val | lkp  | scn  | chk | ext | aply  | sch | c.add | c.upd | c.rm | t.add | t.dep |
| ----------------------- | --- | ---- | ---- | --- | --- | ----- | --- | ----- | ----- | ---- | ----- | ----- |
| (positional)            |     | TERM | FILE | 2×  | 1+  |       |     |       | ID    | ID   | ID    | ID    |
| `--tbx`/`-T` (global)   | ✓   | ✓    | ✓    | ✓   | ✓   | ✓     | ✓   | ✓     | ✓     | ✓    | ✓     | ✓     |
| `--format` (global)     | ✓   | ✓    | ✓    | ✓   | ✓   | ✓     | ✓   | ✓     | ✓     | ✓    | ✓     | ✓     |
| `--verbose` (global)    | ✓   | ✓    | ✓    | ✓   | ✓   | ✓     | ✓   | ✓     | ✓     | ✓    | ✓     | ✓     |
| `--debug` (global)      | ✓   | ✓    | ✓    | ✓   | ✓   | ✓     | ✓   | ✓     | ✓     | ✓    | ✓     | ✓     |
| `--quiet` (global)      | ✓   | ✓    | ✓    | ✓   | ✓   | ✓     | ✓   | ✓     | ✓     | ✓    | ✓     | ✓     |
| `--fields`/`-F`         | ✓   | ✓    | ✓    | ✓   | ✓   |       |     |       |       |      |       |       |
| `--strict`              | ✓   |      |      | ✓   |     |       |     |       |       |      |       |       |
| `--lang`/`-l`           |     | ✓    | ✓    |     | ✓   |       |     | ✓     | ✓     |      | ✓     | ✓     |
| `--source-lang`/`-S`    |     |      |      | ✓   |     |       |     |       |       |      |       |       |
| `--target-lang`         |     |      |      | ✓   |     |       |     |       |       |      |       |       |
| `--context`             |     |      | ✓    | ✓   |     |       |     |       |       |      |       |       |
| `--script`              |     |      |      |     | ✓   |       |     |       |       |      |       |       |
| `--stopwords`           |     |      |      |     | ✓   |       |     |       |       |      |       |       |
| `--min-freq`            |     |      |      |     | ✓   |       |     |       |       |      |       |       |
| `--exclude`/`-x`        |     |      |      |     | ✓   |       |     |       |       |      |       |       |
| `--file`/`-f`           |     |      |      |     |     | ✓ (R) |     |       |       |      |       |       |
| `--prune`               |     |      |      |     |     | ✓     |     |       |       |      |       |       |
| `--command`             |     |      |      |     |     |       | ✓   |       |       |      |       |       |
| `--id`/`-i`             |     |      |      |     |     |       |     | ✓     |       |      |       |       |
| `--subject-field`       |     |      |      |     |     |       |     | ✓     | ✓     |      |       |       |
| `--canonical-lang`      |     |      |      |     |     |       |     | ✓     |       |      |       |       |
| `--term`/`-t`           |     |      |      |     |     |       |     | ✓     | ✓     |      | ✓ (R) | ✓ (R) |
| `--status`/`-s`         |     |      |      |     |     |       |     | ✓     |       |      | ✓     |       |
| `--part-of-speech`/`-p` |     |      |      |     |     |       |     | ✓     |       |      | ✓     |       |
| `--register`/`-r`       |     |      |      |     |     |       |     | ✓     |       |      | ✓     |       |
| `--grammatical-gender`  |     |      |      |     |     |       |     | ✓     |       |      | ✓     |       |
| `--merge`               |     |      |      |     |     |       |     |       | ✓ (M) |      |       |       |
| `--replace`             |     |      |      |     |     |       |     |       | ✓ (M) |      |       |       |
| `--force`               |     |      |      |     |     |       |     |       |       | ✓    |       |       |
| `--dry-run`/`-n`        |     |      |      |     |     | ✓     |     | ✓     | ✓     | ✓    | ✓     | ✓     |
| `--transaction`         |     |      |      |     |     | ✓     |     | ✓     | ✓     | ✓    | ✓     | ✓     |
| `--author`/`-a` (env)   |     |      |      |     |     | ✓     |     | ✓     | ✓     | ✓    | ✓     | ✓     |

Legend: `(R)` = required, `(M)` = mutex pair.

---

# 7. Sign-off checklist

Tick before declaring E1 manual QA pass.

- [ ] Section 1 (pre-flight): both cases pass.
- [ ] Section 2 (root cross-cutting): all P0 cases pass.
- [ ] Section 3 (per-command):
  - [ ] `validate` — every P0 + P1 case passes.
  - [ ] `lookup` — every P0 + P1 case passes.
  - [ ] `scan` — every P0 + P1 case passes.
  - [ ] `check` — every P0 + P1 case passes.
  - [ ] `extract` — every P0 + P1 case passes; both script matrices pass.
  - [ ] `apply` — every P0 + P1 case passes; env-source case passes.
  - [ ] `schema` — every P0 + P1 case passes.
  - [ ] `concept` parent — every P0 + P1 case passes.
  - [ ] `concept add` — every P0 + P1 case passes; every picklist matrix passes.
  - [ ] `concept update` — every P0 case passes; both mutex sentinels surface correctly.
  - [ ] `concept remove` — every P0 + P1 case passes.
  - [ ] `term` parent — every P0 + P1 case passes.
  - [ ] `term add` — every P0 + P1 case passes; every picklist matrix passes.
  - [ ] `term deprecate` — every P0 + P1 case passes.
- [ ] Section 4 (output contract): all four contract cases pass.
- [ ] Section 5 (help system): every command's `--help` returns exit 0.
- [ ] Section 6 (coverage matrix): every ticked cell verified at least once.
- [ ] No undocumented behaviour observed (any new exit code, envelope
      field, or rendered text not described above is filed as a finding).

## On failure

Any failing case must be classified:

- **Surface drift** — wiring problem in the urfave declaration or
  helper. Reopen the relevant E1 ticket and add a regression test
  (golden file) to its `internal/app/testdata/`.
- **Contract violation** — envelope shape, exit code, or stream
  assignment is wrong. File a new ticket tagged
  `e1,bug,blocker` linked to `ter-qxrg` and the offending command's
  ticket.
- **Spec ambiguity** — the test plan and the spec disagree. Update
  the spec **first** (PR), then re-derive the test case from the
  updated spec.

Do **not** silence a failing case by updating the test plan to match
buggy behaviour. The plan is derived from
[`docs/specs/001-cli-surface-stub.md`](../docs/specs/001-cli-surface-stub.md);
the spec wins.
