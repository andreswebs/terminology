---
id: ter-xjpk
status: closed
deps: []
links: []
created: 2026-05-27T13:46:20Z
type: bug
priority: 0
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, bug, e3, validator]
---
# BUG: Crossref validator single-pass ordering breaks idempotency after sorted write

## Summary

The crossref validator in `src/internal/tbx/validate.go` (lines 22–34)
populates `idSeen` left-to-right as it iterates through concepts, then
checks each crossref target against only the IDs seen so far. When
`apply` writes concepts sorted alphabetically, any crossref whose source
sorts before its target produces a false-positive `unresolved_crossref`
warning.

This breaks idempotency: the second `apply` with the same payload fails
with `validation_error` (exit 65) because the file written by the first
apply has concepts in sorted order.

## Reproduce

```sh
# Create a glossary with sefirah (refs tzimtzum) and tzimtzum.
# First apply succeeds. Second apply fails.
$TT apply --tbx "${W}" --file payload.json   # exit 0
$TT apply --tbx "${W}" --file payload.json   # exit 65, validation_error
```

The crossref `sefirah → tzimtzum` fails because `sefirah` sorts before
`tzimtzum` alphabetically, so `tzimtzum` hasn't been added to `idSeen`
when sefirah's crossref is checked.

## Root cause

`src/internal/tbx/validate.go` — single-pass crossref check:

```go
idSeen := make(map[string]int)
for i, c := range g.Concepts {
    idSeen[c.ID] = i  // populated as we go
    for _, cr := range c.CrossRefs {
        if _, found := idSeen[cr.Target]; !found {  // only checks what's seen so far
            // false positive unresolved_crossref
        }
    }
}
```

## Fix

Collect all concept IDs in a first pass, then check crossrefs in a second pass.

## Origin

Pre-existing E3 validator bug surfaced by E8's deterministic sorted output.

## Affected QA tests

- TC-IDEM-001, TC-IDEM-002 (idempotency)
- TC-EQ-001 (concept equality)

## Refs

- QA report: `qa/E8-manual-qa.report.md` (BUG-001)
- Spec: `docs/specs/008-apply.md`


## Notes

**2026-05-27T13:50:49Z**

Fixed by splitting Glossary.Validate() into two passes: pass 1 collects all concept IDs and checks for duplicate_id, pass 2 validates lang tags, missing terms, and crossrefs against the complete ID set. This fixes false-positive unresolved_crossref warnings when a forward-referencing concept sorts alphabetically before its target. Golden file for validate/warnings updated because warning order changed (duplicate_id now emitted before unresolved_crossref). Two new unit tests added: TestValidate_CrossRefForwardReference and TestValidate_TermCrossRefForwardReference.
