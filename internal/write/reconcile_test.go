package write

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/andreswebs/terminology/internal/clock"
	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	_ "github.com/andreswebs/terminology/internal/tbx/linguist"
)

func makeConcept(id, lang, term string) tbx.Concept {
	return tbx.Concept{
		ID: id,
		Languages: map[string]tbx.LangSection{
			lang: {Lang: lang, Terms: []tbx.Term{{Surface: term}}},
		},
	}
}

func makeGlossary(concepts ...tbx.Concept) *tbx.Glossary {
	return &tbx.Glossary{
		Dialect:    tbx.DialectLinguist,
		SourceDesc: "test",
		Concepts:   concepts,
	}
}

func TestReconcile_AddNewConcepts(t *testing.T) {
	g := makeGlossary(makeConcept("existing", "en", "existing"))
	payload := []tbx.Concept{
		makeConcept("existing", "en", "existing"),
		makeConcept("new-one", "en", "new"),
	}

	result, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Added) != 1 || result.Added[0] != "new-one" {
		t.Errorf("Added = %v, want [new-one]", result.Added)
	}
	if len(result.Unchanged) != 1 || result.Unchanged[0] != "existing" {
		t.Errorf("Unchanged = %v, want [existing]", result.Unchanged)
	}
	if len(result.Updated) != 0 {
		t.Errorf("Updated = %v, want []", result.Updated)
	}
	if len(result.Removed) != 0 {
		t.Errorf("Removed = %v, want []", result.Removed)
	}

	found := false
	for _, c := range g.Concepts {
		if c.ID == "new-one" {
			found = true
		}
	}
	if !found {
		t.Error("new concept not added to glossary")
	}
}

func TestReconcile_UpdateChangedConcepts(t *testing.T) {
	g := makeGlossary(makeConcept("c1", "en", "old-term"))
	payload := []tbx.Concept{
		makeConcept("c1", "en", "new-term"),
	}

	result, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Updated) != 1 || result.Updated[0] != "c1" {
		t.Errorf("Updated = %v, want [c1]", result.Updated)
	}

	if g.Concepts[0].Languages["en"].Terms[0].Surface != "new-term" {
		t.Error("concept not replaced in glossary")
	}
}

func TestReconcile_UnchangedConcepts(t *testing.T) {
	g := makeGlossary(makeConcept("c1", "en", "same"))
	payload := []tbx.Concept{
		makeConcept("c1", "en", "same"),
	}

	result, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Unchanged) != 1 || result.Unchanged[0] != "c1" {
		t.Errorf("Unchanged = %v, want [c1]", result.Unchanged)
	}
	if len(result.Updated) != 0 {
		t.Errorf("Updated = %v, want []", result.Updated)
	}
}

func TestReconcile_Prune_RemovesAbsentConcepts(t *testing.T) {
	g := makeGlossary(
		makeConcept("keep", "en", "keep"),
		makeConcept("drop", "en", "drop"),
	)
	payload := []tbx.Concept{
		makeConcept("keep", "en", "keep"),
	}

	result, err := Reconcile(g, payload, true)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Removed) != 1 || result.Removed[0] != "drop" {
		t.Errorf("Removed = %v, want [drop]", result.Removed)
	}

	if len(g.Concepts) != 1 || g.Concepts[0].ID != "keep" {
		t.Errorf("glossary should have 1 concept 'keep', got %d", len(g.Concepts))
	}
}

func TestReconcile_NoPrune_PreservesAbsentConcepts(t *testing.T) {
	g := makeGlossary(
		makeConcept("keep", "en", "keep"),
		makeConcept("absent", "en", "absent"),
	)
	payload := []tbx.Concept{
		makeConcept("keep", "en", "keep"),
	}

	result, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Removed) != 0 {
		t.Errorf("Removed = %v, want []", result.Removed)
	}
	if len(g.Concepts) != 2 {
		t.Error("absent concept should be preserved without --prune")
	}
}

func TestReconcile_Prune_DanglingCrossref_Errors(t *testing.T) {
	c1 := makeConcept("referencer", "en", "referencer")
	c1.CrossRefs = []tbx.CrossRef{{Target: "target"}}
	c2 := makeConcept("target", "en", "target")

	g := makeGlossary(c1, c2)
	payload := []tbx.Concept{c1}

	_, err := Reconcile(g, payload, true)
	if err == nil {
		t.Fatal("expected dangling_crossref error, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "dangling_crossref" {
		t.Errorf("got code %q, want %q", coded.Code(), "dangling_crossref")
	}
}

func TestReconcile_Prune_DanglingCrossref_TermLevel(t *testing.T) {
	c1 := makeConcept("referencer", "en", "referencer")
	c1.Languages["en"] = tbx.LangSection{
		Lang: "en",
		Terms: []tbx.Term{{
			Surface:   "referencer",
			CrossRefs: []tbx.CrossRef{{Target: "target"}},
		}},
	}
	c2 := makeConcept("target", "en", "target")

	g := makeGlossary(c1, c2)
	payload := []tbx.Concept{c1}

	_, err := Reconcile(g, payload, true)
	if err == nil {
		t.Fatal("expected dangling_crossref error for term-level ref, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "dangling_crossref" {
		t.Errorf("got code %q, want %q", coded.Code(), "dangling_crossref")
	}
}

func TestReconcile_WholesaleReplace_DropsOmittedFields(t *testing.T) {
	c := makeConcept("c1", "en", "term")
	c.SubjectField = "original-field"
	c.Notes = []string{"original note"}

	g := makeGlossary(c)

	payload := []tbx.Concept{
		makeConcept("c1", "en", "term-updated"),
	}

	result, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Updated) != 1 {
		t.Fatalf("Updated = %v, want [c1]", result.Updated)
	}

	updated := g.Concepts[0]
	if updated.SubjectField != "" {
		t.Errorf("SubjectField should be empty after wholesale replace, got %q", updated.SubjectField)
	}
	if len(updated.Notes) != 0 {
		t.Errorf("Notes should be empty after wholesale replace, got %v", updated.Notes)
	}
}

func TestReconcile_SortedResults(t *testing.T) {
	g := makeGlossary(
		makeConcept("z-concept", "en", "z"),
		makeConcept("a-concept", "en", "a-old"),
	)
	payload := []tbx.Concept{
		makeConcept("z-concept", "en", "z"),
		makeConcept("a-concept", "en", "a-new"),
		makeConcept("m-concept", "en", "m"),
	}

	result, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if !sort.StringsAreSorted(result.Added) {
		t.Errorf("Added not sorted: %v", result.Added)
	}
	if !sort.StringsAreSorted(result.Updated) {
		t.Errorf("Updated not sorted: %v", result.Updated)
	}
	if !sort.StringsAreSorted(result.Unchanged) {
		t.Errorf("Unchanged not sorted: %v", result.Unchanged)
	}
}

func TestReconcile_Idempotency(t *testing.T) {
	original := makeConcept("c1", "en", "term")
	g := makeGlossary(original)
	payload := []tbx.Concept{makeConcept("c1", "en", "term")}

	r1, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("first Reconcile: %v", err)
	}

	if len(r1.Unchanged) != 1 {
		t.Fatalf("first run: Unchanged = %v, want [c1]", r1.Unchanged)
	}

	r2, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("second Reconcile: %v", err)
	}

	if len(r2.Unchanged) != 1 || r2.Unchanged[0] != "c1" {
		t.Errorf("second run: Unchanged = %v, want [c1]", r2.Unchanged)
	}
	if len(r2.Added) != 0 || len(r2.Updated) != 0 {
		t.Errorf("second run should have zero ops, got Added=%v Updated=%v", r2.Added, r2.Updated)
	}
}

func TestReconcile_Idempotency_WithTransactions(t *testing.T) {
	g := makeGlossary(makeConcept("c1", "en", "term"))
	payload := []tbx.Concept{makeConcept("c1", "en", "new-term")}

	ctx := clock.With(context.Background(), fakeClock{T: time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC)})

	r1, err := ReconcileWithTxn(ctx, g, payload, false, "test-author")
	if err != nil {
		t.Fatalf("first Reconcile: %v", err)
	}
	if len(r1.Updated) != 1 {
		t.Fatalf("first run: Updated = %v, want [c1]", r1.Updated)
	}

	payloadAgain := []tbx.Concept{makeConcept("c1", "en", "new-term")}
	r2, err := ReconcileWithTxn(ctx, g, payloadAgain, false, "test-author")
	if err != nil {
		t.Fatalf("second Reconcile: %v", err)
	}
	if len(r2.Unchanged) != 1 || r2.Unchanged[0] != "c1" {
		t.Errorf("second run: Unchanged = %v, want [c1]", r2.Unchanged)
	}
	if len(r2.Updated) != 0 {
		t.Errorf("second run should not update, got Updated=%v", r2.Updated)
	}
}

func TestReconcile_ValidationFailure_ReturnsApplyValidationError(t *testing.T) {
	g := makeGlossary(makeConcept("c1", "en", "term"))

	bad := makeConcept("c-bad", "en", "bad")
	bad.CrossRefs = []tbx.CrossRef{{Target: "nonexistent"}}

	payload := []tbx.Concept{
		makeConcept("c1", "en", "term"),
		bad,
	}

	_, err := Reconcile(g, payload, false)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var ave *ApplyValidationError
	if !errors.As(err, &ave) {
		t.Fatalf("expected *ApplyValidationError, got %T: %v", err, err)
	}

	if len(ave.Failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(ave.Failures))
	}

	f := ave.Failures[0]
	if f.ConceptID != "c-bad" {
		t.Errorf("failure concept_id = %q, want %q", f.ConceptID, "c-bad")
	}
	if f.Code != "unresolved_crossref" {
		t.Errorf("failure code = %q, want %q", f.Code, "unresolved_crossref")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected Coded interface, got %T", err)
	}
	if coded.Code() != "apply_validation_failed" {
		t.Errorf("error code = %q, want %q", coded.Code(), "apply_validation_failed")
	}

	exitCoded, ok := err.(interface{ ExitCode() int })
	if !ok {
		t.Fatalf("expected ExitCode interface, got %T", err)
	}
	if exitCoded.ExitCode() != 1 {
		t.Errorf("exit code = %d, want 1", exitCoded.ExitCode())
	}
}

func TestReconcile_ValidationFailure_MultipleFailures(t *testing.T) {
	bad1 := makeConcept("bad-a", "en", "bad-a")
	bad1.CrossRefs = []tbx.CrossRef{{Target: "ghost-1"}}

	bad2 := makeConcept("bad-b", "en", "bad-b")
	bad2.CrossRefs = []tbx.CrossRef{{Target: "ghost-2"}}

	g := makeGlossary()
	payload := []tbx.Concept{bad1, bad2}

	_, err := Reconcile(g, payload, false)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var ave *ApplyValidationError
	if !errors.As(err, &ave) {
		t.Fatalf("expected *ApplyValidationError, got %T: %v", err, err)
	}

	if len(ave.Failures) != 2 {
		t.Fatalf("expected 2 failures, got %d: %+v", len(ave.Failures), ave.Failures)
	}
}

func TestReconcile_ValidationFailure_DetailsInterface(t *testing.T) {
	g := makeGlossary(makeConcept("c1", "en", "term"))

	bad := makeConcept("c-bad", "en", "bad")
	bad.CrossRefs = []tbx.CrossRef{{Target: "nonexistent"}}

	payload := []tbx.Concept{
		makeConcept("c1", "en", "term"),
		bad,
	}

	_, err := Reconcile(g, payload, false)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	detailed, ok := err.(interface{ ErrorDetails() any })
	if !ok {
		t.Fatalf("expected Detailed interface, got %T", err)
	}

	details := detailed.ErrorDetails()
	dm, ok := details.(map[string][]output.ApplyFailure)
	if !ok {
		t.Fatalf("details type = %T, want map[string][]output.ApplyFailure", details)
	}

	failures := dm["failures"]
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(failures))
	}
}

func TestReconcile_UpdatePreservesID(t *testing.T) {
	g := makeGlossary(makeConcept("original-id", "en", "old"))
	payload := []tbx.Concept{
		makeConcept("original-id", "en", "new"),
	}

	_, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if g.Concepts[0].ID != "original-id" {
		t.Errorf("ID changed to %q, should be preserved", g.Concepts[0].ID)
	}
}

func TestReconcile_EmptyGlossary_AllAdded(t *testing.T) {
	g := makeGlossary()
	payload := []tbx.Concept{
		makeConcept("a", "en", "alpha"),
		makeConcept("b", "en", "beta"),
	}

	result, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Added) != 2 {
		t.Errorf("Added = %v, want [a, b]", result.Added)
	}
}

func TestReconcile_EmptyPayload_NoPrune(t *testing.T) {
	g := makeGlossary(makeConcept("c1", "en", "term"))
	payload := []tbx.Concept{}

	result, err := Reconcile(g, payload, false)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Added) != 0 && len(result.Updated) != 0 && len(result.Removed) != 0 {
		t.Errorf("empty payload without prune should be no-ops")
	}
	if len(g.Concepts) != 1 {
		t.Error("glossary should be unchanged")
	}
}

func TestReconcile_EmptyPayload_WithPrune(t *testing.T) {
	g := makeGlossary(makeConcept("c1", "en", "term"))
	payload := []tbx.Concept{}

	result, err := Reconcile(g, payload, true)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Removed) != 1 || result.Removed[0] != "c1" {
		t.Errorf("Removed = %v, want [c1]", result.Removed)
	}
	if len(g.Concepts) != 0 {
		t.Error("glossary should be empty after prune")
	}
}
