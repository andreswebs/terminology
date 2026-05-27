package commands

import (
	"context"
	"fmt"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/schema"
	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

func Schema() *urfcli.Command {
	return &urfcli.Command{
		Name:  "schema",
		Usage: "emit a reflective description of the CLI surface, envelopes, and error codes",
		Flags: []urfcli.Flag{
			&urfcli.StringFlag{Name: "command", Usage: "restrict output to one command's entry"},
		},
		Action: schemaAction,
	}
}

func schemaAction(_ context.Context, cmd *urfcli.Command) error {
	root := cmd.Root()
	commandFilter := cmd.String("command")

	commands := schema.WalkCommands(root)
	allEnvelopes := output.AllEnvelopes()
	errorCodes := schema.EnumerateErrors()

	if commandFilter != "" {
		return schemaFiltered(root, commandFilter, commands, allEnvelopes, errorCodes)
	}

	envelopeDescs := make(map[string]schema.EnvelopeDesc, len(allEnvelopes))
	for name, zero := range allEnvelopes {
		envelopeDescs[name] = schema.ReflectEnvelope(zero)
	}

	env := schemaFullEnvelope{
		SchemaVersion: output.SchemaVersion,
		Commands:      commands,
		Envelopes:     envelopeDescs,
		ErrorCodes:    errorCodes,
	}

	if emitErr := output.EmitJSON(root.Writer, env); emitErr != nil {
		return fmt.Errorf("writing output: %w", emitErr)
	}
	return nil
}

func schemaFiltered(root *urfcli.Command, name string, commands []schema.CommandDesc, allEnvelopes map[string]any, errorCodes []schema.ErrorCodeDesc) error {
	var found *schema.CommandDesc
	for i := range commands {
		if f := findCommand(&commands[i], name); f != nil {
			found = f
			break
		}
	}
	if found == nil {
		return terr.Newf(
			"unknown_command", 2,
			"run 'terminology schema' for available commands",
			"unknown command: %s", name,
		)
	}

	env := schemaFilteredEnvelope{
		SchemaVersion: output.SchemaVersion,
		Name:          found.Name,
		Usage:         found.Usage,
		Flags:         found.Flags,
		Arguments:     found.Arguments,
		Commands:      found.Commands,
	}

	if zero, ok := allEnvelopes[name]; ok {
		desc := schema.ReflectEnvelope(zero)
		env.Envelope = &desc
	}

	if codes, ok := output.ExitCodesFor(name); ok {
		env.ExitCodes = codes
	}

	env.ErrorCodes = errorCodes

	if emitErr := output.EmitJSON(root.Writer, env); emitErr != nil {
		return fmt.Errorf("writing output: %w", emitErr)
	}
	return nil
}

func findCommand(desc *schema.CommandDesc, name string) *schema.CommandDesc {
	if desc.Name == name {
		return desc
	}
	for i := range desc.Commands {
		if f := findCommand(&desc.Commands[i], name); f != nil {
			return f
		}
	}
	return nil
}

type schemaFullEnvelope struct {
	SchemaVersion int                            `json:"schema_version"`
	Commands      []schema.CommandDesc           `json:"commands"`
	Envelopes     map[string]schema.EnvelopeDesc `json:"envelopes"`
	ErrorCodes    []schema.ErrorCodeDesc         `json:"error_codes"`
}

type schemaFilteredEnvelope struct {
	SchemaVersion int                    `json:"schema_version"`
	Name          string                 `json:"name"`
	Usage         string                 `json:"usage"`
	Flags         []schema.FlagDesc      `json:"flags"`
	Arguments     []schema.ArgDesc       `json:"arguments,omitempty"`
	Commands      []schema.CommandDesc   `json:"commands,omitempty"`
	Envelope      *schema.EnvelopeDesc   `json:"envelope,omitempty"`
	ExitCodes     []int                  `json:"exit_codes,omitempty"`
	ErrorCodes    []schema.ErrorCodeDesc `json:"error_codes"`
}
