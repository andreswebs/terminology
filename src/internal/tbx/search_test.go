package tbx

import (
	"reflect"
	"testing"
)

func aikidoGlossary() *Glossary {
	return &Glossary{
		Dialect: DialectLinguist,
		Style:   StyleDCT,
		Concepts: []Concept{
			{
				ID:           "kokyu-ho",
				SubjectField: "aikido",
				Languages: map[string]LangSection{
					"ja": {
						Lang: "ja",
						Terms: []Term{
							{
								Surface:              "呼吸法",
								AdministrativeStatus: StatusPreferred,
								Reading:              "こきゅうほう",
								ReadingNote:          "kokyū-ho",
							},
						},
					},
					"en": {
						Lang:  "en",
						Terms: []Term{{Surface: "breathing method", AdministrativeStatus: StatusPreferred}},
					},
				},
			},
			{
				ID:           "irimi-nage",
				SubjectField: "aikido",
				Languages: map[string]LangSection{
					"ja": {
						Lang: "ja",
						Terms: []Term{
							{
								Surface:              "入身投げ",
								AdministrativeStatus: StatusPreferred,
								Reading:              "いりみなげ",
								ReadingNote:          "irimi-nage",
							},
						},
					},
					"en": {
						Lang:  "en",
						Terms: []Term{{Surface: "entering throw", AdministrativeStatus: StatusPreferred}},
					},
				},
			},
			{
				ID:           "katate-dori",
				SubjectField: "aikido",
				Definitions:  []NoteText{{Plain: "A one-handed grab of the wrist."}},
				Languages: map[string]LangSection{
					"ja": {
						Lang: "ja",
						Terms: []Term{
							{
								Surface:              "片手取り",
								AdministrativeStatus: StatusPreferred,
								Reading:              "かたてどり",
								ReadingNote:          "katate-dori",
							},
						},
					},
					"en": {
						Lang:  "en",
						Terms: []Term{{Surface: "single hand grab", AdministrativeStatus: StatusPreferred}},
					},
				},
			},
		},
	}
}

func searchIDs(concepts []Concept) []string {
	ids := make([]string, len(concepts))
	for i, c := range concepts {
		ids[i] = c.ID
	}
	return ids
}

func TestSearch_RomajiWithoutHyphenMatchesReadingNote(t *testing.T) {
	g := aikidoGlossary()
	got := searchIDs(g.Search("katatedori", SearchOptions{}))
	if want := []string{"katate-dori"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Search(katatedori) = %v, want %v", got, want)
	}
}

func TestSearch_DiacriticAndSeparatorFolding(t *testing.T) {
	g := aikidoGlossary()

	if got := searchIDs(g.Search("kokyu", SearchOptions{})); !reflect.DeepEqual(got, []string{"kokyu-ho"}) {
		t.Errorf("Search(kokyu) = %v, want [kokyu-ho] (macron fold)", got)
	}
	if got := searchIDs(g.Search("grab", SearchOptions{})); !reflect.DeepEqual(got, []string{"katate-dori"}) {
		t.Errorf("Search(grab) = %v, want [katate-dori] (substring in en term)", got)
	}
}

func TestSearch_CJKAndKanaQueries(t *testing.T) {
	g := aikidoGlossary()

	if got := searchIDs(g.Search("片手取り", SearchOptions{})); !reflect.DeepEqual(got, []string{"katate-dori"}) {
		t.Errorf("Search(kanji) = %v, want [katate-dori]", got)
	}
	if got := searchIDs(g.Search("かたてどり", SearchOptions{})); !reflect.DeepEqual(got, []string{"katate-dori"}) {
		t.Errorf("Search(hiragana) = %v, want [katate-dori]", got)
	}
}

func TestSearch_DefaultHaystackExcludesDescriptiveText(t *testing.T) {
	g := aikidoGlossary()

	if got := g.Search("wrist", SearchOptions{}); len(got) != 0 {
		t.Errorf("Search(wrist) default = %v, want no hits (definition excluded)", searchIDs(got))
	}
	got := searchIDs(g.Search("wrist", SearchOptions{Include: []string{"definitions"}}))
	if !reflect.DeepEqual(got, []string{"katate-dori"}) {
		t.Errorf("Search(wrist, include definitions) = %v, want [katate-dori]", got)
	}
}

func TestSearch_DeterministicSortByID(t *testing.T) {
	g := aikidoGlossary() // input order: kokyu-ho, irimi-nage, katate-dori
	got := searchIDs(g.Search("e", SearchOptions{}))
	want := []string{"irimi-nage", "katate-dori", "kokyu-ho"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Search(e) = %v, want %v (sorted by concept_id)", got, want)
	}
}

func TestSearch_LangRestrictsHaystack(t *testing.T) {
	g := aikidoGlossary()

	if got := g.Search("grab", SearchOptions{Lang: "ja"}); len(got) != 0 {
		t.Errorf("Search(grab, lang=ja) = %v, want no hits (en-only match excluded)", searchIDs(got))
	}
	if got := searchIDs(g.Search("片手取り", SearchOptions{Lang: "ja"})); !reflect.DeepEqual(got, []string{"katate-dori"}) {
		t.Errorf("Search(kanji, lang=ja) = %v, want [katate-dori]", got)
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	g := aikidoGlossary()
	if got := g.Search("", SearchOptions{}); len(got) != 0 {
		t.Errorf("Search(empty) = %v, want no hits", searchIDs(got))
	}
}

func TestFoldForSearch(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"katate-dori", "katatedori"},
		{"kokyū-ho", "kokyuho"},
		{"Single Hand Grab", "singlehandgrab"},
		{"", ""},
		{"片手取り", "片手取り"},
		{"  ", ""},
	}
	for _, c := range cases {
		if got := foldForSearch(c.in); got != c.want {
			t.Errorf("foldForSearch(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
