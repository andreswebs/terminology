package output_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/output"
)

func TestEmitProjected_EmptyFieldsEmitsFullEnvelope(t *testing.T) {
	env := testEnvelope{SchemaVersion: 1, OK: true, ConceptID: "c1", SubjectField: "law"}

	var buf bytes.Buffer
	if err := output.EmitProjected(&buf, env, ""); err != nil {
		t.Fatalf("EmitProjected(env, \"\") error = %v, want nil", err)
	}

	got := buf.String()
	for _, want := range []string{`"concept_id":"c1"`, `"subject_field":"law"`, `"ok":true`} {
		if !strings.Contains(got, want) {
			t.Errorf("EmitProjected output = %q, want it to contain %q", got, want)
		}
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("EmitProjected output = %q, want trailing newline", got)
	}
}

func TestEmitProjected_ProjectsRequestedFields(t *testing.T) {
	env := testEnvelope{SchemaVersion: 1, OK: true, ConceptID: "c1", SubjectField: "law"}

	var buf bytes.Buffer
	if err := output.EmitProjected(&buf, env, "concept_id"); err != nil {
		t.Fatalf("EmitProjected(env, \"concept_id\") error = %v, want nil", err)
	}

	got := buf.String()
	if !strings.Contains(got, `"concept_id":"c1"`) {
		t.Errorf("EmitProjected output = %q, want it to contain concept_id", got)
	}
	if strings.Contains(got, "subject_field") {
		t.Errorf("EmitProjected output = %q, want subject_field projected out", got)
	}
	// schema_version and ok are always retained by ProjectFields.
	if !strings.Contains(got, `"schema_version":1`) || !strings.Contains(got, `"ok":true`) {
		t.Errorf("EmitProjected output = %q, want schema_version and ok retained", got)
	}
}

func TestEmitProjected_InvalidFieldReturnsCodedError(t *testing.T) {
	env := testEnvelope{SchemaVersion: 1, OK: true}

	var buf bytes.Buffer
	err := output.EmitProjected(&buf, env, "bogus_field")
	if err == nil {
		t.Fatal("EmitProjected(env, \"bogus_field\") error = nil, want ErrInvalidField")
	}
	if !errors.Is(err, output.ErrInvalidField) {
		t.Errorf("EmitProjected error = %v, want errors.Is ErrInvalidField", err)
	}
	if buf.Len() != 0 {
		t.Errorf("EmitProjected wrote %q on validation failure, want no output", buf.String())
	}
}
