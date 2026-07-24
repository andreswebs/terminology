package write

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
	_ "github.com/andreswebs/terminology/internal/tbx/linguist"
)

func TestParseJSONInput_ValidPayload(t *testing.T) {
	input := `{
		"concept_id": "test-concept",
		"subject_field": "linguistics",
		"languages": {
			"en": {
				"preferred": {
					"term": "test"
				}
			}
		}
	}`

	result, err := ParseJSONInput([]byte(input))
	if err != nil {
		t.Fatalf("ParseJSONInput: %v", err)
	}
	if result.ConceptID != "test-concept" {
		t.Errorf("got concept_id %q, want %q", result.ConceptID, "test-concept")
	}
	if result.SubjectField != "linguistics" {
		t.Errorf("got subject_field %q, want %q", result.SubjectField, "linguistics")
	}
	grp, ok := result.Languages["en"]
	if !ok || grp.Preferred == nil || grp.Preferred.Term != "test" {
		t.Errorf("expected en preferred term 'test', got %+v", result.Languages)
	}
}

func TestParseJSONInput_UnknownField_ReturnsInvalidInput(t *testing.T) {
	input := `{"concept_id": "x", "bogus_field": true}`

	_, err := ParseJSONInput([]byte(input))
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "invalid_input" {
		t.Errorf("got code %q, want %q", coded.Code(), "invalid_input")
	}
}

func TestParseTBXFragment_BareConceptEntry(t *testing.T) {
	input := `<conceptEntry id="c-test"
		xmlns="urn:iso:std:iso:30042:ed-2"
		xmlns:min="http://www.tbxinfo.net/ns/min"
		xmlns:basic="http://www.tbxinfo.net/ns/basic"
		xmlns:ling="http://www.tbxinfo.net/ns/linguist">
		<langSec xml:lang="en">
			<termSec>
				<term>hello</term>
			</termSec>
		</langSec>
	</conceptEntry>`

	concepts, err := ParseTBXFragment([]byte(input))
	if err != nil {
		t.Fatalf("ParseTBXFragment: %v", err)
	}
	if len(concepts) != 1 {
		t.Fatalf("got %d concepts, want 1", len(concepts))
	}
	if concepts[0].ID != "c-test" {
		t.Errorf("got ID %q, want %q", concepts[0].ID, "c-test")
	}
	if _, ok := concepts[0].Languages["en"]; !ok {
		t.Error("expected en language section")
	}
}

func TestParseTBXFragment_ConceptEntryList(t *testing.T) {
	input := `<conceptEntryList
		xmlns="urn:iso:std:iso:30042:ed-2"
		xmlns:min="http://www.tbxinfo.net/ns/min"
		xmlns:basic="http://www.tbxinfo.net/ns/basic"
		xmlns:ling="http://www.tbxinfo.net/ns/linguist">
		<conceptEntry id="c1">
			<langSec xml:lang="en">
				<termSec><term>one</term></termSec>
			</langSec>
		</conceptEntry>
		<conceptEntry id="c2">
			<langSec xml:lang="en">
				<termSec><term>two</term></termSec>
			</langSec>
		</conceptEntry>
	</conceptEntryList>`

	concepts, err := ParseTBXFragment([]byte(input))
	if err != nil {
		t.Fatalf("ParseTBXFragment: %v", err)
	}
	if len(concepts) != 2 {
		t.Fatalf("got %d concepts, want 2", len(concepts))
	}
	if concepts[0].ID != "c1" || concepts[1].ID != "c2" {
		t.Errorf("got IDs %q/%q, want c1/c2", concepts[0].ID, concepts[1].ID)
	}
}

func TestParseTBXFragment_FullTBXDocument_Rejected(t *testing.T) {
	input := `<?xml version="1.0"?>
	<tbx type="TBX-Linguist" style="dct"
		xmlns="urn:iso:std:iso:30042:ed-2">
		<tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
		<text><body>
			<conceptEntry id="c1">
				<langSec xml:lang="en"><termSec><term>x</term></termSec></langSec>
			</conceptEntry>
		</body></text>
	</tbx>`

	_, err := ParseTBXFragment([]byte(input))
	if err == nil {
		t.Fatal("expected error for full tbx document, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "invalid_input" {
		t.Errorf("got code %q, want %q", coded.Code(), "invalid_input")
	}
}

func TestParseTBXFragment_MalformedXML_ReturnsInvalidInput(t *testing.T) {
	_, err := ParseTBXFragment([]byte(`<conceptEntry id="x"><not-closed`))
	if err == nil {
		t.Fatal("expected error for malformed XML, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "invalid_input" {
		t.Errorf("got code %q, want %q", coded.Code(), "invalid_input")
	}
}

func TestParseJSONInput_MalformedJSON_ReturnsInvalidInput(t *testing.T) {
	_, err := ParseJSONInput([]byte(`{not json`))
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "invalid_input" {
		t.Errorf("got code %q, want %q", coded.Code(), "invalid_input")
	}
}

func minimalGlossaryFile(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "test.tbx")
	content := `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader>
    <fileDesc>
      <sourceDesc><p>test</p></sourceDesc>
    </fileDesc>
  </tbxHeader>
  <text>
    <body>
      <conceptEntry id="c001">
        <langSec xml:lang="en">
          <termSec>
            <term>widget</term>
          </termSec>
        </langSec>
      </conceptEntry>
    </body>
  </text>
</tbx>`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExecute_ValidationFailure_AbortsSave(t *testing.T) {
	dir := t.TempDir()
	path := minimalGlossaryFile(t, dir)

	mutator := func(g *tbx.Glossary) (*tbx.Concept, error) {
		c := tbx.Concept{
			ID: "c001",
			Languages: map[string]tbx.LangSection{
				"en": {Lang: "en", Terms: []tbx.Term{{Surface: "dup"}}},
			},
		}
		g.Concepts = append(g.Concepts, c)
		return &g.Concepts[len(g.Concepts)-1], nil
	}

	_, err := Execute(path, mutator, false)
	if err == nil {
		t.Fatal("expected validation error for duplicate_id, got nil")
	}

	g, _, loadErr := tbx.Load(path)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	if len(g.Concepts) != 1 {
		t.Errorf("validation failure should not save; got %d concepts, want 1", len(g.Concepts))
	}
}

func TestExecute_MutatorError_Propagated(t *testing.T) {
	dir := t.TempDir()
	path := minimalGlossaryFile(t, dir)

	mutator := func(_ *tbx.Glossary) (*tbx.Concept, error) {
		return nil, ErrNotFound
	}

	_, err := Execute(path, mutator, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "not_found" {
		t.Errorf("got code %q, want %q", coded.Code(), "not_found")
	}
}

func TestExecute_DanglingCrossref_AbortsSave(t *testing.T) {
	dir := t.TempDir()
	path := minimalGlossaryFile(t, dir)

	mutator := func(g *tbx.Glossary) (*tbx.Concept, error) {
		c := tbx.Concept{
			ID: "c-with-ref",
			CrossRefs: []tbx.CrossRef{
				{Target: "nonexistent-concept"},
			},
			Languages: map[string]tbx.LangSection{
				"en": {Lang: "en", Terms: []tbx.Term{{Surface: "ref"}}},
			},
		}
		g.Concepts = append(g.Concepts, c)
		return &g.Concepts[len(g.Concepts)-1], nil
	}

	_, err := Execute(path, mutator, false)
	if err == nil {
		t.Fatal("expected validation error for dangling crossref, got nil")
	}

	g, _, loadErr := tbx.Load(path)
	if loadErr != nil {
		t.Fatalf("re-load: %v", loadErr)
	}
	if len(g.Concepts) != 1 {
		t.Errorf("dangling crossref should not save; got %d concepts, want 1", len(g.Concepts))
	}
}

func TestExecute_DryRun_StillValidates(t *testing.T) {
	dir := t.TempDir()
	path := minimalGlossaryFile(t, dir)

	mutator := func(g *tbx.Glossary) (*tbx.Concept, error) {
		c := tbx.Concept{
			ID:        "c001",
			Languages: map[string]tbx.LangSection{},
		}
		g.Concepts = append(g.Concepts, c)
		return &g.Concepts[len(g.Concepts)-1], nil
	}

	_, err := Execute(path, mutator, true)
	if err == nil {
		t.Fatal("expected validation error for duplicate_id in dry-run, got nil")
	}
}

func TestExecute_LoadError_Propagated(t *testing.T) {
	mutator := func(_ *tbx.Glossary) (*tbx.Concept, error) {
		return nil, nil
	}

	_, err := Execute("/nonexistent/path.tbx", mutator, false)
	if err == nil {
		t.Fatal("expected load error, got nil")
	}
}

func TestExecute_DryRun_DoesNotSave(t *testing.T) {
	dir := t.TempDir()
	path := minimalGlossaryFile(t, dir)

	mutator := func(g *tbx.Glossary) (*tbx.Concept, error) {
		c := tbx.Concept{
			ID: "dry-run-concept",
			Languages: map[string]tbx.LangSection{
				"en": {Lang: "en", Terms: []tbx.Term{{Surface: "dry"}}},
			},
		}
		g.Concepts = append(g.Concepts, c)
		return &g.Concepts[len(g.Concepts)-1], nil
	}

	result, err := Execute(path, mutator, true)
	if err != nil {
		t.Fatalf("Execute dry-run: %v", err)
	}
	if result.ID != "dry-run-concept" {
		t.Errorf("got %q, want %q", result.ID, "dry-run-concept")
	}

	g, _, err := tbx.Load(path)
	if err != nil {
		t.Fatalf("re-load: %v", err)
	}
	if len(g.Concepts) != 1 {
		t.Errorf("dry-run should not save; got %d concepts, want 1", len(g.Concepts))
	}
}

func TestExecute_HappyPath_SavesAndReturnsAffectedConcept(t *testing.T) {
	dir := t.TempDir()
	path := minimalGlossaryFile(t, dir)

	mutator := func(g *tbx.Glossary) (*tbx.Concept, error) {
		c := &tbx.Concept{
			ID:        "new-concept",
			Languages: map[string]tbx.LangSection{},
		}
		g.Concepts = append(g.Concepts, *c)
		return &g.Concepts[len(g.Concepts)-1], nil
	}

	result, err := Execute(path, mutator, false)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.ID != "new-concept" {
		t.Errorf("got concept ID %q, want %q", result.ID, "new-concept")
	}

	g, _, err := tbx.Load(path)
	if err != nil {
		t.Fatalf("re-load: %v", err)
	}
	if len(g.Concepts) != 2 {
		t.Errorf("got %d concepts, want 2", len(g.Concepts))
	}
}

func TestParseTBXFragment_DCA_PersistsDefinitionAndStatus(t *testing.T) {
	input := `<conceptEntry id="irimi-nage">
  <descrip type="subjectField">aikido</descrip>
  <langSec xml:lang="en">
    <descrip type="definition">Nage enters past uke to throw.</descrip>
    <termSec>
      <term>entering throw</term>
      <termNote type="administrativeStatus">preferredTerm-admn-sts</termNote>
    </termSec>
  </langSec>
</conceptEntry>`

	concepts, err := ParseTBXFragment([]byte(input))
	if err != nil {
		t.Fatalf("ParseTBXFragment: %v", err)
	}
	if len(concepts) != 1 {
		t.Fatalf("got %d concepts, want 1", len(concepts))
	}

	c := concepts[0]
	if c.SubjectField != "aikido" {
		t.Errorf("SubjectField = %q, want %q", c.SubjectField, "aikido")
	}

	ls, ok := c.Languages["en"]
	if !ok {
		t.Fatalf("expected en language section")
	}
	if len(ls.Definitions) != 1 || ls.Definitions[0].Plain != "Nage enters past uke to throw." {
		t.Errorf("langSec definitions = %+v, want one entry %q", ls.Definitions, "Nage enters past uke to throw.")
	}
	if len(ls.Terms) != 1 {
		t.Fatalf("got %d terms, want 1", len(ls.Terms))
	}
	if ls.Terms[0].Surface != "entering throw" {
		t.Errorf("term surface = %q, want %q", ls.Terms[0].Surface, "entering throw")
	}
	if ls.Terms[0].AdministrativeStatus != tbx.StatusPreferred {
		t.Errorf("term status = %v, want StatusPreferred", ls.Terms[0].AdministrativeStatus)
	}
}

func TestParseTBXFragment_DCT_PersistsDefinitionAndStatus(t *testing.T) {
	input := `<conceptEntry id="irimi-nage">
  <min:subjectField>aikido</min:subjectField>
  <langSec xml:lang="en">
    <basic:definition>Nage enters past uke to throw.</basic:definition>
    <termSec>
      <term>entering throw</term>
      <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
    </termSec>
  </langSec>
</conceptEntry>`

	concepts, err := ParseTBXFragment([]byte(input))
	if err != nil {
		t.Fatalf("ParseTBXFragment: %v", err)
	}
	if len(concepts) != 1 {
		t.Fatalf("got %d concepts, want 1", len(concepts))
	}

	c := concepts[0]
	if c.SubjectField != "aikido" {
		t.Errorf("SubjectField = %q, want %q", c.SubjectField, "aikido")
	}
	ls := c.Languages["en"]
	if len(ls.Definitions) != 1 || ls.Definitions[0].Plain != "Nage enters past uke to throw." {
		t.Errorf("langSec definitions = %+v, want one entry", ls.Definitions)
	}
	if len(ls.Terms) != 1 || ls.Terms[0].AdministrativeStatus != tbx.StatusPreferred {
		t.Errorf("term status not preserved: %+v", ls.Terms)
	}
}

func TestParseTBXFragment_UnknownElement_FailsClosed(t *testing.T) {
	input := `<conceptEntry id="x">
  <langSec xml:lang="en">
    <descrip type="madeUpThing">nonsense</descrip>
    <termSec><term>x</term></termSec>
  </langSec>
</conceptEntry>`

	_, err := ParseTBXFragment([]byte(input))
	if err == nil {
		t.Fatal("expected error for unknown element, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "invalid_input" {
		t.Errorf("got code %q, want %q", coded.Code(), "invalid_input")
	}
	if !strings.Contains(err.Error(), "madeUpThing") {
		t.Errorf("error message %q does not name the offending element", err.Error())
	}
}
