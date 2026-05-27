package check

import (
	"testing"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
)

func glossaryWith(concepts ...tbx.Concept) *tbx.Glossary {
	return &tbx.Glossary{Concepts: concepts}
}

func concept(id string, langs map[string]tbx.LangSection) tbx.Concept {
	return tbx.Concept{ID: id, Languages: langs}
}

func langSec(lang string, terms ...tbx.Term) tbx.LangSection {
	return tbx.LangSection{Lang: lang, Terms: terms}
}

func term(surface string, status tbx.Status) tbx.Term {
	return tbx.Term{Surface: surface, AdministrativeStatus: status}
}

func TestCheck_NoViolations(t *testing.T) {
	g := glossaryWith(
		concept("tzimtzum", map[string]tbx.LangSection{
			"es": langSec("es", term("tzimtzum", tbx.StatusPreferred)),
			"he": langSec("he", term("צמצום", tbx.StatusPreferred)),
		}),
	)

	src := []byte("El concepto de tzimtzum es central.")
	tgt := []byte("המושג של צמצום הוא מרכזי.")

	result, err := Check(g, src, tgt, "es", "he", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations, got %d: %+v", len(result.Violations), result.Violations)
	}
	if result.ConceptsChecked != 1 {
		t.Errorf("expected 1 concept checked, got %d", result.ConceptsChecked)
	}
}

func TestCheck_MissingViolation(t *testing.T) {
	g := glossaryWith(
		concept("tzimtzum", map[string]tbx.LangSection{
			"es": langSec("es", term("tzimtzum", tbx.StatusPreferred)),
			"he": langSec("he", term("צמצום", tbx.StatusPreferred)),
		}),
	)

	src := []byte("tzimtzum aparece tres veces: tzimtzum y tzimtzum.")
	tgt := []byte("הטקסט הזה לא מכיל את המונח.")

	result, err := Check(g, src, tgt, "es", "he", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}

	v := result.Violations[0]
	if v.Type != "missing" {
		t.Errorf("expected type missing, got %s", v.Type)
	}
	if v.ConceptID != "tzimtzum" {
		t.Errorf("expected concept_id tzimtzum, got %s", v.ConceptID)
	}
	if v.SourceTerm != "tzimtzum" {
		t.Errorf("expected source_term tzimtzum, got %s", v.SourceTerm)
	}
	if v.ExpectedTarget != "צמצום" {
		t.Errorf("expected expected_target צמצום, got %s", v.ExpectedTarget)
	}
	if v.SourceOccurrences != 3 {
		t.Errorf("expected 3 source occurrences, got %d", v.SourceOccurrences)
	}
}

func TestCheck_ForbiddenVariant(t *testing.T) {
	g := glossaryWith(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("contraction", tbx.StatusPreferred)),
			"es": langSec("es",
				term("contracción", tbx.StatusPreferred),
				term("reducción", tbx.StatusDeprecated),
			),
		}),
	)

	src := []byte("The contraction of the infinite.")
	tgt := []byte("La reducción del infinito.")

	result, err := Check(g, src, tgt, "en", "es", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var missing, forbidden int
	for _, v := range result.Violations {
		switch v.Type {
		case "missing":
			missing++
		case "forbidden_variant":
			forbidden++
			if v.ConceptID != "c001" {
				t.Errorf("expected concept_id c001, got %s", v.ConceptID)
			}
			if v.Variant != "reducción" {
				t.Errorf("expected variant reducción, got %s", v.Variant)
			}
			if v.Line == 0 {
				t.Error("expected non-zero line for forbidden_variant")
			}
			if v.Column == 0 {
				t.Error("expected non-zero column for forbidden_variant")
			}
		}
	}

	if missing != 1 {
		t.Errorf("expected 1 missing violation, got %d", missing)
	}
	if forbidden != 1 {
		t.Errorf("expected 1 forbidden_variant violation, got %d", forbidden)
	}
}

func TestCheck_ConceptAbsentFromSource(t *testing.T) {
	g := glossaryWith(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("soul", tbx.StatusPreferred)),
			"es": langSec("es", term("alma", tbx.StatusPreferred)),
		}),
	)

	src := []byte("This text mentions nothing from the glossary.")
	tgt := []byte("Este texto no menciona nada del glosario.")

	result, err := Check(g, src, tgt, "en", "es", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(result.Violations))
	}
	if result.ConceptsChecked != 0 {
		t.Errorf("expected 0 concepts checked, got %d", result.ConceptsChecked)
	}
}

func TestCheck_MultipleConcepts(t *testing.T) {
	g := glossaryWith(
		concept("soul", map[string]tbx.LangSection{
			"en": langSec("en", term("soul", tbx.StatusPreferred)),
			"es": langSec("es", term("alma", tbx.StatusPreferred)),
		}),
		concept("light", map[string]tbx.LangSection{
			"en": langSec("en", term("light", tbx.StatusPreferred)),
			"es": langSec("es",
				term("luz", tbx.StatusPreferred),
				term("lumbre", tbx.StatusDeprecated),
			),
		}),
		concept("vessel", map[string]tbx.LangSection{
			"en": langSec("en", term("vessel", tbx.StatusPreferred)),
			"es": langSec("es", term("vasija", tbx.StatusPreferred)),
		}),
	)

	src := []byte("The soul receives light through the vessel.")
	tgt := []byte("El alma recibe lumbre a través de la vasija.")

	result, err := Check(g, src, tgt, "en", "es", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ConceptsChecked != 3 {
		t.Errorf("expected 3 concepts checked, got %d", result.ConceptsChecked)
	}

	violations := make(map[string][]output.CheckViolation)
	for _, v := range result.Violations {
		violations[v.Type] = append(violations[v.Type], v)
	}

	if len(violations["missing"]) != 1 {
		t.Errorf("expected 1 missing violation, got %d", len(violations["missing"]))
	} else if violations["missing"][0].ConceptID != "light" {
		t.Errorf("expected missing for concept light, got %s", violations["missing"][0].ConceptID)
	}

	if len(violations["forbidden_variant"]) != 1 {
		t.Errorf("expected 1 forbidden_variant, got %d", len(violations["forbidden_variant"]))
	} else if violations["forbidden_variant"][0].Variant != "lumbre" {
		t.Errorf("expected variant lumbre, got %s", violations["forbidden_variant"][0].Variant)
	}
}

func TestCheck_CodeBlocksExcluded(t *testing.T) {
	g := glossaryWith(
		concept("tzimtzum", map[string]tbx.LangSection{
			"es": langSec("es", term("tzimtzum", tbx.StatusPreferred)),
			"he": langSec("he", term("צמצום", tbx.StatusPreferred)),
		}),
	)

	src := []byte("El tzimtzum es central.\n")
	tgt := []byte("```\nצמצום\n```\n")

	result, err := Check(g, src, tgt, "es", "he", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation (missing), got %d", len(result.Violations))
	}
	if result.Violations[0].Type != "missing" {
		t.Errorf("expected missing violation, got %s", result.Violations[0].Type)
	}
}

func TestCheck_ContextWindow(t *testing.T) {
	g := glossaryWith(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("contraction", tbx.StatusPreferred)),
			"es": langSec("es",
				term("contracción", tbx.StatusPreferred),
				term("reducción", tbx.StatusDeprecated),
			),
		}),
	)

	src := []byte("The contraction of the infinite.")
	tgt := []byte("La reducción del infinito.")

	result, err := Check(g, src, tgt, "en", "es", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, v := range result.Violations {
		if v.Type == "forbidden_variant" && v.Context == "" {
			t.Error("expected non-empty context for forbidden_variant")
		}
	}
}

func TestCheck_UnspecifiedStatusTreatedAsPreferred(t *testing.T) {
	g := glossaryWith(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("word", tbx.StatusUnspecified)),
			"es": langSec("es", term("palabra", tbx.StatusUnspecified)),
		}),
	)

	src := []byte("The word is here.")
	tgt := []byte("La palabra está aquí.")

	result, err := Check(g, src, tgt, "en", "es", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(result.Violations))
	}
	if result.ConceptsChecked != 1 {
		t.Errorf("expected 1 concept checked, got %d", result.ConceptsChecked)
	}
}

func TestCheck_SupersededVariant(t *testing.T) {
	g := glossaryWith(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("word", tbx.StatusPreferred)),
			"es": langSec("es",
				term("palabra", tbx.StatusPreferred),
				term("vocablo", tbx.StatusSuperseded),
			),
		}),
	)

	src := []byte("The word is here.")
	tgt := []byte("El vocablo está aquí.")

	result, err := Check(g, src, tgt, "en", "es", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var forbidden int
	for _, v := range result.Violations {
		if v.Type == "forbidden_variant" {
			forbidden++
		}
	}
	if forbidden != 1 {
		t.Errorf("expected 1 forbidden_variant, got %d", forbidden)
	}
}

func TestCheck_NoTargetLanguageInGlossary(t *testing.T) {
	g := glossaryWith(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("soul", tbx.StatusPreferred)),
		}),
	)

	src := []byte("The soul is eternal.")
	tgt := []byte("El alma es eterna.")

	result, err := Check(g, src, tgt, "en", "es", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations for concept without target lang, got %d", len(result.Violations))
	}
	if result.ConceptsChecked != 0 {
		t.Errorf("expected 0 concepts checked, got %d", result.ConceptsChecked)
	}
}

func TestCheck_AdmittedTargetWarning(t *testing.T) {
	g := glossaryWith(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("soul", tbx.StatusPreferred)),
			"es": langSec("es",
				term("alma", tbx.StatusPreferred),
				term("ánima", tbx.StatusAdmitted),
			),
		}),
	)

	src := []byte("The soul is eternal.")
	tgt := []byte("El ánima es eterna.")

	result, err := Check(g, src, tgt, "en", "es", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation (missing), got %d: %+v", len(result.Violations), result.Violations)
	}
	if result.Violations[0].Type != "missing" {
		t.Errorf("expected missing violation, got %s", result.Violations[0].Type)
	}

	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}
	w := result.Warnings[0]
	if w.Type != "admitted_variant" {
		t.Errorf("expected warning type admitted_variant, got %s", w.Type)
	}
	if w.ConceptID != "c001" {
		t.Errorf("expected concept_id c001, got %s", w.ConceptID)
	}
	if w.Variant != "ánima" {
		t.Errorf("expected variant ánima, got %s", w.Variant)
	}
	if w.Line == 0 {
		t.Error("expected non-zero line for admitted_variant warning")
	}
	if w.Column == 0 {
		t.Error("expected non-zero column for admitted_variant warning")
	}
}

func TestCheck_AdmittedTargetViolation_Strict(t *testing.T) {
	g := glossaryWith(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("soul", tbx.StatusPreferred)),
			"es": langSec("es",
				term("alma", tbx.StatusPreferred),
				term("ánima", tbx.StatusAdmitted),
			),
		}),
	)

	src := []byte("The soul is eternal.")
	tgt := []byte("El ánima es eterna.")

	result, err := Check(g, src, tgt, "en", "es", 80, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Warnings) != 0 {
		t.Errorf("expected 0 warnings under strict, got %d", len(result.Warnings))
	}

	var missing, admitted int
	for _, v := range result.Violations {
		switch v.Type {
		case "missing":
			missing++
		case "admitted_variant":
			admitted++
			if v.ConceptID != "c001" {
				t.Errorf("expected concept_id c001, got %s", v.ConceptID)
			}
			if v.Variant != "ánima" {
				t.Errorf("expected variant ánima, got %s", v.Variant)
			}
			if v.Line == 0 {
				t.Error("expected non-zero line")
			}
		}
	}

	if missing != 1 {
		t.Errorf("expected 1 missing violation, got %d", missing)
	}
	if admitted != 1 {
		t.Errorf("expected 1 admitted_variant violation, got %d", admitted)
	}
}

func TestCheck_AdmittedAlongsidePreferred(t *testing.T) {
	g := glossaryWith(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en", term("soul", tbx.StatusPreferred)),
			"es": langSec("es",
				term("alma", tbx.StatusPreferred),
				term("ánima", tbx.StatusAdmitted),
			),
		}),
	)

	src := []byte("The soul is eternal.")
	tgt := []byte("El alma y el ánima son eternos.")

	result, err := Check(g, src, tgt, "en", "es", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations (preferred present), got %d: %+v",
			len(result.Violations), result.Violations)
	}

	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 admitted_variant warning, got %d", len(result.Warnings))
	}
	if result.Warnings[0].Type != "admitted_variant" {
		t.Errorf("expected warning type admitted_variant, got %s", result.Warnings[0].Type)
	}
	if result.Warnings[0].Variant != "ánima" {
		t.Errorf("expected variant ánima, got %s", result.Warnings[0].Variant)
	}
}

func TestSortViolations_PositionalSorted(t *testing.T) {
	vs := []output.CheckViolation{
		{Type: "forbidden_variant", ConceptID: "c002", Line: 10, Column: 5},
		{Type: "forbidden_variant", ConceptID: "c001", Line: 3, Column: 12},
	}
	sortViolations(vs)
	if vs[0].Line != 3 || vs[0].Column != 12 {
		t.Errorf("expected (3,12) first, got (%d,%d)", vs[0].Line, vs[0].Column)
	}
	if vs[1].Line != 10 || vs[1].Column != 5 {
		t.Errorf("expected (10,5) second, got (%d,%d)", vs[1].Line, vs[1].Column)
	}
}

func TestSortViolations_MissingAtTail(t *testing.T) {
	vs := []output.CheckViolation{
		{Type: "missing", ConceptID: "c001"},
		{Type: "forbidden_variant", ConceptID: "c002", Line: 5, Column: 1},
	}
	sortViolations(vs)
	if vs[0].Type != "forbidden_variant" {
		t.Errorf("expected forbidden_variant first, got %s", vs[0].Type)
	}
	if vs[1].Type != "missing" {
		t.Errorf("expected missing last, got %s", vs[1].Type)
	}
}

func TestSortViolations_MissingTailByConceptID(t *testing.T) {
	vs := []output.CheckViolation{
		{Type: "missing", ConceptID: "tzimtzum"},
		{Type: "missing", ConceptID: "sefirah"},
	}
	sortViolations(vs)
	if vs[0].ConceptID != "sefirah" {
		t.Errorf("expected sefirah first, got %s", vs[0].ConceptID)
	}
	if vs[1].ConceptID != "tzimtzum" {
		t.Errorf("expected tzimtzum second, got %s", vs[1].ConceptID)
	}
}

func TestSortViolations_SamePositionTiebreak(t *testing.T) {
	vs := []output.CheckViolation{
		{Type: "forbidden_variant", ConceptID: "zebra", Line: 5, Column: 3},
		{Type: "forbidden_variant", ConceptID: "alpha", Line: 5, Column: 3},
	}
	sortViolations(vs)
	if vs[0].ConceptID != "alpha" {
		t.Errorf("expected alpha first at same position, got %s", vs[0].ConceptID)
	}
	if vs[1].ConceptID != "zebra" {
		t.Errorf("expected zebra second at same position, got %s", vs[1].ConceptID)
	}
}

func TestSortWarnings_PositionalSorted(t *testing.T) {
	ws := []output.CheckWarning{
		{Type: "admitted_variant", ConceptID: "c002", Line: 10, Column: 5},
		{Type: "admitted_variant", ConceptID: "c001", Line: 3, Column: 12},
	}
	sortWarnings(ws)
	if ws[0].Line != 3 || ws[0].Column != 12 {
		t.Errorf("expected (3,12) first, got (%d,%d)", ws[0].Line, ws[0].Column)
	}
	if ws[1].Line != 10 || ws[1].Column != 5 {
		t.Errorf("expected (10,5) second, got (%d,%d)", ws[1].Line, ws[1].Column)
	}
}

func TestSortWarnings_NonPositionalAtTail(t *testing.T) {
	ws := []output.CheckWarning{
		{Type: "admitted_variant", ConceptID: "c001"},
		{Type: "admitted_variant", ConceptID: "c002", Line: 5, Column: 1},
	}
	sortWarnings(ws)
	if ws[0].Line != 5 {
		t.Errorf("expected positional warning first, got line=%d", ws[0].Line)
	}
	if ws[1].Line != 0 {
		t.Errorf("expected non-positional warning last, got line=%d", ws[1].Line)
	}
}

func TestSortWarnings_SamePositionTiebreak(t *testing.T) {
	ws := []output.CheckWarning{
		{Type: "admitted_variant", ConceptID: "zebra", Line: 5, Column: 3},
		{Type: "admitted_variant", ConceptID: "alpha", Line: 5, Column: 3},
	}
	sortWarnings(ws)
	if ws[0].ConceptID != "alpha" {
		t.Errorf("expected alpha first at same position, got %s", ws[0].ConceptID)
	}
	if ws[1].ConceptID != "zebra" {
		t.Errorf("expected zebra second at same position, got %s", ws[1].ConceptID)
	}
}

func TestCheck_AdmittedSourceTermTriggersCheck(t *testing.T) {
	g := glossaryWith(
		concept("c001", map[string]tbx.LangSection{
			"en": langSec("en",
				term("soul", tbx.StatusPreferred),
				term("spirit", tbx.StatusAdmitted),
			),
			"es": langSec("es", term("alma", tbx.StatusPreferred)),
		}),
	)

	src := []byte("The spirit is eternal.")
	tgt := []byte("El alma es eterna.")

	result, err := Check(g, src, tgt, "en", "es", 80, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(result.Violations))
	}
	if result.ConceptsChecked != 1 {
		t.Errorf("expected 1 concept checked, got %d", result.ConceptsChecked)
	}
}
