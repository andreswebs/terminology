package write

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// FieldError describes a single JSON decode failure in terms an agent can act
// on: which path failed, what was expected, and what was actually present.
type FieldError struct {
	// Path is the dotted location of the offending value (e.g. "definitions"
	// or "concepts.definitions"). Empty when the decoder could not attribute
	// the failure to a specific field.
	Path string `json:"path,omitempty"`
	// Expected names the type the schema requires at Path (e.g. "string",
	// "object", "array"). Empty for syntax and unknown-field failures.
	Expected string `json:"expected,omitempty"`
	// Actual names the JSON value kind that was found at Path (e.g. "object",
	// "number", "bool"). Empty for syntax and unknown-field failures.
	Actual string `json:"actual,omitempty"`
	// Kind classifies the failure: "type_mismatch", "unknown_field", or
	// "syntax".
	Kind string `json:"kind"`
	// Offset is the byte offset into the payload where the failure was
	// detected. Zero when the decoder did not report a position.
	Offset int `json:"offset,omitempty"`
}

// InvalidInputError is a field-level JSON decode error. It reuses the
// code/exit/hint of ErrInvalidInput and unwraps to it (so
// errors.Is(err, ErrInvalidInput) holds), while adding a specific message and
// structured FieldError details for the error envelope.
type InvalidInputError struct {
	msg    string
	detail FieldError
}

func (e *InvalidInputError) Error() string { return e.msg }

// Code returns the error code shared with ErrInvalidInput.
func (e *InvalidInputError) Code() string { return ErrInvalidInput.Code() }

// ExitCode returns the process exit code shared with ErrInvalidInput.
func (e *InvalidInputError) ExitCode() int { return ErrInvalidInput.ExitCode() }

// Hint returns the remediation hint shared with ErrInvalidInput.
func (e *InvalidInputError) Hint() string { return ErrInvalidInput.Hint() }

// Unwrap returns the ErrInvalidInput sentinel so errors.Is classification and
// exit-code extraction keep working against the shared sentinel identity.
func (e *InvalidInputError) Unwrap() error { return ErrInvalidInput }

// ErrorDetails satisfies output.Detailed so the envelope emits error.details.
func (e *InvalidInputError) ErrorDetails() any { return e.detail }

// describeJSONError classifies an encoding/json decode error into a
// field-level InvalidInputError. Errors it does not recognize fall back to the
// generic ErrInvalidInput wrap, so no decode error is left worse than before.
func describeJSONError(err error) error {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		path := typeErr.Field
		if path == "" {
			path = typeErr.Struct
		}
		expected := friendlyType(typeErr.Type)
		actual := typeErr.Value
		msg := fmt.Sprintf("expected %s, got %s", expected, actual)
		if path != "" {
			msg = fmt.Sprintf("%s: %s", path, msg)
		}
		return &InvalidInputError{
			msg: msg,
			detail: FieldError{
				Path:     path,
				Expected: expected,
				Actual:   actual,
				Kind:     "type_mismatch",
				Offset:   int(typeErr.Offset),
			},
		}
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return &InvalidInputError{
			msg: fmt.Sprintf("malformed JSON at offset %d", syntaxErr.Offset),
			detail: FieldError{
				Kind:   "syntax",
				Offset: int(syntaxErr.Offset),
			},
		}
	}

	// Truncated input (e.g. "{") surfaces as io.ErrUnexpectedEOF rather than a
	// *json.SyntaxError, but is still a syntax-level problem.
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return &InvalidInputError{
			msg:    "malformed JSON: unexpected end of input",
			detail: FieldError{Kind: "syntax"},
		}
	}

	if name, ok := unknownFieldName(err); ok {
		return &InvalidInputError{
			msg: fmt.Sprintf("unknown field %q", name),
			detail: FieldError{
				Path: name,
				Kind: "unknown_field",
			},
		}
	}

	return ErrInvalidInput.Wrap(err)
}

// unknownFieldName extracts the field name from the DisallowUnknownFields
// error. encoding/json exposes no typed error for this case, so the stable
// message format `json: unknown field "<name>"` is matched by prefix.
func unknownFieldName(err error) (string, bool) {
	const prefix = `json: unknown field "`
	msg := err.Error()
	if !strings.HasPrefix(msg, prefix) {
		return "", false
	}
	rest := msg[len(prefix):]
	end := strings.IndexByte(rest, '"')
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}

// friendlyType maps a Go reflect.Type to a JSON-flavored type name for
// human-facing messages.
func friendlyType(t reflect.Type) string {
	if t == nil {
		return "value"
	}
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	case reflect.Pointer:
		return friendlyType(t.Elem())
	default:
		return t.String()
	}
}
