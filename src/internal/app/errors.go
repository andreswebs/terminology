package app

import "github.com/andreswebs/terminology/internal/terr"

var ErrConflictingVerbosity = terr.New(
	"conflicting_verbosity", 2,
	"--verbose, --debug, and --quiet are mutually exclusive",
	"pass at most one of --verbose, --debug, --quiet",
)
