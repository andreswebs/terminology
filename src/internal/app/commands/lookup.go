package commands

import (
	"context"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/write"
	urfcli "github.com/urfave/cli/v3"
)

// Lookup constructs the "lookup" command, which looks up a term across all
// languages in the TBX file.
func Lookup() *urfcli.Command {
	return &urfcli.Command{
		Name:      "lookup",
		Usage:     "look up a term across all languages in the TBX file",
		ArgsUsage: "TERM",
		Arguments: []urfcli.Argument{
			&urfcli.StringArg{Name: "term", UsageText: "TERM"},
		},
		Flags: []urfcli.Flag{
			langFlag(false, "restrict to a language section"),
			readFieldsFlag(),
		},
		Before: argBounds(1, 1),
		Action: lookupAction,
	}
}

func lookupAction(_ context.Context, cmd *urfcli.Command) error {
	path, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	term := cmd.StringArg("term")
	if err := sanitizeTerm(term); err != nil {
		return err
	}

	lang := cmd.String("lang")
	if lang != "" {
		if err := sanitizeLangTag(lang); err != nil {
			return err
		}
	}

	g, _, err := tbx.Load(path)
	if err != nil {
		return wrapLoadError(err)
	}

	matches := g.Lookup(term, lang)

	env := output.LookupEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		Results:       buildLookupResults(matches),
	}

	if err := output.EmitProjected(cmd.Root().Writer, env, cmd.String("fields")); err != nil {
		return err
	}

	if len(matches) == 0 {
		return lookupNotFound()
	}

	return nil
}

func buildLookupResults(matches []tbx.LookupMatch) []output.WriteResult {
	results := make([]output.WriteResult, 0, len(matches))
	for _, m := range matches {
		results = append(results, write.ConceptToWriteResult(m.Concept))
	}
	return results
}

func lookupNotFound() error {
	return &lookupNotFoundError{}
}

type lookupNotFoundError struct{}

func (e *lookupNotFoundError) Error() string { return "no results found" }
func (e *lookupNotFoundError) ExitCode() int { return 1 }
func (e *lookupNotFoundError) Code() string  { return "not_found" }
func (e *lookupNotFoundError) Hint() string  { return "" }
