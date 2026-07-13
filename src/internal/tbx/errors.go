package tbx

import "github.com/andreswebs/terminology/internal/terr"

// ErrUnsupportedDialect indicates the TBX dialect is not supported.
var ErrUnsupportedDialect = terr.New(
	"unsupported_dialect", 65,
	"supported: TBX-Linguist",
	"unsupported TBX dialect",
)

// ErrTBXLocked indicates the TBX file is locked by another process.
var ErrTBXLocked = terr.New(
	"tbx_locked", 3,
	"another terminology process is writing; retry",
	"TBX file is locked by another process",
)

// ErrNoTBXPath indicates no TBX file path was provided.
var ErrNoTBXPath = terr.New(
	"no_tbx_path", 2,
	"set --tbx or TERMINOLOGY_TBX",
	"no TBX file path provided",
)

// ErrValidationError indicates TBX validation failed.
var ErrValidationError = terr.New(
	"validation_error", 65,
	"check the TBX file structure and content",
	"TBX validation failed",
)

// ErrDangerousDoctype indicates a DOCTYPE declaration with entity
// declarations or external IDs was rejected.
var ErrDangerousDoctype = terr.New(
	"invalid_input", 65,
	"only bare <!DOCTYPE tbx> is accepted; entity declarations and external IDs are rejected",
	"dangerous DOCTYPE declaration",
)

// ErrNestingTooDeep indicates the XML nesting depth exceeds the maximum.
var ErrNestingTooDeep = terr.New(
	"invalid_input", 65,
	"nesting_too_deep",
	"XML nesting depth exceeds maximum of 256 levels",
)

// ErrInputTooLarge indicates the input exceeds the maximum allowed size.
var ErrInputTooLarge = terr.New(
	"input_too_large", 2,
	"split the input into smaller files or batches",
	"input exceeds maximum size",
)
