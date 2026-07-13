package commands

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

// Init constructs the "init" command, which creates a minimal valid
// TBX-Linguist skeleton at the target path.
func Init() *urfcli.Command {
	return &urfcli.Command{
		Name:  "init",
		Usage: "create a minimal valid TBX-Linguist skeleton at the target path",
		Flags: []urfcli.Flag{
			&urfcli.StringFlag{
				Name:     "source-lang",
				Usage:    "BCP 47 source language for the new glossary (root xml:lang)",
				Required: true,
			},
			&urfcli.StringFlag{
				Name:  "title",
				Usage: "optional title for <titleStmt>",
			},
			dryRunFlag("render the skeleton without writing it"),
		},
		Action: initAction,
	}
}

func initAction(_ context.Context, cmd *urfcli.Command) error {
	tbxPath, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	sourceLang := cmd.String("source-lang")
	if err := sanitizeLangTag(sourceLang); err != nil {
		return err
	}

	title := cmd.String("title")
	if title != "" {
		if hasControlChars(title) {
			return terr.Newf("invalid_input", 65,
				"titles must not contain control characters",
				"title contains control characters")
		}
	}

	dryRun := cmd.Bool("dry-run")

	if _, statErr := os.Stat(tbxPath); statErr == nil {
		return terr.Newf("io_error", 3,
			"remove the existing file or choose a different --tbx path",
			"refusing to overwrite existing file %s", tbxPath)
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return terr.Newf("io_error", 3, "", "%s", statErr)
	}

	g := &tbx.Glossary{
		Dialect:    tbx.DialectLinguist,
		Style:      tbx.StyleDCT,
		SourceLang: sourceLang,
		Header: tbx.Header{
			Title: title,
		},
	}

	if !dryRun {
		if err := tbx.Save(tbxPath, g); err != nil {
			return terr.Newf("io_error", 3, "", "%s", err)
		}
	}

	env := output.InitEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		SourceLang:    sourceLang,
		Title:         title,
		DryRun:        dryRun,
	}

	if emitErr := output.EmitJSON(cmd.Root().Writer, env); emitErr != nil {
		return fmt.Errorf("writing output: %w", emitErr)
	}

	return nil
}
