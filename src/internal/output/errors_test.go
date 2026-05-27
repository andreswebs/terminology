package output_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/terr"
)

func TestEmitError_JSON_TerrCoded(t *testing.T) {
	var buf bytes.Buffer
	err := terr.Newf("test_error", 65, "try again", "something went wrong")
	output.EmitError(&buf, "json", err)

	want := `{"schema_version":1,"ok":false,"error":{"code":"test_error","message":"something went wrong","hint":"try again"}}` + "\n"
	if got := buf.String(); got != want {
		t.Errorf("EmitError JSON =\n%s\nwant:\n%s", got, want)
	}
}

func TestEmitError_JSON_OmitsEmptyHint(t *testing.T) {
	var buf bytes.Buffer
	err := terr.New("x", 2, "", "msg")
	output.EmitError(&buf, "json", err)

	if bytes.Contains(buf.Bytes(), []byte(`"hint"`)) {
		t.Errorf("expected no hint key in output, got: %s", buf.String())
	}
}

func TestEmitError_JSON_UnknownError(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New("boom")
	output.EmitError(&buf, "json", err)

	want := `{"schema_version":1,"ok":false,"error":{"code":"internal_error","message":"boom"}}` + "\n"
	if got := buf.String(); got != want {
		t.Errorf("EmitError JSON fallback =\n%s\nwant:\n%s", got, want)
	}
}

func TestEmitError_Text_TerrCoded(t *testing.T) {
	var buf bytes.Buffer
	err := terr.Newf("test_error", 65, "try again", "something went wrong")
	output.EmitError(&buf, "text", err)

	want := "✗ something went wrong\n  hint: try again\n"
	if got := buf.String(); got != want {
		t.Errorf("EmitError text =\n%q\nwant:\n%q", got, want)
	}
}

func TestEmitError_Text_OmitsEmptyHint(t *testing.T) {
	var buf bytes.Buffer
	err := terr.New("x", 2, "", "something failed")
	output.EmitError(&buf, "text", err)

	want := "✗ something failed\n"
	if got := buf.String(); got != want {
		t.Errorf("EmitError text no hint =\n%q\nwant:\n%q", got, want)
	}
}

func TestEmitError_FallbackFormat(t *testing.T) {
	var buf bytes.Buffer
	err := terr.Newf("test_error", 65, "try again", "something went wrong")
	output.EmitError(&buf, "yaml", err)

	var bufJSON bytes.Buffer
	output.EmitError(&bufJSON, "json", err)

	if got, want := buf.String(), bufJSON.String(); got != want {
		t.Errorf("unknown format should fall back to JSON:\ngot:  %s\nwant: %s", got, want)
	}
}

func TestExitCodeFor_Coded(t *testing.T) {
	err := terr.New("x", 65, "", "m")
	if got := output.ExitCodeFor(err); got != 65 {
		t.Errorf("ExitCodeFor coded = %d, want 65", got)
	}
}

func TestExitCodeFor_NonCoded(t *testing.T) {
	err := errors.New("plain")
	if got := output.ExitCodeFor(err); got != 1 {
		t.Errorf("ExitCodeFor plain = %d, want 1", got)
	}
}

func TestExitCodeFor_UrfaveErrors(t *testing.T) {
	cases := []struct {
		name    string
		message string
		want    int
	}{
		{"unknown_flag", `flag provided but not defined: -bogus`, 2},
		{"invalid_value", `invalid value "yaml" for flag -format: invalid value yaml; accepted: json, text`, 2},
		{"required_flag", `Required flag "file" not set`, 2},
		{"required_flags_plural", `Required flags "lang, term" not set`, 2},
		{"parse_error", `could not parse "abc" as int64 value from SOME_ENV for flag foo: strconv.ParseInt`, 2},
		{"duplicate_flag", `cant duplicate this flag`, 2},
		{"mutually_exclusive", `option foo cannot be set along with option bar`, 2},
		{"insufficient_args", `sufficient count of arg files not provided, given 0 expected 1`, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := errors.New(tc.message)
			if got := output.ExitCodeFor(err); got != tc.want {
				t.Errorf("ExitCodeFor(%q) = %d, want %d", tc.message, got, tc.want)
			}
		})
	}
}

func TestEmitError_UrfaveUnknownFlag(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New(`flag provided but not defined: -bogus`)
	output.EmitError(&buf, "json", err)

	var env map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
		t.Fatalf("not valid JSON: %v", jsonErr)
	}
	errObj := env["error"].(map[string]any)
	if code := errObj["code"].(string); code != "unknown_flag" {
		t.Errorf("code = %q, want %q", code, "unknown_flag")
	}
}

func TestEmitError_UrfaveInvalidValue(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New(`invalid value "yaml" for flag -format: bad`)
	output.EmitError(&buf, "json", err)

	var env map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
		t.Fatalf("not valid JSON: %v", jsonErr)
	}
	errObj := env["error"].(map[string]any)
	if code := errObj["code"].(string); code != "invalid_value" {
		t.Errorf("code = %q, want %q", code, "invalid_value")
	}
}

func TestEmitError_UrfaveRequiredFlag(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New(`Required flag "file" not set`)
	output.EmitError(&buf, "json", err)

	var env map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
		t.Fatalf("not valid JSON: %v", jsonErr)
	}
	errObj := env["error"].(map[string]any)
	if code := errObj["code"].(string); code != "missing_required_flag" {
		t.Errorf("code = %q, want %q", code, "missing_required_flag")
	}
}

func TestEmitError_UrfaveInsufficientArgs(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New(`sufficient count of arg files not provided, given 0 expected 1`)
	output.EmitError(&buf, "json", err)

	var env map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
		t.Fatalf("not valid JSON: %v", jsonErr)
	}
	errObj := env["error"].(map[string]any)
	if code := errObj["code"].(string); code != "missing_argument" {
		t.Errorf("code = %q, want %q", code, "missing_argument")
	}
}

func TestErrInvalidField_Coded(t *testing.T) {
	var coded terr.Coded = output.ErrInvalidField
	if got := coded.Code(); got != "invalid_field" {
		t.Errorf("Code() = %q, want %q", got, "invalid_field")
	}
	if got := coded.ExitCode(); got != 2 {
		t.Errorf("ExitCode() = %d, want 2", got)
	}
}

func TestErrInvalidField_HintAndMessage(t *testing.T) {
	if got := output.ErrInvalidField.Hint(); got == "" || got != "see `terminology schema --command CMD` for valid paths" {
		t.Errorf("Hint() = %q, want hint containing schema reference", got)
	}
	if got := output.ErrInvalidField.Error(); got != "unknown field path" {
		t.Errorf("Error() = %q, want %q", got, "unknown field path")
	}
}

func TestErrInvalidField_WrapPreservesCode(t *testing.T) {
	cause := errors.New("bad path: concpet_id")
	wrapped := output.ErrInvalidField.Wrap(cause)

	var coded terr.Coded
	if !errors.As(wrapped, &coded) {
		t.Fatal("wrapped error does not satisfy terr.Coded")
	}
	if got := coded.Code(); got != "invalid_field" {
		t.Errorf("wrapped Code() = %q, want %q", got, "invalid_field")
	}
	if !errors.Is(wrapped, cause) {
		t.Error("wrapped error does not unwrap to cause")
	}
}

type detailedError struct {
	code    string
	exit    int
	hint    string
	msg     string
	details any
}

func (e *detailedError) Error() string     { return e.msg }
func (e *detailedError) Code() string      { return e.code }
func (e *detailedError) ExitCode() int     { return e.exit }
func (e *detailedError) Hint() string      { return e.hint }
func (e *detailedError) ErrorDetails() any { return e.details }

func TestEmitError_JSON_WithDetails(t *testing.T) {
	var buf bytes.Buffer
	failures := []output.ApplyFailure{
		{ConceptID: "bad-concept", Code: "dangling_crossref", Message: "unresolved ref"},
	}
	err := &detailedError{
		code:    "apply_validation_failed",
		exit:    1,
		hint:    "fix per-concept errors in failures[] and retry",
		msg:     "1 concept(s) failed validation; no changes written",
		details: map[string]any{"failures": failures},
	}
	output.EmitError(&buf, "json", err)

	var env map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr != nil {
		t.Fatalf("not valid JSON: %v\nraw: %s", jsonErr, buf.String())
	}

	errObj, ok := env["error"].(map[string]any)
	if !ok {
		t.Fatalf("missing error object: %s", buf.String())
	}
	if code := errObj["code"].(string); code != "apply_validation_failed" {
		t.Errorf("code = %q, want %q", code, "apply_validation_failed")
	}

	details, ok := errObj["details"].(map[string]any)
	if !ok {
		t.Fatalf("missing details: %s", buf.String())
	}
	failuresArr, ok := details["failures"].([]any)
	if !ok || len(failuresArr) != 1 {
		t.Fatalf("expected 1 failure, got: %v", details["failures"])
	}
	f := failuresArr[0].(map[string]any)
	if f["concept_id"] != "bad-concept" {
		t.Errorf("failure concept_id = %v, want bad-concept", f["concept_id"])
	}
}

func TestEmitError_JSON_NoDetailsWhenNotDetailed(t *testing.T) {
	var buf bytes.Buffer
	err := terr.Newf("test_error", 65, "hint", "msg")
	output.EmitError(&buf, "json", err)

	if bytes.Contains(buf.Bytes(), []byte(`"details"`)) {
		t.Errorf("non-detailed error should not have details key: %s", buf.String())
	}
}

func TestSchemaVersion_IsOne(t *testing.T) {
	if output.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", output.SchemaVersion)
	}
}
