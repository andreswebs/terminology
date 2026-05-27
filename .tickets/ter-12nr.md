---
id: ter-12nr
status: closed
deps: [ter-bedf]
links: []
created: 2026-05-24T01:05:20Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, validation, tier2]
---
# E3.T9 — Tier-2: unknown_element detection

## Goal

Add `unknown_element` detection to the linguist reader. When the reader encounters an XML element that is not part of the TBX-Linguist supported set (Core + Min + Basic + Linguist modules), it records the element name for potential warning emission.

In **lenient mode** (default): unknown elements are **silent** — no warning emitted. The reader skips them via `dec.Skip()`.
In **strict mode** (`--strict`): unknown elements are promoted to **warnings** with code `"unknown_element"`.

The spec says: `--strict` promotes "out-of-set element" from silent to warning.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"--strict", §"Warning codes" (`unknown_element`)
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"Required structural elements" — "tolerate unknown elements"

## Files to modify

- `src/internal/tbx/linguist/reader.go` — track unknown elements during decode
- `src/internal/tbx/linguist/reader_test.go` — test unknown element detection

## Implementation

The reader's decode functions have `default:` cases that call `dec.Skip()` for unrecognized elements. To support `unknown_element` warnings:

1. **Add a `strict bool` parameter to `Decode()`** — or alternatively, always collect unknown elements and let the caller decide whether to emit warnings. The second approach is better: the reader reports what it found, the validate command filters based on `--strict`.

Recommended approach — always collect unknown elements as warnings:

```go
func (lr *LinguistReader) Decode(r io.Reader) (*tbx.Glossary, []tbx.Warning, error) {
    // In each default: case, instead of just dec.Skip():
    warnings = append(warnings, tbx.Warning{
        Code:    "unknown_element",
        Message: fmt.Sprintf("concept %q: unknown element <%s>", conceptID, se.Name.Local),
        ConceptID: conceptID,
    })
    dec.Skip()
}
```

The validate command (T12) decides whether to include these warnings in the envelope based on `--strict`. In lenient mode, `unknown_element` warnings are filtered out before output.

**Known elements by namespace:**
- Core (`urn:iso:std:iso:30042:ed-2`): `tbx`, `tbxHeader`, `fileDesc`, `sourceDesc`, `p`, `text`, `body`, `back`, `conceptEntry`, `langSec`, `termSec`, `term`
- Min (`http://www.tbxinfo.net/ns/min`): `subjectField`, `administrativeStatus`, `partOfSpeech`
- Basic (`http://www.tbxinfo.net/ns/basic`): `definition`, `crossReference`, `externalCrossReference`, `xGraphic`, `source`, `note`, `customerSubset`, `projectSubset`, `context`, `termNote`
- Linguist (`http://www.tbxinfo.net/ns/linguist`): `grammaticalGender`, `grammaticalNumber`, `register`, `termType`, `termLocation`, `geographicalUsage`, `transferComment`, `reading`, `readingNote`, `administrativeStatusTag`, `adminGrp`, `transacGrp`, `transac`, `transacNote`

Elements not matching any of these (namespace, localName) pairs are considered unknown.

## Test fixture to create

`src/internal/tbx/linguist/testdata/malformed/unknown-element.tbx` — a DCT file with `<custom:foo>bar</custom:foo>` inside a `<conceptEntry>`.

## TDD cycles

### Cycle 1 — Clean file produces no unknown_element warnings
RED: Decode `minimal-dct.tbx`. Assert no warnings with code `"unknown_element"`.
GREEN: Ensure default cases don't emit unknown_element for known elements.

### Cycle 2 — Unknown element collected
RED: Decode `unknown-element.tbx` with `<custom:foo>bar</custom:foo>` inside conceptEntry. Assert 1 warning with code `"unknown_element"` and concept_id.
GREEN: Add warning emission in the default case of `decodeConceptDCT` / `decodeConceptChild`.

### Cycle 3 — Unknown element at term level
RED: Fixture with unknown element inside `<termSec>`. Assert `unknown_element` warning.
GREEN: Add warning emission in `decodeTermDCT` default case.

### Cycle 4 — Unknown element at langSec level
RED: Fixture with unknown element inside `<langSec>`. Assert `unknown_element` warning.
GREEN: Add warning emission in `decodeLangSecDCT` default case.

### Cycle 5 — Multiple unknown elements produce multiple warnings
RED: Fixture with 2 unknown elements. Assert 2 `unknown_element` warnings.
GREEN: Already passing.

## Deviation note

The current implementation skips unknown elements silently via `dec.Skip()` in default cases. The `unknown_element` warning code is not implemented. This ticket adds unknown element tracking.

Changes needed:
1. In each `default:` case of decode functions (`decodeConceptDCT`, `decodeConceptDCA`, `decodeLangSecDCT`, `decodeLangSecDCA`, `decodeTermDCT`, `decodeTermDCA`), emit a `tbx.Warning{Code: "unknown_element", ...}` before calling `dec.Skip()`
2. Thread these warnings up to the `Decode()` return value
3. Create test fixture with unknown elements

The validate command (T12) will filter unknown_element warnings based on `--strict`.

## Out of scope

- Filtering unknown_element warnings by strict mode (T10/T12)
- Known element set definition as a data structure (inline in decode dispatch is fine)
- DCA-style unknown element detection (follow same pattern as DCT)

## Acceptance

- `make build` passes
- Reader collects `unknown_element` warnings for elements outside the supported set
- Clean files produce no `unknown_element` warnings
- Unknown elements are still skipped (no error, no panic)
- Warnings carry concept_id and element name


## Notes

**2026-05-25T18:36:52Z**

Implemented unknown_element warning detection in the linguist reader. All 9 default: cases in decode functions (decodeConceptDCT/DCA, decodeLangSecDCT/DCA, decodeTermDCT/DCA, decodeAdminGrp DCT/DCA branches) now emit tbx.Warning{Code: "unknown_element"} before calling dec.Skip(). ConceptID is set on all warnings at the decodeConceptEntry level (avoids threading conceptID through every decode function signature). Two helper functions: unknownElementWarning (DCT, uses se.Name.Local) and unknownElementWarningDCA (includes type attribute in message). TransacGrp defaults are unchanged (doesn't return warnings — out of scope). Added 6 new test cases covering DCT, DCA, multi-concept, adminGrp DCT/DCA, and clean file. Renamed existing tests from TestDecode_UnknownElementsSkipped to TestDecode_UnknownElements_DCT/DCA for consistency.
