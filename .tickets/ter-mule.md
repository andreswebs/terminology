---
id: ter-mule
status: closed
deps: [ter-8zuv, ter-a52o, ter-bbbs, ter-12nr, ter-bedf]
links: []
created: 2026-05-24T01:05:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, validation, strict]
---
# E3.T10 — legacy_form_normalized + --strict promotions

## Goal

Implement two cross-cutting validation behaviors:

1. **`legacy_form_normalized` warning** — info-only warning emitted when the reader normalizes a legacy form on read. Two normalization cases exist (from E2.T9):
   - Bare admin status (`preferredTerm` → `preferredTerm-admn-sts`)
   - Legacy register (`usageRegister` → `register`)
   
   This warning is **`--strict` only** — in lenient mode, normalization happens silently.

2. **`--strict` promotions** — the `--strict` flag changes how two warning types are routed:
   - `unknown_element`: silent → **warning** (from T9)
   - `unresolved_crossref`: warning → **error** (from T7)

Both behaviors are wired in the reader and/or the validate command layer.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"--strict", §"Warning codes" (`legacy_form_normalized`)
- Normalization: `src/internal/tbx/linguist/normalize.go` (E2.T9)

## Files to modify

- `src/internal/tbx/linguist/reader.go` — emit `legacy_form_normalized` warnings during normalization
- `src/internal/tbx/linguist/reader_test.go` — test legacy_form_normalized emission
- `src/internal/tbx/validate.go` — ensure `Validate(strict)` routes `unresolved_crossref` correctly (already done in T7, this ticket verifies the strict behavior)

## Implementation

### legacy_form_normalized

In the reader's normalization points (where `normalizeStatus` and `normalizeRegister` are called), check if the value was changed. If so, emit a warning:

```go
original := statusText
normalized := normalizeStatus(statusText)
if original != statusToString(normalized) {
    warnings = append(warnings, tbx.Warning{
        Code:      "legacy_form_normalized",
        Message:   fmt.Sprintf("concept %q: %q normalized to canonical form", conceptID, original),
        ConceptID: conceptID,
    })
}
```

The `legacy_form_normalized` warning is always collected by the reader. The validate command filters it:
- In lenient mode: `legacy_form_normalized` warnings are **suppressed** (not included in the envelope).
- In strict mode: `legacy_form_normalized` warnings are **included** in the envelope.

### --strict promotions summary

| Warning code | Lenient | Strict |
|---|---|---|
| `unknown_element` | silent (filtered from output) | warning (included) |
| `unresolved_crossref` | warning | error |
| `legacy_form_normalized` | silent (filtered from output) | warning (included) |

The validate command (T12) handles the filtering. The reader always collects all warnings regardless of mode.

## Test fixture

`src/internal/tbx/linguist/testdata/normalized/legacy-forms.tbx` — file with bare `preferredTerm` status (no `-admn-sts` suffix) and `usageRegister`. This fixture should already exist from E2.

## TDD cycles

### Cycle 1 — Bare admin status triggers legacy_form_normalized
RED: Decode a file with `<min:administrativeStatus>preferredTerm</min:administrativeStatus>`. Assert 1 warning with code `"legacy_form_normalized"`.
GREEN: Add normalization check in the reader.

### Cycle 2 — Legacy register triggers legacy_form_normalized
RED: Decode a file with `usageRegister`. Assert `legacy_form_normalized` warning.
GREEN: Add normalization check for register.

### Cycle 3 — Canonical form produces no legacy warning
RED: Decode `minimal-dct.tbx` (uses `preferredTerm-admn-sts`). Assert no `legacy_form_normalized` warnings.
GREEN: Already passing — normalization is a no-op for canonical forms.

### Cycle 4 — Strict mode includes legacy warnings in validate
RED: Integration test: run validate with `--strict` on a file with legacy forms. Assert `legacy_form_normalized` appears in the warnings array.
GREEN: Validate command includes `legacy_form_normalized` when strict.

### Cycle 5 — Lenient mode suppresses legacy warnings
RED: Run validate without `--strict` on the same file. Assert `legacy_form_normalized` does NOT appear.
GREEN: Validate command filters `legacy_form_normalized` when lenient.

### Cycle 6 — Strict promotes unresolved_crossref to error
RED: Run `Validate(true)` on a glossary with unresolved cross-ref. Assert the issue is in `res.Errors`, not `res.Warnings`.
GREEN: Already implemented in T7 — this cycle confirms the behavior.

## Deviation note

The current implementation does NOT emit `legacy_form_normalized` warnings. Normalization happens silently in `normalizeStatus` and `normalizeRegister` (in `linguist/normalize.go`). The strict-mode promotion of `unresolved_crossref` from warning to error IS implemented in `validate.go`.

Changes needed:
1. Add `legacy_form_normalized` warning emission in the reader at normalization points
2. The validate command (T12) needs filtering logic for strict-only warnings (`unknown_element`, `legacy_form_normalized`)

The `unknown_element` strict promotion is implemented in T9 (reader always collects) + T12 (command filters).

## Out of scope

- Normalization functions themselves (E2.T9)
- Other warning codes
- The validate command action (T12 — handles filtering)

## Acceptance

- `make build` passes
- Reader emits `legacy_form_normalized` when a value is normalized
- Canonical forms produce no legacy warnings
- Strict mode includes `legacy_form_normalized` and `unknown_element` in output
- Lenient mode suppresses `legacy_form_normalized` and `unknown_element`
- Strict mode promotes `unresolved_crossref` from warning to error


## Notes

**2026-05-25T18:49:55Z**

Implementation gap: the reader does NOT emit legacy_form_normalized warnings. Normalization happens silently in normalizeStatus/normalizeRegister. The validate command (commands/validate.go) does NOT filter unknown_element or legacy_form_normalized warnings by --strict mode — all warnings pass through unfiltered. Two changes needed: (1) reader must emit legacy_form_normalized warnings at normalization points (comparing before/after values), (2) validate command must filter strict-only warnings (unknown_element, legacy_form_normalized) in lenient mode. The unresolved_crossref strict→error promotion IS already implemented in validate.go.

**2026-05-25T19:11:00Z**

Implemented legacy_form_normalized warning emission and --strict filtering. Changes: (1) normalize.go: added isLegacyStatus() and isLegacyRegister() helpers. (2) reader.go: all 4 normalization points (DCT/DCA × status/register) now emit legacy_form_normalized warnings when input is a legacy form. (3) validate.go: added isStrictOnly() filter — lenient mode suppresses unknown_element and legacy_form_normalized warnings, strict mode includes them. (4) Updated TestDecode_LegacyFormsNormalized to expect 5 legacy warnings from the fixture. (5) Added tests: TestDecode_LegacyAdminStatus_EmitsLegacyWarning, TestDecode_LegacyRegister_EmitsLegacyWarning, TestDecode_CanonicalForms_NoLegacyWarning, TestDecode_LegacyAdminStatus_DCA_EmitsLegacyWarning, TestValidate_Lenient_SuppressesStrictOnlyWarnings, TestValidate_Strict_IncludesStrictOnlyWarnings, TestValidate_Strict_PromotesUnresolvedCrossrefToError. (6) Added test fixture with-legacy-and-unknown.tbx.
