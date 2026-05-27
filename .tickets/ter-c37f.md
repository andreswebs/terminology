---
id: ter-c37f
status: closed
deps: []
links: []
created: 2026-05-27T15:17:06Z
type: task
priority: 1
assignee: Andre Silva
parent: ter-nd3x
tags: [e9, hardening, security]
---
# E9.T4 — Bounded reads (io.LimitedReader on all external inputs)

Wrap every external input with io.LimitedReader to cap resource consumption. Caps per spec: TBX 50 MB, markdown/stdin/extract-per-file 10 MB. Declare ErrInputTooLarge sentinel (exit 2, code input_too_large).

## Acceptance Criteria

- TBX loading (tbx.Load / linguist.Reader.Decode): capped at 50 MB via io.LimitedReader
- Markdown files (scan, check): capped at 10 MB
- Stdin payloads (apply --file -, concept add JSON stdin, term add/deprecate stdin): capped at 10 MB
- Extract per-file: capped at 10 MB
- Apply --file (non-stdin): capped at 10 MB
- ErrInputTooLarge sentinel: code "input_too_large", exit 2, hint "split the input into smaller files or batches"
- Hitting the cap returns structured error with the cap value echoed in the message
- Unit tests for each cap (mock reader that exceeds limit)
- Golden test for input_too_large error envelope
- make build passes


## Notes

**2026-05-27T16:54:43Z**

Implemented bounded reads (io.LimitedReader) on all external inputs. ErrInputTooLarge sentinel in tbx/errors.go (code: input_too_large, exit 2). ReadBounded/ReadFileBounded helpers in tbx/bounded.go. TBX loading capped at 50MB, markdown/stdin/payload at 10MB. Wired into: tbx.Load, scan, check, extract, concept add/update stdin, apply --file, apply stdin. Added wrapLoadError helper so input_too_large passes through instead of being masked by ErrValidationError. Unit tests in tbx/bounded_test.go, integration tests for all 5 commands, golden test for the error envelope. Schema golden files regenerated. Side effect: ErrDangerousDoctype now passes through correctly with its own code instead of being masked — doctype_entity golden file updated.
