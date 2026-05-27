---
id: ter-usuw
status: closed
deps: [ter-zwxt]
links: []
created: 2026-05-22T19:41:34Z
type: task
priority: 1
assignee: Andre Silva
parent: ter-qxrg
tags: [e1, task, foundation, output]
---
# E1.T2 — Foundation: internal/output (envelope renderer)

## Goal

Stand up `internal/output` — the single sink that converts a `terr.Coded` (or any `error`) into the documented JSON or text error envelope, plus the resolver for the process exit code. This is the only package that knows the envelope shape.

## Refs

- E1 spec: [docs/specs/001-cli-surface-stub.md](docs/specs/001-cli-surface-stub.md) §"Under-construction contract" + §"main.go shape"
- Error-handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md) §"Layer 3 — Envelope emission" + §"Format-specific error rendering"
- Schema-source-of-truth ADR: [docs/adr/schema-source-of-truth.md](docs/adr/schema-source-of-truth.md) §"Envelope \`schema_version\`"
- Testing ADR: [docs/adr/testing.md](docs/adr/testing.md) §"Envelope conformance"

## Depends on

- T1 (`internal/terr`).

## Files to create

- `src/internal/output/errors.go` — `EmitError`, `ExitCodeFor`
- `src/internal/output/version.go` — `const SchemaVersion = 1`
- `src/internal/output/errors_test.go` — unit tests for the renderer + exit resolver
- `src/internal/output/conformance_test.go` — helper invariant (Assert every envelope carries `schema_version` and exactly one of `{ok:true}` or `{ok:false, error:{…}}`)

## API

```go
package output

const SchemaVersion = 1

// EmitError writes the structured error envelope to w in either
// \"json\" or \"text\" format. Unknown formats fall back to JSON.
// If err satisfies terr.Coded, its code/message/hint populate the envelope.
// Otherwise the fallback envelope is {code: \"internal_error\", message: err.Error()}
// (no hint).
func EmitError(w io.Writer, format string, err error)

// ExitCodeFor returns err.ExitCode() if err satisfies terr.Coded;
// otherwise 1.
func ExitCodeFor(err error) int
```

### JSON envelope shape (`format == \"json\"\`)

```json
{
  "schema_version": 1,
  "ok": false,
  "error": {
    "code": "under_construction",
    "message": "terminology validate is not implemented yet",
    "hint": "track progress in .tickets/ or rebuild from a newer commit"
  }
}
```

- `hint` is **omitted** (not empty-string) when `Hint()` returns `\"\"\`. Use `json:\",omitempty\"` on the struct field.
- One JSON object per call, followed by a single `\n\`.
- Deterministic key order via struct definition (encoding/json preserves declaration order).

### Text envelope shape (`format == \"text\"\`)

```
✗ <message>
  hint: <hint>
```

- Single line for the message (prefixed with `✗ ` and a single space).
- Continuation line for the hint, prefixed with exactly **two spaces** then `hint: `. Omitted when Hint() is empty.
- Trailing newline.

## TDD plan

Tests live in `src/internal/output/errors_test.go` (renderer/resolver) and `src/internal/output/conformance_test.go` (invariant helper).

Vertical slices:

1. **RED** `TestEmitError_JSON_TerrCoded` — pass `terr.UnderConstruction(\"validate\")` and `format=\"json\"`; assert byte-exact output (use byte literal). **GREEN** implement `EmitError` JSON branch + struct types.
2. **RED** `TestEmitError_JSON_OmitsEmptyHint` — construct *terr.E via `terr.New(\"x\", 2, \"\", \"msg\")`; assert output has no `hint` key. **GREEN** ensure `omitempty` tag.
3. **RED** `TestEmitError_JSON_UnknownError` — pass a plain `errors.New(\"boom\")`; assert envelope has code=`internal_error`, ok=false, no hint key, message=\"boom\". **GREEN** add fallback branch.
4. **RED** `TestEmitError_Text_TerrCoded` — assert two-line output (✗ line + hint continuation). **GREEN** implement text branch.
5. **RED** `TestEmitError_Text_OmitsEmptyHint` — assert single line, no continuation. **GREEN** conditional hint emission.
6. **RED** `TestEmitError_FallbackFormat` — `format=\"yaml\"` (unknown) routes to JSON. **GREEN** default branch.
7. **RED** `TestExitCodeFor_Coded` — pass `terr.New(\"x\", 65, \"\", \"m\")`; assert 65. **GREEN** implement via errors.As.
8. **RED** `TestExitCodeFor_NonCoded` — plain error; assert 1. **GREEN** else branch.
9. **RED** `TestSchemaVersion_IsOne` — sanity: `SchemaVersion == 1`.

### Conformance helper (`conformance_test.go`)

```go
// AssertEnvelopeShape unmarshals raw into a generic map and asserts:
//   - top-level \"schema_version\" is an int (== SchemaVersion)
//   - top-level \"ok\" is a bool
//   - if ok==false, top-level \"error\" exists with string \"code\" and \"message\"
//   - if ok==true, no top-level \"error\" key
// Used by golden runners in T5+ to enforce the envelope invariant
// without restating it in every golden file.
func AssertEnvelopeShape(t *testing.T, raw []byte)
```

One test (`TestAssertEnvelopeShape`) covers the success path against an EmitError output and the failure path against malformed JSON.

## Acceptance

- `make build` clean from project root
- `cd src && go test ./internal/output/...` passes
- No new dependencies (stdlib only: `encoding/json`, `fmt`, `io`)

## Out of scope

- Success-path envelopes (T5 introduces the first stub envelope; later epics introduce real result envelopes).
- The reflective `terminology schema` walker.
- `--fields` projection (lives in `internal/output/fields.go` — landed by E4).


## Notes

**2026-05-22T20:42:01Z**

Implemented internal/output with three files:
- version.go: SchemaVersion = 1
- errors.go: EmitError(w, format, err) and ExitCodeFor(err) — JSON/text envelope rendering with terr.Coded extraction and internal_error fallback
- errors_test.go + conformance_test.go: 14 tests covering JSON/text output, omitempty hint, unknown format fallback, exit code resolution, and envelope shape validation helper (AssertEnvelopeShape)

Design notes:
- EmitError takes format as a parameter (not from cli.Context) to keep the package dependency-free from urfave
- AssertEnvelopeShape uses t.Errorf (not Fatalf) so it can be tested with bare *testing.T for negative cases
- Write return values use _, _ = pattern to satisfy errcheck linter (stderr writes at exit boundary have no recovery path)
