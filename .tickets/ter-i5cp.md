---
id: ter-i5cp
status: closed
deps: [ter-c4ra]
links: []
created: 2026-05-26T13:48:49Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-2sqs
tags: [e5, task, match, policy]
---
# E5.T1 — Match policy types

## Goal

Define the match policy types that control normalization behavior for the matcher. The `Policy` struct encodes per-axis decisions (case-fold, diacritics, niqqud, normalization form) and a per-language override table maps language tags to their policies.

## Refs

- E5 spec: [docs/specs/005-matcher.md](docs/specs/005-matcher.md) §"v1 match policy" and §"Per-language overrides"
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md)

## Files to create / modify

- `src/internal/match/policy.go` — Policy struct, Form enum, baseline default, per-language overrides
- `src/internal/match/policy_test.go` — tests

## Behavior

```go
package match

type Form int

const (
    NFC Form = iota
    NFKD
)

type Policy struct {
    CaseFold       bool
    FoldDiacritics bool
    StripNiqqud    bool
    Normalize      Form
}

var Baseline = Policy{CaseFold: true, FoldDiacritics: false, StripNiqqud: false, Normalize: NFC}

func PolicyFor(lang string) Policy
```

### v1 defaults

| Axis          | Default                         |
|---------------|---------------------------------|
| Case          | Unicode default case-fold       |
| Diacritics    | Strict (no folding)             |
| Niqqud        | Strip from both sides (Hebrew)  |
| Normalization | NFC                             |

### Per-language overrides

```go
var byLanguage = map[string]Policy{
    "he": {CaseFold: true, FoldDiacritics: false, StripNiqqud: true, Normalize: NFC},
}
```

`PolicyFor(lang)` returns the language-specific policy if one exists, otherwise the baseline.

## TDD cycles

### Cycle 1 — Baseline policy
RED: `PolicyFor("")` returns baseline with `CaseFold: true`, `StripNiqqud: false`.
GREEN: Return baseline default.

### Cycle 2 — Hebrew override
RED: `PolicyFor("he")` returns policy with `StripNiqqud: true`.
GREEN: Lookup in byLanguage map.

### Cycle 3 — Unknown language falls back to baseline
RED: `PolicyFor("fr")` returns baseline.
GREEN: Map miss returns baseline.

### Cycle 4 — Form enum
RED: Assert `NFC` and `NFKD` are distinct values.
GREEN: Const block definition.

## Acceptance

- `make build` passes
- Policy struct exported with all four axes
- `PolicyFor` returns language-specific overrides or baseline
- Hebrew override strips niqqud


## Notes

**2026-05-26T14:09:52Z**

Implemented match.Policy struct, Form enum (NFC/NFKD), Baseline default, per-language override map with Hebrew entry (StripNiqqud: true), and PolicyFor() lookup function. Created src/internal/match/policy.go and policy_test.go. All tests pass, make build clean.
