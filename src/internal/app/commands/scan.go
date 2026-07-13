package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/andreswebs/terminology/internal/markdown"
	"github.com/andreswebs/terminology/internal/match"
	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

// Scan constructs the "scan" command, which finds all glossary term
// occurrences in a markdown file.
func Scan() *urfcli.Command {
	return &urfcli.Command{
		Name:      "scan",
		Usage:     "find all glossary term occurrences in a markdown file",
		ArgsUsage: "FILE",
		Arguments: []urfcli.Argument{
			&urfcli.StringArg{Name: "file", UsageText: "FILE"},
		},
		Flags: []urfcli.Flag{
			langFlag(false, "restrict to a language section"),
			&urfcli.IntFlag{Name: "context", Value: 80, Usage: "context window around each match (chars)"},
			readFieldsFlag(),
		},
		Before: argBounds(1, 1),
		Action: scanAction,
	}
}

func scanAction(_ context.Context, cmd *urfcli.Command) error {
	tbxPath, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	displayPath := cmd.StringArg("file")
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}
	filePath, err := sanitizePath(displayPath, cwd)
	if err != nil {
		return err
	}

	g, _, err := tbx.Load(tbxPath)
	if err != nil {
		return wrapLoadError(err)
	}
	data, err := tbx.ReadFileBounded(filePath, tbx.MaxMarkdownSize)
	if err != nil {
		if coded, ok := err.(interface{ Code() string }); ok && coded.Code() == "input_too_large" {
			return err
		}
		return terr.Newf("io_error", 3, "", "reading %s: %s", displayPath, err)
	}

	lang := markdown.FrontmatterLang(data)
	if lang == "" {
		lang = cmd.String("lang")
	}

	matches, err := match.ScanText(g, data, lang, int(cmd.Int("context")))
	if err != nil {
		return fmt.Errorf("building matcher: %w", err)
	}

	scanMatches := make([]output.ScanMatch, 0, len(matches))
	conceptSet := make(map[string]struct{})
	for _, m := range matches {
		scanMatches = append(scanMatches, output.ScanMatch{
			ConceptID: m.ConceptID,
			Term:      m.Term,
			Lang:      m.Lang,
			Status:    m.Status,
			Line:      m.Line,
			Column:    m.Column,
			Context:   m.Context,
		})
		conceptSet[m.ConceptID] = struct{}{}
	}

	env := output.ScanEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		File:          displayPath,
		Matches:       scanMatches,
		Summary: output.ScanSummary{
			TotalMatches:   len(scanMatches),
			UniqueConcepts: len(conceptSet),
		},
	}

	if err := output.EmitProjected(cmd.Root().Writer, env, cmd.String("fields")); err != nil {
		return err
	}

	return nil
}
