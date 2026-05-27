package linguist

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
)

func TestXmlEscape(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello", "hello"},
		{"a & b", "a &amp; b"},
		{"<tag>", "&lt;tag&gt;"},
		{`say "hi"`, "say &quot;hi&quot;"},
		{`<a & "b">`, "&lt;a &amp; &quot;b&quot;&gt;"},
		{"already &amp; escaped", "already &amp;amp; escaped"},
		{"", ""},
	}

	for _, tc := range tests {
		got := xmlEscape(tc.in)
		if got != tc.want {
			t.Errorf("xmlEscape(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestXmlBuilder(t *testing.T) {
	var buf bytes.Buffer
	b := &xmlBuilder{w: &buf}

	b.line("hello %s", "world")
	b.line("line %d", 2)

	if b.err != nil {
		t.Fatalf("unexpected error: %v", b.err)
	}

	want := "hello world\nline 2\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestXmlBuilder_StickyError(t *testing.T) {
	fw := &failWriter{failAfter: 5}
	b := &xmlBuilder{w: fw}

	b.line("this is a long line that will cause failure")
	if b.err == nil {
		t.Fatal("expected error after write to failing writer")
	}

	firstErr := b.err
	b.line("this should be a no-op")
	if b.err != firstErr {
		t.Error("error changed after subsequent write; should be sticky")
	}
}

type failWriter struct {
	written   int
	failAfter int
}

func (fw *failWriter) Write(p []byte) (int, error) {
	if fw.written+len(p) > fw.failAfter {
		return 0, errors.New("write failed")
	}
	fw.written += len(p)
	return len(p), nil
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status tbx.Status
		want   string
	}{
		{tbx.StatusPreferred, "preferredTerm-admn-sts"},
		{tbx.StatusAdmitted, "admittedTerm-admn-sts"},
		{tbx.StatusDeprecated, "deprecatedTerm-admn-sts"},
		{tbx.StatusSuperseded, "supersededTerm-admn-sts"},
		{tbx.StatusUnspecified, ""},
	}

	for _, tc := range tests {
		got := statusString(tc.status)
		if got != tc.want {
			t.Errorf("statusString(%d) = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestStatusOrder(t *testing.T) {
	if statusOrder(tbx.StatusPreferred) >= statusOrder(tbx.StatusAdmitted) {
		t.Error("preferred should sort before admitted")
	}
	if statusOrder(tbx.StatusAdmitted) >= statusOrder(tbx.StatusDeprecated) {
		t.Error("admitted should sort before deprecated")
	}
	if statusOrder(tbx.StatusDeprecated) >= statusOrder(tbx.StatusSuperseded) {
		t.Error("deprecated should sort before superseded")
	}
	if statusOrder(tbx.StatusSuperseded) >= statusOrder(tbx.StatusUnspecified) {
		t.Error("superseded should sort before unspecified")
	}
}

func TestEncode_MinimalGlossary(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "hello",
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang: "en",
						Terms: []tbx.Term{
							{Surface: "hello", AdministrativeStatus: tbx.StatusPreferred},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `<?xml version="1.0" encoding="UTF-8"?>`)
	assertContains(t, out, `type="TBX-Linguist"`)
	assertContains(t, out, `style="dct"`)
	assertContains(t, out, `<p>Test</p>`)
	assertContains(t, out, `<conceptEntry id="hello">`)
	assertContains(t, out, `<langSec xml:lang="en">`)
	assertContains(t, out, `<term>hello</term>`)
	assertContains(t, out, `preferredTerm-admn-sts`)

	if !strings.HasSuffix(out, "</tbx>\n") {
		t.Error("output should end with </tbx> followed by a single newline")
	}
}

func TestEncode_DefaultSourceDesc(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	assertContains(t, buf.String(), `<p>Terminology glossary</p>`)
}

func TestEncode_ConceptLevelElements(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID:             "test",
				SubjectField:   "philosophy",
				Definitions:    []tbx.NoteText{{Plain: "A test concept"}},
				CrossRefs:      []tbx.CrossRef{{Target: "other", Label: "see also"}},
				ExternalRefs:   []string{"https://example.com/ref"},
				Graphics:       []string{"https://example.com/img.png"},
				Sources:        []string{"Encyclopedia"},
				CustomerSubset: "academic",
				ProjectSubset:  "translations",
				Notes:          []string{"A note"},
				Languages:      map[string]tbx.LangSection{},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `<min:subjectField>philosophy</min:subjectField>`)
	assertContains(t, out, `<basic:definition>A test concept</basic:definition>`)
	assertContains(t, out, `<basic:crossReference target="other">see also</basic:crossReference>`)
	assertContains(t, out, `<min:externalCrossReference target="https://example.com/ref"/>`)
	assertContains(t, out, `<basic:xGraphic target="https://example.com/img.png"/>`)
	assertContains(t, out, `<basic:source>Encyclopedia</basic:source>`)
	assertContains(t, out, `<min:customerSubset>academic</min:customerSubset>`)
	assertContains(t, out, `<basic:projectSubset>translations</basic:projectSubset>`)
	assertContains(t, out, `<note>A note</note>`)
}

func TestEncode_CrossRefWithoutLabel(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID:        "test",
				CrossRefs: []tbx.CrossRef{{Target: "other-concept"}},
				Languages: map[string]tbx.LangSection{},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	assertContains(t, buf.String(), `<basic:crossReference>other-concept</basic:crossReference>`)
}

func TestEncode_TermLevelElements(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang: "en",
						Terms: []tbx.Term{
							{
								Surface:              "completeness",
								AdministrativeStatus: tbx.StatusPreferred,
								PartOfSpeech:         "noun",
								GrammaticalGender:    "neuter",
								GrammaticalNumber:    "singular",
								Register:             "technicalRegister",
								TermType:             "fullForm",
								TermLocation:         "checkBox",
								GeographicalUsage:    "North America",
								Contexts:             []tbx.NoteText{{Plain: "In formal logic"}},
								TransferComment:      "Translate carefully",
								Sources:              []string{"Academic corpus"},
								CustomerSubset:       "term-customer",
								ProjectSubset:        "term-project",
								ExternalRefs:         []string{"https://example.com/term-ref"},
								CrossRefs:            []tbx.CrossRef{{Target: "related", Label: "related term"}},
								Notes:                []string{"Term-level note"},
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `<term>completeness</term>`)
	assertContains(t, out, `<min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>`)
	assertContains(t, out, `<min:partOfSpeech>noun</min:partOfSpeech>`)
	assertContains(t, out, `<basic:grammaticalGender>neuter</basic:grammaticalGender>`)
	assertContains(t, out, `<ling:grammaticalNumber>singular</ling:grammaticalNumber>`)
	assertContains(t, out, `<ling:register>technicalRegister</ling:register>`)
	assertContains(t, out, `<basic:termType>fullForm</basic:termType>`)
	assertContains(t, out, `<basic:termLocation>checkBox</basic:termLocation>`)
	assertContains(t, out, `<basic:geographicalUsage>North America</basic:geographicalUsage>`)
	assertContains(t, out, `<basic:context>In formal logic</basic:context>`)
	assertContains(t, out, `<ling:transferComment>Translate carefully</ling:transferComment>`)
	assertContains(t, out, `<basic:source>Academic corpus</basic:source>`)
	assertContains(t, out, `<min:customerSubset>term-customer</min:customerSubset>`)
	assertContains(t, out, `<basic:projectSubset>term-project</basic:projectSubset>`)
	assertContains(t, out, `<min:externalCrossReference target="https://example.com/term-ref"/>`)
	assertContains(t, out, `<basic:crossReference target="related">related term</basic:crossReference>`)
	assertContains(t, out, `<note>Term-level note</note>`)
}

func TestEncode_AdminGrp(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"he": {
						Lang: "he",
						Terms: []tbx.Term{
							{
								Surface:     "בינה",
								Reading:     "binah",
								ReadingNote: "Ashkenazi pronunciation",
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `<adminGrp>`)
	assertContains(t, out, `<ling:reading>binah</ling:reading>`)
	assertContains(t, out, `<ling:readingNote>Ashkenazi pronunciation</ling:readingNote>`)
	assertContains(t, out, `</adminGrp>`)
}

func TestEncode_AdminGrp_ReadingOnly(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"he": {
						Lang: "he",
						Terms: []tbx.Term{
							{Surface: "בינה", Reading: "binah"},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `<adminGrp>`)
	assertContains(t, out, `<ling:reading>binah</ling:reading>`)
	assertNotContains(t, out, `readingNote`)
}

func TestEncode_NoAdminGrp_WhenEmpty(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang: "en",
						Terms: []tbx.Term{
							{Surface: "hello"},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	assertNotContains(t, buf.String(), `<adminGrp>`)
}

func TestEncode_Transactions(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Transactions: []tbx.Transaction{
					{Type: "origination", Date: "2026-05-21T12:00:00Z", Responsibility: "andre"},
				},
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang: "en",
						Terms: []tbx.Term{
							{
								Surface: "test",
								Transactions: []tbx.Transaction{
									{Type: "modification", Date: "2026-05-22T08:30:00Z", Responsibility: "reviewer"},
								},
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `<transacGrp>`)
	assertContains(t, out, `<basic:transactionType>origination</basic:transactionType>`)
	assertContains(t, out, `<date>2026-05-21T12:00:00Z</date>`)
	assertContains(t, out, `<basic:responsibility>andre</basic:responsibility>`)
	assertContains(t, out, `<basic:transactionType>modification</basic:transactionType>`)
	assertContains(t, out, `<basic:responsibility>reviewer</basic:responsibility>`)
}

func TestEncode_TransactionPartialFields(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Transactions: []tbx.Transaction{
					{Type: "origination"},
				},
				Languages: map[string]tbx.LangSection{},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `<basic:transactionType>origination</basic:transactionType>`)
	assertNotContains(t, out, `<date>`)
	assertNotContains(t, out, `<basic:responsibility>`)
}

func TestSortedConcepts(t *testing.T) {
	cs := []tbx.Concept{
		{ID: "zebra"},
		{ID: "alpha"},
		{ID: "middle"},
	}

	sorted := sortedConcepts(cs)

	if sorted[0].ID != "alpha" || sorted[1].ID != "middle" || sorted[2].ID != "zebra" {
		t.Errorf("got order %s, %s, %s; want alpha, middle, zebra",
			sorted[0].ID, sorted[1].ID, sorted[2].ID)
	}

	if cs[0].ID != "zebra" {
		t.Error("sortedConcepts should not mutate the input slice")
	}
}

func TestSortedLangs(t *testing.T) {
	langs := map[string]tbx.LangSection{
		"he": {Lang: "he"},
		"en": {Lang: "en"},
		"es": {Lang: "es"},
	}

	sorted := sortedLangs(langs)

	if len(sorted) != 3 || sorted[0] != "en" || sorted[1] != "es" || sorted[2] != "he" {
		t.Errorf("got %v, want [en es he]", sorted)
	}
}

func TestSortedTerms_ByStatus(t *testing.T) {
	terms := []tbx.Term{
		{Surface: "deprecated", AdministrativeStatus: tbx.StatusDeprecated},
		{Surface: "preferred", AdministrativeStatus: tbx.StatusPreferred},
		{Surface: "admitted", AdministrativeStatus: tbx.StatusAdmitted},
		{Surface: "unspecified", AdministrativeStatus: tbx.StatusUnspecified},
		{Surface: "superseded", AdministrativeStatus: tbx.StatusSuperseded},
	}

	sorted := sortedTerms(terms)

	wantOrder := []string{"preferred", "admitted", "deprecated", "superseded", "unspecified"}
	for i, want := range wantOrder {
		if sorted[i].Surface != want {
			t.Errorf("position %d: got %q, want %q", i, sorted[i].Surface, want)
		}
	}

	if terms[0].Surface != "deprecated" {
		t.Error("sortedTerms should not mutate the input slice")
	}
}

func TestSortedTerms_StableWithinStatus(t *testing.T) {
	terms := []tbx.Term{
		{Surface: "first-preferred", AdministrativeStatus: tbx.StatusPreferred},
		{Surface: "second-preferred", AdministrativeStatus: tbx.StatusPreferred},
		{Surface: "third-preferred", AdministrativeStatus: tbx.StatusPreferred},
	}

	sorted := sortedTerms(terms)

	if sorted[0].Surface != "first-preferred" ||
		sorted[1].Surface != "second-preferred" ||
		sorted[2].Surface != "third-preferred" {
		t.Errorf("stable sort violated: got %s, %s, %s",
			sorted[0].Surface, sorted[1].Surface, sorted[2].Surface)
	}
}

func TestEncode_ConceptSortOrder(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{ID: "zebra", Languages: map[string]tbx.LangSection{}},
			{ID: "alpha", Languages: map[string]tbx.LangSection{}},
			{ID: "middle", Languages: map[string]tbx.LangSection{}},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	alphaIdx := strings.Index(out, `id="alpha"`)
	middleIdx := strings.Index(out, `id="middle"`)
	zebraIdx := strings.Index(out, `id="zebra"`)

	if alphaIdx > middleIdx || middleIdx > zebraIdx {
		t.Errorf("concepts not sorted by ID: alpha@%d, middle@%d, zebra@%d",
			alphaIdx, middleIdx, zebraIdx)
	}
}

func TestEncode_LanguageSortOrder(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"he": {Lang: "he", Terms: []tbx.Term{{Surface: "בינה"}}},
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "binah"}}},
					"es": {Lang: "es", Terms: []tbx.Term{{Surface: "biná"}}},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	enIdx := strings.Index(out, `xml:lang="en"`)
	esIdx := strings.Index(out, `xml:lang="es"`)
	heIdx := strings.Index(out, `xml:lang="he"`)

	if enIdx > esIdx || esIdx > heIdx {
		t.Errorf("languages not sorted: en@%d, es@%d, he@%d", enIdx, esIdx, heIdx)
	}
}

func TestEncode_TermSortOrderInOutput(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang: "en",
						Terms: []tbx.Term{
							{Surface: "deprecated-form", AdministrativeStatus: tbx.StatusDeprecated},
							{Surface: "preferred-form", AdministrativeStatus: tbx.StatusPreferred},
							{Surface: "admitted-form", AdministrativeStatus: tbx.StatusAdmitted},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	prefIdx := strings.Index(out, "preferred-form")
	admIdx := strings.Index(out, "admitted-form")
	depIdx := strings.Index(out, "deprecated-form")

	if prefIdx > admIdx || admIdx > depIdx {
		t.Errorf("terms not sorted by status: preferred@%d, admitted@%d, deprecated@%d",
			prefIdx, admIdx, depIdx)
	}
}

func TestEncode_SelfClosingElements(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID:           "test",
				ExternalRefs: []string{"https://example.com/ref"},
				Graphics:     []string{"https://example.com/img.png"},
				Languages:    map[string]tbx.LangSection{},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `<min:externalCrossReference target="https://example.com/ref"/>`)
	assertContains(t, out, `<basic:xGraphic target="https://example.com/img.png"/>`)
}

func TestEncode_XmlEscapeInOutput(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "A & B",
		Concepts: []tbx.Concept{
			{
				ID:           "test",
				SubjectField: `"quoted" & <tagged>`,
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang: "en",
						Terms: []tbx.Term{
							{Surface: `a "term" with <special> & chars`},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `<p>A &amp; B</p>`)
	assertContains(t, out, `&quot;quoted&quot; &amp; &lt;tagged&gt;`)
	assertContains(t, out, `a &quot;term&quot; with &lt;special&gt; &amp; chars`)
}

func TestEncode_LangSecDefinitionsAndSources(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang:        "en",
						Definitions: []tbx.NoteText{{Plain: "English-level definition"}},
						Sources:     []string{"Oxford English Dictionary"},
						Terms:       []tbx.Term{{Surface: "test"}},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `<basic:definition>English-level definition</basic:definition>`)
	assertContains(t, out, `<basic:source>Oxford English Dictionary</basic:source>`)
}

func TestEncode_UnspecifiedStatusOmitted(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang:  "en",
						Terms: []tbx.Term{{Surface: "test", AdministrativeStatus: tbx.StatusUnspecified}},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	assertNotContains(t, buf.String(), `administrativeStatus`)
}

func TestEncode_EmptyFieldsOmitted(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang:  "en",
						Terms: []tbx.Term{{Surface: "test"}},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertNotContains(t, out, `partOfSpeech`)
	assertNotContains(t, out, `grammaticalGender`)
	assertNotContains(t, out, `grammaticalNumber`)
	assertNotContains(t, out, `register`)
	assertNotContains(t, out, `termType`)
	assertNotContains(t, out, `termLocation`)
	assertNotContains(t, out, `geographicalUsage`)
	assertNotContains(t, out, `transferComment`)
	assertNotContains(t, out, `subjectField`)
	assertNotContains(t, out, `customerSubset`)
	assertNotContains(t, out, `projectSubset`)
}

func TestEncode_Deterministic(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID:           "beta",
				SubjectField: "test",
				Languages: map[string]tbx.LangSection{
					"he": {Lang: "he", Terms: []tbx.Term{{Surface: "בטא"}}},
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "beta"}}},
				},
			},
			{
				ID: "alpha",
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "alpha"}}},
				},
			},
		},
	}

	w := NewWriter()
	var buf1, buf2 bytes.Buffer
	if err := w.Encode(&buf1, g); err != nil {
		t.Fatalf("first encode: %v", err)
	}
	if err := w.Encode(&buf2, g); err != nil {
		t.Fatalf("second encode: %v", err)
	}

	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		t.Error("two encodes of the same glossary produced different output")
	}
}

func TestEncode_WriterError(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "test"}}},
				},
			},
		},
	}

	fw := &failWriter{failAfter: 10}
	w := NewWriter()
	err := w.Encode(fw, g)
	if err == nil {
		t.Error("expected error from failing writer")
	}
}

func TestEncode_NamespaceDeclarations(t *testing.T) {
	g := &tbx.Glossary{SourceDesc: "Test", Concepts: []tbx.Concept{}}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out := buf.String()
	assertContains(t, out, `xmlns="urn:iso:std:iso:30042:ed-2"`)
	assertContains(t, out, `xmlns:min="http://www.tbxinfo.net/ns/min"`)
	assertContains(t, out, `xmlns:basic="http://www.tbxinfo.net/ns/basic"`)
	assertContains(t, out, `xmlns:ling="http://www.tbxinfo.net/ns/linguist"`)
}

func TestEncode_TwoSpaceIndent(t *testing.T) {
	g := &tbx.Glossary{
		SourceDesc: "Test",
		Concepts: []tbx.Concept{
			{
				ID: "test",
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang:  "en",
						Terms: []tbx.Term{{Surface: "test"}},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	lines := strings.SplitSeq(buf.String(), "\n")
	for line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)
		if indent > 0 && indent%2 != 0 {
			t.Errorf("odd indent (%d spaces) on line: %q", indent, line)
		}
	}
}

func TestEncode_LFLineEndings(t *testing.T) {
	g := &tbx.Glossary{SourceDesc: "Test", Concepts: []tbx.Concept{}}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if strings.Contains(buf.String(), "\r") {
		t.Error("output contains carriage return characters; expected LF only")
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("output does not contain %q", substr)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("output should not contain %q", substr)
	}
}
