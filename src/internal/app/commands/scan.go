package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/andreswebs/terminology/internal/markdown"
	"github.com/andreswebs/terminology/internal/match"
	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

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
	policy := match.PolicyFor(lang)

	matcher, err := match.New(g, lang, policy)
	if err != nil {
		return fmt.Errorf("building matcher: %w", err)
	}

	var spans []markdown.Span
	for s := range markdown.Spans(data) {
		spans = append(spans, s)
	}

	contextSize := cmd.Int("context")
	matches := matcher.Scan(data, spans, int(contextSize))

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

	fieldsStr := cmd.String("fields")
	if fieldsStr != "" {
		fields, fErr := output.ValidateFields(fieldsStr, env)
		if fErr != nil {
			return fErr
		}

		jsonData, mErr := json.Marshal(env)
		if mErr != nil {
			return fmt.Errorf("marshaling output: %w", mErr)
		}

		projected, pErr := output.ProjectFields(jsonData, fields)
		if pErr != nil {
			return fmt.Errorf("projecting fields: %w", pErr)
		}

		if _, wErr := cmd.Root().Writer.Write(projected); wErr != nil {
			return fmt.Errorf("writing output: %w", wErr)
		}
		if _, wErr := cmd.Root().Writer.Write([]byte("\n")); wErr != nil {
			return fmt.Errorf("writing output: %w", wErr)
		}
	} else {
		if emitErr := output.EmitJSON(cmd.Root().Writer, env); emitErr != nil {
			return fmt.Errorf("writing output: %w", emitErr)
		}
	}

	return nil
}
