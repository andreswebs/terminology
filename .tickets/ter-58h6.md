---
id: ter-58h6
status: closed
deps: []
links: []
created: 2026-05-26T03:20:00Z
type: chore
priority: 3
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, chore, qa]
---
# E4.CHORE — Update QA plan for --lang and --fields semantics

## Summary

The E4 manual QA plan has test cases that assume different semantics from
the implementation for `--lang` on lookup and `--fields` path resolution.
The implementation is correct; the test plan needs updating.

## Changes needed

### Finding 1 — TC-LANG-LKP-001: `--lang` restricts search scope

The test expects `lookup "tzimtzum" --lang he` to find the concept (matched
via the English term) and filter the output to only show the Hebrew section.

The actual behavior: `--lang he` restricts the search to Hebrew-language
terms only. Since "tzimtzum" is the English term, it returns empty. The
Hebrew term "צמצום" must be used to match within Hebrew.

**Fix**: Update TC-LANG-LKP-001 to use the Hebrew term, or rephrase the
test to verify that `--lang` restricts the search scope.

### Finding 2 — TC-FLD-001, TC-FLD-003, TC-FLD-EXT-001: envelope-relative paths

The tests use bare result-level paths (`concept_id`, `term,frequency`).
The implementation requires envelope-relative paths
(`results.concept_id`, `candidates.term,candidates.frequency`).

**Fix**: Update all `--fields` test cases to use the full envelope-relative
paths. The `terminology schema --command CMD` output confirms these paths.

### Test plan errata

Also fix the `hebrew-frontmatter.md` corpus setup where `## cat >` uses
`##` (markdown heading) instead of a bare `cat` command.

## Refs

- QA report: [qa/E4-manual-qa.report.md](qa/E4-manual-qa.report.md)
  Findings 1, 2
- QA plan: [qa/E4-manual-qa.md](qa/E4-manual-qa.md) TC-LANG-LKP-001,
  TC-FLD-001, TC-FLD-003, TC-FLD-EXT-001

