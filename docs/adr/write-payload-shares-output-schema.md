# Cross-cutting — apply/write input reuses the output envelope schema

> **Status**: APPROVED. The `apply` and write-command JSON input payloads
> deliberately deserialize into the same `internal/output` types the binary
> emits. Do not split them into a separate `write`-owned input type.

## The decision

`internal/write` imports `internal/output` and parses input JSON into
`output.WriteResult` (and `WriteTerm`, `WriteCrossRef`, `WriteTermGroup`):

- `write.ParseJSONInput` → `*output.WriteResult` (write-command `--format json`
  stdin, e.g. `concept add`).
- `write.ParseApplyJSON` → `ApplyPayload{ Concepts []output.WriteResult }` →
  `WriteResultToConcept` (the `apply --file` path).

The same types are produced for emission by
`commands/write_helpers.go:buildWriteResult`. A future architecture review will
see the "display DTO doubling as input payload" and the `write → output` import
and propose a dedicated `write.Payload` input type. We considered it and
**rejected it.**

## Why not

- **Input == output is a documented contract.**
  [`cli-design.md` §"write commands"](../cli-design.md) states the stdin payload
  schema matches the JSON shape the binary emits, so a read-modify-write
  round-trip is straightforward. A single shared type guarantees that by
  construction; two types would not.
- **No real seam yet.** There is exactly one input shape and one output shape,
  and they are intentionally identical. One adapter is a hypothetical seam; two
  justify a real one. Nothing varies across this seam, so splitting is premature.
- **It fails the deletion test.** A `write.Payload` mirror would not remove the
  field set — it would duplicate it and add a parity test to keep input and
  output from drifting. Complexity moves and grows rather than vanishing.
- **The coupling is one-way and acyclic.** `output` never imports `write`; it is
  a leaf display package. The `write → output` arrow is mildly inverted in
  layering but causes no cycle and no real friction.
- **Envelope types are ADR-anchored in `output`.**
  [schema-source-of-truth](schema-source-of-truth.md) names
  `internal/output/types.go` as the canonical envelope contract and the surface
  the reflective `terminology schema` command walks. Relocating these types
  would fight that decision.

## Revisit when

The input schema needs to diverge from the emitted schema — e.g. accepting a
lenient superset on input while emitting a strict canonical shape. At that point
a real seam exists and a `write`-owned input type (with an explicit mapping to
`output.WriteResult`) becomes justified.
