package write

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
)

type Mutator func(g *tbx.Glossary) (*tbx.Concept, error)

var fatalWarningCodes = map[string]bool{
	"duplicate_id":        true,
	"unresolved_crossref": true,
	"missing_term":        true,
	"invalid_lang_tag":    true,
}

func Execute(path string, mutate Mutator, dryRun bool) (*tbx.Concept, error) {
	g, _, err := tbx.Load(path)
	if err != nil {
		return nil, err
	}

	affected, err := mutate(g)
	if err != nil {
		return nil, err
	}

	if err := validateForWrite(g); err != nil {
		return nil, err
	}

	if !dryRun {
		if err := tbx.Save(path, g); err != nil {
			return nil, err
		}
	}

	return affected, nil
}

func ParseTBXFragment(data []byte) ([]tbx.Concept, error) {
	rootName, err := firstElementName(data)
	if err != nil {
		return nil, ErrInvalidInput.Wrap(err)
	}

	switch rootName {
	case "tbx":
		return nil, ErrInvalidInput.Wrap(
			fmt.Errorf("full <tbx> document not accepted; use apply --file for full-file ingest"),
		)
	case "conceptEntry", "conceptEntryList":
		// accepted
	default:
		return nil, ErrInvalidInput.Wrap(
			fmt.Errorf("unexpected root element <%s>; expected <conceptEntry> or <conceptEntryList>", rootName),
		)
	}

	wrapped := wrapInTBXShell(data, rootName)

	r, err := tbx.ReaderForDialect(tbx.DialectLinguist)
	if err != nil {
		return nil, fmt.Errorf("getting reader: %w", err)
	}

	g, _, err := r.Decode(bytes.NewReader(wrapped))
	if err != nil {
		return nil, ErrInvalidInput.Wrap(err)
	}

	return g.Concepts, nil
}

func firstElementName(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = true
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return "", fmt.Errorf("no XML element found")
		}
		if err != nil {
			return "", err
		}
		if se, ok := tok.(xml.StartElement); ok {
			return se.Name.Local, nil
		}
	}
}

const tbxShellPrefix = `<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader>
    <fileDesc>
      <sourceDesc><p>fragment import</p></sourceDesc>
    </fileDesc>
  </tbxHeader>
  <text>
    <body>
`

const tbxShellSuffix = `
    </body>
  </text>
</tbx>`

func wrapInTBXShell(fragment []byte, rootName string) []byte {
	var buf bytes.Buffer
	buf.WriteString(tbxShellPrefix)

	if rootName == "conceptEntryList" {
		buf.Write(extractListInner(fragment))
	} else {
		buf.Write(fragment)
	}

	buf.WriteString(tbxShellSuffix)
	return buf.Bytes()
}

func extractListInner(data []byte) []byte {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = true

	var collecting bool
	var entries []byte

	for {
		off := dec.InputOffset()
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if !collecting && t.Name.Local == "conceptEntryList" {
				collecting = true
				continue
			}
			if collecting && t.Name.Local == "conceptEntry" {
				entryStart := off
				if err := dec.Skip(); err != nil {
					break
				}
				entryEnd := dec.InputOffset()
				entries = append(entries, data[entryStart:entryEnd]...)
				entries = append(entries, '\n')
			}
		}
	}
	return entries
}

func ParseJSONInput(data []byte) (*output.WriteResult, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var result output.WriteResult
	if err := dec.Decode(&result); err != nil {
		return nil, ErrInvalidInput.Wrap(err)
	}
	return &result, nil
}

func validateForWrite(g *tbx.Glossary) error {
	res := g.Validate(false)

	if len(res.Errors) > 0 {
		return tbx.ErrValidationError.Wrap(
			fmt.Errorf("%d validation error(s)", len(res.Errors)),
		)
	}

	for _, w := range res.Warnings {
		if fatalWarningCodes[w.Code] {
			return tbx.ErrValidationError.Wrap(
				fmt.Errorf("post-mutation validation: %s", w.Message),
			)
		}
	}

	return nil
}

type ApplyValidationError struct {
	Failures []output.ApplyFailure
}

func (e *ApplyValidationError) Error() string {
	return fmt.Sprintf("%d concept(s) failed validation; no changes written", len(e.Failures))
}

func (e *ApplyValidationError) Code() string  { return ErrApplyValidationFailed.Code() }
func (e *ApplyValidationError) ExitCode() int { return ErrApplyValidationFailed.ExitCode() }
func (e *ApplyValidationError) Hint() string  { return ErrApplyValidationFailed.Hint() }
func (e *ApplyValidationError) ErrorDetails() any {
	return map[string][]output.ApplyFailure{"failures": e.Failures}
}

func validateForApply(g *tbx.Glossary) error {
	res := g.Validate(false)

	if len(res.Errors) > 0 {
		return tbx.ErrValidationError.Wrap(
			fmt.Errorf("%d validation error(s)", len(res.Errors)),
		)
	}

	var failures []output.ApplyFailure
	for _, w := range res.Warnings {
		if fatalWarningCodes[w.Code] {
			failures = append(failures, output.ApplyFailure{
				ConceptID: w.ConceptID,
				Code:      w.Code,
				Message:   w.Message,
			})
		}
	}

	if len(failures) > 0 {
		return &ApplyValidationError{Failures: failures}
	}

	return nil
}
