package output

import (
	"encoding/json"
	"fmt"
	"io"
)

func EmitJSON(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	return err
}

// EmitProjected writes envelope to w as JSON. When fieldsStr is empty it emits
// the full envelope via EmitJSON; otherwise it validates the requested field
// paths against envelope, projects the marshaled JSON down to those paths, and
// writes the result. A field-validation failure is returned unwrapped so its
// exit-code semantics reach the caller.
func EmitProjected(w io.Writer, envelope any, fieldsStr string) error {
	if fieldsStr == "" {
		return EmitJSON(w, envelope)
	}

	fields, err := ValidateFields(fieldsStr, envelope)
	if err != nil {
		return err
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshaling output: %w", err)
	}

	projected, err := ProjectFields(data, fields)
	if err != nil {
		return fmt.Errorf("projecting fields: %w", err)
	}

	if _, err := w.Write(projected); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}
