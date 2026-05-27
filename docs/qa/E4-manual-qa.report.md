# E4 Manual QA Report — Read commands

> **Date**: 2026-05-26
> **Tester**: Claude (automated execution of manual QA plan)
> **Binary**: `terminology` version `cc45821`
> **Platform**: macOS darwin-arm64
> **Verdict**: **PASS**

## Summary

46 test cases executed across 18 sections. 45 passed, 1 N/A, 0 failed.

**Re-test history**: The initial run found 5 failures. Three were caused
by the test plan using incorrect `--lang` and `--fields` semantics
(fixed in ter-58h6). The remaining 2 were code bugs — `exit_codes`
missing from per-command schema view (ter-ix1i) and foreign-script
heuristic ignoring frontmatter language (ter-vri7). Both are now fixed
and verified.

## Environment

- macOS (Darwin 25.5.0, arm64)
- Go toolchain present
- `jq` available
- Binary built via `make build` — exit 0 (after fixing `reflect.Ptr` →
  `reflect.Pointer` lint issues in `schema.go` and `fields.go`)

## Results by section

### 1. Lookup — basic matching

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-LKP-001 | P0 | PASS | Exit 0, ok true, results non-empty, schema_version 1 |
| TC-LKP-002 | P0 | PASS | Exit 0, case-insensitive match works (TZIMTZUM → tzimtzum) |
| TC-LKP-003 | P1 | N/A | "café" not in fixture — NFC normalization applied but no match to verify against |
| TC-LKP-004 | P0 | PASS | Exit 1, ok true, results [] |
| TC-LKP-005 | P0 | PASS | Exit 2, `no_tbx_path` |

### 2. Lookup — --lang filter

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-LANG-LKP-001 | P1 | PASS | Exit 0, Hebrew term matched within `he` section; cross-language check confirms English term returns empty with `--lang he` |
| TC-LANG-LKP-002 | P1 | PASS | Exit 1, results [] when term doesn't exist in French |

### 3. Lookup — envelope shape

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-LKP-ENV-001 | P0 | PASS | Keys: `["ok","results","schema_version"]` |
| TC-LKP-ENV-002 | P1 | PASS | Result has `concept_id`, `languages`; language entry has `preferred` |

### 4. --fields projection

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-FLD-001 | P1 | PASS | `--fields results.concept_id` projects correctly; `languages` absent |
| TC-FLD-002 | P0 | PASS | Exit 2, `invalid_field` with hint mentioning `schema` |
| TC-FLD-003 | P2 | PASS | `--fields results.languages.*.preferred.term` — wildcard traverses map keys |

### 5. Schema — full output

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-SCH-001 | P0 | PASS | Has `schema_version`, `commands`, `envelopes`, `error_codes`; version 1 |
| TC-SCH-002 | P0 | PASS | Commands: apply, check, concept, extract, help, lookup, scan, schema, term, validate |
| TC-SCH-003 | P1 | PASS | Contains `validation_error`, `no_tbx_path`, `invalid_field` |
| TC-SCH-004 | P1 | PASS | Envelopes has `validate`, `lookup`, `extract` |
| TC-SCH-005 | P0 | PASS | Exit 0 without `--tbx` |

### 6. Schema — --command filter

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-SCH-CMD-001 | P1 | PASS | name=validate, has flags, has envelope |
| TC-SCH-CMD-002 | P2 | PASS | `--strict` in flags list (also `fields`, `help`) |
| TC-SCH-CMD-003 | P0 | PASS | Exit 2, `unknown_command` |

### 7. Extract — capitalized phrases

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-CAP-001 | P0 | PASS | Holy Temple and Dead Sea Scrolls found |
| TC-CAP-002 | P2 | PASS | Bare "The" not present |
| TC-CAP-003 | P1 | PASS | Holy Temple frequency=2 |

### 8. Extract — foreign-script tokens

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-FST-001 | P0 | PASS | צמצום detected as foreign_script candidate |
| TC-FST-002 | P1 | PASS | `--script hebrew` returns only Hebrew candidates (`["צמצום"]`) |
| TC-FST-003 | P2 | PASS | `--script latin` excludes all Hebrew tokens |

### 9. Extract — high-frequency tokens

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-FREQ-001 | P1 | PASS | `--min-freq 10` filters all high_frequency candidates; 5 non-frequency candidates remain |
| TC-FREQ-002 | P1 | PASS | Stopwords excluded from high_frequency candidates |
| TC-FREQ-003 | P2 | PASS | All high_frequency candidates have frequency >= 3 |

### 10. Extract — markdown awareness

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-EXT-MD-001 | P0 | PASS | `getUserById` excluded (fenced code block), Holy Temple found |
| TC-EXT-MD-002 | P1 | PASS | Inline code `getUserById` excluded |

### 11. Extract — --exclude glossary terms

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-EXCL-001 | P1 | PASS | צמצום excluded after `--exclude minimal-dct.tbx` |
| TC-EXCL-002 | P2 | PASS | Holy Temple still present after exclude |

### 12. Extract — language detection

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-LANG-DET-001 | P1 | PASS | Exit 0, ok true; Latin tokens correctly flagged as foreign_script with `lang: he` frontmatter |
| TC-LANG-DET-002 | P2 | PASS | Exit 0, ok true with `--lang es` |
| TC-LANG-DET-003 | P2 | PASS | Exit 0, ok true (default en) |

### 13. Extract — envelope shape

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-EXT-ENV-001 | P0 | PASS | Has `schema_version`, `ok`, `candidates`; version 1, ok true |
| TC-EXT-ENV-002 | P1 | PASS | Candidate has `term`, `frequency`, `heuristic` |
| TC-EXT-ENV-003 | P0 | PASS | Exit 1, ok true, candidates [] |

### 14. Extract — exit codes

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-EXT-EXIT-001 | P0 | PASS | Exit 3 (`io_error`) — non-zero as required |
| TC-EXT-EXIT-002 | P0 | PASS | Exit 2, `missing_argument` |

### 15. --fields on extract

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-FLD-EXT-001 | P2 | PASS | `--fields candidates.term,candidates.frequency` projects correctly; `heuristic` absent |

### 16. Schema — error code detail

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-SCH-ERR-001 | P1 | PASS | Error code entry has `code`, `exit_code`, `message` |
| TC-SCH-ERR-002 | P2 | PASS | `exit_codes` present: `[0,1,2,3,65]` |

### 17. Stream routing

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-STREAM-001 | P0 | PASS | stdout non-empty, stderr empty |
| TC-STREAM-002 | P0 | PASS | stdout empty, stderr non-empty |

### 18. Regression — validate

| Case | Priority | Result | Notes |
| --- | --- | --- | --- |
| TC-REG-001 | P0 | PASS | Exit 0, ok true, schema_version 1 |

## Findings

No findings. All previously reported issues (ter-ix1i, ter-vri7) have
been fixed and verified.

## Test plan errata

**TC-LKP-003** — The fixture `all-categories-dct.tbx` does not contain a
term with combining characters (e.g. "café"). NFC normalization is
applied (verified by code inspection and unit tests), but this manual
test case cannot verify it without a suitable fixture. Marked N/A.

**hebrew-frontmatter.md setup** — The original test plan had a typo in
the corpus setup (`## cat >` instead of `cat >`, and bare `## lang: he`
instead of YAML frontmatter). Fixed in ter-58h6.

**TC-LANG-LKP-001** — Original test assumed `--lang` filters output;
actual behavior is that `--lang` restricts the search scope. Test plan
updated to use the Hebrew term and verify cross-language scoping. Fixed
in ter-58h6.

**TC-FLD-001, TC-FLD-003, TC-FLD-EXT-001** — Original tests used bare
result-level paths (`concept_id`, `term,frequency`). Implementation
requires envelope-relative paths (`results.concept_id`,
`candidates.term,candidates.frequency`). Test plan updated. Fixed in
ter-58h6.

## Sign-off checklist

- [x] Section 1 (lookup basic): all P0 pass; TC-LKP-003 N/A (no fixture).
- [x] Section 2 (lookup --lang): all cases pass.
- [x] Section 3 (lookup envelope): all cases pass.
- [x] Section 4 (--fields): all cases pass.
- [x] Section 5 (schema full): all cases pass.
- [x] Section 6 (schema --command): all cases pass.
- [x] Section 7 (capitalized phrases): all cases pass.
- [x] Section 8 (foreign-script): all cases pass.
- [x] Section 9 (high-frequency): all cases pass.
- [x] Section 10 (markdown awareness): all cases pass.
- [x] Section 11 (--exclude): all cases pass.
- [x] Section 12 (language detection): all cases pass.
- [x] Section 13 (extract envelope): all cases pass.
- [x] Section 14 (extract exit codes): all cases pass.
- [x] Section 15 (--fields on extract): all cases pass.
- [x] Section 16 (schema error detail): all cases pass.
- [x] Section 17 (stream routing): all cases pass.
- [x] Section 18 (regression): validate still works.
