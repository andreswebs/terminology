package commands

import (
	"context"

	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

var errNoTermSubcommand = terr.New(
	"no_subcommand", 2,
	"run 'terminology term --help' for available subcommands",
	"no subcommand specified",
)

// Term constructs the "term" command, which groups the subcommands for adding
// or deprecating a term within an existing concept.
func Term() *urfcli.Command {
	return &urfcli.Command{
		Name:  "term",
		Usage: "add or deprecate a term within an existing concept",
		Action: func(_ context.Context, _ *urfcli.Command) error {
			return errNoTermSubcommand
		},
		Commands: []*urfcli.Command{
			TermAdd(),
			TermDeprecate(),
		},
	}
}
