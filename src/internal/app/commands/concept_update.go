package commands

import (
	"context"
	"fmt"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/terr"
	"github.com/andreswebs/terminology/internal/write"
	urfcli "github.com/urfave/cli/v3"
)

var errMergeReplaceMutex = terr.New(
	"merge_replace_mutex", 2,
	"--merge and --replace are mutually exclusive",
	"pass exactly one of --merge or --replace",
)

var errMergeReplaceRequired = terr.New(
	"merge_replace_required", 2,
	"either --merge or --replace must be supplied",
	"pass exactly one of --merge or --replace",
)

func requireMergeXorReplace(ctx context.Context, cmd *urfcli.Command) (context.Context, error) {
	merge := cmd.Bool("merge")
	replace := cmd.Bool("replace")
	switch {
	case merge && replace:
		return ctx, errMergeReplaceMutex
	case !merge && !replace:
		return ctx, errMergeReplaceRequired
	}
	return ctx, nil
}

func ConceptUpdate() *urfcli.Command {
	flags := []urfcli.Flag{
		&urfcli.BoolFlag{Name: "merge", Usage: "merge supplied fields with existing (preserve unspecified)"},
		&urfcli.BoolFlag{Name: "replace", Usage: "replace entire concept content (except id)"},
		&urfcli.StringFlag{Name: "subject-field", Usage: "concept-level subjectField"},
		langFlag(false, "language tag for the term being added/updated"),
		termFlag(false, "surface form of the term"),
		pickFlag("status", "s", "administrative status", tbx.AdminStatus),
		pickFlag("part-of-speech", "p", "part of speech", tbx.PartOfSpeech),
		pickFlag("register", "r", "register", tbx.Register),
		pickFlag("grammatical-gender", "", "grammatical gender", tbx.GrammaticalGender),
	}
	flags = append(flags, writeFlags("validate and preview without writing")...)
	return &urfcli.Command{
		Name:      "update",
		Usage:     "modify an existing concept",
		ArgsUsage: "ID",
		Arguments: []urfcli.Argument{
			&urfcli.StringArg{Name: "id", UsageText: "ID"},
		},
		Flags:  flags,
		Before: chainBefore(argBounds(1, 1), requireMergeXorReplace),
		Action: conceptUpdateAction,
	}
}

func conceptUpdateAction(ctx context.Context, cmd *urfcli.Command) error {
	tbxPath, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	targetID := cmd.StringArg("id")
	if err := sanitizeConceptID(targetID); err != nil {
		return err
	}
	if lang := cmd.String("lang"); lang != "" {
		if err := sanitizeLangTag(lang); err != nil {
			return err
		}
	}
	if term := cmd.String("term"); term != "" {
		if err := sanitizeTerm(term); err != nil {
			return err
		}
	}
	mergeMode := cmd.Bool("merge")
	dryRun := cmd.Bool("dry-run")
	wantTxn := cmd.Bool("transaction")
	author := cmd.String("author")

	payload, err := parseConceptUpdateInput(cmd)
	if err != nil {
		return err
	}

	mutator := func(g *tbx.Glossary) (*tbx.Concept, error) {
		idx, err := write.ConceptIndex(g, targetID)
		if err != nil {
			return nil, err
		}

		existing := &g.Concepts[idx]

		if mergeMode {
			write.MergeConcept(existing, payload)
		} else {
			write.ReplaceConcept(existing, payload)
		}

		if wantTxn {
			txn := write.NewTransaction(ctx, author)
			existing.Transactions = append(existing.Transactions, txn)
		}

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

func parseConceptUpdateInput(cmd *urfcli.Command) (*tbx.Concept, error) {
	stdinData, hasStdin, err := readStdinIfAvailable()
	if err != nil {
		return nil, err
	}
	if hasStdin {
		return parseConceptFromStdin(stdinData, cmd)
	}
	return parseUpdateFromFlags(cmd)
}

func parseUpdateFromFlags(cmd *urfcli.Command) (*tbx.Concept, error) {
	lang := cmd.String("lang")
	term := cmd.String("term")

	if lang == "" || term == "" {
		return nil, write.ErrInvalidInput.Wrap(
			fmt.Errorf("--lang and --term are required when not piping stdin"),
		)
	}

	t := tbx.Term{
		Surface:              term,
		AdministrativeStatus: tbx.ParseStatus(cmd.String("status")),
		PartOfSpeech:         cmd.String("part-of-speech"),
		Register:             cmd.String("register"),
		GrammaticalGender:    cmd.String("grammatical-gender"),
	}

	c := tbx.Concept{
		SubjectField: cmd.String("subject-field"),
		Languages: map[string]tbx.LangSection{
			lang: {Lang: lang, Terms: []tbx.Term{t}},
		},
	}

	return &c, nil
}
