package write

import (
	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
)

// ConceptToWriteResult serializes a tbx.Concept into the canonical
// output.WriteResult shape shared by every read and write command. It is the
// single Concept-to-JSON serializer; the inverse is WriteResultToConcept.
//
// Concept-level fields not represented in WriteResult (Graphics,
// CustomerSubset, ProjectSubset) and langSec-level Sources are intentionally
// omitted from the canonical shape; they do not survive a WriteResult
// round-trip.
func ConceptToWriteResult(c tbx.Concept) output.WriteResult {
	r := output.WriteResult{
		ConceptID:    c.ID,
		SubjectField: c.SubjectField,
		Languages:    make(map[string]output.WriteTermGroup, len(c.Languages)),
	}

	for _, d := range c.Definitions {
		r.Definitions = append(r.Definitions, d.Plain)
	}
	for _, cr := range c.CrossRefs {
		r.CrossRefs = append(r.CrossRefs, output.WriteCrossRef{Target: cr.Target, Label: cr.Label})
	}
	r.ExternalRefs = c.ExternalRefs
	r.Sources = c.Sources
	r.Notes = c.Notes

	for tag, ls := range c.Languages {
		var grp output.WriteTermGroup
		for _, d := range ls.Definitions {
			grp.Definitions = append(grp.Definitions, d.Plain)
		}
		for _, t := range ls.Terms {
			wt := TermToWriteTerm(t)
			switch t.AdministrativeStatus {
			case tbx.StatusPreferred:
				grp.Preferred = &wt
			case tbx.StatusAdmitted:
				grp.Admitted = append(grp.Admitted, wt)
			case tbx.StatusDeprecated:
				grp.Deprecated = append(grp.Deprecated, wt)
			case tbx.StatusSuperseded:
				grp.Superseded = append(grp.Superseded, wt)
			default:
				if grp.Preferred == nil {
					grp.Preferred = &wt
				} else {
					grp.Admitted = append(grp.Admitted, wt)
				}
			}
		}
		r.Languages[tag] = grp
	}

	return r
}

// TermToWriteTerm serializes a tbx.Term into the canonical output.WriteTerm
// shape. The inverse is WriteTermToTBXTerm.
func TermToWriteTerm(t tbx.Term) output.WriteTerm {
	wt := output.WriteTerm{
		Term:                 t.Surface,
		AdministrativeStatus: t.AdministrativeStatus.String(),
		PartOfSpeech:         t.PartOfSpeech,
		GrammaticalGender:    t.GrammaticalGender,
		GrammaticalNumber:    t.GrammaticalNumber,
		Register:             t.Register,
		TermType:             t.TermType,
		TermLocation:         t.TermLocation,
		GeographicalUsage:    t.GeographicalUsage,
		TransferComment:      t.TransferComment,
		Reading:              t.Reading,
		ReadingNote:          t.ReadingNote,
		Sources:              t.Sources,
		CustomerSubset:       t.CustomerSubset,
		ProjectSubset:        t.ProjectSubset,
		ExternalRefs:         t.ExternalRefs,
		Notes:                t.Notes,
	}

	for _, ctx := range t.Contexts {
		wt.Contexts = append(wt.Contexts, ctx.Plain)
	}
	for _, cr := range t.CrossRefs {
		wt.CrossRefs = append(wt.CrossRefs, output.WriteCrossRef{Target: cr.Target, Label: cr.Label})
	}

	return wt
}
