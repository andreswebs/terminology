# E3 Manual QA Report — terminology validate

> **Date**: 2026-05-25
> **Tester**: Claude (automated execution of manual QA plan)
> **Binary**: `terminology` version `1b95435`
> **Platform**: macOS darwin-arm64
> **Verdict**: **PASS**

## Summary

38 test cases executed across 10 sections (plus 2 traceability aliases).
All 38 passed.

**Re-test note**: An earlier run against version `6d9bb2f` failed 4 cases
(1 P0, 1 P1, 2 P2). Bug tickets ter-9dho and ter-97c1 were filed. Both
bugs have been fixed; this report reflects the re-test against `1b95435`.

## Environment

- macOS (Darwin 25.5.0, arm64)
- Go toolchain present
- `jq` available
- Binary built via `make build` — exit 0

## Results by section

### 1. Tier-1 — Well-formedness

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-T1-001 | P0 | PASS | Exit 65, `validation_error`, malformed XML rejected |
| TC-T1-002 | P0 | PASS | Exit 65, `validation_error`, empty file rejected |
| TC-T1-003 | P0 | PASS | Exit 65, `validation_error`, missing `<text><body>` rejected |
| TC-T1-004 | P1 | PASS | Tier-1 failure → stdout empty |

### 2. Tier-2 — Dialect/schema checks

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-PICK-001 | P0 | PASS | `invalid_picklist` warning on `bogusStatus`, exit 1 |
| TC-PICK-002 | P0 | PASS | Clean fixture: no picklist warnings |
| TC-PICK-003 | P1 | PASS | Legacy admin status forms accepted (no false positives) |
| TC-UNK-001 | P0 | PASS | `unknown_element` surfaced under `--strict` |
| TC-UNK-002 | P0 | PASS | `unknown_element` suppressed in lenient mode |

### 3. Tier-3 — Semantic checks

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-DUP-001 | P0 | PASS | `duplicate_id` warning present |
| TC-DUP-002 | P0 | PASS | `concepts == 2` (raw count, not deduplicated) |
| TC-DUP-003 | P1 | PASS | `duplicate_id` carries `concept_id == "c1"` |
| TC-LANG-001 | P0 | PASS | `invalid_lang_tag` warning on bad tag, exit 1 |
| TC-LANG-002 | P1 | PASS | Valid BCP 47 tags: no `invalid_lang_tag` |
| TC-LANG-003 | P1 | PASS | `invalid_lang_tag` carries `concept_id == "kabbalah"` |
| TC-TERM-001 | P0 | PASS | `missing_term` warning on empty langSec |
| TC-XREF-001 | P0 | PASS | `unresolved_crossref` warning in lenient mode |
| TC-XREF-002 | P0 | PASS | `--strict` promotes to error, exit 65 |
| TC-XREF-003 | P1 | PASS | Clean fixture: no `unresolved_crossref` |

### 4. `--strict` promotions

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-STRICT-001 | P0 | PASS | Clean file passes strict, exit 0 |
| TC-STRICT-002 | P0 | PASS | `unresolved_crossref` → error, exit 65 |
| TC-STRICT-003 | P0 | PASS | (alias for TC-UNK-001) `unknown_element` surfaced |
| TC-STRICT-004 | P0 | PASS | (alias for TC-UNK-002) `unknown_element` suppressed |
| TC-STRICT-005 | P1 | PASS | `legacy_form_normalized` appears in strict |
| TC-STRICT-006 | P1 | PASS | `legacy_form_normalized` suppressed in lenient |

### 5. Warning shape & line/column tracking

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-SHAPE-001 | P0 | PASS | Every warning has `code` and `message` |
| TC-SHAPE-002 | P1 | PASS | `concept_id` fields are non-empty |
| TC-LINE-001 | P1 | PASS | `unresolved_crossref` has `line == 12` |
| TC-LINE-002 | P2 | PASS | `unresolved_crossref` has `column == 29` |
| TC-LINE-003 | P2 | PASS | Two warnings have distinct line values (`12`, `22`) |

### 6. Warning codes — completeness

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-CODES-001 | P1 | PASS | Found: `duplicate_id`, `legacy_form_normalized`, `unresolved_crossref` — all spec-defined |

**Note on TC-CODES-001**: The test plan's shell loop has a quoting issue
where `for code in ${ALL_CODES}` fails to iterate when `ALL_CODES` is
set via command substitution in a single compound block. The corrected
approach uses a `while read` loop. This is a test-plan ergonomic issue,
not a code bug. All observed codes are spec-defined.

### 7. Envelope fidelity

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-ENV-001 | P0 | PASS | `ok == true`, `warnings == []`, `concepts >= 1`, exit 0 |
| TC-ENV-002 | P0 | PASS | `ok == true`, `warnings` non-empty, exit 1 |
| TC-ENV-003 | P0 | PASS | `languages == (languages | sort)` — sorted ASCII |
| TC-ENV-004 | P1 | PASS | `schema_version == 1` |
| TC-ENV-005 | P1 | PASS | `languages` and `warnings` are arrays, never null |

### 8. Exit codes

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-EXIT-001 | P0 | PASS | Exit 0 for clean file |
| TC-EXIT-002 | P0 | PASS | Exit 1 for warnings |
| TC-EXIT-003 | P0 | PASS | Exit 65 for malformed XML |
| TC-EXIT-004 | P0 | PASS | Exit 65 for strict errors |
| TC-EXIT-005 | P0 | PASS | Exit 2 for missing TBX path |

### 9. Tier sequencing

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-SEQ-001 | P0 | PASS | Tier-1 failure → stdout empty, stderr has error |
| TC-SEQ-002 | P1 | PASS | Both `invalid_picklist` (tier-2) and `duplicate_id` (tier-3) appear |

### 10. Stream routing

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-STREAM-001 | P0 | PASS | Success → stdout only, stderr empty |
| TC-STREAM-002 | P0 | PASS | Error → stderr only, stdout empty |
| TC-STREAM-003 | P1 | PASS | Warnings envelope → stdout, exit 1 |

## Findings

None. All previously reported findings have been resolved:

- **Finding 1 (TC-T1-003, P0)**: Fixed — missing `<text><body>` now
  rejected at tier-1 with exit 65. See ticket ter-9dho.
- **Findings 2–4 (TC-LINE-001/002/003, P1/P2)**: Fixed — `line` and
  `column` fields now populated on all warnings, including model-level
  warnings from `Glossary.Validate()`. See ticket ter-97c1.

## Test plan errata

**TC-CODES-001** — The shell loop `for code in ${ALL_CODES}` does not
iterate correctly when `ALL_CODES` is set via command substitution in a
compound brace block. A `while IFS= read -r code` loop over the output
works correctly. All observed codes are spec-defined. This is a
test-plan ergonomic issue, not a code bug.

## Sign-off checklist

- [x] Section 1 (tier-1 well-formedness): all cases pass.
- [x] Section 2 (tier-2 dialect checks): all cases pass.
- [x] Section 3 (tier-3 semantic checks): all cases pass.
- [x] Section 4 (`--strict` promotions): all cases pass.
- [x] Section 5 (warning shape & line/col): all cases pass.
- [x] Section 6 (warning codes): all codes are spec-defined.
- [x] Section 7 (envelope fidelity): all shapes correct.
- [x] Section 8 (exit codes): 0/1/2/65 all correct.
- [x] Section 9 (tier sequencing): tier-1 short-circuits; tiers 2+3 aggregate.
- [x] Section 10 (stream routing): success → stdout, errors → stderr.
- [x] No undocumented behaviour observed.
