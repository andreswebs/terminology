package linguist

import (
	"encoding/xml"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
)

func dctElem(ns, local string) xml.StartElement {
	return xml.StartElement{Name: xml.Name{Space: ns, Local: local}}
}

func dcaElem(local, typ string) xml.StartElement {
	se := xml.StartElement{Name: xml.Name{Local: local}}
	if typ != "" {
		se.Attr = []xml.Attr{{Name: xml.Name{Local: "type"}, Value: typ}}
	}
	return se
}

func TestDialectKeys(t *testing.T) {
	tests := []struct {
		name string
		key  elementKey
		se   xml.StartElement
		want string
	}{
		// Concept level — DCT
		{"dct/concept/subjectField", dialectDCT.conceptKey, dctElem(nsMin, "subjectField"), "subjectField"},
		{"dct/concept/definition", dialectDCT.conceptKey, dctElem(nsBase, "definition"), "definition"},
		{"dct/concept/crossReference", dialectDCT.conceptKey, dctElem(nsBase, "crossReference"), "crossReference"},
		{"dct/concept/externalCrossReference", dialectDCT.conceptKey, dctElem(nsMin, "externalCrossReference"), "externalCrossReference"},
		{"dct/concept/xGraphic", dialectDCT.conceptKey, dctElem(nsBase, "xGraphic"), "xGraphic"},
		{"dct/concept/source", dialectDCT.conceptKey, dctElem(nsBase, "source"), "source"},
		{"dct/concept/customerSubset", dialectDCT.conceptKey, dctElem(nsMin, "customerSubset"), "customerSubset"},
		{"dct/concept/projectSubset", dialectDCT.conceptKey, dctElem(nsBase, "projectSubset"), "projectSubset"},
		{"dct/concept/wrongNamespace", dialectDCT.conceptKey, dctElem(nsBase, "subjectField"), ""},
		{"dct/concept/unknown", dialectDCT.conceptKey, dctElem(nsBase, "bogus"), ""},

		// Concept level — DCA
		{"dca/concept/subjectField", dialectDCA.conceptKey, dcaElem("descrip", "subjectField"), "subjectField"},
		{"dca/concept/definition", dialectDCA.conceptKey, dcaElem("descrip", "definition"), "definition"},
		{"dca/concept/crossReference", dialectDCA.conceptKey, dcaElem("ref", "crossReference"), "crossReference"},
		{"dca/concept/externalCrossReference", dialectDCA.conceptKey, dcaElem("xref", "externalCrossReference"), "externalCrossReference"},
		{"dca/concept/xGraphic", dialectDCA.conceptKey, dcaElem("xref", "xGraphic"), "xGraphic"},
		{"dca/concept/source", dialectDCA.conceptKey, dcaElem("admin", "source"), "source"},
		{"dca/concept/customerSubset", dialectDCA.conceptKey, dcaElem("admin", "customerSubset"), "customerSubset"},
		{"dca/concept/projectSubset", dialectDCA.conceptKey, dcaElem("admin", "projectSubset"), "projectSubset"},
		{"dca/concept/wrongElement", dialectDCA.conceptKey, dcaElem("admin", "subjectField"), ""},
		{"dca/concept/missingType", dialectDCA.conceptKey, dcaElem("descrip", ""), ""},

		// LangSec level
		{"dct/langSec/definition", dialectDCT.langSecKey, dctElem(nsBase, "definition"), "definition"},
		{"dct/langSec/source", dialectDCT.langSecKey, dctElem(nsBase, "source"), "source"},
		{"dct/langSec/unknown", dialectDCT.langSecKey, dctElem(nsMin, "subjectField"), ""},
		{"dca/langSec/definition", dialectDCA.langSecKey, dcaElem("descrip", "definition"), "definition"},
		{"dca/langSec/source", dialectDCA.langSecKey, dcaElem("admin", "source"), "source"},
		{"dca/langSec/unknown", dialectDCA.langSecKey, dcaElem("admin", "customerSubset"), ""},

		// Term level — DCT
		{"dct/term/administrativeStatus", dialectDCT.termKey, dctElem(nsMin, "administrativeStatus"), "administrativeStatus"},
		{"dct/term/partOfSpeech", dialectDCT.termKey, dctElem(nsMin, "partOfSpeech"), "partOfSpeech"},
		{"dct/term/grammaticalGender", dialectDCT.termKey, dctElem(nsBase, "grammaticalGender"), "grammaticalGender"},
		{"dct/term/grammaticalNumber", dialectDCT.termKey, dctElem(nsLing, "grammaticalNumber"), "grammaticalNumber"},
		{"dct/term/register", dialectDCT.termKey, dctElem(nsLing, "register"), "register"},
		{"dct/term/termType", dialectDCT.termKey, dctElem(nsBase, "termType"), "termType"},
		{"dct/term/termLocation", dialectDCT.termKey, dctElem(nsBase, "termLocation"), "termLocation"},
		{"dct/term/geographicalUsage", dialectDCT.termKey, dctElem(nsBase, "geographicalUsage"), "geographicalUsage"},
		{"dct/term/context", dialectDCT.termKey, dctElem(nsBase, "context"), "context"},
		{"dct/term/transferComment", dialectDCT.termKey, dctElem(nsLing, "transferComment"), "transferComment"},
		{"dct/term/source", dialectDCT.termKey, dctElem(nsBase, "source"), "source"},
		{"dct/term/customerSubset", dialectDCT.termKey, dctElem(nsMin, "customerSubset"), "customerSubset"},
		{"dct/term/projectSubset", dialectDCT.termKey, dctElem(nsBase, "projectSubset"), "projectSubset"},
		{"dct/term/externalCrossReference", dialectDCT.termKey, dctElem(nsMin, "externalCrossReference"), "externalCrossReference"},
		{"dct/term/crossReference", dialectDCT.termKey, dctElem(nsBase, "crossReference"), "crossReference"},
		{"dct/term/unknown", dialectDCT.termKey, dctElem(nsBase, "bogus"), ""},

		// Term level — DCA
		{"dca/term/administrativeStatus", dialectDCA.termKey, dcaElem("termNote", "administrativeStatus"), "administrativeStatus"},
		{"dca/term/partOfSpeech", dialectDCA.termKey, dcaElem("termNote", "partOfSpeech"), "partOfSpeech"},
		{"dca/term/grammaticalGender", dialectDCA.termKey, dcaElem("termNote", "grammaticalGender"), "grammaticalGender"},
		{"dca/term/grammaticalNumber", dialectDCA.termKey, dcaElem("termNote", "grammaticalNumber"), "grammaticalNumber"},
		{"dca/term/register", dialectDCA.termKey, dcaElem("termNote", "register"), "register"},
		{"dca/term/termType", dialectDCA.termKey, dcaElem("termNote", "termType"), "termType"},
		{"dca/term/termLocation", dialectDCA.termKey, dcaElem("termNote", "termLocation"), "termLocation"},
		{"dca/term/geographicalUsage", dialectDCA.termKey, dcaElem("termNote", "geographicalUsage"), "geographicalUsage"},
		{"dca/term/context", dialectDCA.termKey, dcaElem("descrip", "context"), "context"},
		{"dca/term/transferComment", dialectDCA.termKey, dcaElem("termNote", "transferComment"), "transferComment"},
		{"dca/term/source", dialectDCA.termKey, dcaElem("admin", "source"), "source"},
		{"dca/term/customerSubset", dialectDCA.termKey, dcaElem("admin", "customerSubset"), "customerSubset"},
		{"dca/term/projectSubset", dialectDCA.termKey, dcaElem("admin", "projectSubset"), "projectSubset"},
		{"dca/term/externalCrossReference", dialectDCA.termKey, dcaElem("xref", "externalCrossReference"), "externalCrossReference"},
		{"dca/term/crossReference", dialectDCA.termKey, dcaElem("ref", "crossReference"), "crossReference"},
		{"dca/term/unknown", dialectDCA.termKey, dcaElem("termNote", "bogus"), ""},

		// AdminGrp level
		{"dct/admin/reading", dialectDCT.adminKey, dctElem(nsLing, "reading"), "reading"},
		{"dct/admin/readingNote", dialectDCT.adminKey, dctElem(nsLing, "readingNote"), "readingNote"},
		{"dct/admin/unknown", dialectDCT.adminKey, dctElem(nsBase, "reading"), ""},
		{"dca/admin/reading", dialectDCA.adminKey, dcaElem("admin", "reading"), "reading"},
		{"dca/admin/readingNote", dialectDCA.adminKey, dcaElem("adminNote", "readingNote"), "readingNote"},
		{"dca/admin/unknown", dialectDCA.adminKey, dcaElem("admin", "readingNote"), ""},

		// TransacGrp level
		{"dct/transac/transactionType", dialectDCT.transacKey, dctElem(nsBase, "transactionType"), "transactionType"},
		{"dct/transac/responsibility", dialectDCT.transacKey, dctElem(nsBase, "responsibility"), "responsibility"},
		{"dct/transac/unknown", dialectDCT.transacKey, dctElem(nsMin, "transactionType"), ""},
		{"dca/transac/transactionType", dialectDCA.transacKey, dcaElem("transac", "transactionType"), "transactionType"},
		{"dca/transac/responsibility", dialectDCA.transacKey, dcaElem("transacNote", "responsibility"), "responsibility"},
		{"dca/transac/unknown", dialectDCA.transacKey, dcaElem("transac", "responsibility"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key(tt.se); got != tt.want {
				t.Errorf("key(%+v) = %q, want %q", tt.se.Name, got, tt.want)
			}
		})
	}
}

func TestDialectFor(t *testing.T) {
	if got := dialectFor(tbx.StyleDCA); got.warnUnknown == nil {
		t.Fatal("dialectFor(DCA) returned zero dialect")
	}
	// DCA surfaces the type attribute in unknown-element warnings.
	dca := dialectFor(tbx.StyleDCA).warnUnknown(dcaElem("descrip", "mystery"))
	if want := `unknown element <descrip type="mystery">`; dca.Message != want {
		t.Errorf("DCA warnUnknown message = %q, want %q", dca.Message, want)
	}
	// DCT never includes a type attribute.
	dct := dialectFor(tbx.StyleDCT).warnUnknown(dctElem(nsBase, "mystery"))
	if want := "unknown element <mystery>"; dct.Message != want {
		t.Errorf("DCT warnUnknown message = %q, want %q", dct.Message, want)
	}
}
