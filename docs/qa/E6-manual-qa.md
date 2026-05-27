# Test Plan: E6 — scan & check manual QA

> **Status**: ready to execute once E6 is completed
> Run end-to-end in a single sitting.

## Purpose

E6's deliverable is the **check command** and the **scan frontmatter language
resolution update**. The check command verifies a translated target file against
a source file using the glossary: it detects missing preferred terms,
forbidden (deprecated/superseded) variants, and optionally admitted variants
under `--strict`.

This is **check + scan-update QA**. E5 QA already covered the scan command
surface, matcher pipeline, normalization, and envelope shape. E6 tests the
check algorithm, violation types, language resolution (frontmatter > flag >
error), `--strict` semantics, violation ordering, and the scan frontmatter
update.

## Scope

### In scope

- `check SRC TGT` — verify translated target against source given a glossary.
- Violation types: `missing`, `forbidden_variant`, `admitted_variant` (strict).
- Language resolution: frontmatter `lang:` → CLI flag → `ErrLanguageRequired`.
- `--strict` promoting admitted variants from warnings to violations.
- `--context N` controlling the context window on violations.
- `--fields` projection on check output.
- Violation ordering: `(line, column)` in TGT, missing at tail by `concept_id`.
- Check envelope shape (`source`, `target`, `violations`, `warnings`, `summary`).
- Exit codes: 0 (clean), 1 (violations), 2 (usage error), 3 (I/O error).
- Error envelopes: `no_tbx_path`, `language_required`, `invalid_field`.
- Scan frontmatter language detection (frontmatter > `--lang` > all languages).

### Out of scope

- Matcher pipeline internals (normalization, boundary, Aho-Corasick) — covered by E5 QA.
- Scan command surface (envelope shape, `--context`, `--fields`) — covered by E5 QA.
- Validation logic (`validate` command) — covered by E3 QA.
- Read commands (`lookup`, `schema`, `extract`) — covered by E4 QA.
- Write commands — E7/E8.
- Performance budget — deferred to E9.

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

# 3. Smoke check.
$TT --version
$TT --help

# 4. Verify fixtures exist.
ls "${FIXTURES}/canonical/minimal-dct.tbx" \
   "${FIXTURES}/canonical/all-categories-dct.tbx"
```

### Test corpus setup

Create test markdown, TBX, and fixture files for check and scan tests:

````sh
export QA_TMP=$(mktemp -d /tmp/e6-qa-XXXXXX)

# --- Glossary fixture ---
# Multi-language glossary with preferred, admitted, deprecated terms.
cat > "${QA_TMP}/glossary.tbx" <<'TBXEOF'
<?xml version="1.0" encoding="UTF-8"?>
<?xml-model href="https://raw.githubusercontent.com/LTAC-Global/TBX-Linguist_Module/master/Schema/TBXcheckerTBX-Linguist.sch" type="application/xml" schematypens="http://purl.oclc.org/dml/schematron"?>
<tbx style="dct" type="TBX-Linguist" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader>
    <fileDesc><sourceDesc><p>E6 QA fixture</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="tzimtzum">
      <min:subjectField>kabbalah</min:subjectField>
      <langSec xml:lang="en">
        <termSec>
          <term>tzimtzum</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
      <langSec xml:lang="es">
        <termSec>
          <term>tzimtzum</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
        <termSec>
          <term>contraccion</term>
          <min:administrativeStatus>admittedTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
      <langSec xml:lang="he">
        <termSec>
          <term>צמצום</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
        <termSec>
          <term>כיווץ</term>
          <min:administrativeStatus>deprecatedTerm-admn-sts</min:administrativeStatus>
        </termSec>
        <termSec>
          <term>התכווצות</term>
          <min:administrativeStatus>admittedTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
    <conceptEntry id="sefirah">
      <min:subjectField>kabbalah</min:subjectField>
      <langSec xml:lang="en">
        <termSec>
          <term>sefirah</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
      <langSec xml:lang="es">
        <termSec>
          <term>sefirah</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
      <langSec xml:lang="he">
        <termSec>
          <term>ספירה</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
        <termSec>
          <term>ספירא</term>
          <min:administrativeStatus>supersededTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
    <conceptEntry id="razon-historica">
      <min:subjectField>philosophy</min:subjectField>
      <langSec xml:lang="en">
        <termSec>
          <term>historical reason</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
      <langSec xml:lang="es">
        <termSec>
          <term>razón histórica</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
      <langSec xml:lang="he">
        <termSec>
          <term>תבונה היסטורית</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
TBXEOF

# --- Source: Spanish with glossary terms ---
cat > "${QA_TMP}/source.md" <<'EOF'
---
lang: es
---
# El Concepto de Tzimtzum

El concepto de tzimtzum es central en el pensamiento cabalístico.

Cada sefirah representa un atributo divino.

La razón histórica es un concepto fundamental en la obra de Ortega.
EOF

# --- Target: Hebrew clean (all preferred terms present) ---
# IMPORTANT: Hebrew terms must appear as standalone words (no ה prefix)
# because the matcher uses \p{L} word boundaries — "הצמצום" does NOT
# match glossary term "צמצום" since ה is a Letter character.
cat > "${QA_TMP}/target-clean.md" <<'EOF'
---
lang: he
---
# מושג צמצום

צמצום הוא מושג מרכזי במחשבה הקבלית.

כל ספירה מייצגת תכונה אלוהית.

תבונה היסטורית היא מושג יסודי ביצירתו של אורטגה.
EOF

# --- Target: Hebrew missing a preferred term ---
# "תבונה היסטורית" is absent — replaced with non-glossary paraphrase.
cat > "${QA_TMP}/target-missing.md" <<'EOF'
---
lang: he
---
# מושג צמצום

צמצום הוא מושג מרכזי במחשבה הקבלית.

כל ספירה מייצגת תכונה אלוהית.

הרעיון ההיסטורי הוא מושג יסודי ביצירתו של אורטגה.
EOF

# --- Target: Hebrew with a deprecated variant ---
# כיווץ (deprecated for tzimtzum) appears; צמצום (preferred) also present.
cat > "${QA_TMP}/target-forbidden.md" <<'EOF'
---
lang: he
---
# מושג צמצום

כיווץ הוא מושג מרכזי במחשבה הקבלית. צמצום הוא תהליך.

כל ספירה מייצגת תכונה אלוהית.

תבונה היסטורית היא מושג יסודי ביצירתו של אורטגה.
EOF

# --- Target: Hebrew with a superseded variant ---
# ספירא (superseded for sefirah) appears; ספירה (preferred) is absent.
cat > "${QA_TMP}/target-superseded.md" <<'EOF'
---
lang: he
---
# מושג צמצום

צמצום הוא מושג מרכזי במחשבה הקבלית.

כל ספירא מייצגת תכונה אלוהית.

תבונה היסטורית היא מושג יסודי ביצירתו של אורטגה.
EOF

# --- Target: Hebrew with an admitted variant ---
# התכווצות (admitted for tzimtzum) appears alongside צמצום (preferred).
# Admitted does NOT satisfy the preferred-term check, so the preferred
# must also be present to avoid a "missing" violation.
cat > "${QA_TMP}/target-admitted.md" <<'EOF'
---
lang: he
---
# מושג צמצום

התכווצות הוא מושג מרכזי במחשבה הקבלית. צמצום הוא תהליך.

כל ספירה מייצגת תכונה אלוהית.

תבונה היסטורית היא מושג יסודי ביצירתו של אורטגה.
EOF

# --- Source: no frontmatter ---
cat > "${QA_TMP}/source-nofm.md" <<'EOF'
# El Concepto de Tzimtzum

El concepto de tzimtzum es central.
EOF

# --- Target: no frontmatter ---
# Uses standalone צמצום (no ה prefix) so it matches the glossary term.
cat > "${QA_TMP}/target-nofm.md" <<'EOF'
# מושג צמצום

צמצום הוא מרכזי.
EOF

# --- Target: Hebrew with multiple violations for ordering test ---
# כיווץ (deprecated for tzimtzum) and ספירא (superseded for sefirah)
# are positional violations. צמצום, ספירה, and תבונה היסטורית are all
# absent, producing three "missing" violations sorted by concept_id.
cat > "${QA_TMP}/target-multi.md" <<'EOF'
---
lang: he
---
# טקסט בדיקה

כיווץ הוא מושג מרכזי במחשבה הקבלית.

כל ספירא מייצגת תכונה אלוהית.

הרעיון ההיסטורי הוא מושג יסודי.
EOF

# --- Scan fixture: file with frontmatter ---
cat > "${QA_TMP}/scan-fm.md" <<'EOF'
---
lang: he
---
# טקסט עברי

מושג צמצום הוא מרכזי. כל ספירה מייצגת תכונה.
EOF

# --- Scan fixture: file without frontmatter ---
# Contains English-only "historical reason" and Spanish-only "razón histórica"
# to verify multi-language matching when no --lang filter is applied.
cat > "${QA_TMP}/scan-nofm.md" <<'EOF'
# Mixed Text

The concept of tzimtzum is central. Each sefirah represents an attribute.
The historical reason is a key concept.
El concepto de razón histórica es fundamental.
EOF
````

If any setup step fails, **stop**: build or fixture tree is broken.

## Entry criteria

- All E6 tickets closed.
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

| Risk                                                  | Mitigation                                             |
| ----------------------------------------------------- | ------------------------------------------------------ |
| Language resolution silently picks wrong language      | TC-CHK-LANG-001/002/003/004 test each precedence level |
| Missing violation not emitted for absent concept       | TC-CHK-MISS-001 verifies missing detection             |
| Forbidden variant position (line/col) off-by-one       | TC-CHK-FORB-001 spot-checks reported positions         |
| Admitted variant not promoted under --strict           | TC-CHK-STRICT-001/002 verify both modes                |
| Violation ordering breaks on mixed positional+missing  | TC-CHK-ORD-001 verifies sort contract                  |
| Frontmatter detection breaks scan's existing behavior  | TC-SCAN-FM-001/002/003 verify scan frontmatter update  |
| Check envelope diverges from spec                      | TC-CHK-ENV-001/002 verify keys and types               |
| Code regions stripped asymmetrically                   | TC-CHK-CODE-001 verifies symmetric stripping           |
| Hebrew ה prefix fails word-boundary match              | All Hebrew fixtures use standalone terms (no prefix)    |

## Conventions

Same as E2/E3/E4/E5 QA: see [E2-manual-qa.md](E2-manual-qa.md) §Conventions for
envelope shapes, exit code map, and test case format.

### Exit code map (E6 surface)

| Code | Meaning     | Source                                                          |
| ---- | ----------- | --------------------------------------------------------------- |
| 0    | success     | Check clean (no violations), scan completed                     |
| 1    | violations  | Check found violations (recoverable)                            |
| 2    | usage error | `no_tbx_path`, `language_required`, `invalid_field`             |
| 3    | I/O error   | File not readable                                               |

---

# 1. Check — clean check (no violations)

## TC-CHK-001 — clean check exits 0 with ok true P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-clean.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.violations == []' out.json
jq -e '.schema_version == 1' out.json
```

- **exit**: `0`
- `ok` is `true`.
- `violations` is `[]`.
- `schema_version` is `1`.

## TC-CHK-002 — clean check with frontmatter language detection P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-clean.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- Languages resolved from frontmatter (`lang: es` and `lang: he`), no flags needed.

---

# 2. Check — missing violation

## TC-CHK-MISS-001 — missing preferred target term P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-missing.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == false' out.json
jq -e '.violations | length > 0' out.json
jq -e '[.violations[] | select(.type == "missing")] | length > 0' out.json
# "razón histórica" is in source, "תבונה היסטורית" is missing in target
MISSING=$(jq '[.violations[] | select(.type == "missing")]' out.json)
echo "missing violations: ${MISSING}"
jq -e '[.violations[] | select(.type == "missing")][0] | has("concept_id", "source_term", "expected_target", "source_occurrences")' out.json
```

- **exit**: `1`
- `ok` is `false`.
- At least one `missing` violation present.
- Missing violation carries `concept_id`, `source_term`, `expected_target`, `source_occurrences`.

## TC-CHK-MISS-002 — concepts absent from source are ignored P1

```sh
# Create a source that only mentions tzimtzum, not sefirah or razón histórica.
cat > "${QA_TMP}/source-partial.md" <<'EOF'
---
lang: es
---
# Solo Tzimtzum

El concepto de tzimtzum es central.
EOF

$TT check "${QA_TMP}/source-partial.md" "${QA_TMP}/target-clean.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# sefirah and razón histórica should NOT produce violations since they're
# not in the source.
jq -e '[.violations[] | select(.concept_id == "sefirah")] | length == 0' out.json
jq -e '[.violations[] | select(.concept_id == "razon-historica")] | length == 0' out.json
```

- **exit**: `0`
- Concepts not present in source produce no violations.

---

# 3. Check — forbidden variant

## TC-CHK-FORB-001 — deprecated variant in target P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-forbidden.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == false' out.json
jq -e '[.violations[] | select(.type == "forbidden_variant")] | length > 0' out.json
# כיווץ is deprecated for tzimtzum
FORB=$(jq '[.violations[] | select(.type == "forbidden_variant")][0]' out.json)
echo "forbidden: ${FORB}"
jq -e '[.violations[] | select(.type == "forbidden_variant")][0] | has("concept_id", "variant", "line", "column", "context")' out.json
```

- **exit**: `1`
- At least one `forbidden_variant` violation.
- Violation carries `concept_id`, `variant`, `line`, `column`, `context`.

## TC-CHK-FORB-002 — superseded variant in target P1

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-superseded.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == false' out.json
# ספירא is superseded for sefirah
jq -e '[.violations[] | select(.type == "forbidden_variant" and .concept_id == "sefirah")] | length > 0' out.json
```

- **exit**: `1`
- Superseded variant `ספירא` produces a `forbidden_variant` violation.

## TC-CHK-FORB-003 — forbidden variant line/column accuracy P1

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-forbidden.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
FORB_LINE=$(jq '[.violations[] | select(.type == "forbidden_variant")][0].line' out.json)
FORB_COL=$(jq '[.violations[] | select(.type == "forbidden_variant")][0].column' out.json)
echo "forbidden at line=${FORB_LINE} col=${FORB_COL}"
# Verify the line in the target file contains the variant.
ACTUAL_LINE=$(sed -n "${FORB_LINE}p" "${QA_TMP}/target-forbidden.md")
echo "source line: ${ACTUAL_LINE}"
echo "${ACTUAL_LINE}" | grep -q "כיווץ" && echo "variant found on reported line: ok"
```

- Reported `line` in target file contains the forbidden variant.
- Column position is plausible.

---

# 4. Check — --strict and admitted variants

## TC-CHK-STRICT-001 — admitted variant as warning (non-strict) P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-admitted.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.violations == []' out.json
jq -e '.warnings | length > 0' out.json
jq -e '.warnings[0].type == "admitted_variant"' out.json
jq -e '.warnings[0] | has("concept_id", "variant", "line", "column")' out.json
```

- **exit**: `0` (admitted is a warning, not a violation, without `--strict`).
- `violations` is `[]` — no violations because all preferred terms are present.
- `warnings` contains at least one `admitted_variant` for `התכווצות`.
- Admitted variant warning carries positional data (`line`, `column`).

## TC-CHK-STRICT-002 — admitted variant as violation (strict) P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-admitted.md" \
  --tbx "${QA_TMP}/glossary.tbx" --strict >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == false' out.json
jq -e '[.violations[] | select(.type == "admitted_variant")] | length > 0' out.json
# התכווצות is admitted for tzimtzum
ADMITTED=$(jq '[.violations[] | select(.type == "admitted_variant")][0]' out.json)
echo "admitted violation: ${ADMITTED}"
```

- **exit**: `1`
- `admitted_variant` violation present.

## TC-CHK-STRICT-003 — strict does not affect missing/forbidden detection P1

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-forbidden.md" \
  --tbx "${QA_TMP}/glossary.tbx" --strict >out.json 2>err.json
echo "exit=$?"
jq -e '[.violations[] | select(.type == "forbidden_variant")] | length > 0' out.json
```

- **exit**: `1`
- `--strict` does not suppress or alter `forbidden_variant` detection.

---

# 5. Check — language resolution

## TC-CHK-LANG-001 — frontmatter detected for both files P0

```sh
# source.md has lang: es, target-clean.md has lang: he — no flags needed.
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-clean.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- Languages resolved from frontmatter without flags.

## TC-CHK-LANG-002 — flag-based language when no frontmatter P0

```sh
$TT check "${QA_TMP}/source-nofm.md" "${QA_TMP}/target-nofm.md" \
  --tbx "${QA_TMP}/glossary.tbx" \
  --source-lang es --target-lang he >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- Languages resolved from `--source-lang` and `--target-lang` flags.

## TC-CHK-LANG-003 — frontmatter overrides flag P1

```sh
# source.md has lang: es in frontmatter. Pass --source-lang he to conflict.
# Frontmatter should win: source scanned as Spanish.
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-clean.md" \
  --tbx "${QA_TMP}/glossary.tbx" \
  --source-lang he --target-lang he >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- Frontmatter `lang: es` overrides `--source-lang he`.
- Check still finds Spanish source terms and Hebrew target terms.

## TC-CHK-LANG-004 — missing language exits 2 with language_required P0

```sh
$TT check "${QA_TMP}/source-nofm.md" "${QA_TMP}/target-nofm.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "language_required"' err.json
jq -e '.error | has("hint")' err.json
```

- **exit**: `2`
- Error code `language_required`.
- Hint present explaining how to supply the language.

## TC-CHK-LANG-005 — missing source lang only P1

```sh
# Target has frontmatter, source does not, no --source-lang flag.
$TT check "${QA_TMP}/source-nofm.md" "${QA_TMP}/target-clean.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "language_required"' err.json
```

- **exit**: `2`
- Error specifically about source language.

## TC-CHK-LANG-006 — missing target lang only P1

```sh
# Source has frontmatter, target does not, no --target-lang flag.
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-nofm.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "language_required"' err.json
```

- **exit**: `2`
- Error specifically about target language.

---

# 6. Check — violation ordering

## TC-CHK-ORD-001 — positional violations sorted by (line, column), missing at tail P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-multi.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == false' out.json
# target-multi.md has:
#   כיווץ (deprecated for tzimtzum) — forbidden_variant, positional
#   ספירא (superseded for sefirah)  — forbidden_variant, positional
#   missing: צמצום, ספירה, תבונה היסטורית (all absent preferred terms)
# Expect: 2 positional + 3 missing = 5 violations.
# Positional sorted by (line, column), missing at tail by concept_id.
python3 -c "
import json
data = json.load(open('out.json'))
vs = data['violations']
positional = [v for v in vs if v['type'] != 'missing']
missing = [v for v in vs if v['type'] == 'missing']
assert len(positional) >= 2, f'Expected >=2 positional, got {len(positional)}'
assert len(missing) >= 2, f'Expected >=2 missing, got {len(missing)}'
# Positional sorted by (line, column)
pos_keys = [(v['line'], v['column']) for v in positional]
assert pos_keys == sorted(pos_keys), f'Positional not sorted: {pos_keys}'
print(f'positional sorted: ok ({len(positional)} violations)')
# Missing at tail
if positional and missing:
    assert vs.index(missing[0]) > vs.index(positional[-1]), 'Missing not at tail'
    print('missing at tail: ok')
# Missing sorted by concept_id
if len(missing) > 1:
    ids = [v['concept_id'] for v in missing]
    assert ids == sorted(ids), f'Missing not sorted by concept_id: {ids}'
    print('missing sorted by concept_id: ok')
print('PASS')
"
```

- Positional violations (`forbidden_variant`) sorted by `(line, column)`.
- `missing` violations appear after all positional violations.
- Within missing group, sorted by `concept_id` ASCII.

---

# 7. Check — context window

## TC-CHK-CTX-001 — default context on forbidden variant P1

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-forbidden.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
# Default --context is 80 (40 chars per side).
FORB=$(jq '[.violations[] | select(.type == "forbidden_variant")][0]' out.json)
CTX_LEN=$(echo "${FORB}" | jq '.context | length')
VARIANT_LEN=$(echo "${FORB}" | jq '.variant | length')
echo "context_len=${CTX_LEN} variant_len=${VARIANT_LEN}"
OVERHEAD=$((CTX_LEN - VARIANT_LEN))
echo "overhead=${OVERHEAD}"
test "${OVERHEAD}" -le 86 && echo "within default budget: ok"
```

- Context overhead (context length minus variant length) at most 86 (80 + 6 for ellipsis).

## TC-CHK-CTX-002 — custom context window P2

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-forbidden.md" \
  --tbx "${QA_TMP}/glossary.tbx" --context 40 >out.json 2>err.json
echo "exit=$?"
FORB=$(jq '[.violations[] | select(.type == "forbidden_variant")][0]' out.json)
CTX_LEN=$(echo "${FORB}" | jq '.context | length')
VARIANT_LEN=$(echo "${FORB}" | jq '.variant | length')
OVERHEAD=$((CTX_LEN - VARIANT_LEN))
echo "overhead=${OVERHEAD}"
test "${OVERHEAD}" -le 46 && echo "within custom budget: ok"
```

- With `--context 40`, overhead at most 46 (40 + 6).

---

# 8. Check — --fields projection

## TC-CHK-FLD-001 — project specific violation fields P1

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-forbidden.md" \
  --tbx "${QA_TMP}/glossary.tbx" \
  --fields violations.concept_id,violations.type >out.json 2>err.json
echo "exit=$?"
jq -e '.violations[0] | has("concept_id", "type")' out.json
jq -e '.violations[0] | has("context") | not' out.json
jq -e '.violations[0] | has("variant") | not' out.json
```

- Only `concept_id` and `type` present in projected violations.
- Other fields absent.

## TC-CHK-FLD-002 — invalid field returns error P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-clean.md" \
  --tbx "${QA_TMP}/glossary.tbx" \
  --fields concpet_id >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "invalid_field"' err.json
```

- **exit**: `2`
- Error code `invalid_field`.

---

# 9. Check — envelope shape

## TC-CHK-ENV-001 — envelope has correct top-level keys P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-clean.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e 'has("schema_version", "ok", "source", "target", "violations", "warnings", "summary")' out.json
jq -e '.schema_version == 1' out.json
jq -e '.ok == true' out.json
jq -e '.source | type == "string"' out.json
jq -e '.target | type == "string"' out.json
jq -e '.violations | type == "array"' out.json
jq -e '.warnings | type == "array"' out.json
```

- Top-level keys: `schema_version`, `ok`, `source`, `target`, `violations`, `warnings`, `summary`.
- Types correct.

## TC-CHK-ENV-002 — violation field shapes match spec P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-forbidden.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
# forbidden_variant shape
jq -e '[.violations[] | select(.type == "forbidden_variant")][0] |
  has("type", "concept_id", "variant", "line", "column", "context")' out.json
jq -e '[.violations[] | select(.type == "forbidden_variant")][0].line | type == "number"' out.json
jq -e '[.violations[] | select(.type == "forbidden_variant")][0].column | type == "number"' out.json
```

- `forbidden_variant` carries: `type`, `concept_id`, `variant`, `line`, `column`, `context`.

## TC-CHK-ENV-003 — missing violation field shape P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-missing.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
jq -e '[.violations[] | select(.type == "missing")][0] |
  has("type", "concept_id", "source_term", "expected_target", "source_occurrences")' out.json
jq -e '[.violations[] | select(.type == "missing")][0].source_occurrences | type == "number"' out.json
```

- `missing` carries: `type`, `concept_id`, `source_term`, `expected_target`, `source_occurrences`.

## TC-CHK-ENV-004 — summary shape P1

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-forbidden.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
jq -e '.summary | has("violations", "warnings", "concepts_checked")' out.json
jq -e '.summary.violations | type == "number"' out.json
jq -e '.summary.warnings | type == "number"' out.json
jq -e '.summary.concepts_checked | type == "number"' out.json
jq -e '.summary.violations == (.violations | length)' out.json
```

- Summary has `violations`, `warnings`, `concepts_checked`.
- `summary.violations` equals the length of the `violations` array.

---

# 10. Check — error cases

## TC-CHK-ERR-001 — no TBX path P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-clean.md" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "no_tbx_path"' err.json
```

- **exit**: `2`
- Error code `no_tbx_path`.

## TC-CHK-ERR-002 — source file not found P0

```sh
$TT check /nonexistent/source.md "${QA_TMP}/target-clean.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `3` (I/O error).

## TC-CHK-ERR-003 — target file not found P0

```sh
$TT check "${QA_TMP}/source.md" /nonexistent/target.md \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `3` (I/O error).

## TC-CHK-ERR-004 — missing positional arguments P0

```sh
$TT check --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `2` (usage error — missing SRC and TGT).

## TC-CHK-ERR-005 — only one positional argument P2

```sh
$TT check "${QA_TMP}/source.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `2` (usage error — missing TGT).

---

# 11. Check — TERMINOLOGY_TBX env var

## TC-CHK-ENVVAR-001 — TBX resolved from environment variable P1

```sh
TERMINOLOGY_TBX="${QA_TMP}/glossary.tbx" \
  $TT check "${QA_TMP}/source.md" "${QA_TMP}/target-clean.md" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- TBX path resolved from `TERMINOLOGY_TBX` when `--tbx` is not provided.

---

# 12. Check — code region stripping

## TC-CHK-CODE-001 — terms in code blocks ignored symmetrically P1

```sh
# Source with term in code block only.
cat > "${QA_TMP}/source-code.md" <<'SRCEOF'
---
lang: es
---
# Notas técnicas

El concepto de tzimtzum es central.

```python
sefirah = compute_sefirah()
```
SRCEOF

# Target with sefirah only in code block.
cat > "${QA_TMP}/target-code.md" <<'TGTEOF'
---
lang: he
---
# הערות טכניות

צמצום הוא מושג מרכזי.

```python
sefirah = compute_sefirah()
```
TGTEOF

$TT check "${QA_TMP}/source-code.md" "${QA_TMP}/target-code.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# sefirah in code blocks should not count as source occurrences,
# so no missing violation for sefirah concept.
jq -e '[.violations[] | select(.concept_id == "sefirah")] | length == 0' out.json
```

- **exit**: `0`
- Code block terms ignored in both SRC and TGT symmetrically.
- No false-positive `missing` for terms only in code.

---

# 13. Check — stream routing

## TC-CHK-STREAM-001 — success envelope on stdout, nothing on stderr P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-clean.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
test -s out.json && echo "stdout: non-empty"
test ! -s err.json && echo "stderr: empty"
```

- stdout has the JSON envelope.
- stderr is empty on success.

## TC-CHK-STREAM-002 — violation envelope on stdout, summary on stderr P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-forbidden.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
test -s out.json && echo "stdout: non-empty"
test -s err.json && echo "stderr: non-empty"
jq -e '.error.code == "violations"' err.json
```

- Violations produce exit 1 with the full results envelope on stdout.
- stderr carries a short violation summary envelope (`"code": "violations"`).

## TC-CHK-STREAM-003 — error envelope on stderr, nothing on stdout P0

```sh
$TT check "${QA_TMP}/source.md" "${QA_TMP}/target-clean.md" >out.json 2>err.json
echo "exit=$?"
test ! -s out.json && echo "stdout: empty"
test -s err.json && echo "stderr: non-empty"
```

- Error envelope on stderr.
- stdout is empty on error.

---

# 14. Scan — frontmatter language resolution

## TC-SCAN-FM-001 — scan respects frontmatter lang P0

```sh
$TT scan "${QA_TMP}/scan-fm.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# File has lang: he in frontmatter — should only match Hebrew terms.
jq -e '.matches | length > 0' out.json
jq -e '[.matches[].lang] | unique == ["he"]' out.json
```

- **exit**: `0`
- Only Hebrew matches — frontmatter `lang: he` restricts the scan.

## TC-SCAN-FM-002 — scan falls back to --lang flag P0

```sh
$TT scan "${QA_TMP}/scan-nofm.md" --tbx "${QA_TMP}/glossary.tbx" --lang es >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '[.matches[].lang] | unique == ["es"]' out.json
```

- **exit**: `0`
- No frontmatter, `--lang es` used — only Spanish matches.

## TC-SCAN-FM-003 — scan without frontmatter or flag scans all languages P0

```sh
$TT scan "${QA_TMP}/scan-nofm.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.matches | length > 0' out.json
LANGS=$(jq -r '[.matches[].lang] | unique | sort | join(",")' out.json)
echo "languages matched: ${LANGS}"
# Glossary has es and en langSecs — scan-nofm.md has text in both.
jq -e '[.matches[].lang] | unique | length > 1' out.json
```

- **exit**: `0`
- Matches across multiple languages when neither frontmatter nor flag present.

## TC-SCAN-FM-004 — frontmatter overrides --lang flag P1

```sh
# File has lang: he, but pass --lang es.
$TT scan "${QA_TMP}/scan-fm.md" --tbx "${QA_TMP}/glossary.tbx" --lang es >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# Frontmatter wins: should match Hebrew, not Spanish.
jq -e '[.matches[].lang] | unique == ["he"]' out.json
```

- **exit**: `0`
- Frontmatter `lang: he` overrides `--lang es`.

---

# 15. Regression — previous commands still work

## TC-REG-001 — validate unaffected P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.schema_version == 1' out.json
```

- **exit**: `0`
- Validate command still works after E6 changes.

## TC-REG-002 — lookup unaffected P0

```sh
TERM="tzimtzum"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM}" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.results | length > 0' out.json
```

- **exit**: `0`
- Lookup command still works after E6 changes.

## TC-REG-003 — extract unaffected P1

```sh
$TT extract "${QA_TMP}/source.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- Extract command still works.

## TC-REG-004 — scan basic behavior unaffected P0

```sh
$TT scan "${QA_TMP}/scan-nofm.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.matches | length > 0' out.json
```

- **exit**: `0`
- Scan command still works with files that have no frontmatter.

## TC-REG-005 — schema includes check command P1

```sh
$TT schema >out.json 2>err.json
CMDS=$(jq -r '[.commands[].name] | sort | join(",")' out.json)
echo "commands=${CMDS}"
echo "${CMDS}" | grep -q "check" && echo "check: ok"
echo "${CMDS}" | grep -q "scan" && echo "scan: ok"
```

- Schema output includes both `scan` and `check` commands.

---

# Cleanup

```sh
rm -rf "${QA_TMP}"
rm -f out.json err.json
```

---

# Test case summary

| Section                               | Cases  | Priority |
| ------------------------------------- | ------ | -------- |
| Check — clean check                   | 2      | P0       |
| Check — missing violation             | 2      | P0–P1   |
| Check — forbidden variant             | 3      | P0–P1   |
| Check — --strict + admitted           | 3      | P0–P1   |
| Check — language resolution           | 6      | P0–P1   |
| Check — violation ordering            | 1      | P0       |
| Check — context window                | 2      | P1–P2   |
| Check — --fields projection           | 2      | P0–P1   |
| Check — envelope shape                | 4      | P0–P1   |
| Check — error cases                   | 5      | P0–P2   |
| Check — TERMINOLOGY_TBX env var       | 1      | P1       |
| Check — code region stripping         | 1      | P1       |
| Check — stream routing                | 3      | P0       |
| Scan — frontmatter language           | 4      | P0–P1   |
| Regression — previous commands        | 5      | P0–P1   |
| **Total**                             | **44** |          |
