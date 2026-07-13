package write

import "github.com/andreswebs/terminology/internal/terr"

// ErrInvalidID reports that a concept ID could not be derived from the term.
var ErrInvalidID = terr.New(
	"invalid_id", 65,
	"supply an explicit --id when the term has no Latin/numeric characters",
	"derived concept ID is empty",
)

// ErrDuplicateID reports that a concept with the given ID already exists.
var ErrDuplicateID = terr.New(
	"duplicate_id", 65,
	"use 'concept update' to modify an existing concept",
	"concept ID already exists",
)

// ErrNotFound reports that no concept matched the given ID.
var ErrNotFound = terr.New(
	"not_found", 65,
	"use 'concept add' to create a new concept, or check the ID",
	"concept not found",
)

// ErrDanglingCrossref reports that a write would leave cross-references
// pointing at a removed concept.
var ErrDanglingCrossref = terr.New(
	"dangling_crossref", 65,
	"remove or update cross-references before removing the target concept, or use --force",
	"write would leave dangling cross-references",
)

// ErrInvalidInput reports that a payload is malformed or in an unsupported
// format.
var ErrInvalidInput = terr.New(
	"invalid_input", 65,
	"check the JSON payload structure; use 'terminology schema' for the expected shape",
	"stdin payload is malformed or unsupported",
)

// ErrApplyValidationFailed reports that an apply payload failed validation and
// no changes were written.
var ErrApplyValidationFailed = terr.New(
	"apply_validation_failed", 1,
	"fix per-concept errors in failures[] and retry",
	"apply payload failed validation; no changes written",
)
