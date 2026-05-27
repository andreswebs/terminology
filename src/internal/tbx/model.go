package tbx

type Glossary struct {
	Dialect    Dialect
	Style      Style
	SourceDesc string
	Concepts   []Concept
}

type Concept struct {
	ID             string
	StartLine      int
	StartCol       int
	SubjectField   string
	Definitions    []NoteText
	CrossRefs      []CrossRef
	ExternalRefs   []string
	Graphics       []string
	Sources        []string
	CustomerSubset string
	ProjectSubset  string
	Transactions   []Transaction
	Notes          []string
	Languages      map[string]LangSection
}

type LangSection struct {
	Lang        string
	StartLine   int
	StartCol    int
	Definitions []NoteText
	Sources     []string
	Terms       []Term
}

type Term struct {
	Surface              string
	AdministrativeStatus Status
	PartOfSpeech         string
	GrammaticalGender    string
	GrammaticalNumber    string
	Register             string
	TermType             string
	TermLocation         string
	GeographicalUsage    string
	Contexts             []NoteText
	TransferComment      string
	Reading              string
	ReadingNote          string
	Sources              []string
	CustomerSubset       string
	ProjectSubset        string
	ExternalRefs         []string
	CrossRefs            []CrossRef
	Transactions         []Transaction
	Notes                []string
}

type Status int

const (
	StatusUnspecified Status = iota
	StatusPreferred
	StatusAdmitted
	StatusDeprecated
	StatusSuperseded
)

func (s Status) String() string {
	switch s {
	case StatusPreferred:
		return "preferredTerm-admn-sts"
	case StatusAdmitted:
		return "admittedTerm-admn-sts"
	case StatusDeprecated:
		return "deprecatedTerm-admn-sts"
	case StatusSuperseded:
		return "supersededTerm-admn-sts"
	default:
		return ""
	}
}

func ParseStatus(s string) Status {
	switch s {
	case "preferredTerm-admn-sts", "preferredTerm":
		return StatusPreferred
	case "admittedTerm-admn-sts", "admittedTerm":
		return StatusAdmitted
	case "deprecatedTerm-admn-sts", "deprecatedTerm":
		return StatusDeprecated
	case "supersededTerm-admn-sts", "supersededTerm":
		return StatusSuperseded
	default:
		return StatusUnspecified
	}
}

type CrossRef struct {
	Target string
	Label  string
}

type Transaction struct {
	Type           string
	Date           string
	Responsibility string
}

type NoteText struct {
	Plain string
	Raw   string
}
