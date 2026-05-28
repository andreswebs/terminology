package linguist

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/tbx/lineindex"
)

const (
	nsMin  = "http://www.tbxinfo.net/ns/min"
	nsBase = "http://www.tbxinfo.net/ns/basic"
	nsLing = "http://www.tbxinfo.net/ns/linguist"
)

type decodeCtx struct {
	dec   *xml.Decoder
	li    *lineindex.Index
	d     dialect
	depth int
}

func (dc *decodeCtx) skip() error {
	depth := 1
	for depth > 0 {
		tok, err := dc.dec.Token()
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			dc.depth++
			depth++
			if dc.depth > maxNestingDepth {
				return fmt.Errorf("XML nesting depth exceeds maximum of %d levels", maxNestingDepth)
			}
		case xml.EndElement:
			dc.depth--
			depth--
		}
	}
	return nil
}

func (dc *decodeCtx) token() (xml.Token, error) {
	tok, err := dc.dec.Token()
	if err != nil {
		return tok, err
	}
	switch tok.(type) {
	case xml.StartElement:
		dc.depth++
		if dc.depth > maxNestingDepth {
			return nil, fmt.Errorf("XML nesting depth exceeds maximum of %d levels", maxNestingDepth)
		}
	case xml.EndElement:
		dc.depth--
	}
	return tok, nil
}

func (dc *decodeCtx) pos() (int, int) {
	if dc.li == nil {
		return 0, 0
	}
	return dc.li.Position(int(dc.dec.InputOffset()))
}

type LinguistReader struct{}

func NewReader() *LinguistReader {
	return &LinguistReader{}
}

const maxNestingDepth = 256

func (lr *LinguistReader) Decode(r io.Reader) (*tbx.Glossary, []tbx.Warning, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("reading input: %w", err)
	}

	if err := tbx.CheckDoctype(data); err != nil {
		return nil, nil, err
	}

	li, err := lineindex.New(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("building line index: %w", err)
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = true
	dc := &decodeCtx{
		dec: dec,
		li:  li,
	}

	var style tbx.Style
	var sourceDesc string
	var warnings []tbx.Warning
	var concepts []tbx.Concept
	var inBody bool

	for {
		tok, err := dc.token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("xml parse: %w", err)
		}

		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		switch {
		case se.Name.Local == "tbx":
			style = detectStyle(se)
			dc.d = dialectFor(style)
		case se.Name.Local == "p":
			text, err := readCharData(dc, se)
			if err != nil {
				return nil, nil, err
			}
			if sourceDesc == "" {
				sourceDesc = text
			}
		case se.Name.Local == "body":
			inBody = true
		case inBody && se.Name.Local == "conceptEntry":
			cLine, cCol := dc.pos()
			c, ws, err := decodeConceptEntry(dc, se)
			if err != nil {
				return nil, nil, err
			}
			c.StartLine, c.StartCol = cLine, cCol
			concepts = append(concepts, c)
			warnings = append(warnings, ws...)
		}
	}

	if !inBody {
		return nil, nil, fmt.Errorf("missing required <text><body> structure")
	}

	g := &tbx.Glossary{
		Dialect:    tbx.DialectLinguist,
		Style:      style,
		SourceDesc: sourceDesc,
		Concepts:   concepts,
	}

	return g, warnings, nil
}

func detectStyle(se xml.StartElement) tbx.Style {
	for _, a := range se.Attr {
		if a.Name.Local == "style" {
			if a.Value == "dca" {
				return tbx.StyleDCA
			}
		}
	}
	return tbx.StyleDCT
}

func decodeConceptEntry(dc *decodeCtx, start xml.StartElement) (tbx.Concept, []tbx.Warning, error) {
	c := tbx.Concept{
		Languages: make(map[string]tbx.LangSection),
	}
	var warnings []tbx.Warning

	for _, a := range start.Attr {
		if a.Name.Local == "id" {
			c.ID = a.Value
		}
	}

	for {
		tok, err := dc.token()
		if err != nil {
			return c, warnings, fmt.Errorf("reading conceptEntry: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			ws, err := decodeConceptChild(dc, t, &c)
			if err != nil {
				return c, warnings, err
			}
			warnings = append(warnings, ws...)

		case xml.EndElement:
			if t.Name.Local == "conceptEntry" {
				for i := range warnings {
					if warnings[i].ConceptID == "" {
						warnings[i].ConceptID = c.ID
					}
				}
				return c, warnings, nil
			}
		}
	}
}

func decodeConceptChild(dc *decodeCtx, se xml.StartElement, c *tbx.Concept) ([]tbx.Warning, error) {
	name := se.Name.Local

	if name == "langSec" {
		lsLine, lsCol := dc.pos()
		return decodeLangSec(dc, se, c, lsLine, lsCol)
	}

	if name == "transacGrp" {
		tx, ws, err := decodeTransacGrp(dc, se)
		if err != nil {
			return nil, err
		}
		c.Transactions = append(c.Transactions, tx)
		return ws, nil
	}

	if name == "note" {
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		c.Notes = append(c.Notes, text)
		return nil, nil
	}

	return decodeConceptFields(dc, se, c)
}

func decodeConceptFields(dc *decodeCtx, se xml.StartElement, c *tbx.Concept) ([]tbx.Warning, error) {
	switch dc.d.conceptKey(se) {
	case "subjectField":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		c.SubjectField = text

	case "definition":
		nt, err := readNoteText(dc, se)
		if err != nil {
			return nil, err
		}
		c.Definitions = append(c.Definitions, nt)

	case "crossReference":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		cr := tbx.CrossRef{Target: text}
		for _, a := range se.Attr {
			if a.Name.Local == "target" {
				cr.Target = a.Value
				cr.Label = text
			}
		}
		c.CrossRefs = append(c.CrossRefs, cr)

	case "externalCrossReference":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		target := text
		for _, a := range se.Attr {
			if a.Name.Local == "target" {
				target = a.Value
			}
		}
		c.ExternalRefs = append(c.ExternalRefs, target)

	case "xGraphic":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		target := text
		for _, a := range se.Attr {
			if a.Name.Local == "target" {
				target = a.Value
			}
		}
		c.Graphics = append(c.Graphics, target)

	case "source":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		c.Sources = append(c.Sources, text)

	case "customerSubset":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		c.CustomerSubset = text

	case "projectSubset":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		c.ProjectSubset = text

	default:
		w := dc.d.warnUnknown(se)
		w.Line, w.Col = dc.pos()
		if err := dc.skip(); err != nil {
			return nil, err
		}
		return []tbx.Warning{w}, nil
	}

	return nil, nil
}

func decodeLangSec(dc *decodeCtx, start xml.StartElement, c *tbx.Concept, startLine, startCol int) ([]tbx.Warning, error) {
	lang := ""
	for _, a := range start.Attr {
		if a.Name.Local == "lang" {
			lang = a.Value
		}
	}

	ls := tbx.LangSection{Lang: lang, StartLine: startLine, StartCol: startCol}
	var warnings []tbx.Warning

	for {
		tok, err := dc.token()
		if err != nil {
			return warnings, fmt.Errorf("reading langSec: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "termSec":
				term, ws, err := decodeTermSec(dc, t)
				if err != nil {
					return warnings, err
				}
				ls.Terms = append(ls.Terms, term)
				warnings = append(warnings, ws...)
			case "transacGrp":
				if err := dc.skip(); err != nil {
					return warnings, err
				}
			default:
				ws, err := decodeLangSecFields(dc, t, &ls)
				if err != nil {
					return warnings, err
				}
				warnings = append(warnings, ws...)
			}

		case xml.EndElement:
			if t.Name.Local == "langSec" {
				c.Languages[lang] = ls
				return warnings, nil
			}
		}
	}
}

func decodeLangSecFields(dc *decodeCtx, se xml.StartElement, ls *tbx.LangSection) ([]tbx.Warning, error) {
	switch dc.d.langSecKey(se) {
	case "definition":
		nt, err := readNoteText(dc, se)
		if err != nil {
			return nil, err
		}
		ls.Definitions = append(ls.Definitions, nt)
	case "source":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		ls.Sources = append(ls.Sources, text)
	default:
		w := dc.d.warnUnknown(se)
		w.Line, w.Col = dc.pos()
		if err := dc.skip(); err != nil {
			return nil, err
		}
		return []tbx.Warning{w}, nil
	}
	return nil, nil
}

func decodeTermSec(dc *decodeCtx, start xml.StartElement) (tbx.Term, []tbx.Warning, error) {
	var term tbx.Term
	var warnings []tbx.Warning

	for {
		tok, err := dc.token()
		if err != nil {
			return term, warnings, fmt.Errorf("reading termSec: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "term":
				text, err := readCharData(dc, t)
				if err != nil {
					return term, warnings, err
				}
				term.Surface = text
			case "adminGrp":
				ws, err := decodeAdminGrp(dc, t, &term)
				if err != nil {
					return term, warnings, err
				}
				warnings = append(warnings, ws...)
			case "transacGrp":
				tx, ws, err := decodeTransacGrp(dc, t)
				if err != nil {
					return term, warnings, err
				}
				term.Transactions = append(term.Transactions, tx)
				warnings = append(warnings, ws...)
			case "note":
				text, err := readCharData(dc, t)
				if err != nil {
					return term, warnings, err
				}
				term.Notes = append(term.Notes, text)
			default:
				ws, err := decodeTermFields(dc, t, &term)
				if err != nil {
					return term, warnings, err
				}
				warnings = append(warnings, ws...)
			}

		case xml.EndElement:
			if t.Name.Local == "termSec" {
				return term, warnings, nil
			}
		}
	}
}

func decodeTermFields(dc *decodeCtx, se xml.StartElement, term *tbx.Term) ([]tbx.Warning, error) {
	switch dc.d.termKey(se) {
	case "administrativeStatus":
		line, col := dc.pos()
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.AdministrativeStatus = normalizeStatus(text)
		var warnings []tbx.Warning
		if w := checkPicklist(text, "administrativeStatus", tbx.AdminStatus()); w != nil {
			w.Line, w.Col = line, col
			warnings = append(warnings, *w)
		}
		if isLegacyStatus(text) {
			warnings = append(warnings, tbx.Warning{
				Code:    "legacy_form_normalized",
				Message: fmt.Sprintf("%q normalized to canonical form", text),
				Line:    line, Col: col,
			})
		}
		if len(warnings) > 0 {
			return warnings, nil
		}

	case "partOfSpeech":
		line, col := dc.pos()
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.PartOfSpeech = text
		if w := checkPicklist(text, "partOfSpeech", tbx.PartOfSpeech()); w != nil {
			w.Line, w.Col = line, col
			return []tbx.Warning{*w}, nil
		}

	case "grammaticalGender":
		line, col := dc.pos()
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.GrammaticalGender = text
		if w := checkPicklist(text, "grammaticalGender", tbx.GrammaticalGender()); w != nil {
			w.Line, w.Col = line, col
			return []tbx.Warning{*w}, nil
		}

	case "grammaticalNumber":
		line, col := dc.pos()
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.GrammaticalNumber = text
		if w := checkPicklist(text, "grammaticalNumber", tbx.GrammaticalNumber()); w != nil {
			w.Line, w.Col = line, col
			return []tbx.Warning{*w}, nil
		}

	case "register":
		line, col := dc.pos()
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.Register = normalizeRegister(text)
		var warnings []tbx.Warning
		if w := checkPicklist(text, "register", tbx.Register()); w != nil {
			w.Line, w.Col = line, col
			warnings = append(warnings, *w)
		}
		if isLegacyRegister(text) {
			warnings = append(warnings, tbx.Warning{
				Code:    "legacy_form_normalized",
				Message: fmt.Sprintf("%q normalized to canonical form", text),
				Line:    line, Col: col,
			})
		}
		if len(warnings) > 0 {
			return warnings, nil
		}

	case "termType":
		line, col := dc.pos()
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.TermType = text
		if w := checkPicklist(text, "termType", tbx.TermType()); w != nil {
			w.Line, w.Col = line, col
			return []tbx.Warning{*w}, nil
		}

	case "termLocation":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.TermLocation = text

	case "geographicalUsage":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.GeographicalUsage = text

	case "context":
		nt, err := readNoteText(dc, se)
		if err != nil {
			return nil, err
		}
		term.Contexts = append(term.Contexts, nt)

	case "transferComment":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.TransferComment = text

	case "source":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.Sources = append(term.Sources, text)

	case "customerSubset":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.CustomerSubset = text

	case "projectSubset":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		term.ProjectSubset = text

	case "externalCrossReference":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		target := text
		for _, a := range se.Attr {
			if a.Name.Local == "target" {
				target = a.Value
			}
		}
		term.ExternalRefs = append(term.ExternalRefs, target)

	case "crossReference":
		text, err := readCharData(dc, se)
		if err != nil {
			return nil, err
		}
		cr := tbx.CrossRef{Target: text}
		for _, a := range se.Attr {
			if a.Name.Local == "target" {
				cr.Target = a.Value
				cr.Label = text
			}
		}
		term.CrossRefs = append(term.CrossRefs, cr)

	default:
		w := dc.d.warnUnknown(se)
		w.Line, w.Col = dc.pos()
		if err := dc.skip(); err != nil {
			return nil, err
		}
		return []tbx.Warning{w}, nil
	}
	return nil, nil
}

func decodeAdminGrp(dc *decodeCtx, start xml.StartElement, term *tbx.Term) ([]tbx.Warning, error) {
	var warnings []tbx.Warning
	for {
		tok, err := dc.token()
		if err != nil {
			return warnings, fmt.Errorf("reading adminGrp: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch dc.d.adminKey(t) {
			case "reading":
				text, err := readCharData(dc, t)
				if err != nil {
					return warnings, err
				}
				term.Reading = text
			case "readingNote":
				text, err := readCharData(dc, t)
				if err != nil {
					return warnings, err
				}
				term.ReadingNote = text
			default:
				w := dc.d.warnUnknown(t)
				w.Line, w.Col = dc.pos()
				warnings = append(warnings, w)
				if err := dc.skip(); err != nil {
					return warnings, err
				}
			}

		case xml.EndElement:
			if t.Name.Local == "adminGrp" {
				return warnings, nil
			}
		}
	}
}

func decodeTransacGrp(dc *decodeCtx, start xml.StartElement) (tbx.Transaction, []tbx.Warning, error) {
	var tx tbx.Transaction
	var warnings []tbx.Warning

	for {
		tok, err := dc.token()
		if err != nil {
			return tx, warnings, fmt.Errorf("reading transacGrp: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "date" {
				text, err := readCharData(dc, t)
				if err != nil {
					return tx, warnings, err
				}
				tx.Date = text
				continue
			}
			switch dc.d.transacKey(t) {
			case "transactionType":
				line, col := dc.pos()
				text, err := readCharData(dc, t)
				if err != nil {
					return tx, warnings, err
				}
				tx.Type = text
				if w := checkPicklist(text, "transactionType", tbx.TransactionType()); w != nil {
					w.Line, w.Col = line, col
					warnings = append(warnings, *w)
				}
			case "responsibility":
				text, err := readCharData(dc, t)
				if err != nil {
					return tx, warnings, err
				}
				tx.Responsibility = text
			default:
				if err := dc.skip(); err != nil {
					return tx, warnings, err
				}
			}

		case xml.EndElement:
			if t.Name.Local == "transacGrp" {
				return tx, warnings, nil
			}
		}
	}
}

func readCharData(dc *decodeCtx, start xml.StartElement) (string, error) {
	var sb strings.Builder
	for {
		tok, err := dc.token()
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", start.Name.Local, err)
		}
		switch t := tok.(type) {
		case xml.CharData:
			sb.Write(t)
		case xml.StartElement:
			inner, err := readCharData(dc, t)
			if err != nil {
				return "", err
			}
			sb.WriteString(inner)
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return strings.TrimSpace(sb.String()), nil
			}
		}
	}
}

func readNoteText(dc *decodeCtx, start xml.StartElement) (tbx.NoteText, error) {
	var plain, raw strings.Builder
	inlineDepth := 0
	for {
		tok, err := dc.token()
		if err != nil {
			return tbx.NoteText{}, fmt.Errorf("reading %s: %w", start.Name.Local, err)
		}
		switch t := tok.(type) {
		case xml.CharData:
			plain.Write(t)
			raw.Write(t)
		case xml.StartElement:
			inlineDepth++
			raw.WriteString("<" + t.Name.Local)
			for _, a := range t.Attr {
				fmt.Fprintf(&raw, ` %s="%s"`, a.Name.Local, a.Value)
			}
			raw.WriteString(">")
		case xml.EndElement:
			if inlineDepth == 0 && t.Name.Local == start.Name.Local {
				return tbx.NoteText{
					Plain: strings.TrimSpace(plain.String()),
					Raw:   strings.TrimSpace(raw.String()),
				}, nil
			}
			if inlineDepth > 0 {
				inlineDepth--
				raw.WriteString("</" + t.Name.Local + ">")
			}
		}
	}
}

func attrVal(se xml.StartElement, name string) string {
	for _, a := range se.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func checkPicklist(value, fieldName string, accepted []string) *tbx.Warning {
	if slices.Contains(accepted, value) {
		return nil
	}
	return &tbx.Warning{
		Code:    "invalid_picklist",
		Message: fmt.Sprintf("%s value %q not in accepted set", fieldName, value),
	}
}

func unknownElementWarning(se xml.StartElement) tbx.Warning {
	return tbx.Warning{
		Code:    "unknown_element",
		Message: fmt.Sprintf("unknown element <%s>", se.Name.Local),
	}
}

func unknownElementWarningDCA(se xml.StartElement) tbx.Warning {
	typ := attrVal(se, "type")
	if typ != "" {
		return tbx.Warning{
			Code:    "unknown_element",
			Message: fmt.Sprintf("unknown element <%s type=%q>", se.Name.Local, typ),
		}
	}
	return tbx.Warning{
		Code:    "unknown_element",
		Message: fmt.Sprintf("unknown element <%s>", se.Name.Local),
	}
}
