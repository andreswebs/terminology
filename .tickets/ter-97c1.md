---
id: ter-97c1
status: closed
deps: []
links: [ter-w6kr]
created: 2026-05-25T19:30:31Z
type: bug
priority: 1
assignee: Andre Silva
parent: ter-told
tags: [e3, bug, lineindex]
---
# E3.BUG — Line/column not populated on Glossary.Validate() warnings

## Summary

Reader warnings emitted from `Glossary.Validate()` (tier-3 semantic
checks) have `line == 0` and `column == 0` in the validate envelope JSON
output. The fields are omitted entirely due to `omitempty`. The E3 spec
acceptance criteria state "warnings carry concept_id + line/col."

## QA reference

E3 manual QA report Findings 2–4: `qa/E3-manual-qa.report.md`
Test cases: TC-LINE-001 (P1), TC-LINE-002 (P2), TC-LINE-003 (P2)

## Reproduction

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"
APP_FIXTURES="src/internal/app/testdata/fixtures"

# TC-LINE-001: unresolved_crossref has line == 0
${TT} --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null \
  | jq '.warnings[] | select(.code == "unresolved_crossref") | {line, column}'
# Actual: {"line": null, "column": null}
# Expected: line > 0, column > 0

# TC-LINE-003: two crossref warnings on different lines, both lack line data
MULTI_TBX=$(mktemp /tmp/multi-XXXXXX.tbx)
cat > "${MULTI_TBX}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:basic="http://www.tbxinfo.net/ns/basic">
  <tbxHeader>
    <fileDesc><sourceDesc><p>multi warning</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <basic:crossReference target="missing1">see missing1</basic:crossReference>
      <langSec xml:lang="en">
        <termSec><term>alpha</term></termSec>
      </langSec>
    </conceptEntry>
    <conceptEntry id="c2">
      <basic:crossReference target="missing2">see missing2</basic:crossReference>
      <langSec xml:lang="en">
        <termSec><term>beta</term></termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
XML
${TT} --tbx "${MULTI_TBX}" validate 2>/dev/null | jq '.warnings'
# Actual: both warnings have no line/column fields
rm -f "${MULTI_TBX}"
```

## Affected warning codes

- `unresolved_crossref` — tier-3, emitted by `Glossary.Validate()`
- `duplicate_id` — tier-3, emitted by `Glossary.Validate()`
- `invalid_lang_tag` — tier-3, emitted by `Glossary.Validate()`
- `missing_term` — tier-3, emitted by `Glossary.Validate()`

All four tier-3 warning codes operate on the in-memory `Glossary` model
after parsing is complete, so they have no access to XML source positions.

## Root cause

E3.T11 (ter-w6kr) wired LineIndex into the linguist reader for
**reader-emitted** warnings (`unknown_element`, `invalid_picklist`,
`legacy_form_normalized`), which are detected during XML streaming. The
tier-3 semantic warnings (`unresolved_crossref`, `duplicate_id`,
`invalid_lang_tag`, `missing_term`) are emitted by `Glossary.Validate()`,
which operates on the post-parse domain model and has no access to source
byte offsets.

ter-w6kr explicitly scoped out model-level warnings: "Out of scope:
Line/col for Glossary.Validate() warnings (these operate on the
in-memory model, not the XML stream — line/col would require storing
offsets during decode, which is a larger change)."

However, the E3 spec acceptance criteria and QA plan expect line/col on
all warnings, not just reader-emitted ones.

## Fix direction

Store the XML byte offset of each element during decode in a side channel
(e.g., a map from concept ID to byte offset in the `Glossary` or a
parallel structure). When `Glossary.Validate()` emits a warning for a
concept, look up the stored offset and call `LineIndex.Position()` to
populate `Line`/`Col`.

Alternatively, store `(line, col)` directly on `Concept` or in a
`SourcePositions` map during decode, so Validate can populate warnings
without needing the LineIndex at validation time.

## Related tickets

- ter-w6kr (E3.T11 — LineIndex wiring into reader) — wired line/col
  for reader-emitted warnings only; model-level warnings explicitly
  out of scope.
- ter-cplu (E2.T5 — Line index for reader position tracking) — the
  lineindex package itself.

## Refs

- E3 spec: docs/specs/003-validate-command.md §"Line/column tracking"
- E3 epic acceptance: "warnings carry concept_id + line/col"
- E3 QA plan: qa/E3-manual-qa.md §TC-LINE-001, TC-LINE-002, TC-LINE-003


## Notes

**2026-05-25T19:42:44Z**

Fix: Added StartLine/StartCol fields to Concept and LangSection in the domain model. The reader captures XML source positions during decode (via dc.pos() on conceptEntry and langSec start elements). Validate() populates Line/Col on all four tier-3 warning codes: duplicate_id and unresolved_crossref use Concept position; invalid_lang_tag and missing_term use LangSection position for more precise pointing. Updated golden test file for validate/warnings. Added unit tests for each warning code with position verification, an integration test loading the fixture through tbx.Load(), and a reader test verifying positions are populated on decoded Concept/LangSection structs.
