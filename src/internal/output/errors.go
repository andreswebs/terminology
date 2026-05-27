package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/andreswebs/terminology/internal/terr"
)

var ErrInvalidField = terr.New(
	"invalid_field", 2,
	"see `terminology schema --command CMD` for valid paths",
	"unknown field path",
)

type errorEnvelope struct {
	SchemaVersion int          `json:"schema_version"`
	OK            bool         `json:"ok"`
	Error         *errorDetail `json:"error"`
}

type Detailed interface {
	ErrorDetails() any
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
	Details any    `json:"details,omitempty"`
}

func EmitError(w io.Writer, format string, err error) {
	code, message, hint := "internal_error", err.Error(), ""

	var coded terr.Coded
	if errors.As(err, &coded) {
		code = coded.Code()
		message = coded.Error()
		hint = coded.Hint()
	} else if usageCode, usageHint := classifyUsageError(err); usageCode != "" {
		code = usageCode
		hint = usageHint
	}

	var details any
	var detailed Detailed
	if errors.As(err, &detailed) {
		details = detailed.ErrorDetails()
	}

	switch format {
	case "text":
		_, _ = fmt.Fprintf(w, "✗ %s\n", message)
		if hint != "" {
			_, _ = fmt.Fprintf(w, "  hint: %s\n", hint)
		}
	default:
		env := errorEnvelope{
			SchemaVersion: SchemaVersion,
			OK:            false,
			Error: &errorDetail{
				Code:    code,
				Message: message,
				Hint:    hint,
				Details: details,
			},
		}
		data, _ := json.Marshal(env)
		_, _ = w.Write(data)
		_, _ = w.Write([]byte("\n"))
	}
}

func ExitCodeFor(err error) int {
	var coded terr.Coded
	if errors.As(err, &coded) {
		return coded.ExitCode()
	}
	if code, _ := classifyUsageError(err); code != "" {
		return 2
	}
	return 1
}

func classifyUsageError(err error) (code string, hint string) {
	msg := err.Error()
	switch {
	case strings.HasPrefix(msg, "flag provided but not defined"):
		return "unknown_flag", "check available flags with --help"
	case strings.HasPrefix(msg, "invalid value") || strings.HasPrefix(msg, "could not parse"):
		return "invalid_value", ""
	case strings.HasPrefix(msg, "Required flag") || strings.HasPrefix(msg, "Required flags"):
		return "missing_required_flag", "check required flags with --help"
	case msg == "cant duplicate this flag":
		return "duplicate_flag", ""
	case strings.HasPrefix(msg, "option") && strings.Contains(msg, "cannot be set along with option"):
		return "mutually_exclusive_flags", ""
	case strings.HasPrefix(msg, "one of these flags needs to be provided"):
		return "missing_required_flag", "check required flags with --help"
	case strings.HasPrefix(msg, "sufficient count of arg"):
		return "missing_argument", "check required arguments with --help"
	}
	return "", ""
}
