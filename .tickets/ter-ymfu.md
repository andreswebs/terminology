---
id: ter-ymfu
status: closed
deps: []
links: []
created: 2026-05-26T15:31:21Z
type: bug
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, bug, match, multiword]
---
# E5.BUG — Multi-word terms not matched across line breaks

## Summary

Multi-word glossary terms are not matched when the words span a soft line
break within the same markdown paragraph. The term "tzimtzum primordial"
matches on a single line but fails when split as "tzimtzum\nprimordial"
across lines 3–4.

## QA reference

E5 manual QA report Finding F2: `qa/E5-manual-qa.report.md`
Test case: TC-SCAN-MW-001 (P1)

## Reproduction

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"
QA_TMP=$(mktemp -d)

# linebreak.md content:
# Line 3: "The concept of tzimtzum"
# Line 4: "primordial extends the basic idea of divine contraction."

# Multi-word term NOT matched across line break:
${TT} scan "${QA_TMP}/linebreak.md" --tbx "${QA_TMP}/glossary.tbx" \
  | jq '[.matches[] | select(.concept_id == "tzimtzum-primordial")] | length'
# Actual: 0
# Expected: 1

# Same term matched on single line in corpus.md:
${TT} scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" \
  | jq '[.matches[] | select(.concept_id == "tzimtzum-primordial")] | length'
# Actual: 1 (correct)
```

## Root cause

`markdown.Spans()` produces spans that end at line boundaries within a
paragraph. The matcher's whitespace normalization (`\s+` → single space)
operates per-span, so the newline between "tzimtzum" and "primordial"
acts as a span boundary rather than being collapsed.

Evidence: the context string for the "tzimtzum" match on line 3 is
`"The concept of tzimtzum"` — it ends at the line boundary, confirming
the span does not continue to line 4.

## Fix direction

Option A: Have `markdown.Spans()` join soft line breaks within the same
paragraph into a single span, so whitespace collapse can operate across
them.

Option B: Have the matcher concatenate consecutive spans from the same
paragraph before scanning, preserving the offset mapping.

Option A is simpler and more correct — a soft line break in markdown is
semantically equivalent to a space, not a paragraph break.

## Impact

Any multi-word glossary term split across a soft line break in markdown
prose will not be detected. This affects real-world documents where line
wrapping is common.

## Affected components

- `src/internal/markdown/` — span generation (if fix is Option A)
- `src/internal/match/match.go` — `Scan()` method (if fix is Option B)

## Refs

- E5 spec: docs/specs/005-matcher.md §"Normalization pipeline" — whitespace collapse
- E4 markdown package: src/internal/markdown/
- E5 QA plan: qa/E5-manual-qa.md §TC-SCAN-MW-001


## Notes

**2026-05-26T15:42:51Z**

Fix: Option A from ticket — merged consecutive sibling Text nodes connected by soft line breaks into single spans in markdown.Spans(). The walkText function now manually iterates children, detecting runs of Text nodes where SoftLineBreak() is true, and yields src[first.Start:last.Stop] as a single span. This preserves 1:1 byte offset mapping since the gap between consecutive sibling text nodes in goldmark is exactly the newline character. The extract package (all 3 heuristics: capitalized, foreign-script, high-frequency) needed a parallel fix: added Span.lineColAt(byteOffset) helper to correctly compute line/col by walking through newlines, replacing the naive span.Col + offset calculation. Golden test files regenerated. Hard line breaks (two trailing spaces) are not merged — goldmark's ast.Text.SoftLineBreak() correctly distinguishes the two cases.
