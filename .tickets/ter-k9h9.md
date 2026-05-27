---
id: ter-k9h9
status: closed
deps: [ter-oddb, ter-6z5g]
links: []
created: 2026-05-24T00:26:53Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, reader, linguist]
---
# E2.T8 — Linguist reader: adminGrp + transacGrp decoders

## Goal

Implement the adminGrp and transacGrp decoders for the Linguist reader. These handle administrative groups (reading/readingNote) and transaction groups (transactionType/date/responsibility) at both the concept and term level, in both DCT and DCA styles.

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Architecture" (reader.go)
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"TBX-Linguist dialect" — adminGrp contains reading + readingNote; transacGrp contains transactionType + date + responsibility
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md) §"Transaction timestamps" — timestamps are RFC3339 UTC

## Files to modify

- \`src/internal/tbx/linguist/reader.go\` — add decodeAdminGrp and decodeTransacGrp

## Files to create

- \`src/internal/tbx/linguist/testdata/canonical/with-transactions.tbx\`
- \`src/internal/tbx/linguist/testdata/canonical/full-features.tbx\`

## Functions to implement

```go
func decodeAdminGrp(dec *xml.Decoder, start xml.StartElement, term *tbx.Term, style tbx.Style) ([]tbx.Warning, error)
func decodeTransacGrp(dec *xml.Decoder, start xml.StartElement, style tbx.Style) (tbx.Transaction, error)
```

Design notes:
- **adminGrp** contains reading/readingNote data categories:
  - DCT: \`<ling:reading>\` and \`<ling:readingNote>\` namespace-qualified elements
  - DCA: \`<admin type="reading">\` and \`<adminNote type="readingNote">\`
- **transacGrp** contains transaction metadata:
  - DCT: \`<basic:transactionType>\`, \`<date>\`, \`<basic:responsibility>\`
  - DCA: \`<transac type="transactionType">\`, \`<date>\`, \`<transacNote type="responsibility">\`
- **transacGrp appears at both concept and term level**. The T6 decoder already routes \`<transacGrp>\` at the concept level to this function, and the term decoder routes it similarly.
- **langSec-level transacGrp** is skipped in v1 (the model has no field for it). The langSec decoder in T6 handles this by calling dec.Skip() when it encounters transacGrp.

## Test fixtures

### with-transactions.tbx
A DCT fixture with transaction groups at both concept and term level:
- Concept-level: origination transaction with date and responsibility
- Term-level: modification transaction

### full-features.tbx
A DCT fixture exercising all features including adminGrp (reading + readingNote), contexts, transferComment, notes, definitions — the comprehensive feature coverage fixture.

## TDD cycles

### Cycle 1 — DCT transacGrp at concept level
RED: Fixture with \`<transacGrp>\` inside conceptEntry. Assert c.Transactions[0].Type=="origination", Date matches, Responsibility set.
GREEN: Implement decodeTransacGrp for DCT style.

### Cycle 2 — DCT transacGrp at term level
RED: Fixture with \`<transacGrp>\` inside termSec. Assert term.Transactions[0] is populated.
GREEN: Wire term-level transacGrp routing (already partially done in T6).

### Cycle 3 — DCT adminGrp (reading + readingNote)
RED: Fixture with \`<adminGrp>\` containing \`<ling:reading>\` and \`<ling:readingNote>\`. Assert term.Reading and term.ReadingNote.
GREEN: Implement decodeAdminGrp for DCT style.

### Cycle 4 — DCA transacGrp
RED: DCA fixture with \`<transac type="transactionType">\` and \`<transacNote type="responsibility">\`. Assert transaction fields.
GREEN: Add DCA branch to decodeTransacGrp.

### Cycle 5 — DCA adminGrp
RED: DCA fixture with \`<admin type="reading">\` and \`<adminNote type="readingNote">\`. Assert reading fields.
GREEN: Add DCA branch to decodeAdminGrp.

### Cycle 6 — Unknown elements inside groups
RED: adminGrp containing an unknown child element. Assert no error, reading/readingNote still parsed.
GREEN: Default case calls dec.Skip().

## Out of scope

- Transaction timestamp validation (E3/E9)
- Writer transaction output (T10)
- Normalization (T9)

## Acceptance

- \`make build\` passes
- transacGrp decodes correctly at concept and term level in both DCT and DCA
- adminGrp decodes reading + readingNote in both DCT and DCA
- Unknown children within groups are skipped
- Test fixtures exercise all adminGrp and transacGrp paths


## Notes

**2026-05-25T13:33:32Z**

decodeAdminGrp and decodeTransacGrp were already implemented in reader.go (likely done during T6/T7). This ticket's work was adding comprehensive test coverage:

- TestDecode_TransacGrp_ConceptLevel: DCT concept-level transaction fields
- TestDecode_TransacGrp_TermLevel: DCT term-level transaction fields  
- TestDecode_AdminGrp_ReadingAndReadingNote: DCT reading/readingNote from full-features.tbx
- TestDecode_TransacGrp_DCA: DCA transactions at both concept and term level (inline fixture)
- TestDecode_AdminGrp_DCA: DCA reading/readingNote (inline fixture)
- TestDecode_UnknownElementsInsideGroups: Unknown children in transacGrp and adminGrp are skipped
- TestDecode_LangSecTransacGrp_Skipped: langSec-level transacGrp is skipped per v1 model

Updated with-transactions.tbx to include term-level transacGrp (modification transaction with date and responsibility).
