# Cross-cutting — Term-field vocabulary stays in parallel switches

> **Status**: APPROVED. The TBX-Linguist term/concept field vocabulary is
> deliberately restated across the reader, the dialect key maps, and the
> writer. Do not unify them behind a single field table.

## The decision

The set of TBX-Linguist data-category fields (e.g. `administrativeStatus`,
`register`, `crossReference`) appears in three places in
[`internal/tbx/linguist`](../../src/internal/tbx/linguist):

- `dialect.go` — `dialectDCT` / `dialectDCA` key closures map an element to a
  semantic field key (the **read** direction, per dialect).
- `reader.go` — `decodeTermFields` / `decodeConceptFields` switch on that key
  to decode, with **heterogeneous** per-field behaviour (picklist validation,
  legacy-form normalization, line/col warnings, scalar vs slice vs attribute
  vs `NoteText`).
- `writer.go` — `writeTermSec` / `writeConceptEntry` emit each field (the
  **write** direction), also heterogeneous (scalar guard, slice loop,
  `adminGrp` grouping, self-closing `target=` attributes), and **DCT-only** by
  the [determinism ADR](determinism.md) (output is always canonical DCT).

A periodic architecture review will see these three parallel blocks and
propose collapsing them into one `key → {namespace, localName, accessors,
emit}` table. We considered it and **rejected it.**

## Why not

- **The drift it would prevent is already caught.** The round-trip suite
  (`TestRoundTrip_Canonical`, `TestDCAtoCanonicalDCT`) locks the read↔write
  element spelling byte-for-byte; `TestRoundTrip_CoversEveryTermField`
  asserts every `tbx.Term` field is exercised by a canonical fixture, so a new
  field cannot be added without round-trip coverage forcing all three sites
  into agreement.
- **It fails the deletion test.** The per-field bodies are genuinely
  heterogeneous on both the read and write sides. A table would not delete that
  complexity — it would relocate it into per-field closures, trading readable
  switches for an indirection with no net leverage.
- **It only shares half.** The writer emits DCT exclusively, so the DCA spelling
  lives in `dialect.go` alone; a "shared" table would unify only the DCT half.
- **The writer cannot be dialect-parameterized.** Canonical-DCT output is
  mandated by the [determinism ADR](determinism.md), so the writer's spelling is
  not a free variable that a shared table could own.

## The guard instead

Coverage — not deduplication — is the real risk, and it is closed by the
reflection guard `TestRoundTrip_CoversEveryTermField` in
[`roundtrip_test.go`](../../src/internal/tbx/linguist/roundtrip_test.go). Adding
a field to `tbx.Term` fails that test until a fixture exercises it, which then
engages the byte-exact round-trip tests to lock its reader/dialect/writer
spelling.
