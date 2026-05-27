---
id: ter-bbbs
status: closed
deps: [ter-43jt, ter-bedf]
links: []
created: 2026-05-24T01:05:19Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, validation, tier2]
---
# E3.T8 — Tier-2: invalid_picklist validation on read

## Goal

Add `invalid_picklist` warning emission to the linguist reader. When the reader encounters a picklist value (administrative status, part-of-speech, register, grammatical gender, grammatical number, term type, transaction type) that is not in the accepted set from `tbx.picklist.go`, it emits a warning with code `"invalid_picklist"`.

The reader still accepts the value and stores it in the domain model — it does not reject the file. The warning is informational, telling the consumer that a value was encountered that doesn't match the known set.

This is a tier-2 (schema/dialect) check. It runs as the reader decodes the file and produces warnings that are aggregated with tier-3 warnings in the validate envelope.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Scope" tier 2, §"Warning codes" (`invalid_picklist`), §"Picklist values"
- Picklist source: `src/internal/tbx/picklist.go` (T1)
- Reader: `src/internal/tbx/linguist/reader.go`

## Files to modify

- `src/internal/tbx/linguist/reader.go` — add picklist validation in decode functions
- `src/internal/tbx/linguist/reader_test.go` — test invalid picklist detection

## Implementation

The reader's `Decode()` method already returns `(*Glossary, []Warning, error)`. Currently, warnings are only collected from `decodeAdminGrp`. This ticket extends warning collection to all picklist-valued decode points.

At each decode point where a picklist value is read, build a lookup set from the corresponding picklist function and check membership:

```go
func validatePicklist(value, fieldName, conceptID string, accepted []string) *tbx.Warning {
    for _, a := range accepted {
        if value == a {
            return nil
        }
    }
    return &tbx.Warning{
        Code:      "invalid_picklist",
        Message:   fmt.Sprintf("concept %q: %s value %q not in accepted set", conceptID, fieldName, value),
        ConceptID: conceptID,
    }
}
```

Picklist validation points in the reader:
- `<min:administrativeStatus>` → `tbx.AdminStatus()`
- `<min:partOfSpeech>` → `tbx.PartOfSpeech()`
- `<ling:grammaticalGender>` → `tbx.GrammaticalGender()`
- `<ling:grammaticalNumber>` → `tbx.GrammaticalNumber()`
- `<ling:register>` → `tbx.Register()` (after normalization from E2.T9)
- `<ling:termType>` → `tbx.TermType()`
- Transaction `type` attribute → `tbx.TransactionType()`

Design notes:
- **Validation happens after normalization** — the reader normalizes legacy forms (e.g. `usageRegister` → `register`) before validating. The picklist check runs on the normalized value.
- **Build lookup sets once per decode call**, not per element — cache the sets at the start of `Decode()` to avoid repeated allocation.
- **Thread warnings through the decode chain** — each decode function returns `[]tbx.Warning` which gets appended to the top-level warnings slice.

## Test fixture to create

`src/internal/tbx/linguist/testdata/malformed/invalid-picklist.tbx` — a DCT file with one term having `<min:partOfSpeech>frobnicator</min:partOfSpeech>` (not in the accepted set).

## TDD cycles

### Cycle 1 — Valid picklist, no warning
RED: Decode `minimal-dct.tbx` (uses `noun` for partOfSpeech). Assert no `invalid_picklist` warnings.
GREEN: Add picklist validation (passes because value is valid).

### Cycle 2 — Invalid partOfSpeech produces warning
RED: Decode `invalid-picklist.tbx` with `partOfSpeech=frobnicator`. Assert 1 `invalid_picklist` warning with concept_id and field name in message.
GREEN: Wire picklist validation into `decodeTermDCT` for `partOfSpeech`.

### Cycle 3 — Invalid administrativeStatus
RED: Fixture with `<min:administrativeStatus>bogus</min:administrativeStatus>`. Assert `invalid_picklist` warning.
GREEN: Wire picklist validation into `decodeTermDCT`/`decodeAdminGrp` for administrativeStatus.

### Cycle 4 — Multiple invalid values produce multiple warnings
RED: Fixture with invalid partOfSpeech AND invalid register. Assert 2 `invalid_picklist` warnings.
GREEN: Already passing — each validation point emits independently.

### Cycle 5 — Valid after normalization
RED: Fixture with `usageRegister` (legacy form). After normalization to `register`, value may or may not be in the accepted set. Assert the correct behavior (warning only if the normalized value is not in the set).
GREEN: Ensure validation runs after normalization.

## Deviation note

The current implementation does NOT perform picklist validation on read. The reader silently accepts any value for picklist fields. The `invalid_picklist` warning code mentioned in the spec is not implemented. This ticket adds the validation layer to the reader.

The main changes needed:
1. Add a `validatePicklist` helper function in `linguist/reader.go`
2. Wire it into `decodeTermDCT`, `decodeTermDCA`, `decodeAdminGrp`, and `decodeTransacGrp`
3. Thread warnings from each decode function up to the top-level `Decode()` return
4. Create test fixture with invalid picklist values

## Out of scope

- CLI flag validators for picklists (E7)
- `--strict` interaction with invalid_picklist (not specified — invalid_picklist is always a warning)
- Modifying picklist.go values (T1)

## Acceptance

- `make build` passes
- Reader emits `invalid_picklist` warnings for out-of-set values
- Valid picklist values produce no warnings
- Warnings carry concept_id and identify which field has the invalid value
- Reader still stores the value in the model (warning, not error)


## Notes

**2026-05-25T18:49:35Z**

Implemented invalid_picklist warning emission in the linguist reader. Added checkPicklist() helper in reader.go. Wired picklist validation into decodeTermDCT, decodeTermDCA, and decodeTransacGrp for all 7 picklist fields: administrativeStatus, partOfSpeech, grammaticalGender, grammaticalNumber, register, termType, transactionType. Changed decodeTransacGrp signature from (Transaction, error) to (Transaction, []Warning, error) to support warning propagation. Updated all callers. Validation runs on raw text before normalization; legacy forms (e.g. preferredTerm, usageRegister) are in the picklist so they pass without warning. Values are still stored in the model regardless of validation result. Added 7 test cases and a fixture file at testdata/malformed/invalid-picklist.tbx. make build passes clean.
