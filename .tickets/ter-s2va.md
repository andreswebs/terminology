---
id: ter-s2va
status: closed
deps: [ter-6z5g]
links: []
created: 2026-05-24T00:21:44Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, model, foundation]
---
# E2.T2 — Style, dialect, and warning types

## Goal

Define the style enum (DCT/DCA), the dialect identifier type, and the reader warning struct. These are small but foundational types referenced by every other E2 package.

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Architecture" (style.go, warning.go), §"Reader warnings"
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md) §"XML output" (DCT vs DCA style)

## Files to create

- `src/internal/tbx/style.go`
- `src/internal/tbx/warning.go`
- `src/internal/tbx/style_test.go`

## Type definitions

### style.go

```go
package tbx

type Style int

const (
    StyleDCT Style = iota
    StyleDCA
)

func (s Style) String() string {
    switch s {
    case StyleDCT:
        return "dct"
    case StyleDCA:
        return "dca"
    default:
        return "unknown"
    }
}

type Dialect string

const DialectLinguist Dialect = "TBX-Linguist"
```

Design notes:
- **Style is an iota enum** with a String() method for use in XML attributes (`style="dct"`).
- **Dialect is a string type**, not an enum, so future dialects (TBX-Basic, TBX-Min) register as new constants without modifying existing code.
- **DialectLinguist** is the only concrete dialect in v1.

### warning.go

```go
package tbx

type Warning struct {
    Code      string
    Message   string
    ConceptID string
    Line, Col int
}
```

Design notes:
- **Line, Col fields** are populated by the reader when position tracking is available (via the lineindex package from T5). When position is unavailable, they default to 0.
- **Code field** uses snake_case codes matching the project's error code convention: `legacy_form_normalized`, `unknown_element`, `duplicate_id`, etc.
- Warnings are data, not errors — the reader returns them alongside the parsed glossary. This keeps the reader free of logging-policy concerns.

## TDD cycles

### Cycle 1 — Style enum
RED: Test StyleDCT.String() == "dct", StyleDCA.String() == "dca".
GREEN: Implement Style type and String() method.

### Cycle 2 — Dialect constant
RED: Test that DialectLinguist == Dialect("TBX-Linguist").
GREEN: Define Dialect type and constant.

### Cycle 3 — Warning struct
RED: Test that a Warning{Code: "test", Message: "msg", ConceptID: "c1", Line: 10, Col: 5} has all fields accessible.
GREEN: Define Warning struct.

## Out of scope

- Reader/Writer interfaces (T4)
- Error sentinels (T3)
- Model types (T1)

## Acceptance

- `make build` passes
- Style enum has DCT=0, DCA=1 with correct String() output
- DialectLinguist constant equals "TBX-Linguist"
- Warning struct has all 5 fields (Code, Message, ConceptID, Line, Col)


## Notes

**2026-05-25T13:08:20Z**

Implementation already existed from prior work (style.go, warning.go). Added style_test.go with table-driven tests for Style.String() (DCT, DCA, unknown), iota value assertions (DCT=0, DCA=1), DialectLinguist constant equality check, Warning struct field access test (all 5 fields), and zero-value default test. make build passes.
