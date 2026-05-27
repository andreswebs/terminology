# Test Plan: E3 — terminology validate manual QA

> **Status**: ready to execute once E3 is completed
> Run end-to-end in a single sitting.

## Purpose

E3's deliverable is **the `terminology validate` command end-to-end**: three
validation tiers (well-formedness, dialect/schema, semantic), the `--strict`
promotion mode, picklist validation on read, unknown-element detection,
line/column tracking in warnings, and the full validate envelope with correct
exit codes.

This is **validation-logic QA**. E2 QA already covered I/O-layer concerns
(dialect detection, legacy normalization, reader/writer round-trip, error
sentinels, DCA/DCT parity). E3 tests the validation tiers, warning codes,
`--strict` promotions, and the validate command's envelope fidelity.

## Scope

### In scope

- Tier-1 well-formedness: malformed XML → exit 65; missing required
  structural elements → exit 65; tier 1 short-circuits (no tier-2/3).
- Tier-2 dialect checks: `invalid_picklist` warning on bad picklist
  values; `unknown_element` warning (strict only).
- Tier-3 semantic checks: `duplicate_id`, `invalid_lang_tag`,
  `missing_term`, `unresolved_crossref`.
- `--strict` promotions: `unknown_element` silent→warning;
  `unresolved_crossref` warning→error.
- `legacy_form_normalized` info-only warning (strict only).
- Warning shape: `code`, `message`, `concept_id`, `line`, `column`.
- Line/column tracking via LineIndex (line/col > 0 on reader warnings).
- Exit codes: 0 (clean), 1 (warnings present), 65 (errors / unparseable).
- Validate envelope: `schema_version`, `ok`, `concepts`, `languages`,
  `warnings`.
- `concepts` count is raw (as-found), not deduplicated.
- `languages` sorted ASCII byte order.
- Picklist functions: all accepted values accessible from `picklist.go`.

### Out of scope

- I/O-layer correctness (dialect detection, DCA/DCT parity, reader/writer
  round-trip) — covered by E2 QA.
- `tbx.Save()` / atomic write / advisory lock — E7.
- Commands beyond `validate` (`lookup`, `schema`, etc.).
- IANA-registry validation of BCP 47 tags — spec explicitly excludes it.
- Line/col on `Glossary.Validate()` warnings (post-parse model; deferred).

## Environment & preconditions

- macOS, Linux, or Windows shell.
- Go toolchain installed (for `make build`).
- `jq` required for envelope assertions.
- Working directory: project root.

### Pre-flight setup

```sh
# 1. Build the binary.
make build

# 2. Bind convenience variables.
export TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"
export FIXTURES="src/internal/tbx/linguist/testdata"
export APP_FIXTURES="src/internal/app/testdata/fixtures"

# 3. Smoke check.
$TT --version
$TT --help

# 4. Verify fixtures exist.
ls "${FIXTURES}/canonical/minimal-dct.tbx" \
   "${FIXTURES}/canonical/all-categories-dct.tbx" \
   "${FIXTURES}/normalized/legacy-forms.tbx" \
   "${APP_FIXTURES}/with-warnings.tbx"
```

If any of steps 1–4 fails, **stop**: build or fixture tree is broken.

## Entry criteria

- All E3 tickets closed.
- `make build` exits 0.
- `cd src && go test ./...` exits 0.

## Exit criteria

- Every **P0** test case passes — no exceptions.
- Every **P1** test case passes — no exceptions.
- Every **P2** test case passes, OR a follow-up ticket is filed with a
  reproducer.
- Every **P3** test case is run and recorded; failures noted but not
  blocking.

## Risk areas

| Risk                                                | Mitigation                                          |
| --------------------------------------------------- | --------------------------------------------------- |
| Tier-1 doesn't short-circuit on malformed XML       | TC-T1-001/002 verify exit 65 and no tier-2/3 output |
| `--strict` promotion mismatch                       | TC-STRICT-\* cover both directions                  |
| `unknown_element` leaks into lenient mode           | TC-UNK-002 verifies suppression                     |
| Picklist validation false-positives on legacy forms | TC-PICK-003 tests legacy values accepted            |
| Line/col always zero on reader warnings             | TC-LINE-001/002 assert line > 0                     |
| Raw concept count masks duplicates                  | TC-DUP-002 verifies raw count preserved             |
| Warning codes drift from spec                       | TC-CODES-001 enumerates all seven                   |

## Conventions

Same as E2 QA: see [E2-manual-qa.md](E2-manual-qa.md) §Conventions for
envelope shapes, exit code map, and test case format.

### Exit code map (E3 surface)

| Code | Meaning          | Source                                              |
| ---- | ---------------- | --------------------------------------------------- |
| 0    | success          | Valid TBX, no warnings                              |
| 1    | warnings         | Valid TBX, warnings present                         |
| 2    | usage error      | `no_tbx_path` (missing `--tbx` / `TERMINOLOGY_TBX`) |
| 65   | validation_error | Tier-1 failure, or `--strict` errors                |

---

# 1. Tier-1 — Well-formedness

## TC-T1-001 — malformed XML is rejected P0

```sh
BAD_XML=$(mktemp /tmp/bad-XXXXXX.tbx)
echo '<tbx type="TBX-Linguist"><unclosed' > "${BAD_XML}"
$TT --tbx "${BAD_XML}" validate 2>err.json >/dev/null
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
rm -f "${BAD_XML}"
```

- **exit**: `65`
- **jq_exit**: `0`
- Tier-1 short-circuits — no validate envelope on stdout.

## TC-T1-002 — empty file is rejected P0

```sh
EMPTY_TBX=$(mktemp /tmp/empty-XXXXXX.tbx)
: > "${EMPTY_TBX}"
$TT --tbx "${EMPTY_TBX}" validate 2>err.json >/dev/null
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
rm -f "${EMPTY_TBX}"
```

- **exit**: `65`
- **jq_exit**: `0`

## TC-T1-003 — valid XML, missing required structure P0

```sh
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
$TT --tbx "${BAD_STRUCT}" validate 2>err.json >/dev/null
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
rm -f "${BAD_STRUCT}"
```

- **exit**: `65`
- **jq_exit**: `0`
- XML parses but required `<text><body>` structure missing.

## TC-T1-004 — tier-1 failure has no validate envelope on stdout P1

```sh
BAD_XML=$(mktemp /tmp/bad-XXXXXX.tbx)
echo '<tbx type="TBX-Linguist"><unclosed' > "${BAD_XML}"
$TT --tbx "${BAD_XML}" validate >out.txt 2>/dev/null
echo "stdout_empty=$([ ! -s out.txt ] && echo yes || echo no)"
rm -f "${BAD_XML}" out.txt
```

- **stdout_empty**: `yes`
- Tier-1 short-circuits — stdout is empty, error goes to stderr only.

---

# 2. Tier-2 — Dialect/schema checks

## TC-PICK-001 — invalid picklist value emits warning P0

Create a fixture with an invalid `administrativeStatus` value:

```sh
PICK_TBX=$(mktemp /tmp/pick-XXXXXX.tbx)
cat > "${PICK_TBX}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader>
    <fileDesc><sourceDesc><p>picklist test</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>test</term>
          <min:administrativeStatus>bogusStatus</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
XML
$TT --tbx "${PICK_TBX}" validate 2>/dev/null \
  | jq -e '.ok == true and (.warnings[] | select(.code == "invalid_picklist") | .code) == "invalid_picklist"'
echo "jq_exit=$?"
echo "exit=$?"
rm -f "${PICK_TBX}"
```

- **jq_exit**: `0`
- **exit**: `1` (warnings present).

## TC-PICK-002 — valid picklist values produce no warning P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '(.warnings | map(select(.code == "invalid_picklist")) | length) == 0'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- A clean fixture has no picklist warnings.

## TC-PICK-003 — legacy admin status forms are accepted P1

```sh
$TT --tbx "${FIXTURES}/normalized/legacy-forms.tbx" validate 2>/dev/null \
  | jq -e '(.warnings | map(select(.code == "invalid_picklist")) | length) == 0'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Legacy bare forms (`preferredTerm`, `admittedTerm`, etc.) are valid
  picklist values — no false-positive `invalid_picklist` warning.

## TC-UNK-001 — unknown element reported in strict mode P0

Create a fixture with an element outside the supported set:

```sh
UNK_TBX=$(mktemp /tmp/unk-XXXXXX.tbx)
cat > "${UNK_TBX}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader>
    <fileDesc><sourceDesc><p>unknown element test</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <madeUpElement>surprise</madeUpElement>
      <langSec xml:lang="en">
        <termSec>
          <term>test</term>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
XML
$TT --tbx "${UNK_TBX}" validate --strict 2>/dev/null \
  | jq -e '(.warnings[] | select(.code == "unknown_element") | .code) == "unknown_element"'
echo "jq_exit=$?"
rm -f "${UNK_TBX}"
```

- **jq_exit**: `0`
- `--strict` surfaces `unknown_element` as a warning.

## TC-UNK-002 — unknown element suppressed in lenient mode P0

```sh
UNK_TBX=$(mktemp /tmp/unk-XXXXXX.tbx)
cat > "${UNK_TBX}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader>
    <fileDesc><sourceDesc><p>unknown element test</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <madeUpElement>surprise</madeUpElement>
      <langSec xml:lang="en">
        <termSec>
          <term>test</term>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
XML
$TT --tbx "${UNK_TBX}" validate 2>/dev/null \
  | jq -e '(.warnings | map(select(.code == "unknown_element")) | length) == 0'
echo "jq_exit=$?"
rm -f "${UNK_TBX}"
```

- **jq_exit**: `0`
- Without `--strict`, unknown elements are tolerated silently.

---

# 3. Tier-3 — Semantic checks

## TC-DUP-001 — duplicate concept ID emits warning P0

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null \
  | jq -e '(.warnings[] | select(.code == "duplicate_id") | .code) == "duplicate_id"'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- `with-warnings.tbx` has two `<conceptEntry id="c1">`.

## TC-DUP-002 — raw concept count includes duplicates P0

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null \
  | jq -e '.concepts == 2'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- `concepts: 2` — as-found count, not deduplicated.

## TC-DUP-003 — duplicate_id warning carries concept_id P1

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null \
  | jq -e '.warnings[] | select(.code == "duplicate_id") | .concept_id == "c1"'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-LANG-001 — invalid BCP 47 tag emits warning P0

```sh
LANG_TBX=$(mktemp /tmp/lang-XXXXXX.tbx)
cat > "${LANG_TBX}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader>
    <fileDesc><sourceDesc><p>lang tag test</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec><term>hello</term></termSec>
      </langSec>
      <langSec xml:lang="not a valid tag!!!">
        <termSec><term>bad</term></termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
XML
$TT --tbx "${LANG_TBX}" validate 2>/dev/null \
  | jq -e '(.warnings[] | select(.code == "invalid_lang_tag") | .code) == "invalid_lang_tag"'
echo "jq_exit=$?"
rm -f "${LANG_TBX}"
```

- **jq_exit**: `0`
- Exit: `1` (warnings present).

## TC-LANG-002 — valid BCP 47 tags produce no warning P1

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '(.warnings | map(select(.code == "invalid_lang_tag")) | length) == 0'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-LANG-003 — invalid_lang_tag carries concept_id P1

```sh
LANG_TBX=$(mktemp /tmp/lang-XXXXXX.tbx)
cat > "${LANG_TBX}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader>
    <fileDesc><sourceDesc><p>lang tag test</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="kabbalah">
      <langSec xml:lang="!!!">
        <termSec><term>bad</term></termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
XML
$TT --tbx "${LANG_TBX}" validate 2>/dev/null \
  | jq -e '.warnings[] | select(.code == "invalid_lang_tag") | .concept_id == "kabbalah"'
echo "jq_exit=$?"
rm -f "${LANG_TBX}"
```

- **jq_exit**: `0`

## TC-TERM-001 — missing term in langSec emits warning P0

```sh
MISS_TBX=$(mktemp /tmp/miss-XXXXXX.tbx)
cat > "${MISS_TBX}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader>
    <fileDesc><sourceDesc><p>missing term test</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec><term>hello</term></termSec>
      </langSec>
      <langSec xml:lang="he">
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
XML
$TT --tbx "${MISS_TBX}" validate 2>/dev/null \
  | jq -e '(.warnings[] | select(.code == "missing_term") | .code) == "missing_term"'
echo "jq_exit=$?"
rm -f "${MISS_TBX}"
```

- **jq_exit**: `0`

## TC-XREF-001 — unresolved crossref emits warning (lenient) P0

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null \
  | jq -e '(.warnings[] | select(.code == "unresolved_crossref") | .code) == "unresolved_crossref"'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- `with-warnings.tbx` has `crossReference target="nonexistent"`.

## TC-XREF-002 — unresolved crossref promoted to error in strict P0

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate --strict 2>err.json >/dev/null
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
```

- **exit**: `65`
- **jq_exit**: `0`
- `--strict` promotes `unresolved_crossref` from warning to error.

## TC-XREF-003 — resolved crossrefs produce no warning P1

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '(.warnings | map(select(.code == "unresolved_crossref")) | length) == 0'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

---

# 4. `--strict` promotions

## TC-STRICT-001 — strict clean file passes P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate --strict 2>/dev/null \
  | jq -e '.ok == true and .warnings == []'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Exit: `0` — `--strict` doesn't reject a clean file.

## TC-STRICT-002 — strict promotes unresolved_crossref to error P0

(Same as TC-XREF-002 — listed here for traceability.)

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate --strict 2>err.json >/dev/null
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
```

- **exit**: `65`
- **jq_exit**: `0`

## TC-STRICT-003 — strict surfaces unknown_element P0

(Same as TC-UNK-001 — listed here for traceability.)

Uses a fixture with an element outside the supported set.
See TC-UNK-001 for the full script.

- `--strict` surfaces `unknown_element` as a warning.

## TC-STRICT-004 — lenient suppresses unknown_element P0

(Same as TC-UNK-002 — listed here for traceability.)

- Without `--strict`, unknown elements are tolerated silently.

## TC-STRICT-005 — legacy_form_normalized appears in strict only P1

```sh
$TT --tbx "${FIXTURES}/normalized/legacy-forms.tbx" validate --strict 2>/dev/null \
  | jq -e '(.warnings | map(select(.code == "legacy_form_normalized")) | length) > 0'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- `legacy_form_normalized` is info-only under `--strict`.

## TC-STRICT-006 — legacy_form_normalized suppressed in lenient P1

```sh
$TT --tbx "${FIXTURES}/normalized/legacy-forms.tbx" validate 2>/dev/null \
  | jq -e '(.warnings | map(select(.code == "legacy_form_normalized")) | length) == 0'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Lenient mode: no `legacy_form_normalized` warnings.

---

# 5. Warning shape & line/column tracking

## TC-SHAPE-001 — every warning has code and message P0

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null \
  | jq -e '.warnings | all(has("code") and has("message"))'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-SHAPE-002 — warnings with concept_id carry a non-empty value P1

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null \
  | jq -e '.warnings | map(select(.concept_id)) | all(.concept_id != "")'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-LINE-001 — reader warnings carry line > 0 P1

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null \
  | jq -e '.warnings[] | select(.code == "unresolved_crossref") | .line > 0'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- LineIndex wiring populates `line` on reader-emitted warnings.

## TC-LINE-002 — reader warnings carry column > 0 P2

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null \
  | jq -e '.warnings[] | select(.code == "unresolved_crossref") | .column > 0'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-LINE-003 — multiple warnings have distinct line values P2

```sh
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
$TT --tbx "${MULTI_TBX}" validate 2>/dev/null \
  | jq -e '[.warnings[] | select(.code == "unresolved_crossref") | .line] | unique | length > 1'
echo "jq_exit=$?"
rm -f "${MULTI_TBX}"
```

- **jq_exit**: `0`
- Two warnings on different source lines have distinct `.line` values.

---

# 6. Warning codes — completeness

## TC-CODES-001 — all seven warning codes are spec-defined P1

Verify that the validate command produces only spec-defined warning
codes. Run all fixtures and collect distinct codes:

```sh
KNOWN_CODES="duplicate_id invalid_lang_tag invalid_picklist legacy_form_normalized missing_term unknown_element unresolved_crossref"

ALL_CODES=$({
  $TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null
  $TT --tbx "${FIXTURES}/normalized/legacy-forms.tbx" validate --strict 2>/dev/null
} | jq -r '.warnings[].code' 2>/dev/null | sort -u | tr '\n' ' ' | sed 's/ $//')

echo "found codes: ${ALL_CODES}"
for code in ${ALL_CODES}; do
  if echo "${KNOWN_CODES}" | grep -qw "$code"; then
    echo "ok: ${code}"
  else
    echo "FAIL: unknown code ${code}"
  fi
done
```

- Every code is one of the seven spec-defined codes.
- No unknown codes appear.

---

# 7. Envelope fidelity

## TC-ENV-001 — clean file: ok=true, warnings=[] P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and .warnings == [] and .concepts >= 1 and (.languages | length) >= 1'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Exit: `0`.

## TC-ENV-002 — warning-bearing: ok=true, warnings non-empty P0

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and (.warnings | length) > 0'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Exit: `1`.

## TC-ENV-003 — languages sorted ASCII byte order P0

```sh
$TT --tbx "${FIXTURES}/canonical/rich-dct.tbx" validate 2>/dev/null \
  | jq -e '.languages == (.languages | sort)'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- `["en", "es", "he"]` — already sorted.

## TC-ENV-004 — schema_version is 1 P1

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '.schema_version == 1'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-ENV-005 — languages and warnings are never null P1

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '(.languages | type) == "array" and (.warnings | type) == "array"'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

---

# 8. Exit codes

## TC-EXIT-001 — exit 0 for clean file P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate >/dev/null 2>/dev/null
echo "exit=$?"
```

- **exit**: `0`

## TC-EXIT-002 — exit 1 for warnings P0

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate >/dev/null 2>/dev/null
echo "exit=$?"
```

- **exit**: `1`

## TC-EXIT-003 — exit 65 for malformed XML P0

```sh
BAD_XML=$(mktemp /tmp/bad-XXXXXX.tbx)
echo '<broken' > "${BAD_XML}"
$TT --tbx "${BAD_XML}" validate >/dev/null 2>/dev/null
echo "exit=$?"
rm -f "${BAD_XML}"
```

- **exit**: `65`

## TC-EXIT-004 — exit 65 for strict errors P0

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate --strict >/dev/null 2>/dev/null
echo "exit=$?"
```

- **exit**: `65`

## TC-EXIT-005 — exit 2 for missing TBX path P0

```sh
$TT validate >/dev/null 2>/dev/null
echo "exit=$?"
```

- **exit**: `2`

---

# 9. Tier sequencing

## TC-SEQ-001 — tier-1 failure prevents tier-2/3 output P0

A file that fails tier-1 (malformed XML) should produce no validate
envelope — only the error envelope on stderr.

```sh
BAD_XML=$(mktemp /tmp/bad-XXXXXX.tbx)
echo '<broken' > "${BAD_XML}"
$TT --tbx "${BAD_XML}" validate >out.txt 2>err.json
echo "stdout_empty=$([ ! -s out.txt ] && echo yes || echo no)"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
rm -f "${BAD_XML}" out.txt err.json
```

- **stdout_empty**: `yes`
- **jq_exit**: `0`

## TC-SEQ-002 — tier-2 failure does not prevent tier-3 P1

A file with both a picklist violation (tier-2) and a semantic violation
(tier-3) should report both warnings in a single envelope.

```sh
BOTH_TBX=$(mktemp /tmp/both-XXXXXX.tbx)
cat > "${BOTH_TBX}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min">
  <tbxHeader>
    <fileDesc><sourceDesc><p>tiers 2+3</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec>
          <term>alpha</term>
          <min:administrativeStatus>bogusStatus</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
    <conceptEntry id="c1">
      <langSec xml:lang="en">
        <termSec><term>beta</term></termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
XML
$TT --tbx "${BOTH_TBX}" validate 2>/dev/null \
  | jq -e '(.warnings | map(.code) | unique | sort) as $codes | ($codes | contains(["duplicate_id"])) and ($codes | contains(["invalid_picklist"]))'
echo "jq_exit=$?"
rm -f "${BOTH_TBX}"
```

- **jq_exit**: `0`
- Both tier-2 (`invalid_picklist`) and tier-3 (`duplicate_id`) appear.

---

# 10. Stream routing

## TC-STREAM-001 — validate success → stdout only P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate >out.txt 2>err.txt
echo "stdout_has_json=$(jq -e '.ok == true' out.txt >/dev/null 2>&1 && echo yes || echo no)"
echo "stderr_empty=$([ ! -s err.txt ] && echo yes || echo no)"
rm -f out.txt err.txt
```

- **stdout_has_json**: `yes`
- **stderr_empty**: `yes`

## TC-STREAM-002 — validate error → stderr only P0

```sh
$TT validate >out.txt 2>err.txt
echo "stderr_has_json=$(jq -e '.ok == false' err.txt >/dev/null 2>&1 && echo yes || echo no)"
echo "stdout_empty=$([ ! -s out.txt ] && echo yes || echo no)"
rm -f out.txt err.txt
```

- **stderr_has_json**: `yes`
- **stdout_empty**: `yes`

## TC-STREAM-003 — warnings envelope → stdout P1

```sh
$TT --tbx "${APP_FIXTURES}/with-warnings.tbx" validate >out.txt 2>err.txt
echo "stdout_has_warnings=$(jq -e '(.warnings | length) > 0' out.txt >/dev/null 2>&1 && echo yes || echo no)"
rm -f out.txt err.txt
```

- **stdout_has_warnings**: `yes`
- Exit: `1`.

---

# 11. Sign-off checklist

Tick before declaring E3 manual QA pass.

- [ ] Section 1 (tier-1 well-formedness): all P0 + P1 cases pass;
      malformed XML rejected; tier-1 short-circuits.
- [ ] Section 2 (tier-2 dialect checks): picklist validation works;
      unknown elements handled by `--strict`.
- [ ] Section 3 (tier-3 semantic checks): duplicate_id, invalid_lang_tag,
      missing_term, unresolved_crossref all detected correctly.
- [ ] Section 4 (`--strict` promotions): unresolved_crossref promoted to
      error; unknown_element surfaced as warning; legacy_form_normalized
      appears only in strict.
- [ ] Section 5 (warning shape): code + message present; concept_id
      populated; line/col > 0 on reader warnings.
- [ ] Section 6 (warning codes): all codes are spec-defined.
- [ ] Section 7 (envelope fidelity): shapes correct; languages sorted;
      never null.
- [ ] Section 8 (exit codes): 0/1/2/65 all correct.
- [ ] Section 9 (tier sequencing): tier-1 short-circuits; tiers 2+3
      aggregate.
- [ ] Section 10 (stream routing): success → stdout, errors → stderr.
- [ ] No undocumented behaviour observed.

## On failure

Any failing case must be classified:

- **Validation-logic bug** — tier-2/3 check is missing or wrong. File a
  ticket tagged `e3,bug` linked to `ter-told`.
- **Promotion bug** — `--strict` fails to promote or incorrectly
  promotes. File a ticket tagged `e3,bug,strict`.
- **Envelope bug** — missing field, wrong type, null instead of `[]`.
  File a ticket tagged `e3,bug,envelope`.
- **Spec ambiguity** — test plan and spec disagree. Update the spec
  **first** (PR), then re-derive the test case.

Do **not** silence a failing case by updating the test plan to match
buggy behaviour. The plan is derived from
[`docs/specs/003-validate-command.md`](../docs/specs/003-validate-command.md);
the spec wins.
