// Package commands defines the terminology CLI subcommands and the shared
// flags, argument validation, and sanitization helpers they rely on.
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
	"github.com/andreswebs/terminology/internal/write"
	urfcli "github.com/urfave/cli/v3"
)

// Apply constructs the "apply" command, which reconciles a declarative JSON or
// TBX payload against the glossary.
func Apply() *urfcli.Command {
	flags := []urfcli.Flag{
		&urfcli.StringFlag{
			Name:      "file",
			Aliases:   []string{"f"},
			Usage:     "path to JSON or TBX payload; '-' for stdin",
			Required:  true,
			TakesFile: true,
		},
		&urfcli.BoolFlag{Name: "prune", Usage: "remove concepts absent from payload"},
	}
	flags = append(flags, writeFlags("preview without modifying file")...)
	return &urfcli.Command{
		Name:   "apply",
		Usage:  "reconcile a declarative payload against the glossary",
		Flags:  flags,
		Action: applyAction,
	}
}

func applyAction(ctx context.Context, cmd *urfcli.Command) error {
	tbxPath, err := tbxPathFromRoot(cmd)
	if err != nil {
		return err
	}

	filePath := cmd.String("file")
	if filePath != "-" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return fmt.Errorf("getting working directory: %w", cwdErr)
		}
		filePath, err = sanitizePath(filePath, cwd)
		if err != nil {
			return err
		}
	}

	prune := cmd.Bool("prune")
	dryRun := cmd.Bool("dry-run")
	wantTxn := cmd.Bool("transaction")
	author := cmd.String("author")

	payload, err := write.LoadApplyFile(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return terr.Newf("io_error", 3, "", "%s", err)
		}
		return err
	}

	for _, c := range payload {
		if c.ID != "" {
			if err := sanitizeConceptID(c.ID); err != nil {
				return err
			}
		}
		for lang := range c.Languages {
			if err := sanitizeLangTag(lang); err != nil {
				return err
			}
		}
	}

	lockPath := tbxPath + ".lock"
	unlock, err := tbx.AcquireLock(lockPath)
	if err != nil {
		return err
	}
	defer unlock()

	g, err := loadTBXForWrite(tbxPath)
	if err != nil {
		return err
	}

	var result *write.ReconcileResult
	if wantTxn {
		result, err = write.ReconcileWithTxn(ctx, g, payload, prune, author)
	} else {
		result, err = write.Reconcile(g, payload, prune)
	}
	if err != nil {
		return err
	}

	if !dryRun {
		if err := tbx.SaveLocked(tbxPath, g); err != nil {
			return err
		}
	}

	env := output.ApplyEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		Applied: output.ApplyResult{
			Added:     result.Added,
			Updated:   result.Updated,
			Removed:   result.Removed,
			Unchanged: result.Unchanged,
		},
	}

	if emitErr := output.EmitJSON(cmd.Root().Writer, env); emitErr != nil {
		return fmt.Errorf("writing output: %w", emitErr)
	}

	return nil
}
