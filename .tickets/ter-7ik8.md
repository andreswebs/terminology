---
id: ter-7ik8
status: closed
deps: [ter-x3kh, ter-smzv, ter-2cqg, ter-iu37, ter-ppld]
links: []
created: 2026-05-26T19:31:25Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, golden, testing]
---
# E7.T12 — Golden CLI tests for write commands

Golden CLI tests for all 5 write commands using the existing runGolden harness. Each test captures argv + stdin → stdout/stderr/exit-code triples with byte-for-byte golden files. Tests need a writable TBX fixture (copy to temp dir before each test since writes mutate). Fake clock injection for deterministic transaction timestamps. Cover: happy paths (flags, JSON stdin, TBX fragment), error cases (duplicate_id, not_found, dangling_crossref, invalid_input), --dry-run, --transaction/--author, --merge/--replace, --force, ID derivation, ID stability.

## Acceptance Criteria

- make build passes
- Golden tests for concept add (flags, JSON stdin, TBX fragment, duplicate_id, dry-run, transaction)
- Golden tests for concept update (merge, replace, not_found, ID stability)
- Golden tests for concept remove (clean, dangling_crossref, force, not_found)
- Golden tests for term add (new langSec, existing langSec, not_found)
- Golden tests for term deprecate (happy, not_found)
- All tests use byte-for-byte golden comparison
- Exit codes verified
- Fake clock for deterministic timestamps


## Notes

**2026-05-26T23:30:43Z**

Implemented 20 golden CLI tests for all 5 write commands in write_golden_test.go. Added runGoldenCtx variant to golden_test.go for context injection (fake clock). Tests cover: concept add (flags, dry-run, duplicate_id, transaction with fake clock, JSON stdin, TBX fragment, ID derivation), concept update (merge, replace, not_found, ID stability), concept remove (clean, not_found, dangling_crossref, force), term add (existing langSec, new langSec, not_found), term deprecate (happy, not_found). Helper functions: copyFixture (copies TBX to temp dir), pipeStdin (os.Pipe for stdin injection), fakeClock (deterministic timestamps). All byte-for-byte golden comparison with correct exit codes verified.
