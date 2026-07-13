package commands

import (
	"context"
	"strings"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/write"
	urfcli "github.com/urfave/cli/v3"
)

var searchIncludeValues = []string{"definitions", "notes", "contexts", "subject_field"}

// Search returns the `search` command: a discovery finder that matches the
// query as a diacritic- and separator-insensitive normalized substring across
// concept ids, term surfaces, and per-term readings. It is distinct from the
// strict exact `lookup` finder.
func Search() *urfcli.Command {
	return &urfcli.Command{
		Name:      "search",
		Usage:     "discover concepts by normalized substring across terms, readings, and id",
		ArgsUsage: "QUERY",
		Arguments: []urfcli.Argument{
			&urfcli.StringArg{Name: "query", UsageText: "QUERY"},
		},
		Flags: []urfcli.Flag{
			langFlag(false, "restrict the search to this language's fields"),
			searchIncludeFlag(),
			readFieldsFlag(),
		},
		Before: argBounds(1, 1),
		Action: searchAction,
	}
}

func searchIncludeFlag() urfcli.Flag {
	set := make(map[string]bool, len(searchIncludeValues))
	for _, v := range searchIncludeValues {
		set[v] = true
	}
	return &urfcli.StringFlag{
		Name:  "include",
		Usage: "widen the haystack (comma-separated): " + strings.Join(searchIncludeValues, ", "),
		Validator: func(val string) error {
			for _, part := range splitInclude(val) {
				if !set[part] {
					return urfcli.Exit("invalid include value "+part+"; accepted: "+strings.Join(searchIncludeValues, ", "), 2)
				}
			}
			return nil
		},
	}
}

func splitInclude(val string) []string {
	if val == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(val, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func searchAction(_ context.Context, cmd *urfcli.Command) error {
	path, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	query := cmd.StringArg("query")
	if err := sanitizeTerm(query); err != nil {
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

	matches := g.Search(query, tbx.SearchOptions{Lang: lang, Include: splitInclude(cmd.String("include"))})

	results := make([]output.WriteResult, 0, len(matches))
	for _, c := range matches {
		results = append(results, write.ConceptToWriteResult(c))
	}

	env := output.SearchEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		Results:       results,
	}

	if err := output.EmitProjected(cmd.Root().Writer, env, cmd.String("fields")); err != nil {
		return err
	}

	if len(matches) == 0 {
		return lookupNotFound()
	}

	return nil
}
