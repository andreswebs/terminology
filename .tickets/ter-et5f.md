---
id: ter-et5f
status: closed
deps: [ter-oddb, ter-6z5g]
links: []
created: 2026-05-24T00:24:51Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, reader, linguist]
---
# E2.T7 — Linguist DCA reader

## Goal

Add the DCA (Data Category as Attribute) decoding paths to the Linguist reader. DCA is the alternative TBX-Linguist style where data categories are expressed as \`type\` attributes on generic elements (\`<descrip type="subjectField">\`, \`<termNote type="partOfSpeech">\`) rather than namespace-qualified element names.

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Architecture" (DCT + DCA on input)
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"TBX-Linguist dialect" (DCA element mapping)

## Files to modify

- \`src/internal/tbx/linguist/reader.go\` — add DCA decode functions

## Files to create

- \`src/internal/tbx/linguist/testdata/canonical/minimal-dca.tbx\` (test fixture)

## Functions to implement

```go
func decodeConceptDCA(dec *xml.Decoder, se xml.StartElement, c *tbx.Concept) ([]tbx.Warning, error)
func decodeLangSecDCA(dec *xml.Decoder, se xml.StartElement, ls *tbx.LangSection) ([]tbx.Warning, error)
func decodeTermDCA(dec *xml.Decoder, se xml.StartElement, term *tbx.Term) ([]tbx.Warning, error)
```

Design notes:
- **DCA dispatch uses \`(localName, type)\` pairs** instead of \`(namespace, localName)\`. Example: \`<descrip type="subjectField">\` maps to \`Concept.SubjectField\`.
- **Element-to-type mapping** (DCA):
  - Concept level: \`descrip[subjectField]\`, \`descrip[definition]\`, \`ref[crossReference]\`, \`xref[externalCrossReference]\`, \`xref[xGraphic]\`, \`admin[source|customerSubset|projectSubset]\`
  - LangSec level: \`descrip[definition]\`, \`admin[source]\`
  - Term level: \`termNote[administrativeStatus|partOfSpeech|grammaticalGender|grammaticalNumber|register|termType|termLocation|geographicalUsage|transferComment]\`, \`descrip[context]\`, \`admin[source|customerSubset|projectSubset]\`, \`xref[externalCrossReference]\`, \`ref[crossReference]\`
- The \`decodeConceptChild\` function (from T6) already dispatches to DCA vs DCT based on \`style\`. This ticket implements the DCA branches.
- The \`attrVal\` helper (from T6) extracts the \`type\` attribute.

## Minimal DCA fixture

Create \`src/internal/tbx/linguist/testdata/canonical/minimal-dca.tbx\` — same semantic content as minimal-dct.tbx but in DCA style:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dca" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader>
    <fileDesc>
      <sourceDesc>
        <p>Minimal test fixture</p>
      </sourceDesc>
    </fileDesc>
  </tbxHeader>
  <text>
    <body>
      <conceptEntry id="tzimtzum">
        <descrip type="subjectField">kabbalah</descrip>
        <langSec xml:lang="en">
          <termSec>
            <term>tzimtzum</term>
            <termNote type="administrativeStatus">preferredTerm-admn-sts</termNote>
            <termNote type="partOfSpeech">noun</termNote>
          </termSec>
        </langSec>
        <langSec xml:lang="he">
          <termSec>
            <term>צמצום</term>
            <termNote type="administrativeStatus">preferredTerm-admn-sts</termNote>
            <termNote type="partOfSpeech">noun</termNote>
          </termSec>
        </langSec>
      </conceptEntry>
    </body>
  </text>
</tbx>
```

## TDD cycles

### Cycle 1 — Detect DCA style
RED: Decode minimal-dca.tbx. Assert g.Style==StyleDCA.
GREEN: detectStyle (from T6) already handles this — verify it works with DCA input.

### Cycle 2 — DCA concept fields
RED: Assert concept ID=="tzimtzum", SubjectField=="kabbalah".
GREEN: Implement decodeConceptDCA for \`descrip[subjectField]\`.

### Cycle 3 — DCA term fields
RED: Assert en term Surface=="tzimtzum", AdministrativeStatus==StatusPreferred, PartOfSpeech=="noun".
GREEN: Implement decodeTermDCA for \`termNote[administrativeStatus]\` and \`termNote[partOfSpeech]\`.

### Cycle 4 — All DCA term-level categories
RED: Fixture with all DCA term elements. Assert each field matches.
GREEN: Implement remaining decodeTermDCA cases.

### Cycle 5 — All DCA concept-level categories
RED: Fixture with definition, crossReference, etc. at concept level using DCA syntax.
GREEN: Implement remaining decodeConceptDCA cases.

### Cycle 6 — DCA langSec categories
RED: Fixture with langSec-level definition and source in DCA style.
GREEN: Implement decodeLangSecDCA.

### Cycle 7 — DCA unknown elements skipped
RED: DCA fixture with unknown \`<descrip type="custom">...</descrip>\`. Assert no error.
GREEN: Default case calls dec.Skip().

## Out of scope

- DCA writer (spec says v1 only writes DCT — T10)
- adminGrp/transacGrp DCA paths (T8)
- Normalization (T9)

## Acceptance

- \`make build\` passes
- DCA fixtures decode to the same domain model as equivalent DCT fixtures
- All DCA data categories at concept, langSec, and term level are handled
- Unknown DCA elements are skipped without error


## Notes

**2026-05-25T13:28:39Z**

DCA reader code was already implemented in reader.go (decodeConceptDCA, decodeLangSecDCA, decodeTermDCA, plus DCA branches in decodeAdminGrp and decodeTransacGrp). This ticket completed the test coverage: created all-categories-dca.tbx comprehensive fixture (DCA equivalent of all-categories-dct.tbx), added DCA-specific tests for all term/concept/langSec categories reusing shared assertion helpers, added unknown DCA element skipping test, and expanded DCA→DCT roundtrip test to cover the comprehensive fixture. Refactored existing test assertions into reusable helpers (assertAllTermCategories, assertAllConceptCategories, assertLangSecCategories) shared by both DCT and DCA tests. Updated TestRoundTrip_Canonical to skip all *-dca.tbx files (not just minimal-dca.tbx).
