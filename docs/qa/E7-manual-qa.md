# Test Plan: E7 — Write commands manual QA

> **Status**: ready to execute once E7 is completed
> Run end-to-end in a single sitting.

## Purpose

E7's deliverable is **five granular mutation commands** on top of E2's
writer: `concept add`, `concept update ID`, `concept remove ID`,
`term add ID`, `term deprecate ID`. Plus shared write infrastructure:
concept-ID derivation, transaction records, dry-run preview, and
pre-write validation.

This is **write command QA**. E3 QA covered `validate`, E4 QA covered
`lookup`/`schema`/`extract`, E5 QA covered `scan`, E6 QA covered `check`.
E7 tests the mutation surface: flag input, JSON stdin, TBX fragment stdin,
merge vs replace semantics, transaction record placement, concept-ID
derivation and stability, dry-run fidelity, pre-write validation, file
locking, and the full set of write-specific error sentinels.

## Scope

### In scope

- `concept add` — create a new concept (flags, JSON stdin, TBX fragment stdin).
- `concept update ID` — modify existing concept (`--merge` overlay, `--replace` wholesale).
- `concept remove ID` — delete a concept (default refuses on dangling crossref; `--force` overrides).
- `term add ID` — append a term to a concept's language section.
- `term deprecate ID` — set a term's status to `deprecatedTerm-admn-sts`.
- Common write affordances: `--dry-run`, `--transaction`, `--author` / `TERMINOLOGY_AUTHOR`.
- Input layering: flags → JSON stdin → TBX fragment stdin (auto-detected).
- Concept-ID derivation (NFKD → casefold → `[a-z0-9]` → hyphen → truncate 64).
- ID stability: renaming preferred term does NOT re-derive ID.
- Transaction record placement (concept-level for concept commands, termSec-level for term commands).
- Pre-write validation: full E3 pipeline on in-memory result before rename.
- Concurrency: `${TBX}.lock` held across read→modify→write; non-blocking, fails fast.
- Error sentinels: `duplicate_id`, `not_found`, `dangling_crossref`, `invalid_id`, `invalid_input`, `invalid_value`.
- Envelope shape: success outputs concept in lookup-style shape.
- Exit codes: 0 (success), 2 (usage error), 3 (I/O / lock), 65 (validation / data error).

### Out of scope

- Bulk declarative `apply` — E8.
- Matcher pipeline and scan internals — covered by E5 QA.
- Check command — covered by E6 QA.
- Read commands (`lookup`, `schema`, `extract`) — covered by E4 QA.
- Validation logic internals — covered by E3 QA.
- Input hardening (path traversal, control chars) — E9.
- Performance budget — E9.

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

Create a mutable TBX glossary and helper fixtures for write command tests.
Each test that mutates copies the base glossary first.

````sh
export QA_TMP=$(mktemp -d /tmp/e7-qa-XXXXXX)

# --- Base glossary fixture ---
# Three concepts: tzimtzum (en, es, he), sefirah (en, he with crossref),
# razon-historica (en, es). sefirah cross-references tzimtzum.
cat > "${QA_TMP}/glossary.tbx" <<'TBXEOF'
<?xml version="1.0" encoding="UTF-8"?>
<?xml-model href="https://raw.githubusercontent.com/LTAC-Global/TBX-Linguist_Module/master/Schema/TBXcheckerTBX-Linguist.sch" type="application/xml" schematypens="http://purl.oclc.org/dml/schematron"?>
<tbx style="dct" type="TBX-Linguist" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader>
    <fileDesc><sourceDesc><p>E7 QA fixture</p></sourceDesc></fileDesc>
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
      <langSec xml:lang="es">
        <termSec>
          <term>tzimtzum</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
      <langSec xml:lang="he">
        <termSec>
          <term>צמצום</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
    <conceptEntry id="sefirah">
      <min:subjectField>kabbalah</min:subjectField>
      <basic:crossReference target="tzimtzum">related kabbalistic concept</basic:crossReference>
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
          <term>ספירה</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
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
    </conceptEntry>
  </body></text>
</tbx>
TBXEOF

# --- JSON payload: new concept for stdin tests ---
# JSON payload struct = lookup output struct (round-trip).
cat > "${QA_TMP}/new-concept.json" <<'JSONEOF'
{
  "concept_id": "ein-sof",
  "subject_field": "kabbalah",
  "languages": {
    "en": {
      "preferred": {
        "term": "Ein Sof",
        "administrative_status": "preferredTerm-admn-sts"
      }
    },
    "he": {
      "preferred": {
        "term": "אין סוף",
        "administrative_status": "preferredTerm-admn-sts"
      }
    }
  }
}
JSONEOF

# --- JSON payload: concept update merge ---
cat > "${QA_TMP}/update-merge.json" <<'JSONEOF'
{
  "languages": {
    "fr": {
      "preferred": {
        "term": "tsimtsoum",
        "administrative_status": "preferredTerm-admn-sts"
      }
    }
  }
}
JSONEOF

# --- JSON payload: concept update replace ---
cat > "${QA_TMP}/update-replace.json" <<'JSONEOF'
{
  "subject_field": "mysticism",
  "languages": {
    "en": {
      "preferred": {
        "term": "tzimtzum",
        "administrative_status": "preferredTerm-admn-sts"
      }
    }
  }
}
JSONEOF

# --- TBX fragment: bare conceptEntry ---
cat > "${QA_TMP}/fragment.xml" <<'XMLEOF'
<conceptEntry id="malkhut">
  <min:subjectField>kabbalah</min:subjectField>
  <langSec xml:lang="en">
    <termSec>
      <term>malkhut</term>
      <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
    </termSec>
  </langSec>
</conceptEntry>
XMLEOF

# --- TBX fragment: full <tbx> document (should be rejected) ---
cat > "${QA_TMP}/full-tbx-fragment.xml" <<'XMLEOF'
<tbx style="dct" type="TBX-Linguist" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader>
    <fileDesc><sourceDesc><p>Bad fragment</p></sourceDesc></fileDesc>
  </tbxHeader>
  <text><body>
    <conceptEntry id="bad-concept">
      <langSec xml:lang="en">
        <termSec>
          <term>bad</term>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
XMLEOF

# --- Helper: copy glossary for a mutable test ---
work() {
  local DEST="${QA_TMP}/work-$(date +%s%N).tbx"
  cp "${QA_TMP}/glossary.tbx" "${DEST}"
  echo "${DEST}"
}
````

If any setup step fails, **stop**: build or fixture tree is broken.

## Entry criteria

- All E7 tickets closed.
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

| Risk                                                       | Mitigation                                                |
| ---------------------------------------------------------- | --------------------------------------------------------- |
| ID derivation produces wrong slug for non-Latin terms       | TC-ADD-ID-003 tests empty derivation for Hebrew-only      |
| Merge overlays wrong language or drops existing terms       | TC-UPD-MERGE-001/002/003/005/006 test each overlay rule   |
| Merge appends duplicate instead of overlaying existing term | TC-UPD-MERGE-003 tests (Surface, status) term matching    |
| Replace accidentally changes concept ID                     | TC-UPD-REPL-002 verifies ID preserved                    |
| Crossref refusal misses indirect references                 | TC-RM-XREF-001/002 test with/without --force             |
| Transaction placed at wrong scope (concept vs termSec)      | TC-TXN-001/003 verify placement per command type          |
| Preferred-term rename silently re-derives ID                | TC-STAB-001/002 verify ID unchanged after rename          |
| Pre-write validation misses invalid state                   | TC-PREVAL-001 verifies full pipeline runs before rename   |
| File lock not released on error, blocks subsequent writes   | TC-LOCK-001 verifies lock error and recovery              |
| JSON stdin unknown fields silently ignored                  | TC-ERR-005 verifies invalid_input for malformed JSON      |
| TBX fragment accepts full <tbx> document                    | TC-ADD-TBX-003 verifies rejection with invalid_input      |
| Dry-run mutates file despite preview-only intent            | TC-DRY-002 verifies file byte-identical after dry-run     |
| Missing --author with --transaction pollutes envelope       | TC-TXN-002 verifies WARN log, no responsibility in record |
| TERMINOLOGY_AUTHOR env var not resolved                     | TC-TXN-004 verifies env var path for author               |
| term deprecate misreports not_found level (concept vs lang) | TC-TERM-DEP-003 tests nonexistent langSec specifically    |

## Conventions

Same as E2/E3/E4/E5/E6 QA: see [E2-manual-qa.md](E2-manual-qa.md) §Conventions for
envelope shapes, exit code map, and test case format.

### Exit code map (E7 surface)

| Code | Meaning          | Source                                                                     |
| ---- | ---------------- | -------------------------------------------------------------------------- |
| 0    | success          | Write completed (concept created/updated/removed, term added/deprecated)   |
| 2    | usage error      | `no_tbx_path`, merge/replace mutex, missing args                           |
| 3    | I/O error        | File not found, `tbx_locked`                                               |
| 65   | validation error | `duplicate_id`, `not_found`, `dangling_crossref`, `invalid_id`, `invalid_input`, `validation_error` |

### Mutation test pattern

Every test that mutates the glossary **must** copy the base fixture first.
The `work` shell function creates a timestamped copy and prints its path:

```sh
W=$(work)
$TT concept add --tbx "${W}" --lang en --term "new term" --status preferredTerm-admn-sts
```

---

# 1. concept add — flag input

## TC-ADD-001 — basic concept add with flags P0

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --id "tikkun" --subject-field "kabbalah" \
  --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.schema_version == 1' out.json
jq -e '.result.concept_id == "tikkun"' out.json
jq -e '.result.subject_field == "kabbalah"' out.json
```

- **exit**: `0`
- `ok` is `true`.
- Output concept has `concept_id: "tikkun"` and `subject_field: "kabbalah"`.

## TC-ADD-002 — concept add with multiple picklist flags P1

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --id "devekut" --subject-field "kabbalah" \
  --lang en --term "devekut" --status preferredTerm-admn-sts \
  --part-of-speech noun --register "technicalRegister" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.result.concept_id == "devekut"' out.json
```

- **exit**: `0`
- Picklist values accepted and normalized.

## TC-ADD-003 — concept add persists to file P0

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
# Verify the concept is in the file by looking it up.
$TT lookup "tikkun" --tbx "${W}" >lookup.json 2>/dev/null
echo "exit=$?"
jq -e '.ok == true' lookup.json
jq -e '.results | length > 0' lookup.json
jq -e '.results[0].concept_id == "tikkun"' lookup.json
```

- **exit**: `0` for both commands.
- `lookup` confirms the concept was persisted to the TBX file.

---

# 2. concept add — ID derivation

## TC-ADD-ID-001 — auto-derived ID from preferred term P0

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --lang en --term "Divine Light" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# "Divine Light" → NFKD → casefold → "divine-light"
jq -e '.result.concept_id == "divine-light"' out.json
```

- **exit**: `0`
- ID derived as `divine-light` (spaces → hyphen, casefolded).

## TC-ADD-ID-002 — canonical-lang flag controls derivation P1

Multi-lang flag input only keeps the last `--lang`/`--term`/`--status`
triple. Use JSON stdin to supply both languages.

```sh
W=$(work)
echo '{"languages":{"en":{"preferred":{"term":"primordial contraction","administrative_status":"preferredTerm-admn-sts"}},"es":{"preferred":{"term":"contracción primordial","administrative_status":"preferredTerm-admn-sts"}}}}' | \
  $TT concept add --tbx "${W}" --canonical-lang es >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# With --canonical-lang es, ID derived from Spanish preferred term.
# "contracción primordial" → NFKD → drop combining → "contraccion-primordial"
jq -e '.result.concept_id == "contraccion-primordial"' out.json
```

- **exit**: `0`
- ID derived from the Spanish term, not English, due to `--canonical-lang es`.

## TC-ADD-ID-003 — derivation fails for non-Latin-only term P1

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --lang he --term "אור" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "invalid_id"' err.json
```

- **exit**: `65`
- Hebrew-only term produces empty slug after derivation.
- Error `invalid_id` with hint to use `--id`.

---

# 3. concept add — JSON stdin input

## TC-ADD-JSON-001 — concept add from JSON stdin P0

```sh
W=$(work)
cat "${QA_TMP}/new-concept.json" | \
  $TT concept add --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.result.concept_id == "ein-sof"' out.json
# Verify both languages present.
$TT lookup "ein-sof" --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.results[0].concept_id == "ein-sof"' lookup.json
```

- **exit**: `0`
- Concept created from JSON payload with both `en` and `he` terms.

## TC-ADD-JSON-002 — JSON stdin with unknown field rejected P1

```sh
W=$(work)
echo '{"concept_id": "bad", "unknown_field": true, "languages": {}}' | \
  $TT concept add --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "invalid_input"' err.json
```

- **exit**: `65`
- Unknown JSON field produces `invalid_input`.

---

# 4. concept add — TBX fragment stdin

## TC-ADD-TBX-001 — bare conceptEntry accepted P0

TBX fragment stdin is auto-detected; no `--format` flag needed.

```sh
W=$(work)
cat "${QA_TMP}/fragment.xml" | \
  $TT concept add --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.result.concept_id == "malkhut"' out.json
# Verify persisted.
$TT lookup "malkhut" --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `0`
- Bare `<conceptEntry>` accepted and persisted.

## TC-ADD-TBX-002 — conceptEntryList wrapper accepted P1

```sh
W=$(work)
cat <<'XMLEOF' | $TT concept add --tbx "${W}" >out.json 2>err.json
<conceptEntryList>
  <conceptEntry id="binah">
    <min:subjectField>kabbalah</min:subjectField>
    <langSec xml:lang="en">
      <termSec>
        <term>binah</term>
        <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
      </termSec>
    </langSec>
  </conceptEntry>
</conceptEntryList>
XMLEOF
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- `<conceptEntryList>` wrapper accepted.

## TC-ADD-TBX-003 — full tbx document rejected P0

```sh
W=$(work)
cat "${QA_TMP}/full-tbx-fragment.xml" | \
  $TT concept add --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "invalid_input"' err.json
```

- **exit**: `65`
- Full `<tbx>` document rejected with `invalid_input`.

---

# 5. concept update — merge semantics

## TC-UPD-MERGE-001 — merge adds new language section P0

```sh
W=$(work)
# Add French langSec to tzimtzum via JSON stdin merge.
cat "${QA_TMP}/update-merge.json" | \
  $TT concept update tzimtzum --tbx "${W}" --merge >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# French should be present.
$TT lookup tzimtzum --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
# Existing en, es, he langSecs should still be present.
```

- **exit**: `0`
- French langSec added via merge.
- Existing `en`, `es`, `he` langSecs preserved.

## TC-UPD-MERGE-002 — merge preserves absent languages P0

```sh
W=$(work)
# Merge only touches 'fr'; verify 'en' terms untouched.
cat "${QA_TMP}/update-merge.json" | \
  $TT concept update tzimtzum --tbx "${W}" --merge >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# Verify English terms still have both 'tzimtzum' (preferred) and 'contraction' (deprecated).
$TT lookup tzimtzum --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `0`
- English langSec untouched by French-only merge payload.

## TC-UPD-MERGE-003 — merge overlays existing term matched by (Surface, status) P1

```sh
W=$(work)
# tzimtzum has en preferred "tzimtzum". Merge with same surface+status
# but add part-of-speech via JSON stdin.
echo '{"languages":{"en":{"preferred":{"term":"tzimtzum","administrative_status":"preferredTerm-admn-sts","part_of_speech":"noun"}}}}' | \
  $TT concept update tzimtzum --tbx "${W}" --merge >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# The existing preferred "tzimtzum" term should have part_of_speech overlaid,
# not a duplicate appended.
```

- **exit**: `0`
- Existing term matched by `(Surface, status)` has fields overlaid.
- No duplicate term appended.

## TC-UPD-MERGE-004 — merge via TBX fragment stdin P2

```sh
W=$(work)
# Add a French langSec to tzimtzum via TBX fragment (auto-detected).
cat <<'XMLEOF' | $TT concept update tzimtzum --tbx "${W}" --merge >out.json 2>err.json
<conceptEntry id="tzimtzum">
  <langSec xml:lang="fr">
    <termSec>
      <term>tsimtsoum</term>
      <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
    </termSec>
  </langSec>
</conceptEntry>
XMLEOF
echo "exit=$?"
jq -e '.ok == true' out.json
# French should be present, existing langSecs preserved.
$TT lookup tzimtzum --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `0`
- TBX fragment accepted for concept update merge.
- Existing langSecs preserved, French added.

## TC-UPD-MERGE-005 — merge with flag input P1

Flag-based merge requires `--lang`/`--term`/`--status` alongside
`--subject-field`; without them the command attempts to read stdin.

```sh
W=$(work)
$TT concept update tzimtzum --tbx "${W}" --merge \
  --subject-field "mysticism" \
  --lang en --term "tzimtzum" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.result.subject_field == "mysticism"' out.json
# Languages should all still be present.
```

- **exit**: `0`
- Subject field updated, languages preserved.

---

# 6. concept update — replace semantics

## TC-UPD-REPL-001 — replace replaces entire concept content P0

```sh
W=$(work)
cat "${QA_TMP}/update-replace.json" | \
  $TT concept update tzimtzum --tbx "${W}" --replace >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.result.subject_field == "mysticism"' out.json
# After replace, only English should remain (payload only has en).
# Spanish and Hebrew langSecs should be gone.
$TT lookup tzimtzum --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `0`
- Concept content replaced wholesale with payload.
- Original `es` and `he` langSecs removed.

## TC-UPD-REPL-002 — replace preserves concept ID P0

```sh
W=$(work)
cat "${QA_TMP}/update-replace.json" | \
  $TT concept update tzimtzum --tbx "${W}" --replace >out.json 2>err.json
echo "exit=$?"
jq -e '.result.concept_id == "tzimtzum"' out.json
```

- **exit**: `0`
- ID remains `tzimtzum` despite wholesale replacement.

---

# 7. concept update — merge/replace mutex

## TC-UPD-MUTEX-001 — both --merge and --replace is usage error P0

```sh
W=$(work)
$TT concept update tzimtzum --tbx "${W}" --merge --replace \
  --subject-field "kabbalah" >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `2`
- Usage error: `--merge` and `--replace` are mutually exclusive.

## TC-UPD-MUTEX-002 — neither --merge nor --replace is usage error P0

```sh
W=$(work)
$TT concept update tzimtzum --tbx "${W}" \
  --subject-field "kabbalah" >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `2`
- Usage error: exactly one of `--merge` or `--replace` required.

---

# 8. concept remove — basic

## TC-RM-001 — remove concept without crossrefs P0

```sh
W=$(work)
$TT concept remove razon-historica --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# Verify concept is gone.
$TT lookup razon-historica --tbx "${W}" >lookup.json 2>/dev/null
LOOKUP_EXIT=$?
echo "lookup exit=${LOOKUP_EXIT}"
test "${LOOKUP_EXIT}" -eq 1 && echo "not found: ok"
```

- **exit**: `0` for remove.
- `lookup` returns exit `1` (not found) confirming deletion.

## TC-RM-002 — remove nonexistent concept fails P0

```sh
W=$(work)
$TT concept remove nonexistent-id --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "not_found"' err.json
```

- **exit**: `65`
- Error code `not_found`.

---

# 9. concept remove — dangling crossref

## TC-RM-XREF-001 — remove concept with inbound crossref refused P0

```sh
W=$(work)
# sefirah has a crossReference targeting tzimtzum.
$TT concept remove tzimtzum --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "dangling_crossref"' err.json
# File should be untouched.
$TT lookup tzimtzum --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `65`
- Error `dangling_crossref` because sefirah references tzimtzum.
- File untouched.

## TC-RM-XREF-002 — remove with --force overrides refusal P0

```sh
W=$(work)
$TT concept remove tzimtzum --tbx "${W}" --force >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# tzimtzum should be gone.
$TT lookup tzimtzum --tbx "${W}" >lookup.json 2>/dev/null
LOOKUP_EXIT=$?
test "${LOOKUP_EXIT}" -eq 1 && echo "removed: ok"
```

- **exit**: `0`
- Concept removed despite inbound crossref.

## TC-RM-XREF-003 — validate after --force shows unresolved_crossref P1

```sh
W=$(work)
$TT concept remove tzimtzum --tbx "${W}" --force >out.json 2>err.json
# Now validate — sefirah still references tzimtzum which is gone.
$TT validate --tbx "${W}" >val.json 2>val-err.json
echo "validate exit=$?"
jq -e '[.warnings[] | select(.code == "unresolved_crossref")] | length > 0' val.json
```

- `validate` surfaces `unresolved_crossref` warning for sefirah's dangling reference.

---

# 10. term add

## TC-TERM-ADD-001 — add term to existing concept and langSec P0

```sh
W=$(work)
$TT term add tzimtzum --tbx "${W}" \
  --lang en --term "divine withdrawal" --status admittedTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# Verify term was added.
$TT lookup tzimtzum --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `0`
- Term `divine withdrawal` added to tzimtzum's English langSec.

## TC-TERM-ADD-002 — add term creates new langSec if needed P0

```sh
W=$(work)
$TT term add tzimtzum --tbx "${W}" \
  --lang fr --term "tsimtsoum" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# French langSec should now exist for tzimtzum.
$TT lookup tzimtzum --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `0`
- French langSec created automatically.

## TC-TERM-ADD-003 — add term to nonexistent concept fails P0

```sh
W=$(work)
$TT term add nonexistent-id --tbx "${W}" \
  --lang en --term "test" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "not_found"' err.json
```

- **exit**: `65`
- Error `not_found`.

## TC-TERM-ADD-004 — invalid picklist value rejected P1

```sh
W=$(work)
$TT term add tzimtzum --tbx "${W}" \
  --lang en --term "test" --status "badstatus" >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `2`
- Invalid picklist value rejected at the urfave validator layer.

---

# 11. term deprecate

## TC-TERM-DEP-001 — deprecate existing term P0

```sh
W=$(work)
$TT term deprecate sefirah --tbx "${W}" \
  --lang en --term "sephirah" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# Verify the term's status changed.
$TT lookup sefirah --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `0`
- `sephirah` status changed to `deprecatedTerm-admn-sts`.

## TC-TERM-DEP-002 — deprecate nonexistent term fails P0

```sh
W=$(work)
$TT term deprecate sefirah --tbx "${W}" \
  --lang en --term "nonexistent" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "not_found"' err.json
```

- **exit**: `65`
- Error `not_found` for term not in the given langSec.

## TC-TERM-DEP-003 — deprecate term in nonexistent langSec fails P1

```sh
W=$(work)
# tzimtzum exists but has no French langSec.
$TT term deprecate tzimtzum --tbx "${W}" \
  --lang fr --term "tsimtsoum" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "not_found"' err.json
```

- **exit**: `65`
- Error `not_found` for missing language section.

## TC-TERM-DEP-004 — deprecate term in nonexistent concept fails P1

```sh
W=$(work)
$TT term deprecate no-such-concept --tbx "${W}" \
  --lang en --term "test" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "not_found"' err.json
```

- **exit**: `65`
- Error `not_found` for missing concept.

---

# 12. Dry-run

## TC-DRY-001 — dry-run shows final-state preview P0

```sh
W=$(work)
$TT concept add --tbx "${W}" --dry-run \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.result.concept_id == "tikkun"' out.json
```

- **exit**: `0`
- Output shows the concept as it would appear after the write.

## TC-DRY-002 — dry-run does not modify file P0

```sh
W=$(work)
MD5_BEFORE=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
$TT concept add --tbx "${W}" --dry-run \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
MD5_AFTER=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
echo "before=${MD5_BEFORE} after=${MD5_AFTER}"
test "${MD5_BEFORE}" = "${MD5_AFTER}" && echo "file unchanged: ok"
# Concept should NOT be findable.
$TT lookup "tikkun" --tbx "${W}" >lookup.json 2>/dev/null
LOOKUP_EXIT=$?
test "${LOOKUP_EXIT}" -eq 1 && echo "not persisted: ok"
```

- File checksum identical before and after dry-run.
- `lookup` confirms concept was NOT persisted.

## TC-DRY-003 — dry-run on concept update does not modify file P1

```sh
W=$(work)
MD5_BEFORE=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
$TT concept update tzimtzum --tbx "${W}" --merge --dry-run \
  --lang fr --term "tsimtsoum" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
MD5_AFTER=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
test "${MD5_BEFORE}" = "${MD5_AFTER}" && echo "file unchanged: ok"
```

- **exit**: `0`
- File checksum identical — update was preview only.

## TC-DRY-004 — dry-run on concept remove does not modify file P2

```sh
W=$(work)
MD5_BEFORE=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
$TT concept remove razon-historica --tbx "${W}" --dry-run >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
MD5_AFTER=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
test "${MD5_BEFORE}" = "${MD5_AFTER}" && echo "file unchanged: ok"
# Concept should still be findable.
$TT lookup razon-historica --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `0`
- File checksum identical — concept NOT removed.
- `lookup` confirms concept still present.

## TC-DRY-005 — dry-run runs pre-write validation P1

```sh
W=$(work)
# Try to add a concept with an ID that already exists — dry-run should catch it.
$TT concept add --tbx "${W}" --dry-run \
  --id "tzimtzum" --lang en --term "duplicate" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "duplicate_id"' err.json
```

- **exit**: `65`
- Dry-run catches `duplicate_id` without modifying the file.

---

# 13. Transaction records

## TC-TXN-001 — --transaction adds transacGrp to concept P0

Transaction records are written to the TBX file but not included in
the JSON response. Verify via grep in the written file.

```sh
W=$(work)
$TT concept add --tbx "${W}" --transaction --author "Andre Silva" \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# Transaction record in the TBX file.
grep -q "modification" "${W}" && echo "transactionType: ok"
grep -q "Andre Silva" "${W}" && echo "responsibility: ok"
grep -q "<date>" "${W}" && echo "date: ok"
```

- **exit**: `0`
- TBX file contains `<transacGrp>` with `modification` type, `Andre Silva` responsibility, and `<date>`.

## TC-TXN-002 — --transaction without --author omits responsibility, warns P0

```sh
W=$(work)
$TT concept add --tbx "${W}" --transaction \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# responsibility should be absent from the TBX file's transacGrp.
grep -q "responsibility" "${W}" && echo "responsibility found: FAIL" || echo "no responsibility: ok"
# stderr should contain a WARN about missing author.
grep -qi "warn" err.json && echo "author warning: ok"
```

- **exit**: `0`
- TBX file has `<transacGrp>` without `<basic:responsibility>`.
- WARN-level log on stderr about missing author.

## TC-TXN-003 — term add places transaction at termSec scope P1

```sh
W=$(work)
$TT term add tzimtzum --tbx "${W}" --transaction --author "Andre Silva" \
  --lang en --term "divine withdrawal" --status admittedTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# Transaction should be inside termSec (after the term), not at concept level.
# Verify transacGrp appears after the new term in the file.
grep -n "divine withdrawal" "${W}"
grep -n "transacGrp" "${W}"
```

- **exit**: `0`
- `<transacGrp>` appears inside the `<termSec>` containing `divine withdrawal`, not at concept level.

## TC-TXN-004 — TERMINOLOGY_AUTHOR env var resolves author P1

```sh
W=$(work)
TERMINOLOGY_AUTHOR="Env Author" $TT concept add --tbx "${W}" --transaction \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
grep -q "Env Author" "${W}" && echo "env author: ok"
```

- **exit**: `0`
- Author resolved from `TERMINOLOGY_AUTHOR` env var.
- `--author` flag takes precedence if both are set (tested implicitly by TC-TXN-001).

## TC-TXN-005 — --author without --transaction is ignored P1

```sh
W=$(work)
$TT concept add --tbx "${W}" --author "Andre Silva" \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# No transaction record should be present in the TBX file.
grep -q "transac" "${W}" && echo "transacGrp found: FAIL" || echo "no transacGrp: ok"
```

- **exit**: `0`
- No `<transacGrp>` in the TBX file when `--transaction` is not set.
- `--author` silently ignored.

---

# 14. ID stability

## TC-STAB-001 — concept update with new preferred term keeps ID P0

```sh
W=$(work)
# Rename tzimtzum's preferred English term.
echo '{"languages":{"en":{"preferred":{"term":"divine contraction","administrative_status":"preferredTerm-admn-sts"}}}}' | \
  $TT concept update tzimtzum --tbx "${W}" --merge >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.result.concept_id == "tzimtzum"' out.json
```

- **exit**: `0`
- ID remains `tzimtzum` despite preferred term changing to `divine contraction`.

## TC-STAB-002 — term add with preferred status does not change ID P0

```sh
W=$(work)
$TT term add tzimtzum --tbx "${W}" \
  --lang en --term "divine contraction" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.result.concept_id == "tzimtzum"' out.json
```

- **exit**: `0`
- ID remains `tzimtzum` after adding a new preferred term via `term add`.

---

# 15. Pre-write validation

## TC-PREVAL-001 — validation catches invalid state before write P0

```sh
W=$(work)
# Try to add a concept with duplicate ID.
$TT concept add --tbx "${W}" \
  --id "tzimtzum" --lang en --term "dup" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "duplicate_id"' err.json
# File should be untouched.
$TT validate --tbx "${W}" >val.json 2>/dev/null
echo "validate exit=$?"
jq -e '.ok == true' val.json
```

- **exit**: `65`
- Error `duplicate_id`.
- File untouched — validate still passes.

## TC-PREVAL-002 — dangling crossref in payload caught before write P1

```sh
W=$(work)
# Add a concept that references a nonexistent concept.
echo '{"concept_id":"bad-refs","subject_field":"test","cross_refs":[{"target":"nonexistent-concept","label":"test"}],"languages":{"en":{"preferred":{"term":"bad refs","administrative_status":"preferredTerm-admn-sts"}}}}' | \
  $TT concept add --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
```

- **exit**: `65`
- Unresolved crossref caught by pre-write validation (`validation_error`).

---

# 16. Concurrency — file locking

## TC-LOCK-001 — locked file fails fast with tbx_locked P1

The lock mechanism uses `fcntl`/`flock` advisory locking — file
existence alone does not simulate a held lock. On Linux, use `flock(1)`:

```sh
W=$(work)
# Hold an exclusive flock in the background.
flock -x "${W}.lock" sleep 5 &
LOCK_PID=$!
sleep 0.5
$TT concept add --tbx "${W}" \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "tbx_locked"' err.json
kill "${LOCK_PID}" 2>/dev/null
wait "${LOCK_PID}" 2>/dev/null
```

- **exit**: `3` if lock is held.
- Error code `tbx_locked`.
- **macOS note**: `flock(1)` is not available. SKIP on macOS; test on
  Linux or write a small Go helper that holds the lock.

---

# 17. Envelope shape

## TC-ENV-001 — success envelope has correct top-level keys P0

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e 'has("schema_version", "ok", "result")' out.json
jq -e '.schema_version == 1' out.json
jq -e '.ok == true' out.json
jq -e '.result | type == "object"' out.json
```

- Top-level keys: `schema_version`, `ok`, `result`.
- `result` is an object containing the concept.

## TC-ENV-002 — concept shape in output matches lookup P0

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --id "tikkun" --subject-field "kabbalah" \
  --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
# Concept should have concept_id, subject_field, languages.
jq -e '.result | has("concept_id")' out.json
jq -e '.result.concept_id == "tikkun"' out.json
jq -e '.result | has("subject_field")' out.json
jq -e '.result | has("languages")' out.json
```

- Output concept has at least `concept_id`, `subject_field`, `languages`.
- Matches the lookup output shape.

## TC-ENV-003 — error envelope on stderr P0

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --id "tzimtzum" --lang en --term "dup" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == false' err.json
jq -e '.error | has("code", "message")' err.json
jq -e '.error.code == "duplicate_id"' err.json
```

- Error envelope on stderr with `ok: false`, `error.code`, `error.message`.

---

# 18. Error cases

## TC-ERR-001 — duplicate_id on concept add P0

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --id "tzimtzum" --lang en --term "dup" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "duplicate_id"' err.json
```

- **exit**: `65`
- Error code `duplicate_id`.

## TC-ERR-002 — not_found on concept update P0

```sh
W=$(work)
$TT concept update no-such-id --tbx "${W}" --merge \
  --lang en --term "test" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "not_found"' err.json
```

- **exit**: `65`
- Error code `not_found`.

## TC-ERR-003 — not_found on concept remove P0

```sh
W=$(work)
$TT concept remove no-such-id --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "not_found"' err.json
```

- **exit**: `65`

## TC-ERR-004 — invalid_id for empty derivation P1

```sh
W=$(work)
# Hebrew-only term without --id produces empty slug.
$TT concept add --tbx "${W}" \
  --lang he --term "אור" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "invalid_id"' err.json
```

- **exit**: `65`
- Empty ID derivation rejected with `invalid_id` and hint to use `--id`.

## TC-ERR-005 — invalid_input on malformed JSON stdin P0

```sh
W=$(work)
echo '{not valid json' | \
  $TT concept add --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "invalid_input"' err.json
```

- **exit**: `65`
- Malformed JSON rejected with `invalid_input`.

## TC-ERR-006 — no_tbx_path when --tbx omitted P0

```sh
$TT concept add --id "test" --lang en --term "test" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "no_tbx_path"' err.json
```

- **exit**: `2`
- Error code `no_tbx_path`.

## TC-ERR-007 — TERMINOLOGY_TBX env var resolves path P1

```sh
W=$(work)
TERMINOLOGY_TBX="${W}" $TT concept add \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- TBX path resolved from `TERMINOLOGY_TBX` env var.

---

# 19. Stream routing

## TC-STREAM-001 — success: envelope on stdout, stderr empty P0

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
test -s out.json && echo "stdout: non-empty"
test ! -s err.json && echo "stderr: empty"
```

- stdout has the JSON envelope.
- stderr is empty on success.

## TC-STREAM-002 — error: envelope on stderr, stdout empty P0

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --id "tzimtzum" --lang en --term "dup" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
test ! -s out.json && echo "stdout: empty"
test -s err.json && echo "stderr: non-empty"
```

- stderr has the error envelope.
- stdout is empty on error.

## TC-STREAM-003 — dry-run: preview on stdout, stderr empty P1

```sh
W=$(work)
$TT concept add --tbx "${W}" --dry-run \
  --id "tikkun" --lang en --term "tikkun" --status preferredTerm-admn-sts >out.json 2>err.json
echo "exit=$?"
test -s out.json && echo "stdout: non-empty"
test ! -s err.json && echo "stderr: empty"
```

- Dry-run preview on stdout, stderr empty.

---

# 20. Regression — previous commands still work

## TC-REG-001 — validate unaffected P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.schema_version == 1' out.json
```

- **exit**: `0`
- Validate command still works after E7 changes.

## TC-REG-002 — lookup unaffected P0

```sh
TERM="tzimtzum"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM}" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.results | length > 0' out.json
```

- **exit**: `0`
- Lookup command still works.

## TC-REG-003 — scan unaffected P0

```sh
cat > "${QA_TMP}/reg-scan.md" <<'EOF'
The concept of tzimtzum is central.
EOF
$TT scan "${QA_TMP}/reg-scan.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.matches | length > 0' out.json
```

- **exit**: `0`
- Scan command still works.

## TC-REG-004 — extract unaffected P1

```sh
cat > "${QA_TMP}/reg-extract.md" <<'EOF'
---
lang: es
---
El concepto de tzimtzum es central. Cada sefirah es un atributo.
EOF
$TT extract "${QA_TMP}/reg-extract.md" --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- Extract command still works.

## TC-REG-005 — check unaffected P0

```sh
cat > "${QA_TMP}/reg-source.md" <<'EOF'
---
lang: es
---
El concepto de tzimtzum es central.
EOF
cat > "${QA_TMP}/reg-target.md" <<'EOF'
---
lang: he
---
צמצום הוא מושג מרכזי.
EOF
$TT check "${QA_TMP}/reg-source.md" "${QA_TMP}/reg-target.md" \
  --tbx "${QA_TMP}/glossary.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- Check command still works.

## TC-REG-006 — schema includes write commands P1

```sh
$TT schema >out.json 2>err.json
echo "exit=$?"
# Write commands are subcommands of "concept" and "term" parent commands.
CONCEPT_SUBS=$(jq -r '.commands[] | select(.name == "concept") | [.commands[]?.name] | sort | join(",")' out.json)
TERM_SUBS=$(jq -r '.commands[] | select(.name == "term") | [.commands[]?.name] | sort | join(",")' out.json)
echo "concept subcommands: ${CONCEPT_SUBS}"
echo "term subcommands: ${TERM_SUBS}"
echo "${CONCEPT_SUBS}" | grep -q "add" && echo "concept add: ok"
echo "${CONCEPT_SUBS}" | grep -q "update" && echo "concept update: ok"
echo "${CONCEPT_SUBS}" | grep -q "remove" && echo "concept remove: ok"
echo "${TERM_SUBS}" | grep -q "add" && echo "term add: ok"
echo "${TERM_SUBS}" | grep -q "deprecate" && echo "term deprecate: ok"
```

- Schema has `concept` parent with `add`, `update`, `remove` subcommands.
- Schema has `term` parent with `add`, `deprecate` subcommands.

---

# Cleanup

```sh
rm -rf "${QA_TMP}"
rm -f out.json err.json lookup.json val.json val-err.json
```

---

# Test case summary

| Section                               | Cases  | Priority |
| ------------------------------------- | ------ | -------- |
| concept add — flag input              | 3      | P0–P1   |
| concept add — ID derivation           | 3      | P0–P1   |
| concept add — JSON stdin              | 2      | P0–P1   |
| concept add — TBX fragment stdin      | 3      | P0–P1   |
| concept update — merge                | 5      | P0–P2   |
| concept update — replace              | 2      | P0       |
| concept update — merge/replace mutex  | 2      | P0       |
| concept remove — basic                | 2      | P0       |
| concept remove — dangling crossref    | 3      | P0–P1   |
| term add                              | 4      | P0–P1   |
| term deprecate                        | 4      | P0–P1   |
| Dry-run                               | 5      | P0–P2   |
| Transaction records                   | 5      | P0–P1   |
| ID stability                          | 2      | P0       |
| Pre-write validation                  | 2      | P0–P1   |
| Concurrency — file locking            | 1      | P1       |
| Envelope shape                        | 3      | P0       |
| Error cases                           | 7      | P0–P1   |
| Stream routing                        | 3      | P0–P1   |
| Regression — previous commands        | 6      | P0–P1   |
| **Total**                             | **67** |          |
