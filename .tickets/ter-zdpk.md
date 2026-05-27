---
id: ter-zdpk
status: closed
deps: []
links: []
created: 2026-05-27T15:16:58Z
type: task
priority: 0
assignee: Andre Silva
parent: ter-nd3x
tags: [e9, hardening, security]
---
# E9.T3 — XML parser hardening (DOCTYPE policy, nesting depth)

Harden the TBX XML parser: add DOCTYPE policy enforcement (streaming pre-scan before xml.Decoder), nesting depth cap (256 levels), and xml.Decoder.Strict=true. Applies to all XML decode paths: linguist.Reader.Decode, detectDialect, firstElementName, extractListInner.

## Acceptance Criteria

- DOCTYPE policy: accept bare <!DOCTYPE tbx>; reject any DOCTYPE with internal subset (entity declarations) or SYSTEM/PUBLIC external IDs
- Streaming pre-scan runs before handing bytes to xml.Decoder — standard decoder doesn't surface DOCTYPE shape
- xml.Decoder.Strict = true on all decode paths
- Nesting depth cap: 256 levels; exceeding returns invalid_input with nesting_too_deep hint
- Depth counter wraps xml.Decoder token reads in linguist.Reader.Decode
- Unit tests for accepted DOCTYPE (bare), rejected DOCTYPE (entities, SYSTEM, PUBLIC), and nesting depth overflow
- Golden test for nesting_too_deep error envelope
- Existing tests continue to pass (TBX files with bare <!DOCTYPE tbx> still load)
- make build passes


## Notes

**2026-05-27T16:41:31Z**

Implemented XML parser hardening with three measures:

1. **DOCTYPE policy** (`tbx/harden.go`): `CheckDoctype(data)` scans raw bytes before XML decoding. Accepts bare `<!DOCTYPE tbx>`, rejects DOCTYPE with internal subset (`[` bracket), SYSTEM, or PUBLIC external IDs. Wired into both `tbx.Load()` (file path) and `linguist.Reader.Decode()` (io.Reader).

2. **Nesting depth cap** (`linguist/reader.go`): `decodeCtx` gained a `depth int` field and a `token()` method that wraps `xml.Decoder.Token()` with depth tracking. All token reads and `Skip()` calls go through `dc.token()` / `dc.skip()` respectively, enforcing a 256-level cap. Exceeding returns an error with 'nesting depth' message. `readCharData` and `readNoteText` were changed from `*xml.Decoder` to `*decodeCtx` parameter to participate in depth tracking.

3. **`xml.Decoder.Strict = true`** on all 4 decode paths: `linguist.Reader.Decode`, `detectDialect`, `firstElementName`, `extractListInner`.

New sentinels: `ErrDangerousDoctype` and `ErrNestingTooDeep` in `tbx/errors.go` (both code: invalid_input, exit 65). Schema golden files regenerated.

Unit tests in `tbx/harden_test.go` and `linguist/reader_test.go`. Golden CLI tests: `validate/doctype_entity`, `validate/nesting_too_deep`.

All existing tests pass. make build clean.
