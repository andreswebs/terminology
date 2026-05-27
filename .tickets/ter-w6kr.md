---
id: ter-w6kr
status: closed
deps: [ter-cplu, ter-bedf]
links: [ter-97c1]
created: 2026-05-24T01:05:21Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-told
tags: [e3, task, lineindex, reader]
---
# E3.T11 — LineIndex wiring into reader

## Goal

Wire the `lineindex.LineIndex` package (from E2.T5) into the linguist reader so that `Warning.Line` and `Warning.Col` fields are populated with the source position where each issue was detected.

The spec describes a streaming newline-index wrapper: the reader wraps the input `io.Reader` with `LineIndex.Wrap(r)`, which counts newline byte offsets as bytes pass through. When a warning needs to be emitted, the reader calls `LineIndex.Position(decoder.InputOffset())` to get `(line, col)`.

## Refs

- E3 spec: [docs/specs/003-validate-command.md](docs/specs/003-validate-command.md) §"Line/column tracking"
- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Architecture" (lineindex.go)
- LineIndex package ticket: E2.T5 (ter-cplu)

## Dependencies

- **E2.T5 (ter-cplu)** — the `internal/tbx/lineindex` package must exist before this ticket can be implemented.

## Files to modify

- `src/internal/tbx/linguist/reader.go` — wrap input with LineIndex, populate Warning.Line/Col
- `src/internal/tbx/linguist/reader_test.go` — test that warnings carry line/col

## LineIndex API (from E2.T5)

```go
package lineindex

type LineIndex struct {
    newlines []int64
}

func (li *LineIndex) Wrap(r io.Reader) io.Reader
func (li *LineIndex) Position(offset int64) (line, col int)
```

- `Wrap` returns an `io.Reader` that passes bytes through while recording newline offsets.
- `Position` does a binary search on the newline offsets to map a byte offset to `(line, col)` where line is 1-based and col is byte offset from the start of the line.

## Implementation

In `LinguistReader.Decode()`:

```go
func (lr *LinguistReader) Decode(r io.Reader) (*tbx.Glossary, []tbx.Warning, error) {
    li := &lineindex.LineIndex{}
    wrapped := li.Wrap(r)
    dec := xml.NewDecoder(wrapped)

    // ... existing decode logic ...
    // At each warning emission point, capture position:
    // line, col := li.Position(dec.InputOffset())
    // w.Line = line
    // w.Col = col
}
```

Key implementation notes:
- **`dec.InputOffset()`** returns the byte offset of the decoder's current position in the input stream. Call it BEFORE decoding the element that triggers the warning (or immediately at the start of element processing).
- **Offsets reference already-streamed bytes** — the LineIndex is always up-to-date at lookup time because the decoder has already streamed past the relevant bytes.
- **All warning emission points must be updated** — duplicate_id, invalid_lang_tag, missing_term (in Glossary.Validate), unknown_element, invalid_picklist, legacy_form_normalized (in reader). For reader-emitted warnings, Position is called directly. For Validate-emitted warnings, the reader would need to store element offsets in the domain model or a side channel — this may be deferred.

## TDD cycles

### Cycle 1 — LineIndex wrapping
RED: Decode `minimal-dct.tbx` with LineIndex wrapping. Assert no errors (wrapping doesn't break decoding).
GREEN: Add LineIndex wrapping in `Decode()`.

### Cycle 2 — Reader warnings carry line/col
RED: Decode a fixture that triggers an `invalid_picklist` warning. Assert `Warning.Line > 0` and `Warning.Col > 0`.
GREEN: Call `li.Position(dec.InputOffset())` at warning emission points.

### Cycle 3 — Line/col accuracy
RED: Decode a fixture where the warning-triggering element is on a known line (e.g. line 12). Assert `Warning.Line == 12`.
GREEN: Verify InputOffset is captured at the right moment (before or after element decode).

### Cycle 4 — Multiple warnings have distinct positions
RED: Fixture with 2 warnings on different lines. Assert they have different `Line` values.
GREEN: Already passing if each warning captures its own offset.

## Deviation note

The current implementation does NOT populate `Warning.Line` or `Warning.Col` — they are always zero. The `Warning` struct (defined in E2) already has `Line int` and `Col int` fields, but no code sets them. The `lineindex` package (E2.T5) does not exist yet.

This ticket depends on E2.T5 being completed first. Once LineIndex exists, the reader wraps its input and populates positions at warning emission points.

## Out of scope

- The lineindex package itself (E2.T5)
- Line/col for `Glossary.Validate()` warnings (these operate on the in-memory model, not the XML stream — line/col would require storing offsets during decode, which is a larger change)
- Line/col for error messages (separate concern)

## Acceptance

- `make build` passes
- Reader wraps input with `LineIndex.Wrap(r)`
- Reader-emitted warnings (unknown_element, invalid_picklist, legacy_form_normalized) carry line/col
- Clean files produce no warnings (line/col is irrelevant)
- Decode behavior is unchanged (wrapping is transparent)


## Notes

**2026-05-25T18:50:02Z**

The lineindex package exists (src/internal/tbx/lineindex/lineindex.go with Wrap and Position methods). However, the linguist reader does NOT use it — no import of lineindex, no call to Wrap, no call to Position, no InputOffset usage. Warning.Line and Warning.Col remain zero on all reader-emitted warnings. This ticket is not started. Once implemented: wrap the input io.Reader with lineindex.Wrap in Decode(), capture dec.InputOffset() at each warning emission point, and call li.Position() to populate Line/Col.

**2026-05-25T18:59:55Z**

Implemented lineindex wiring into linguist reader. Key changes:

1. Added decodeCtx struct bundling *xml.Decoder + *lineindex.Index with pos() method
2. Decode() now buffers input via io.ReadAll, builds lineindex.Index, creates xml.Decoder from the same buffer
3. All decode functions (decodeConceptEntry, decodeConceptChild, decodeLangSec, decodeTermSec, decodeAdminGrp, decodeTransacGrp, and all DCT/DCA variants) changed from *xml.Decoder to *decodeCtx
4. readCharData and readNoteText kept with *xml.Decoder (they don't emit warnings)
5. For unknown_element warnings: dc.pos() called at emission point (after StartElement token, offset is right after the opening tag)
6. For invalid_picklist warnings: position captured BEFORE readCharData (since readCharData advances the offset past the closing tag)
7. Five new tests: unknown_element line/col, invalid_picklist line/col, line accuracy on known fixture, multiple warnings with distinct positions, DCA style warning positions

The lineindex.New() API reads all data (no streaming Wrap method despite the ticket spec mentioning one). The approach buffers input once with io.ReadAll, then creates both the Index and the Decoder from bytes.NewReader — the double-read through lineindex.New is negligible for TBX file sizes.
