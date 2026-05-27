---
id: ter-u2b3
status: closed
deps: [ter-xby1, ter-pwd0, ter-slnb]
links: []
created: 2026-05-27T00:36:05Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, task, write, reconciliation]
---
# E8.T4 — Reconciliation algorithm

Implement the core apply reconciliation algorithm that diffs payload against the current glossary.

## Spec refs

- [008-apply.md §Algorithm](docs/specs/008-apply.md)
- [008-apply.md §Update rule](docs/specs/008-apply.md)
- [008-apply.md §--prune and dangling cross-references](docs/specs/008-apply.md)
- [008-apply.md §Idempotency](docs/specs/008-apply.md)

## Scope

### Reconciliation function

func Reconcile(g *tbx.Glossary, payload []tbx.Concept, prune bool) (*ReconcileResult, error)

Algorithm:
1. Build a map of current glossary concepts by ID.
2. For each payload concept:
   - Not in glossary → add to "added" set.
   - In glossary, content differs (via ConceptsEqual) → replace (wholesale). Add to "updated" set.
   - In glossary, content matches → add to "unchanged" set.
3. If prune: every glossary concept absent from payload → candidate for removal.
   - For each removal candidate, check if any remaining concept (payload-present or preserved) has a crossReference targeting it.
   - If dangling crossref found → refuse with dangling_crossref (entire operation aborted, file untouched).
   - Otherwise → add to "removed" set.
4. Apply mutations to glossary in-memory: append new concepts, replace updated ones, remove pruned ones.

### Update rule

Wholesale replace: the payload concept replaces the existing concept entirely (except ID is preserved from the existing entry). Payload-omitted fields are dropped.

### Transaction records

If --transaction is set, append a transacGrp (modification type) to each added and updated concept. Unchanged concepts get no transaction.

### Validation

After all mutations, validate the resulting in-memory glossary (g.Validate(false)). If validation fails, abort with ErrApplyValidationFailed containing the failures[] detail array. File is never partially written.

### ReconcileResult type

type ReconcileResult struct {
    Added     []string  // concept IDs, sorted ASCII
    Updated   []string
    Removed   []string
    Unchanged []string
}

All lists sorted ASCII byte order by concept_id per determinism ADR.

### Idempotency

Running the same payload twice yields zero ops on second run. The transac-strip equality compare guarantees transaction records from the first run don't flip the second run to "update".

## Acceptance Criteria

- make build passes
- Reconciliation correctly categorizes: add, update, unchanged, remove
- Wholesale replace on update (payload-omitted fields dropped)
- --prune removes absent concepts; no prune preserves them
- Dangling crossref on prune → error, file untouched
- Transaction records appended only to added/updated concepts
- Post-mutation validation; failures produce ErrApplyValidationFailed with details
- All ID lists sorted ASCII byte order
- Idempotency: same payload twice → all unchanged on second run
- Unit tests for each scenario


## Notes

**2026-05-27T12:16:54Z**

Implemented Reconcile() and ReconcileWithTxn() in internal/write/reconcile.go with 16 unit tests in reconcile_test.go. The reconciliation algorithm: (1) builds a map of current concepts by ID, (2) categorizes payload concepts as add/update/unchanged using ConceptsEqual, (3) optionally prunes absent concepts with dangling-crossref checking, (4) validates the resulting glossary via validateForWrite, (5) returns sorted ID lists. ReconcileWithTxn adds transaction records to added/updated concepts only. Wholesale replace preserves the existing concept ID. The crossref check uses findCrossRefsToInSlice (parallel to concept_remove's findCrossRefsTo) scanning both concept-level and term-level CrossRefs. The fakeClock type was already declared in transaction_test.go, so reconcile_test.go reuses it without redeclaring.
