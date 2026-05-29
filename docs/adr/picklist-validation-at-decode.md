# Cross-cutting — picklist validation happens at decode, not in Validate

> **Status**: APPROVED. `invalid_picklist` warnings are emitted by the
> TBX-Linguist decoder, not by `Glossary.Validate`. Do not move picklist
> checking into `internal/tbx/validate.go`.

## The decision

Picklist conformance (a term's `administrativeStatus`, `partOfSpeech`,
`register`, etc. being a member of the accepted set in
[`tbx/picklist.go`](../../src/internal/tbx/picklist.go)) is checked at
**decode time** by `checkPicklist` in
[`linguist/reader.go`](../../src/internal/tbx/linguist/reader.go), which emits
`invalid_picklist` warnings as part of the `(glossary, []Warning, error)` decode
result. `Glossary.Validate` performs only **structural** checks (duplicate ids,
unresolved cross-references, missing terms, malformed language tags) and
deliberately does **not** re-check picklists.

A future architecture review will see that `validate.go` does not validate
picklists and propose extracting `checkPicklist` into `tbx` so `Validate` can
reuse it. We considered it and **rejected it.**

## Why not

- **It would double-report on the file path.** `terminology validate` (and the
  write/apply validation gates) run `Validate` on an *already-decoded*
  glossary. Decode has already emitted `invalid_picklist` for every offending
  value, so having `Validate` re-check would surface each violation twice.
- **The only genuine gap is narrow and is a behaviour question, not a refactor.**
  Glossaries built in memory without decoding — the `apply` JSON payload path
  (`WriteResultToConcept`) — bypass decode-time picklist checks. Write-command
  flags are already constrained by `pickFlag` at the CLI layer. Whether `apply`
  should reject or warn on out-of-set picklist values is a deliberate feature
  decision, to be made on its own merits, not smuggled in via a dedup.
- **No vocabulary is duplicated by keeping it at decode.** The accepted sets live
  once in `tbx/picklist.go`; the decoder imports them. Validation *placement* is
  not vocabulary duplication.

## Related

The canonical `Status` ↔ string mapping is owned solely by `tbx`
(`ParseStatus`, `Status.String()`); the linguist decoder/encoder call those
rather than carrying their own copies. The legacy-form helpers
(`isLegacyStatus`, `normalizeRegister`, `isLegacyRegister`) remain in
`linguist` because legacy spellings are a TBX-Linguist dialect concern with no
`tbx` equivalent.

## Revisit when

`apply` (or another non-decode construction path) must enforce picklist
membership. At that point add the check at that path's validation gate — still
not in the shared `Validate`, to keep the file path single-reported.
