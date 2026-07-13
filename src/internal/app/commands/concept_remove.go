package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/write"
	urfcli "github.com/urfave/cli/v3"
)

// ConceptRemove constructs the "concept remove" command, which deletes a
// concept entry.
func ConceptRemove() *urfcli.Command {
	flags := []urfcli.Flag{
		&urfcli.BoolFlag{Name: "force", Usage: "remove even if other concepts cross-reference this ID"},
	}
	flags = append(flags, writeFlags("validate and preview without writing")...)
	return &urfcli.Command{
		Name:      "remove",
		Usage:     "delete a concept entry",
		ArgsUsage: "ID",
		Arguments: []urfcli.Argument{
			&urfcli.StringArg{Name: "id", UsageText: "ID"},
		},
		Flags:  flags,
		Before: argBounds(1, 1),
		Action: conceptRemoveAction,
	}
}

func conceptRemoveAction(ctx context.Context, cmd *urfcli.Command) error {
	tbxPath, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	targetID := cmd.StringArg("id")
	if err := sanitizeConceptID(targetID); err != nil {
		return err
	}
	force := cmd.Bool("force")
	dryRun := cmd.Bool("dry-run")
	wantTxn := cmd.Bool("transaction")
	author := cmd.String("author")

	g, err := loadTBXForWrite(tbxPath)
	if err != nil {
		return err
	}

	idx, err := write.ConceptIndex(g, targetID)
	if err != nil {
		return err
	}

	if !force {
		if refs := write.CrossRefsTo(g, targetID); len(refs) > 0 {
			return write.ErrDanglingCrossref.Wrap(
				fmt.Errorf("concept(s) %s reference %q", strings.Join(refs, ", "), targetID),
			)
		}
	}

	removed := g.Concepts[idx]

	if wantTxn {
		txn := write.NewTransaction(ctx, author)
		removed.Transactions = append(removed.Transactions, txn)
	}

	g.Concepts = append(g.Concepts[:idx], g.Concepts[idx+1:]...)

	if !dryRun {
		if err := tbx.Save(tbxPath, g); err != nil {
			return err
		}
	}

	env := output.WriteEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		Result:        buildWriteResult(removed),
	}

	if emitErr := output.EmitJSON(cmd.Root().Writer, env); emitErr != nil {
		return fmt.Errorf("writing output: %w", emitErr)
	}

	return nil
}
