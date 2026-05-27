---
id: ter-dgow
status: closed
deps: [ter-nur0]
links: []
created: 2026-05-27T00:36:37Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, task, testing, golden]
---
# E8.T6 — Golden CLI tests for apply

Golden CLI tests for the apply command covering all major scenarios.

## Spec refs

- [008-apply.md](docs/specs/008-apply.md)
- [testing.md §Golden CLI tests](docs/adr/testing.md)

## Scope

Test fixtures in src/internal/app/commands/testdata/apply/:

### Fixtures needed

- Base glossary TBX file with 2-3 concepts (crossrefs between them)
- JSON payload: add new concept
- JSON payload: update existing concept (content change)
- JSON payload: mix of add, update, unchanged
- JSON payload: all unchanged (idempotency)
- JSON payload + --prune: remove absent concept
- JSON payload + --prune: dangling crossref refusal
- TBX fragment payload
- Malformed JSON payload
- Dry-run scenario

### Golden file triples per scenario

- .stdout (ApplyEnvelope JSON)
- .stderr (empty on success, error envelope on failure)
- .exit (exit code)

### Test scenarios

1. apply-add: new concept via JSON → added:[id], exit 0
2. apply-update: changed concept → updated:[id], exit 0
3. apply-unchanged: identical concept → unchanged:[id], exit 0
4. apply-mixed: add + update + unchanged → correct categorization, exit 0
5. apply-idempotent: run same payload on already-converged file → all unchanged, exit 0
6. apply-prune: absent concept removed → removed:[id], exit 0
7. apply-prune-crossref: prune would dangle → dangling_crossref, exit 65
8. apply-tbx-fragment: TBX input → correct categorization, exit 0
9. apply-dry-run: file unchanged after dry-run, exit 0
10. apply-invalid-json: malformed payload → invalid_input, exit 65
11. apply-validation-failed: payload with bad data → apply_validation_failed, exit 1
12. apply-no-tbx: missing --tbx → no_tbx_path, exit 2
13. apply-transaction: --transaction adds transacGrp records

### Pattern

Follow existing golden test infrastructure in commands_test.go / golden_test.go. Use fake clock for transaction timestamps.

## Acceptance Criteria

- make build passes (including make test)
- All golden test scenarios pass
- Golden files committed as byte-for-byte reference
- Covers success, error, dry-run, prune, idempotency, transaction paths
- Tests exercise both JSON and TBX payload formats


## Notes

**2026-05-27T12:32:51Z**

Implemented 13 golden CLI tests for the apply command in src/internal/app/apply_golden_test.go, covering: add, update, unchanged, mixed (add+update+unchanged), idempotent (double-apply), prune, prune with dangling crossref (exit 65), TBX fragment input, dry-run, invalid JSON (exit 65), no TBX path (exit 2), transaction with fake clock, and validation error via dangling crossref in payload (exit 65). The validation_failed scenario (exit 1 via ErrApplyValidationFailed) was replaced with validation_error because ErrApplyValidationFailed is defined but not yet used in the reconciliation code path — only ErrValidationError (exit 65) fires via validateForWrite. All golden files committed as byte-for-byte references. Helper writePayloadFile added for creating temp payload files. Pattern follows write_golden_test.go conventions: copyFixture for mutable TBX, writeCtx for fake clock, pipeStdin not needed (apply uses --file not stdin).
