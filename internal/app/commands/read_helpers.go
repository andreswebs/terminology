package commands

import (
	"sort"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/write"
)

// conceptsToResults serializes every concept in g into the canonical
// WriteResult shape, sorted by concept id for deterministic output (per the
// determinism ADR).
func conceptsToResults(g *tbx.Glossary) []output.WriteResult {
	results := make([]output.WriteResult, 0, len(g.Concepts))
	for _, c := range g.Concepts {
		results = append(results, write.ConceptToWriteResult(c))
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].ConceptID < results[j].ConceptID
	})
	return results
}

// restrictLang, when lang is non-empty, trims each result's Languages map to
// only the requested tag. Concepts with no such section keep a stable presence
// with an empty map rather than being dropped.
func restrictLang(results []output.WriteResult, lang string) {
	if lang == "" {
		return
	}
	for i := range results {
		filtered := make(map[string]output.WriteTermGroup, 1)
		if grp, ok := results[i].Languages[lang]; ok {
			filtered[lang] = grp
		}
		results[i].Languages = filtered
	}
}

// reduceToPreferred projects each result down to the fields `list` emits:
// concept_id, subject_field, and per-language preferred term surface only.
func reduceToPreferred(results []output.WriteResult) []output.WriteResult {
	reduced := make([]output.WriteResult, 0, len(results))
	for _, r := range results {
		langs := make(map[string]output.WriteTermGroup, len(r.Languages))
		for tag, grp := range r.Languages {
			var out output.WriteTermGroup
			if grp.Preferred != nil {
				out.Preferred = &output.WriteTerm{Term: grp.Preferred.Term}
			}
			langs[tag] = out
		}
		reduced = append(reduced, output.WriteResult{
			ConceptID:    r.ConceptID,
			SubjectField: r.SubjectField,
			Languages:    langs,
		})
	}
	return reduced
}
