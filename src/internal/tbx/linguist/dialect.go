package linguist

import (
	"encoding/xml"

	"github.com/andreswebs/terminology/internal/tbx"
)

// elementKey maps an XML element to a semantic field key, or "" when the
// element is not recognised at the current decoding level.
type elementKey func(xml.StartElement) string

// dialect parameterises the decoder over the two TBX-Linguist styles. DCT
// distinguishes data categories by namespaced element names; DCA distinguishes
// them by a type attribute on generic elements. Each elementKey resolves a
// style-specific element to a shared field key, so the decoder switches on
// field keys instead of forking per style.
type dialect struct {
	conceptKey  elementKey
	langSecKey  elementKey
	termKey     elementKey
	adminKey    elementKey
	transacKey  elementKey
	warnUnknown func(xml.StartElement) tbx.Warning
}

func dialectFor(style tbx.Style) dialect {
	if style == tbx.StyleDCA {
		return dialectDCA
	}
	return dialectDCT
}

var dialectDCT = dialect{
	conceptKey: func(se xml.StartElement) string {
		switch ns, name := se.Name.Space, se.Name.Local; {
		case ns == nsMin && name == "subjectField":
			return "subjectField"
		case ns == nsBase && name == "definition":
			return "definition"
		case ns == nsBase && name == "crossReference":
			return "crossReference"
		case ns == nsMin && name == "externalCrossReference":
			return "externalCrossReference"
		case ns == nsBase && name == "xGraphic":
			return "xGraphic"
		case ns == nsBase && name == "source":
			return "source"
		case ns == nsMin && name == "customerSubset":
			return "customerSubset"
		case ns == nsBase && name == "projectSubset":
			return "projectSubset"
		}
		return ""
	},
	langSecKey: func(se xml.StartElement) string {
		switch ns, name := se.Name.Space, se.Name.Local; {
		case ns == nsBase && name == "definition":
			return "definition"
		case ns == nsBase && name == "source":
			return "source"
		}
		return ""
	},
	termKey: func(se xml.StartElement) string {
		switch ns, name := se.Name.Space, se.Name.Local; {
		case ns == nsMin && name == "administrativeStatus":
			return "administrativeStatus"
		case ns == nsMin && name == "partOfSpeech":
			return "partOfSpeech"
		case ns == nsBase && name == "grammaticalGender":
			return "grammaticalGender"
		case ns == nsLing && name == "grammaticalNumber":
			return "grammaticalNumber"
		case ns == nsLing && name == "register":
			return "register"
		case ns == nsBase && name == "termType":
			return "termType"
		case ns == nsBase && name == "termLocation":
			return "termLocation"
		case ns == nsBase && name == "geographicalUsage":
			return "geographicalUsage"
		case ns == nsBase && name == "context":
			return "context"
		case ns == nsLing && name == "transferComment":
			return "transferComment"
		case ns == nsBase && name == "source":
			return "source"
		case ns == nsMin && name == "customerSubset":
			return "customerSubset"
		case ns == nsBase && name == "projectSubset":
			return "projectSubset"
		case ns == nsMin && name == "externalCrossReference":
			return "externalCrossReference"
		case ns == nsBase && name == "crossReference":
			return "crossReference"
		}
		return ""
	},
	adminKey: func(se xml.StartElement) string {
		switch ns, name := se.Name.Space, se.Name.Local; {
		case ns == nsLing && name == "reading":
			return "reading"
		case ns == nsLing && name == "readingNote":
			return "readingNote"
		}
		return ""
	},
	transacKey: func(se xml.StartElement) string {
		switch ns, name := se.Name.Space, se.Name.Local; {
		case ns == nsBase && name == "transactionType":
			return "transactionType"
		case ns == nsBase && name == "responsibility":
			return "responsibility"
		}
		return ""
	},
	warnUnknown: unknownElementWarning,
}

var dialectDCA = dialect{
	conceptKey: func(se xml.StartElement) string {
		switch name, typ := se.Name.Local, attrVal(se, "type"); {
		case name == "descrip" && typ == "subjectField":
			return "subjectField"
		case name == "descrip" && typ == "definition":
			return "definition"
		case name == "ref" && typ == "crossReference":
			return "crossReference"
		case name == "xref" && typ == "externalCrossReference":
			return "externalCrossReference"
		case name == "xref" && typ == "xGraphic":
			return "xGraphic"
		case name == "admin" && typ == "source":
			return "source"
		case name == "admin" && typ == "customerSubset":
			return "customerSubset"
		case name == "admin" && typ == "projectSubset":
			return "projectSubset"
		}
		return ""
	},
	langSecKey: func(se xml.StartElement) string {
		switch name, typ := se.Name.Local, attrVal(se, "type"); {
		case name == "descrip" && typ == "definition":
			return "definition"
		case name == "admin" && typ == "source":
			return "source"
		}
		return ""
	},
	termKey: func(se xml.StartElement) string {
		switch name, typ := se.Name.Local, attrVal(se, "type"); {
		case name == "termNote" && typ == "administrativeStatus":
			return "administrativeStatus"
		case name == "termNote" && typ == "partOfSpeech":
			return "partOfSpeech"
		case name == "termNote" && typ == "grammaticalGender":
			return "grammaticalGender"
		case name == "termNote" && typ == "grammaticalNumber":
			return "grammaticalNumber"
		case name == "termNote" && typ == "register":
			return "register"
		case name == "termNote" && typ == "termType":
			return "termType"
		case name == "termNote" && typ == "termLocation":
			return "termLocation"
		case name == "termNote" && typ == "geographicalUsage":
			return "geographicalUsage"
		case name == "descrip" && typ == "context":
			return "context"
		case name == "termNote" && typ == "transferComment":
			return "transferComment"
		case name == "admin" && typ == "source":
			return "source"
		case name == "admin" && typ == "customerSubset":
			return "customerSubset"
		case name == "admin" && typ == "projectSubset":
			return "projectSubset"
		case name == "xref" && typ == "externalCrossReference":
			return "externalCrossReference"
		case name == "ref" && typ == "crossReference":
			return "crossReference"
		}
		return ""
	},
	adminKey: func(se xml.StartElement) string {
		switch name, typ := se.Name.Local, attrVal(se, "type"); {
		case name == "admin" && typ == "reading":
			return "reading"
		case name == "adminNote" && typ == "readingNote":
			return "readingNote"
		}
		return ""
	},
	transacKey: func(se xml.StartElement) string {
		switch name, typ := se.Name.Local, attrVal(se, "type"); {
		case name == "transac" && typ == "transactionType":
			return "transactionType"
		case name == "transacNote" && typ == "responsibility":
			return "responsibility"
		}
		return ""
	},
	warnUnknown: unknownElementWarningDCA,
}
