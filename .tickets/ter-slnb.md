---
id: ter-slnb
status: closed
deps: [ter-6dsf]
links: []
created: 2026-05-27T00:35:43Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, task, write, parsing]
---
# E8.T3 — Apply payload parsing (JSON + TBX + format detection)

Parse apply payloads from JSON and TBX formats, with format auto-detection.

## Spec refs

- [008-apply.md §Payload format](docs/specs/008-apply.md)
- [008-apply.md §Format selection precedence](docs/specs/008-apply.md)

## Scope

### JSON payload parser

Parse {"concepts": [...]} where each concept matches the WriteResult struct shape (same as lookup output). Uses existing output.WriteResult type and writeResultToConcept converter.

New type for the JSON wrapper:

type ApplyPayload struct {
    Concepts []output.WriteResult \`json:"concepts"\`
}

Strict decoding (DisallowUnknownFields) — unknown fields produce invalid_input.

### TBX payload parser

Reuse write.ParseTBXFragment which already handles <conceptEntry> and <conceptEntryList> fragments.

### Format detection

Decision: use extension + content sniffing (no --format flag for input).

1. File path with .json extension → JSON
2. File path with .tbx or .xml extension → TBX fragment
3. Stdin (--file -) or no extension match → content sniffing via looksLikeXML()
4. If content sniffing fails → invalid_input with hint

### File loading

Read --file flag value:
- If "-" → read stdin
- Otherwise → read file at path
- File not found → exit 3 (I/O error)

## Acceptance Criteria

- make build passes
- JSON payload parsed into []tbx.Concept via {"concepts": [...]}
- TBX payload parsed via existing ParseTBXFragment
- Extension-based detection works for .json, .tbx, .xml
- Content sniffing works for stdin
- Unknown JSON fields rejected with invalid_input
- Unit tests for JSON parsing, TBX parsing, format detection


## Notes

**2026-05-27T12:11:37Z**

Implemented apply payload parsing in internal/write/apply.go with: (1) ParseApplyJSON — parses {"concepts": [...]} wrapper into []tbx.Concept using DisallowUnknownFields for strict validation; (2) DetectPayloadFormat — extension-based (.json/.tbx/.xml) + content sniffing (first non-whitespace byte) for stdin/unknown extensions; (3) LoadApplyFile — reads file from path or stdin, auto-detects format, dispatches to ParseApplyJSON or existing ParseTBXFragment; (4) Exported WriteResultToConcept and WriteTermToTBXTerm (moved from commands/concept_add.go to write package to avoid duplication, concept_add.go now calls write.WriteResultToConcept). All unit tests in apply_test.go. make build passes clean.
