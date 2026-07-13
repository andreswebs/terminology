package commands

import (
	"context"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	urfcli "github.com/urfave/cli/v3"
)

// List returns the `list` command: a projected enumeration of the glossary,
// emitting each concept as its id, subject field, and per-language preferred
// term only, sorted by id. It is equivalent to
// `export --fields concept_id,subject_field,languages.*.preferred.term`.
func List() *urfcli.Command {
	return &urfcli.Command{
		Name:  "list",
		Usage: "enumerate concepts as id + preferred term per language",
		Flags: []urfcli.Flag{
			langFlag(false, "restrict emitted language sections to this tag"),
			readFieldsFlag(),
		},
		Action: listAction,
	}
}

func listAction(_ context.Context, cmd *urfcli.Command) error {
	path, err := tbxPathFromRoot(cmd)
	if err != nil {
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

	results := conceptsToResults(g)
	restrictLang(results, lang)
	results = reduceToPreferred(results)

	env := output.ListEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		Concepts:      results,
	}

	return output.EmitProjected(cmd.Root().Writer, env, cmd.String("fields"))
}
