package tbx_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
	_ "github.com/andreswebs/terminology/internal/tbx/linguist"
)

func minimalGlossary() *tbx.Glossary {
	return &tbx.Glossary{
		Dialect: tbx.DialectLinguist,
		Style:   tbx.StyleDCT,
		Concepts: []tbx.Concept{
			{
				ID: "tzimtzum",
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "tzimtzum", AdministrativeStatus: tbx.StatusPreferred}}},
					"he": {Lang: "he", Terms: []tbx.Term{{Surface: "צמצום", AdministrativeStatus: tbx.StatusPreferred}}},
				},
			},
		},
	}
}

func TestValidate_CleanGlossary(t *testing.T) {
	g := minimalGlossary()
	res := g.Validate(false)

	if res.Concepts != 1 {
		t.Errorf("Concepts = %d, want 1", res.Concepts)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("Warnings = %v, want empty", res.Warnings)
	}
	if len(res.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", res.Errors)
	}
	if len(res.Languages) != 2 {
		t.Fatalf("Languages = %v, want [en he]", res.Languages)
	}
	if res.Languages[0] != "en" || res.Languages[1] != "he" {
		t.Errorf("Languages = %v, want [en he]", res.Languages)
	}
}

func TestValidate_DuplicateID(t *testing.T) {
	g := minimalGlossary()
	g.Concepts = append(g.Concepts, tbx.Concept{
		ID:        "tzimtzum",
		Languages: map[string]tbx.LangSection{"en": {Lang: "en", Terms: []tbx.Term{{Surface: "dup"}}}},
	})

	res := g.Validate(false)
	if res.Concepts != 2 {
		t.Errorf("Concepts = %d, want 2", res.Concepts)
	}

	found := false
	for _, w := range res.Warnings {
		if w.Code == "duplicate_id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected duplicate_id warning")
	}
}

func TestValidate_DuplicateID_CarriesPosition(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID:        "c1",
				StartLine: 5, StartCol: 7,
				Languages: map[string]tbx.LangSection{"en": {Lang: "en", Terms: []tbx.Term{{Surface: "x"}}}},
			},
			{
				ID:        "c1",
				StartLine: 12, StartCol: 7,
				Languages: map[string]tbx.LangSection{"en": {Lang: "en", Terms: []tbx.Term{{Surface: "y"}}}},
			},
		},
	}

	res := g.Validate(false)
	var w *tbx.Warning
	for i := range res.Warnings {
		if res.Warnings[i].Code == "duplicate_id" {
			w = &res.Warnings[i]
			break
		}
	}
	if w == nil {
		t.Fatal("expected duplicate_id warning")
	}
	if w.Line != 12 {
		t.Errorf("Line = %d, want 12", w.Line)
	}
	if w.Col != 7 {
		t.Errorf("Col = %d, want 7", w.Col)
	}
}

func TestValidate_UnresolvedCrossRef_CarriesPosition(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID:        "c1",
				StartLine: 8, StartCol: 5,
				CrossRefs: []tbx.CrossRef{{Target: "missing"}},
				Languages: map[string]tbx.LangSection{"en": {Lang: "en", Terms: []tbx.Term{{Surface: "x"}}}},
			},
		},
	}

	res := g.Validate(false)
	var w *tbx.Warning
	for i := range res.Warnings {
		if res.Warnings[i].Code == "unresolved_crossref" {
			w = &res.Warnings[i]
			break
		}
	}
	if w == nil {
		t.Fatal("expected unresolved_crossref warning")
	}
	if w.Line != 8 {
		t.Errorf("Line = %d, want 8", w.Line)
	}
	if w.Col != 5 {
		t.Errorf("Col = %d, want 5", w.Col)
	}
}

func TestValidate_InvalidLangTag_CarriesLangSecPosition(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID:        "c1",
				StartLine: 3, StartCol: 5,
				Languages: map[string]tbx.LangSection{
					"not valid!!!": {
						Lang:      "not valid!!!",
						StartLine: 10, StartCol: 9,
						Terms: []tbx.Term{{Surface: "x"}},
					},
				},
			},
		},
	}

	res := g.Validate(false)
	var w *tbx.Warning
	for i := range res.Warnings {
		if res.Warnings[i].Code == "invalid_lang_tag" {
			w = &res.Warnings[i]
			break
		}
	}
	if w == nil {
		t.Fatal("expected invalid_lang_tag warning")
	}
	if w.Line != 10 {
		t.Errorf("Line = %d, want 10 (langSec position)", w.Line)
	}
	if w.Col != 9 {
		t.Errorf("Col = %d, want 9 (langSec position)", w.Col)
	}
}

func TestValidate_MissingTerm_CarriesLangSecPosition(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID:        "c1",
				StartLine: 3, StartCol: 5,
				Languages: map[string]tbx.LangSection{
					"en": {
						Lang:      "en",
						StartLine: 7, StartCol: 9,
						Terms: []tbx.Term{},
					},
				},
			},
		},
	}

	res := g.Validate(false)
	var w *tbx.Warning
	for i := range res.Warnings {
		if res.Warnings[i].Code == "missing_term" {
			w = &res.Warnings[i]
			break
		}
	}
	if w == nil {
		t.Fatal("expected missing_term warning")
	}
	if w.Line != 7 {
		t.Errorf("Line = %d, want 7 (langSec position)", w.Line)
	}
	if w.Col != 9 {
		t.Errorf("Col = %d, want 9 (langSec position)", w.Col)
	}
}

func TestValidate_InvalidLangTag(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID: "c1",
				Languages: map[string]tbx.LangSection{
					"not a lang!!!": {Lang: "not a lang!!!", Terms: []tbx.Term{{Surface: "x"}}},
				},
			},
		},
	}

	res := g.Validate(false)
	found := false
	for _, w := range res.Warnings {
		if w.Code == "invalid_lang_tag" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected invalid_lang_tag warning")
	}
}

func TestValidate_MissingTerm(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID: "c1",
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{}},
				},
			},
		},
	}

	res := g.Validate(false)
	found := false
	for _, w := range res.Warnings {
		if w.Code == "missing_term" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected missing_term warning")
	}
}

func TestValidate_UnresolvedCrossRef_Lenient(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID:        "c1",
				CrossRefs: []tbx.CrossRef{{Target: "nonexistent"}},
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "x"}}},
				},
			},
		},
	}

	res := g.Validate(false)
	found := false
	for _, w := range res.Warnings {
		if w.Code == "unresolved_crossref" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected unresolved_crossref warning (lenient)")
	}
	if len(res.Errors) != 0 {
		t.Error("expected no errors in lenient mode")
	}
}

func TestValidate_UnresolvedCrossRef_Strict(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID:        "c1",
				CrossRefs: []tbx.CrossRef{{Target: "nonexistent"}},
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "x"}}},
				},
			},
		},
	}

	res := g.Validate(true)
	found := false
	for _, w := range res.Errors {
		if w.Code == "unresolved_crossref" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected unresolved_crossref error (strict)")
	}
}

func testFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "internal", "app", "testdata", "fixtures", name)
}

func TestValidate_LoadedFile_WarningsCarryLineCol(t *testing.T) {
	g, _, err := tbx.Load(testFixturePath("with-warnings.tbx"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	res := g.Validate(false)
	for _, w := range res.Warnings {
		if w.Line == 0 {
			t.Errorf("warning %q (concept %q): Line = 0, want > 0", w.Code, w.ConceptID)
		}
		if w.Col == 0 {
			t.Errorf("warning %q (concept %q): Col = 0, want > 0", w.Code, w.ConceptID)
		}
	}
}

func TestValidate_CrossRefForwardReference(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID:        "alpha",
				CrossRefs: []tbx.CrossRef{{Target: "beta"}},
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "alpha"}}},
				},
			},
			{
				ID: "beta",
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "beta"}}},
				},
			},
		},
	}

	res := g.Validate(false)
	if len(res.Warnings) != 0 {
		t.Errorf("Warnings = %v, want empty (forward crossref should resolve)", res.Warnings)
	}
	if len(res.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", res.Errors)
	}
}

func TestValidate_TermCrossRefForwardReference(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID: "alpha",
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{
						{Surface: "alpha", CrossRefs: []tbx.CrossRef{{Target: "beta"}}},
					}},
				},
			},
			{
				ID: "beta",
				Languages: map[string]tbx.LangSection{
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "beta"}}},
				},
			},
		},
	}

	res := g.Validate(false)
	if len(res.Warnings) != 0 {
		t.Errorf("Warnings = %v, want empty (forward term crossref should resolve)", res.Warnings)
	}
	if len(res.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", res.Errors)
	}
}

func TestValidate_LanguagesSortedASCII(t *testing.T) {
	g := &tbx.Glossary{
		Concepts: []tbx.Concept{
			{
				ID: "c1",
				Languages: map[string]tbx.LangSection{
					"he": {Lang: "he", Terms: []tbx.Term{{Surface: "x"}}},
					"en": {Lang: "en", Terms: []tbx.Term{{Surface: "y"}}},
					"es": {Lang: "es", Terms: []tbx.Term{{Surface: "z"}}},
				},
			},
		},
	}

	res := g.Validate(false)
	want := []string{"en", "es", "he"}
	if len(res.Languages) != len(want) {
		t.Fatalf("Languages = %v, want %v", res.Languages, want)
	}
	for i, l := range res.Languages {
		if l != want[i] {
			t.Errorf("Languages[%d] = %q, want %q", i, l, want[i])
		}
	}
}
