# E7 Manual QA Report — Write commands

> **Date**: 2026-05-26
> **Tester**: Claude (automated execution of manual QA plan)
> **Binary**: `terminology` version `125885a`
> **Platform**: macOS darwin-arm64
> **Verdict**: **PASS**

## Summary

67 test cases executed across 20 sections. 66 passed, 1 skipped
(TC-LOCK-001: file locking uses fcntl advisory locks, untestable on
macOS without `flock(1)`). No code bugs found.

## Environment

- macOS (Darwin 25.5.0, arm64)
- Go toolchain present
- `jq` available
- Binary built via `make build` — exit 0 (version `125885a`)

## Results by section

### 1. concept add — flag input

| Case       | Priority | Result | Notes                                                              |
| ---------- | -------- | ------ | ------------------------------------------------------------------ |
| TC-ADD-001 | P0       | PASS   | Exit 0, ok true, concept_id "tikkun", subject_field "kabbalah"     |
| TC-ADD-002 | P1       | PASS   | Exit 0, picklist values (noun, technicalRegister) accepted         |
| TC-ADD-003 | P0       | PASS   | Exit 0, lookup confirms concept persisted to TBX file              |

### 2. concept add — ID derivation

| Case          | Priority | Result | Notes                                                         |
| ------------- | -------- | ------ | ------------------------------------------------------------- |
| TC-ADD-ID-001 | P0       | PASS   | Exit 0, "Divine Light" derived as "divine-light"              |
| TC-ADD-ID-002 | P1       | PASS   | Exit 0, --canonical-lang es derives "contraccion-primordial" via JSON stdin |
| TC-ADD-ID-003 | P1       | PASS   | Exit 65, Hebrew-only "אור" produces invalid_id               |

### 3. concept add — JSON stdin input

| Case            | Priority | Result | Notes                                                         |
| --------------- | -------- | ------ | ------------------------------------------------------------- |
| TC-ADD-JSON-001 | P0       | PASS   | Exit 0, concept "ein-sof" created from JSON stdin, both langs |
| TC-ADD-JSON-002 | P1       | PASS   | Exit 65, unknown field "unknown_field" produces invalid_input |

### 4. concept add — TBX fragment stdin

| Case           | Priority | Result | Notes                                                     |
| -------------- | -------- | ------ | --------------------------------------------------------- |
| TC-ADD-TBX-001 | P0       | PASS   | Exit 0, bare conceptEntry auto-detected and persisted     |
| TC-ADD-TBX-002 | P1       | PASS   | Exit 0, conceptEntryList wrapper auto-detected            |
| TC-ADD-TBX-003 | P0       | PASS   | Exit 65, full tbx document rejected with invalid_input    |

### 5. concept update — merge semantics

| Case             | Priority | Result | Notes                                                      |
| ---------------- | -------- | ------ | ---------------------------------------------------------- |
| TC-UPD-MERGE-001 | P0       | PASS   | Exit 0, French langSec added, en/es/he preserved          |
| TC-UPD-MERGE-002 | P0       | PASS   | Exit 0, English terms untouched by French-only merge       |
| TC-UPD-MERGE-003 | P1       | PASS   | Exit 0, existing term matched by (Surface, status), fields overlaid |
| TC-UPD-MERGE-004 | P2       | PASS   | Exit 0, TBX fragment auto-detected for update merge       |
| TC-UPD-MERGE-005 | P1       | PASS   | Exit 0, subject_field updated to "mysticism", langs kept   |

### 6. concept update — replace semantics

| Case            | Priority | Result | Notes                                                      |
| --------------- | -------- | ------ | ---------------------------------------------------------- |
| TC-UPD-REPL-001 | P0       | PASS   | Exit 0, content replaced; only en remains, es/he removed   |
| TC-UPD-REPL-002 | P0       | PASS   | Exit 0, concept_id remains "tzimtzum" despite replacement  |

### 7. concept update — merge/replace mutex

| Case              | Priority | Result | Notes                                                  |
| ----------------- | -------- | ------ | ------------------------------------------------------ |
| TC-UPD-MUTEX-001  | P0       | PASS   | Exit 2, merge_replace_mutex error                      |
| TC-UPD-MUTEX-002  | P0       | PASS   | Exit 2, merge_replace_required error                   |

### 8. concept remove — basic

| Case      | Priority | Result | Notes                                                   |
| --------- | -------- | ------ | ------------------------------------------------------- |
| TC-RM-001 | P0       | PASS   | Exit 0, razon-historica removed; lookup returns exit 1  |
| TC-RM-002 | P0       | PASS   | Exit 65, not_found for nonexistent concept              |

### 9. concept remove — dangling crossref

| Case           | Priority | Result | Notes                                                       |
| -------------- | -------- | ------ | ----------------------------------------------------------- |
| TC-RM-XREF-001 | P0       | PASS   | Exit 65, dangling_crossref; tzimtzum still in file          |
| TC-RM-XREF-002 | P0       | PASS   | Exit 0, --force removes tzimtzum despite inbound crossref   |
| TC-RM-XREF-003 | P1       | PASS   | validate exit 1, ok true, unresolved_crossref warning found |

### 10. term add

| Case            | Priority | Result | Notes                                                  |
| --------------- | -------- | ------ | ------------------------------------------------------ |
| TC-TERM-ADD-001 | P0       | PASS   | Exit 0, "divine withdrawal" added to en langSec        |
| TC-TERM-ADD-002 | P0       | PASS   | Exit 0, French langSec created automatically           |
| TC-TERM-ADD-003 | P0       | PASS   | Exit 65, not_found for nonexistent concept              |
| TC-TERM-ADD-004 | P1       | PASS   | Exit 2, invalid_value for "badstatus"                   |

### 11. term deprecate

| Case            | Priority | Result | Notes                                                        |
| --------------- | -------- | ------ | ------------------------------------------------------------ |
| TC-TERM-DEP-001 | P0       | PASS   | Exit 0, sephirah moved to deprecated in output              |
| TC-TERM-DEP-002 | P0       | PASS   | Exit 65, not_found for nonexistent term                      |
| TC-TERM-DEP-003 | P1       | PASS   | Exit 65, not_found for nonexistent langSec (fr on tzimtzum)  |
| TC-TERM-DEP-004 | P1       | PASS   | Exit 65, not_found for nonexistent concept                   |

### 12. Dry-run

| Case       | Priority | Result | Notes                                                     |
| ---------- | -------- | ------ | --------------------------------------------------------- |
| TC-DRY-001 | P0       | PASS   | Exit 0, ok true, concept_id "tikkun" in output            |
| TC-DRY-002 | P0       | PASS   | File checksum identical; lookup confirms not persisted     |
| TC-DRY-003 | P1       | PASS   | Exit 0, file unchanged after dry-run concept update       |
| TC-DRY-004 | P2       | PASS   | Exit 0, file unchanged after dry-run concept remove       |
| TC-DRY-005 | P1       | PASS   | Exit 65, dry-run catches duplicate_id                     |

### 13. Transaction records

| Case       | Priority | Result | Notes                                                            |
| ---------- | -------- | ------ | ---------------------------------------------------------------- |
| TC-TXN-001 | P0       | PASS   | Exit 0, transacGrp in TBX file with modification type and author |
| TC-TXN-002 | P0       | PASS   | Exit 0, no responsibility in file, WARN on stderr                |
| TC-TXN-003 | P1       | PASS   | Exit 0, transacGrp inside termSec (line 57), not concept level   |
| TC-TXN-004 | P1       | PASS   | Exit 0, TERMINOLOGY_AUTHOR="Env Author" resolved in file        |
| TC-TXN-005 | P1       | PASS   | Exit 0, no transacGrp when --transaction not set                 |

### 14. ID stability

| Case        | Priority | Result | Notes                                                           |
| ----------- | -------- | ------ | --------------------------------------------------------------- |
| TC-STAB-001 | P0       | PASS   | Exit 0, concept_id remains "tzimtzum" after preferred rename    |
| TC-STAB-002 | P0       | PASS   | Exit 0, concept_id remains "tzimtzum" after term add preferred  |

### 15. Pre-write validation

| Case          | Priority | Result | Notes                                                         |
| ------------- | -------- | ------ | ------------------------------------------------------------- |
| TC-PREVAL-001 | P0       | PASS   | Exit 65, duplicate_id; validate confirms file untouched       |
| TC-PREVAL-002 | P1       | PASS   | Exit 65, validation_error; crossref in payload caught before write |

### 16. Concurrency — file locking

| Case        | Priority | Result | Notes                                                            |
| ----------- | -------- | ------ | ---------------------------------------------------------------- |
| TC-LOCK-001 | P1       | SKIP   | Lock is fcntl-based; flock(1) unavailable on macOS               |

### 17. Envelope shape

| Case       | Priority | Result | Notes                                                          |
| ---------- | -------- | ------ | -------------------------------------------------------------- |
| TC-ENV-001 | P0       | PASS   | Top-level: schema_version (1), ok (true), result (object)      |
| TC-ENV-002 | P0       | PASS   | result has concept_id, subject_field, languages                 |
| TC-ENV-003 | P0       | PASS   | stderr: ok false, error.code "duplicate_id", error.message set |

### 18. Error cases

| Case       | Priority | Result | Notes                                                  |
| ---------- | -------- | ------ | ------------------------------------------------------ |
| TC-ERR-001 | P0       | PASS   | Exit 65, duplicate_id                                  |
| TC-ERR-002 | P0       | PASS   | Exit 65, not_found on concept update                   |
| TC-ERR-003 | P0       | PASS   | Exit 65, not_found on concept remove                   |
| TC-ERR-004 | P1       | PASS   | Exit 65, invalid_id for empty derivation               |
| TC-ERR-005 | P0       | PASS   | Exit 65, invalid_input for malformed JSON              |
| TC-ERR-006 | P0       | PASS   | Exit 2, no_tbx_path                                   |
| TC-ERR-007 | P1       | PASS   | Exit 0, TERMINOLOGY_TBX env var resolves path          |

### 19. Stream routing

| Case          | Priority | Result | Notes                                              |
| ------------- | -------- | ------ | -------------------------------------------------- |
| TC-STREAM-001 | P0       | PASS   | stdout non-empty, stderr empty on success          |
| TC-STREAM-002 | P0       | PASS   | stdout empty, stderr non-empty on error (exit 65)  |
| TC-STREAM-003 | P1       | PASS   | stdout non-empty, stderr empty on dry-run          |

### 20. Regression — previous commands still work

| Case       | Priority | Result | Notes                                                            |
| ---------- | -------- | ------ | ---------------------------------------------------------------- |
| TC-REG-001 | P0       | PASS   | Exit 0, validate ok true, schema_version 1                      |
| TC-REG-002 | P0       | PASS   | Exit 0, lookup finds tzimtzum                                   |
| TC-REG-003 | P0       | PASS   | Exit 0, scan finds matches                                      |
| TC-REG-004 | P1       | PASS   | Exit 0, extract works                                           |
| TC-REG-005 | P0       | PASS   | Exit 0, check ok true                                           |
| TC-REG-006 | P1       | PASS   | Schema has concept (add/update/remove) and term (add/deprecate)  |

## Findings

No open findings. All test cases pass or are skipped with justification.

**TC-LOCK-001 (SKIP)**: File locking uses `fcntl` advisory locks. The
`flock(1)` utility required to simulate a held lock is not available on
macOS (it is a Linux `util-linux` tool). This test should be executed on
Linux or via a dedicated Go test binary that holds the lock.

## Sign-off checklist

- [x] Section 1 (concept add — flag input): all pass.
- [x] Section 2 (concept add — ID derivation): all pass.
- [x] Section 3 (concept add — JSON stdin): all pass.
- [x] Section 4 (concept add — TBX fragment): all pass.
- [x] Section 5 (concept update — merge): all pass.
- [x] Section 6 (concept update — replace): all pass.
- [x] Section 7 (concept update — mutex): all pass.
- [x] Section 8 (concept remove — basic): all pass.
- [x] Section 9 (concept remove — crossref): all pass.
- [x] Section 10 (term add): all pass.
- [x] Section 11 (term deprecate): all pass.
- [x] Section 12 (dry-run): all pass.
- [x] Section 13 (transaction records): all pass.
- [x] Section 14 (ID stability): all pass.
- [x] Section 15 (pre-write validation): all pass.
- [x] Section 16 (concurrency): SKIP (TC-LOCK-001, macOS limitation).
- [x] Section 17 (envelope shape): all pass.
- [x] Section 18 (error cases): all pass.
- [x] Section 19 (stream routing): all pass.
- [x] Section 20 (regression): all pass.
