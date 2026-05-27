---
id: ter-2kny
status: closed
deps: []
links: []
created: 2026-05-27T15:17:36Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-nd3x
tags: [e9, hardening, perf]
---
# E9.T7 — Perf budget tests + make perf target

Implement performance budget tests behind the 'perf' build tag. Generate synthetic TBX/markdown fixtures for stress testing. Add make perf target. Budgets are hardcoded constants in test files — ratcheted via PR review, not config file.

## Acceptance Criteria

- Perf tests in internal/<pkg>/perf_test.go behind //go:build perf tag
- Targets per spec:
  - validate: 10k-concept TBX < 500ms
  - lookup: 10k-concept TBX < 50ms
  - scan: 200-concept TBX × 100KB markdown < 100ms
  - scan: 5000-concept TBX × 5MB markdown < 5s
  - check: same as scan × 2 < 10s
  - extract: 1MB markdown corpus < 2s
- Synthetic fixture generation: Go test helpers that produce 10k-concept TBX and large markdown files (not committed as static files — generated at test time)
- make perf target runs: go test -tags perf -run TestPerf ./src/internal/...
- Budget constants hardcoded in test files (no config file)
- Tests fail if budget exceeded (t.Fatalf with timing info)
- make build passes (perf tests excluded from default build/test since they use build tag)


## Notes

**2026-05-27T17:12:23Z**

Implemented perf budget tests behind //go:build perf tag. Files created: internal/tbx/perf_test.go (validate 10K <500ms, lookup 10K <50ms), internal/match/perf_test.go (scan 200×100KB <100ms, scan 5000×5MB <5s), internal/check/perf_test.go (check 5000×5MB×2 <10s), internal/extract/perf_test.go (extract 1MB <2s). Added make perf target. Synthetic fixtures generated at test time using helper functions (generateGlossary, generateMarkdown, generateCorpus). Term density parameter controls how sparse/dense term occurrences are in generated markdown — dense fixtures can cause quadratic blowup in match count processing. All budgets pass with margin (actual times ~10-50% of budget on CI-like ARM64).
