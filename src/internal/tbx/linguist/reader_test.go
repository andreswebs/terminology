package linguist_test

import (
	"os"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/tbx/linguist"
)

func TestDecode_MinimalDCT(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/minimal-dct.tbx")

	if g.Dialect != tbx.DialectLinguist {
		t.Errorf("dialect = %q, want %q", g.Dialect, tbx.DialectLinguist)
	}
	if g.Style != tbx.StyleDCT {
		t.Errorf("style = %v, want DCT", g.Style)
	}

	assertMinimalConcept(t, g)
}

func TestDecode_HeaderPreservation(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/header-es-he.tbx")

	if g.SourceLang != "es" {
		t.Errorf("SourceLang = %q, want %q", g.SourceLang, "es")
	}
	if g.Header.Title != "Glosario de cábala" {
		t.Errorf("Header.Title = %q, want %q", g.Header.Title, "Glosario de cábala")
	}
	if len(g.Header.SourceDescs) != 1 || g.Header.SourceDescs[0] != "Glosario es/he para textos académicos" {
		t.Errorf("Header.SourceDescs = %v, want one paragraph", g.Header.SourceDescs)
	}
	if g.SourceDesc != "Glosario es/he para textos académicos" {
		t.Errorf("SourceDesc = %q, want first paragraph", g.SourceDesc)
	}
}

func TestDecode_MinimalDCA(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/minimal-dca.tbx")

	if g.Style != tbx.StyleDCA {
		t.Errorf("style = %v, want DCA", g.Style)
	}

	assertMinimalConcept(t, g)
}

func TestDecode_LegacyFormsNormalized(t *testing.T) {
	f, err := os.Open("testdata/normalized/legacy-forms.tbx")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	r := linguist.NewReader()
	g, warnings, err := r.Decode(f)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(g.Concepts) != 1 {
		t.Fatalf("concepts = %d, want 1", len(g.Concepts))
	}

	en := g.Concepts[0].Languages["en"]
	if len(en.Terms) != 5 {
		t.Fatalf("terms = %d, want 5", len(en.Terms))
	}

	wantStatus := []tbx.Status{
		tbx.StatusPreferred,
		tbx.StatusAdmitted,
		tbx.StatusDeprecated,
		tbx.StatusSuperseded,
		tbx.StatusUnspecified,
	}
	for i, want := range wantStatus {
		if en.Terms[i].AdministrativeStatus != want {
			t.Errorf("term[%d] %q status = %d, want %d",
				i, en.Terms[i].Surface, en.Terms[i].AdministrativeStatus, want)
		}
	}

	if en.Terms[4].Register != "register" {
		t.Errorf("legacy usageRegister not normalized: got %q", en.Terms[4].Register)
	}

	lws := filterWarnings(warnings, "legacy_form_normalized")
	if len(lws) != 5 {
		t.Errorf("legacy_form_normalized warnings = %d, want 5 (4 bare statuses + 1 usageRegister)", len(lws))
	}
}

func decodeFixture(t *testing.T, path string) *tbx.Glossary {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	r := linguist.NewReader()
	g, warnings, err := r.Decode(f)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %d", len(warnings))
	}
	return g
}

func TestDecode_AllTermCategories(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/all-categories-dct.tbx")
	assertAllTermCategories(t, g)
}

func TestDecode_AllTermCategories_DCA(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/all-categories-dca.tbx")
	assertAllTermCategories(t, g)
}

func TestDecode_AllConceptCategories(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/all-categories-dct.tbx")
	assertAllConceptCategories(t, g)
}

func TestDecode_AllConceptCategories_DCA(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/all-categories-dca.tbx")
	assertAllConceptCategories(t, g)
}

func TestDecode_LangSecCategories(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/all-categories-dct.tbx")
	assertLangSecCategories(t, g)
}

func TestDecode_LangSecCategories_DCA(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/all-categories-dca.tbx")
	assertLangSecCategories(t, g)
}

func TestDecode_UnknownElements_DCA(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dca" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <descrip type="subjectField">test</descrip>
      <descrip type="custom">unknown concept field</descrip>
      <langSec xml:lang="en">
        <admin type="unknownAdmin">ignored</admin>
        <termSec>
          <term>test</term>
          <termNote type="unknownCategory">also ignored</termNote>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(g.Concepts) != 1 {
		t.Fatalf("concepts = %d, want 1", len(g.Concepts))
	}
	if g.Concepts[0].SubjectField != "test" {
		t.Errorf("SubjectField = %q, want %q", g.Concepts[0].SubjectField, "test")
	}
	if g.Concepts[0].Languages["en"].Terms[0].Surface != "test" {
		t.Errorf("term Surface = %q, want %q",
			g.Concepts[0].Languages["en"].Terms[0].Surface, "test")
	}

	unknowns := filterWarnings(warnings, "unknown_element")
	if len(unknowns) != 3 {
		t.Fatalf("unknown_element warnings = %d, want 3", len(unknowns))
	}

	for _, w := range unknowns {
		if w.ConceptID != "c1" {
			t.Errorf("warning ConceptID = %q, want %q", w.ConceptID, "c1")
		}
	}

	if !containsSubstring(unknowns[0].Message, "custom") {
		t.Errorf("warning[0].Message = %q, want substring %q", unknowns[0].Message, "custom")
	}
	if !containsSubstring(unknowns[1].Message, "unknownAdmin") {
		t.Errorf("warning[1].Message = %q, want substring %q", unknowns[1].Message, "unknownAdmin")
	}
	if !containsSubstring(unknowns[2].Message, "unknownCategory") {
		t.Errorf("warning[2].Message = %q, want substring %q", unknowns[2].Message, "unknownCategory")
	}
}

func assertAllTermCategories(t *testing.T, g *tbx.Glossary) {
	t.Helper()

	if len(g.Concepts) != 1 {
		t.Fatalf("concepts = %d, want 1", len(g.Concepts))
	}

	en, ok := g.Concepts[0].Languages["en"]
	if !ok {
		t.Fatal("missing language section 'en'")
	}
	if len(en.Terms) != 1 {
		t.Fatalf("terms = %d, want 1", len(en.Terms))
	}

	term := en.Terms[0]

	checks := []struct {
		name string
		got  string
		want string
	}{
		{"Surface", term.Surface, "completeness"},
		{"PartOfSpeech", term.PartOfSpeech, "noun"},
		{"GrammaticalGender", term.GrammaticalGender, "neuter"},
		{"GrammaticalNumber", term.GrammaticalNumber, "singular"},
		{"Register", term.Register, "technicalRegister"},
		{"TermType", term.TermType, "fullForm"},
		{"TermLocation", term.TermLocation, "checkBox"},
		{"GeographicalUsage", term.GeographicalUsage, "North America"},
		{"TransferComment", term.TransferComment, "Translate carefully"},
		{"Reading", term.Reading, "kuhm-pleet-nis"},
		{"ReadingNote", term.ReadingNote, "IPA approximation"},
		{"CustomerSubset", term.CustomerSubset, "term-customer"},
		{"ProjectSubset", term.ProjectSubset, "term-project"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}

	if term.AdministrativeStatus != tbx.StatusPreferred {
		t.Errorf("AdministrativeStatus = %d, want %d", term.AdministrativeStatus, tbx.StatusPreferred)
	}

	if len(term.Contexts) != 1 || term.Contexts[0].Plain != "In the context of formal logic" {
		t.Errorf("Contexts = %v, want [{Plain: 'In the context of formal logic'}]", term.Contexts)
	}
	if len(term.Sources) != 1 || term.Sources[0] != "Academic corpus" {
		t.Errorf("Sources = %v, want [Academic corpus]", term.Sources)
	}
	if len(term.ExternalRefs) != 1 || term.ExternalRefs[0] != "https://example.com/term-ref" {
		t.Errorf("ExternalRefs = %v, want [https://example.com/term-ref]", term.ExternalRefs)
	}
	if len(term.CrossRefs) != 1 {
		t.Fatalf("CrossRefs = %d, want 1", len(term.CrossRefs))
	}
	if term.CrossRefs[0].Target != "related-term" || term.CrossRefs[0].Label != "related term" {
		t.Errorf("CrossRef = {%q, %q}, want {related-term, related term}",
			term.CrossRefs[0].Target, term.CrossRefs[0].Label)
	}
	if len(term.Notes) != 1 || term.Notes[0] != "Term-level note" {
		t.Errorf("Notes = %v, want [Term-level note]", term.Notes)
	}
}

func assertAllConceptCategories(t *testing.T, g *tbx.Glossary) {
	t.Helper()

	c := g.Concepts[0]

	if c.ID != "complete" {
		t.Errorf("ID = %q, want %q", c.ID, "complete")
	}
	if c.SubjectField != "philosophy" {
		t.Errorf("SubjectField = %q, want %q", c.SubjectField, "philosophy")
	}
	if len(c.Definitions) != 1 || c.Definitions[0].Plain != "A concept with all data categories" {
		t.Errorf("Definitions = %v, want plain='A concept with all data categories'", c.Definitions)
	}
	if len(c.CrossRefs) != 1 {
		t.Fatalf("CrossRefs = %d, want 1", len(c.CrossRefs))
	}
	if c.CrossRefs[0].Target != "related" || c.CrossRefs[0].Label != "see also" {
		t.Errorf("CrossRef = {%q, %q}, want {related, see also}",
			c.CrossRefs[0].Target, c.CrossRefs[0].Label)
	}
	if len(c.ExternalRefs) != 1 || c.ExternalRefs[0] != "https://example.com/ref" {
		t.Errorf("ExternalRefs = %v, want [https://example.com/ref]", c.ExternalRefs)
	}
	if len(c.Graphics) != 1 || c.Graphics[0] != "https://example.com/img.png" {
		t.Errorf("Graphics = %v, want [https://example.com/img.png]", c.Graphics)
	}
	if len(c.Sources) != 1 || c.Sources[0] != "Encyclopedia of Philosophy" {
		t.Errorf("Sources = %v, want [Encyclopedia of Philosophy]", c.Sources)
	}
	if c.CustomerSubset != "academic" {
		t.Errorf("CustomerSubset = %q, want %q", c.CustomerSubset, "academic")
	}
	if c.ProjectSubset != "translations" {
		t.Errorf("ProjectSubset = %q, want %q", c.ProjectSubset, "translations")
	}
	if len(c.Notes) != 1 || c.Notes[0] != "General note" {
		t.Errorf("Notes = %v, want [General note]", c.Notes)
	}
}

func assertLangSecCategories(t *testing.T, g *tbx.Glossary) {
	t.Helper()

	en := g.Concepts[0].Languages["en"]

	if len(en.Definitions) != 1 || en.Definitions[0].Plain != "English-level definition" {
		t.Errorf("LangSec Definitions = %v, want plain='English-level definition'", en.Definitions)
	}
	if len(en.Sources) != 1 || en.Sources[0] != "Oxford English Dictionary" {
		t.Errorf("LangSec Sources = %v, want [Oxford English Dictionary]", en.Sources)
	}
}

func TestDecode_NoteTextInlineXML(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:basic="http://www.tbxinfo.net/ns/basic">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <basic:definition>The <hi>bold</hi> concept</basic:definition>
      <langSec xml:lang="en">
        <termSec><term>test</term></termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, _, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(g.Concepts) != 1 {
		t.Fatalf("concepts = %d, want 1", len(g.Concepts))
	}
	defs := g.Concepts[0].Definitions
	if len(defs) != 1 {
		t.Fatalf("definitions = %d, want 1", len(defs))
	}
	if defs[0].Plain != "The bold concept" {
		t.Errorf("Plain = %q, want %q", defs[0].Plain, "The bold concept")
	}
	if defs[0].Raw != "The <hi>bold</hi> concept" {
		t.Errorf("Raw = %q, want %q", defs[0].Raw, "The <hi>bold</hi> concept")
	}
}

func TestDecode_UnknownElements_DCT(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:custom="http://example.com/custom">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <min:subjectField>test</min:subjectField>
      <custom:unknown>some value</custom:unknown>
      <langSec xml:lang="en">
        <custom:extraField>ignored</custom:extraField>
        <termSec>
          <term>test</term>
          <custom:termExtra>also ignored</custom:termExtra>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(g.Concepts) != 1 {
		t.Fatalf("concepts = %d, want 1", len(g.Concepts))
	}
	if g.Concepts[0].SubjectField != "test" {
		t.Errorf("SubjectField = %q, want %q", g.Concepts[0].SubjectField, "test")
	}
	if g.Concepts[0].Languages["en"].Terms[0].Surface != "test" {
		t.Errorf("term Surface = %q, want %q",
			g.Concepts[0].Languages["en"].Terms[0].Surface, "test")
	}

	unknowns := filterWarnings(warnings, "unknown_element")
	if len(unknowns) != 3 {
		t.Fatalf("unknown_element warnings = %d, want 3", len(unknowns))
	}

	for _, w := range unknowns {
		if w.ConceptID != "c1" {
			t.Errorf("warning ConceptID = %q, want %q", w.ConceptID, "c1")
		}
	}

	if !containsSubstring(unknowns[0].Message, "unknown") {
		t.Errorf("warning[0].Message = %q, want substring %q", unknowns[0].Message, "unknown")
	}
	if !containsSubstring(unknowns[1].Message, "extraField") {
		t.Errorf("warning[1].Message = %q, want substring %q", unknowns[1].Message, "extraField")
	}
	if !containsSubstring(unknowns[2].Message, "termExtra") {
		t.Errorf("warning[2].Message = %q, want substring %q", unknowns[2].Message, "termExtra")
	}
}

func TestDecode_TransacGrp_ConceptLevel(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/with-transactions.tbx")

	if len(g.Concepts) != 1 {
		t.Fatalf("concepts = %d, want 1", len(g.Concepts))
	}
	c := g.Concepts[0]

	if len(c.Transactions) != 1 {
		t.Fatalf("concept transactions = %d, want 1", len(c.Transactions))
	}
	tx := c.Transactions[0]

	if tx.Type != "origination" {
		t.Errorf("Type = %q, want %q", tx.Type, "origination")
	}
	if tx.Date != "2026-05-21T12:00:00Z" {
		t.Errorf("Date = %q, want %q", tx.Date, "2026-05-21T12:00:00Z")
	}
	if tx.Responsibility != "andre" {
		t.Errorf("Responsibility = %q, want %q", tx.Responsibility, "andre")
	}
}

func TestDecode_TransacGrp_TermLevel(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/with-transactions.tbx")

	en := g.Concepts[0].Languages["en"]
	if len(en.Terms) < 1 {
		t.Fatal("expected at least 1 term")
	}
	term := en.Terms[0]

	if len(term.Transactions) != 1 {
		t.Fatalf("term transactions = %d, want 1", len(term.Transactions))
	}
	tx := term.Transactions[0]

	if tx.Type != "modification" {
		t.Errorf("Type = %q, want %q", tx.Type, "modification")
	}
	if tx.Date != "2026-05-22T08:30:00Z" {
		t.Errorf("Date = %q, want %q", tx.Date, "2026-05-22T08:30:00Z")
	}
	if tx.Responsibility != "reviewer" {
		t.Errorf("Responsibility = %q, want %q", tx.Responsibility, "reviewer")
	}
}

func TestDecode_AdminGrp_ReadingAndReadingNote(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/full-features.tbx")

	he := g.Concepts[0].Languages["he"]
	if len(he.Terms) != 1 {
		t.Fatalf("he terms = %d, want 1", len(he.Terms))
	}
	term := he.Terms[0]

	if term.Reading != "binah" {
		t.Errorf("Reading = %q, want %q", term.Reading, "binah")
	}
	if term.ReadingNote != "Ashkenazi pronunciation" {
		t.Errorf("ReadingNote = %q, want %q", term.ReadingNote, "Ashkenazi pronunciation")
	}
}

func TestDecode_TransacGrp_DCA(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dca" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <descrip type="subjectField">test</descrip>
      <transacGrp>
        <transac type="transactionType">origination</transac>
        <date>2026-05-20T10:00:00Z</date>
        <transacNote type="responsibility">alice</transacNote>
      </transacGrp>
      <langSec xml:lang="en">
        <termSec>
          <term>test</term>
          <transacGrp>
            <transac type="transactionType">modification</transac>
            <date>2026-05-21T14:00:00Z</date>
            <transacNote type="responsibility">bob</transacNote>
          </transacGrp>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, _, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	c := g.Concepts[0]
	if len(c.Transactions) != 1 {
		t.Fatalf("concept transactions = %d, want 1", len(c.Transactions))
	}
	tx := c.Transactions[0]
	if tx.Type != "origination" {
		t.Errorf("concept tx Type = %q, want %q", tx.Type, "origination")
	}
	if tx.Date != "2026-05-20T10:00:00Z" {
		t.Errorf("concept tx Date = %q, want %q", tx.Date, "2026-05-20T10:00:00Z")
	}
	if tx.Responsibility != "alice" {
		t.Errorf("concept tx Responsibility = %q, want %q", tx.Responsibility, "alice")
	}

	term := c.Languages["en"].Terms[0]
	if len(term.Transactions) != 1 {
		t.Fatalf("term transactions = %d, want 1", len(term.Transactions))
	}
	ttx := term.Transactions[0]
	if ttx.Type != "modification" {
		t.Errorf("term tx Type = %q, want %q", ttx.Type, "modification")
	}
	if ttx.Date != "2026-05-21T14:00:00Z" {
		t.Errorf("term tx Date = %q, want %q", ttx.Date, "2026-05-21T14:00:00Z")
	}
	if ttx.Responsibility != "bob" {
		t.Errorf("term tx Responsibility = %q, want %q", ttx.Responsibility, "bob")
	}
}

func TestDecode_AdminGrp_DCA(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dca" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="he">
        <termSec>
          <term>בינה</term>
          <adminGrp>
            <admin type="reading">binah</admin>
            <adminNote type="readingNote">Ashkenazi pronunciation</adminNote>
          </adminGrp>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, _, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	term := g.Concepts[0].Languages["he"].Terms[0]
	if term.Reading != "binah" {
		t.Errorf("Reading = %q, want %q", term.Reading, "binah")
	}
	if term.ReadingNote != "Ashkenazi pronunciation" {
		t.Errorf("ReadingNote = %q, want %q", term.ReadingNote, "Ashkenazi pronunciation")
	}
}

func TestDecode_UnknownElementsInsideGroups(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <transacGrp>
        <basic:transactionType>origination</basic:transactionType>
        <date>2026-05-20T10:00:00Z</date>
        <basic:responsibility>andre</basic:responsibility>
        <basic:unknownChild>ignored</basic:unknownChild>
      </transacGrp>
      <langSec xml:lang="en">
        <termSec>
          <term>test</term>
          <adminGrp>
            <ling:reading>test-reading</ling:reading>
            <ling:readingNote>test-note</ling:readingNote>
            <ling:unknownField>also ignored</ling:unknownField>
          </adminGrp>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, _, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	c := g.Concepts[0]
	if len(c.Transactions) != 1 {
		t.Fatalf("transactions = %d, want 1", len(c.Transactions))
	}
	if c.Transactions[0].Type != "origination" {
		t.Errorf("tx Type = %q, want %q", c.Transactions[0].Type, "origination")
	}

	term := c.Languages["en"].Terms[0]
	if term.Reading != "test-reading" {
		t.Errorf("Reading = %q, want %q", term.Reading, "test-reading")
	}
	if term.ReadingNote != "test-note" {
		t.Errorf("ReadingNote = %q, want %q", term.ReadingNote, "test-note")
	}
}

func TestDecode_LangSecTransacGrp_Skipped(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <transacGrp>
          <basic:transactionType>modification</basic:transactionType>
          <date>2026-05-20T10:00:00Z</date>
        </transacGrp>
        <termSec>
          <term>test</term>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, _, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(g.Concepts) != 1 {
		t.Fatalf("concepts = %d, want 1", len(g.Concepts))
	}
	if g.Concepts[0].Languages["en"].Terms[0].Surface != "test" {
		t.Errorf("term not parsed after skipped langSec transacGrp")
	}
}

func assertMinimalConcept(t *testing.T, g *tbx.Glossary) {
	t.Helper()

	if len(g.Concepts) != 1 {
		t.Fatalf("concepts = %d, want 1", len(g.Concepts))
	}

	c := g.Concepts[0]
	if c.ID != "tzimtzum" {
		t.Errorf("concept ID = %q, want %q", c.ID, "tzimtzum")
	}
	if c.SubjectField != "kabbalah" {
		t.Errorf("subject field = %q, want %q", c.SubjectField, "kabbalah")
	}

	if len(c.Languages) != 2 {
		t.Fatalf("languages = %d, want 2", len(c.Languages))
	}

	en, ok := c.Languages["en"]
	if !ok {
		t.Fatal("missing language section 'en'")
	}
	if len(en.Terms) != 1 {
		t.Fatalf("en terms = %d, want 1", len(en.Terms))
	}
	if en.Terms[0].Surface != "tzimtzum" {
		t.Errorf("en term = %q, want %q", en.Terms[0].Surface, "tzimtzum")
	}
	if en.Terms[0].AdministrativeStatus != tbx.StatusPreferred {
		t.Errorf("en term status = %d, want %d", en.Terms[0].AdministrativeStatus, tbx.StatusPreferred)
	}
	if en.Terms[0].PartOfSpeech != "noun" {
		t.Errorf("en term POS = %q, want %q", en.Terms[0].PartOfSpeech, "noun")
	}

	he, ok := c.Languages["he"]
	if !ok {
		t.Fatal("missing language section 'he'")
	}
	if len(he.Terms) != 1 {
		t.Fatalf("he terms = %d, want 1", len(he.Terms))
	}
	if he.Terms[0].Surface != "צמצום" {
		t.Errorf("he term = %q, want %q", he.Terms[0].Surface, "צמצום")
	}
}

func TestDecode_UnknownElements_MultipleConcepts(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:custom="http://example.com/custom">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="alpha">
      <min:subjectField>test</min:subjectField>
      <custom:foo>bar</custom:foo>
      <langSec xml:lang="en">
        <termSec>
          <term>alpha</term>
        </termSec>
      </langSec>
    </conceptEntry>
    <conceptEntry id="beta">
      <min:subjectField>test</min:subjectField>
      <langSec xml:lang="en">
        <termSec>
          <term>beta</term>
          <custom:baz>qux</custom:baz>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	_, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	unknowns := filterWarnings(warnings, "unknown_element")
	if len(unknowns) != 2 {
		t.Fatalf("unknown_element warnings = %d, want 2", len(unknowns))
	}

	if unknowns[0].ConceptID != "alpha" {
		t.Errorf("warning[0].ConceptID = %q, want %q", unknowns[0].ConceptID, "alpha")
	}
	if unknowns[1].ConceptID != "beta" {
		t.Errorf("warning[1].ConceptID = %q, want %q", unknowns[1].ConceptID, "beta")
	}
}

func TestDecode_UnknownElements_AdminGrp(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist"
     xmlns:custom="http://example.com/custom">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <min:subjectField>test</min:subjectField>
      <langSec xml:lang="en">
        <termSec>
          <term>test</term>
          <adminGrp>
            <ling:reading>reading</ling:reading>
            <custom:extra>unknown in adminGrp</custom:extra>
          </adminGrp>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	_, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	unknowns := filterWarnings(warnings, "unknown_element")
	if len(unknowns) != 1 {
		t.Fatalf("unknown_element warnings = %d, want 1", len(unknowns))
	}
	if !containsSubstring(unknowns[0].Message, "extra") {
		t.Errorf("warning.Message = %q, want substring %q", unknowns[0].Message, "extra")
	}
	if unknowns[0].ConceptID != "c1" {
		t.Errorf("warning.ConceptID = %q, want %q", unknowns[0].ConceptID, "c1")
	}
}

func TestDecode_UnknownElements_AdminGrp_DCA(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dca" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <descrip type="subjectField">test</descrip>
      <langSec xml:lang="en">
        <termSec>
          <term>test</term>
          <adminGrp>
            <admin type="reading">reading</admin>
            <admin type="customField">unknown in adminGrp</admin>
          </adminGrp>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	_, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	unknowns := filterWarnings(warnings, "unknown_element")
	if len(unknowns) != 1 {
		t.Fatalf("unknown_element warnings = %d, want 1", len(unknowns))
	}
	if !containsSubstring(unknowns[0].Message, "customField") {
		t.Errorf("warning.Message = %q, want substring %q", unknowns[0].Message, "customField")
	}
}

func TestDecode_CleanFile_NoUnknownWarnings(t *testing.T) {
	g := decodeFixture(t, "testdata/canonical/all-categories-dct.tbx")
	if g == nil {
		t.Fatal("expected non-nil glossary")
	}
}

func decodeString(t *testing.T, input string) (*tbx.Glossary, []tbx.Warning) {
	t.Helper()
	r := linguist.NewReader()
	g, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	return g, warnings
}

func TestDecode_InvalidPicklist_PartOfSpeech(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <min:subjectField>test</min:subjectField>
      <langSec xml:lang="en">
        <termSec>
          <term>example</term>
          <min:partOfSpeech>frobnicator</min:partOfSpeech>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	g, warnings := decodeString(t, input)

	if g.Concepts[0].Languages["en"].Terms[0].PartOfSpeech != "frobnicator" {
		t.Error("value should still be stored in model")
	}

	pws := filterWarnings(warnings, "invalid_picklist")
	if len(pws) != 1 {
		t.Fatalf("invalid_picklist warnings = %d, want 1", len(pws))
	}
	if pws[0].ConceptID != "c1" {
		t.Errorf("ConceptID = %q, want %q", pws[0].ConceptID, "c1")
	}
	if !containsSubstring(pws[0].Message, "partOfSpeech") {
		t.Errorf("Message = %q, want substring %q", pws[0].Message, "partOfSpeech")
	}
	if !containsSubstring(pws[0].Message, "frobnicator") {
		t.Errorf("Message = %q, want substring %q", pws[0].Message, "frobnicator")
	}
}

func TestDecode_InvalidPicklist_AdminStatus(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>example</term>
          <min:administrativeStatus>bogus</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	g, warnings := decodeString(t, input)

	if g.Concepts[0].Languages["en"].Terms[0].AdministrativeStatus != tbx.StatusUnspecified {
		t.Error("invalid status should normalize to StatusUnspecified")
	}

	pws := filterWarnings(warnings, "invalid_picklist")
	if len(pws) != 1 {
		t.Fatalf("invalid_picklist warnings = %d, want 1", len(pws))
	}
	if !containsSubstring(pws[0].Message, "administrativeStatus") {
		t.Errorf("Message = %q, want substring %q", pws[0].Message, "administrativeStatus")
	}
}

func TestDecode_InvalidPicklist_DCA(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dca" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>example</term>
          <termNote type="partOfSpeech">frobnicator</termNote>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	g, warnings := decodeString(t, input)

	if g.Concepts[0].Languages["en"].Terms[0].PartOfSpeech != "frobnicator" {
		t.Error("value should still be stored in model")
	}

	pws := filterWarnings(warnings, "invalid_picklist")
	if len(pws) != 1 {
		t.Fatalf("invalid_picklist warnings = %d, want 1", len(pws))
	}
	if pws[0].ConceptID != "c1" {
		t.Errorf("ConceptID = %q, want %q", pws[0].ConceptID, "c1")
	}
}

func TestDecode_InvalidPicklist_Multiple(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>example</term>
          <min:partOfSpeech>badpos</min:partOfSpeech>
          <ling:register>badregister</ling:register>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	_, warnings := decodeString(t, input)

	pws := filterWarnings(warnings, "invalid_picklist")
	if len(pws) != 2 {
		t.Fatalf("invalid_picklist warnings = %d, want 2", len(pws))
	}
	if !containsSubstring(pws[0].Message, "partOfSpeech") {
		t.Errorf("warning[0].Message = %q, want substring %q", pws[0].Message, "partOfSpeech")
	}
	if !containsSubstring(pws[1].Message, "register") {
		t.Errorf("warning[1].Message = %q, want substring %q", pws[1].Message, "register")
	}
}

func TestDecode_InvalidPicklist_LegacyFormNoWarning(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>example</term>
          <min:administrativeStatus>preferredTerm</min:administrativeStatus>
          <ling:register>usageRegister</ling:register>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	g, warnings := decodeString(t, input)

	term := g.Concepts[0].Languages["en"].Terms[0]
	if term.AdministrativeStatus != tbx.StatusPreferred {
		t.Errorf("admin status = %d, want StatusPreferred", term.AdministrativeStatus)
	}
	if term.Register != "register" {
		t.Errorf("register = %q, want %q", term.Register, "register")
	}

	pws := filterWarnings(warnings, "invalid_picklist")
	if len(pws) != 0 {
		t.Errorf("invalid_picklist warnings = %d, want 0 (legacy forms are in picklist)", len(pws))
	}
}

func TestDecode_InvalidPicklist_TransactionType(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:basic="http://www.tbxinfo.net/ns/basic">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <transacGrp>
        <basic:transactionType>badtype</basic:transactionType>
        <date>2026-05-20T10:00:00Z</date>
      </transacGrp>
      <langSec xml:lang="en">
        <termSec>
          <term>example</term>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	g, warnings := decodeString(t, input)

	if g.Concepts[0].Transactions[0].Type != "badtype" {
		t.Error("value should still be stored in model")
	}

	pws := filterWarnings(warnings, "invalid_picklist")
	if len(pws) != 1 {
		t.Fatalf("invalid_picklist warnings = %d, want 1", len(pws))
	}
	if !containsSubstring(pws[0].Message, "transactionType") {
		t.Errorf("Message = %q, want substring %q", pws[0].Message, "transactionType")
	}
}

func TestDecode_InvalidPicklist_Fixture(t *testing.T) {
	f, err := os.Open("testdata/malformed/invalid-picklist.tbx")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	r := linguist.NewReader()
	g, warnings, err := r.Decode(f)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(g.Concepts) != 1 {
		t.Fatalf("concepts = %d, want 1", len(g.Concepts))
	}

	pws := filterWarnings(warnings, "invalid_picklist")
	if len(pws) != 7 {
		t.Fatalf("invalid_picklist warnings = %d, want 7", len(pws))
	}

	fields := []string{
		"transactionType",
		"administrativeStatus",
		"partOfSpeech",
		"grammaticalGender",
		"grammaticalNumber",
		"register",
		"termType",
	}
	for i, field := range fields {
		if !containsSubstring(pws[i].Message, field) {
			t.Errorf("warning[%d].Message = %q, want substring %q", i, pws[i].Message, field)
		}
	}
}

func filterWarnings(ws []tbx.Warning, code string) []tbx.Warning {
	var out []tbx.Warning
	for _, w := range ws {
		if w.Code == code {
			out = append(out, w)
		}
	}
	return out
}

func containsSubstring(s, sub string) bool {
	return strings.Contains(s, sub)
}

func TestDecode_LegacyAdminStatus_EmitsLegacyWarning(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>example</term>
          <min:administrativeStatus>preferredTerm</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if g.Concepts[0].Languages["en"].Terms[0].AdministrativeStatus != tbx.StatusPreferred {
		t.Error("status should still normalize correctly")
	}

	lws := filterWarnings(warnings, "legacy_form_normalized")
	if len(lws) != 1 {
		t.Fatalf("legacy_form_normalized warnings = %d, want 1", len(lws))
	}
	if lws[0].ConceptID != "c1" {
		t.Errorf("ConceptID = %q, want %q", lws[0].ConceptID, "c1")
	}
	if !containsSubstring(lws[0].Message, "preferredTerm") {
		t.Errorf("Message = %q, want substring %q", lws[0].Message, "preferredTerm")
	}
}

func TestDecode_LegacyRegister_EmitsLegacyWarning(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>example</term>
          <ling:register>usageRegister</ling:register>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if g.Concepts[0].Languages["en"].Terms[0].Register != "register" {
		t.Error("register should normalize to canonical form")
	}

	lws := filterWarnings(warnings, "legacy_form_normalized")
	if len(lws) != 1 {
		t.Fatalf("legacy_form_normalized warnings = %d, want 1", len(lws))
	}
	if !containsSubstring(lws[0].Message, "usageRegister") {
		t.Errorf("Message = %q, want substring %q", lws[0].Message, "usageRegister")
	}
}

func TestDecode_LegacyAdminStatus_DCA_EmitsLegacyWarning(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dca" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>example</term>
          <termNote type="administrativeStatus">admittedTerm</termNote>
          <termNote type="register">usageRegister</termNote>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	term := g.Concepts[0].Languages["en"].Terms[0]
	if term.AdministrativeStatus != tbx.StatusAdmitted {
		t.Error("status should normalize correctly in DCA")
	}
	if term.Register != "register" {
		t.Error("register should normalize correctly in DCA")
	}

	lws := filterWarnings(warnings, "legacy_form_normalized")
	if len(lws) != 2 {
		t.Fatalf("legacy_form_normalized warnings = %d, want 2", len(lws))
	}
}

func TestDecode_CanonicalForms_NoLegacyWarning(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>example</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
          <ling:register>technicalRegister</ling:register>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	_, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	lws := filterWarnings(warnings, "legacy_form_normalized")
	if len(lws) != 0 {
		t.Errorf("legacy_form_normalized warnings = %d, want 0", len(lws))
	}
}

func TestDecode_EmptyBody_ReturnsEmptyGlossary(t *testing.T) {
	f, err := os.Open("testdata/malformed/empty-body.tbx")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	r := linguist.NewReader()
	g, warnings, err := r.Decode(f)
	if err != nil {
		t.Fatalf("expected no error for empty body, got: %v", err)
	}
	if len(g.Concepts) != 0 {
		t.Errorf("concepts = %d, want 0", len(g.Concepts))
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %d, want 0", len(warnings))
	}
}

func TestDecode_MalformedXML_ReturnsError(t *testing.T) {
	f, err := os.Open("testdata/malformed/bad-xml.tbx")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	r := linguist.NewReader()
	_, _, err = r.Decode(f)
	if err == nil {
		t.Fatal("expected error for malformed XML, got nil")
	}
}

func TestDecode_MissingBody_ReturnsError(t *testing.T) {
	f, err := os.Open("testdata/malformed/missing-body.tbx")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	r := linguist.NewReader()
	_, _, err = r.Decode(f)
	if err == nil {
		t.Fatal("expected error for missing <text><body>, got nil")
	}
}

func TestDecode_InvalidPicklist_WarningCarriesLineCol(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>test</term>
          <min:partOfSpeech>invalidPOS</min:partOfSpeech>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	_, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	pl := filterWarnings(warnings, "invalid_picklist")
	if len(pl) != 1 {
		t.Fatalf("expected 1 invalid_picklist warning, got %d", len(pl))
	}
	if pl[0].Line == 0 {
		t.Errorf("Line = 0, want > 0")
	}
	if pl[0].Col == 0 {
		t.Errorf("Col = 0, want > 0")
	}
}

func TestDecode_WarningLineAccuracy(t *testing.T) {
	input := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + // line 1
		"<tbx type=\"TBX-Linguist\" style=\"dct\" xml:lang=\"en\"\n" + // line 2
		"     xmlns=\"urn:iso:std:iso:30042:ed-2\"\n" + // line 3
		"     xmlns:min=\"http://www.tbxinfo.net/ns/min\"\n" + // line 4
		"     xmlns:custom=\"http://example.com/custom\">\n" + // line 5
		"  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>\n" + // line 6
		"  <text><body>\n" + // line 7
		"    <conceptEntry id=\"c1\">\n" + // line 8
		"      <min:subjectField>test</min:subjectField>\n" + // line 9
		"      <langSec xml:lang=\"en\">\n" + // line 10
		"        <termSec>\n" + // line 11
		"          <term>test</term>\n" + // line 12
		"          <custom:unknown>value</custom:unknown>\n" + // line 13
		"        </termSec>\n" + // line 14
		"      </langSec>\n" + // line 15
		"    </conceptEntry>\n" + // line 16
		"  </body></text>\n" + // line 17
		"</tbx>"

	r := linguist.NewReader()
	_, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	unk := filterWarnings(warnings, "unknown_element")
	if len(unk) != 1 {
		t.Fatalf("expected 1 unknown_element warning, got %d", len(unk))
	}
	if unk[0].Line != 13 {
		t.Errorf("Line = %d, want 13", unk[0].Line)
	}
}

func TestDecode_MultipleWarnings_DistinctPositions(t *testing.T) {
	input := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<tbx type=\"TBX-Linguist\" style=\"dct\" xml:lang=\"en\"\n" +
		"     xmlns=\"urn:iso:std:iso:30042:ed-2\"\n" +
		"     xmlns:min=\"http://www.tbxinfo.net/ns/min\"\n" +
		"     xmlns:custom=\"http://example.com/custom\">\n" +
		"  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>\n" +
		"  <text><body>\n" +
		"    <conceptEntry id=\"c1\">\n" +
		"      <custom:first>v1</custom:first>\n" + // line 9 — unknown_element
		"      <langSec xml:lang=\"en\">\n" +
		"        <termSec>\n" +
		"          <term>test</term>\n" +
		"          <min:partOfSpeech>badPOS</min:partOfSpeech>\n" + // line 13 — invalid_picklist
		"        </termSec>\n" +
		"      </langSec>\n" +
		"    </conceptEntry>\n" +
		"  </body></text>\n" +
		"</tbx>"

	r := linguist.NewReader()
	_, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(warnings) < 2 {
		t.Fatalf("expected at least 2 warnings, got %d", len(warnings))
	}

	unk := filterWarnings(warnings, "unknown_element")
	pl := filterWarnings(warnings, "invalid_picklist")
	if len(unk) != 1 || len(pl) != 1 {
		t.Fatalf("expected 1 unknown + 1 picklist, got %d + %d", len(unk), len(pl))
	}

	if unk[0].Line == pl[0].Line {
		t.Errorf("warnings on same line %d, expected distinct lines", unk[0].Line)
	}
	if unk[0].Line != 9 {
		t.Errorf("unknown_element Line = %d, want 9", unk[0].Line)
	}
	if pl[0].Line != 13 {
		t.Errorf("invalid_picklist Line = %d, want 13", pl[0].Line)
	}
}

func TestDecode_DCA_WarningCarriesLineCol(t *testing.T) {
	input := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<tbx type=\"TBX-Linguist\" style=\"dca\" xml:lang=\"en\"\n" +
		"     xmlns=\"urn:iso:std:iso:30042:ed-2\">\n" +
		"  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>\n" +
		"  <text><body>\n" +
		"    <conceptEntry id=\"c1\">\n" +
		"      <descrip type=\"subjectField\">test</descrip>\n" +
		"      <descrip type=\"custom\">unknown</descrip>\n" + // line 8 — unknown_element
		"      <langSec xml:lang=\"en\">\n" +
		"        <termSec>\n" +
		"          <term>test</term>\n" +
		"        </termSec>\n" +
		"      </langSec>\n" +
		"    </conceptEntry>\n" +
		"  </body></text>\n" +
		"</tbx>"

	r := linguist.NewReader()
	_, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	unk := filterWarnings(warnings, "unknown_element")
	if len(unk) != 1 {
		t.Fatalf("expected 1 unknown_element warning, got %d", len(unk))
	}
	if unk[0].Line != 8 {
		t.Errorf("Line = %d, want 8", unk[0].Line)
	}
	if unk[0].Col == 0 {
		t.Errorf("Col = 0, want > 0")
	}
}

func TestDecode_UnknownElement_WarningCarriesLineCol(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:custom="http://example.com/custom">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <min:subjectField>test</min:subjectField>
      <custom:unknown>some value</custom:unknown>
      <langSec xml:lang="en">
        <termSec>
          <term>test</term>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	_, warnings, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	unk := filterWarnings(warnings, "unknown_element")
	if len(unk) != 1 {
		t.Fatalf("expected 1 unknown_element warning, got %d", len(unk))
	}
	if unk[0].Line == 0 {
		t.Errorf("Line = 0, want > 0")
	}
	if unk[0].Col == 0 {
		t.Errorf("Col = 0, want > 0")
	}
}

func TestDecode_ConceptAndLangSec_CarrySourcePositions(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader>
    <fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec><term>alpha</term></termSec>
      </langSec>
    </conceptEntry>
    <conceptEntry id="c2">
      <langSec xml:lang="he">
        <termSec><term>beta</term></termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, _, err := r.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(g.Concepts) != 2 {
		t.Fatalf("got %d concepts, want 2", len(g.Concepts))
	}

	c1 := g.Concepts[0]
	if c1.StartLine == 0 {
		t.Error("c1.StartLine = 0, want > 0")
	}

	c2 := g.Concepts[1]
	if c2.StartLine == 0 {
		t.Error("c2.StartLine = 0, want > 0")
	}
	if c2.StartLine <= c1.StartLine {
		t.Errorf("c2.StartLine (%d) should be > c1.StartLine (%d)", c2.StartLine, c1.StartLine)
	}

	enLS := c1.Languages["en"]
	if enLS.StartLine == 0 {
		t.Error("en langSec StartLine = 0, want > 0")
	}
	if enLS.StartLine <= c1.StartLine {
		t.Errorf("en langSec StartLine (%d) should be > c1.StartLine (%d)", enLS.StartLine, c1.StartLine)
	}

	heLS := c2.Languages["he"]
	if heLS.StartLine == 0 {
		t.Error("he langSec StartLine = 0, want > 0")
	}
}

func TestDecode_NestingTooDeep(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min" xmlns:basic="http://www.tbxinfo.net/ns/basic" xmlns:ling="http://www.tbxinfo.net/ns/linguist">`)
	sb.WriteString(`<tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>`)
	sb.WriteString(`<text><body><conceptEntry id="c1"><langSec xml:lang="en"><termSec><term>`)
	for range 260 {
		sb.WriteString("<x>")
	}
	for range 260 {
		sb.WriteString("</x>")
	}
	sb.WriteString(`</term></termSec></langSec></conceptEntry></body></text></tbx>`)

	r := linguist.NewReader()
	_, _, err := r.Decode(strings.NewReader(sb.String()))
	if err == nil {
		t.Fatal("expected nesting depth error, got nil")
	}
	if !strings.Contains(err.Error(), "nesting depth") {
		t.Errorf("expected nesting depth error, got: %v", err)
	}
}

func TestDecode_DoctypeBareTBX(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE tbx>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min" xmlns:basic="http://www.tbxinfo.net/ns/basic" xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en"><termSec><term>hello</term></termSec></langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	g, _, err := r.Decode(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("bare DOCTYPE should be accepted: %v", err)
	}
	if len(g.Concepts) != 1 {
		t.Errorf("concepts = %d, want 1", len(g.Concepts))
	}
}

func TestDecode_DoctypeWithEntitiesRejected(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE tbx [<!ENTITY xxe "boom">]>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en"><termSec><term>hello</term></termSec></langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	_, _, err := r.Decode(strings.NewReader(xml))
	if err == nil {
		t.Fatal("DOCTYPE with entity declarations should be rejected")
	}
}

func TestDecode_DoctypeWithSystemRejected(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE tbx SYSTEM "http://evil.com/xxe.dtd">
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en"><termSec><term>hello</term></termSec></langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	_, _, err := r.Decode(strings.NewReader(xml))
	if err == nil {
		t.Fatal("DOCTYPE with SYSTEM should be rejected")
	}
}

func TestDecode_DoctypeWithPublicRejected(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE tbx PUBLIC "-//TBX//DTD" "http://example.com/tbx.dtd">
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en"><termSec><term>hello</term></termSec></langSec>
    </conceptEntry>
  </body></text>
</tbx>`

	r := linguist.NewReader()
	_, _, err := r.Decode(strings.NewReader(xml))
	if err == nil {
		t.Fatal("DOCTYPE with PUBLIC should be rejected")
	}
}
