---
id: ter-ix1i
status: closed
deps: []
links: [ter-5sps]
created: 2026-05-26T03:19:29Z
type: bug
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, bug, schema]
---
# E4.BUG — exit_codes missing from per-command schema view

## Summary

`terminology schema --command validate` does not include an `exit_codes`
array in the per-command view. The E4 QA plan (TC-SCH-ERR-002) expects it.

## Expected

```json
{
  "name": "validate",
  "flags": [...],
  "envelope": {...},
  "exit_codes": [0, 1, 2, 65]
}
```

## Actual

`exit_codes` key is absent from the per-command schema output.

## Steps to reproduce

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"
${TT} schema --command validate | jq 'has("exit_codes")'
# returns false
```

## Refs

- QA report: [qa/E4-manual-qa.report.md](qa/E4-manual-qa.report.md) Finding 3
- QA plan: [qa/E4-manual-qa.md](qa/E4-manual-qa.md) TC-SCH-ERR-002
- Schema walker: `src/internal/schema/schema.go`

## Fix

Add per-command exit code enumeration to the schema command walker. Each
command should declare its possible exit codes so the reflective walker
can include them in the single-command view.


## Notes

**2026-05-26T03:47:09Z**

Added per-command exit_codes to the schema filtered view. Implementation: (1) Added exitCodes registry in output/registry.go (parallel to envelopes registry — RegisterExitCodes, ExitCodesFor, AllExitCodes). (2) Registered exit codes for all commands in output/types.go init(). (3) Added ExitCodes []int field to schemaFilteredEnvelope and wired it in schemaFiltered. (4) Updated golden tests. Exit codes per command: validate [0,1,2,3,65], lookup [0,1,2,3,65], extract [0,1,2,3], scan [0,2,3,65], check [0,1,2,3,65], schema [0,2], write commands [0,2,3,65].
