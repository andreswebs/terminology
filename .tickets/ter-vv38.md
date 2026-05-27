---
id: ter-vv38
status: closed
deps: [ter-y57h]
links: []
created: 2026-05-26T17:24:30Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-7fyo
tags: [e6, task, check, strict]
---
# E6.T5 — Check --strict + admitted_variant

## Goal

Extend the check algorithm to handle admitted variants. Without `--strict`, admitted variants in TGT generate warnings. With `--strict`, they become `admitted_variant` violations.

## Refs

- E6 spec: [docs/specs/006-scan-check.md](docs/specs/006-scan-check.md) §"--strict"
- CLI design: [docs/cli-design.md](docs/cli-design.md) §"--strict semantics" table
- TBX status: `StatusAdmitted` in `src/internal/tbx/model.go`

## Files to create / modify

- `src/internal/check/check.go` — extend `Check` to handle admitted terms
- `src/internal/check/check_test.go` — tests for --strict behavior

## Behavior

For each concept with source occurrences in SRC, after checking preferred and deprecated terms:

1. Find admitted target terms for this concept in `tgtLang`.
2. Scan TGT matches for admitted terms.
3. If `strict == false`: each admitted match → `CheckWarning{Type: "admitted_variant", ...}`.
4. If `strict == true`: each admitted match → `CheckViolation{Type: "admitted_variant", Variant: ..., Line: ..., Column: ..., Context: ...}`.

The `--strict` flag on `check` is distinct from `--strict` on `validate` — different commands, different `*cli.BoolFlag` declarations, different semantics. This ticket only handles the `check` side.

### Interaction with `missing` detection

An admitted variant in TGT does NOT satisfy the "preferred term present" check. If only admitted variants are found (no preferred), the concept still generates a `missing` violation. The admitted match generates its own warning (or violation under --strict) independently.

## TDD cycles

### Cycle 1 — Admitted as warning (non-strict)
RED: TGT has admitted variant "sephirah" instead of preferred "sefirah" → `missing` violation + `admitted_variant` warning.
GREEN: Detect admitted terms, emit warning.

### Cycle 2 — Admitted as violation (strict)
RED: Same scenario with `strict=true` → `missing` violation + `admitted_variant` violation (no warning).
GREEN: Switch warning → violation based on strict flag.

### Cycle 3 — Admitted alongside preferred
RED: TGT has both preferred "sefirah" and admitted "sephirah" → no `missing` violation, but still `admitted_variant` warning.
GREEN: Admitted detection is independent of missing detection.

## Acceptance

- `make build` passes
- Without `--strict`: admitted variants produce warnings, not violations
- With `--strict`: admitted variants produce `admitted_variant` violations
- Admitted variants do not satisfy the preferred-term presence check
- `ok` field in envelope: `true` when only warnings, `false` when violations present

## Notes

**2026-05-26T18:22:41Z**

Implemented admitted_variant detection in check algorithm. Changes: (1) Extended CheckWarning in output/types.go with Variant, Line, Column, Context fields (omitempty) for positional admitted_variant warnings. (2) Added admitted term tracking in check.go alongside existing forbidden hits. When strict=false, admitted target matches produce CheckWarning; when strict=true, they produce CheckViolation. (3) Admitted variants do NOT satisfy preferred-term presence — a missing violation still fires independently. (4) Three new tests: AdmittedTargetWarning, AdmittedTargetViolation_Strict, AdmittedAlongsidePreferred. (5) Regenerated schema full golden file since CheckWarning struct changed.
