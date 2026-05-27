package schema

import (
	"testing"

	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

func TestWalkCommands_BasicTree(t *testing.T) {
	root := &urfcli.Command{
		Name:  "root",
		Usage: "test root",
		Commands: []*urfcli.Command{
			{
				Name:  "sub1",
				Usage: "first subcommand",
				Flags: []urfcli.Flag{
					&urfcli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "a name"},
				},
			},
			{
				Name:  "sub2",
				Usage: "second subcommand",
			},
		},
	}

	descs := WalkCommands(root)

	if len(descs) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(descs))
	}

	if descs[0].Name != "sub1" {
		t.Errorf("expected sub1, got %s", descs[0].Name)
	}
	if descs[0].Usage != "first subcommand" {
		t.Errorf("expected 'first subcommand', got %s", descs[0].Usage)
	}
	if len(descs[0].Flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(descs[0].Flags))
	}
	if descs[0].Flags[0].Name != "name" {
		t.Errorf("expected flag 'name', got %s", descs[0].Flags[0].Name)
	}
	if len(descs[0].Flags[0].Aliases) != 1 || descs[0].Flags[0].Aliases[0] != "n" {
		t.Errorf("expected alias [n], got %v", descs[0].Flags[0].Aliases)
	}

	if descs[1].Name != "sub2" {
		t.Errorf("expected sub2, got %s", descs[1].Name)
	}
}

func TestWalkCommands_FlagTypes(t *testing.T) {
	root := &urfcli.Command{
		Name: "root",
		Commands: []*urfcli.Command{
			{
				Name: "cmd",
				Flags: []urfcli.Flag{
					&urfcli.StringFlag{Name: "str", Value: "default"},
					&urfcli.BoolFlag{Name: "verbose"},
					&urfcli.IntFlag{Name: "count", Value: 5},
				},
			},
		},
	}

	descs := WalkCommands(root)

	if len(descs) != 1 {
		t.Fatalf("expected 1 command, got %d", len(descs))
	}
	flags := descs[0].Flags
	if len(flags) != 3 {
		t.Fatalf("expected 3 flags, got %d", len(flags))
	}

	tests := []struct {
		name   string
		typ    string
		defVal string
	}{
		{"str", "string", "default"},
		{"verbose", "bool", ""},
		{"count", "int", "5"},
	}

	for i, tc := range tests {
		if flags[i].Name != tc.name {
			t.Errorf("[%d] name: got %s, want %s", i, flags[i].Name, tc.name)
		}
		if flags[i].Type != tc.typ {
			t.Errorf("[%d] type: got %s, want %s", i, flags[i].Type, tc.typ)
		}
		if flags[i].Default != tc.defVal {
			t.Errorf("[%d] default: got %q, want %q", i, flags[i].Default, tc.defVal)
		}
	}
}

func TestWalkCommands_RequiredFlag(t *testing.T) {
	root := &urfcli.Command{
		Name: "root",
		Commands: []*urfcli.Command{
			{
				Name: "cmd",
				Flags: []urfcli.Flag{
					&urfcli.StringFlag{Name: "file", Required: true},
				},
			},
		},
	}

	descs := WalkCommands(root)
	if !descs[0].Flags[0].Required {
		t.Error("expected flag to be required")
	}
}

func TestWalkCommands_Arguments(t *testing.T) {
	root := &urfcli.Command{
		Name: "root",
		Commands: []*urfcli.Command{
			{
				Name: "cmd",
				Arguments: []urfcli.Argument{
					&urfcli.StringArg{Name: "term"},
				},
			},
			{
				Name: "variadic",
				Arguments: []urfcli.Argument{
					&urfcli.StringArgs{Name: "files", Min: 1, Max: -1},
				},
			},
		},
	}

	descs := WalkCommands(root)

	if len(descs[0].Arguments) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(descs[0].Arguments))
	}
	if descs[0].Arguments[0].Name != "term" {
		t.Errorf("expected arg 'term', got %s", descs[0].Arguments[0].Name)
	}

	if len(descs[1].Arguments) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(descs[1].Arguments))
	}
	if descs[1].Arguments[0].Min != 1 || descs[1].Arguments[0].Max != -1 {
		t.Errorf("expected min=1 max=-1, got min=%d max=%d",
			descs[1].Arguments[0].Min, descs[1].Arguments[0].Max)
	}
}

func TestWalkCommands_NestedSubcommands(t *testing.T) {
	root := &urfcli.Command{
		Name: "root",
		Commands: []*urfcli.Command{
			{
				Name: "parent",
				Commands: []*urfcli.Command{
					{Name: "child1", Usage: "first child"},
					{Name: "child2", Usage: "second child"},
				},
			},
		},
	}

	descs := WalkCommands(root)
	if len(descs) != 1 {
		t.Fatalf("expected 1 top-level command, got %d", len(descs))
	}
	if descs[0].Name != "parent" {
		t.Errorf("expected 'parent', got %s", descs[0].Name)
	}
	if len(descs[0].Commands) != 2 {
		t.Fatalf("expected 2 subcommands, got %d", len(descs[0].Commands))
	}
	if descs[0].Commands[0].Name != "child1" {
		t.Errorf("expected 'child1', got %s", descs[0].Commands[0].Name)
	}
}

func TestReflectEnvelope_SimpleStruct(t *testing.T) {
	type SimpleEnvelope struct {
		SchemaVersion int    `json:"schema_version"`
		OK            bool   `json:"ok"`
		Message       string `json:"message"`
	}

	desc := ReflectEnvelope(SimpleEnvelope{})

	if len(desc.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(desc.Fields))
	}

	tests := []struct {
		name string
		typ  string
	}{
		{"schema_version", "number"},
		{"ok", "boolean"},
		{"message", "string"},
	}

	for i, tc := range tests {
		if desc.Fields[i].Name != tc.name {
			t.Errorf("[%d] name: got %s, want %s", i, desc.Fields[i].Name, tc.name)
		}
		if desc.Fields[i].Type != tc.typ {
			t.Errorf("[%d] type: got %s, want %s", i, desc.Fields[i].Type, tc.typ)
		}
	}
}

func TestReflectEnvelope_NestedStruct(t *testing.T) {
	type Inner struct {
		Term string `json:"term"`
	}
	type Outer struct {
		OK      bool    `json:"ok"`
		Results []Inner `json:"results"`
	}

	desc := ReflectEnvelope(Outer{})

	if len(desc.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(desc.Fields))
	}

	resultsField := desc.Fields[1]
	if resultsField.Name != "results" {
		t.Errorf("expected 'results', got %s", resultsField.Name)
	}
	if resultsField.Type != "array" {
		t.Errorf("expected type 'array', got %s", resultsField.Type)
	}
	if len(resultsField.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(resultsField.Children))
	}
	if resultsField.Children[0].Name != "term" {
		t.Errorf("expected child 'term', got %s", resultsField.Children[0].Name)
	}
}

func TestReflectEnvelope_MapField(t *testing.T) {
	type TermGroup struct {
		Preferred string `json:"preferred"`
	}
	type Env struct {
		Languages map[string]TermGroup `json:"languages"`
	}

	desc := ReflectEnvelope(Env{})

	if len(desc.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(desc.Fields))
	}

	langField := desc.Fields[0]
	if langField.Type != "object" {
		t.Errorf("expected type 'object', got %s", langField.Type)
	}
	if len(langField.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(langField.Children))
	}
	if langField.Children[0].Name != "preferred" {
		t.Errorf("expected child 'preferred', got %s", langField.Children[0].Name)
	}
}

func TestReflectEnvelope_PointerField(t *testing.T) {
	type Inner struct {
		Name string `json:"name"`
	}
	type Env struct {
		Item *Inner `json:"item"`
	}

	desc := ReflectEnvelope(Env{})

	if len(desc.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(desc.Fields))
	}
	if desc.Fields[0].Type != "object" {
		t.Errorf("expected type 'object', got %s", desc.Fields[0].Type)
	}
	if len(desc.Fields[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(desc.Fields[0].Children))
	}
}

func TestReflectEnvelope_OmitsUnexported(t *testing.T) {
	type Env struct {
		OK     bool   `json:"ok"`
		hidden string //nolint:unused
	}

	desc := ReflectEnvelope(Env{})
	if len(desc.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(desc.Fields))
	}
}

func TestReflectEnvelope_JSONTagDash(t *testing.T) {
	type Env struct {
		OK      bool `json:"ok"`
		Ignored int  `json:"-"`
	}

	desc := ReflectEnvelope(Env{})
	if len(desc.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(desc.Fields))
	}
}

func TestEnumerateErrors_ReturnsSentinels(t *testing.T) {
	_ = terr.New("test_sentinel_a", 1, "hint a", "message a")
	_ = terr.New("test_sentinel_b", 65, "", "message b")

	descs := EnumerateErrors()

	codeSet := make(map[string]bool, len(descs))
	for _, d := range descs {
		if d.Code == "" {
			t.Error("empty code")
		}
		if d.ExitCode == 0 {
			t.Errorf("code %s has exit code 0", d.Code)
		}
		if d.Message == "" {
			t.Errorf("code %s has empty message", d.Code)
		}
		codeSet[d.Code] = true
	}

	if !codeSet["test_sentinel_a"] {
		t.Error("expected test_sentinel_a in results")
	}
	if !codeSet["test_sentinel_b"] {
		t.Error("expected test_sentinel_b in results")
	}

	for _, d := range descs {
		if d.Code == "test_sentinel_a" {
			if d.Hint != "hint a" {
				t.Errorf("expected hint 'hint a', got %q", d.Hint)
			}
			if d.ExitCode != 1 {
				t.Errorf("expected exit 1, got %d", d.ExitCode)
			}
		}
		if d.Code == "test_sentinel_b" {
			if d.Hint != "" {
				t.Errorf("expected empty hint, got %q", d.Hint)
			}
		}
	}
}

func TestEnumerateErrors_SortedByCode(t *testing.T) {
	descs := EnumerateErrors()
	for i := 1; i < len(descs); i++ {
		if descs[i].Code < descs[i-1].Code {
			t.Errorf("not sorted: %s before %s", descs[i-1].Code, descs[i].Code)
		}
	}
}

func TestEnumerateErrors_ExcludesNewf(t *testing.T) {
	_ = terr.Newf("test_dynamic_only", 99, "", "should not appear")

	descs := EnumerateErrors()
	for _, d := range descs {
		if d.Code == "test_dynamic_only" {
			t.Error("Newf-created errors should not appear in registry")
		}
	}
}

func TestReflectEnvelope_MapWithPointerValues(t *testing.T) {
	type Inner struct {
		Term string `json:"term"`
	}
	type Env struct {
		Languages map[string]*Inner `json:"languages"`
	}

	desc := ReflectEnvelope(Env{})
	langField := desc.Fields[0]
	if len(langField.Children) != 1 {
		t.Fatalf("expected 1 child for map[string]*Inner, got %d", len(langField.Children))
	}
	if langField.Children[0].Name != "term" {
		t.Errorf("expected child 'term', got %s", langField.Children[0].Name)
	}
}
