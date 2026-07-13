package write

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/output"
)

// gap2Payload is the field-feedback reproduction: definitions declared as an
// array of objects where the schema requires an array of strings.
const gap2Payload = `{"concept_id":"x","definitions":[{"text":"d","lang":"en"}],` +
	`"languages":{"en":{"preferred":{"term":"x","administrative_status":"preferredTerm-admn-sts"}}}}`

func codedFrom(t *testing.T, err error) interface {
	Code() string
	Error() string
} {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	coded, ok := err.(interface {
		Code() string
		Error() string
	})
	if !ok {
		t.Fatalf("error %v (%T) does not implement Coded", err, err)
	}
	return coded
}

func TestParseJSONInput_TypeMismatch_NamesPath(t *testing.T) {
	_, err := ParseJSONInput([]byte(gap2Payload))
	coded := codedFrom(t, err)

	if coded.Code() != "invalid_input" {
		t.Errorf("Code() = %q, want invalid_input", coded.Code())
	}
	msg := coded.Error()
	for _, want := range []string{"definitions", "expected string", "object"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message %q does not contain %q", msg, want)
		}
	}
}

func TestParseJSONInput_UnknownKey_Named(t *testing.T) {
	payload := `{"concpet_id":"x","languages":{}}`
	_, err := ParseJSONInput([]byte(payload))
	coded := codedFrom(t, err)

	if coded.Code() != "invalid_input" {
		t.Errorf("Code() = %q, want invalid_input", coded.Code())
	}
	if !strings.Contains(coded.Error(), "concpet_id") {
		t.Errorf("message %q does not name the unknown field concpet_id", coded.Error())
	}
}

func TestParseJSONInput_Syntax_Distinguished(t *testing.T) {
	_, err := ParseJSONInput([]byte("{"))
	coded := codedFrom(t, err)

	if coded.Code() != "invalid_input" {
		t.Errorf("Code() = %q, want invalid_input", coded.Code())
	}
	msg := coded.Error()
	if !strings.Contains(msg, "malformed") && !strings.Contains(msg, "offset") {
		t.Errorf("message %q does not indicate a syntax problem", msg)
	}
}

func TestParseJSONInput_StructuredDetails(t *testing.T) {
	_, err := ParseJSONInput([]byte(gap2Payload))

	// Sentinel identity is preserved for classification/exit-code paths.
	if !errors.Is(err, ErrInvalidInput) {
		t.Error("errors.Is(err, ErrInvalidInput) = false, want true")
	}

	var detailed output.Detailed
	if !errors.As(err, &detailed) {
		t.Fatalf("error %v (%T) does not implement output.Detailed", err, err)
	}
	fe, ok := detailed.ErrorDetails().(FieldError)
	if !ok {
		t.Fatalf("ErrorDetails() = %T, want FieldError", detailed.ErrorDetails())
	}
	if !strings.Contains(fe.Path, "definitions") {
		t.Errorf("FieldError.Path = %q, want to contain definitions", fe.Path)
	}
	if fe.Expected != "string" {
		t.Errorf("FieldError.Expected = %q, want string", fe.Expected)
	}
	if fe.Actual != "object" {
		t.Errorf("FieldError.Actual = %q, want object", fe.Actual)
	}
	if fe.Kind != "type_mismatch" {
		t.Errorf("FieldError.Kind = %q, want type_mismatch", fe.Kind)
	}
}

func TestParseJSONInput_DetailsReachEnvelope(t *testing.T) {
	_, err := ParseJSONInput([]byte(gap2Payload))

	var buf bytes.Buffer
	output.EmitError(&buf, "json", err)

	var env struct {
		Error struct {
			Code    string `json:"code"`
			Details struct {
				Path     string `json:"path"`
				Expected string `json:"expected"`
				Actual   string `json:"actual"`
				Kind     string `json:"kind"`
			} `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.Error.Code != "invalid_input" {
		t.Errorf("envelope code = %q, want invalid_input", env.Error.Code)
	}
	if !strings.Contains(env.Error.Details.Path, "definitions") {
		t.Errorf("envelope details.path = %q, want to contain definitions", env.Error.Details.Path)
	}
	if env.Error.Details.Expected != "string" {
		t.Errorf("envelope details.expected = %q, want string", env.Error.Details.Expected)
	}
}

func TestParseApplyJSON_PerConceptPath(t *testing.T) {
	payload := `{"concepts":[{"concept_id":"x","definitions":[{"text":"d"}],"languages":{}}]}`
	_, err := ParseApplyJSON([]byte(payload))
	coded := codedFrom(t, err)

	if coded.Code() != "invalid_input" {
		t.Errorf("Code() = %q, want invalid_input", coded.Code())
	}
	msg := coded.Error()
	for _, want := range []string{"concepts", "definitions"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message %q does not contain %q", msg, want)
		}
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Error("errors.Is(err, ErrInvalidInput) = false, want true")
	}
}
