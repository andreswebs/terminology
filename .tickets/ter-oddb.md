---
id: ter-oddb
status: closed
deps: [ter-ydak, ter-s2va, ter-anta, ter-cplu, ter-6z5g]
links: []
created: 2026-05-24T00:24:50Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, reader, linguist]
---
# E2.T6 — Linguist DCT reader (core decoder)

## Goal

Implement the DCT (Data Category as Tag) reader for TBX-Linguist. This is the core decoder that parses a TBX-Linguist XML file in DCT style into the domain model. DCT is the primary style: element names carry semantic meaning directly (e.g. \`<min:subjectField>\`, \`<basic:definition>\`).

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Architecture" (linguist/reader.go), §"Reader warnings", §"Hand-rolled XML emitter" (namespace URIs), §"Round-trip fidelity for unknown elements"
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"TBX-Linguist dialect" (namespace URIs, element names)
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md) §"Read/write canonicalization model" — reader preserves input order

## Files to create

- \`src/internal/tbx/linguist/reader.go\`
- \`src/internal/tbx/linguist/reader_test.go\`
- \`src/internal/tbx/linguist/testdata/canonical/minimal-dct.tbx\` (test fixture)

## API and structure

```go
package linguist

import (
    "encoding/xml"
    "io"
    "github.com/andreswebs/terminology/internal/tbx"
)

const (
    nsTBX  = "urn:iso:std:iso:30042:ed-2"
    nsMin  = "http://www.tbxinfo.net/ns/min"
    nsBase = "http://www.tbxinfo.net/ns/basic"
    nsLing = "http://www.tbxinfo.net/ns/linguist"
)

type LinguistReader struct{}

func NewReader() *LinguistReader

func (lr *LinguistReader) Decode(r io.Reader) (*tbx.Glossary, []tbx.Warning, error)
```

Internal functions (DCT-specific):
- \`detectStyle(se xml.StartElement) tbx.Style\` — reads \`@style\` attribute from \`<tbx>\` element
- \`decodeConceptEntry(dec *xml.Decoder, start xml.StartElement, style tbx.Style) (tbx.Concept, []tbx.Warning, error)\`
- \`decodeConceptChild(dec, se, *Concept, style)\` — dispatches to DCT or DCA path based on style
- \`decodeConceptDCT(dec, se, *Concept)\` — handles namespace-qualified concept-level elements
- \`decodeLangSec(dec, start, *Concept, style)\` — parses \`<langSec>\`, dispatches children
- \`decodeLangSecDCT(dec, se, *LangSection)\` — handles langSec-level DCT elements
- \`decodeTermSec(dec, start, style)\` — parses \`<termSec>\`, dispatches children
- \`decodeTermDCT(dec, se, *Term)\` — handles all term-level DCT data categories

Helper functions (shared with DCA in T7):
- \`readCharData(dec, start) (string, error)\` — reads text content, handles inline elements
- \`readNoteText(dec, start) (NoteText, error)\` — reads text preserving both plain and raw (with inline XML)
- \`attrVal(se, name) string\` — extracts attribute value by local name

Design notes:
- **Streaming XML decoder** (\`encoding/xml\`) — not DOM-based. Memory stays proportional to document depth, not size.
- **Namespace-aware dispatch**: DCT elements are identified by \`(namespace, localName)\` pair, e.g. \`(nsMin, "subjectField")\`.
- **Unknown elements are skipped** via \`dec.Skip()\` — per spec, round-trip fidelity for unknown elements is out of scope for v1.
- **Reader preserves input order** in the model — sorting happens only on write (per determinism ADR).
- **Line/col tracking**: If the lineindex package (T5) is available, the reader should capture \`dec.InputOffset()\` before decoding each element and use it to populate Warning.Line/Col. If T5 is not yet available, leave Line/Col as zero — the fields exist on Warning (T2) for future use.

## Minimal DCT fixture

Create \`src/internal/tbx/linguist/testdata/canonical/minimal-dct.tbx\`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
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
        <min:subjectField>kabbalah</min:subjectField>
        <langSec xml:lang="en">
          <termSec>
            <term>tzimtzum</term>
            <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
            <min:partOfSpeech>noun</min:partOfSpeech>
          </termSec>
        </langSec>
        <langSec xml:lang="he">
          <termSec>
            <term>צמצום</term>
            <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
            <min:partOfSpeech>noun</min:partOfSpeech>
          </termSec>
        </langSec>
      </conceptEntry>
    </body>
  </text>
</tbx>
```

## TDD cycles

### Cycle 1 — Decode minimal DCT
RED: Load minimal-dct.tbx, assert g.Dialect==DialectLinguist, g.Style==StyleDCT, len(g.Concepts)==1.
GREEN: Implement Decode skeleton: parse \`<tbx>\` for style, set dialect, iterate to \`<body>\`, decode conceptEntry stubs.

### Cycle 2 — Concept fields
RED: Assert concept ID=="tzimtzum", SubjectField=="kabbalah".
GREEN: Implement decodeConceptEntry and decodeConceptDCT for subjectField.

### Cycle 3 — LangSection parsing
RED: Assert len(c.Languages)==2, "en" and "he" keys exist.
GREEN: Implement decodeLangSec with xml:lang extraction.

### Cycle 4 — Term parsing
RED: Assert en term Surface=="tzimtzum", AdministrativeStatus==StatusPreferred, PartOfSpeech=="noun".
GREEN: Implement decodeTermSec and decodeTermDCT for status + POS.

### Cycle 5 — All term-level data categories
RED: Create a fixture with all DCT term elements (grammaticalGender, grammaticalNumber, register, termType, termLocation, geographicalUsage, context, transferComment, source, customerSubset, projectSubset, externalCrossReference, crossReference). Assert each field is populated.
GREEN: Implement remaining decodeTermDCT cases.

### Cycle 6 — Concept-level data categories
RED: Fixture with definition, crossReference, externalCrossReference, xGraphic, source, customerSubset, projectSubset, note at concept level. Assert each field.
GREEN: Implement remaining decodeConceptDCT cases.

### Cycle 7 — LangSec-level data categories
RED: Fixture with definition and source at langSec level.
GREEN: Implement decodeLangSecDCT.

### Cycle 8 — readNoteText preserves inline XML
RED: Definition containing \`<hi>bold</hi>\` inline markup. Assert NoteText.Plain=="bold" (stripped), NoteText.Raw contains "\<hi>bold\</hi>".
GREEN: Implement readNoteText with plain/raw tracking.

### Cycle 9 — Unknown elements skipped without error
RED: Fixture with an unknown element \`<custom:foo>bar</custom:foo>\` inside conceptEntry. Assert no error, no panic, concept parsed correctly.
GREEN: Ensure default case calls dec.Skip().

### Cycle 10 — No warnings for clean DCT
RED: Parse minimal fixture, assert len(warnings)==0.
GREEN: Already passing — confirms baseline.

## Deviation note

The current implementation puts all reader code (DCT + DCA + helpers) in a single 951-line file. The spec's architecture shows \`linguist/reader.go\`. This ticket covers only the DCT decoder and shared helpers. DCA paths are added in T7. adminGrp/transacGrp decoders are added in T8. The file structure stays as \`linguist/reader.go\` — all reader code lives in one file, but the implementation is done incrementally.

## Out of scope

- DCA decoder paths (T7)
- adminGrp/transacGrp decoders (T8)
- Normalization of legacy forms (T9 — the reader calls normalizeStatus/normalizeRegister, but those functions are defined in T9)
- Writer (T10)
- File I/O, dialect dispatch (T12)

## Acceptance

- \`make build\` passes
- LinguistReader satisfies tbx.Reader interface
- Minimal DCT fixture decodes correctly with all fields
- All DCT data categories at concept, langSec, and term level are handled
- Unknown elements are skipped without error
- readCharData and readNoteText handle inline XML


## Notes

**2026-05-25T13:21:58Z**

T6 complete. The reader implementation (reader.go, 952 lines) was already in place from prior work covering DCT+DCA+adminGrp+transacGrp. This task added comprehensive test coverage for the DCT reader per the ticket's TDD cycles: (1) all-categories-dct.tbx fixture exercising every supported data category, (2) TestDecode_AllTermCategories — explicit assertions for all 13+ term-level fields, (3) TestDecode_AllConceptCategories — all concept-level categories (crossRef, externalRef, xGraphic, source, customerSubset, projectSubset, notes), (4) TestDecode_LangSecCategories — langSec-level definition and source, (5) TestDecode_NoteTextInlineXML — verifies readNoteText preserves both Plain (stripped) and Raw (with inline XML), (6) TestDecode_UnknownElementsSkipped — unknown namespaced elements at concept/langSec/term levels are silently skipped. The new fixture also passes round-trip testing. make build passes clean.
