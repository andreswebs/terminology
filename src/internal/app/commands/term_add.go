package commands

import (
	"context"
	"fmt"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/write"
	urfcli "github.com/urfave/cli/v3"
)

// TermAdd constructs the "term add" command, which adds a term to an existing
// concept's langSec.
func TermAdd() *urfcli.Command {
	flags := []urfcli.Flag{
		langFlag(true, "language tag"),
		termFlag(true, "surface form"),
		pickFlag("status", "s", "administrative status", tbx.AdminStatus),
		pickFlag("part-of-speech", "p", "part of speech", tbx.PartOfSpeech),
		pickFlag("register", "r", "register", tbx.Register),
		pickFlag("grammatical-gender", "", "grammatical gender", tbx.GrammaticalGender),
	}
	flags = append(flags, writeFlags("validate and preview without writing")...)
	return &urfcli.Command{
		Name:      "add",
		Usage:     "add a term to an existing concept's langSec",
		ArgsUsage: "ID",
		Arguments: []urfcli.Argument{
			&urfcli.StringArg{Name: "id", UsageText: "ID"},
		},
		Flags:  flags,
		Before: argBounds(1, 1),
		Action: termAddAction,
	}
}

func termAddAction(ctx context.Context, cmd *urfcli.Command) error {
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

	t := tbx.Term{
		Surface:              termSurface,
		AdministrativeStatus: tbx.ParseStatus(cmd.String("status")),
		PartOfSpeech:         cmd.String("part-of-speech"),
		Register:             cmd.String("register"),
		GrammaticalGender:    cmd.String("grammatical-gender"),
	}

	mutator := func(g *tbx.Glossary) (*tbx.Concept, error) {
		idx, err := write.ConceptIndex(g, targetID)
		if err != nil {
			return nil, err
		}

		existing := &g.Concepts[idx]

		if existing.Languages == nil {
			existing.Languages = make(map[string]tbx.LangSection)
		}

		ls, ok := existing.Languages[lang]
		if !ok {
			ls = tbx.LangSection{Lang: lang}
		}

		if wantTxn {
			txn := write.NewTransaction(ctx, author)
			t.Transactions = append(t.Transactions, txn)
		}

		ls.Terms = append(ls.Terms, t)
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
