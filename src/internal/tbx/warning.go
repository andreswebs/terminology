package tbx

// Warning is a diagnostic produced during loading or validation, locating the
// affected concept and source position.
type Warning struct {
	Code      string
	Message   string
	ConceptID string
	Line, Col int
}
