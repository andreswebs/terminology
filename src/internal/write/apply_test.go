package write

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/andreswebs/terminology/internal/tbx/linguist"
)

func TestParseApplyJSON_ValidPayload(t *testing.T) {
	input := `{
		"concepts": [
			{
				"concept_id": "tzimtzum",
				"subject_field": "kabbalah",
				"languages": {
					"he": {"preferred": {"term": "צמצום"}},
					"es": {"preferred": {"term": "tzimtzum"}}
				}
			},
			{
				"concept_id": "sefirot",
				"languages": {
					"he": {"preferred": {"term": "ספירות"}}
				}
			}
		]
	}`

	concepts, err := ParseApplyJSON([]byte(input))
	if err != nil {
		t.Fatalf("ParseApplyJSON: %v", err)
	}
	if len(concepts) != 2 {
		t.Fatalf("got %d concepts, want 2", len(concepts))
	}
	if concepts[0].ID != "tzimtzum" {
		t.Errorf("concepts[0].ID = %q, want %q", concepts[0].ID, "tzimtzum")
	}
	if concepts[0].SubjectField != "kabbalah" {
		t.Errorf("concepts[0].SubjectField = %q, want %q", concepts[0].SubjectField, "kabbalah")
	}
	if _, ok := concepts[0].Languages["he"]; !ok {
		t.Error("expected he language section on concepts[0]")
	}
	if _, ok := concepts[0].Languages["es"]; !ok {
		t.Error("expected es language section on concepts[0]")
	}
	if concepts[1].ID != "sefirot" {
		t.Errorf("concepts[1].ID = %q, want %q", concepts[1].ID, "sefirot")
	}
}

func TestParseApplyJSON_UnknownField_ReturnsInvalidInput(t *testing.T) {
	input := `{"concepts": [{"concept_id": "x", "bogus": true}]}`

	_, err := ParseApplyJSON([]byte(input))
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}

	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "invalid_input" {
		t.Errorf("got code %q, want %q", coded.Code(), "invalid_input")
	}
}

func TestParseApplyJSON_MalformedJSON(t *testing.T) {
	_, err := ParseApplyJSON([]byte(`{not json at all`))
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "invalid_input" {
		t.Errorf("got code %q, want %q", coded.Code(), "invalid_input")
	}
}

func TestParseApplyJSON_EmptyConcepts(t *testing.T) {
	input := `{"concepts": []}`
	concepts, err := ParseApplyJSON([]byte(input))
	if err != nil {
		t.Fatalf("ParseApplyJSON: %v", err)
	}
	if len(concepts) != 0 {
		t.Errorf("got %d concepts, want 0", len(concepts))
	}
}

func TestDetectPayloadFormat_Extension(t *testing.T) {
	tests := []struct {
		path string
		want PayloadFormat
	}{
		{"payload.json", FormatJSON},
		{"payload.JSON", FormatJSON},
		{"glossary.tbx", FormatTBX},
		{"glossary.TBX", FormatTBX},
		{"glossary.xml", FormatTBX},
		{"glossary.XML", FormatTBX},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := DetectPayloadFormat(tt.path, nil)
			if err != nil {
				t.Fatalf("DetectPayloadFormat(%q): %v", tt.path, err)
			}
			if got != tt.want {
				t.Errorf("DetectPayloadFormat(%q) = %d, want %d", tt.path, got, tt.want)
			}
		})
	}
}

func TestDetectPayloadFormat_ContentSniffing(t *testing.T) {
	tests := []struct {
		name string
		path string
		data string
		want PayloadFormat
	}{
		{"stdin_json", "-", `{"concepts": []}`, FormatJSON},
		{"stdin_xml", "-", `<conceptEntry id="x">`, FormatTBX},
		{"stdin_json_whitespace", "-", "  \n\t{}", FormatJSON},
		{"stdin_xml_whitespace", "-", "  \n\t<foo/>", FormatTBX},
		{"unknown_ext_json", "payload.txt", `{"concepts": []}`, FormatJSON},
		{"unknown_ext_xml", "payload.txt", `<conceptEntry/>`, FormatTBX},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectPayloadFormat(tt.path, []byte(tt.data))
			if err != nil {
				t.Fatalf("DetectPayloadFormat: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDetectPayloadFormat_EmptyData_Error(t *testing.T) {
	_, err := DetectPayloadFormat("-", []byte{})
	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}
	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "invalid_input" {
		t.Errorf("got code %q, want %q", coded.Code(), "invalid_input")
	}
}

func TestDetectPayloadFormat_UnrecognizedContent_Error(t *testing.T) {
	_, err := DetectPayloadFormat("-", []byte("hello world"))
	if err == nil {
		t.Fatal("expected error for unrecognized content, got nil")
	}
	coded, ok := err.(interface{ Code() string })
	if !ok {
		t.Fatalf("expected terr.Coded, got %T", err)
	}
	if coded.Code() != "invalid_input" {
		t.Errorf("got code %q, want %q", coded.Code(), "invalid_input")
	}
}

func TestParseApplyJSON_RichConcept(t *testing.T) {
	input := `{
		"concepts": [
			{
				"concept_id": "tzimtzum",
				"subject_field": "kabbalah",
				"definitions": ["contraction of divine light"],
				"sources": ["zohar"],
				"notes": ["key concept"],
				"languages": {
					"he": {
						"preferred": {"term": "צמצום", "part_of_speech": "noun"},
						"deprecated": [{"term": "tzimtzoum"}]
					},
					"es": {
						"preferred": {"term": "tzimtzum"},
						"admitted": [{"term": "tsimtsum"}]
					}
				}
			}
		]
	}`

	concepts, err := ParseApplyJSON([]byte(input))
	if err != nil {
		t.Fatalf("ParseApplyJSON: %v", err)
	}
	if len(concepts) != 1 {
		t.Fatalf("got %d concepts, want 1", len(concepts))
	}

	c := concepts[0]
	if c.SubjectField != "kabbalah" {
		t.Errorf("SubjectField = %q, want %q", c.SubjectField, "kabbalah")
	}
	if len(c.Definitions) != 1 || c.Definitions[0].Plain != "contraction of divine light" {
		t.Errorf("Definitions = %v, want 1 definition", c.Definitions)
	}
	if len(c.Sources) != 1 || c.Sources[0] != "zohar" {
		t.Errorf("Sources = %v", c.Sources)
	}
	if len(c.Notes) != 1 || c.Notes[0] != "key concept" {
		t.Errorf("Notes = %v", c.Notes)
	}

	heLs, ok := c.Languages["he"]
	if !ok {
		t.Fatal("expected he language section")
	}
	if len(heLs.Terms) != 2 {
		t.Fatalf("he terms = %d, want 2", len(heLs.Terms))
	}

	esLs, ok := c.Languages["es"]
	if !ok {
		t.Fatal("expected es language section")
	}
	if len(esLs.Terms) != 2 {
		t.Fatalf("es terms = %d, want 2", len(esLs.Terms))
	}
}

func TestLoadApplyFile_JSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload.json")
	content := `{
		"concepts": [
			{
				"concept_id": "tzimtzum",
				"languages": {
					"he": {"preferred": {"term": "צמצום"}}
				}
			}
		]
	}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	concepts, err := LoadApplyFile(path)
	if err != nil {
		t.Fatalf("LoadApplyFile: %v", err)
	}
	if len(concepts) != 1 {
		t.Fatalf("got %d concepts, want 1", len(concepts))
	}
	if concepts[0].ID != "tzimtzum" {
		t.Errorf("got ID %q, want %q", concepts[0].ID, "tzimtzum")
	}
}

func TestLoadApplyFile_TBXFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload.tbx")
	content := `<conceptEntry id="c-test"
		xmlns="urn:iso:std:iso:30042:ed-2"
		xmlns:min="http://www.tbxinfo.net/ns/min"
		xmlns:basic="http://www.tbxinfo.net/ns/basic"
		xmlns:ling="http://www.tbxinfo.net/ns/linguist">
		<langSec xml:lang="en">
			<termSec><term>test</term></termSec>
		</langSec>
	</conceptEntry>`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	concepts, err := LoadApplyFile(path)
	if err != nil {
		t.Fatalf("LoadApplyFile: %v", err)
	}
	if len(concepts) != 1 {
		t.Fatalf("got %d concepts, want 1", len(concepts))
	}
	if concepts[0].ID != "c-test" {
		t.Errorf("got ID %q, want %q", concepts[0].ID, "c-test")
	}
}

func TestLoadApplyFile_MissingFile(t *testing.T) {
	_, err := LoadApplyFile("/nonexistent/payload.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadApplyFile_UnknownExtensionSniffsJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload.txt")
	content := `{"concepts": [{"concept_id": "x", "languages": {"en": {"preferred": {"term": "test"}}}}]}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	concepts, err := LoadApplyFile(path)
	if err != nil {
		t.Fatalf("LoadApplyFile: %v", err)
	}
	if len(concepts) != 1 {
		t.Fatalf("got %d concepts, want 1", len(concepts))
	}
}

func TestLoadApplyFile_UnknownExtensionSniffsTBX(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload.dat")
	content := `<conceptEntry id="c-sniff"
		xmlns="urn:iso:std:iso:30042:ed-2"
		xmlns:min="http://www.tbxinfo.net/ns/min"
		xmlns:basic="http://www.tbxinfo.net/ns/basic"
		xmlns:ling="http://www.tbxinfo.net/ns/linguist">
		<langSec xml:lang="en">
			<termSec><term>sniff</term></termSec>
		</langSec>
	</conceptEntry>`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	concepts, err := LoadApplyFile(path)
	if err != nil {
		t.Fatalf("LoadApplyFile: %v", err)
	}
	if len(concepts) != 1 {
		t.Fatalf("got %d concepts, want 1", len(concepts))
	}
	if concepts[0].ID != "c-sniff" {
		t.Errorf("got ID %q, want %q", concepts[0].ID, "c-sniff")
	}
}
