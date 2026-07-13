package write

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
)

// ApplyPayload is the JSON envelope accepted by the `apply` command, matching
// the shape emitted by `export` so exported output can be piped straight back
// in.
type ApplyPayload struct {
	// SchemaVersion and OK are accepted but ignored so that the full envelope
	// emitted by `export` (schema_version + ok + concepts) is directly
	// apply-consumable, enabling `terminology export | terminology apply -f -`.
	SchemaVersion int                  `json:"schema_version"`
	OK            bool                 `json:"ok"`
	Concepts      []output.WriteResult `json:"concepts"`
}

// ParseApplyJSON decodes an ApplyPayload and converts its write results into
// concepts, rejecting unknown fields.
func ParseApplyJSON(data []byte) ([]tbx.Concept, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var payload ApplyPayload
	if err := dec.Decode(&payload); err != nil {
		return nil, describeJSONError(err)
	}

	concepts := make([]tbx.Concept, 0, len(payload.Concepts))
	for i := range payload.Concepts {
		concepts = append(concepts, ResultToConcept(&payload.Concepts[i]))
	}

	return concepts, nil
}

// ResultToConcept converts an output.WriteResult into a tbx.Concept,
// expanding each language group's terms by administrative status.
func ResultToConcept(wr *output.WriteResult) tbx.Concept {
	c := tbx.Concept{
		ID:           wr.ConceptID,
		SubjectField: wr.SubjectField,
		Languages:    make(map[string]tbx.LangSection),
	}

	for _, d := range wr.Definitions {
		c.Definitions = append(c.Definitions, tbx.NoteText{Plain: d})
	}
	for _, cr := range wr.CrossRefs {
		c.CrossRefs = append(c.CrossRefs, tbx.CrossRef{Target: cr.Target, Label: cr.Label})
	}
	c.ExternalRefs = wr.ExternalRefs
	c.Sources = wr.Sources
	c.Notes = wr.Notes

	for lang, grp := range wr.Languages {
		ls := tbx.LangSection{Lang: lang}
		for _, d := range grp.Definitions {
			ls.Definitions = append(ls.Definitions, tbx.NoteText{Plain: d})
		}
		if grp.Preferred != nil {
			ls.Terms = append(ls.Terms, TermToTBXTerm(*grp.Preferred, tbx.StatusPreferred))
		}
		for _, at := range grp.Admitted {
			ls.Terms = append(ls.Terms, TermToTBXTerm(at, tbx.StatusAdmitted))
		}
		for _, dt := range grp.Deprecated {
			ls.Terms = append(ls.Terms, TermToTBXTerm(dt, tbx.StatusDeprecated))
		}
		for _, st := range grp.Superseded {
			ls.Terms = append(ls.Terms, TermToTBXTerm(st, tbx.StatusSuperseded))
		}
		c.Languages[lang] = ls
	}

	return c
}

// TermToTBXTerm converts an output.WriteTerm into a tbx.Term, applying
// defaultStatus unless the term overrides its administrative status.
func TermToTBXTerm(wt output.WriteTerm, defaultStatus tbx.Status) tbx.Term {
	status := defaultStatus
	if wt.AdministrativeStatus != "" {
		status = tbx.ParseStatus(wt.AdministrativeStatus)
	}

	t := tbx.Term{
		Surface:              wt.Term,
		AdministrativeStatus: status,
		PartOfSpeech:         wt.PartOfSpeech,
		GrammaticalGender:    wt.GrammaticalGender,
		GrammaticalNumber:    wt.GrammaticalNumber,
		Register:             wt.Register,
		TermType:             wt.TermType,
		TermLocation:         wt.TermLocation,
		GeographicalUsage:    wt.GeographicalUsage,
		TransferComment:      wt.TransferComment,
		Reading:              wt.Reading,
		ReadingNote:          wt.ReadingNote,
		Sources:              wt.Sources,
		CustomerSubset:       wt.CustomerSubset,
		ProjectSubset:        wt.ProjectSubset,
		ExternalRefs:         wt.ExternalRefs,
		Notes:                wt.Notes,
	}

	for _, ctx := range wt.Contexts {
		t.Contexts = append(t.Contexts, tbx.NoteText{Plain: ctx})
	}
	for _, cr := range wt.CrossRefs {
		t.CrossRefs = append(t.CrossRefs, tbx.CrossRef{Target: cr.Target, Label: cr.Label})
	}

	return t
}

// PayloadFormat identifies the serialization format of an apply payload.
type PayloadFormat int

// Recognized apply payload formats.
const (
	FormatJSON PayloadFormat = iota
	FormatTBX
)

// DetectPayloadFormat determines a payload's format from the file extension
// when available, falling back to sniffing the first non-whitespace byte.
func DetectPayloadFormat(filePath string, data []byte) (PayloadFormat, error) {
	if filePath != "-" && filePath != "" {
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".json":
			return FormatJSON, nil
		case ".tbx", ".xml":
			return FormatTBX, nil
		}
	}

	return sniffFormat(data)
}

func sniffFormat(data []byte) (PayloadFormat, error) {
	for _, b := range data {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '<':
			return FormatTBX, nil
		case '{', '[':
			return FormatJSON, nil
		default:
			return 0, ErrInvalidInput.Wrap(
				fmt.Errorf("cannot detect payload format; first non-whitespace byte is %q", b),
			)
		}
	}
	return 0, ErrInvalidInput.Wrap(fmt.Errorf("empty payload"))
}

// LoadApplyFile reads the payload at filePath (or stdin for "-"), detects its
// format, and parses it into concepts.
func LoadApplyFile(filePath string) ([]tbx.Concept, error) {
	data, err := readPayloadFile(filePath)
	if err != nil {
		return nil, err
	}

	format, err := DetectPayloadFormat(filePath, data)
	if err != nil {
		return nil, err
	}

	switch format {
	case FormatJSON:
		return ParseApplyJSON(data)
	case FormatTBX:
		return ParseTBXFragment(data)
	default:
		return nil, ErrInvalidInput.Wrap(fmt.Errorf("unsupported payload format"))
	}
}

func readPayloadFile(filePath string) ([]byte, error) {
	if filePath == "-" {
		return readStdin()
	}
	data, err := tbx.ReadFileBounded(filePath, tbx.MaxPayloadSize)
	if err != nil {
		if _, ok := err.(interface{ Code() string }); ok {
			return nil, err
		}
		return nil, fmt.Errorf("reading payload file: %w", err)
	}
	return data, nil
}

func readStdin() ([]byte, error) {
	data, err := tbx.ReadBounded(os.Stdin, tbx.MaxStdinSize)
	if err != nil {
		if _, ok := err.(interface{ Code() string }); ok {
			return nil, err
		}
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	return data, nil
}
