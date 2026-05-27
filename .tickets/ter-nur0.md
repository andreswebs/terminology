---
id: ter-nur0
status: closed
deps: [ter-u2b3]
links: []
created: 2026-05-27T00:36:23Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, task, command, apply]
---
# E8.T5 — Apply command action

Wire the apply command action, replacing the underConstruction stub. This is the integration point connecting payload parsing, reconciliation, and output emission.

## Spec refs

- [008-apply.md](docs/specs/008-apply.md)
- [008-apply.md §Concurrency](docs/specs/008-apply.md)

## Scope

### Command action (commands/apply.go)

Replace underConstruction with applyAction:

1. Resolve TBX path (--tbx or TERMINOLOGY_TBX).
2. Load payload from --file (path or stdin).
3. Detect format (extension + content sniffing).
4. Parse payload into []tbx.Concept.
5. Load glossary from TBX file.
6. Run reconciliation (with --prune flag).
7. If --transaction, inject transaction records into added/updated concepts.
8. Validate resulting glossary.
9. If validation fails → emit apply_validation_failed error envelope to stderr with failures[] details. Exit 1.
10. If --dry-run → skip file write.
11. Otherwise → save via tbx.Save (atomic rename with advisory lock).
12. Emit ApplyEnvelope to stdout. Exit 0.

### Lock scope

Apply holds the ${TBX}.lock advisory lock for the entire read→modify→validate→write window. tbx.Save already acquires the lock. For apply, we need to acquire it before Load and hold through Save, since the read+write must be atomic.

This requires a new execution path (not write.Execute, which loads+saves internally). The apply action should:
- Acquire lock
- Load glossary
- Run reconciliation
- Validate
- Write (if not dry-run)
- Release lock

### Error handling

- invalid_input (exit 65): malformed payload, unknown fields
- apply_validation_failed (exit 1): post-mutation validation failed, with failures[] detail
- no_tbx_path (exit 2): no --tbx and no TERMINOLOGY_TBX
- I/O errors (exit 3): file not found, permission denied
- dangling_crossref (exit 65): --prune would leave dangling refs

### Output

Success: ApplyEnvelope with applied.{added, updated, removed, unchanged} lists.
Failure: standard error envelope on stderr.

## Acceptance Criteria

- make build passes
- Apply command no longer returns underConstruction
- Payload loaded from file path and stdin (--file -)
- Format auto-detected from extension and content
- Reconciliation applied with correct add/update/remove/unchanged categorization
- --prune removes absent concepts
- --dry-run produces output without modifying file
- --transaction appends transacGrp to modified concepts
- Lock held across entire read-modify-write window
- Error envelopes match spec shape
- Exit codes: 0 (success), 1 (validation failed), 2 (usage), 3 (I/O), 65 (data error)


## Notes

**2026-05-27T12:24:39Z**

Implemented applyAction replacing underConstruction stub. Key decisions: (1) Exported AcquireLock from tbx/lock.go and added SaveLocked to tbx/io.go for lock-spanning read-modify-write. SaveLocked skips lock acquisition since the caller already holds it. (2) Apply uses a manual pipeline (like concept remove) rather than write.Execute, because it handles multiple concepts, reconciliation, and prune logic. (3) Payload loading delegates to write.LoadApplyFile which handles file/stdin, format detection (extension + content sniffing), and JSON/TBX parsing. (4) Reconciliation delegates to write.Reconcile or write.ReconcileWithTxn depending on --transaction flag. (5) Removed stub_test.go and commands/stub.go since apply was the last stub command. (6) Added 10 tests covering: add, dry-run, prune, idempotent unchanged, update detection, transaction, TBX fragment input, dangling crossref on prune, no-tbx-path error, missing-file error.
