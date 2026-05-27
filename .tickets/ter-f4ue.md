---
id: ter-f4ue
status: closed
deps: []
links: []
created: 2026-05-26T15:31:44Z
type: bug
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, bug, scan, context]
---
# E5.BUG — --context flag has no effect on context window size

## Summary

The `--context` flag on the `scan` command is parsed and appears in
`--help` output with a default of 80, but the value has no effect on the
context window size in match output. All context strings are identical
regardless of the `--context` value.

## QA reference

E5 manual QA report Finding F3: `qa/E5-manual-qa.report.md`
Test case: TC-SCAN-CTX-002 (P1)

## Reproduction

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"
QA_TMP=$(mktemp -d)
# (see qa/E5-manual-qa.md §"Test corpus setup" for fixtures)

# Default context (80):
${TT} scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" \
  | jq -r '.matches[1].context'
# Output: "The concept of tzimtzum is central to Kabbalistic thought. It d..."
# Length: 66

# Custom context (10):
${TT} scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" --context 10 \
  | jq -r '.matches[1].context'
# Output: "The concept of tzimtzum is central to Kabbalistic thought. It d..."
# Length: 66 (same — flag ignored)

# Custom context (40):
${TT} scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" --context 40 \
  | jq -r '.matches[1].context'
# Output: identical to above
```

## Root cause

The `--context` flag value is read from the CLI but not passed through to
the context extraction logic in either the scan command action or the
matcher's `Scan()` method. The context appears to default to the full
span/line text with fixed-length ellipsis truncation, ignoring the
user-supplied value.

## Fix direction

Thread the `--context` value from `scanAction` through to the context
extraction step. The context window should be `N` characters total
(centered on the match), with ellipsis markers when truncated. Check both:

1. `src/internal/app/commands/scan.go` — `scanAction` reads the flag but
   may not pass it to the matcher or envelope builder.
2. `src/internal/match/match.go` — `Scan()` or a helper function that
   extracts context around each match.

## Affected components

- `src/internal/app/commands/scan.go` — flag reading and passthrough
- `src/internal/match/match.go` — context extraction logic

## Refs

- E5 spec: docs/specs/005-matcher.md §"Context window"
- CLI design: docs/cli-design.md §"terminology scan" — `--context N`
- E5 QA plan: qa/E5-manual-qa.md §TC-SCAN-CTX-001, TC-SCAN-CTX-002


## Notes

**2026-05-26T15:46:38Z**

Fixed --context flag passthrough. The bug was that scanAction never read cmd.Int("context") and Scan() / extractContext() hardcoded window=40. Fix: (1) added contextSize int param to Matcher.Scan() with 0→80 default, (2) changed extractContext to accept contextSize and compute window=contextSize/2, (3) read cmd.Int("context") in scanAction and pass as int(contextSize). Added 2 new unit tests (TestScan_ContextWindowCustomSize, TestScan_ContextWindowZeroUsesDefault). Updated context_window golden file.
