# Test Plan: E8 — `terminology apply` manual QA

> **Status**: ready to execute once E8 is completed
> Run end-to-end in a single sitting.

## Purpose

E8's deliverable is the **apply command**: a bulk declarative write that
reconciles a desired-state payload against the current glossary. It computes
the minimal set of add/update/remove operations and applies them
all-or-nothing. Idempotent by construction.

This is **apply command QA**. E7 QA covered the granular write commands
(`concept add/update/remove`, `term add/deprecate`), shared write
infrastructure (dry-run, transaction records, ID derivation), and error
sentinels. E8 tests the bulk reconciliation surface: full-state patch model,
concept equality (transacGrp stripping), wholesale replace on update, `--prune`
semantics, dangling crossref refusal, payload formats (JSON + TBX fragment),
format auto-detection, idempotency, concurrency (lock scope), the
`apply_validation_failed` error sentinel with `failures[]` details, and the
apply-specific output envelope.

## Scope

### In scope

- `apply --file PAYLOAD` — reconcile payload against glossary.
- Patch model: add (new in payload), update (changed), unchanged (identical), remove (absent + `--prune`).
- Concept equality: canonicalized XML byte-identical after stripping `<transacGrp>`.
- Update rule: wholesale replace — payload IS authoritative; omitted fields dropped.
- `--prune` removes glossary concepts absent from payload.
- `--prune` + dangling crossref → refused with `dangling_crossref`, file untouched.
- Payload formats: JSON (`{"concepts": [...]}`) and TBX fragment (`<conceptEntryList>`).
- Format auto-detection: extension-based (`.json` → JSON, `.tbx`/`.xml` → TBX) + content sniffing for stdin.
- `--file -` reads from stdin.
- `--dry-run` preview without writing.
- `--transaction` / `--author` transaction records on added/updated concepts only.
- Idempotency: same payload twice → all unchanged on second run.
- Concurrency: advisory lock held across read→modify→validate→write.
- Error sentinels: `apply_validation_failed` (exit 1, with `failures[]`), `invalid_input` (exit 65), `dangling_crossref` (exit 65), `no_tbx_path` (exit 2).
- Output envelope: `schema_version`, `ok`, `applied` (added/updated/removed/unchanged), `warnings`.
- All ID lists sorted ASCII byte order.
- Exit codes: 0 (success), 1 (validation failed), 2 (usage error), 3 (I/O), 65 (data error).

### Out of scope

- Granular write commands (`concept add/update/remove`, `term add/deprecate`) — covered by E7 QA.
- ID derivation and stability — covered by E7 QA.
- Validation internals — covered by E3 QA.
- Matcher pipeline — covered by E5 QA.
- Check command — covered by E6 QA.
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

Create a mutable TBX glossary and payload fixtures for apply command tests.
Each test that mutates copies the base glossary first.

````sh
export QA_TMP=$(mktemp -d /tmp/e8-qa-XXXXXX)

# --- Base glossary fixture ---
# Four concepts: tzimtzum (en, es, he), sefirah (en, he with crossref → tzimtzum),
# razon-historica (en, es), malkhut (en).
# sefirah cross-references tzimtzum.
cat > "${QA_TMP}/glossary.tbx" <<'TBXEOF'
<?xml version="1.0" encoding="UTF-8"?>
<?xml-model href="https://raw.githubusercontent.com/LTAC-Global/TBX-Linguist_Module/master/Schema/TBXcheckerTBX-Linguist.sch" type="application/xml" schematypens="http://purl.oclc.org/dml/schematron"?>
<tbx style="dct" type="TBX-Linguist" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader>
    <fileDesc><sourceDesc><p>E8 QA fixture</p></sourceDesc></fileDesc>
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
    <conceptEntry id="malkhut">
      <min:subjectField>kabbalah</min:subjectField>
      <langSec xml:lang="en">
        <termSec>
          <term>malkhut</term>
          <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
        </termSec>
      </langSec>
    </conceptEntry>
  </body></text>
</tbx>
TBXEOF

# --- JSON payload: all four concepts unchanged (idempotency test) ---
# Matches the glossary exactly — apply should produce all unchanged.
cat > "${QA_TMP}/payload-unchanged.json" <<'JSONEOF'
{
  "concepts": [
    {
      "concept_id": "tzimtzum",
      "subject_field": "kabbalah",
      "languages": {
        "en": { "preferred": { "term": "tzimtzum", "administrative_status": "preferredTerm-admn-sts" } },
        "es": { "preferred": { "term": "tzimtzum", "administrative_status": "preferredTerm-admn-sts" } },
        "he": { "preferred": { "term": "צמצום", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "sefirah",
      "subject_field": "kabbalah",
      "cross_refs": [{"target": "tzimtzum", "label": "related kabbalistic concept"}],
      "languages": {
        "en": { "preferred": { "term": "sefirah", "administrative_status": "preferredTerm-admn-sts" } },
        "he": { "preferred": { "term": "ספירה", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "razon-historica",
      "subject_field": "philosophy",
      "languages": {
        "en": { "preferred": { "term": "historical reason", "administrative_status": "preferredTerm-admn-sts" } },
        "es": { "preferred": { "term": "razón histórica", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "malkhut",
      "subject_field": "kabbalah",
      "languages": {
        "en": { "preferred": { "term": "malkhut", "administrative_status": "preferredTerm-admn-sts" } }
      }
    }
  ]
}
JSONEOF

# --- JSON payload: add one new concept ---
cat > "${QA_TMP}/payload-add.json" <<'JSONEOF'
{
  "concepts": [
    {
      "concept_id": "tzimtzum",
      "subject_field": "kabbalah",
      "languages": {
        "en": { "preferred": { "term": "tzimtzum", "administrative_status": "preferredTerm-admn-sts" } },
        "es": { "preferred": { "term": "tzimtzum", "administrative_status": "preferredTerm-admn-sts" } },
        "he": { "preferred": { "term": "צמצום", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "sefirah",
      "subject_field": "kabbalah",
      "cross_refs": [{"target": "tzimtzum", "label": "related kabbalistic concept"}],
      "languages": {
        "en": { "preferred": { "term": "sefirah", "administrative_status": "preferredTerm-admn-sts" } },
        "he": { "preferred": { "term": "ספירה", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "razon-historica",
      "subject_field": "philosophy",
      "languages": {
        "en": { "preferred": { "term": "historical reason", "administrative_status": "preferredTerm-admn-sts" } },
        "es": { "preferred": { "term": "razón histórica", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "malkhut",
      "subject_field": "kabbalah",
      "languages": {
        "en": { "preferred": { "term": "malkhut", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "binah",
      "subject_field": "kabbalah",
      "languages": {
        "en": { "preferred": { "term": "binah", "administrative_status": "preferredTerm-admn-sts" } },
        "he": { "preferred": { "term": "בינה", "administrative_status": "preferredTerm-admn-sts" } }
      }
    }
  ]
}
JSONEOF

# --- JSON payload: update one concept (change subject_field + drop a langSec) ---
cat > "${QA_TMP}/payload-update.json" <<'JSONEOF'
{
  "concepts": [
    {
      "concept_id": "tzimtzum",
      "subject_field": "mysticism",
      "languages": {
        "en": { "preferred": { "term": "tzimtzum", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "sefirah",
      "subject_field": "kabbalah",
      "cross_refs": [{"target": "tzimtzum", "label": "related kabbalistic concept"}],
      "languages": {
        "en": { "preferred": { "term": "sefirah", "administrative_status": "preferredTerm-admn-sts" } },
        "he": { "preferred": { "term": "ספירה", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "razon-historica",
      "subject_field": "philosophy",
      "languages": {
        "en": { "preferred": { "term": "historical reason", "administrative_status": "preferredTerm-admn-sts" } },
        "es": { "preferred": { "term": "razón histórica", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "malkhut",
      "subject_field": "kabbalah",
      "languages": {
        "en": { "preferred": { "term": "malkhut", "administrative_status": "preferredTerm-admn-sts" } }
      }
    }
  ]
}
JSONEOF

# --- JSON payload: mixed add + update + unchanged ---
cat > "${QA_TMP}/payload-mixed.json" <<'JSONEOF'
{
  "concepts": [
    {
      "concept_id": "tzimtzum",
      "subject_field": "mysticism",
      "languages": {
        "en": { "preferred": { "term": "tzimtzum", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "sefirah",
      "subject_field": "kabbalah",
      "cross_refs": [{"target": "tzimtzum", "label": "related kabbalistic concept"}],
      "languages": {
        "en": { "preferred": { "term": "sefirah", "administrative_status": "preferredTerm-admn-sts" } },
        "he": { "preferred": { "term": "ספירה", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "razon-historica",
      "subject_field": "philosophy",
      "languages": {
        "en": { "preferred": { "term": "historical reason", "administrative_status": "preferredTerm-admn-sts" } },
        "es": { "preferred": { "term": "razón histórica", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "malkhut",
      "subject_field": "kabbalah",
      "languages": {
        "en": { "preferred": { "term": "malkhut", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "tiferet",
      "subject_field": "kabbalah",
      "languages": {
        "en": { "preferred": { "term": "tiferet", "administrative_status": "preferredTerm-admn-sts" } }
      }
    }
  ]
}
JSONEOF

# --- JSON payload: prune scenario (only two concepts; absent ones get removed) ---
cat > "${QA_TMP}/payload-prune.json" <<'JSONEOF'
{
  "concepts": [
    {
      "concept_id": "tzimtzum",
      "subject_field": "kabbalah",
      "languages": {
        "en": { "preferred": { "term": "tzimtzum", "administrative_status": "preferredTerm-admn-sts" } },
        "es": { "preferred": { "term": "tzimtzum", "administrative_status": "preferredTerm-admn-sts" } },
        "he": { "preferred": { "term": "צמצום", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "sefirah",
      "subject_field": "kabbalah",
      "cross_refs": [{"target": "tzimtzum", "label": "related kabbalistic concept"}],
      "languages": {
        "en": { "preferred": { "term": "sefirah", "administrative_status": "preferredTerm-admn-sts" } },
        "he": { "preferred": { "term": "ספירה", "administrative_status": "preferredTerm-admn-sts" } }
      }
    }
  ]
}
JSONEOF

# --- JSON payload: prune + dangling crossref ---
# Only tzimtzum present; sefirah (which refs tzimtzum) absent → prune would
# remove sefirah, but sefirah refs tzimtzum which IS present. That's fine.
# To trigger dangling: remove tzimtzum but keep sefirah (which refs it).
cat > "${QA_TMP}/payload-prune-dangle.json" <<'JSONEOF'
{
  "concepts": [
    {
      "concept_id": "sefirah",
      "subject_field": "kabbalah",
      "cross_refs": [{"target": "tzimtzum", "label": "related kabbalistic concept"}],
      "languages": {
        "en": { "preferred": { "term": "sefirah", "administrative_status": "preferredTerm-admn-sts" } },
        "he": { "preferred": { "term": "ספירה", "administrative_status": "preferredTerm-admn-sts" } }
      }
    },
    {
      "concept_id": "malkhut",
      "subject_field": "kabbalah",
      "languages": {
        "en": { "preferred": { "term": "malkhut", "administrative_status": "preferredTerm-admn-sts" } }
      }
    }
  ]
}
JSONEOF

# --- JSON payload: validation failure (crossref to nonexistent concept) ---
cat > "${QA_TMP}/payload-invalid.json" <<'JSONEOF'
{
  "concepts": [
    {
      "concept_id": "bad-concept",
      "subject_field": "kabbalah",
      "cross_refs": [{"target": "nonexistent", "label": "broken ref"}],
      "languages": {
        "en": { "preferred": { "term": "bad", "administrative_status": "preferredTerm-admn-sts" } }
      }
    }
  ]
}
JSONEOF

# --- Malformed JSON ---
cat > "${QA_TMP}/payload-malformed.json" <<'JSONEOF'
{not valid json at all
JSONEOF

# --- JSON without concepts key ---
cat > "${QA_TMP}/payload-no-concepts.json" <<'JSONEOF'
{"items": []}
JSONEOF

# --- TBX fragment payload ---
cat > "${QA_TMP}/payload-tbx.tbx" <<'XMLEOF'
<conceptEntryList>
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
  <conceptEntry id="malkhut">
    <min:subjectField>kabbalah</min:subjectField>
    <langSec xml:lang="en">
      <termSec>
        <term>malkhut</term>
        <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
      </termSec>
    </langSec>
  </conceptEntry>
  <conceptEntry id="tiferet">
    <min:subjectField>kabbalah</min:subjectField>
    <langSec xml:lang="en">
      <termSec>
        <term>tiferet</term>
        <min:administrativeStatus>preferredTerm-admn-sts</min:administrativeStatus>
      </termSec>
    </langSec>
  </conceptEntry>
</conceptEntryList>
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

- All E8 tickets closed.
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

| Risk                                                        | Mitigation                                                   |
| ----------------------------------------------------------- | ------------------------------------------------------------ |
| Concept equality misses transacGrp at termSec level         | TC-EQ-001/002/003 test equality across transaction scenarios |
| Wholesale replace drops crossrefs or corrupts ID            | TC-UPD-001/002 verify replace semantics and ID preservation  |
| Prune removes concept with inbound crossref                 | TC-PRUNE-002 verifies dangling_crossref refusal              |
| Idempotency broken by transaction records                   | TC-IDEM-001/002 verify transac-strip makes second run clean  |
| Format detection fails for stdin without extension           | TC-FMT-003 tests content sniffing on stdin                   |
| Lock not held across full read→write window                 | TC-LOCK-001 verifies lock error under contention             |
| apply_validation_failed missing failures[] detail            | TC-VALFAIL-001 verifies error envelope shape                 |
| Output lists not sorted ASCII byte order                    | TC-SORT-001 checks sorted order                              |
| Dry-run silently writes file                                | TC-DRY-001/002 verify file checksum unchanged                |
| Payload-absent fields retained instead of dropped on update | TC-UPD-001 checks only payload-present fields survive        |
| All-or-nothing violated (partial write)                     | TC-ATOMIC-001 confirms no change on validation failure       |

## Conventions

Same as E2/E3/E4/E5/E6/E7 QA: see [E2-manual-qa.md](E2-manual-qa.md) §Conventions for
envelope shapes, exit code map, and test case format.

### Exit code map (E8 surface)

| Code | Meaning              | Source                                                       |
| ---- | -------------------- | ------------------------------------------------------------ |
| 0    | success              | Apply completed, reconciliation applied                      |
| 1    | validation failed    | `apply_validation_failed` — recoverable, with `failures[]`   |
| 2    | usage error          | `no_tbx_path`, missing `--file`                              |
| 3    | I/O error            | File not found, `tbx_locked`                                 |
| 65   | data error           | `invalid_input`, `dangling_crossref`                         |

### Mutation test pattern

Every test that mutates the glossary **must** copy the base fixture first.
The `work` shell function creates a timestamped copy and prints its path:

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-unchanged.json"
```

---

# 1. Basic apply — add

## TC-ADD-001 — apply adds new concept P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.added | length == 1' out.json
jq -e '.applied.added[0] == "binah"' out.json
jq -e '.applied.unchanged | length == 4' out.json
```

- **exit**: `0`
- `applied.added` contains `["binah"]`.
- `applied.unchanged` contains the other 4 concepts.
- `applied.updated` and `applied.removed` are empty.

## TC-ADD-002 — added concept persisted to file P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" >out.json 2>err.json
$TT lookup "binah" --tbx "${W}" >lookup.json 2>/dev/null
echo "exit=$?"
jq -e '.ok == true' lookup.json
jq -e '.results[0].concept_id == "binah"' lookup.json
```

- **exit**: `0`
- `lookup` confirms `binah` was persisted.

---

# 2. Basic apply — update

## TC-UPD-001 — wholesale replace on update P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-update.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.updated | length == 1' out.json
jq -e '.applied.updated[0] == "tzimtzum"' out.json
# Verify wholesale replace: only en remains, es and he dropped.
$TT lookup "tzimtzum" --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
jq -e '.results[0].subject_field == "mysticism"' lookup.json
```

- **exit**: `0`
- `applied.updated` contains `["tzimtzum"]`.
- After replace, `subject_field` is `"mysticism"`.
- Only English langSec survives; es and he dropped (payload omitted them).

## TC-UPD-002 — update preserves concept ID P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-update.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.applied.updated[0] == "tzimtzum"' out.json
$TT lookup "tzimtzum" --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.results[0].concept_id == "tzimtzum"' lookup.json
```

- **exit**: `0`
- Concept ID `tzimtzum` preserved after wholesale replace.

---

# 3. Basic apply — unchanged

## TC-UNCH-001 — identical concept classified as unchanged P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-unchanged.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.unchanged | length == 4' out.json
jq -e '.applied.added | length == 0' out.json
jq -e '.applied.updated | length == 0' out.json
jq -e '.applied.removed | length == 0' out.json
```

- **exit**: `0`
- All 4 concepts classified as unchanged.
- No adds, updates, or removes.

---

# 4. Mixed apply — add + update + unchanged

## TC-MIX-001 — mixed reconciliation P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-mixed.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.added == ["tiferet"]' out.json
jq -e '.applied.updated == ["tzimtzum"]' out.json
jq -e '.applied.unchanged | length == 3' out.json
jq -e '.applied.removed | length == 0' out.json
```

- **exit**: `0`
- `tiferet` added, `tzimtzum` updated (subject_field changed + langSecs dropped), 3 unchanged.
- No removes (no `--prune`).

---

# 5. Idempotency

## TC-IDEM-001 — second run yields all unchanged P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" >out1.json 2>err1.json
echo "first exit=$?"
jq -e '.applied.added == ["binah"]' out1.json
# Second run with same payload.
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" >out2.json 2>err2.json
echo "second exit=$?"
jq -e '.ok == true' out2.json
jq -e '.applied.unchanged | length == 5' out2.json
jq -e '.applied.added | length == 0' out2.json
jq -e '.applied.updated | length == 0' out2.json
```

- **exit**: `0` both runs.
- Second run: all 5 concepts unchanged.

## TC-IDEM-002 — idempotent after --transaction P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" \
  --transaction --author "QA Tester" >out1.json 2>err1.json
echo "first exit=$?"
jq -e '.applied.added == ["binah"]' out1.json
# Second run with same payload (no --transaction this time).
# TransacGrp from first run should NOT flip concepts to "updated"
# because equality strips transacGrp.
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" >out2.json 2>err2.json
echo "second exit=$?"
jq -e '.ok == true' out2.json
jq -e '.applied.unchanged | length == 5' out2.json
jq -e '.applied.updated | length == 0' out2.json
```

- **exit**: `0` both runs.
- Transaction records from first run do not cause second run to see "updated".
- Verifies transacGrp-stripping equality.

---

# 6. Concept equality

## TC-EQ-001 — concepts with only transacGrp difference are equal P1

This is implicitly tested by TC-IDEM-002. Verify explicitly by adding
a concept with a transaction, then re-applying the same payload.

```sh
W=$(work)
# Add a concept with transaction records via granular command.
$TT concept add --tbx "${W}" --transaction --author "QA Tester" \
  --id "binah" --lang en --term "binah" --status preferredTerm-admn-sts \
  --subject-field kabbalah < /dev/null >dev_null 2>&1
# Now apply a payload containing binah without transactions.
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# binah should be unchanged (transacGrp stripped for comparison).
jq -e '[.applied.unchanged[] | select(. == "binah")] | length == 1' out.json
```

- **exit**: `0`
- `binah` classified as unchanged despite having transacGrp in the file.

## TC-EQ-002 — concepts differing in non-transacGrp fields are not equal P1

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-update.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.applied.updated == ["tzimtzum"]' out.json
```

- **exit**: `0`
- `tzimtzum` classified as updated (subject_field differs).

## TC-EQ-003 — unchanged despite different payload field ordering P1

Different field ordering in the payload should still produce "unchanged"
because the canonical writer normalizes order.

```sh
W=$(work)
# Apply the unchanged payload — all should be unchanged regardless of
# JSON field ordering (the JSON parser produces the same Go struct).
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-unchanged.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.applied.unchanged | length == 4' out.json
```

- **exit**: `0`
- Canonical comparison handles field ordering.

---

# 7. Prune

## TC-PRUNE-001 — prune removes absent concepts P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-prune.json" --prune >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.removed | sort == ["malkhut", "razon-historica"]' out.json
jq -e '.applied.unchanged | sort == ["sefirah", "tzimtzum"]' out.json
# Verify removed concepts are gone.
$TT lookup "razon-historica" --tbx "${W}" >dev_null 2>&1
LOOKUP_EXIT=$?
test "${LOOKUP_EXIT}" -ne 0 && echo "razon-historica removed: ok"
$TT lookup "malkhut" --tbx "${W}" >dev_null 2>&1
LOOKUP_EXIT=$?
test "${LOOKUP_EXIT}" -ne 0 && echo "malkhut removed: ok"
```

- **exit**: `0`
- `razon-historica` and `malkhut` removed (absent from payload).
- `tzimtzum` and `sefirah` unchanged (present in payload).

## TC-PRUNE-002 — prune refuses on dangling crossref P0

```sh
W=$(work)
# Payload has sefirah (refs tzimtzum) and malkhut, but NOT tzimtzum.
# --prune would remove tzimtzum, but sefirah references it → refuse.
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-prune-dangle.json" --prune >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == false' err.json
jq -e '.error.code == "dangling_crossref"' err.json
# File untouched — all 4 concepts still present.
$TT lookup "tzimtzum" --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `65`
- Error `dangling_crossref` because pruning `tzimtzum` would break sefirah's crossref.
- File untouched.

## TC-PRUNE-003 — without --prune, absent concepts preserved P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-prune.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.removed | length == 0' out.json
# Absent concepts still in file (no prune).
$TT lookup "razon-historica" --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `0`
- Without `--prune`, absent concepts are not removed.

---

# 8. Payload formats

## TC-FMT-001 — JSON payload from file path P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-unchanged.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- `.json` extension → JSON format auto-detected.

## TC-FMT-002 — TBX fragment payload from .tbx file P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-tbx.tbx" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.added == ["tiferet"]' out.json
jq -e '.applied.unchanged | length == 4' out.json
```

- **exit**: `0`
- `.tbx` extension → TBX fragment format auto-detected.
- `tiferet` added, others unchanged.

## TC-FMT-003 — JSON payload from stdin P1

```sh
W=$(work)
cat "${QA_TMP}/payload-unchanged.json" | \
  $TT apply --tbx "${W}" --file - >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.unchanged | length == 4' out.json
```

- **exit**: `0`
- `--file -` reads from stdin; JSON content-sniffed.

## TC-FMT-004 — TBX fragment from stdin P1

```sh
W=$(work)
cat "${QA_TMP}/payload-tbx.tbx" | \
  $TT apply --tbx "${W}" --file - >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.added == ["tiferet"]' out.json
```

- **exit**: `0`
- `--file -` reads from stdin; XML content-sniffed.

## TC-FMT-005 — .xml extension treated as TBX P2

```sh
W=$(work)
cp "${QA_TMP}/payload-tbx.tbx" "${QA_TMP}/payload-tbx.xml"
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-tbx.xml" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- `.xml` extension → TBX fragment format auto-detected.

---

# 9. Dry-run

## TC-DRY-001 — dry-run shows reconciliation result P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" --dry-run >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.added == ["binah"]' out.json
jq -e '.applied.unchanged | length == 4' out.json
```

- **exit**: `0`
- Output shows reconciliation result (same as non-dry-run).

## TC-DRY-002 — dry-run does not modify file P0

```sh
W=$(work)
MD5_BEFORE=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" --dry-run >out.json 2>err.json
MD5_AFTER=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
echo "before=${MD5_BEFORE} after=${MD5_AFTER}"
test "${MD5_BEFORE}" = "${MD5_AFTER}" && echo "file unchanged: ok"
# Verify binah was NOT persisted.
$TT lookup "binah" --tbx "${W}" >lookup.json 2>/dev/null
LOOKUP_EXIT=$?
test "${LOOKUP_EXIT}" -ne 0 && echo "not persisted: ok"
```

- File checksum identical before and after dry-run.
- `lookup` confirms `binah` was NOT persisted.

## TC-DRY-003 — dry-run with --prune does not modify file P1

```sh
W=$(work)
MD5_BEFORE=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-prune.json" --prune --dry-run >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.removed | length == 2' out.json
MD5_AFTER=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
test "${MD5_BEFORE}" = "${MD5_AFTER}" && echo "file unchanged: ok"
# Removed concepts still in file.
$TT lookup "razon-historica" --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `0`
- Dry-run reports removes but file untouched.

---

# 10. Transaction records

## TC-TXN-001 — transaction records on added and updated concepts P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-mixed.json" \
  --transaction --author "QA Tester" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
# tiferet was added, tzimtzum was updated — both should have transacGrp.
grep -q "tiferet" "${W}" && echo "tiferet in file: ok"
# Count transacGrp occurrences associated with modified concepts.
grep -c "transacGrp" "${W}"
grep -q "QA Tester" "${W}" && echo "author in file: ok"
```

- **exit**: `0`
- `<transacGrp>` present for added (`tiferet`) and updated (`tzimtzum`) concepts.
- `QA Tester` appears as responsibility.

## TC-TXN-002 — no transaction on unchanged concepts P1

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-unchanged.json" \
  --transaction --author "QA Tester" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.applied.unchanged | length == 4' out.json
# No transacGrp should be added when all concepts are unchanged.
grep -c "transacGrp" "${W}" || echo "no transacGrp: ok"
```

- **exit**: `0`
- No `<transacGrp>` added when all concepts are unchanged.

## TC-TXN-003 — --transaction without --author warns P1

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" \
  --transaction >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
grep -q "transacGrp" "${W}" && echo "transacGrp present: ok"
grep -q "responsibility" "${W}" && echo "responsibility found: FAIL" || echo "no responsibility: ok"
```

- **exit**: `0`
- `<transacGrp>` present without `<basic:responsibility>`.
- WARN on stderr about missing author.

---

# 11. All-or-nothing atomicity

## TC-ATOMIC-001 — validation failure leaves file untouched P0

```sh
W=$(work)
MD5_BEFORE=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-invalid.json" >out.json 2>err.json
echo "exit=$?"
# Validation should fail (crossref to nonexistent concept).
jq -e '.ok == false' err.json
jq -e '.error.code == "apply_validation_failed"' err.json
# File must be untouched.
MD5_AFTER=$(md5 -q "${W}" 2>/dev/null || md5sum "${W}" | cut -d' ' -f1)
test "${MD5_BEFORE}" = "${MD5_AFTER}" && echo "file unchanged: ok"
# All original concepts still present.
$TT lookup "tzimtzum" --tbx "${W}" >lookup.json 2>/dev/null
jq -e '.ok == true' lookup.json
```

- **exit**: `1`
- Error code `apply_validation_failed`.
- File checksum identical — untouched after validation failure.
- All original concepts still present.

---

# 12. apply_validation_failed envelope

## TC-VALFAIL-001 — error envelope has failures[] detail P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-invalid.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == false' err.json
jq -e '.error.code == "apply_validation_failed"' err.json
jq -e '.error.details.failures | type == "array"' err.json
jq -e '.error.details.failures | length > 0' err.json
jq -e '.error.details.failures[0] | has("concept_id", "code", "message")' err.json
```

- **exit**: `1`
- Error code `apply_validation_failed`.
- `error.details.failures` is a non-empty array.
- Each failure has `concept_id`, `code`, `message`.

---

# 13. Error cases

## TC-ERR-001 — malformed JSON payload P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-malformed.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "invalid_input"' err.json
```

- **exit**: `65`
- Error code `invalid_input`.

## TC-ERR-002 — JSON payload with unknown fields P1

```sh
W=$(work)
echo '{"concepts": [{"concept_id": "x", "unknown_field": true, "languages": {"en": {"preferred": {"term": "x", "administrative_status": "preferredTerm-admn-sts"}}}}]}' > "${QA_TMP}/payload-unknown-field.json"
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-unknown-field.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "invalid_input"' err.json
```

- **exit**: `65`
- Unknown JSON field rejected with `invalid_input`.

## TC-ERR-003 — no --tbx flag P0

```sh
$TT apply --file "${QA_TMP}/payload-unchanged.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "no_tbx_path"' err.json
```

- **exit**: `2`
- Error code `no_tbx_path`.

## TC-ERR-004 — nonexistent payload file P1

```sh
W=$(work)
$TT apply --tbx "${W}" --file "/nonexistent/path.json" >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `3`
- I/O error for nonexistent file.

## TC-ERR-005 — TERMINOLOGY_TBX env var resolves path P1

```sh
W=$(work)
TERMINOLOGY_TBX="${W}" $TT apply --file "${QA_TMP}/payload-unchanged.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- TBX path resolved from `TERMINOLOGY_TBX` env var.

## TC-ERR-006 — missing --file flag P0

```sh
W=$(work)
$TT apply --tbx "${W}" >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `2`
- `--file` is required.

---

# 14. Output envelope shape

## TC-ENV-001 — success envelope has correct top-level keys P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-unchanged.json" >out.json 2>err.json
echo "exit=$?"
jq -e 'has("schema_version", "ok", "applied")' out.json
jq -e '.schema_version == 1' out.json
jq -e '.ok == true' out.json
jq -e '.applied | has("added", "updated", "removed", "unchanged")' out.json
```

- Top-level keys: `schema_version`, `ok`, `applied`.
- `applied` has `added`, `updated`, `removed`, `unchanged`.

## TC-ENV-002 — lists are arrays, never null P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-unchanged.json" >out.json 2>err.json
jq -e '.applied.added | type == "array"' out.json
jq -e '.applied.updated | type == "array"' out.json
jq -e '.applied.removed | type == "array"' out.json
jq -e '.applied.unchanged | type == "array"' out.json
```

- All four lists are arrays, even when empty (not `null`).

## TC-ENV-003 — error envelope on stderr P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-malformed.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == false' err.json
jq -e '.error | has("code", "message")' err.json
```

- Error envelope on stderr with `ok: false`, `error.code`, `error.message`.

---

# 15. Determinism — sorted output

## TC-SORT-001 — ID lists sorted ASCII byte order P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-mixed.json" >out.json 2>err.json
echo "exit=$?"
# Unchanged should contain: malkhut, razon-historica, sefirah (sorted).
UNCHANGED=$(jq -r '.applied.unchanged | join(",")' out.json)
SORTED=$(jq -r '.applied.unchanged | sort | join(",")' out.json)
echo "unchanged=${UNCHANGED} sorted=${SORTED}"
test "${UNCHANGED}" = "${SORTED}" && echo "unchanged sorted: ok"
```

- All lists (`added`, `updated`, `removed`, `unchanged`) sorted ASCII byte order.

---

# 16. Concurrency — file locking

## TC-LOCK-001 — locked file fails fast with tbx_locked P1

```sh
W=$(work)
# Hold an exclusive flock in the background.
flock -x "${W}.lock" sleep 5 &
LOCK_PID=$!
sleep 0.5
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-unchanged.json" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "tbx_locked"' err.json
kill "${LOCK_PID}" 2>/dev/null
wait "${LOCK_PID}" 2>/dev/null
```

- **exit**: `3`
- Error code `tbx_locked`.
- **macOS note**: `flock(1)` is not available. SKIP on macOS; test on
  Linux or write a small Go helper that holds the lock.

---

# 17. Stream routing

## TC-STREAM-001 — success: envelope on stdout, stderr empty P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-unchanged.json" >out.json 2>err.json
echo "exit=$?"
test -s out.json && echo "stdout: non-empty"
test ! -s err.json && echo "stderr: empty"
```

- stdout has the JSON envelope.
- stderr is empty on success.

## TC-STREAM-002 — error: envelope on stderr, stdout empty P0

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-malformed.json" >out.json 2>err.json
echo "exit=$?"
test ! -s out.json && echo "stdout: empty"
test -s err.json && echo "stderr: non-empty"
```

- stderr has the error envelope.
- stdout is empty on error.

## TC-STREAM-003 — dry-run: preview on stdout, stderr empty P1

```sh
W=$(work)
$TT apply --tbx "${W}" --file "${QA_TMP}/payload-add.json" --dry-run >out.json 2>err.json
echo "exit=$?"
test -s out.json && echo "stdout: non-empty"
test ! -s err.json && echo "stderr: empty"
```

- Dry-run preview on stdout, stderr empty.

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
- Validate command still works after E8 changes.

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

## TC-REG-004 — concept add still works P0

```sh
W=$(work)
$TT concept add --tbx "${W}" \
  --id "binah" --lang en --term "binah" --status preferredTerm-admn-sts < /dev/null >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- Granular write commands still work after E8 changes.

## TC-REG-005 — check still works P0

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

## TC-REG-006 — schema includes apply command P1

```sh
$TT schema >out.json 2>err.json
echo "exit=$?"
jq -e '[.commands[] | select(.name == "apply")] | length == 1' out.json
jq -e '.commands[] | select(.name == "apply") | .flags | map(.name) | contains(["file", "prune"])' out.json
```

- Schema has `apply` command with `--file` and `--prune` flags.

---

# Cleanup

```sh
rm -rf "${QA_TMP}"
rm -f out.json err.json out1.json out2.json err1.json err2.json lookup.json val.json dev_null
```

---

# Test case summary

| Section                                | Cases  | Priority |
| -------------------------------------- | ------ | -------- |
| Basic apply — add                      | 2      | P0       |
| Basic apply — update                   | 2      | P0       |
| Basic apply — unchanged                | 1      | P0       |
| Mixed apply                            | 1      | P0       |
| Idempotency                            | 2      | P0       |
| Concept equality                       | 3      | P1       |
| Prune                                  | 3      | P0       |
| Payload formats                        | 5      | P0–P2   |
| Dry-run                                | 3      | P0–P1   |
| Transaction records                    | 3      | P0–P1   |
| All-or-nothing atomicity               | 1      | P0       |
| apply_validation_failed envelope       | 1      | P0       |
| Error cases                            | 6      | P0–P1   |
| Output envelope shape                  | 3      | P0       |
| Determinism — sorted output            | 1      | P0       |
| Concurrency — file locking             | 1      | P1       |
| Stream routing                         | 3      | P0–P1   |
| Regression — previous commands         | 6      | P0–P1   |
| **Total**                              | **47** |          |
