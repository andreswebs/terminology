package app_test

import (
	"testing"

	"github.com/andreswebs/terminology/internal/app"
	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/schema"
)

func TestWalkCommands_RealTree(t *testing.T) {
	root := app.Root()
	descs := schema.WalkCommands(root)

	names := make(map[string]bool, len(descs))
	for _, d := range descs {
		names[d.Name] = true
	}

	want := []string{"validate", "lookup", "scan", "check", "extract", "apply", "concept", "term", "schema"}
	for _, name := range want {
		if !names[name] {
			t.Errorf("missing command %q", name)
		}
	}
}

func TestWalkCommands_ConceptSubcommands(t *testing.T) {
	root := app.Root()
	descs := schema.WalkCommands(root)

	var concept *schema.CommandDesc
	for i := range descs {
		if descs[i].Name == "concept" {
			concept = &descs[i]
			break
		}
	}
	if concept == nil {
		t.Fatal("concept command not found")
	}

	subNames := make(map[string]bool, len(concept.Commands))
	for _, sub := range concept.Commands {
		subNames[sub.Name] = true
	}

	for _, name := range []string{"add", "update", "remove"} {
		if !subNames[name] {
			t.Errorf("missing concept subcommand %q", name)
		}
	}
}

func TestReflectEnvelope_ValidateEnvelope(t *testing.T) {
	desc := schema.ReflectEnvelope(output.ValidateEnvelope{})

	fieldNames := make(map[string]bool, len(desc.Fields))
	for _, f := range desc.Fields {
		fieldNames[f.Name] = true
	}

	for _, name := range []string{"schema_version", "ok", "concepts", "languages", "warnings"} {
		if !fieldNames[name] {
			t.Errorf("missing field %q in ValidateEnvelope", name)
		}
	}
}

func TestReflectEnvelope_LookupEnvelope(t *testing.T) {
	desc := schema.ReflectEnvelope(output.LookupEnvelope{})

	fieldNames := make(map[string]bool, len(desc.Fields))
	for _, f := range desc.Fields {
		fieldNames[f.Name] = true
	}

	for _, name := range []string{"schema_version", "ok", "results"} {
		if !fieldNames[name] {
			t.Errorf("missing field %q in LookupEnvelope", name)
		}
	}

	var resultsField *schema.FieldDesc
	for i := range desc.Fields {
		if desc.Fields[i].Name == "results" {
			resultsField = &desc.Fields[i]
			break
		}
	}
	if resultsField == nil {
		t.Fatal("results field not found")
	}
	if resultsField.Type != "array" {
		t.Errorf("expected array, got %s", resultsField.Type)
	}
	if len(resultsField.Children) == 0 {
		t.Error("expected children for results field")
	}
}

func TestEnumerateErrors_AllKnownSentinels(t *testing.T) {
	descs := schema.EnumerateErrors()

	codes := make(map[string]bool, len(descs))
	for _, d := range descs {
		codes[d.Code] = true
	}

	want := []string{
		"unsupported_dialect",
		"tbx_locked",
		"no_tbx_path",
		"validation_error",
		"no_subcommand",
	}

	for _, code := range want {
		if !codes[code] {
			t.Errorf("missing error code %q", code)
		}
	}
}

func TestReflectEnvelope_AllRegistered(t *testing.T) {
	all := output.AllEnvelopes()

	for cmd, zero := range all {
		desc := schema.ReflectEnvelope(zero)
		if len(desc.Fields) == 0 {
			t.Errorf("envelope for %q has no fields", cmd)
		}

		hasOK := false
		for _, f := range desc.Fields {
			if f.Name == "ok" {
				hasOK = true
			}
		}
		if !hasOK {
			t.Errorf("envelope for %q missing 'ok' field", cmd)
		}
	}
}
