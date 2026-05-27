---
id: ter-rb0i
status: closed
deps: []
links: [ter-qxrg, ter-smt4]
created: 2026-05-25T14:19:34Z
type: bug
priority: 0
assignee: Andre Silva
tags: [e1, bug, blocker, surface-drift]
---
# E1.BUG — Positional argument bounds (Min/Max) not enforced

Commands that declare Min/Max on urfave Arguments do not enforce positional
bounds. Missing or excess positionals pass through to the stub (exit 75)
instead of being rejected with exit 2.

## Affected commands

- lookup: Min:1, Max:1 — neither enforced
- scan: Min:1, Max:1 — neither enforced
- check: Max:2 only — Min:2 IS enforced (exits 1, see ter-smt4), but Max not
- concept update: Min:1 — not enforced
- concept remove: Min:1 — not enforced
- term add: Min:1 — not enforced
- term deprecate: Min:1 — not enforced

## Reproduction

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"

$TT lookup 2>/dev/null; echo "exit=$?"                         # want 2, got 75
$TT lookup tzimtzum extra 2>/dev/null; echo "exit=$?"          # want 2, got 75
$TT scan 2>/dev/null; echo "exit=$?"                           # want 2, got 75
$TT scan a.md b.md 2>/dev/null; echo "exit=$?"                 # want 2, got 75
$TT check a.md b.md c.md 2>/dev/null; echo "exit=$?"           # want 2, got 75
$TT concept update --merge 2>/dev/null; echo "exit=$?"         # want 2, got 75
$TT concept remove 2>/dev/null; echo "exit=$?"                 # want 2, got 75
$TT term add --lang es --term tzim 2>/dev/null; echo "exit=$?" # want 2, got 75
$TT term deprecate --lang es --term tzim 2>/dev/null; echo "exit=$?" # want 2, got 75
```

## Affected test cases

TC-LOOKUP-002, TC-LOOKUP-003, TC-SCAN-002, TC-SCAN-003, TC-CHECK-004,
TC-CONCEPT-UPDATE-003, TC-CONCEPT-REMOVE-002, TC-TERM-ADD-003, TC-TERM-DEP-003

## Fix direction

Either urfave's Arguments Min/Max fields are not being set in the command
definitions, or urfave v3 does not enforce them automatically and explicit
before-action validation is needed. Verify the command definitions first;
if Min/Max is declared but not enforced, add a shared before-action hook
that checks len(args) against the declared bounds and returns a usage error.


## Notes

**2026-05-25T14:35:28Z**

Fix: added argBounds(min, max) Before hook in commands/argcheck.go. urfave v3 StringArg.Parse() silently accepts 0 args and StringArgs doesn't reject excess beyond Max. The hook returns terr.New sentinels with codes 'missing_argument' / 'excess_arguments' (exit 2). For concept update, chainBefore() composes argBounds with requireMergeXorReplace so arg check runs first. All 9 affected commands now validated. Table-driven tests in TestArgBounds_ExitCode2 and TestArgBounds_ErrorCodes.
