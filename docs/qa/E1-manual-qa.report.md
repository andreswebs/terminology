# E1 Manual QA Report — CLI surface stub

> **Date**: 2026-05-25
> **Tester**: Claude (automated execution of manual QA plan)
> **Binary**: `terminology` version `6183118-dirty`
> **Platform**: macOS darwin-arm64
> **Verdict**: **PASS**

## Summary

Full E1 QA plan executed. All P0 and P1 test cases pass. All P2 and P3
cases pass. No blocking findings.

**Re-test note**: An earlier run against version `f227ce1` failed with
two systemic issues (urfave errors exiting 1 instead of 2, and positional
bounds not enforced). Both have been fixed. Additionally, E2 and E3 are
now fully implemented, so `validate` no longer returns the
`under_construction` stub — those test cases are marked N/A.

## Environment

- macOS (Darwin 25.5.0, arm64)
- Go toolchain present
- `jq` available
- Binary built via `make build` — exit 0

## Section 1: Pre-flight

| Case | Result | Notes |
| --- | --- | --- |
| TC-PRE-001 | **PASS** | `make build` exits 0, binary produced |
| TC-PRE-002 | **PASS** | `terminology version 6183118-dirty`, exit 0 |

## Section 2: Root cross-cutting

| Case | Result | Notes |
| --- | --- | --- |
| TC-ROOT-001 | **PASS** | `--help` exit 0, lists all commands + global flags |
| TC-ROOT-002 | **PASS** | `-h` same as TC-ROOT-001 |
| TC-ROOT-003 | **PASS** | `--version` exit 0 |
| TC-ROOT-004 | **PASS** | `-v` exit 0, same as `--version` |
| TC-ROOT-005 | **PASS** | Bare invocation exit 2, `no_subcommand` envelope |
| TC-ROOT-006 | **PASS** | Exit 2, `unknown_subcommand`, message mentions `bogus` |
| TC-ROOT-007 | **PASS** | Exit 2, unknown flag rejected |
| TC-ROOT-008 | **N/A** | E2+ TBX validation intercepts (exit 65 `validation_error` — file doesn't exist). Flag parses. |
| TC-ROOT-009 | **N/A** | Same as TC-ROOT-008 for `-T` |
| TC-ROOT-010 | **N/A** | Same — `TERMINOLOGY_TBX` env is wired (validation fires) |
| TC-ROOT-FMT-001 | **N/A** | `validate` now requires `--tbx` (E2+), returns `no_tbx_path` not `under_construction` |
| TC-ROOT-FMT-002 | **PASS** | `--format text` works; renders `✗` prefix + hint correctly; content is `no_tbx_path` rather than stub |
| TC-ROOT-FMT-003 | **PASS** | `--format yaml` rejected, exit 2, `invalid_value` |

### Verbosity matrix

| Case | Result | Notes |
| --- | --- | --- |
| TC-ROOT-VERB-001 | **PASS** | `--verbose lookup tzimtzum` exit 75, `under_construction` |
| TC-ROOT-VERB-002 | **PASS** | `--debug lookup tzimtzum` exit 75, `under_construction` |
| TC-ROOT-VERB-003 | **PASS** | `--quiet lookup tzimtzum` exit 75, `under_construction` |
| TC-ROOT-VERB-004 | **PASS** | Exit 2, `conflicting_verbosity` |
| TC-ROOT-VERB-005 | **PASS** | Exit 2, `conflicting_verbosity` |
| TC-ROOT-VERB-006 | **PASS** | Exit 2, `conflicting_verbosity` |
| TC-ROOT-VERB-007 | **PASS** | Exit 2, `conflicting_verbosity` |

**Note on VERB-001/002/003**: The QA plan uses `validate` as the test
command, but `validate` now requires `--tbx` (E2+). Verified with
`lookup tzimtzum` instead, which is still a stub.

## Section 3: Per-command surface

### 3.1 `validate`

| Case | Result | Notes |
| --- | --- | --- |
| TC-VALIDATE-001 | **N/A** | E2+ requires `--tbx`; returns `no_tbx_path` exit 2 (not `under_construction` exit 75) |
| TC-VALIDATE-002 | **N/A** | Same — `--strict` accepted but `no_tbx_path` fires |
| TC-VALIDATE-003 | **N/A** | Same |
| TC-VALIDATE-004 | **N/A** | Same |
| TC-VALIDATE-005 | **PASS** | Unknown flag rejected, exit 2 |
| TC-VALIDATE-006 | **PASS** | `--help` exit 0 |

### 3.2 `lookup`

| Case | Result | Notes |
| --- | --- | --- |
| TC-LOOKUP-001 | **PASS** | Exit 75, `under_construction` |
| TC-LOOKUP-002 | **PASS** | Missing positional rejected, exit 2, `missing_argument` |
| TC-LOOKUP-003 | **PASS** | Too many positionals rejected, exit 2, `excess_arguments` |
| TC-LOOKUP-004 | **PASS** | `--lang` exit 75 |
| TC-LOOKUP-005 | **PASS** | `-l` exit 75 |
| TC-LOOKUP-006 | **PASS** | `--fields` / `-F` exit 75 |
| TC-LOOKUP-007 | **PASS** | `--help` exit 0 |

### 3.3 `scan`

| Case | Result | Notes |
| --- | --- | --- |
| TC-SCAN-001 | **PASS** | Exit 75, `under_construction` |
| TC-SCAN-002 | **PASS** | Missing positional rejected, exit 2, `missing_argument` |
| TC-SCAN-003 | **PASS** | Too many positionals rejected, exit 2, `excess_arguments` |
| TC-SCAN-004 | **PASS** | `--lang` / `-l` exit 75 |
| TC-SCAN-005 | **PASS** | `--context 120` exit 75 |
| TC-SCAN-006 | **PASS** | `--context bogus` rejected, exit 2 |
| TC-SCAN-007 | **PASS** | `--fields` / `-F` exit 75 |
| TC-SCAN-008 | **PASS** | `--help` exit 0 |

### 3.4 `check`

| Case | Result | Notes |
| --- | --- | --- |
| TC-CHECK-001 | **PASS** | Exit 75, `under_construction` |
| TC-CHECK-002 | **PASS** | One positional rejected, exit 2, `missing_argument` |
| TC-CHECK-003 | **PASS** | No positionals rejected, exit 2, `missing_argument` |
| TC-CHECK-004 | **PASS** | Too many positionals rejected, exit 2, `excess_arguments` |
| TC-CHECK-005 | **PASS** | `--source-lang` / `-S` exit 75 |
| TC-CHECK-006 | **PASS** | `--target-lang` exit 75 |
| TC-CHECK-007 | **PASS** | `--strict` exit 75 |
| TC-CHECK-008 | **PASS** | `--context` exit 75 |
| TC-CHECK-009 | **PASS** | `--fields` / `-F` exit 75 |
| TC-CHECK-010 | **PASS** | `--help` exit 0 |

### 3.5 `extract`

| Case | Result | Notes |
| --- | --- | --- |
| TC-EXTRACT-001 | **PASS** | Exit 75, `under_construction` |
| TC-EXTRACT-002 | **PASS** | Variadic exit 75 |
| TC-EXTRACT-003 | **PASS** | No positionals rejected, exit 2, `missing_argument` |
| TC-EXTRACT-004 | **PASS** | `--exclude` / `-x` exit 75 |
| TC-EXTRACT-005 | **PASS** | All valid `--script` values exit 75 |
| TC-EXTRACT-006 | **PASS** | Invalid `--script klingon` rejected, exit 2 |
| TC-EXTRACT-007 | **PASS** | `--lang` / `-l` exit 75 |
| TC-EXTRACT-008 | **PASS** | `--stopwords` exit 75 |
| TC-EXTRACT-009 | **PASS** | Valid int exit 75; bogus int rejected exit 2 |
| TC-EXTRACT-010 | **PASS** | `--fields` / `-F` exit 75 |
| TC-EXTRACT-011 | **PASS** | `--help` exit 0 |

### 3.6 `apply`

| Case | Result | Notes |
| --- | --- | --- |
| TC-APPLY-001 | **PASS** | `--file -` exit 75, `under_construction` |
| TC-APPLY-002 | **PASS** | `--file PATH` exit 75 |
| TC-APPLY-003 | **PASS** | `-f` exit 75 |
| TC-APPLY-004 | **PASS** | Missing `--file` rejected, exit 2, `missing_required_flag` |
| TC-APPLY-005 | **PASS** | `--prune` exit 75 |
| TC-APPLY-006 | **PASS** | `--dry-run` / `-n` exit 75 |
| TC-APPLY-007 | **PASS** | `--transaction` exit 75 |
| TC-APPLY-008 | **PASS** | `--author` / `-a` exit 75 |
| TC-APPLY-009 | **PASS** | `TERMINOLOGY_AUTHOR` env exit 75 |
| TC-APPLY-010 | **PASS** | `--help` exit 0 |

### 3.7 `schema`

| Case | Result | Notes |
| --- | --- | --- |
| TC-SCHEMA-001 | **PASS** | Exit 75, `under_construction` |
| TC-SCHEMA-002 | **PASS** | `--command` exit 75 |
| TC-SCHEMA-003 | **PASS** | Unknown flag rejected, exit 2 |
| TC-SCHEMA-004 | **PASS** | `--help` exit 0 |

### 3.8 `concept` parent

| Case | Result | Notes |
| --- | --- | --- |
| TC-CONCEPT-PARENT-001 | **PASS** | Bare parent exit 2 |
| TC-CONCEPT-PARENT-002 | **PASS** | Unknown subcommand exit 2 |
| TC-CONCEPT-PARENT-003 | **PASS** | `--help` exit 0 |

### 3.9 `concept add`

| Case | Result | Notes |
| --- | --- | --- |
| TC-CONCEPT-ADD-001 | **PASS** | Exit 75, `under_construction` |
| TC-CONCEPT-ADD-002 | **PASS** | `--id` / `-i` exit 75 |
| TC-CONCEPT-ADD-003 | **PASS** | `--subject-field` exit 75 |
| TC-CONCEPT-ADD-004 | **PASS** | `--canonical-lang` exit 75 |
| TC-CONCEPT-ADD-005 | **PASS** | `--lang/-l` + `--term/-t` exit 75 |
| TC-CONCEPT-ADD-PICK-001 | **PASS** | All modern `--status` values exit 75 |
| TC-CONCEPT-ADD-PICK-002 | **PASS** | All legacy bare `--status` values exit 75 |
| TC-CONCEPT-ADD-PICK-003 | **PASS** | Invalid `--status klingon` rejected, exit 2 |
| TC-CONCEPT-ADD-PICK-004 | **PASS** | `-s` short exit 75 |
| TC-CONCEPT-ADD-PICK-005 | **PASS** | All valid part-of-speech exit 75; `frobnicator` rejected exit 2 |
| TC-CONCEPT-ADD-PICK-006 | **PASS** | `-p` short exit 75 |
| TC-CONCEPT-ADD-PICK-007 | **PASS** | All valid registers exit 75; `something-else` rejected exit 2 |
| TC-CONCEPT-ADD-PICK-008 | **PASS** | `-r` short exit 75 |
| TC-CONCEPT-ADD-PICK-009 | **PASS** | All valid genders exit 75; `non-binary` rejected exit 2 |
| TC-CONCEPT-ADD-010 | **PASS** | All write affordances exit 75 |
| TC-CONCEPT-ADD-011 | **PASS** | `TERMINOLOGY_AUTHOR` env exit 75 |
| TC-CONCEPT-ADD-012 | **PASS** | `--help` exit 0 |

### 3.10 `concept update`

| Case | Result | Notes |
| --- | --- | --- |
| TC-CONCEPT-UPDATE-001 | **PASS** | `--merge` exit 75, `under_construction` |
| TC-CONCEPT-UPDATE-002 | **PASS** | `--replace` exit 75 |
| TC-CONCEPT-UPDATE-MUTEX-001 | **PASS** | Both flags: exit 2, `merge_replace_mutex` |
| TC-CONCEPT-UPDATE-MUTEX-002 | **PASS** | Neither flag: exit 2, `merge_replace_required` |
| TC-CONCEPT-UPDATE-003 | **PASS** | Missing positional rejected, exit 2, `missing_argument` |
| TC-CONCEPT-UPDATE-004 | **PASS** | `--subject-field` exit 75 |
| TC-CONCEPT-UPDATE-005 | **PASS** | `--lang/-l` + `--term/-t` exit 75 |
| TC-CONCEPT-UPDATE-006 | **PASS** | All write affordances exit 75 |
| TC-CONCEPT-UPDATE-007 | **PASS** | `--help` exit 0 |

### 3.11 `concept remove`

| Case | Result | Notes |
| --- | --- | --- |
| TC-CONCEPT-REMOVE-001 | **PASS** | Exit 75, `under_construction` |
| TC-CONCEPT-REMOVE-002 | **PASS** | Missing positional rejected, exit 2, `missing_argument` |
| TC-CONCEPT-REMOVE-003 | **PASS** | `--force` exit 75 |
| TC-CONCEPT-REMOVE-004 | **PASS** | All write affordances exit 75 |
| TC-CONCEPT-REMOVE-005 | **PASS** | `--help` exit 0 |

### 3.12 `term` parent

| Case | Result | Notes |
| --- | --- | --- |
| TC-TERM-PARENT-001 | **PASS** | Bare parent exit 2 |
| TC-TERM-PARENT-002 | **PASS** | Unknown subcommand exit 2 |
| TC-TERM-PARENT-003 | **PASS** | `--help` exit 0 |

### 3.13 `term add`

| Case | Result | Notes |
| --- | --- | --- |
| TC-TERM-ADD-001 | **PASS** | Exit 75, `under_construction` |
| TC-TERM-ADD-002 | **PASS** | Short aliases exit 75 |
| TC-TERM-ADD-003 | **PASS** | Missing positional rejected, exit 2, `missing_argument` |
| TC-TERM-ADD-004 | **PASS** | Missing `--lang` rejected, exit 2, `missing_required_flag` |
| TC-TERM-ADD-005 | **PASS** | Missing `--term` rejected, exit 2, `missing_required_flag` |
| TC-TERM-ADD-006 | **PASS** | Missing both required flags rejected, exit 2 |
| TC-TERM-ADD-PICK-001 | **PASS** | All modern + legacy `--status` values exit 75; invalid rejected exit 2 |
| TC-TERM-ADD-PICK-002 | **PASS** | `-s` short exit 75 |
| TC-TERM-ADD-PICK-003 | **PASS** | All part-of-speech values exit 75; invalid rejected exit 2 |
| TC-TERM-ADD-PICK-004 | **PASS** | `-p` short exit 75 |
| TC-TERM-ADD-PICK-005 | **PASS** | All register values exit 75; invalid rejected exit 2 |
| TC-TERM-ADD-PICK-006 | **PASS** | `-r` short exit 75 |
| TC-TERM-ADD-PICK-007 | **PASS** | All gender values exit 75; invalid rejected exit 2 |
| TC-TERM-ADD-007 | **PASS** | All write affordances exit 75 |
| TC-TERM-ADD-008 | **PASS** | `TERMINOLOGY_AUTHOR` env exit 75 |
| TC-TERM-ADD-009 | **PASS** | `--help` exit 0 |

### 3.14 `term deprecate`

| Case | Result | Notes |
| --- | --- | --- |
| TC-TERM-DEP-001 | **PASS** | Exit 75, `under_construction` |
| TC-TERM-DEP-002 | **PASS** | Short aliases exit 75 |
| TC-TERM-DEP-003 | **PASS** | Missing positional rejected, exit 2, `missing_argument` |
| TC-TERM-DEP-004 | **PASS** | Missing `--lang` rejected, exit 2, `missing_required_flag` |
| TC-TERM-DEP-005 | **PASS** | Missing `--term` rejected, exit 2, `missing_required_flag` |
| TC-TERM-DEP-006 | **PASS** | All write affordances exit 75 |
| TC-TERM-DEP-007 | **PASS** | `TERMINOLOGY_AUTHOR` env exit 75 |
| TC-TERM-DEP-008 | **PASS** | `--help` exit 0 |

## Section 4: Output contract

| Case | Result | Notes |
| --- | --- | --- |
| TC-CONTRACT-001 | **PASS** | All stub paths produce empty stdout |
| TC-CONTRACT-002 | **PASS** | All stub paths produce valid envelope on stderr (`schema_version=1`, `ok=false`, `under_construction`) |
| TC-CONTRACT-003 | **PASS** | All exit codes match: `--help`/`--version` exit 0; stubs exit 75; usage errors exit 2 |
| TC-CONTRACT-004 | **PASS** | Text format renders `✗` prefix + `  hint:` continuation line |
| TC-CONTRACT-005 | **PASS** | `concept update tzimtzum` in text mode: `✗` + `  hint:` — two lines, correct shape |


## Section 5: Help system

**PASS** — All 16 help paths (`--help` and `-h`) exit 0 on every command
and subcommand, including deep subcommands (`concept add`, `concept update`,
`concept remove`, `term add`, `term deprecate`).


## Findings

None. All previously reported findings have been resolved:

- **Finding 1 (urfave exit codes)**: Fixed — all urfave-origin errors
  now exit 2 with appropriate error codes (`missing_argument`,
  `excess_arguments`, `missing_required_flag`, `invalid_value`).
- **Finding 2 (positional bounds)**: Fixed — `Min`/`Max` on arguments
  are now enforced on all commands.
- **Finding 3 (validate requires --tbx)**: Still applies as expected
  post-E2 behavior. `validate` test cases that expected the
  `under_construction` stub are marked N/A.
- **Finding 4 (bogus subcommand naming)**: Fixed — `$TT bogus` now
  returns `unknown_subcommand` with message `"unknown subcommand: bogus"`.

## Test plan errata

**TC-ROOT-VERB-001/002/003** — The plan uses `validate` as the test
command, but `validate` now requires `--tbx` (E2+). Verified with
`lookup tzimtzum` instead.

## Sign-off checklist

- [x] Section 1 (pre-flight): both cases pass.
- [x] Section 2 (root cross-cutting): all P0 cases pass.
- [x] Section 3 (per-command):
  - [x] `validate` — N/A for stub tests (E2+); unknown flag + help pass.
  - [x] `lookup` — all P0 + P1 cases pass.
  - [x] `scan` — all P0 + P1 cases pass.
  - [x] `check` — all P0 + P1 cases pass.
  - [x] `extract` — all P0 + P1 cases pass; both script matrices pass.
  - [x] `apply` — all P0 + P1 cases pass; env-source case passes.
  - [x] `schema` — all P0 + P1 cases pass.
  - [x] `concept` parent — all P0 + P1 cases pass.
  - [x] `concept add` — all P0 + P1 cases pass; every picklist matrix passes.
  - [x] `concept update` — all P0 cases pass; both mutex sentinels surface correctly.
  - [x] `concept remove` — all P0 + P1 cases pass.
  - [x] `term` parent — all P0 + P1 cases pass.
  - [x] `term add` — all P0 + P1 cases pass; every picklist matrix passes.
  - [x] `term deprecate` — all P0 + P1 cases pass.
- [x] Section 4 (output contract): all five contract cases pass.
- [x] Section 5 (help system): all `--help` paths exit 0.
- [x] Section 6 (coverage matrix): every ticked cell verified at least once.
- [x] No undocumented behaviour observed.
