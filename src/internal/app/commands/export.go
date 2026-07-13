package commands

import (
	"context"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	urfcli "github.com/urfave/cli/v3"
)

// Export returns the `export` command: it emits every concept in the glossary
// in the canonical WriteResult shape, sorted by concept id. The output is
// apply-consumable, enabling a read-modify-write round-trip.
func Export() *urfcli.Command {
	return &urfcli.Command{
		Name:  "export",
		Usage: "emit every concept in the canonical shape (apply-consumable)",
		Flags: []urfcli.Flag{
			langFlag(false, "restrict emitted language sections to this tag"),
			readFieldsFlag(),
		},
		Action: exportAction,
	}
}

func exportAction(_ context.Context, cmd *urfcli.Command) error {
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

	env := output.ExportEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		Concepts:      results,
	}

	return output.EmitProjected(cmd.Root().Writer, env, cmd.String("fields"))
}
