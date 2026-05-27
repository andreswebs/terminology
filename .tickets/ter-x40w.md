---
id: ter-x40w
status: closed
deps: [ter-ydak, ter-s2va, ter-anta, ter-6z5g]
links: []
created: 2026-05-24T00:28:07Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, writer, linguist]
---
# E2.T10 — Linguist canonical DCT writer

## Goal

Implement the canonical DCT writer for TBX-Linguist. This is a hand-rolled XML emitter (~250 LOC) that produces byte-deterministic TBX-Linguist DCT output from the domain model. \`encoding/xml\`'s emitter is unsuitable for canonical output (nondeterministic namespace prefixes, whitespace issues, no self-closing control).

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Hand-rolled XML emitter", §"DCA-on-emit" (v1 ships only DCT writer)
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md) §"XML output (TBX writes)" — the canonical rules this writer must follow
- Testing ADR: [docs/adr/testing.md](docs/adr/testing.md) §"Golden CLI tests" — byte-for-byte comparison

## Files to create

- \`src/internal/tbx/linguist/writer.go\`
- \`src/internal/tbx/linguist/writer_test.go\`

## API and structure

```go
package linguist

type LinguistWriter struct{}

func NewWriter() *LinguistWriter

func (lw *LinguistWriter) Encode(w io.Writer, g *tbx.Glossary) error
```

Internal functions:
- \`writeConceptEntry(b *xmlBuilder, c tbx.Concept)\`
- \`writeLangSec(b *xmlBuilder, ls tbx.LangSection)\`
- \`writeTermSec(b *xmlBuilder, term tbx.Term)\`
- \`writeTransaction(b *xmlBuilder, tx tbx.Transaction, indent int)\`

Sorting functions (determinism):
- \`sortedConcepts(cs []tbx.Concept) []tbx.Concept\` — by ID, ASCII byte sort
- \`sortedLangs(langs map[string]tbx.LangSection) []string\` — by xml:lang, ASCII byte sort
- \`sortedTerms(terms []tbx.Term) []tbx.Term\` — by status (preferred→admitted→deprecated→superseded→unspecified), stable secondary key = declaration order

Helper functions:
- \`statusString(s tbx.Status) string\` — Status enum → canonical string (\`preferredTerm-admn-sts\`, etc.)
- \`statusOrder(s tbx.Status) int\` — Status → sort rank

xmlBuilder:
```go
type xmlBuilder struct {
    w   io.Writer
    err error
}

func (b *xmlBuilder) line(format string, args ...any)

func xmlEscape(s string) string
```

Design notes:
- **xmlBuilder** accumulates a sticky error — \`line()\` becomes a no-op after first write failure. The caller checks \`b.err\` once at the end. This eliminates per-line error handling boilerplate.
- **xmlEscape** handles \`& < > "\` — the four XML-special characters. Single quotes are not escaped because attribute values use double quotes.
- **Canonical output rules** (from determinism ADR):
  - Fixed namespace prefixes: \`min:\`, \`basic:\`, \`ling:\`
  - Element order matches TBX-Linguist schema
  - Concepts sorted by ID (ASCII)
  - Languages sorted by xml:lang (ASCII)
  - Terms sorted by status rank (stable sort preserves declaration order within same status)
  - 2-space indent, LF line endings, single trailing newline
  - UTF-8, no BOM
  - Self-closing for empty elements (used for externalCrossReference, xGraphic with target attr)
  - XML comments stripped
- **Always writes DCT style** regardless of input style. DCA→DCT conversion happens naturally: the domain model is style-agnostic, the writer always emits DCT namespace-qualified elements.
- **statusString always writes canonical \`-admn-sts\` forms** — legacy bare forms only exist on read, never on write.
- **SourceDesc defaults to "Terminology glossary"** if empty.

## TDD cycles

### Cycle 1 — xmlBuilder and xmlEscape
RED: Write a few lines through xmlBuilder. Assert output matches expected string. Test xmlEscape for \`&\`, \`<\`, \`>\`, \`"\`.
GREEN: Implement xmlBuilder and xmlEscape.

### Cycle 2 — Minimal glossary
RED: Encode a Glossary with one concept, two languages, one term each (the minimal-dct domain model). Assert output matches the expected XML.
GREEN: Implement Encode, writeConceptEntry, writeLangSec, writeTermSec skeleton.

### Cycle 3 — Concept-level elements
RED: Glossary with subjectField, definition, crossReference, note, etc. Assert each element appears in output.
GREEN: Implement all concept-level element writes.

### Cycle 4 — Term-level elements
RED: Term with all data categories (POS, gender, number, register, termType, etc.). Assert each element appears.
GREEN: Implement all term-level element writes.

### Cycle 5 — adminGrp output
RED: Term with Reading and ReadingNote. Assert \`<adminGrp>\` wraps them.
GREEN: Implement adminGrp output logic.

### Cycle 6 — Transaction output
RED: Concept and term with transactions. Assert \`<transacGrp>\` at correct indent levels.
GREEN: Implement writeTransaction with indent parameter.

### Cycle 7 — Concept sorting
RED: Glossary with concepts ["zebra", "alpha", "middle"]. Assert output order is alpha → middle → zebra.
GREEN: Implement sortedConcepts.

### Cycle 8 — Language sorting
RED: Concept with languages {"he", "en", "es"}. Assert langSec order is en → es → he.
GREEN: Implement sortedLangs.

### Cycle 9 — Term sorting by status
RED: LangSec with terms [deprecated, preferred, admitted]. Assert output order is preferred → admitted → deprecated.
GREEN: Implement sortedTerms with statusOrder.

### Cycle 10 — Term sorting stability
RED: LangSec with two preferred terms in specific declaration order. Assert output preserves declaration order within same status.
GREEN: Use sort.SliceStable.

### Cycle 11 — StatusString
RED: Test statusString for all 5 Status values. StatusUnspecified returns "".
GREEN: Implement statusString.

### Cycle 12 — Empty elements
RED: ExternalCrossReference with target attr. Assert self-closing \`<min:externalCrossReference target="..."/>\`.
GREEN: Implement self-closing logic.

### Cycle 13 — Sticky error
RED: Encode to a writer that fails after N bytes. Assert Encode returns error, no panic.
GREEN: xmlBuilder.line no-ops after first error.

## Out of scope

- DCA writer (spec explicitly says DCA-on-emit is out of scope for v1)
- File I/O (Save, atomic write — T13)
- Round-trip testing (T14)

## Acceptance

- \`make build\` passes
- LinguistWriter satisfies tbx.Writer interface
- Output is byte-deterministic: same input → same bytes
- All canonical ordering rules enforced (concepts by ID, langs by tag, terms by status)
- All TBX-Linguist data categories emitted correctly
- 2-space indent, LF endings, trailing newline
- xmlEscape handles all XML-special characters


## Notes

**2026-05-25T13:37:46Z**

Implemented writer_test.go with 30 dedicated unit tests covering all 13 TDD cycles from the ticket. The writer implementation (writer.go) was already in place from prior work. Tests cover: xmlBuilder and xmlEscape helpers, statusString/statusOrder, minimal glossary encoding, concept-level elements, term-level elements (all data categories), adminGrp output, transactions (concept and term level, partial fields), sorting (concepts by ID, languages by tag, terms by status with stable sort), self-closing elements, XML escaping in output, default sourceDesc, omission of empty/unspecified fields, deterministic output, writer error propagation, namespace declarations, 2-space indent, and LF line endings. All tests pass, make build clean.
