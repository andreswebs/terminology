package commands

import (
	"context"
	"fmt"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/write"
	urfcli "github.com/urfave/cli/v3"
)

func TermDeprecate() *urfcli.Command {
	flags := []urfcli.Flag{
		langFlag(true, "language tag"),
		termFlag(true, "surface form"),
	}
	flags = append(flags, writeFlags("validate and preview without writing")...)
	return &urfcli.Command{
		Name:      "deprecate",
		Usage:     "set an existing term's administrativeStatus to deprecatedTerm-admn-sts",
		ArgsUsage: "ID",
		Arguments: []urfcli.Argument{
			&urfcli.StringArg{Name: "id", UsageText: "ID"},
		},
		Flags:  flags,
		Before: argBounds(1, 1),
		Action: termDeprecateAction,
	}
}

func termDeprecateAction(ctx context.Context, cmd *urfcli.Command) error {
	tbxPath, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	targetID := cmd.StringArg("id")
	if err := sanitizeConceptID(targetID); err != nil {
		return err
	}

	lang := cmd.String("lang")
	if err := sanitizeLangTag(lang); err != nil {
		return err
	}

	termSurface := cmd.String("term")
	if err := sanitizeTerm(termSurface); err != nil {
		return err
	}
	dryRun := cmd.Bool("dry-run")
	wantTxn := cmd.Bool("transaction")
	author := cmd.String("author")

	mutator := func(g *tbx.Glossary) (*tbx.Concept, error) {
		idx := -1
		for i := range g.Concepts {
			if g.Concepts[i].ID == targetID {
				idx = i
				break
			}
		}
		if idx == -1 {
			return nil, write.ErrNotFound.Wrap(
				fmt.Errorf("concept %q not found", targetID),
			)
		}

		existing := &g.Concepts[idx]

		ls, ok := existing.Languages[lang]
		if !ok {
			return nil, write.ErrNotFound.Wrap(
				fmt.Errorf("language %q not found in concept %q", lang, targetID),
			)
		}

		termIdx := -1
		for i := range ls.Terms {
			if ls.Terms[i].Surface == termSurface {
				termIdx = i
				break
			}
		}
		if termIdx == -1 {
			return nil, write.ErrNotFound.Wrap(
				fmt.Errorf("term %q not found in language %q of concept %q", termSurface, lang, targetID),
			)
		}

		ls.Terms[termIdx].AdministrativeStatus = tbx.StatusDeprecated

		if wantTxn {
			txn := write.NewTransaction(ctx, author)
			ls.Terms[termIdx].Transactions = append(ls.Terms[termIdx].Transactions, txn)
		}

		existing.Languages[lang] = ls

		return existing, nil
	}

	affected, err := write.Execute(tbxPath, mutator, dryRun)
	if err != nil {
		return err
	}

	env := output.WriteEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		Result:        buildWriteResult(*affected),
	}

	if emitErr := output.EmitJSON(cmd.Root().Writer, env); emitErr != nil {
		return fmt.Errorf("writing output: %w", emitErr)
	}

	return nil
}
