---
id: ter-kqsv
status: closed
deps: []
links: []
created: 2026-05-27T12:54:13Z
type: chore
priority: 3
assignee: Andre Silva
parent: ter-jfqg
tags: [e8, chore, cleanup]
---
# E8.T8 — Remove under-construction scaffolding

All 12 commands are now implemented. Remove the `terr.UnderConstruction()` scaffolding that was introduced in E1 for stub commands. It is dead code.

## What to remove

| Item | File | Lines | Action |
|------|------|-------|--------|
| `UnderConstruction()` func | `src/internal/terr/terr.go` | 58–65 | Delete function |
| `TestUnderConstruction_NotRegistered` | `src/internal/terr/terr_test.go` | 95–102 | Delete test |
| `TestUnderConstruction` | `src/internal/terr/terr_test.go` | 119–149 | Delete test |
| Output error tests using `UnderConstruction` as fixture | `src/internal/output/errors_test.go` | 15, 47, 69–77 | Replace fixture with a real sentinel |
| Conformance test using `UnderConstruction` as fixture | `src/internal/output/conformance_test.go` | ~72 | Replace fixture with a real sentinel |
| Exit code 75 row | `docs/cli-design.md` | ~33 | Keep but mark as historical (scaffolding pattern, no longer emitted) |
| `under_construction` mentions | `docs/adr/error-handling.md` | 170, 194–196 | Remove references |

## What to keep (do NOT touch)

- `docs/specs/001-cli-surface-stub.md` — historical record of E1 design
- `qa/E1-manual-qa.md` — historical QA plan
- `docs/learnings.md` — historical learnings entries

## Acceptance criteria

- `terr.UnderConstruction()` function no longer exists in source
- No test references `UnderConstruction`
- Output tests still pass using a replacement error fixture
- Exit code 75 row in `cli-design.md` annotated as historical (scaffolding pattern, no longer emitted)
- ADR `error-handling.md` no longer references `under_construction`
- `make build` passes (fmt-check, vet, lint, test, compile)


## Notes

**2026-05-27T13:04:02Z**

Removed terr.UnderConstruction() function and its two tests from terr package. Replaced UnderConstruction fixtures in output/errors_test.go (3 usages) and output/conformance_test.go (1 usage) with terr.Newf()-based test fixtures. Marked exit code 75 as historical in cli-design.md. Removed under_construction row from error-handling.md exit code table and the urfave integration paragraph referencing the stub pattern. make build passes clean.
