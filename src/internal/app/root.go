package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/andreswebs/terminology/internal/app/commands"
	"github.com/andreswebs/terminology/internal/logctx"
	"github.com/andreswebs/terminology/internal/tbx"
	_ "github.com/andreswebs/terminology/internal/tbx/linguist"
	"github.com/andreswebs/terminology/internal/terr"
	"github.com/andreswebs/terminology/internal/version"
	urfcli "github.com/urfave/cli/v3"
	"golang.org/x/term"
)

func Root() *urfcli.Command {
	return &urfcli.Command{
		Name:    "terminology",
		Usage:   "agent-driven, terminology-focused academic translation",
		Version: version.Current(),
		Flags:   globalFlags(),
		Before:  beforeHook,
		Action:  rootAction,
		Commands: []*urfcli.Command{
			commands.Init(),
			commands.Validate(),
			commands.Lookup(),
			commands.Search(),
			commands.Export(),
			commands.Show(),
			commands.List(),
			commands.Scan(),
			commands.Check(),
			commands.Extract(),
			commands.Apply(),
			commands.Concept(),
			commands.Term(),
			commands.Schema(),
		},
		ExitErrHandler: func(_ context.Context, _ *urfcli.Command, _ error) {},
	}
}

func globalFlags() []urfcli.Flag {
	return []urfcli.Flag{
		&urfcli.StringFlag{
			Name:      "tbx",
			Aliases:   []string{"T"},
			Usage:     "path to TBX glossary file",
			Sources:   urfcli.EnvVars("TERMINOLOGY_TBX"),
			TakesFile: true,
		},
		&urfcli.StringFlag{
			Name:      "format",
			Usage:     "output format: json or text",
			Value:     "json",
			Validator: enumValidator(tbx.Format()),
		},
		&urfcli.BoolFlag{Name: "verbose", Usage: "INFO-level diagnostics"},
		&urfcli.BoolFlag{Name: "debug", Usage: "DEBUG-level diagnostics"},
		&urfcli.BoolFlag{Name: "quiet", Usage: "ERROR-only diagnostics"},
	}
}

func enumValidator(allowed []string) func(string) error {
	set := make(map[string]bool, len(allowed))
	for _, v := range allowed {
		set[v] = true
	}
	return func(val string) error {
		if !set[val] {
			return urfcli.Exit("invalid value "+val+"; accepted: json, text", 2)
		}
		return nil
	}
}

func beforeHook(ctx context.Context, cmd *urfcli.Command) (context.Context, error) {
	n := boolCount(cmd.Bool("verbose"), cmd.Bool("debug"), cmd.Bool("quiet"))
	if n > 1 {
		return ctx, ErrConflictingVerbosity
	}

	level := slog.LevelWarn
	switch {
	case cmd.Bool("debug"):
		level = slog.LevelDebug
	case cmd.Bool("verbose"):
		level = slog.LevelInfo
	case cmd.Bool("quiet"):
		level = slog.LevelError
	}

	var h slog.Handler
	opts := &slog.HandlerOptions{Level: level}
	if term.IsTerminal(int(os.Stderr.Fd())) {
		h = slog.NewTextHandler(os.Stderr, opts)
	} else {
		h = slog.NewJSONHandler(os.Stderr, opts)
	}

	logger := slog.New(h).With(
		"command", cmd.FullName(),
		"run_id", logctx.NewRunID(),
		"version", version.Current(),
	)

	return logctx.With(ctx, logger), nil
}

var errNoSubcommand = terr.New(
	"no_subcommand", 2,
	"run 'terminology --help' for available commands",
	"no subcommand specified",
)

func rootAction(_ context.Context, cmd *urfcli.Command) error {
	if args := cmd.Args(); args.Len() > 0 {
		return terr.Newf(
			"unknown_subcommand", 2,
			"run 'terminology --help' for available commands",
			"unknown subcommand: %s", args.First(),
		)
	}
	return errNoSubcommand
}

func boolCount(vals ...bool) int {
	n := 0
	for _, v := range vals {
		if v {
			n++
		}
	}
	return n
}
