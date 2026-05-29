package linguist

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/andreswebs/terminology/internal/tbx"
)

type LinguistWriter struct{}

func NewWriter() *LinguistWriter {
	return &LinguistWriter{}
}

func (lw *LinguistWriter) Encode(w io.Writer, g *tbx.Glossary) error {
	b := &xmlBuilder{w: w}

	lang := g.SourceLang
	if lang == "" {
		lang = "en"
	}

	b.line(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.line(`<tbx type="TBX-Linguist" style="dct" xml:lang="%s" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min" xmlns:basic="http://www.tbxinfo.net/ns/basic" xmlns:ling="http://www.tbxinfo.net/ns/linguist">`, xmlEscape(lang))
	writeHeader(b, g)
	b.line(`  <text>`)
	b.line(`    <body>`)

	concepts := sortedConcepts(g.Concepts)
	for _, c := range concepts {
		writeConceptEntry(b, c)
	}

	b.line(`    </body>`)
	b.line(`  </text>`)
	b.line(`</tbx>`)

	return b.err
}

func writeHeader(b *xmlBuilder, g *tbx.Glossary) {
	h := g.Header
	sourceDescs := h.SourceDescs
	if len(sourceDescs) == 0 {
		desc := g.SourceDesc
		if desc == "" {
			desc = "Terminology glossary"
		}
		sourceDescs = []string{desc}
	}

	b.line(`  <tbxHeader>`)
	b.line(`    <fileDesc>`)
	if h.Title != "" {
		b.line(`      <titleStmt>`)
		b.line(`        <title>%s</title>`, xmlEscape(h.Title))
		b.line(`      </titleStmt>`)
	}
	if len(h.PublicationStmts) > 0 {
		b.line(`      <publicationStmt>`)
		for _, p := range h.PublicationStmts {
			b.line(`        <p>%s</p>`, xmlEscape(p))
		}
		b.line(`      </publicationStmt>`)
	}
	b.line(`      <sourceDesc>`)
	for _, p := range sourceDescs {
		b.line(`        <p>%s</p>`, xmlEscape(p))
	}
	b.line(`      </sourceDesc>`)
	if len(h.EncodingDescs) > 0 {
		b.line(`      <encodingDesc>`)
		for _, p := range h.EncodingDescs {
			b.line(`        <p>%s</p>`, xmlEscape(p))
		}
		b.line(`      </encodingDesc>`)
	}
	if len(h.RevisionDescs) > 0 {
		b.line(`      <revisionDesc>`)
		for _, p := range h.RevisionDescs {
			b.line(`        <p>%s</p>`, xmlEscape(p))
		}
		b.line(`      </revisionDesc>`)
	}
	b.line(`    </fileDesc>`)
	b.line(`  </tbxHeader>`)
}

func writeConceptEntry(b *xmlBuilder, c tbx.Concept) {
	b.line(`      <conceptEntry id="%s">`, c.ID)

	if c.SubjectField != "" {
		b.line(`        <min:subjectField>%s</min:subjectField>`, xmlEscape(c.SubjectField))
	}
	for _, d := range c.Definitions {
		b.line(`        <basic:definition>%s</basic:definition>`, xmlEscape(d.Plain))
	}
	for _, cr := range c.CrossRefs {
		if cr.Label != "" {
			b.line(`        <basic:crossReference target="%s">%s</basic:crossReference>`, xmlEscape(cr.Target), xmlEscape(cr.Label))
		} else {
			b.line(`        <basic:crossReference>%s</basic:crossReference>`, xmlEscape(cr.Target))
		}
	}
	for _, er := range c.ExternalRefs {
		b.line(`        <min:externalCrossReference target="%s"/>`, xmlEscape(er))
	}
	for _, gr := range c.Graphics {
		b.line(`        <basic:xGraphic target="%s"/>`, xmlEscape(gr))
	}
	for _, s := range c.Sources {
		b.line(`        <basic:source>%s</basic:source>`, xmlEscape(s))
	}
	if c.CustomerSubset != "" {
		b.line(`        <min:customerSubset>%s</min:customerSubset>`, xmlEscape(c.CustomerSubset))
	}
	if c.ProjectSubset != "" {
		b.line(`        <basic:projectSubset>%s</basic:projectSubset>`, xmlEscape(c.ProjectSubset))
	}
	for _, n := range c.Notes {
		b.line(`        <note>%s</note>`, xmlEscape(n))
	}
	for _, tx := range c.Transactions {
		writeTransaction(b, tx, 8)
	}

	langs := sortedLangs(c.Languages)
	for _, lang := range langs {
		ls := c.Languages[lang]
		writeLangSec(b, ls)
	}

	b.line(`      </conceptEntry>`)
}

func writeLangSec(b *xmlBuilder, ls tbx.LangSection) {
	b.line(`        <langSec xml:lang="%s">`, ls.Lang)

	for _, d := range ls.Definitions {
		b.line(`          <basic:definition>%s</basic:definition>`, xmlEscape(d.Plain))
	}
	for _, s := range ls.Sources {
		b.line(`          <basic:source>%s</basic:source>`, xmlEscape(s))
	}

	terms := sortedTerms(ls.Terms)
	for _, term := range terms {
		writeTermSec(b, term)
	}

	b.line(`        </langSec>`)
}

func writeTermSec(b *xmlBuilder, term tbx.Term) {
	b.line(`          <termSec>`)
	b.line(`            <term>%s</term>`, xmlEscape(term.Surface))

	if term.AdministrativeStatus != tbx.StatusUnspecified {
		b.line(`            <min:administrativeStatus>%s</min:administrativeStatus>`, term.AdministrativeStatus.String())
	}
	if term.PartOfSpeech != "" {
		b.line(`            <min:partOfSpeech>%s</min:partOfSpeech>`, xmlEscape(term.PartOfSpeech))
	}
	if term.GrammaticalGender != "" {
		b.line(`            <basic:grammaticalGender>%s</basic:grammaticalGender>`, xmlEscape(term.GrammaticalGender))
	}
	if term.GrammaticalNumber != "" {
		b.line(`            <ling:grammaticalNumber>%s</ling:grammaticalNumber>`, xmlEscape(term.GrammaticalNumber))
	}
	if term.Register != "" {
		b.line(`            <ling:register>%s</ling:register>`, xmlEscape(term.Register))
	}
	if term.TermType != "" {
		b.line(`            <basic:termType>%s</basic:termType>`, xmlEscape(term.TermType))
	}
	if term.TermLocation != "" {
		b.line(`            <basic:termLocation>%s</basic:termLocation>`, xmlEscape(term.TermLocation))
	}
	if term.GeographicalUsage != "" {
		b.line(`            <basic:geographicalUsage>%s</basic:geographicalUsage>`, xmlEscape(term.GeographicalUsage))
	}
	for _, ctx := range term.Contexts {
		b.line(`            <basic:context>%s</basic:context>`, xmlEscape(ctx.Plain))
	}
	if term.TransferComment != "" {
		b.line(`            <ling:transferComment>%s</ling:transferComment>`, xmlEscape(term.TransferComment))
	}
	if term.Reading != "" || term.ReadingNote != "" {
		b.line(`            <adminGrp>`)
		if term.Reading != "" {
			b.line(`              <ling:reading>%s</ling:reading>`, xmlEscape(term.Reading))
		}
		if term.ReadingNote != "" {
			b.line(`              <ling:readingNote>%s</ling:readingNote>`, xmlEscape(term.ReadingNote))
		}
		b.line(`            </adminGrp>`)
	}
	for _, s := range term.Sources {
		b.line(`            <basic:source>%s</basic:source>`, xmlEscape(s))
	}
	if term.CustomerSubset != "" {
		b.line(`            <min:customerSubset>%s</min:customerSubset>`, xmlEscape(term.CustomerSubset))
	}
	if term.ProjectSubset != "" {
		b.line(`            <basic:projectSubset>%s</basic:projectSubset>`, xmlEscape(term.ProjectSubset))
	}
	for _, er := range term.ExternalRefs {
		b.line(`            <min:externalCrossReference target="%s"/>`, xmlEscape(er))
	}
	for _, cr := range term.CrossRefs {
		if cr.Label != "" {
			b.line(`            <basic:crossReference target="%s">%s</basic:crossReference>`, xmlEscape(cr.Target), xmlEscape(cr.Label))
		} else {
			b.line(`            <basic:crossReference>%s</basic:crossReference>`, xmlEscape(cr.Target))
		}
	}
	for _, n := range term.Notes {
		b.line(`            <note>%s</note>`, xmlEscape(n))
	}
	for _, tx := range term.Transactions {
		writeTransaction(b, tx, 12)
	}

	b.line(`          </termSec>`)
}

func writeTransaction(b *xmlBuilder, tx tbx.Transaction, indent int) {
	pad := strings.Repeat(" ", indent)
	b.line(`%s<transacGrp>`, pad)
	if tx.Type != "" {
		b.line(`%s  <basic:transactionType>%s</basic:transactionType>`, pad, xmlEscape(tx.Type))
	}
	if tx.Date != "" {
		b.line(`%s  <date>%s</date>`, pad, xmlEscape(tx.Date))
	}
	if tx.Responsibility != "" {
		b.line(`%s  <basic:responsibility>%s</basic:responsibility>`, pad, xmlEscape(tx.Responsibility))
	}
	b.line(`%s</transacGrp>`, pad)
}

func sortedConcepts(cs []tbx.Concept) []tbx.Concept {
	sorted := make([]tbx.Concept, len(cs))
	copy(sorted, cs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

func sortedLangs(langs map[string]tbx.LangSection) []string {
	keys := make([]string, 0, len(langs))
	for k := range langs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedTerms(terms []tbx.Term) []tbx.Term {
	sorted := make([]tbx.Term, len(terms))
	copy(sorted, terms)
	sort.SliceStable(sorted, func(i, j int) bool {
		return statusOrder(sorted[i].AdministrativeStatus) < statusOrder(sorted[j].AdministrativeStatus)
	})
	return sorted
}

func statusOrder(s tbx.Status) int {
	switch s {
	case tbx.StatusPreferred:
		return 0
	case tbx.StatusAdmitted:
		return 1
	case tbx.StatusDeprecated:
		return 2
	case tbx.StatusSuperseded:
		return 3
	default:
		return 4
	}
}

type xmlBuilder struct {
	w   io.Writer
	err error
}

func (b *xmlBuilder) line(format string, args ...any) {
	if b.err != nil {
		return
	}
	s := fmt.Sprintf(format, args...)
	_, b.err = fmt.Fprintln(b.w, s)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
