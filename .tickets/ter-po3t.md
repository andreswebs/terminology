---
id: ter-po3t
status: open
deps: []
links: []
created: 2026-07-12T11:18:13Z
type: bug
priority: 1
assignee: Andre Silva
tags: [field-feedback, write, tbx, fragment, bug]
---
# BUG-1 — TBX fragment input silently drops definitions and statuses

# BUG-1 — TBX fragment input silently drops definitions and statuses (false success)

## Severity

Bug — silent data loss / false success. An agent building a glossary from TBX
fragments is told the write succeeded (`ok:true`, exit 0) while definitions and
administrative statuses are discarded. This is the most dangerous defect class.

Source: field feedback in `.local/tmp/issues.md` (BUG-1), build `b675e4d`.

## Reproduction

```sh
terminology init --tbx g.tbx --source-lang en
cat <<'EOF' | terminology concept add --tbx g.tbx
<conceptEntry id="irimi-nage">
  <descrip type="subjectField">aikido</descrip>
  <langSec xml:lang="en">
    <descrip type="definition">Nage enters past uke to throw.</descrip>
    <termSec>
      <term>entering throw</term>
      <termNote type="administrativeStatus">preferredTerm-admn-sts</termNote>
    </termSec>
  </langSec>
</conceptEntry>
EOF
grep -c definition g.tbx   # -> 0  (BUG: definition dropped, yet exit 0 / ok:true)
```

Expected: either the definition + administrative status are persisted, or the
command fails closed with a clear `invalid_input` error that names the dropped
elements. Never `ok:true` after discarding recognized input.

## Root cause (verified against source)

The fragment reproduction above is **DCA-style** TBX-Linguist: it uses the
generic carrier elements `<descrip type="...">` and `<termNote type="...">`
rather than the namespaced DCT elements (`basic:definition`,
`min:administrativeStatus`). Two independent defects combine:

1. **The fragment shell hardcodes `style="dct"`.**
   [`src/internal/write/write.go:104-122`](src/internal/write/write.go#L104-L122)
   defines `tbxShellPrefix` with `style="dct"` baked in, and
   [`wrapInTBXShell`](src/internal/write/write.go#L124-L136) always uses it.
   The linguist reader picks the dialect from that attribute:
   [`detectStyle`](src/internal/tbx/linguist/reader.go#L314-L323) returns
   `StyleDCT`, so
   [`dialectFor`](src/internal/tbx/linguist/dialect.go#L27-L32) selects
   `dialectDCT` ([`dialect.go:34`](src/internal/tbx/linguist/dialect.go#L34)).
   Under DCT, the DCA carriers `<descrip>` / `<termNote>` are unrecognized:
   [`decodeConceptFields`](src/internal/tbx/linguist/reader.go#L393-L480),
   [`decodeLangSecFields`](src/internal/tbx/linguist/reader.go#L530-L553), and
   [`decodeTermFields`](src/internal/tbx/linguist/reader.go#L609-L790) all fall
   into their `default:` branch, emit an `unknown_element` warning, and `skip()`
   the element. `dialectDCA`
   ([`dialect.go:121`](src/internal/tbx/linguist/dialect.go#L121)) already knows
   how to map every one of these elements — it is simply never selected.

2. **`ParseTBXFragment` discards the reader's warnings.**
   [`src/internal/write/write.go:79`](src/internal/write/write.go#L79) calls
   `g, _, err := r.Decode(...)` — the `[]tbx.Warning` slice (which holds the
   `unknown_element` warnings for every dropped element) is thrown away. So even
   the signal that elements were dropped never reaches the caller, and the
   command returns `ok:true`.

Net effect: the concept is written with only `<term>`; the concept-level
subjectField, the definition, and the term's administrative status are gone,
with no error and no warning.

Both entry points are affected, because both funnel through `ParseTBXFragment`:
`concept add` via
[`parseConceptFromTBXStdin`](src/internal/app/commands/concept_add.go#L139-L166),
and `apply` via
[`LoadApplyFile`](src/internal/write/apply.go#L148-L167).

## Fix design

Make the fragment path both **correct** (DCA fragments parse) and **safe**
(nothing is ever silently dropped):

1. **Detect the fragment's style and wrap accordingly.** Add
   `detectFragmentStyle(data []byte) tbx.Style` in the `write` package. Scan the
   fragment's start elements; if any is a DCA-specific generic carrier
   (`descrip`, `termNote`, `admin`, `adminNote`, `ref`, `xref`, `transac`,
   `transacNote`) — DCT never uses these — return `tbx.StyleDCA`; otherwise
   `tbx.StyleDCT`. Parameterize the shell so `wrapInTBXShell(fragment, rootName,
   style)` emits `style="dca"` or `style="dct"`. This lets the existing
   `dialectDCA` machinery map `<descrip>` / `<termNote>` correctly, with no
   changes to the reader.

2. **Fail closed on any remaining dropped element.** Stop discarding the
   reader's warnings in `ParseTBXFragment`. After `Decode`, inspect the returned
   warnings; if any has `Code == "unknown_element"`, return an
   `ErrInvalidInput`-coded error whose message names the dropped element(s)
   (the warning `Message` already contains the element name, e.g.
   `unknown element <descrip type="madeUpThing">`). Only `unknown_element` is
   fatal here — leave benign warnings such as `legacy_form_normalized` and
   `invalid_picklist` non-fatal (they are surfaced through the normal validate
   path). This guarantees the "never `ok:true` after discarding input" invariant
   even for genuinely unsupported elements, and preserves the tool's
   agent-first, fail-closed philosophy (mirrors the existing
   `fatalWarningCodes` gate in
   [`validateForWrite`](src/internal/write/write.go#L183-L201)).

Keep `ParseTBXFragment`'s signature `([]tbx.Concept, error)` unchanged so both
callers (`concept add`, `apply`) inherit the fix with no edits. Reuse the
existing `ErrInvalidInput` sentinel
([`src/internal/write/errors.go:29-33`](src/internal/write/errors.go#L29-L33)),
wrapping with `%w` semantics via its `.Wrap(...)` method so the error code,
hint, and exit 65 are preserved.

### Go conventions

- Exported `ParseTBXFragment` keeps its doc-less internal helper style; the new
  `detectFragmentStyle` is unexported (lower-case) and lives beside the other
  fragment helpers in `write.go`.
- Wrap the underlying cause with the sentinel's `.Wrap(err)` — do not swallow
  and re-`fmt.Errorf` a new string that loses `errors.Is` identity.
- Early-return on error; keep the happy path at minimal indentation.
- Use `slices`/`strings` from stdlib for element-name membership checks; no new
  deps.
- `gofmt` clean; `MixedCaps`; initialisms (`ID`, `XML`, `TBX`) stay upper.

## TDD plan (vertical slices — one test, one change, repeat)

Public interface under test: `write.ParseTBXFragment([]byte)`. Add cases to
[`src/internal/write/write_test.go`](src/internal/write/write_test.go). Do NOT
write all tests up front; go RED→GREEN per cycle.

### Cycle 1 (tracer) — DCA fragment persists definition + status

RED: feed the BUG-1 reproduction fragment (DCA: `<descrip type="definition">`,
`<termNote type="administrativeStatus">`, concept-level `<descrip
type="subjectField">`). Assert the returned `tbx.Concept` has:
`SubjectField == "aikido"`, one concept-level `Definitions` entry with
`Plain == "Nage enters past uke to throw."`, and the `en` langSec's single term
with `Surface == "entering throw"` and `AdministrativeStatus ==
tbx.StatusPreferred`. Fails today (all enrichment dropped).

GREEN: add `detectFragmentStyle`; thread the detected style through
`wrapInTBXShell`; select `style="dca"` for this input.

### Cycle 2 — DCT fragment still parses (regression guard)

RED: feed a namespaced DCT fragment (`<basic:definition>`,
`<min:administrativeStatus>` with the min/basic/ling namespaces declared on the
`conceptEntry` or relying on the shell's declarations). Assert the definition
and status survive. Confirms style detection defaults to DCT and does not
regress the existing path.

GREEN: ensure `detectFragmentStyle` returns `tbx.StyleDCT` when no DCA carrier
is present.

### Cycle 3 — genuinely unknown element fails closed (never silent)

RED: feed a fragment containing an unrecognized carrier, e.g. `<descrip
type="madeUpThing">x</descrip>` inside a langSec (DCA) — or a bogus namespaced
element under DCT. Assert `ParseTBXFragment` returns a non-nil error whose
`Code()` is `"invalid_input"` and whose message names the offending element
(contains `madeUpThing`). Assert no concept is returned as a success.

GREEN: capture the warnings from `Decode`; if any `unknown_element` remains,
return `ErrInvalidInput.Wrap(...)` listing the dropped element name(s).

### Cycle 4 — end-to-end CLI guard

RED: add a CLI-level test (see the golden/table tests in
[`src/internal/app/write_golden_test.go`](src/internal/app/write_golden_test.go)
and `src/internal/app/commands/*_test.go`) driving `concept add` with the BUG-1
DCA fragment on stdin against a temp glossary. Assert exit 0, `ok:true`, and
that reading the file back shows the definition and `preferredTerm-admn-sts`
persisted. Add a second case: the unknown-element fragment yields exit 65 and an
`error.code == "invalid_input"` envelope (no file mutation).

GREEN: covered by Cycles 1–3; wire the test.

### Refactor

- Extract the DCA-carrier name set to a package-level `var` (documented) if it
  reads cleanly; dedupe any shell-building duplication.
- Run tests after each step. Never refactor while RED.

## Documentation updates (part of this ticket)

Update [`docs/skills/terminology/references/write-details.md`](docs/skills/terminology/references/write-details.md)
§"3. TBX fragment stdin" (the "Caveat — lossy parser" block, ~lines 180–185):
remove the lossy-parser warning and document the new behavior — DCA and DCT
fragments both round-trip definitions/statuses/readings, and unsupported
elements fail closed with `invalid_input` naming the element. If
[`docs/cli-design.md`](docs/cli-design.md) documents fragment ingest, align it
too. Run `markdownlint-cli2 --config ~/.markdownlint.yaml --fix` on any edited
markdown.

## Acceptance criteria

- The exact BUG-1 reproduction persists both the definition and the
  `preferredTerm-admn-sts` administrative status; `grep -c definition g.tbx`
  returns ≥ 1; exit 0.
- A DCT (namespaced) fragment continues to parse with no regression.
- A fragment with any unrecognized element fails closed: exit 65,
  `{"ok":false,"error":{"code":"invalid_input",...}}`, message names the dropped
  element, and the glossary file is not modified.
- Both `concept add` and `apply` fragment paths benefit (shared
  `ParseTBXFragment`); no caller signature changes.
- `write-details.md` no longer describes silent fragment loss.
- `make build` passes (fmt-check, vet, lint, test) from the project root.

## Files to touch

- `src/internal/write/write.go` — `detectFragmentStyle`, parameterized
  `wrapInTBXShell`, warning capture + fail-closed in `ParseTBXFragment`.
- `src/internal/write/write_test.go` — Cycles 1–3.
- `src/internal/app/commands/*_test.go` or
  `src/internal/app/write_golden_test.go` — Cycle 4 CLI guard.
- `docs/skills/terminology/references/write-details.md` (and `docs/cli-design.md`
  if it covers fragments).

## Validation

Run `make build` from the project root. For a tighter loop use `make test`
(package: `go test ./src/internal/write/... ./src/internal/app/...`) then finish
with a full `make build`. Do not silence lint with `_ =`.

