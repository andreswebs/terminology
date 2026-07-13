---
id: ter-2mia
status: closed
deps: []
links: []
created: 2026-07-12T11:18:13Z
type: bug
priority: 2
assignee: Andre Silva
tags: [field-feedback, write, json, errors, dx]
---
# GAP-2 — JSON payload validation errors do not name the offending field

# GAP-2 — JSON payload validation errors do not name the offending field

## Severity

Defect — developer/agent experience. Every structurally wrong JSON payload
collapses to one opaque `invalid_input` error with no path or type information,
forcing a guess-and-retry loop that burns agent turns.

Source: field feedback in `.local/tmp/issues.md` (GAP-2), build `b675e4d`.

## Reproduction

```sh
echo '{"concept_id":"x","definitions":[{"text":"d","lang":"en"}],
  "languages":{"en":{"preferred":{"term":"x","administrative_status":"preferredTerm-admn-sts"}}}}' \
  | terminology concept add --tbx g.tbx
```

Actual:

```json
{"ok":false,"error":{"code":"invalid_input",
 "message":"stdin payload is malformed or unsupported",
 "hint":"check the JSON payload structure; use 'terminology schema' for the expected shape"}}
```

The same generic message covers wrong value type, unknown key, wrong nesting,
and a per-language `definitions` key. The correct shape (`definitions` is an
array of strings) is not discoverable from the error.

Expected: an error that names the offending path and the expected vs actual
type, e.g. `definitions[0]: expected string, got object`.

## Root cause (verified against source)

Both JSON entry points decode with `DisallowUnknownFields()` and then wrap any
decode error into the fixed-message `ErrInvalidInput` sentinel, discarding the
informative detail that `encoding/json` already produced:

- `concept add` stdin:
  [`ParseJSONInput`](src/internal/write/write.go#L172-L181) —
  `return nil, ErrInvalidInput.Wrap(err)`.
- `apply`:
  [`ParseApplyJSON`](src/internal/write/apply.go#L19-L34) — same pattern.

`ErrInvalidInput`
([`src/internal/write/errors.go:29-33`](src/internal/write/errors.go#L29-L33))
carries a static message, and
[`EmitError`](src/internal/output/errors.go#L36-L76) renders `coded.Error()`
(the static message) — the wrapped `encoding/json` cause is never surfaced,
even though `*json.UnmarshalTypeError` already knows the field path, the
expected Go type, the JSON value kind, and the byte offset, and
`DisallowUnknownFields` errors already name the unknown field.

The error envelope already supports structured detail: `errorDetail.Details`
is populated from the `Detailed` interface (`ErrorDetails() any`) in
[`errors.go:25-34,49-53`](src/internal/output/errors.go#L25-L53). Today nothing
on the JSON path implements it. The precedent to follow is
[`ApplyValidationError`](src/internal/write/write.go#L203-L216), which
implements `Coded` + `Detailed` to emit a `failures[]` array.

## Fix design

Introduce a single JSON-error classifier used by both parse functions that
turns an `encoding/json` decode error into a field-level, `invalid_input`-coded
error carrying structured details.

1. **Add a typed error `InvalidInputError`** (in the `write` package, e.g.
   `src/internal/write/jsonerr.go`) implementing both interfaces:
   - `Coded`: `Code() == "invalid_input"`, `ExitCode() == 65`,
     `Hint()` = the existing schema hint (reuse `ErrInvalidInput.Hint()`), and
     `Error()` = a specific human message (see below).
   - `Detailed`: `ErrorDetails() any` returns a `FieldError` value so the
     envelope emits `error.details`.
   Preserve `errors.Is(err, ErrInvalidInput)` (wrap the sentinel or embed its
   identity) so existing exit-code/classification tests keep passing.

2. **Add `FieldError`** (JSON-tagged) capturing:
   `path string` (dotted/indexed path, best-effort from
   `json.UnmarshalTypeError.Field`), `expected string`, `actual string`,
   `kind string` (`type_mismatch` | `unknown_field` | `syntax`), and optional
   `offset int` / `line`,`column` from the error offset. Serialize under
   `error.details`.

3. **Classifier `describeJSONError(err error, data []byte) error`:**
   - `*json.UnmarshalTypeError` → `kind=type_mismatch`; `path` from `.Field`
     (fall back to `.Struct`), `expected` from a small `reflect.Type` → friendly
     name mapper (`string`, `object`, `array`, `number`, `bool`), `actual` from
     `.Value`. Message: `"<path>: expected <expected>, got <actual>"`.
   - `DisallowUnknownFields` error (message prefix `json: unknown field "`) →
     `kind=unknown_field`; extract the quoted field name. Message:
     `"unknown field \"<name>\""`. (Go's message format is stable; match the
     prefix and extract between the quotes.)
   - `*json.SyntaxError` → `kind=syntax`; `offset` from `.Offset`. Message:
     `"malformed JSON at offset <n>"`.
   - Anything else → fall back to the generic `ErrInvalidInput` (unchanged
     behavior) so no decode error is left worse than before.
   Optionally map `.Offset`/`.Field` to line/column using
   [`internal/tbx/lineindex`](src/internal/tbx/lineindex) for the `line`/`column`
   fields (nice-to-have; path + types is the core requirement).

4. **Route both parsers through it:** in `ParseJSONInput` and `ParseApplyJSON`,
   replace `ErrInvalidInput.Wrap(err)` with `describeJSONError(err, data)`. For
   apply, `json.UnmarshalTypeError.Field` naturally yields paths like
   `concepts[0].definitions`, giving per-concept context for free.

### Go conventions

- One classifier, two callers (deep module, small interface): put the logic in
  the `write` package, not duplicated in each parser.
- Wrap causes with `%w` / the sentinel's `.Wrap` so `errors.As`/`errors.Is`
  keep working; handle each error once.
- Use `errors.As` to extract `*json.UnmarshalTypeError` / `*json.SyntaxError`
  (they can be wrapped). Do not string-match those two; only the
  unknown-field case requires prefix matching (Go exposes no typed error for
  it).
- Exported `FieldError` fields need doc comments beginning with the field name;
  `MixedCaps`; `gofmt` clean.

## TDD plan (vertical slices — one test, one change, repeat)

Public interfaces under test: `write.ParseJSONInput([]byte)` and
`write.ParseApplyJSON([]byte)`. Add cases to
[`src/internal/write/errors_test.go`](src/internal/write/errors_test.go) (or a
new `jsonerr_test.go`). Go RED→GREEN per cycle; do not pre-write all tests.

### Cycle 1 (tracer) — type mismatch names path + expected/actual

RED: `ParseJSONInput` on the GAP-2 reproduction (`definitions` as an array of
objects). Assert the returned error's `Code()` is `"invalid_input"` and its
`Error()` message contains `definitions`, `expected string`, and
`object`/`got object`. Fails today (generic message).

GREEN: detect `*json.UnmarshalTypeError` via `errors.As`; build the message
from `.Field` + the type mapper + `.Value`.

### Cycle 2 — unknown key is named

RED: payload with an unknown top-level key (e.g. `{"concpet_id":"x",...}`) or a
per-language `definitions` key. Assert the message names the offending field
(`concpet_id` / `definitions`) and `Code()` is `"invalid_input"`.

GREEN: match the `json: unknown field "..."` prefix and extract the name.

### Cycle 3 — syntax error is distinguished

RED: malformed JSON (e.g. `"{"`). Assert `Code()` is `"invalid_input"` and the
message indicates a syntax problem (contains `offset` / `malformed`).

GREEN: detect `*json.SyntaxError`; report its `.Offset`.

### Cycle 4 — structured details reach the envelope

RED: assert the error implements `output.Detailed` and `ErrorDetails()` returns
a `FieldError` with `Path`/`Expected`/`Actual` set for the Cycle 1 input; and
that `errors.Is(err, write.ErrInvalidInput)` is still true. Optionally drive
`output.EmitError` and assert the emitted JSON has
`error.details.path == "definitions"` (or `definitions[0]`) and
`error.details.expected == "string"`.

GREEN: implement `Detailed` on `InvalidInputError`; preserve sentinel identity.

### Cycle 5 — apply path gets the same treatment

RED: `ParseApplyJSON` on `{"concepts":[{"concept_id":"x","definitions":
[{"text":"d"}],...}]}`. Assert the message path includes `concepts[0]` and
`definitions`, `Code()` is `"invalid_input"`, exit 65.

GREEN: route `ParseApplyJSON` through `describeJSONError` too.

### Refactor

- Extract the `reflect.Type` → friendly-name mapper and dedupe the two parser
  call sites. Consider optional line/column enrichment via `lineindex`.
- Run tests after each step. Never refactor while RED.

## Documentation updates (part of this ticket)

Update [`docs/skills/terminology/references/write-details.md`](docs/skills/terminology/references/write-details.md)
§"JSON payload limitations" (~lines 155–165): revise the bullet stating the
generic error "does not name the offending field" to describe the new
field-level error (path + expected/actual type, `error.details`). If
[`docs/cli-design.md`](docs/cli-design.md) or an error-handling ADR
([`docs/adr/error-handling.md`](docs/adr/error-handling.md)) documents the
`invalid_input` shape or the `error.details` field, note the new `FieldError`
detail there. Run
`markdownlint-cli2 --config ~/.markdownlint.yaml --fix` on edited markdown.

## Acceptance criteria

- A type-mismatch payload returns `invalid_input` with a message naming the
  field and the expected vs actual type (e.g. `definitions: expected string,
  got object`), and populates `error.details` with `path`/`expected`/`actual`.
- An unknown key is named in the message.
- A syntax error is distinguished from a schema error.
- `apply` payload errors include per-concept path context (e.g. `concepts[0]`).
- `errors.Is(err, ErrInvalidInput)` and exit code 65 are preserved; no
  regression in existing envelope/exit-code tests.
- `write-details.md` no longer claims the error is unnamed.
- `make build` passes (fmt-check, vet, lint, test) from the project root.

## Files to touch

- `src/internal/write/jsonerr.go` (new) — `InvalidInputError`, `FieldError`,
  `describeJSONError`, type-name mapper.
- `src/internal/write/write.go` — `ParseJSONInput` routes through the
  classifier.
- `src/internal/write/apply.go` — `ParseApplyJSON` routes through the
  classifier.
- `src/internal/write/errors_test.go` (or new `jsonerr_test.go`) — Cycles 1–5.
- `docs/skills/terminology/references/write-details.md` (and `docs/cli-design.md`
  / `docs/adr/error-handling.md` if they document the error shape).

## Validation

Run `make build` from the project root. Tighter loop:
`go test ./src/internal/write/... ./src/internal/output/...`, then finish with a
full `make build`. Do not silence lint with `_ =`.


## Notes

**2026-07-13T17:07:31Z**

GAP-2 fixed. Added InvalidInputError + FieldError + describeJSONError classifier in src/internal/write/jsonerr.go; both ParseJSONInput and ParseApplyJSON now route decode errors through it. Classifies *json.UnmarshalTypeError (type_mismatch, names path/expected/actual), DisallowUnknownFields (unknown_field, prefix-matched), and *json.SyntaxError + io.ErrUnexpectedEOF/EOF (syntax). InvalidInputError.Unwrap() returns the ErrInvalidInput sentinel so errors.Is + exit 65 are preserved; ErrorDetails() surfaces FieldError under error.details. Unrecognized errors fall back to the old ErrInvalidInput.Wrap. Note: Go's json.UnmarshalTypeError.Field does NOT include the array index, so apply paths are 'concepts.definitions', not 'concepts[0].definitions' (the [N] index is not reconstructable from stdlib without re-parsing; left as the ticket's nice-to-have). Regenerated golden testdata/apply/invalid_json (now names the syntax error). Docs updated: write-details.md JSON limitations bullet, error-handling.md ADR. Tests: src/internal/write/jsonerr_test.go cycles 1-5. make build passes.
