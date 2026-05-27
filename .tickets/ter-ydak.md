---
id: ter-ydak
status: closed
deps: [ter-6z5g]
links: []
created: 2026-05-24T00:21:40Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, model, foundation]
---
# E2.T1 — Domain model types (model.go)

## Goal

Stand up the dialect-agnostic domain model types that every other package in the codebase consumes. These are the internal representation of a TBX glossary — independent of any XML shape.

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Domain types", §"Public API of model types"
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"internal/tbx"
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md) §"Sort orders summary" (term ordering by Status)

## Files to create

- `src/internal/tbx/model.go`
- `src/internal/tbx/model_test.go`

## Type definitions

```go
package tbx

type Glossary struct {
    Dialect    Dialect
    Style      Style
    SourceDesc string
    Concepts   []Concept
}

type Concept struct {
    ID             string
    SubjectField   string
    Definitions    []NoteText
    CrossRefs      []CrossRef
    ExternalRefs   []string
    Graphics       []string
    Sources        []string
    CustomerSubset string
    ProjectSubset  string
    Transactions   []Transaction
    Notes          []string
    Languages      map[string]LangSection
}

type LangSection struct {
    Lang        string
    Definitions []NoteText
    Sources     []string
    Terms       []Term
}

type Term struct {
    Surface              string
    AdministrativeStatus Status
    PartOfSpeech         string
    GrammaticalGender    string
    GrammaticalNumber    string
    Register             string
    TermType             string
    TermLocation         string
    GeographicalUsage    string
    Contexts             []NoteText
    TransferComment      string
    Reading              string
    ReadingNote          string
    Sources              []string
    CustomerSubset       string
    ProjectSubset        string
    ExternalRefs         []string
    CrossRefs            []CrossRef
    Transactions         []Transaction
    Notes                []string
}

type Status int

const (
    StatusUnspecified Status = iota
    StatusPreferred
    StatusAdmitted
    StatusDeprecated
    StatusSuperseded
)

type CrossRef struct {
    Target string
    Label  string
}

type Transaction struct {
    Type           string
    Date           string
    Responsibility string
}

type NoteText struct {
    Plain string
    Raw   string
}
```

Design notes:
- **Exported fields, not accessors.** Go convention for value types. Trivial to inspect in tests. Setters arrive only when an invariant needs protecting (e.g. when concept add enforces ID derivation in E7).
- **`Languages` is `map[string]LangSection`** keyed by BCP 47 tag. The map provides O(1) lookup by language for read commands.
- **`Status` is an iota enum**, not a string. The reader normalizes string forms into enum values; the writer converts back. This makes status-based sorting trivial.
- **`NoteText` has Plain + Raw** to support inline XML in definitions/contexts while keeping a plain-text accessor for matching and display.

## TDD cycles

### Cycle 1 — Status enum
RED: Test that StatusPreferred through StatusSuperseded have distinct non-zero int values, and StatusUnspecified is zero.
GREEN: Define the Status iota const block.

### Cycle 2 — Struct instantiation
RED: Test that a Glossary can be constructed with Concepts containing LangSections and Terms, and fields are accessible. Verify a Concept's Languages map can be populated and read back.
GREEN: Define all struct types.

### Cycle 3 — CrossRef and Transaction
RED: Test that CrossRef{Target: "x", Label: "y"} and Transaction{Type: "origination", Date: "2026-01-01T00:00:00Z", Responsibility: "author"} round-trip through field access.
GREEN: Already passing from cycle 2; this cycle confirms the types are correct.

## Out of scope

- Validation logic (E3)
- Concept-ID derivation (E7)
- Style/Dialect types (T2)
- Any methods on these types beyond field access

## Acceptance

- `make build` passes
- All struct types are exported with exported fields
- Status enum has 5 values (Unspecified + 4 named statuses)
- No methods on model types (plain data)


## Notes

**2026-05-25T13:06:26Z**

Created model_test.go with tests covering all TDD cycles from the ticket: Status enum values (iota ordering, distinct non-zero values), Glossary construction with nested Concepts/LangSections/Terms including Hebrew/English multi-language map access, all Concept fields (definitions with NoteText Plain+Raw, cross-refs, external refs, graphics, sources, transactions, notes, customer/project subsets), CrossRef and Transaction field access, all Term fields (20 fields including morphology, register, reading, contexts), LangSection fields, and NoteText plain vs raw. Implementation was already in place from prior work; this ticket added the test coverage.
