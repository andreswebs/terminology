package tbx

import "testing"

func TestStatusEnum(t *testing.T) {
	t.Parallel()

	if StatusUnspecified != 0 {
		t.Errorf("StatusUnspecified = %d, want 0", StatusUnspecified)
	}

	values := []struct {
		name   string
		status Status
		want   int
	}{
		{"StatusPreferred", StatusPreferred, 1},
		{"StatusAdmitted", StatusAdmitted, 2},
		{"StatusDeprecated", StatusDeprecated, 3},
		{"StatusSuperseded", StatusSuperseded, 4},
	}

	for _, v := range values {
		if int(v.status) != v.want {
			t.Errorf("%s = %d, want %d", v.name, v.status, v.want)
		}
	}

	seen := map[Status]bool{}
	for _, v := range values {
		if seen[v.status] {
			t.Errorf("duplicate Status value: %d", v.status)
		}
		seen[v.status] = true
	}
}

func TestParseStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  Status
	}{
		{"preferredTerm-admn-sts", StatusPreferred},
		{"preferredTerm", StatusPreferred},
		{"admittedTerm-admn-sts", StatusAdmitted},
		{"admittedTerm", StatusAdmitted},
		{"deprecatedTerm-admn-sts", StatusDeprecated},
		{"deprecatedTerm", StatusDeprecated},
		{"supersededTerm-admn-sts", StatusSuperseded},
		{"supersededTerm", StatusSuperseded},
		{"", StatusUnspecified},
		{"bogus", StatusUnspecified},
	}

	for _, tt := range tests {
		got := ParseStatus(tt.input)
		if got != tt.want {
			t.Errorf("ParseStatus(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status Status
		want   string
	}{
		{StatusPreferred, "preferredTerm-admn-sts"},
		{StatusAdmitted, "admittedTerm-admn-sts"},
		{StatusDeprecated, "deprecatedTerm-admn-sts"},
		{StatusSuperseded, "supersededTerm-admn-sts"},
		{StatusUnspecified, ""},
	}

	for _, tt := range tests {
		got := tt.status.String()
		if got != tt.want {
			t.Errorf("Status(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestGlossaryConstruction(t *testing.T) {
	t.Parallel()

	g := Glossary{
		Dialect:    DialectLinguist,
		Style:      StyleDCT,
		SourceDesc: "test glossary",
		Concepts: []Concept{
			{
				ID:           "tzimtzum",
				SubjectField: "kabbalah",
				Languages: map[string]LangSection{
					"he": {
						Lang: "he",
						Terms: []Term{
							{
								Surface:              "צמצום",
								AdministrativeStatus: StatusPreferred,
								PartOfSpeech:         "noun",
							},
						},
					},
					"en": {
						Lang: "en",
						Terms: []Term{
							{
								Surface:              "tzimtzum",
								AdministrativeStatus: StatusPreferred,
							},
							{
								Surface:              "contraction",
								AdministrativeStatus: StatusDeprecated,
							},
						},
					},
				},
			},
		},
	}

	if g.Dialect != DialectLinguist {
		t.Errorf("Dialect = %q, want %q", g.Dialect, DialectLinguist)
	}
	if g.Style != StyleDCT {
		t.Errorf("Style = %v, want %v", g.Style, StyleDCT)
	}
	if g.SourceDesc != "test glossary" {
		t.Errorf("SourceDesc = %q, want %q", g.SourceDesc, "test glossary")
	}
	if len(g.Concepts) != 1 {
		t.Fatalf("len(Concepts) = %d, want 1", len(g.Concepts))
	}

	c := g.Concepts[0]
	if c.ID != "tzimtzum" {
		t.Errorf("Concept.ID = %q, want %q", c.ID, "tzimtzum")
	}
	if c.SubjectField != "kabbalah" {
		t.Errorf("Concept.SubjectField = %q, want %q", c.SubjectField, "kabbalah")
	}

	he, ok := c.Languages["he"]
	if !ok {
		t.Fatal("Languages[\"he\"] not found")
	}
	if he.Lang != "he" {
		t.Errorf("LangSection.Lang = %q, want %q", he.Lang, "he")
	}
	if len(he.Terms) != 1 {
		t.Fatalf("len(he.Terms) = %d, want 1", len(he.Terms))
	}
	if he.Terms[0].Surface != "צמצום" {
		t.Errorf("Term.Surface = %q, want %q", he.Terms[0].Surface, "צמצום")
	}
	if he.Terms[0].AdministrativeStatus != StatusPreferred {
		t.Errorf("Term.AdministrativeStatus = %d, want StatusPreferred", he.Terms[0].AdministrativeStatus)
	}
	if he.Terms[0].PartOfSpeech != "noun" {
		t.Errorf("Term.PartOfSpeech = %q, want %q", he.Terms[0].PartOfSpeech, "noun")
	}

	en, ok := c.Languages["en"]
	if !ok {
		t.Fatal("Languages[\"en\"] not found")
	}
	if len(en.Terms) != 2 {
		t.Fatalf("len(en.Terms) = %d, want 2", len(en.Terms))
	}
	if en.Terms[1].AdministrativeStatus != StatusDeprecated {
		t.Errorf("en.Terms[1].AdministrativeStatus = %d, want StatusDeprecated", en.Terms[1].AdministrativeStatus)
	}
}

func TestConceptAllFields(t *testing.T) {
	t.Parallel()

	c := Concept{
		ID:             "test-concept",
		SubjectField:   "mysticism",
		Definitions:    []NoteText{{Plain: "A concept", Raw: "<hi>A</hi> concept"}},
		CrossRefs:      []CrossRef{{Target: "other-id", Label: "see also"}},
		ExternalRefs:   []string{"https://example.com"},
		Graphics:       []string{"https://example.com/image.png"},
		Sources:        []string{"Book of Formation"},
		CustomerSubset: "cust1",
		ProjectSubset:  "proj1",
		Transactions: []Transaction{{
			Type:           "origination",
			Date:           "2026-01-01T00:00:00Z",
			Responsibility: "author",
		}},
		Notes:     []string{"editorial note"},
		Languages: map[string]LangSection{},
	}

	if len(c.Definitions) != 1 {
		t.Fatalf("len(Definitions) = %d, want 1", len(c.Definitions))
	}
	if c.Definitions[0].Plain != "A concept" {
		t.Errorf("Definitions[0].Plain = %q, want %q", c.Definitions[0].Plain, "A concept")
	}
	if c.Definitions[0].Raw != "<hi>A</hi> concept" {
		t.Errorf("Definitions[0].Raw = %q, want %q", c.Definitions[0].Raw, "<hi>A</hi> concept")
	}

	if len(c.CrossRefs) != 1 {
		t.Fatalf("len(CrossRefs) = %d, want 1", len(c.CrossRefs))
	}
	if c.CrossRefs[0].Target != "other-id" {
		t.Errorf("CrossRef.Target = %q, want %q", c.CrossRefs[0].Target, "other-id")
	}
	if c.CrossRefs[0].Label != "see also" {
		t.Errorf("CrossRef.Label = %q, want %q", c.CrossRefs[0].Label, "see also")
	}

	if len(c.ExternalRefs) != 1 || c.ExternalRefs[0] != "https://example.com" {
		t.Errorf("ExternalRefs = %v, want [https://example.com]", c.ExternalRefs)
	}
	if len(c.Graphics) != 1 || c.Graphics[0] != "https://example.com/image.png" {
		t.Errorf("Graphics = %v, want [https://example.com/image.png]", c.Graphics)
	}
	if len(c.Sources) != 1 || c.Sources[0] != "Book of Formation" {
		t.Errorf("Sources = %v, want [Book of Formation]", c.Sources)
	}
	if c.CustomerSubset != "cust1" {
		t.Errorf("CustomerSubset = %q, want %q", c.CustomerSubset, "cust1")
	}
	if c.ProjectSubset != "proj1" {
		t.Errorf("ProjectSubset = %q, want %q", c.ProjectSubset, "proj1")
	}
	if len(c.Notes) != 1 || c.Notes[0] != "editorial note" {
		t.Errorf("Notes = %v, want [editorial note]", c.Notes)
	}
}

func TestCrossRefAndTransaction(t *testing.T) {
	t.Parallel()

	cr := CrossRef{Target: "x", Label: "y"}
	if cr.Target != "x" {
		t.Errorf("CrossRef.Target = %q, want %q", cr.Target, "x")
	}
	if cr.Label != "y" {
		t.Errorf("CrossRef.Label = %q, want %q", cr.Label, "y")
	}

	tx := Transaction{
		Type:           "origination",
		Date:           "2026-01-01T00:00:00Z",
		Responsibility: "author",
	}
	if tx.Type != "origination" {
		t.Errorf("Transaction.Type = %q, want %q", tx.Type, "origination")
	}
	if tx.Date != "2026-01-01T00:00:00Z" {
		t.Errorf("Transaction.Date = %q, want %q", tx.Date, "2026-01-01T00:00:00Z")
	}
	if tx.Responsibility != "author" {
		t.Errorf("Transaction.Responsibility = %q, want %q", tx.Responsibility, "author")
	}
}

func TestTermAllFields(t *testing.T) {
	t.Parallel()

	term := Term{
		Surface:              "tzimtzum",
		AdministrativeStatus: StatusPreferred,
		PartOfSpeech:         "noun",
		GrammaticalGender:    "masculine",
		GrammaticalNumber:    "singular",
		Register:             "technicalRegister",
		TermType:             "fullForm",
		TermLocation:         "menuItem",
		GeographicalUsage:    "IL",
		Contexts:             []NoteText{{Plain: "context text", Raw: "context text"}},
		TransferComment:      "keep as loanword",
		Reading:              "tsimtsum",
		ReadingNote:          "Ashkenazi pronunciation",
		Sources:              []string{"source1"},
		CustomerSubset:       "cust1",
		ProjectSubset:        "proj1",
		ExternalRefs:         []string{"https://example.com/term"},
		CrossRefs:            []CrossRef{{Target: "related-id", Label: "variant of"}},
		Transactions:         []Transaction{{Type: "modification", Date: "2026-06-01T00:00:00Z", Responsibility: "editor"}},
		Notes:                []string{"term note"},
	}

	if term.Surface != "tzimtzum" {
		t.Errorf("Surface = %q, want %q", term.Surface, "tzimtzum")
	}
	if term.AdministrativeStatus != StatusPreferred {
		t.Errorf("AdministrativeStatus = %d, want StatusPreferred", term.AdministrativeStatus)
	}
	if term.GrammaticalGender != "masculine" {
		t.Errorf("GrammaticalGender = %q, want %q", term.GrammaticalGender, "masculine")
	}
	if term.GrammaticalNumber != "singular" {
		t.Errorf("GrammaticalNumber = %q, want %q", term.GrammaticalNumber, "singular")
	}
	if term.Register != "technicalRegister" {
		t.Errorf("Register = %q, want %q", term.Register, "technicalRegister")
	}
	if term.TermType != "fullForm" {
		t.Errorf("TermType = %q, want %q", term.TermType, "fullForm")
	}
	if term.TermLocation != "menuItem" {
		t.Errorf("TermLocation = %q, want %q", term.TermLocation, "menuItem")
	}
	if term.GeographicalUsage != "IL" {
		t.Errorf("GeographicalUsage = %q, want %q", term.GeographicalUsage, "IL")
	}
	if len(term.Contexts) != 1 || term.Contexts[0].Plain != "context text" {
		t.Errorf("Contexts = %v, want [{Plain: context text}]", term.Contexts)
	}
	if term.TransferComment != "keep as loanword" {
		t.Errorf("TransferComment = %q, want %q", term.TransferComment, "keep as loanword")
	}
	if term.Reading != "tsimtsum" {
		t.Errorf("Reading = %q, want %q", term.Reading, "tsimtsum")
	}
	if term.ReadingNote != "Ashkenazi pronunciation" {
		t.Errorf("ReadingNote = %q, want %q", term.ReadingNote, "Ashkenazi pronunciation")
	}
	if len(term.Sources) != 1 || term.Sources[0] != "source1" {
		t.Errorf("Sources = %v, want [source1]", term.Sources)
	}
	if len(term.ExternalRefs) != 1 || term.ExternalRefs[0] != "https://example.com/term" {
		t.Errorf("ExternalRefs = %v, want [https://example.com/term]", term.ExternalRefs)
	}
	if len(term.CrossRefs) != 1 || term.CrossRefs[0].Target != "related-id" {
		t.Errorf("CrossRefs = %v, want [{Target: related-id}]", term.CrossRefs)
	}
	if len(term.Transactions) != 1 || term.Transactions[0].Type != "modification" {
		t.Errorf("Transactions = %v, want [{Type: modification}]", term.Transactions)
	}
	if len(term.Notes) != 1 || term.Notes[0] != "term note" {
		t.Errorf("Notes = %v, want [term note]", term.Notes)
	}
}

func TestLangSectionFields(t *testing.T) {
	t.Parallel()

	ls := LangSection{
		Lang:        "es",
		Definitions: []NoteText{{Plain: "definición", Raw: "definición"}},
		Sources:     []string{"RAE"},
		Terms: []Term{
			{Surface: "tzimtzum", AdministrativeStatus: StatusPreferred},
		},
	}

	if ls.Lang != "es" {
		t.Errorf("Lang = %q, want %q", ls.Lang, "es")
	}
	if len(ls.Definitions) != 1 || ls.Definitions[0].Plain != "definición" {
		t.Errorf("Definitions = %v, want [{Plain: definición}]", ls.Definitions)
	}
	if len(ls.Sources) != 1 || ls.Sources[0] != "RAE" {
		t.Errorf("Sources = %v, want [RAE]", ls.Sources)
	}
	if len(ls.Terms) != 1 || ls.Terms[0].Surface != "tzimtzum" {
		t.Errorf("Terms = %v, want [{Surface: tzimtzum}]", ls.Terms)
	}
}

func TestNoteTextPlainAndRaw(t *testing.T) {
	t.Parallel()

	nt := NoteText{
		Plain: "A definition with emphasis",
		Raw:   "A definition with <hi>emphasis</hi>",
	}

	if nt.Plain != "A definition with emphasis" {
		t.Errorf("NoteText.Plain = %q, want %q", nt.Plain, "A definition with emphasis")
	}
	if nt.Raw != "A definition with <hi>emphasis</hi>" {
		t.Errorf("NoteText.Raw = %q, want %q", nt.Raw, "A definition with <hi>emphasis</hi>")
	}
}
