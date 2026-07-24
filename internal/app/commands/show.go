package commands

import (
	"context"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/write"
	urfcli "github.com/urfave/cli/v3"
)

// Show returns the `show` command: it emits a single concept, identified by its
// id, in the canonical WriteResult shape. A missing concept exits 1 not_found.
func Show() *urfcli.Command {
	return &urfcli.Command{
		Name:      "show",
		Usage:     "emit a single concept by id in the canonical shape",
		ArgsUsage: "CONCEPT_ID",
		Arguments: []urfcli.Argument{
			&urfcli.StringArg{Name: "concept-id", UsageText: "CONCEPT_ID"},
		},
		Flags: []urfcli.Flag{
			readFieldsFlag(),
		},
		Before: argBounds(1, 1),
		Action: showAction,
	}
}

func showAction(_ context.Context, cmd *urfcli.Command) error {
	path, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	id := cmd.StringArg("concept-id")
	if err := sanitizeConceptID(id); err != nil {
		return err
	}

	g, _, err := tbx.Load(path)
	if err != nil {
		return wrapLoadError(err)
	}

	idx, err := write.ConceptIndex(g, id)
	if err != nil {
		return lookupNotFound()
	}

	env := output.ShowEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		Concept:       write.ConceptToWriteResult(g.Concepts[idx]),
	}

	return output.EmitProjected(cmd.Root().Writer, env, cmd.String("fields"))
}
