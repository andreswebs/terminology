---
id: ter-fswn
status: closed
deps: []
links: []
created: 2026-05-27T15:16:37Z
type: task
priority: 0
assignee: Andre Silva
parent: ter-nd3x
tags: [e9, hardening]
---
# E9.T1 — Input sanitization functions (sanitize.go)

Implement the four sanitization functions in internal/app/commands/sanitize.go with comprehensive test table in sanitize_test.go. Functions are unexported, called at the command-action boundary.

## Acceptance Criteria

- sanitizeConceptID(s) rejects control characters, path traversals, percent-encoded segments, embedded query params (?/#)
- sanitizeLangTag(s) validates BCP 47 format, rejects control characters and percent-encoded segments
- sanitizePath(s, baseDir) returns cleaned absolute path within baseDir; rejects .., %2e, percent-encoded segments, embedded query params; resolves symlinks and reapplies prefix check
- sanitizeTerm(s) rejects control characters
- Each function returns a terr-coded error (invalid_id, invalid_lang_tag, invalid_path, invalid_term) with exit 65
- Test table in sanitize_test.go is the authoritative spec of accepted vs rejected patterns
- New terr sentinels declared for invalid_lang_tag, invalid_path, invalid_term (invalid_id already exists in write/errors.go — reuse or relocate)
- make build passes


## Notes

**2026-05-27T15:27:15Z**

Implemented sanitize.go and sanitize_test.go in internal/app/commands/. Four unexported functions: sanitizeConceptID (rejects control chars, path traversals, percent-encoded segments, query params), sanitizeLangTag (validates BCP 47 structure, rejects control chars and percent-encoded segments), sanitizePath (rejects control chars, percent-encoded segments, query params, path traversals; resolves absolute path within baseDir; checks symlink escape), sanitizeTerm (rejects control chars). Four terr.New sentinels: ErrInvalidSanitizeID (invalid_id, exit 65), ErrInvalidLangTag (invalid_lang_tag, exit 65), ErrInvalidPath (invalid_path, exit 65), ErrInvalidTerm (invalid_term, exit 65). BCP 47 validation uses a lightweight structural check (2-8 alpha primary tag, 1-8 alphanumeric subtags) rather than importing x/text/language, keeping the commands package free of that dependency. Schema golden files regenerated since new sentinels changed terr.All() output.
