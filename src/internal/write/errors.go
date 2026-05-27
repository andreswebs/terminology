package write

import "github.com/andreswebs/terminology/internal/terr"

var ErrInvalidID = terr.New(
	"invalid_id", 65,
	"supply an explicit --id when the term has no Latin/numeric characters",
	"derived concept ID is empty",
)

var ErrDuplicateID = terr.New(
	"duplicate_id", 65,
	"use 'concept update' to modify an existing concept",
	"concept ID already exists",
)

var ErrNotFound = terr.New(
	"not_found", 65,
	"use 'concept add' to create a new concept, or check the ID",
	"concept not found",
)

var ErrDanglingCrossref = terr.New(
	"dangling_crossref", 65,
	"remove or update cross-references before removing the target concept, or use --force",
	"write would leave dangling cross-references",
)

var ErrInvalidInput = terr.New(
	"invalid_input", 65,
	"check the JSON payload structure; use 'terminology schema' for the expected shape",
	"stdin payload is malformed or unsupported",
)

var ErrApplyValidationFailed = terr.New(
	"apply_validation_failed", 1,
	"fix per-concept errors in failures[] and retry",
	"apply payload failed validation; no changes written",
)
