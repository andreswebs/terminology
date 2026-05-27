# Test Plan: E5 — Matcher manual QA

> **Status**: ready to execute once E5 is completed
> Run end-to-end in a single sitting.

## Purpose

E5's deliverable is the **matcher pipeline** and the **scan command** that
exposes it. The matcher is the algorithmic core: Aho-Corasick multi-pattern
matching over canonical pre-normalized text, with word-boundary validation,
longest-match-at-same-start filtering, and status tagging.

This is **matcher + scan QA**. E4 QA already covered `lookup`, `schema`,
`extract`, and `--fields` projection. E5 tests the normalization pipeline
(case-fold, niqqud strip, whitespace collapse), boundary checking, match
filtering, the scan command surface (`--lang`, `--context`, `--fields`),
and the scan envelope shape.

## Scope

### In scope

- `scan FILE` — find all glossary term occurrences in a markdown file.
- Case-insensitive matching (Unicode default case-fold).
- Hebrew niqqud stripping (vowel points removed from both glossary and text).
- Diacritics strict (accented characters must match exactly).
- Multi-word term matching across whitespace/line breaks.
- Code block exclusion (fenced + inline) via `internal/markdown`.
- Word-boundary validation (`\p{L}` and `\p{N}` as word characters).
- Longest-match-at-same-start filtering.
- Status tagging (preferred/admitted/deprecated/superseded/unspecified).
- `--lang` language filter.
- `--context N` context window control (default 80).
- `--fields` projection on scan output.
- Scan envelope shape (`file`, `matches`, `summary`).
- Exit codes: 0 (always for successful scans), 2 (usage error), 3 (I/O error).
- Error envelopes: `no_tbx_path`, `invalid_field`, file-not-found.

### Out of scope

- Validation logic (`validate` command) — covered by E3 QA.
- Read commands (`lookup`, `schema`, `extract`) — covered by E4 QA.
- Check command (`check SRC TGT`) — E6.
- Write commands (`concept`, `term`, `apply`) — E7/E8.
- CJK/Thai segmentation — explicitly out of v1 scope.
- Performance budget — deferred to E9.
- `--format text` rendering — tested only for JSON in E5.

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

Create test markdown and TBX files for scan tests:

````sh
export QA_TMP=$(mktemp -d /tmp/e5-qa-XXXXXX)

# --- Glossary fixture ---
# Multi-language glossary with various term statuses.
cat > "${QA_TMP}/glossary.tbx" <<'TBXEOF'
<?xml version="1.0" encoding="UTF-8"?>
<?xml-model href="https://raw.githubusercontent.com/LTAC-Global/TBX-Linguist_Module/master/Schema/TBXcheckerTBX-Linguist.sch" type="application/xml" schematypens="http://purl.oclc.org/dml/schematron"?>
<tbx style="dct" type="TBX-Linguist" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader>
    <fileDesc><sourceDesc><p>E5 QA fixture</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="tzimtzum">
      <min:subjectField>kabbalah</min:subjectField>
      <langSec xml:lang="en">
        <termSec>
          <term>tzimtzum</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
        <termSec>
          <term>contraction</term>
          <min:administrativeStatus>deprecatedTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
      <langSec xml:lang="he">
        <termSec>
          <term>צמצום</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
      <langSec xml:lang="es">
        <termSec>
          <term>tzimtzum</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
    <conceptEntry id="tzimtzum-primordial">
      <min:subjectField>kabbalah</min:subjectField>
      <langSec xml:lang="en">
        <termSec>
          <term>tzimtzum primordial</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
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
        <termSec>
          <term>sephirah</term>
          <min:administrativeStatus>admittedTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
      <langSec xml:lang="he">
        <termSec>
          <term>סְפִירָה</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
    <conceptEntry id="razon-historica">
      <min:subjectField>philosophy</min:subjectField>
      <langSec xml:lang="es">
        <termSec>
          <term>razón histórica</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
TBXEOF

# --- Corpus: happy-path with multiple terms ---
cat > "${QA_TMP}/corpus.md" <<'EOF'
# The Concept of Tzimtzum

The concept of tzimtzum is central to Kabbalistic thought. It describes
the divine contraction that preceded creation.

In Hebrew texts, the term צמצום appears frequently. The idea of
tzimtzum primordial extends the basic concept further.

Each sefirah represents a divine attribute. The sephirah system
organizes these attributes into a tree structure.
EOF

# --- Corpus: empty (no glossary terms) ---
cat > "${QA_TMP}/empty.md" <<'EOF'
# Ordinary Text

This document contains no terminology from the glossary.
Just plain everyday language about nothing in particular.
EOF

# --- Corpus: terms only inside code blocks ---
cat > "${QA_TMP}/code-only.md" <<'EOF'
# Technical Notes

Some implementation details:

```python
def tzimtzum_algorithm():
    sefirah = compute_sefirah()
    return sefirah
```

The `tzimtzum` variable name is used in the code above.
No glossary terms appear in regular prose here.
EOF

# --- Corpus: case variants ---
cat > "${QA_TMP}/case-mix.md" <<'EOF'
# Mixed Case

The term TZIMTZUM appears in uppercase. Later, Tzimtzum appears
title-cased. Finally tzimtzum in lowercase.
EOF

# --- Corpus: Hebrew with niqqud ---
cat > "${QA_TMP}/niqqud.md" <<'EOF'
# Hebrew Text

The term סְפִירָה appears with niqqud vowel points.
The plain form ספירה should also be matched.
The concept of צמצום is central.
EOF

# --- Corpus: multi-word across line break ---
cat > "${QA_TMP}/linebreak.md" <<'EOF'
# Line Break Test

The concept of tzimtzum
primordial extends the basic idea of divine contraction.
EOF

# --- Corpus: boundary edge cases ---
cat > "${QA_TMP}/boundary.md" <<'EOF'
# Boundary Tests

The word tzimtzum appears normally.
But pretzimtzumx is not a valid match (embedded in word).
And (tzimtzum) should match inside parentheses.
Also "tzimtzum" inside quotes.
And 3tzimtzum should not match (digit boundary).
EOF

# --- Corpus: Spanish with accented term ---
cat > "${QA_TMP}/spanish.md" <<'EOF'
# Texto Filosófico

La razón histórica es un concepto fundamental en la obra de Ortega.
La razon historica sin acentos es diferente.
EOF

# --- Corpus: status tagging ---
cat > "${QA_TMP}/deprecated.md" <<'EOF'
# Terminology Usage

The contraction of the divine light created space. The tzimtzum
concept explains this process.
EOF
````

If any setup step fails, **stop**: build or fixture tree is broken.

## Entry criteria

- All E5 tickets closed.
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

| Risk                                               | Mitigation                                            |
| -------------------------------------------------- | ----------------------------------------------------- |
| Case-fold produces wrong canonical for Hebrew      | TC-SCAN-CASE-001 tests uppercase matching             |
| Niqqud stripping misses combining mark ranges      | TC-SCAN-NIQ-001 tests niqqud vs plain form            |
| Diacritics incorrectly folded (should be strict)   | TC-SCAN-DIAC-001 verifies razón ≠ razon               |
| Multi-word match fails across line breaks          | TC-SCAN-MW-001 tests term spanning two lines          |
| Code block terms leak into matches                 | TC-SCAN-CODE-001 verifies code-only corpus empty      |
| Boundary check on original text fails cross-script | TC-SCAN-BND-001/002 test punctuation and digit bounds |
| Longest-match filter drops the wrong match         | TC-SCAN-LM-001 verifies longer phrase wins            |
| Offset map produces wrong line/column              | TC-SCAN-POS-001 spot-checks reported positions        |
| Context window truncation or off-by-one            | TC-SCAN-CTX-001/002 verify per-side padding semantics |
| Scan envelope shape diverges from spec             | TC-SCAN-ENV-001/002 verify keys and types             |

## Conventions

Same as E2/E3/E4 QA: see [E2-manual-qa.md](E2-manual-qa.md) §Conventions for
envelope shapes, exit code map, and test case format.

### Exit code map (E5 surface)

| Code | Meaning     | Source                                                  |
| ---- | ----------- | ------------------------------------------------------- |
| 0    | success     | Scan completed (even with zero matches — informational) |
| 2    | usage error | `no_tbx_path`, `invalid_field`                          |
| 3    | I/O error   | File not readable                                       |

---

# 1. Scan — basic matching

## TC-SCAN-001 — glossary terms found in corpus P0

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.matches | length > 0' out.json
jq -e '.schema_version == 1' out.json
jq -e '[.matches[].term] | contains(["tzimtzum"])' out.json
```

- **exit**: `0`
- `ok` is `true`.
- `schema_version` is `1`.
- `matches` array is non-empty.
- `tzimtzum` appears among matched terms.

## TC-SCAN-002 — no matches returns empty array P0

```sh
$TT scan "${QA_TMP}/empty.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.matches == []' out.json
jq -e '.summary.total_matches == 0' out.json
jq -e '.summary.unique_concepts == 0' out.json
```

- **exit**: `0` (scan is informational, always exit 0).
- `matches` is `[]` (empty array, not null).
- Summary counts are zero.

## TC-SCAN-003 — multiple concepts matched P1

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
CONCEPTS=$(jq -r '[.matches[].concept_id] | unique | sort | join(",")' out.json)
echo "concepts=${CONCEPTS}"
echo "${CONCEPTS}" | grep -q "tzimtzum" && echo "tzimtzum: ok"
echo "${CONCEPTS}" | grep -q "sefirah" && echo "sefirah: ok"
```

- Both `tzimtzum` and `sefirah` concepts found.
- `summary.unique_concepts >= 2`.

---

# 2. Scan — case-insensitive matching

## TC-SCAN-CASE-001 — uppercase term matched P0

```sh
$TT scan "${QA_TMP}/case-mix.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
MATCH_COUNT=$(jq '[.matches[] | select(.concept_id == "tzimtzum")] | length' out.json)
echo "tzimtzum matches=${MATCH_COUNT}"
test "${MATCH_COUNT}" -ge 3 && echo "all case variants matched"
```

- **exit**: `0`
- `TZIMTZUM`, `Tzimtzum`, and `tzimtzum` all produce matches.
- At least 3 matches for concept `tzimtzum`.

---

# 3. Scan — Hebrew niqqud stripping

## TC-SCAN-NIQ-001 — niqqud-bearing term matches plain form P0

```sh
$TT scan "${QA_TMP}/niqqud.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.matches | length > 0' out.json
# The glossary has סְפִירָה (with niqqud); corpus has both forms.
# Both should match because niqqud is stripped from both sides.
HE_MATCHES=$(jq '[.matches[] | select(.concept_id == "sefirah" and .lang == "he")] | length' out.json)
echo "sefirah(he) matches=${HE_MATCHES}"
test "${HE_MATCHES}" -ge 2 && echo "both niqqud and plain form matched"
```

- **exit**: `0`
- Glossary term `סְפִירָה` (with niqqud) matches both `סְפִירָה` and `ספירה`
  in the corpus.
- At least 2 Hebrew sefirah matches.

---

# 4. Scan — diacritics strict

## TC-SCAN-DIAC-001 — accented term matches only exact form P1

```sh
$TT scan "${QA_TMP}/spanish.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
# "razón histórica" (with accents) should match.
jq -e '[.matches[].concept_id] | contains(["razon-historica"])' out.json
MATCH_COUNT=$(jq '[.matches[] | select(.concept_id == "razon-historica")] | length' out.json)
echo "razon-historica matches=${MATCH_COUNT}"
# Only the accented form should match (diacritics strict).
test "${MATCH_COUNT}" -eq 1 && echo "strict diacritics: only exact form matched"
```

- **exit**: `0`
- `razón histórica` (accented) matches.
- `razon historica` (unaccented) does NOT match.
- Exactly 1 match for `razon-historica`.

---

# 5. Scan — code block exclusion

## TC-SCAN-CODE-001 — terms inside code blocks not matched P0

```sh
$TT scan "${QA_TMP}/code-only.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.matches == []' out.json
```

- **exit**: `0`
- `matches` is empty — `tzimtzum` and `sefirah` inside fenced code block
  and inline code are excluded.

## TC-SCAN-CODE-002 — prose terms found alongside code blocks P1

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
jq -e '.matches | length > 0' out.json
```

- Terms in prose paragraphs are found even when the file also contains
  code blocks (corpus.md has no code blocks, so all terms are in prose).

---

# 6. Scan — word boundary validation

## TC-SCAN-BND-001 — embedded substring not matched P0

```sh
$TT scan "${QA_TMP}/boundary.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
# "pretzimtzumx" should NOT produce a match (embedded in word).
LINES=$(jq '[.matches[] | select(.concept_id == "tzimtzum") | .line]' out.json)
echo "match lines=${LINES}"
# Verify the line with "pretzimtzumx" is NOT in the match set.
# That line should be around line 4 in boundary.md.
jq -e '[.matches[] | select(.concept_id == "tzimtzum")] | all(.context | test("pretzimtzumx") | not)' out.json
```

- **exit**: `0`
- `pretzimtzumx` does NOT produce a match — boundary check rejects it.

## TC-SCAN-BND-002 — punctuation boundaries accepted P0

```sh
$TT scan "${QA_TMP}/boundary.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
# "(tzimtzum)" and '"tzimtzum"' should match.
PAREN_MATCH=$(jq '[.matches[] | select(.context | test("\\(tzimtzum\\)"))] | length' out.json)
QUOTE_MATCH=$(jq '[.matches[] | select(.context | test("\"tzimtzum\""))] | length' out.json)
echo "paren=${PAREN_MATCH} quote=${QUOTE_MATCH}"
test "${PAREN_MATCH}" -ge 1 && echo "parentheses boundary: ok"
test "${QUOTE_MATCH}" -ge 1 && echo "quote boundary: ok"
```

- Parentheses and quotes are valid word boundaries.

## TC-SCAN-BND-003 — digit boundary rejects match P1

```sh
$TT scan "${QA_TMP}/boundary.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
# "3tzimtzum" should NOT match (digit is a word character).
jq -e '[.matches[] | select(.context | test("3tzimtzum"))] | length == 0' out.json
```

- `3tzimtzum` does NOT match — `\p{N}` is a word character.

---

# 7. Scan — multi-word term matching

## TC-SCAN-MW-001 — multi-word term across line break P1

```sh
$TT scan "${QA_TMP}/linebreak.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '[.matches[].concept_id] | contains(["tzimtzum-primordial"])' out.json
```

- **exit**: `0`
- `tzimtzum primordial` matched despite spanning a line break.
- Whitespace collapse (`\s+` → single space) enables this.

---

# 8. Scan — longest-match-at-same-start

## TC-SCAN-LM-001 — longer phrase wins over shorter P1

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
# Where "tzimtzum primordial" appears, only the longer match should be emitted.
# The shorter "tzimtzum" should NOT also appear at that same position.
PRIM_LINE=$(jq '[.matches[] | select(.concept_id == "tzimtzum-primordial")][0].line' out.json)
echo "tzimtzum-primordial at line=${PRIM_LINE}"
# No plain "tzimtzum" match should share the same line AND column.
PRIM_COL=$(jq '[.matches[] | select(.concept_id == "tzimtzum-primordial")][0].column' out.json)
CONFLICT=$(jq "[.matches[] | select(.concept_id == \"tzimtzum\" and .line == ${PRIM_LINE} and .column == ${PRIM_COL})] | length" out.json)
echo "conflicts at same position=${CONFLICT}"
test "${CONFLICT}" -eq 0 && echo "longest match wins: ok"
```

- **exit**: `0`
- `tzimtzum primordial` emitted at the overlap position.
- No shorter `tzimtzum` match at the same (line, column).

---

# 9. Scan — status tagging

## TC-SCAN-STATUS-001 — deprecated term carries status P0

```sh
$TT scan "${QA_TMP}/deprecated.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
# "contraction" is deprecated in the glossary.
jq -e '[.matches[] | select(.term == "contraction")][0].status == "deprecated"' out.json
# "tzimtzum" is preferred.
jq -e '[.matches[] | select(.term == "tzimtzum")][0].status == "preferred"' out.json
```

- **exit**: `0`
- `contraction` match has `status: "deprecated"`.
- `tzimtzum` match has `status: "preferred"`.

## TC-SCAN-STATUS-002 — admitted term carries status P1

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
# "sephirah" is admitted in the glossary.
jq -e '[.matches[] | select(.term == "sephirah")][0].status == "admitted"' out.json
```

- `sephirah` match has `status: "admitted"`.

---

# 10. Scan — --lang filter

## TC-SCAN-LANG-001 — --lang restricts to Hebrew matches P0

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" --lang he >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# Only Hebrew-language patterns should be compiled.
jq -e '[.matches[].lang] | unique == ["he"]' out.json
```

- **exit**: `0`
- All matches have `lang: "he"`.
- English terms like `tzimtzum` (en) and `sefirah` (en) are NOT matched.

## TC-SCAN-LANG-002 — --lang with no matches in that language P1

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" --lang fr >out.json 2>err.json
echo "exit=$?"
jq -e '.matches == []' out.json
jq -e '.summary.total_matches == 0' out.json
```

- **exit**: `0` (scan always exits 0).
- No matches when language has no glossary terms.

---

# 11. Scan — --context window

## TC-SCAN-CTX-001 — default context window P1

The `--context N` flag controls **per-side** padding: `N/2` characters
before the match and `N/2` characters after. Total context length is
therefore `term_len + N + up to 6 for ellipsis markers`.

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
# Default is 80 (40 chars per side).
# Overhead = context_len - term_len should be <= 86 (80 + 6 for "..." on each side).
python3 -c "
import json
m = json.load(open('out.json'))['matches']
for x in m:
    overhead = len(x['context']) - len(x['term'])
    ok = 'ok' if overhead <= 86 else 'OVER'
    print(f'  term={x[\"term\"]!r:25s} overhead={overhead:3d} {ok}')
print('PASS' if all(len(x['context']) - len(x['term']) <= 86 for x in m) else 'FAIL')
"
```

- **exit**: `0`
- Context overhead (context length minus term length) is at most 86
  (80 chars padding + up to 6 for ellipsis markers).

## TC-SCAN-CTX-002 — custom context window P1

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" --context 40 >out.json 2>err.json
echo "exit=$?"
# --context 40 → 20 chars per side.
# Overhead should be <= 46 (40 + 6 for "..." on each side).
python3 -c "
import json
m = json.load(open('out.json'))['matches']
for x in m:
    overhead = len(x['context']) - len(x['term'])
    ok = 'ok' if overhead <= 46 else 'OVER'
    print(f'  term={x[\"term\"]!r:25s} overhead={overhead:3d} {ok}')
print('PASS' if all(len(x['context']) - len(x['term']) <= 46 for x in m) else 'FAIL')
"
```

- **exit**: `0`
- Context overhead is at most 46 (40 chars padding + up to 6 for ellipsis).

---

# 12. Scan — --fields projection

## TC-SCAN-FLD-001 — project specific match fields P1

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" --fields matches.concept_id,matches.line >out.json 2>err.json
echo "exit=$?"
jq -e '.matches[0] | has("concept_id", "line")' out.json
jq -e '.matches[0] | has("context") | not' out.json
jq -e '.matches[0] | has("term") | not' out.json
```

- **exit**: `0`
- Only `concept_id` and `line` present in projected matches.
- `context`, `term`, `lang`, `status`, `column` absent.

## TC-SCAN-FLD-002 — invalid field returns error P0

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" --fields concpet_id >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "invalid_field"' err.json
```

- **exit**: `2`
- Error code `invalid_field`.

---

# 13. Scan — envelope shape

## TC-SCAN-ENV-001 — envelope has correct top-level keys P0

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e 'has("schema_version", "ok", "file", "matches", "summary")' out.json
jq -e '.schema_version == 1' out.json
jq -e '.ok == true' out.json
jq -e '.file | test("corpus.md")' out.json
```

- **exit**: `0`
- Top-level keys: `schema_version`, `ok`, `file`, `matches`, `summary`.
- `file` contains the input path.

## TC-SCAN-ENV-002 — match shape matches spec P0

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
jq -e '.matches[0] | has("concept_id", "term", "lang", "status", "line", "column", "context")' out.json
jq -e '.matches[0].line | type == "number"' out.json
jq -e '.matches[0].column | type == "number"' out.json
jq -e '.matches[0].concept_id | type == "string"' out.json
```

- Each match has all 7 required fields.
- `line` and `column` are numbers; `concept_id`, `term`, `lang`, `status`,
  `context` are strings.

## TC-SCAN-ENV-003 — summary shape matches spec P0

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
jq -e '.summary | has("total_matches", "unique_concepts")' out.json
jq -e '.summary.total_matches | type == "number"' out.json
jq -e '.summary.unique_concepts | type == "number"' out.json
jq -e '.summary.total_matches == (.matches | length)' out.json
```

- Summary has `total_matches` and `unique_concepts`.
- `total_matches` equals the length of the `matches` array.

## TC-SCAN-ENV-004 — matches sorted by (line, column) P1

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
# Verify matches are sorted by line, then column.
jq -e '[.matches | [range(0; length - 1)] | map(
  .matches[.].line < .matches[. + 1].line or
  (.matches[.].line == .matches[. + 1].line and .matches[.].column <= .matches[. + 1].column)
) | all] // (.matches | length <= 1)' out.json \
  || python3 -c "
import json, sys
m = json.load(open('out.json'))['matches']
pairs = [(x['line'], x['column']) for x in m]
assert pairs == sorted(pairs), f'Not sorted: {pairs}'
print('sorted: ok')
"
```

- Matches are sorted by `(line, column)` per determinism ADR.

---

# 14. Scan — error cases

## TC-SCAN-ERR-001 — no --tbx flag P0

```sh
$TT scan "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "no_tbx_path"' err.json
```

- **exit**: `2`
- Error envelope on stderr with `code: "no_tbx_path"`.

## TC-SCAN-ERR-002 — file not found P0

```sh
$TT scan /nonexistent/file.md --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
test "$?" -eq 3 || echo "exit=$? (expected 3)"
```

- **exit**: `3` (I/O error).
- Error envelope on stderr.

## TC-SCAN-ERR-003 — no file argument P0

```sh
$TT scan --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `2` (usage error — missing positional argument).

---

# 15. Scan — position accuracy

## TC-SCAN-POS-001 — line and column are correct P1

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
# Pick a known match and verify its position.
# "tzimtzum" first appears on line 3 of corpus.md: "The concept of tzimtzum..."
FIRST=$(jq '[.matches[] | select(.term == "tzimtzum")][0]' out.json)
LINE=$(echo "${FIRST}" | jq '.line')
COL=$(echo "${FIRST}" | jq '.column')
echo "first tzimtzum: line=${LINE} col=${COL}"
# Verify by checking the source file.
ACTUAL_LINE=$(sed -n "${LINE}p" "${QA_TMP}/corpus.md")
echo "source line: ${ACTUAL_LINE}"
echo "${ACTUAL_LINE}" | grep -q "tzimtzum" && echo "term found on reported line: ok"
```

- Reported line number points to a line containing the matched term.
- Column position is plausible (manual spot-check).

---

# 16. Scan — TERMINOLOGY_TBX env var

## TC-SCAN-ENV-VAR-001 — TBX from environment variable P1

```sh
TERMINOLOGY_TBX="${QA_TMP}/glossary.tbx" $TT scan "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.matches | length > 0' out.json
```

- **exit**: `0`
- TBX path resolved from `TERMINOLOGY_TBX` when `--tbx` is not provided.

---

# 17. Stream routing

## TC-STREAM-001 — success envelope on stdout, nothing on stderr P0

```sh
$TT scan "${QA_TMP}/corpus.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
test -s out.json && echo "stdout: non-empty"
test ! -s err.json && echo "stderr: empty"
```

- stdout has the JSON envelope.
- stderr is empty on success.

## TC-STREAM-002 — error envelope on stderr, nothing on stdout P0

```sh
$TT scan "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
test ! -s out.json && echo "stdout: empty"
test -s err.json && echo "stderr: non-empty"
```

- stderr has the error envelope.
- stdout is empty on error.

---

# 18. Regression — previous commands still work

## TC-REG-001 — validate unaffected P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.schema_version == 1' out.json
```

- **exit**: `0`
- Validate command still works after E5 changes.

## TC-REG-002 — lookup unaffected P0

```sh
TERM="tzimtzum"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM}" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.results | length > 0' out.json
```

- **exit**: `0`
- Lookup command still works after E5 changes.

## TC-REG-003 — schema includes scan command P1

```sh
$TT schema >out.json 2>err.json
CMDS=$(jq -r '[.commands[].name] | sort | join(",")' out.json)
echo "commands=${CMDS}"
echo "${CMDS}" | grep -q "scan" && echo "scan: ok"
```

- Schema output includes the `scan` command.

---

# Cleanup

```sh
rm -rf "${QA_TMP}"
rm -f out.json err.json out2.json err2.json
```

---

# Test case summary

| Section                        | Cases  | Priority |
| ------------------------------ | ------ | -------- |
| Scan — basic matching          | 3      | P0–P1    |
| Scan — case-insensitive        | 1      | P0       |
| Scan — niqqud stripping        | 1      | P0       |
| Scan — diacritics strict       | 1      | P1       |
| Scan — code block exclusion    | 2      | P0–P1    |
| Scan — word boundary           | 3      | P0–P1    |
| Scan — multi-word matching     | 1      | P1       |
| Scan — longest-match-at-start  | 1      | P1       |
| Scan — status tagging          | 2      | P0–P1    |
| Scan — --lang filter           | 2      | P0–P1    |
| Scan — --context window        | 2      | P1       |
| Scan — --fields projection     | 2      | P0–P1    |
| Scan — envelope shape          | 4      | P0–P1    |
| Scan — error cases             | 3      | P0       |
| Scan — position accuracy       | 1      | P1       |
| Scan — env var TBX             | 1      | P1       |
| Stream routing                 | 2      | P0       |
| Regression — previous commands | 3      | P0–P1    |
| **Total**                      | **35** |          |
