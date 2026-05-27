---
id: ter-4ory
status: closed
deps: [ter-fswn]
links: []
created: 2026-05-27T15:16:48Z
type: task
priority: 0
assignee: Andre Silva
parent: ter-nd3x
tags: [e9, hardening]
---
# E9.T2 — Wire sanitizers into command action handlers

Call the sanitize* functions at the start of every action handler that accepts concept IDs, language tags, file paths, or terms. Single point of trust per request — inner packages assume clean inputs.

## Acceptance Criteria

- concept add: sanitizeConceptID (if --id given), sanitizeLangTag, sanitizeTerm
- concept update: sanitizeConceptID, sanitizeLangTag (if --lang given), sanitizeTerm (if --term given)
- concept remove: sanitizeConceptID
- term add: sanitizeConceptID, sanitizeLangTag, sanitizeTerm
- term deprecate: sanitizeConceptID, sanitizeLangTag, sanitizeTerm
- apply: sanitizePath for --file (not stdin); concept IDs and lang tags in payload validated post-parse
- scan: sanitizePath for file argument
- check: sanitizePath for source/target arguments, sanitizeLangTag for --lang
- extract: sanitizePath for file arguments
- lookup: sanitizeConceptID for the ID argument, sanitizeLangTag for --lang (if given)
- validate: no additional sanitization needed (TBX path handled by --tbx resolution)
- --tbx flag path: sanitized (reject .., percent-encoding) but NOT sandboxed to CWD (per spec exemption)
- Golden tests added/updated for rejected inputs (exit 65, correct error code)
- make build passes


## Notes

**2026-05-27T15:41:58Z**

Wired sanitize* functions into all 10 command action handlers at the command-action boundary. Changes:

1. sanitize.go: Added sanitizeTBXPath (rejects control chars, .., %, ?# — no CWD sandbox per spec exemption) and tbxPathFromRoot helper (combines --tbx retrieval + sanitization).

2. All commands using --tbx now go through tbxPathFromRoot instead of raw cmd.Root().String('tbx').

3. Per-command sanitization:
   - concept add: sanitizeConceptID (--id), sanitizeLangTag (--lang), sanitizeTerm (--term) — all conditional on flag being set
   - concept update: sanitizeConceptID (positional), sanitizeLangTag/sanitizeTerm (if --lang/--term set)
   - concept remove: sanitizeConceptID (positional)
   - term add: sanitizeConceptID (positional), sanitizeLangTag (--lang), sanitizeTerm (--term)
   - term deprecate: sanitizeConceptID (positional), sanitizeLangTag (--lang), sanitizeTerm (--term)
   - lookup: sanitizeTerm (positional), sanitizeLangTag (--lang if set)
   - scan: sanitizePath (file arg, sandboxed to CWD)
   - check: sanitizePath (both src/tgt args), sanitizeLangTag (--source-lang/--target-lang if set)
   - extract: sanitizePath (all file args)
   - apply: sanitizePath (--file, not stdin), post-parse validation of concept IDs and lang tags in payload
   - validate: tbxPathFromRoot only

4. scan/check/extract keep original user-supplied paths for display in output envelopes, using sanitized absolute paths only for I/O.

5. 15 new golden tests under testdata/sanitize/ covering all rejection paths (invalid_id, invalid_lang_tag, invalid_term, invalid_path, tbx_path_traversal).

6. Updated apply test infrastructure: writePayloadFile and writePayloadInCWD create temp dirs inside testdata/ (within CWD) to satisfy the path sandbox. TestApply_FileNotFound uses a relative within-CWD path instead of absolute /nonexistent/path.json.

7. Regenerated golden files for: apply/file_not_found, scan/file_not_found, check/file_not_found (error messages now include absolute resolved paths from os.ReadFile).

make build passes clean.
