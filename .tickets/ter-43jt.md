---
id: ter-43jt
status: closed
deps: [ter-bedf]
links: []
created: 2026-05-24T01:02:13Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, picklist, foundation]
---
# E3.T1 — Picklist values (picklist.go)

## Goal

Create `internal/tbx/picklist.go` — the single source of truth for all TBX picklist values in the codebase. Every function returns a `[]string` of accepted values. These slices are consumed by two layers: **file-side validation** (E3 — `invalid_picklist` warnings when a reader encounters an out-of-set value) and **CLI flag enums** (E7 — urfave flag validators that reject invalid user input at the command line).

The two-layer design makes picklist drift structurally impossible: both sides import the same constants from the same file.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Picklist values"
- E7 spec: [docs/specs/007-write-commands.md](docs/specs/007-write-commands.md) (consumer of picklists)
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"TBX-Linguist dialect" (accepted values per data category)

## Files to create

- `src/internal/tbx/picklist.go`
- `src/internal/tbx/picklist_test.go`

## API

```go
package tbx

func Format() []string           { return []string{"json", "text"} }
func AdminStatus() []string      { return []string{...} }
func PartOfSpeech() []string     { return []string{...} }
func GrammaticalGender() []string { return []string{...} }
func Register() []string         { return []string{...} }
func GrammaticalNumber() []string { return []string{...} }
func TermType() []string         { return []string{...} }
func TransactionType() []string  { return []string{...} }
func Script() []string           { return []string{...} }
```

Design notes:
- **Functions, not vars** — returning a fresh slice each call prevents caller mutation from corrupting shared state.
- **AdminStatus includes both canonical and legacy bare forms** — the spec shows `preferredTerm-admn-sts` etc. as canonical. The current implementation also includes bare forms (`preferredTerm`, `admittedTerm`, etc.) to support legacy TBX files that use them. Keep both sets so the reader can validate either form.
- **No `Contains(list, val)` helper** — callers build their own set when they need O(1) lookup. This keeps picklist.go a pure data file with no logic.

### Accepted values per picklist

| Picklist | Values |
|----------|--------|
| Format | `json`, `text` |
| AdminStatus | `preferredTerm-admn-sts`, `admittedTerm-admn-sts`, `deprecatedTerm-admn-sts`, `supersededTerm-admn-sts`, `preferredTerm`, `admittedTerm`, `deprecatedTerm`, `supersededTerm` |
| PartOfSpeech | `noun`, `verb`, `adjective`, `adverb`, `other` |
| GrammaticalGender | `masculine`, `feminine`, `neuter`, `other` |
| Register | `colloquialRegister`, `neutralRegister`, `technicalRegister`, `in-houseRegister`, `bench-levelRegister`, `slangRegister`, `vulgarRegister`, `usageRegister` |
| GrammaticalNumber | `singular`, `plural`, `dual`, `mass`, `otherNumber` |
| TermType | `fullForm`, `acronym`, `abbreviation`, `shortForm`, `variant`, `phrase` |
| TransactionType | `origination`, `modification` |
| Script | `latin`, `hebrew`, `cyrillic`, `arabic`, `any` |

## TDD cycles

### Cycle 1 — Non-empty, unique elements
RED: Table test that calls every picklist function and asserts: (a) returned slice is non-empty, (b) no duplicate values.
GREEN: Define all functions returning their value slices.

### Cycle 2 — Canonical values present
RED: Table test asserting specific canonical values are present in each picklist. For AdminStatus, assert all four `-admn-sts` suffixed forms. For PartOfSpeech, assert `noun`, `verb`, `adjective`. For Register, assert `colloquialRegister`, `neutralRegister`, `technicalRegister`. For Script, assert `latin`, `hebrew`, `any`.
GREEN: Already passing from cycle 1 — confirms values are correct.

### Cycle 3 — AdminStatus includes legacy bare forms
RED: Assert AdminStatus() contains `preferredTerm`, `admittedTerm`, `deprecatedTerm`, `supersededTerm` (without the `-admn-sts` suffix).
GREEN: Already passing — confirms legacy form support.

## Deviation note

The current implementation matches this spec. `picklist.go` exists with all 9 functions and the values listed above. `picklist_test.go` has 2 tests covering non-empty/unique and canonical values. No changes are needed to align with the spec.

## Out of scope

- CLI flag validators that consume these picklists (E7)
- `invalid_picklist` warnings in the reader (E3.T8)
- Any `Contains` or lookup helpers

## Acceptance

- `make build` passes
- All 9 picklist functions are exported and return `[]string`
- No duplicates in any list
- AdminStatus includes both canonical (`-admn-sts`) and legacy bare forms
- Functions return fresh slices (not shared package-level vars)


## Notes

**2026-05-25T18:40:48Z**

picklist.go: all 9 functions present with correct values. Tests cover only 6/9 — GrammaticalNumber, TermType, TransactionType missing from both NonEmpty_UniqueElements and ContainCanonicalValues test tables. Also: cli-design.md lists termLocation as a picklist (18 UI-element values per Basic module) but no TermLocation() function exists in picklist.go and it is not in the ticket spec. The reader/writer already handle termLocation as a free string. Decide whether to add TermLocation() for invalid_picklist validation or treat it as out-of-scope for E3.

**2026-05-25T18:42:39Z**

Completed test coverage for all 9 picklist functions. Added GrammaticalNumber, TermType, and TransactionType to both NonEmpty_UniqueElements and ContainCanonicalValues test tables. Added TestAdminStatus_IncludesLegacyBareForms for Cycle 3 (legacy bare form coverage). Added TestPicklists_ReturnFreshSlices to verify functions return fresh slices on each call (not shared backing arrays). All acceptance criteria met: 9 exported functions, no duplicates, legacy forms included, fresh slices confirmed. make build passes clean.
