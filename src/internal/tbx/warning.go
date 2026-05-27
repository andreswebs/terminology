package tbx

type Warning struct {
	Code      string
	Message   string
	ConceptID string
	Line, Col int
}
