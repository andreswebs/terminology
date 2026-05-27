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

		if mergeMode {
			mergeConcept(existing, payload)
		} else {
			replaceConcept(existing, payload)
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

func replaceConcept(existing *tbx.Concept, payload *tbx.Concept) {
	id := existing.ID
	*existing = *payload
	existing.ID = id
}

func mergeConcept(existing *tbx.Concept, payload *tbx.Concept) {
	if payload.SubjectField != "" {
		existing.SubjectField = payload.SubjectField
	}
	if len(payload.Definitions) > 0 {
		existing.Definitions = payload.Definitions
	}
	if len(payload.CrossRefs) > 0 {
		existing.CrossRefs = payload.CrossRefs
	}
	if len(payload.ExternalRefs) > 0 {
		existing.ExternalRefs = payload.ExternalRefs
	}
	if len(payload.Graphics) > 0 {
		existing.Graphics = payload.Graphics
	}
	if len(payload.Sources) > 0 {
		existing.Sources = payload.Sources
	}
	if payload.CustomerSubset != "" {
		existing.CustomerSubset = payload.CustomerSubset
	}
	if payload.ProjectSubset != "" {
		existing.ProjectSubset = payload.ProjectSubset
	}
	if len(payload.Notes) > 0 {
		existing.Notes = payload.Notes
	}

	if existing.Languages == nil {
		existing.Languages = make(map[string]tbx.LangSection)
	}

	for lang, payloadLS := range payload.Languages {
		existingLS, ok := existing.Languages[lang]
		if !ok {
			existing.Languages[lang] = payloadLS
			continue
		}
		mergeLangSection(&existingLS, &payloadLS)
		existing.Languages[lang] = existingLS
	}
}

func mergeLangSection(existing *tbx.LangSection, payload *tbx.LangSection) {
	if len(payload.Definitions) > 0 {
		existing.Definitions = payload.Definitions
	}
	if len(payload.Sources) > 0 {
		existing.Sources = payload.Sources
	}

	for _, pt := range payload.Terms {
		merged := false
		for i := range existing.Terms {
			if existing.Terms[i].Surface == pt.Surface &&
				existing.Terms[i].AdministrativeStatus == pt.AdministrativeStatus {
				mergeTermFields(&existing.Terms[i], &pt)
				merged = true
				break
			}
		}
		if !merged {
			existing.Terms = append(existing.Terms, pt)
		}
	}
}

func mergeTermFields(existing *tbx.Term, payload *tbx.Term) {
	if payload.PartOfSpeech != "" {
		existing.PartOfSpeech = payload.PartOfSpeech
	}
	if payload.GrammaticalGender != "" {
		existing.GrammaticalGender = payload.GrammaticalGender
	}
	if payload.GrammaticalNumber != "" {
		existing.GrammaticalNumber = payload.GrammaticalNumber
	}
	if payload.Register != "" {
		existing.Register = payload.Register
	}
	if payload.TermType != "" {
		existing.TermType = payload.TermType
	}
	if payload.TermLocation != "" {
		existing.TermLocation = payload.TermLocation
	}
	if payload.GeographicalUsage != "" {
		existing.GeographicalUsage = payload.GeographicalUsage
	}
	if payload.TransferComment != "" {
		existing.TransferComment = payload.TransferComment
	}
	if payload.Reading != "" {
		existing.Reading = payload.Reading
	}
	if payload.ReadingNote != "" {
		existing.ReadingNote = payload.ReadingNote
	}
	if payload.CustomerSubset != "" {
		existing.CustomerSubset = payload.CustomerSubset
	}
	if payload.ProjectSubset != "" {
		existing.ProjectSubset = payload.ProjectSubset
	}
	if len(payload.Contexts) > 0 {
		existing.Contexts = payload.Contexts
	}
	if len(payload.Sources) > 0 {
		existing.Sources = payload.Sources
	}
	if len(payload.ExternalRefs) > 0 {
		existing.ExternalRefs = payload.ExternalRefs
	}
	if len(payload.CrossRefs) > 0 {
		existing.CrossRefs = payload.CrossRefs
	}
	if len(payload.Notes) > 0 {
		existing.Notes = payload.Notes
	}
}
