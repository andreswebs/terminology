---
id: ter-pe7f
status: closed
deps: []
links: []
created: 2026-05-24T01:05:24Z
type: task
priority: 3
assignee: Andre Silva
parent: ter-8gyy
tags: [e7, task, flags, retroactive]
---
# E7.R1 — Retroactive: pickFlag wiring for picklist-validated CLI flags

## Goal

Retroactive ticket documenting the `pickFlag()` helper and picklist-validated CLI flags that were implemented alongside E1 command stubs. These flags consume `tbx.picklist.go` (E3.T1) to build urfave validators that reject invalid user input at the command line.

This ticket tracks the existing implementation for spec alignment verification.

## Refs

- E7 spec: [docs/specs/007-write-commands.md](docs/specs/007-write-commands.md) (flag definitions)
- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Picklist values" (two-layer consumption)
- Picklist source: `src/internal/tbx/picklist.go`

## Files (existing)

- `src/internal/app/commands/flags.go` — `pickFlag()` helper, `writeFlags()`, `langFlag()`, `termFlag()`, `dryRunFlag()`, `authorFlag()`, `transactionFlag()`, `readFieldsFlag()`
- `src/internal/app/commands/flags_test.go` — tests for all flag helpers
- `src/internal/app/commands/concept_add.go` — uses `pickFlag` for `--status`, `--part-of-speech`, `--register`, `--grammatical-gender`
- `src/internal/app/commands/term_add.go` — uses same picklist flags

## Current implementation

### pickFlag helper

```go
func pickFlag(name, alias, usage string, valuesFn func() []string) urfcli.Flag {
    allowed := valuesFn()
    set := make(map[string]bool, len(allowed))
    for _, v := range allowed {
        set[v] = true
    }
    f := &urfcli.StringFlag{
        Name:  name,
        Usage: usage,
        Validator: func(val string) error {
            if !set[val] {
                return urfcli.Exit("invalid value "+val+"; accepted: "+strings.Join(allowed, ", "), 2)
            }
            return nil
        },
    }
    if alias != "" {
        f.Aliases = []string{alias}
    }
    return f
}
```

- Takes a `func() []string` callback (the picklist function) — NOT a `[]string` — so the set is built at flag construction time.
- Validator rejects invalid values with exit code 2 and a helpful error listing accepted values.
- Supports optional alias.

### Spec alignment check

| Concern | Spec | Current | Status |
|---------|------|---------|--------|
| pickFlag uses picklist function | ✓ | ✓ `valuesFn func() []string` | Aligned |
| Invalid value rejected | ✓ | ✓ `urfcli.Exit(...)` exit 2 | Aligned |
| Error lists accepted values | ✓ | ✓ `strings.Join(allowed, ", ")` | Aligned |
| `--status` on concept add | ✓ | ✓ `pickFlag("status", "s", ..., tbx.AdminStatus)` | Aligned |
| `--part-of-speech` on concept add | ✓ | ✓ | Aligned |
| `--register` on concept add | ✓ | ✓ | Aligned |
| `--grammatical-gender` on concept add | ✓ | ✓ | Aligned |
| writeFlags (dry-run, transaction, author) | ✓ | ✓ | Aligned |
| author env var `TERMINOLOGY_AUTHOR` | ✓ | ✓ `Sources: urfcli.EnvVars("TERMINOLOGY_AUTHOR")` | Aligned |

### Tests

- `TestPickFlag` — valid value accepted, invalid value rejected with "invalid value" message
- `TestPickFlagNoAlias` — empty alias not included in names
- `TestWriteFlags` — dry-run, transaction, author flags with correct types and aliases
- `TestReadFieldsFlag` — fields flag with alias F
- `TestLangFlag` — required/optional variants
- `TestTermFlag` — required/optional variants
- `TestDryRunFlag` — alias -n
- `TestAuthorFlag` — alias -a, env var source

All tests pass. Implementation fully aligns with the spec.

## Status

This ticket can be closed immediately — the implementation is complete and aligned with the spec. No changes needed.

## Out of scope

- Picklist values themselves (E3.T1)
- Write command actions (E7 — currently stubs returning exit 75)
- Format flag validator (defined inline in `root.go`, not using `pickFlag`)

## Acceptance

- `make build` passes
- `pickFlag()` builds validators from picklist functions
- Invalid values produce exit 2 with accepted-values list
- All write commands use `pickFlag` for picklist-typed flags

