package write

import "github.com/andreswebs/terminology/internal/tbx"

// ReplaceConcept overwrites existing with payload, preserving only the id.
func ReplaceConcept(existing, payload *tbx.Concept) {
	id := existing.ID
	*existing = *payload
	existing.ID = id
}

// MergeConcept overlays the populated fields of payload onto existing: scalar
// fields are replaced when the payload sets them, slice fields when the payload
// is non-empty, and language sections are merged term-by-term.
func MergeConcept(existing, payload *tbx.Concept) {
	if payload.SubjectField != "" {
		existing.SubjectField = payload.SubjectField
	}
	if len(payload.Definitions) > 0 {
		existing.Definitions = payload.Definitions
	}
	if len(payload.CrossRefs) > 0 {
		existing.CrossRefs = payload.CrossRefs
	}
	if len(payload.ExternalRefs) > 0 {
		existing.ExternalRefs = payload.ExternalRefs
	}
	if len(payload.Graphics) > 0 {
		existing.Graphics = payload.Graphics
	}
	if len(payload.Sources) > 0 {
		existing.Sources = payload.Sources
	}
	if payload.CustomerSubset != "" {
		existing.CustomerSubset = payload.CustomerSubset
	}
	if payload.ProjectSubset != "" {
		existing.ProjectSubset = payload.ProjectSubset
	}
	if len(payload.Notes) > 0 {
		existing.Notes = payload.Notes
	}

	if existing.Languages == nil {
		existing.Languages = make(map[string]tbx.LangSection)
	}

	for lang, payloadLS := range payload.Languages {
		existingLS, ok := existing.Languages[lang]
		if !ok {
			existing.Languages[lang] = payloadLS
			continue
		}
		mergeLangSection(&existingLS, &payloadLS)
		existing.Languages[lang] = existingLS
	}
}

func mergeLangSection(existing, payload *tbx.LangSection) {
	if len(payload.Definitions) > 0 {
		existing.Definitions = payload.Definitions
	}
	if len(payload.Sources) > 0 {
		existing.Sources = payload.Sources
	}

	for _, pt := range payload.Terms {
		merged := false
		for i := range existing.Terms {
			if existing.Terms[i].Surface == pt.Surface &&
				existing.Terms[i].AdministrativeStatus == pt.AdministrativeStatus {
				mergeTermFields(&existing.Terms[i], &pt)
				merged = true
				break
			}
		}
		if !merged {
			existing.Terms = append(existing.Terms, pt)
		}
	}
}

func mergeTermFields(existing, payload *tbx.Term) {
	if payload.PartOfSpeech != "" {
		existing.PartOfSpeech = payload.PartOfSpeech
	}
	if payload.GrammaticalGender != "" {
		existing.GrammaticalGender = payload.GrammaticalGender
	}
	if payload.GrammaticalNumber != "" {
		existing.GrammaticalNumber = payload.GrammaticalNumber
	}
	if payload.Register != "" {
		existing.Register = payload.Register
	}
	if payload.TermType != "" {
		existing.TermType = payload.TermType
	}
	if payload.TermLocation != "" {
		existing.TermLocation = payload.TermLocation
	}
	if payload.GeographicalUsage != "" {
		existing.GeographicalUsage = payload.GeographicalUsage
	}
	if payload.TransferComment != "" {
		existing.TransferComment = payload.TransferComment
	}
	if payload.Reading != "" {
		existing.Reading = payload.Reading
	}
	if payload.ReadingNote != "" {
		existing.ReadingNote = payload.ReadingNote
	}
	if payload.CustomerSubset != "" {
		existing.CustomerSubset = payload.CustomerSubset
	}
	if payload.ProjectSubset != "" {
		existing.ProjectSubset = payload.ProjectSubset
	}
	if len(payload.Contexts) > 0 {
		existing.Contexts = payload.Contexts
	}
	if len(payload.Sources) > 0 {
		existing.Sources = payload.Sources
	}
	if len(payload.ExternalRefs) > 0 {
		existing.ExternalRefs = payload.ExternalRefs
	}
	if len(payload.CrossRefs) > 0 {
		existing.CrossRefs = payload.CrossRefs
	}
	if len(payload.Notes) > 0 {
		existing.Notes = payload.Notes
	}
}
