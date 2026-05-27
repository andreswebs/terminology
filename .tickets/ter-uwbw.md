---
id: ter-uwbw
status: closed
deps: []
links: []
created: 2026-05-26T15:30:28Z
type: bug
priority: 1
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, bug, match, niqqud]
---
# E5.BUG — Niqqud matching fails without --lang filter

## Summary

When scanning a markdown file without `--lang`, the plain Hebrew form `ספירה`
(no niqqud) is not matched against the glossary term `סְפִירָה` (with niqqud).
With `--lang he`, both the niqqud and plain forms match correctly.

## QA reference

E5 manual QA report Finding F1: `qa/E5-manual-qa.report.md`
Test case: TC-SCAN-NIQ-001 (P0)

## Reproduction

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"
QA_TMP=$(mktemp -d)

# Create glossary with Hebrew term סְפִירָה (with niqqud)
# and corpus with both סְפִירָה (line 3) and ספירה (line 4)
# (see qa/E5-manual-qa.md §"Test corpus setup" for full fixtures)

# Without --lang: only 1 Hebrew sefirah match (line 3)
${TT} scan "${QA_TMP}/niqqud.md" --tbx "${QA_TMP}/glossary.tbx" \
  | jq '[.matches[] | select(.concept_id == "sefirah" and .lang == "he")] | length'
# Actual: 1
# Expected: 2

# With --lang he: both forms match (lines 3 and 4)
${TT} scan "${QA_TMP}/niqqud.md" --tbx "${QA_TMP}/glossary.tbx" --lang he \
  | jq '[.matches[] | select(.concept_id == "sefirah" and .lang == "he")] | length'
# Actual: 2 (correct)
```

## Root cause hypothesis

When all language patterns are compiled into the AC automaton (no `--lang`
filter), the English terms "sefirah" and "sephirah" (Latin script) coexist
with the Hebrew term `סְפִירָה`. The niqqud-stripped canonical form of
`סְפִירָה` is `ספירה` (5 Hebrew characters, 10 bytes). A pattern collision
or deduplication in `buildAutomaton` may cause the Hebrew canonical pattern
to be dropped or overwritten when multiple language patterns produce
overlapping entries.

With `--lang he`, only Hebrew patterns are compiled, avoiding the collision.

## Fix direction

Investigate `match.New()` and `buildAutomaton()` to determine how patterns
from different languages with different scripts are compiled. Ensure that
patterns with identical canonical bytes but different language/concept
metadata are all preserved (emitting multiple matches or keeping the
correct mapping). The offset map (`Canonical.Map`) should correctly map
back to original text positions regardless of how many patterns share the
same canonical form.

## Affected components

- `src/internal/match/match.go` — `New()` pattern compilation
- `src/internal/match/ac.go` — `buildAutomaton()` pattern deduplication

## Refs

- E5 spec: docs/specs/005-matcher.md §"Normalization pipeline"
- E5 QA plan: qa/E5-manual-qa.md §TC-SCAN-NIQ-001


## Notes

**2026-05-26T15:34:31Z**

Fix: When lang is empty (scan all languages), New() now normalizes each pattern using PolicyFor(ltag) (the language-specific policy) instead of the caller's single Baseline policy. A mergePolicy() function ORs boolean fields from all language policies to produce a merged policy used for text normalization. This ensures Hebrew patterns with niqqud are stripped consistently with how the text is normalized, so plain Hebrew forms in the corpus match glossary terms with niqqud marks. The fix is in match.go (New function) and policy.go (mergePolicy helper). One new test added: TestScan_NiqqudMatchWithoutLangFilter.
