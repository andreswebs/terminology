package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/write"
	urfcli "github.com/urfave/cli/v3"
)

// ConceptAdd constructs the "concept add" command, which creates a new concept
// entry.
func ConceptAdd() *urfcli.Command {
	flags := []urfcli.Flag{
		&urfcli.StringFlag{Name: "id", Aliases: []string{"i"}, Usage: "explicit concept id (otherwise derived from canonical preferred term)"},
		&urfcli.StringFlag{Name: "subject-field", Usage: "concept-level subjectField"},
		&urfcli.StringFlag{Name: "canonical-lang", Usage: "language used for id derivation when --id is omitted"},
		langFlag(false, "language tag for the term being added"),
		termFlag(false, "surface form of the term"),
		pickFlag("status", "s", "administrative status", tbx.AdminStatus),
		pickFlag("part-of-speech", "p", "part of speech", tbx.PartOfSpeech),
		pickFlag("register", "r", "register", tbx.Register),
		pickFlag("grammatical-gender", "", "grammatical gender", tbx.GrammaticalGender),
	}
	flags = append(flags, writeFlags("validate and preview without writing")...)
	return &urfcli.Command{
		Name:   "add",
		Usage:  "create a new concept entry",
		Flags:  flags,
		Action: conceptAddAction,
	}
}

func conceptAddAction(ctx context.Context, cmd *urfcli.Command) error {
	tbxPath, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	if id := cmd.String("id"); id != "" {
		if err := sanitizeConceptID(id); err != nil {
			return err
		}
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

	concept, err := parseConceptAddInput(ctx, cmd)
	if err != nil {
		return err
	}

	dryRun := cmd.Bool("dry-run")
	wantTxn := cmd.Bool("transaction")
	author := cmd.String("author")

	if wantTxn {
		txn := write.NewTransaction(ctx, author)
		concept.Transactions = append(concept.Transactions, txn)
	}

	mutator := func(g *tbx.Glossary) (*tbx.Concept, error) {
		for _, c := range g.Concepts {
			if c.ID == concept.ID {
				return nil, write.ErrDuplicateID.Wrap(
					fmt.Errorf("concept %q already exists", concept.ID),
				)
			}
		}
		g.Concepts = append(g.Concepts, *concept)
		return &g.Concepts[len(g.Concepts)-1], nil
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

func parseConceptAddInput(_ context.Context, cmd *urfcli.Command) (*tbx.Concept, error) {
	stdinData, hasStdin, err := readStdinIfAvailable()
	if err != nil {
		return nil, err
	}
	if hasStdin {
		return parseConceptFromStdin(stdinData, cmd)
	}
	return parseConceptFromFlags(cmd)
}

func readStdinIfAvailable() ([]byte, bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, false, nil
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, false, nil
	}
	data, err := tbx.ReadBounded(os.Stdin, tbx.MaxStdinSize)
	if err != nil {
		return nil, false, err
	}
	if len(data) == 0 {
		return nil, false, nil
	}
	return data, true, nil
}

func parseConceptFromStdin(data []byte, cmd *urfcli.Command) (*tbx.Concept, error) {
	data = trimBOM(data)

	if looksLikeXML(data) {
		return parseConceptFromTBXStdin(data, cmd)
	}
	return parseConceptFromJSONStdin(data, cmd)
}

func parseConceptFromTBXStdin(data []byte, cmd *urfcli.Command) (*tbx.Concept, error) {
	concepts, err := write.ParseTBXFragment(data)
	if err != nil {
		return nil, err
	}
	if len(concepts) == 0 {
		return nil, write.ErrInvalidInput.Wrap(fmt.Errorf("no concept entries found in TBX fragment"))
	}
	if len(concepts) > 1 {
		return nil, write.ErrInvalidInput.Wrap(fmt.Errorf("concept add accepts exactly one concept; got %d (use apply for bulk)", len(concepts)))
	}

	c := concepts[0]

	if idFlag := cmd.String("id"); idFlag != "" {
		c.ID = idFlag
	}

	if c.ID == "" {
		derived, err := deriveConceptID(&c, cmd)
		if err != nil {
			return nil, err
		}
		c.ID = derived
	}

	return &c, nil
}

func parseConceptFromJSONStdin(data []byte, cmd *urfcli.Command) (*tbx.Concept, error) {
	wr, err := write.ParseJSONInput(data)
	if err != nil {
		return nil, err
	}

	c := write.ResultToConcept(wr)

	if idFlag := cmd.String("id"); idFlag != "" {
		c.ID = idFlag
	}

	if c.ID == "" {
		derived, err := deriveConceptID(&c, cmd)
		if err != nil {
			return nil, err
		}
		c.ID = derived
	}

	return &c, nil
}

func parseConceptFromFlags(cmd *urfcli.Command) (*tbx.Concept, error) {
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

	id := cmd.String("id")
	if id != "" {
		c.ID = id
	} else {
		derived, err := deriveConceptID(&c, cmd)
		if err != nil {
			return nil, err
		}
		c.ID = derived
	}

	return &c, nil
}

func deriveConceptID(c *tbx.Concept, cmd *urfcli.Command) (string, error) {
	canonLang := cmd.String("canonical-lang")
	if canonLang == "" {
		canonLang = os.Getenv("TERMINOLOGY_CANONICAL_LANG")
	}
	if canonLang == "" {
		canonLang = "en"
	}

	ls, ok := c.Languages[canonLang]
	if !ok {
		for _, ls2 := range c.Languages {
			ls = ls2
			break
		}
	}

	if len(ls.Terms) == 0 {
		return "", write.ErrInvalidID.Wrap(fmt.Errorf("no terms available for ID derivation"))
	}

	preferred := ls.Terms[0].Surface
	for _, t := range ls.Terms {
		if t.AdministrativeStatus == tbx.StatusPreferred {
			preferred = t.Surface
			break
		}
	}

	return write.DeriveID(preferred)
}

func trimBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

func looksLikeXML(data []byte) bool {
	for _, b := range data {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '<':
			return true
		default:
			return false
		}
	}
	return false
}
