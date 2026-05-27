package commands

import (
	"context"

	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

var errNoConceptSubcommand = terr.New(
	"no_subcommand", 2,
	"run 'terminology concept --help' for available subcommands",
	"no subcommand specified",
)

func Concept() *urfcli.Command {
	return &urfcli.Command{
		Name:  "concept",
		Usage: "create, update, or remove a concept entry",
		Action: func(_ context.Context, _ *urfcli.Command) error {
			return errNoConceptSubcommand
		},
		Commands: []*urfcli.Command{
			ConceptAdd(),
			ConceptUpdate(),
			ConceptRemove(),
		},
	}
}
