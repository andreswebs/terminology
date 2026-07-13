// Package schema introspects the CLI's command tree, JSON envelopes, and error
// codes to produce machine-readable descriptions consumed by the `schema`
// command.
package schema

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

// CommandDesc describes a CLI command, including its flags, positional
// arguments, and any nested subcommands.
type CommandDesc struct {
	Name      string        `json:"name"`
	Usage     string        `json:"usage"`
	Flags     []FlagDesc    `json:"flags"`
	Arguments []ArgDesc     `json:"arguments,omitempty"`
	Commands  []CommandDesc `json:"commands,omitempty"`
}

// FlagDesc describes a single command flag, including its type, default value,
// aliases, and whether it is required.
type FlagDesc struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Default  string   `json:"default,omitempty"`
	Aliases  []string `json:"aliases,omitempty"`
	Required bool     `json:"required,omitempty"`
	Usage    string   `json:"usage"`
}

// ArgDesc describes a positional argument and how many values it accepts.
type ArgDesc struct {
	Name string `json:"name"`
	Min  int    `json:"min"`
	Max  int    `json:"max"`
}

// EnvelopeDesc describes the fields of a JSON output envelope.
type EnvelopeDesc struct {
	Fields []FieldDesc `json:"fields"`
}

// FieldDesc describes a single envelope field, including its JSON type and any
// nested child fields for structured values.
type FieldDesc struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Children []FieldDesc `json:"children,omitempty"`
}

// ErrorCodeDesc describes an error the CLI can emit, including its code, process
// exit code, message, and optional remediation hint.
type ErrorCodeDesc struct {
	Code     string `json:"code"`
	ExitCode int    `json:"exit_code"`
	Message  string `json:"message"`
	Hint     string `json:"hint,omitempty"`
}

// WalkCommands describes every subcommand of root, recursing into nested
// subcommands.
func WalkCommands(root *urfcli.Command) []CommandDesc {
	var descs []CommandDesc
	for _, cmd := range root.Commands {
		descs = append(descs, walkCommand(cmd))
	}
	return descs
}

func walkCommand(cmd *urfcli.Command) CommandDesc {
	desc := CommandDesc{
		Name:  cmd.Name,
		Usage: cmd.Usage,
	}

	for _, f := range cmd.Flags {
		desc.Flags = append(desc.Flags, describeFlag(f))
	}

	for _, a := range cmd.Arguments {
		desc.Arguments = append(desc.Arguments, describeArgument(a))
	}

	for _, sub := range cmd.Commands {
		desc.Commands = append(desc.Commands, walkCommand(sub))
	}

	return desc
}

func describeFlag(f urfcli.Flag) FlagDesc {
	fd := FlagDesc{
		Name:  f.Names()[0],
		Usage: flagUsage(f),
	}

	if len(f.Names()) > 1 {
		fd.Aliases = f.Names()[1:]
	}

	switch tf := f.(type) {
	case *urfcli.StringFlag:
		fd.Type = "string"
		fd.Required = tf.Required
		if tf.Value != "" {
			fd.Default = tf.Value
		}
	case *urfcli.BoolFlag:
		fd.Type = "bool"
		fd.Required = tf.Required
	case *urfcli.IntFlag:
		fd.Type = "int"
		fd.Required = tf.Required
		if tf.Value != 0 {
			fd.Default = fmt.Sprintf("%d", tf.Value)
		}
	case *urfcli.FloatFlag:
		fd.Type = "float"
		fd.Required = tf.Required
	default:
		fd.Type = "unknown"
	}

	return fd
}

func flagUsage(f urfcli.Flag) string {
	rv := reflect.ValueOf(f)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	uf := rv.FieldByName("Usage")
	if uf.IsValid() && uf.Kind() == reflect.String {
		return uf.String()
	}
	return ""
}

func describeArgument(a urfcli.Argument) ArgDesc {
	switch ta := a.(type) {
	case *urfcli.StringArg:
		return ArgDesc{Name: ta.Name, Min: 1, Max: 1}
	case *urfcli.StringArgs:
		return ArgDesc{Name: ta.Name, Min: ta.Min, Max: ta.Max}
	default:
		return ArgDesc{Name: "unknown"}
	}
}

// ReflectEnvelope describes the JSON fields of an envelope type, using the
// struct tags of the given zero value.
func ReflectEnvelope(zero any) EnvelopeDesc {
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return EnvelopeDesc{Fields: reflectFields(t)}
}

func reflectFields(t reflect.Type) []FieldDesc {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	var fields []FieldDesc
	for sf := range t.Fields() {
		sf := sf
		if !sf.IsExported() {
			continue
		}

		tag := sf.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name := jsonFieldName(tag, sf.Name)

		fd := FieldDesc{Name: name, Type: jsonType(sf.Type)}

		elem := sf.Type
		if elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Slice {
			elem = elem.Elem()
			if elem.Kind() == reflect.Pointer {
				elem = elem.Elem()
			}
		}
		if elem.Kind() == reflect.Map {
			valType := elem.Elem()
			if valType.Kind() == reflect.Pointer {
				valType = valType.Elem()
			}
			if valType.Kind() == reflect.Struct {
				fd.Children = reflectFields(valType)
			}
		} else if elem.Kind() == reflect.Struct && elem != reflect.TypeFor[string]() {
			fd.Children = reflectFields(elem)
		}

		fields = append(fields, fd)
	}
	return fields
}

func jsonFieldName(tag, fallback string) string {
	if tag == "" {
		return fallback
	}
	for i := range len(tag) {
		if tag[i] == ',' {
			return tag[:i]
		}
	}
	return tag
}

func jsonType(t reflect.Type) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice:
		return "array"
	case reflect.Map:
		return "object"
	case reflect.Struct:
		return "object"
	default:
		return "unknown"
	}
}

// EnumerateErrors returns descriptions of all known CLI errors, sorted by code.
func EnumerateErrors() []ErrorCodeDesc {
	all := terr.All()
	descs := make([]ErrorCodeDesc, 0, len(all))
	for _, e := range all {
		descs = append(descs, ErrorCodeDesc{
			Code:     e.Code(),
			ExitCode: e.ExitCode(),
			Message:  e.Error(),
			Hint:     e.Hint(),
		})
	}
	sort.Slice(descs, func(i, j int) bool {
		return descs[i].Code < descs[j].Code
	})
	return descs
}
