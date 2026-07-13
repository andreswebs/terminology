// Package app wires the terminology CLI: the root command, its global flags,
// and the assembly of all subcommands.
package app

import "github.com/andreswebs/terminology/internal/terr"

// ErrConflictingVerbosity reports that more than one of --verbose, --debug, or
// --quiet was supplied.
var ErrConflictingVerbosity = terr.New(
	"conflicting_verbosity", 2,
	"--verbose, --debug, and --quiet are mutually exclusive",
	"pass at most one of --verbose, --debug, --quiet",
)
