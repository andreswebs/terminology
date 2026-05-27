---
id: ter-g5id
status: closed
deps: []
links: []
created: 2026-05-27T15:17:25Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-nd3x
tags: [e9, hardening, fuzz]
---
# E9.T6 — Fuzz tests (TBX decoder, matcher, DeriveID)

Add Go native fuzz tests for the three targets specified in the spec. Corpus committed under testdata/fuzz/. Designed for nightly CI runs on Linux (-fuzztime=30s), not PR-level.

## Acceptance Criteria

- FuzzLinguistDecode in internal/tbx/linguist: arbitrary bytes → must never panic, must return error or valid Glossary
- FuzzMatcherScan in internal/match: arbitrary text + small in-memory glossaries → must never panic, must terminate
- FuzzDeriveID in internal/write: arbitrary strings → must never panic, must return valid ID or ErrInvalidID
- Seed corpus committed under testdata/fuzz/ for each target (include known-good inputs + edge cases)
- Each fuzz function runs successfully with go test -fuzz=. -fuzztime=5s locally
- Any crashers found during initial fuzzing are fixed and added to corpus
- make build passes (fuzz tests are standard Go fuzz, no build tags needed to compile)


## Notes

**2026-05-27T17:04:56Z**

Implemented three Go native fuzz tests:

1. FuzzLinguistDecode (internal/tbx/linguist/fuzz_test.go): Feeds arbitrary bytes to linguist.NewReader().Decode(). Seeds include canonical DCT/DCA fixtures plus edge cases (empty, non-XML, bare tags, binary). Validates returned Glossary structure when no error.

2. FuzzMatcherScan (internal/match/fuzz_test.go): Feeds arbitrary text strings through a pre-built Matcher with a small multi-language glossary (en+he, preferred+deprecated terms). Validates match invariants (non-empty ConceptID/Term, positive Line/Column).

3. FuzzDeriveID (internal/write/fuzz_test.go): Feeds arbitrary strings to DeriveID. Validates output invariants: only [a-z0-9-], no leading/trailing hyphens, max 64 codepoints, valid UTF-8. Errors must be ErrInvalidID.

Seed corpus committed under testdata/fuzz/ in each package. All three pass with -fuzztime=5s locally. No crashers found during initial fuzzing. make build passes.
