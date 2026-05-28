package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/andreswebs/terminology/internal/check"
	"github.com/andreswebs/terminology/internal/markdown"
	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

var ErrLanguageRequired = terr.New(
	"language_required", 2,
	"pass --source-lang/--target-lang or add 'lang: LANG' to frontmatter",
	"language not specified",
)

func Check() *urfcli.Command {
	return &urfcli.Command{
		Name:      "check",
		Usage:     "verify a translated target file against the source given the glossary",
		ArgsUsage: "SRC TGT",
		Arguments: []urfcli.Argument{
			&urfcli.StringArgs{Name: "files", Min: 2, Max: 2},
		},
		Flags: []urfcli.Flag{
			&urfcli.StringFlag{Name: "source-lang", Aliases: []string{"S"}, Usage: "source language"},
			&urfcli.StringFlag{Name: "target-lang", Usage: "target language"},
			&urfcli.BoolFlag{Name: "strict", Usage: "admitted variants raise violations"},
			&urfcli.IntFlag{Name: "context", Value: 80, Usage: "context window around each violation (chars)"},
			readFieldsFlag(),
		},
		Before: argBounds(2, 2),
		Action: checkAction,
	}
}

func checkAction(_ context.Context, cmd *urfcli.Command) error {
	tbxPath, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	files := cmd.StringArgs("files")
	srcDisplay, tgtDisplay := files[0], files[1]
	srcPath, err := sanitizePath(srcDisplay, cwd)
	if err != nil {
		return err
	}
	tgtPath, err := sanitizePath(tgtDisplay, cwd)
	if err != nil {
		return err
	}

	if lang := cmd.String("source-lang"); lang != "" {
		if err := sanitizeLangTag(lang); err != nil {
			return err
		}
	}
	if lang := cmd.String("target-lang"); lang != "" {
		if err := sanitizeLangTag(lang); err != nil {
			return err
		}
	}

	g, _, err := tbx.Load(tbxPath)
	if err != nil {
		return wrapLoadError(err)
	}

	srcData, err := tbx.ReadFileBounded(srcPath, tbx.MaxMarkdownSize)
	if err != nil {
		if coded, ok := err.(interface{ Code() string }); ok && coded.Code() == "input_too_large" {
			return err
		}
		return terr.Newf("io_error", 3, "", "reading %s: %s", srcDisplay, err)
	}

	tgtData, err := tbx.ReadFileBounded(tgtPath, tbx.MaxMarkdownSize)
	if err != nil {
		if coded, ok := err.(interface{ Code() string }); ok && coded.Code() == "input_too_large" {
			return err
		}
		return terr.Newf("io_error", 3, "", "reading %s: %s", tgtDisplay, err)
	}

	srcLang := markdown.FrontmatterLang(srcData)
	if srcLang == "" {
		srcLang = cmd.String("source-lang")
	}
	if srcLang == "" {
		return ErrLanguageRequired.Wrap(
			fmt.Errorf("language not specified for %s", srcDisplay),
		)
	}

	tgtLang := markdown.FrontmatterLang(tgtData)
	if tgtLang == "" {
		tgtLang = cmd.String("target-lang")
	}
	if tgtLang == "" {
		return ErrLanguageRequired.Wrap(
			fmt.Errorf("language not specified for %s", tgtDisplay),
		)
	}

	strict := cmd.Bool("strict")
	contextSize := int(cmd.Int("context"))

	result, err := check.Check(g, srcData, tgtData, srcLang, tgtLang, contextSize, strict)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}

	env := output.CheckEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            len(result.Violations) == 0,
		Source:        srcDisplay,
		Target:        tgtDisplay,
		Violations:    result.Violations,
		Warnings:      result.Warnings,
		Summary: output.CheckSummary{
			Violations:      len(result.Violations),
			Warnings:        len(result.Warnings),
			ConceptsChecked: result.ConceptsChecked,
		},
	}

	if err := output.EmitProjected(cmd.Root().Writer, env, cmd.String("fields")); err != nil {
		return err
	}

	if len(result.Violations) > 0 {
		return violationsPresent(len(result.Violations))
	}

	return nil
}

func violationsPresent(n int) error {
	return &violationsError{count: n}
}

type violationsError struct {
	count int
}

func (e *violationsError) Error() string {
	return fmt.Sprintf("%d violation(s)", e.count)
}

func (e *violationsError) ExitCode() int { return 1 }
func (e *violationsError) Code() string  { return "violations" }
func (e *violationsError) Hint() string  { return "" }
