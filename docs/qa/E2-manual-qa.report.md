# E2 Manual QA Report — Domain model & TBX I/O

> **Date**: 2026-05-25
> **Tester**: Claude (automated execution of manual QA plan)
> **Binary**: `terminology` version `d40e637`
> **Platform**: macOS darwin-arm64
> **Verdict**: **PASS**

## Summary

All 45 test cases across 12 sections passed. No findings.

## Environment

- macOS (Darwin 25.5.0, arm64)
- Go toolchain present
- `jq` available
- Binary built via `make build` — exit 0

## Results by section

### 1. Pre-flight

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-PRE-001 | P0 | PASS | `make build` exits 0 |
| TC-PRE-002 | P0 | PASS | `terminology version d40e637` |
| TC-PRE-003 | P0 | PASS | All 8 fixtures present |

### 2. TBX loading — DCT style

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-LOAD-DCT-001 | P0 | PASS | 1 concept, `["en","he"]`, no warnings |
| TC-LOAD-DCT-002 | P0 | PASS | 2 concepts, `["en","es","he"]`, sorted |
| TC-LOAD-DCT-003 | P1 | PASS | full-features loads clean |
| TC-LOAD-DCT-004 | P1 | PASS | with-transactions loads clean |
| TC-LOAD-DCT-005 | P1 | PASS | all-categories-dct loads (warnings expected separately) |

### 3. TBX loading — DCA style

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-LOAD-DCA-001 | P0 | PASS | 1 concept, `["en","he"]`, no warnings |
| TC-LOAD-DCA-002 | P1 | PASS | all-categories-dca loads |
| TC-LOAD-DCA-003 | P0 | PASS | DCT and DCA produce identical model |

### 4. Legacy-form normalization

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-LEGACY-001 | P0 | PASS | 1 concept, `["en"]`, loads clean |
| TC-LEGACY-002 | P1 | PASS | No warnings, exit 0 |

### 5. Error sentinels

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-ERR-001 | P0 | PASS | `no_tbx_path`, exit 2, hint present |
| TC-ERR-002 | P0 | PASS | `validation_error`, exit 65 |
| TC-ERR-003 | P0 | PASS | `validation_error`, exit 65 (empty file) |
| TC-ERR-004 | P0 | PASS | `validation_error`, exit 65 (TBX-Basic) |
| TC-ERR-005 | P1 | PASS | `validation_error`, exit 65 (non-TBX XML) |
| TC-ERR-006 | P1 | PASS | `validation_error`, exit 65 (malformed XML) |
| TC-ERR-007 | P1 | PASS | `validation_error`, exit 65 (no type attr) |
| TC-ERR-008 | P2 | PASS | `validation_error`, exit 65 (directory) |

### 6. Environment variable source

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-ENV-001 | P0 | PASS | `TERMINOLOGY_TBX` feeds `--tbx` |
| TC-ENV-002 | P1 | PASS | `--tbx` flag overrides env var |

### 7. Reader warnings

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-WARN-001 | P0 | PASS | `unresolved_crossref` in warnings array |
| TC-WARN-002 | P0 | PASS | Exit 1 when warnings present |
| TC-WARN-003 | P1 | PASS | Warning has `code` and `message` |
| TC-WARN-004 | P2 | PASS | Warning has non-empty `concept_id` |

### 8. `--strict` flag

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-STRICT-001 | P0 | PASS | Exit 65, `validation_error` |
| TC-STRICT-002 | P1 | PASS | Clean file passes strict, exit 0 |

### 9. Output stream routing

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-STREAM-001 | P0 | PASS | Success → stdout, stderr empty |
| TC-STREAM-002 | P0 | PASS | Error → stderr, stdout empty |
| TC-STREAM-003 | P1 | PASS | Warning envelope → stdout, exit 1 |

### 10. Envelope conformance

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-ENVELOPE-001 | P0 | PASS | All required keys present |
| TC-ENVELOPE-002 | P0 | PASS | `schema_version == 1` |
| TC-ENVELOPE-003 | P1 | PASS | `languages` is array |
| TC-ENVELOPE-004 | P1 | PASS | `warnings` is array |
| TC-ENVELOPE-005 | P0 | PASS | Error envelope has `schema_version`, `ok`, `error.code`, `error.message` |
| TC-ENVELOPE-006 | P1 | PASS | `no_tbx_path` hint is non-empty |

**Note on TC-ENVELOPE-005**: The jq expression in the test plan has a
precedence issue (`and .error | has(...)` pipes the result of `and` into
`.error`, losing context). The corrected expression
`(.error | has("code") and has("message"))` confirms the envelope is
conformant. This is a test-plan typo, not a code bug.

### 11. Text format rendering

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-TEXT-001 | P1 | PASS | `✗ no TBX file path provided` + `hint: set --tbx or TERMINOLOGY_TBX` |
| TC-TEXT-002 | P2 | PASS | `✗ TBX validation failed` + `hint: check the TBX file structure and content` |

### 12. All fixtures matrix

| Fixture | Result |
| --- | --- |
| minimal-dct.tbx | ok |
| minimal-dca.tbx | ok |
| rich-dct.tbx | ok |
| full-features.tbx | ok |
| with-transactions.tbx | ok |
| all-categories-dct.tbx | ok |
| all-categories-dca.tbx | ok |
| legacy-forms.tbx | ok |

All 8 fixtures load successfully.

## Findings

None.

## Test plan errata

**TC-ENVELOPE-005** — The jq assertion
`has("schema_version") and has("ok") and has("error") and .error | has("code") and has("message")`
has an operator precedence bug: the bare `|` after `.error` pipes the
boolean result of the `and` chain (not the object) into `has("code")`.
The corrected form is:
`has("schema_version") and has("ok") and has("error") and (.error | has("code") and has("message"))`.
The underlying envelope is correct.

## Sign-off checklist

- [x] Section 1 (pre-flight): all cases pass.
- [x] Section 2 (DCT loading): all P0 + P1 cases pass.
- [x] Section 3 (DCA loading): all P0 + P1 cases pass; DCT≡DCA check passes.
- [x] Section 4 (legacy normalization): loads without errors or warnings.
- [x] Section 5 (error sentinels): all P0 + P1 + P2 cases pass.
- [x] Section 6 (env var): `TERMINOLOGY_TBX` feeds `--tbx`; flag overrides env.
- [x] Section 7 (reader warnings): warnings surface in envelope; warnings → exit 1.
- [x] Section 8 (`--strict`): promotes warnings to errors (exit 65); clean files pass.
- [x] Section 9 (stream routing): success → stdout, errors → stderr.
- [x] Section 10 (envelope conformance): all shapes correct.
- [x] Section 11 (text format): renders `✗` + `hint:` correctly.
- [x] Section 12 (all fixtures matrix): every fixture loads successfully.
- [x] No undocumented behaviour observed.
