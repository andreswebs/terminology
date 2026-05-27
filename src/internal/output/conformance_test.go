package output_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/terr"
)

// AssertEnvelopeShape unmarshals raw into a generic map and asserts:
//   - top-level "schema_version" is an int (== SchemaVersion)
//   - top-level "ok" is a bool
//   - if ok==false, top-level "error" exists with string "code" and "message"
//   - if ok==true, no top-level "error" key
func AssertEnvelopeShape(t *testing.T, raw []byte) {
	t.Helper()

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Errorf("envelope is not valid JSON: %v", err)
		return
	}

	sv, ok := m["schema_version"]
	if !ok {
		t.Error("envelope missing schema_version")
		return
	}
	if int(sv.(float64)) != output.SchemaVersion {
		t.Errorf("schema_version = %v, want %d", sv, output.SchemaVersion)
	}

	okVal, exists := m["ok"]
	if !exists {
		t.Error("envelope missing ok field")
		return
	}
	okBool, isBool := okVal.(bool)
	if !isBool {
		t.Errorf("ok field is %T, want bool", okVal)
		return
	}

	if okBool {
		if _, has := m["error"]; has {
			t.Error("ok=true envelope must not have error field")
		}
	} else {
		errObj, has := m["error"]
		if !has {
			t.Error("ok=false envelope must have error field")
			return
		}
		errMap, isMap := errObj.(map[string]any)
		if !isMap {
			t.Errorf("error field is %T, want object", errObj)
			return
		}
		if _, has := errMap["code"]; !has {
			t.Error("error object missing code")
		}
		if _, has := errMap["message"]; !has {
			t.Error("error object missing message")
		}
	}
}

func TestAssertEnvelopeShape_ValidErrorEnvelope(t *testing.T) {
	var buf bytes.Buffer
	output.EmitError(&buf, "json", terr.Newf("test_error", 65, "try again", "something went wrong"))
	AssertEnvelopeShape(t, buf.Bytes())
}

func TestAssertEnvelopeShape_ValidErrorEnvelope_NoHint(t *testing.T) {
	var buf bytes.Buffer
	output.EmitError(&buf, "json", terr.New("x", 2, "", "msg"))
	AssertEnvelopeShape(t, buf.Bytes())
}

func TestAssertEnvelopeShape_RejectsMalformedJSON(t *testing.T) {
	tt := &testing.T{}
	AssertEnvelopeShape(tt, []byte(`not json`))
	if !tt.Failed() {
		t.Error("expected AssertEnvelopeShape to fail on malformed JSON")
	}
}

func TestAssertEnvelopeShape_RejectsMissingSchemaVersion(t *testing.T) {
	tt := &testing.T{}
	AssertEnvelopeShape(tt, []byte(`{"ok":false,"error":{"code":"x","message":"m"}}`))
	if !tt.Failed() {
		t.Error("expected AssertEnvelopeShape to fail on missing schema_version")
	}
}

func TestAssertEnvelopeShape_RejectsMissingOk(t *testing.T) {
	tt := &testing.T{}
	AssertEnvelopeShape(tt, []byte(`{"schema_version":1,"error":{"code":"x","message":"m"}}`))
	if !tt.Failed() {
		t.Error("expected AssertEnvelopeShape to fail on missing ok field")
	}
}
