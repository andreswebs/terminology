---
id: ter-9dho
status: closed
deps: []
links: [ter-eedk]
created: 2026-05-25T19:30:06Z
type: bug
priority: 0
assignee: Andre Silva
parent: ter-told
tags: [e3, bug, tier1]
---
# E3.BUG — Tier-1 accepts TBX missing <text><body> structure

## Summary

A structurally valid XML file with the `<tbx>` root element but **no
`<text><body>` element** is accepted by `terminology validate` with exit 0
instead of being rejected at tier-1 with exit 65.

The binary returns an empty validate envelope:

```json
{"schema_version":1,"ok":true,"concepts":0,"languages":[],"warnings":[]}
```

## QA reference

E3 manual QA report Finding 1: `qa/E3-manual-qa.report.md`
Test case: TC-T1-003 (P0)

## Reproduction

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"

BAD_STRUCT=$(mktemp /tmp/bad-struct-XXXXXX.tbx)
cat > "${BAD_STRUCT}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader>
    <fileDesc>
      <sourceDesc><p>missing body</p></sourceDesc>
    </fileDesc>
  </tbxHeader>
</tbx>
XML

${TT} --tbx "${BAD_STRUCT}" validate 2>err.json
echo "exit=$?"
# Actual: exit=0, envelope on stdout with concepts:0
# Expected: exit=65, validation_error on stderr
rm -f "${BAD_STRUCT}" err.json
```

## Expected behavior

Exit 65 with `validation_error` error envelope on stderr. The required
`<text><body>` structure is missing, which should fail tier-1
well-formedness per the E3 spec.

## Actual behavior

Exit 0 with an empty-but-valid success envelope on stdout. The reader
silently produces a `Glossary` with zero concepts.

## Impact

A TBX file missing its entire body element passes validation silently.
Downstream consumers may assume the file was validated and contains data.

## Root cause

The linguist reader's `Decode()` does not check for the presence of
`<text><body>` — if the element is absent, it simply returns an empty
`Glossary` with no error. The reader treats absence of body as "zero
concepts" rather than a structural error.

## Related tickets

- ter-eedk (E3.T4 — Tier-1 well-formedness validation) — this ticket's
  Cycle 3 explicitly specifies "empty body produces empty Glossary (not
  error)" for the `empty-body.tbx` fixture (which has `<body>` present
  but empty). However, the case here is **missing `<body>` entirely**
  (no `<text>` element at all), which should be a tier-1 failure.

## Fix direction

In the linguist reader's `Decode()`, after parsing completes, verify that
the `<text><body>` structure was encountered. If not, return an error
(which `validateAction` will wrap as `ErrValidationError`, producing
exit 65). Alternatively, check during streaming decode that `<text>` and
`<body>` elements were seen before returning.

## Refs

- E3 spec: docs/specs/003-validate-command.md §"Scope" tier 1
- E3 QA plan: qa/E3-manual-qa.md §TC-T1-003


## Notes

**2026-05-25T19:34:24Z**

Fixed: Added inBody check after the XML parsing loop in linguist reader's Decode(). If <text><body> was never encountered, returns an error which validateAction wraps as ErrValidationError (exit 65). Added test fixture missing-body.tbx, unit test TestDecode_MissingBody_ReturnsError in reader_test.go, and golden CLI test TestValidate_MissingBody_Golden. Existing empty-body test (which has <body> present but empty) continues to pass — the distinction is missing vs empty.
