---
id: ter-g887
status: closed
deps: [ter-ydak, ter-6z5g]
links: []
created: 2026-05-24T00:26:54Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, normalize, linguist]
---
# E2.T9 — Legacy-form normalization

## Goal

Implement legacy-form normalization for the TBX-Linguist reader. Two normalization rules run on read:

1. **Administrative status**: bare forms (\`preferredTerm\`) normalize to canonical \`-admn-sts\` suffix forms (\`preferredTerm-admn-sts\`), which then map to the Status enum.
2. **Register**: legacy \`usageRegister\` element/type normalizes to \`register\`.

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Scope" — "Legacy-form normalization on read (\`usageRegister\` → \`register\`; bare \`preferredTerm\` → \`preferredTerm-admn-sts\`; etc.)"
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"Picklist values" — lists both modern and legacy forms

## Files to create

- \`src/internal/tbx/linguist/normalize.go\`
- \`src/internal/tbx/linguist/normalize_test.go\`
- \`src/internal/tbx/linguist/testdata/normalized/legacy-forms.tbx\` (test fixture)

## Functions

```go
package linguist

import "github.com/andreswebs/terminology/internal/tbx"

func normalizeStatus(s string) tbx.Status {
    switch s {
    case "preferredTerm-admn-sts", "preferredTerm":
        return tbx.StatusPreferred
    case "admittedTerm-admn-sts", "admittedTerm":
        return tbx.StatusAdmitted
    case "deprecatedTerm-admn-sts", "deprecatedTerm":
        return tbx.StatusDeprecated
    case "supersededTerm-admn-sts", "supersededTerm":
        return tbx.StatusSuperseded
    default:
        return tbx.StatusUnspecified
    }
}

func normalizeRegister(s string) string {
    if s == "usageRegister" {
        return "register"
    }
    return s
}
```

Design notes:
- **normalizeStatus maps strings to Status enum**, handling both modern (\`-admn-sts\` suffix) and legacy (bare) forms. This is the single normalization point — the reader calls it on every administrativeStatus value.
- **normalizeRegister** maps the legacy element/type name \`usageRegister\` to the canonical \`register\`. The value itself (e.g. \`colloquialRegister\`, \`technicalRegister\`) is kept as-is.
- These functions are **unexported** — consumed only by the reader's decodeTermDCT/decodeTermDCA functions.
- The reader in T6/T7 calls \`normalizeStatus(text)\` where it reads administrativeStatus and \`normalizeRegister(text)\` where it reads the register element.

## Legacy forms fixture

Create a fixture with terms using bare status forms and usageRegister:

```xml
<!-- terms with: preferredTerm (bare), admittedTerm (bare),
     deprecatedTerm (bare), supersededTerm (bare),
     and a term with usageRegister instead of register -->
```

Five terms exercising all four bare status forms plus the register normalization.

## TDD cycles

### Cycle 1 — Modern status forms
RED: Test normalizeStatus for all four \`-admn-sts\` forms. Assert correct Status enum values.
GREEN: Implement the four modern-form cases.

### Cycle 2 — Legacy bare status forms
RED: Test normalizeStatus for bare forms (\`preferredTerm\`, \`admittedTerm\`, \`deprecatedTerm\`, \`supersededTerm\`). Assert same enum values as modern forms.
GREEN: Add bare-form cases to the switch.

### Cycle 3 — Unknown status → StatusUnspecified
RED: Test normalizeStatus("unknown") and normalizeStatus("") both return StatusUnspecified.
GREEN: Default case returns StatusUnspecified.

### Cycle 4 — Register normalization
RED: Test normalizeRegister("usageRegister")=="register", normalizeRegister("colloquialRegister")=="colloquialRegister".
GREEN: Implement normalizeRegister.

### Cycle 5 — Integration with reader
RED: Decode legacy-forms.tbx fixture. Assert term statuses are normalized and register is normalized.
GREEN: Wire normalizeStatus and normalizeRegister calls into the DCT/DCA decoders (T6/T7 should already call these).

## Out of scope

- Writer status rendering (T10 — writes canonical forms only)
- Picklist validation (E3)
- Warning emission for legacy forms (future — the spec mentions \`legacy_form_normalized\` warning code but doesn't require it in E2)

## Acceptance

- \`make build\` passes
- All 8 status string forms (4 modern + 4 legacy) normalize correctly
- Unknown/empty status returns StatusUnspecified
- usageRegister normalizes to register
- Integration test confirms reader applies normalization on real fixture


## Notes

**2026-05-25T13:17:03Z**

T9 was already fully implemented: normalize.go (normalizeStatus + normalizeRegister), normalize_test.go (table-driven tests covering all 8 status forms + unknown/empty + register normalization), legacy-forms.tbx fixture, and integration test TestDecode_LegacyFormsNormalized in reader_test.go. Reader wires in both functions for DCT (lines 507, 535) and DCA (lines 638, 666) paths. make build passes clean.
