package output

import "encoding/json"

func init() {
	RegisterEnvelope("validate", ValidateEnvelope{})
	RegisterEnvelope("lookup", LookupEnvelope{})
	RegisterEnvelope("extract", ExtractEnvelope{})
	RegisterEnvelope("scan", ScanEnvelope{})
	RegisterEnvelope("check", CheckEnvelope{})
	RegisterEnvelope("concept add", WriteEnvelope{})
	RegisterEnvelope("concept update", WriteEnvelope{})
	RegisterEnvelope("concept remove", WriteEnvelope{})
	RegisterEnvelope("term add", WriteEnvelope{})
	RegisterEnvelope("term deprecate", WriteEnvelope{})
	RegisterEnvelope("apply", ApplyEnvelope{})
	RegisterEnvelope("init", InitEnvelope{})
	RegisterEnvelope("search", SearchEnvelope{})
	RegisterEnvelope("export", ExportEnvelope{})
	RegisterEnvelope("show", ShowEnvelope{})
	RegisterEnvelope("list", ListEnvelope{})

	RegisterExitCodes("validate", []int{0, 1, 2, 3, 65})
	RegisterExitCodes("lookup", []int{0, 1, 2, 3, 65})
	RegisterExitCodes("extract", []int{0, 1, 2, 3})
	RegisterExitCodes("scan", []int{0, 2, 3, 65})
	RegisterExitCodes("check", []int{0, 1, 2, 3, 65})
	RegisterExitCodes("schema", []int{0, 2})
	RegisterExitCodes("concept add", []int{0, 2, 3, 65})
	RegisterExitCodes("concept update", []int{0, 2, 3, 65})
	RegisterExitCodes("concept remove", []int{0, 2, 3, 65})
	RegisterExitCodes("term add", []int{0, 2, 3, 65})
	RegisterExitCodes("term deprecate", []int{0, 2, 3, 65})
	RegisterExitCodes("apply", []int{0, 1, 2, 3, 65})
	RegisterExitCodes("init", []int{0, 2, 3, 65})
	RegisterExitCodes("search", []int{0, 1, 2, 3, 65})
	RegisterExitCodes("export", []int{0, 2, 3, 65})
	RegisterExitCodes("show", []int{0, 1, 2, 3, 65})
	RegisterExitCodes("list", []int{0, 2, 3, 65})
}

type ValidateEnvelope struct {
	SchemaVersion int               `json:"schema_version"`
	OK            bool              `json:"ok"`
	Concepts      int               `json:"concepts"`
	Languages     []string          `json:"languages"`
	Warnings      []ValidateWarning `json:"warnings"`
}

type ValidateWarning struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	ConceptID string `json:"concept_id,omitempty"`
	Line      int    `json:"line,omitempty"`
	Col       int    `json:"column,omitempty"`
}

// LookupEnvelope carries the results of the strict exact `lookup` finder. Each
// result is the canonical WriteResult concept shape shared by every read and
// write command, so a lookup hit exposes the same rich fields (definitions,
// readings, contexts, non-preferred variants) that `export`/`show` emit.
type LookupEnvelope struct {
	SchemaVersion int           `json:"schema_version"`
	OK            bool          `json:"ok"`
	Results       []WriteResult `json:"results"`
}

type ExtractEnvelope struct {
	SchemaVersion int                `json:"schema_version"`
	OK            bool               `json:"ok"`
	Candidates    []ExtractCandidate `json:"candidates"`
}

type ExtractCandidate struct {
	Term      string            `json:"term"`
	Frequency int               `json:"frequency"`
	Heuristic string            `json:"heuristic"`
	Locations []ExtractLocation `json:"locations,omitempty"`
}

type ExtractLocation struct {
	File string `json:"file"`
	Line int    `json:"line,omitempty"`
	Col  int    `json:"col,omitempty"`
}

func (e ExtractEnvelope) MarshalJSON() ([]byte, error) {
	type Alias ExtractEnvelope
	a := Alias(e)
	if a.Candidates == nil {
		a.Candidates = []ExtractCandidate{}
	}
	return json.Marshal(a)
}

func (e LookupEnvelope) MarshalJSON() ([]byte, error) {
	type Alias LookupEnvelope
	a := Alias(e)
	if a.Results == nil {
		a.Results = []WriteResult{}
	}
	return json.Marshal(a)
}

type ScanEnvelope struct {
	SchemaVersion int         `json:"schema_version"`
	OK            bool        `json:"ok"`
	File          string      `json:"file"`
	Matches       []ScanMatch `json:"matches"`
	Summary       ScanSummary `json:"summary"`
}

type ScanMatch struct {
	ConceptID string `json:"concept_id"`
	Term      string `json:"term"`
	Lang      string `json:"lang"`
	Status    string `json:"status"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Context   string `json:"context"`
}

type ScanSummary struct {
	TotalMatches   int `json:"total_matches"`
	UniqueConcepts int `json:"unique_concepts"`
}

func (e ScanEnvelope) MarshalJSON() ([]byte, error) {
	type Alias ScanEnvelope
	a := Alias(e)
	if a.Matches == nil {
		a.Matches = []ScanMatch{}
	}
	return json.Marshal(a)
}

type CheckEnvelope struct {
	SchemaVersion int              `json:"schema_version"`
	OK            bool             `json:"ok"`
	Source        string           `json:"source"`
	Target        string           `json:"target"`
	Violations    []CheckViolation `json:"violations"`
	Warnings      []CheckWarning   `json:"warnings"`
	Summary       CheckSummary     `json:"summary"`
}

type CheckViolation struct {
	Type              string `json:"type"`
	ConceptID         string `json:"concept_id"`
	SourceTerm        string `json:"source_term,omitempty"`
	ExpectedTarget    string `json:"expected_target,omitempty"`
	SourceOccurrences int    `json:"source_occurrences,omitempty"`
	Variant           string `json:"variant,omitempty"`
	Line              int    `json:"line,omitempty"`
	Column            int    `json:"column,omitempty"`
	Context           string `json:"context,omitempty"`
}

type CheckWarning struct {
	Type      string `json:"type"`
	ConceptID string `json:"concept_id"`
	Message   string `json:"message,omitempty"`
	Variant   string `json:"variant,omitempty"`
	Line      int    `json:"line,omitempty"`
	Column    int    `json:"column,omitempty"`
	Context   string `json:"context,omitempty"`
}

type CheckSummary struct {
	Violations      int `json:"violations"`
	Warnings        int `json:"warnings"`
	ConceptsChecked int `json:"concepts_checked"`
}

type WriteEnvelope struct {
	SchemaVersion int         `json:"schema_version"`
	OK            bool        `json:"ok"`
	Result        WriteResult `json:"result"`
}

type WriteResult struct {
	ConceptID    string                    `json:"concept_id"`
	SubjectField string                    `json:"subject_field,omitempty"`
	Definitions  []string                  `json:"definitions,omitempty"`
	CrossRefs    []WriteCrossRef           `json:"cross_refs,omitempty"`
	ExternalRefs []string                  `json:"external_refs,omitempty"`
	Sources      []string                  `json:"sources,omitempty"`
	Notes        []string                  `json:"notes,omitempty"`
	Languages    map[string]WriteTermGroup `json:"languages"`
}

type WriteCrossRef struct {
	Target string `json:"target"`
	Label  string `json:"label,omitempty"`
}

type WriteTermGroup struct {
	Definitions []string    `json:"definitions,omitempty"`
	Preferred   *WriteTerm  `json:"preferred,omitempty"`
	Admitted    []WriteTerm `json:"admitted,omitempty"`
	Deprecated  []WriteTerm `json:"deprecated,omitempty"`
	Superseded  []WriteTerm `json:"superseded,omitempty"`
}

type WriteTerm struct {
	Term                 string          `json:"term"`
	AdministrativeStatus string          `json:"administrative_status,omitempty"`
	PartOfSpeech         string          `json:"part_of_speech,omitempty"`
	GrammaticalGender    string          `json:"grammatical_gender,omitempty"`
	GrammaticalNumber    string          `json:"grammatical_number,omitempty"`
	Register             string          `json:"register,omitempty"`
	TermType             string          `json:"term_type,omitempty"`
	TermLocation         string          `json:"term_location,omitempty"`
	GeographicalUsage    string          `json:"geographical_usage,omitempty"`
	TransferComment      string          `json:"transfer_comment,omitempty"`
	Reading              string          `json:"reading,omitempty"`
	ReadingNote          string          `json:"reading_note,omitempty"`
	Contexts             []string        `json:"contexts,omitempty"`
	Sources              []string        `json:"sources,omitempty"`
	CustomerSubset       string          `json:"customer_subset,omitempty"`
	ProjectSubset        string          `json:"project_subset,omitempty"`
	ExternalRefs         []string        `json:"external_refs,omitempty"`
	CrossRefs            []WriteCrossRef `json:"cross_refs,omitempty"`
	Notes                []string        `json:"notes,omitempty"`
}

func (e WriteEnvelope) MarshalJSON() ([]byte, error) {
	type Alias WriteEnvelope
	a := Alias(e)
	if a.Result.Languages == nil {
		a.Result.Languages = make(map[string]WriteTermGroup)
	}
	return json.Marshal(a)
}

type ApplyEnvelope struct {
	SchemaVersion int         `json:"schema_version"`
	OK            bool        `json:"ok"`
	Applied       ApplyResult `json:"applied"`
	Warnings      []string    `json:"warnings"`
}

type ApplyResult struct {
	Added     []string `json:"added"`
	Updated   []string `json:"updated"`
	Removed   []string `json:"removed"`
	Unchanged []string `json:"unchanged"`
}

type ApplyFailure struct {
	ConceptID string `json:"concept_id"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

func (e ApplyEnvelope) MarshalJSON() ([]byte, error) {
	type Alias ApplyEnvelope
	a := Alias(e)
	if a.Applied.Added == nil {
		a.Applied.Added = []string{}
	}
	if a.Applied.Updated == nil {
		a.Applied.Updated = []string{}
	}
	if a.Applied.Removed == nil {
		a.Applied.Removed = []string{}
	}
	if a.Applied.Unchanged == nil {
		a.Applied.Unchanged = []string{}
	}
	if a.Warnings == nil {
		a.Warnings = []string{}
	}
	return json.Marshal(a)
}

type SearchEnvelope struct {
	SchemaVersion int           `json:"schema_version"`
	OK            bool          `json:"ok"`
	Results       []WriteResult `json:"results"`
}

func (e SearchEnvelope) MarshalJSON() ([]byte, error) {
	type Alias SearchEnvelope
	a := Alias(e)
	if a.Results == nil {
		a.Results = []WriteResult{}
	}
	return json.Marshal(a)
}

// ExportEnvelope carries the full glossary as canonical concepts, sorted by
// concept id. The `concepts` payload is exactly what `apply` ingests, enabling
// a read-modify-write round-trip.
type ExportEnvelope struct {
	SchemaVersion int           `json:"schema_version"`
	OK            bool          `json:"ok"`
	Concepts      []WriteResult `json:"concepts"`
}

func (e ExportEnvelope) MarshalJSON() ([]byte, error) {
	type Alias ExportEnvelope
	a := Alias(e)
	if a.Concepts == nil {
		a.Concepts = []WriteResult{}
	}
	return json.Marshal(a)
}

// ShowEnvelope carries a single concept in the canonical shape.
type ShowEnvelope struct {
	SchemaVersion int         `json:"schema_version"`
	OK            bool        `json:"ok"`
	Concept       WriteResult `json:"concept"`
}

func (e ShowEnvelope) MarshalJSON() ([]byte, error) {
	type Alias ShowEnvelope
	a := Alias(e)
	if a.Concept.Languages == nil {
		a.Concept.Languages = make(map[string]WriteTermGroup)
	}
	return json.Marshal(a)
}

// ListEnvelope is the projected enumeration: each concept carries only its id,
// subject field, and per-language preferred term. It is the canonical concept
// element reduced to the fields `export --fields concept_id,subject_field,
// languages.*.preferred.term` would keep.
type ListEnvelope struct {
	SchemaVersion int           `json:"schema_version"`
	OK            bool          `json:"ok"`
	Concepts      []WriteResult `json:"concepts"`
}

func (e ListEnvelope) MarshalJSON() ([]byte, error) {
	type Alias ListEnvelope
	a := Alias(e)
	if a.Concepts == nil {
		a.Concepts = []WriteResult{}
	}
	return json.Marshal(a)
}

type InitEnvelope struct {
	SchemaVersion int    `json:"schema_version"`
	OK            bool   `json:"ok"`
	SourceLang    string `json:"source_lang"`
	Title         string `json:"title,omitempty"`
	DryRun        bool   `json:"dry_run,omitempty"`
}

func (e CheckEnvelope) MarshalJSON() ([]byte, error) {
	type Alias CheckEnvelope
	a := Alias(e)
	if a.Violations == nil {
		a.Violations = []CheckViolation{}
	}
	if a.Warnings == nil {
		a.Warnings = []CheckWarning{}
	}
	return json.Marshal(a)
}
