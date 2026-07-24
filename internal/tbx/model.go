package tbx

// Glossary is an in-memory representation of a TBX terminology collection,
// including its dialect, style, header metadata, and concepts.
type Glossary struct {
	Dialect    Dialect
	Style      Style
	SourceLang string
	Header     Header
	SourceDesc string
	Concepts   []Concept
}

// Header holds the descriptive metadata of a glossary drawn from the TBX
// header section.
type Header struct {
	Title            string
	PublicationStmts []string
	SourceDescs      []string
	EncodingDescs    []string
	RevisionDescs    []string
}

// Concept is a single terminological entry, grouping its language sections
// and shared metadata under one concept ID.
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

// LangSection holds the terms and language-level metadata for one language
// within a concept.
type LangSection struct {
	Lang        string
	StartLine   int
	StartCol    int
	Definitions []NoteText
	Sources     []string
	Terms       []Term
}

// Term is a single term within a language section, carrying its surface form
// and associated linguistic and administrative attributes.
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

// Status is the administrative status of a term.
type Status int

// StatusUnspecified and the following values enumerate the administrative
// statuses a term may carry.
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

// ParseStatus converts a TBX administrative status string into a Status,
// returning StatusUnspecified for unrecognized values.
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

// CrossRef is a reference from a concept or term to another concept.
type CrossRef struct {
	Target string
	Label  string
}

// Transaction records a change event, such as origination or modification,
// with its date and responsible party.
type Transaction struct {
	Type           string
	Date           string
	Responsibility string
}

// NoteText is a piece of descriptive text held both as plain text and in its
// original raw form.
type NoteText struct {
	Plain string
	Raw   string
}
