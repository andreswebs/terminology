package commands

import (
	"context"
	"fmt"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	urfcli "github.com/urfave/cli/v3"
)

func Validate() *urfcli.Command {
	return &urfcli.Command{
		Name:  "validate",
		Usage: "validate a TBX file against the supported subset",
		Flags: []urfcli.Flag{
			&urfcli.BoolFlag{Name: "strict", Usage: "promote unknown elements and unresolved IDREFs to errors"},
			readFieldsFlag(),
		},
		Action: validateAction,
	}
}

func validateAction(_ context.Context, cmd *urfcli.Command) error {
	path, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	g, loadWarnings, err := tbx.Load(path)
	if err != nil {
		return wrapLoadError(err)
	}

	strict := cmd.Bool("strict")
	res := g.Validate(strict)

	if len(res.Errors) > 0 {
		return tbx.ErrValidationError.Wrap(
			fmt.Errorf("%d validation error(s)", len(res.Errors)),
		)
	}

	var warnings []output.ValidateWarning
	for _, w := range loadWarnings {
		if !strict && isStrictOnly(w.Code) {
			continue
		}
		warnings = append(warnings, output.ValidateWarning{
			Code: w.Code, Message: w.Message,
			ConceptID: w.ConceptID, Line: w.Line, Col: w.Col,
		})
	}
	for _, w := range res.Warnings {
		if !strict && isStrictOnly(w.Code) {
			continue
		}
		warnings = append(warnings, output.ValidateWarning{
			Code: w.Code, Message: w.Message,
			ConceptID: w.ConceptID, Line: w.Line, Col: w.Col,
		})
	}

	env := output.ValidateEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		Concepts:      res.Concepts,
		Languages:     res.Languages,
		Warnings:      warnings,
	}

	if env.Languages == nil {
		env.Languages = []string{}
	}
	if env.Warnings == nil {
		env.Warnings = []output.ValidateWarning{}
	}

	if err := output.EmitJSON(cmd.Root().Writer, env); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	if len(warnings) > 0 {
		return warningsPresent(len(warnings))
	}

	return nil
}

func isStrictOnly(code string) bool {
	return code == "unknown_element" || code == "legacy_form_normalized"
}

func warningsPresent(n int) error {
	return &warningsError{count: n}
}

type warningsError struct {
	count int
}

func (e *warningsError) Error() string {
	return fmt.Sprintf("%d warning(s)", e.count)
}

func (e *warningsError) ExitCode() int { return 1 }
func (e *warningsError) Code() string  { return "warnings" }
func (e *warningsError) Hint() string  { return "" }
