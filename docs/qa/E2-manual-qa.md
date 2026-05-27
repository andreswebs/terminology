# Test Plan: E2 — Domain model & TBX I/O manual QA

> **Status**: ready to execute once E2 is completed
> Run end-to-end in a single sitting.

## Purpose

E2's deliverable is **the dialect-agnostic domain model and the first
concrete reader/writer pair for TBX-Linguist**. This QA pass exercises
the E2 surface through the `validate` command — currently the only
command that calls `tbx.Load()` — to verify loading, dialect detection,
error sentinels, legacy-form normalization, reader warnings, and output
stream routing.

This is **I/O-layer QA, not validation-logic QA**. Deep validation
semantics (tier-1/2/3 checks, picklist validation on read, full
`--strict` coverage) land in E3. Anything that isn't the documented
E2 contract (dialect detection, reader/writer round-trip, error
sentinels, domain model fidelity) is out of scope.

## Scope

### In scope

- `tbx.Load()` correctly reads TBX-Linguist DCT files.
- `tbx.Load()` correctly reads TBX-Linguist DCA files.
- Dialect detection: unsupported `@type` → `validation_error`, exit 65.
- Non-TBX and malformed XML inputs → `validation_error`, exit 65.
- File-not-found → `validation_error`, exit 65.
- Empty file → `validation_error`, exit 65.
- Missing TBX path → `no_tbx_path`, exit 2.
- `TERMINOLOGY_TBX` env var feeds `--tbx`.
- Domain model fidelity: concept count, language list, alphabetical
  sort of languages.
- Multi-concept, multi-language, multi-term fixtures load correctly.
- DCA → DCT style normalization on read (input DCA, model is
  style-agnostic).
- Legacy-form normalization: bare admin-status forms and
  `usageRegister` load without errors.
- Reader warnings (e.g. `unresolved_crossref`) surface in the validate
  envelope.
- Warnings present → exit 1 (not exit 0).
- `--strict` promotes warnings to errors → exit 65.
- Success envelopes go to **stdout**; error envelopes go to **stderr**.
- Envelope conformance: success shape
  `{schema_version, ok, concepts, languages, warnings}`.
- All canonical test fixtures load without errors.

### Out of scope

- `tbx.Save()` / atomic write / advisory lock — no write commands
  consume it yet. Tested by unit tests in E2, deferred for CLI QA
  until E7.
- `ErrTBXLocked` — requires concurrent write commands (E7+).
- Round-trip byte stability — requires save (E7+).
- Canonical DCT writer output — requires save (E7+).
- Validation logic (E3): tier-1/2/3 checks, picklist validation on
  read, unknown-element detection, invalid-lang-tag detection.
- Command surfaces beyond `validate` — other commands still return
  `under_construction`.

### Deferred — to be tested when write commands land (E7+)

| E2 deliverable            | First CLI consumer       | QA gate  |
| ------------------------- | ------------------------ | -------- |
| `tbx.Save()`              | `concept add`            | E7 QA    |
| Atomic write              | `concept add`            | E7 QA    |
| Advisory lock             | Concurrent `concept add` | E7 QA    |
| `ErrTBXLocked`            | Concurrent `concept add` | E7 QA    |
| DCT writer determinism    | `concept add` / `apply`  | E7/E8 QA |
| Round-trip byte stability | `concept add` / `apply`  | E7/E8 QA |

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
ls "${FIXTURES}/canonical/" "${FIXTURES}/normalized/"
```

If any of steps 1–4 fails, **stop**: build or fixture tree is broken.

## Entry criteria

- All E2 tickets closed.
- `make build` exits 0.
- `cd src && go test ./...` exits 0.
- Canonical fixtures exist in
  `src/internal/tbx/linguist/testdata/canonical/`.

## Exit criteria

- Every **P0** test case passes — no exceptions.
- Every **P1** test case passes — no exceptions.
- Every **P2** test case passes, OR a follow-up ticket is filed with a
  reproducer.
- Every **P3** test case is run and recorded; failures noted but not
  blocking.

## Risk areas

| Risk                                                | Mitigation                                                    |
| --------------------------------------------------- | ------------------------------------------------------------- |
| DCA reader produces different model than DCT        | TC-LOAD-DCA-001 compares DCT and DCA output for same content  |
| Dialect detection false-positive on non-TBX XML     | TC-ERR-004 / TC-ERR-005 cover non-TBX inputs                  |
| Legacy normalization silently drops data            | TC-LEGACY-001 loads fixture, verifies concept/language counts |
| Reader warnings swallowed or mis-routed             | TC-WARN-001 asserts warnings in stdout envelope               |
| Success envelope on wrong stream (stdout vs stderr) | TC-STREAM-001 / TC-STREAM-002 check stream routing            |
| `--strict` not promoting warnings correctly         | TC-STRICT-001 checks exit 65 on file with known warnings      |
| `TERMINOLOGY_TBX` env not wired                     | TC-ENV-001 verifies env-var fallback                          |

## Conventions

### Success envelope shape

```json
{
  "schema_version": 1,
  "ok": true,
  "concepts": <int>,
  "languages": ["<sorted>", "..."],
  "warnings": []
}
```

- `schema_version` is always `1`.
- `ok` is `true` for clean or warning-bearing runs.
- `concepts` is the number of `<conceptEntry>` elements.
- `languages` is an alphabetically sorted array of BCP 47 tags.
- `warnings` is an array (empty `[]` when no warnings, never `null`).
- Success envelope → **stdout**.

### Warning-bearing envelope shape

```json
{
  "schema_version": 1,
  "ok": true,
  "concepts": <int>,
  "languages": ["..."],
  "warnings": [
    {
      "code": "<warning_code>",
      "message": "<description>",
      "concept_id": "<optional>",
      "line": <optional_int>,
      "column": <optional_int>
    }
  ]
}
```

- Exit code is **1** when warnings are present.
- Envelope goes to **stdout** (warnings are results, not errors).

### Error envelope shape

```json
{
  "schema_version": 1,
  "ok": false,
  "error": {
    "code": "<error_code>",
    "message": "<description>",
    "hint": "<optional>"
  }
}
```

- Error envelope → **stderr**.

### Exit code map (E2 surface)

| Code | Meaning          | E2 source                                              |
| ---- | ---------------- | ------------------------------------------------------ |
| 0    | success          | Valid TBX loaded, no warnings                          |
| 1    | warnings         | Valid TBX loaded, warnings present                     |
| 2    | usage error      | `no_tbx_path` (missing `--tbx` / `TERMINOLOGY_TBX`)    |
| 65   | validation_error | Bad TBX: unsupported dialect, malformed XML, not found |

### How to read a test case

Each case has the form:

```
TC-<MODULE>-NNN — <name>                          P0|P1|P2|P3
$ <argv>                                          (paste-runnable)
exit=<n>
stdout: <expectation>
stderr: <expectation>
```

---

# 1. Pre-flight

## TC-PRE-001 — build cleanly P0

```sh
make build
echo "exit=$?"
```

- **exit**: `0`
- **artefact**: `bin/terminology-<host>-<arch>` exists.

## TC-PRE-002 — binary boots P0

```sh
$TT --version
echo "exit=$?"
```

- **exit**: `0`

## TC-PRE-003 — fixtures exist P0

```sh
ls "${FIXTURES}/canonical/minimal-dct.tbx" \
   "${FIXTURES}/canonical/minimal-dca.tbx" \
   "${FIXTURES}/canonical/rich-dct.tbx" \
   "${FIXTURES}/canonical/full-features.tbx" \
   "${FIXTURES}/canonical/with-transactions.tbx" \
   "${FIXTURES}/canonical/all-categories-dct.tbx" \
   "${FIXTURES}/canonical/all-categories-dca.tbx" \
   "${FIXTURES}/normalized/legacy-forms.tbx"
echo "exit=$?"
```

- **exit**: `0` — all 8 files present.

---

# 2. TBX loading — DCT style

## TC-LOAD-DCT-001 — minimal DCT loads cleanly P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null
echo "exit=$?"
```

- **exit**: `0`
- **stdout**: valid JSON envelope with `ok: true`.

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and .concepts == 1 and .languages == ["en","he"] and .warnings == []'
```

- **jq_exit**: `0`

## TC-LOAD-DCT-002 — rich DCT (multi-concept, multi-language) P0

```sh
$TT --tbx "${FIXTURES}/canonical/rich-dct.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and .concepts == 2 and .languages == ["en","es","he"] and .warnings == []'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- 2 concepts (`malkhut`, `tzimtzum`), 3 languages (`en`, `es`, `he`),
  sorted alphabetically.

## TC-LOAD-DCT-003 — full-features DCT P1

```sh
$TT --tbx "${FIXTURES}/canonical/full-features.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and .concepts >= 1 and (.languages | length) >= 1 and (.warnings | length) == 0'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-LOAD-DCT-004 — with-transactions DCT P1

```sh
$TT --tbx "${FIXTURES}/canonical/with-transactions.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and .concepts >= 1'
echo "jq_exit=$?"
```

- **jq_exit**: `0` — transaction groups don't break loading.

## TC-LOAD-DCT-005 — all-categories DCT P1

```sh
$TT --tbx "${FIXTURES}/canonical/all-categories-dct.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and .concepts >= 1'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Warnings may be present (e.g. `unresolved_crossref`); the load
  itself must succeed.

---

# 3. TBX loading — DCA style

## TC-LOAD-DCA-001 — minimal DCA loads cleanly P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dca.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and .concepts == 1 and .languages == ["en","he"] and .warnings == []'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Same content as `minimal-dct.tbx` — validates that DCA reader
  produces equivalent domain model.

## TC-LOAD-DCA-002 — all-categories DCA P1

```sh
$TT --tbx "${FIXTURES}/canonical/all-categories-dca.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and .concepts >= 1'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-LOAD-DCA-003 — DCT and DCA produce same model P0

Verify that the minimal DCT and DCA fixtures produce identical
validate output (same concept count, same languages).

```sh
dct=$($TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq '{concepts, languages, warnings}')
dca=$($TT --tbx "${FIXTURES}/canonical/minimal-dca.tbx" validate 2>/dev/null \
  | jq '{concepts, languages, warnings}')
if [ "$dct" = "$dca" ]; then echo "PASS"; else echo "FAIL: DCT≠DCA"; fi
```

- **output**: `PASS`

---

# 4. Legacy-form normalization

## TC-LEGACY-001 — legacy status forms load without errors P0

```sh
$TT --tbx "${FIXTURES}/normalized/legacy-forms.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and .concepts == 1 and .languages == ["en"]'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- The fixture contains bare `preferredTerm`, `admittedTerm`,
  `deprecatedTerm`, `supersededTerm`, and `usageRegister` — all must
  normalize on read without producing errors.

## TC-LEGACY-002 — legacy normalization does not produce warnings P1

```sh
$TT --tbx "${FIXTURES}/normalized/legacy-forms.tbx" validate 2>/dev/null \
  | jq -e '.warnings == []'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Exit code: `0` (no warnings, no errors).

---

# 5. Error sentinels

## TC-ERR-001 — no TBX path P0

```sh
$TT validate 2>err.json >/dev/null; echo "exit=$?"
jq -e '.error.code == "no_tbx_path" and .error.hint == "set --tbx or TERMINOLOGY_TBX"' err.json
echo "jq_exit=$?"
```

- **exit**: `2`
- **jq_exit**: `0`
- **stdout**: empty.

## TC-ERR-002 — non-existent file P0

```sh
$TT --tbx /tmp/nonexistent-$(date +%s).tbx validate 2>err.json >/dev/null
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
```

- **exit**: `65`
- **jq_exit**: `0`

## TC-ERR-003 — empty file P0

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

## TC-ERR-004 — unsupported dialect (TBX-Basic) P0

```sh
WRONG_TBX=$(mktemp /tmp/wrong-XXXXXX.tbx)
cat > "${WRONG_TBX}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Basic" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader>
    <fileDesc>
      <sourceDesc><p>wrong dialect</p></sourceDesc>
    </fileDesc>
  </tbxHeader>
  <text><body/></text>
</tbx>
XML
$TT --tbx "${WRONG_TBX}" validate 2>err.json >/dev/null
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
rm -f "${WRONG_TBX}"
```

- **exit**: `65`
- **jq_exit**: `0`
- The underlying `unsupported_dialect` is wrapped by
  `validation_error` in the validate command.

## TC-ERR-005 — non-TBX XML P1

```sh
NOT_TBX=$(mktemp /tmp/not-tbx-XXXXXX.xml)
echo '<?xml version="1.0"?><html><body>not tbx</body></html>' > "${NOT_TBX}"
$TT --tbx "${NOT_TBX}" validate 2>err.json >/dev/null
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
rm -f "${NOT_TBX}"
```

- **exit**: `65`
- **jq_exit**: `0`

## TC-ERR-006 — malformed XML (not well-formed) P1

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

## TC-ERR-007 — TBX element with no type attribute P1

```sh
NO_TYPE=$(mktemp /tmp/notype-XXXXXX.tbx)
cat > "${NO_TYPE}" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<tbx xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2">
  <tbxHeader>
    <fileDesc>
      <sourceDesc><p>no type attr</p></sourceDesc>
    </fileDesc>
  </tbxHeader>
  <text><body/></text>
</tbx>
XML
$TT --tbx "${NO_TYPE}" validate 2>err.json >/dev/null
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
rm -f "${NO_TYPE}"
```

- **exit**: `65`
- **jq_exit**: `0`
- A `<tbx>` element without `@type` is unsupported.

## TC-ERR-008 — directory instead of file P2

```sh
$TT --tbx /tmp validate 2>err.json >/dev/null
echo "exit=$?"
jq -e '.ok == false' err.json
echo "jq_exit=$?"
```

- **exit**: non-zero (`65` for load error or `validation_error`).
- **jq_exit**: `0`

---

# 6. Environment variable source

## TC-ENV-001 — `TERMINOLOGY_TBX` feeds `--tbx` P0

```sh
env -i SHELL=$SHELL HOME=$HOME PATH=$PATH \
  TERMINOLOGY_TBX="${FIXTURES}/canonical/minimal-dct.tbx" \
  $TT validate 2>/dev/null \
  | jq -e '.ok == true and .concepts == 1'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- No `--tbx` flag on the command line; env var provides the path.

## TC-ENV-002 — `--tbx` flag overrides `TERMINOLOGY_TBX` P1

```sh
env -i SHELL=$SHELL HOME=$HOME PATH=$PATH \
  TERMINOLOGY_TBX=/tmp/nonexistent.tbx \
  $TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and .concepts == 1'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- The flag wins over the env var.

---

# 7. Reader warnings

## TC-WARN-001 — unresolved cross-references surface as warnings P0

```sh
$TT --tbx "${FIXTURES}/canonical/all-categories-dct.tbx" validate 2>/dev/null \
  | jq -e '.ok == true and (.warnings | length) > 0 and (.warnings[] | select(.code == "unresolved_crossref") | .code) == "unresolved_crossref"'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Warnings appear in the envelope's `.warnings` array.

## TC-WARN-002 — warnings produce exit 1 P0

```sh
$TT --tbx "${FIXTURES}/canonical/all-categories-dct.tbx" validate 2>/dev/null >/dev/null
echo "exit=$?"
```

- **exit**: `1` (not `0`).

## TC-WARN-003 — warning shape has required fields P1

```sh
$TT --tbx "${FIXTURES}/canonical/all-categories-dct.tbx" validate 2>/dev/null \
  | jq -e '.warnings[0] | has("code") and has("message")'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Each warning has at least `code` and `message`. `concept_id`,
  `line`, `column` are optional.

## TC-WARN-004 — warning has concept_id when applicable P2

```sh
$TT --tbx "${FIXTURES}/canonical/all-categories-dct.tbx" validate 2>/dev/null \
  | jq -e '.warnings[] | select(.code == "unresolved_crossref") | .concept_id != null and .concept_id != ""'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Cross-reference warnings should identify which concept.

---

# 8. `--strict` flag

## TC-STRICT-001 — promotes warnings to errors P0

```sh
$TT --tbx "${FIXTURES}/canonical/all-categories-dct.tbx" validate --strict 2>err.json >/dev/null
echo "exit=$?"
jq -e '.error.code == "validation_error"' err.json
echo "jq_exit=$?"
```

- **exit**: `65`
- **jq_exit**: `0`
- A file that loads with warnings in lenient mode fails in strict mode.

## TC-STRICT-002 — clean file passes in strict mode P1

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate --strict 2>/dev/null \
  | jq -e '.ok == true and .warnings == []'
echo "jq_exit=$?"
```

- **jq_exit**: `0`
- Exit: `0` — `--strict` doesn't reject a clean file.

---

# 9. Output stream routing

## TC-STREAM-001 — success envelope goes to stdout P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate >out.txt 2>err.txt
echo "exit=$?"
echo "stdout_has_json=$(jq -e '.ok == true' out.txt >/dev/null 2>&1 && echo yes || echo no)"
echo "stderr_empty=$([ ! -s err.txt ] && echo yes || echo no)"
rm -f out.txt err.txt
```

- **exit**: `0`
- **stdout_has_json**: `yes`
- **stderr_empty**: `yes`

## TC-STREAM-002 — error envelope goes to stderr P0

```sh
$TT validate >out.txt 2>err.txt
echo "exit=$?"
echo "stderr_has_json=$(jq -e '.ok == false' err.txt >/dev/null 2>&1 && echo yes || echo no)"
echo "stdout_empty=$([ ! -s out.txt ] && echo yes || echo no)"
rm -f out.txt err.txt
```

- **exit**: `2`
- **stderr_has_json**: `yes`
- **stdout_empty**: `yes`

## TC-STREAM-003 — warning-bearing envelope: stdout has payload, stderr has warning exit P1

```sh
$TT --tbx "${FIXTURES}/canonical/all-categories-dct.tbx" validate >out.txt 2>err.txt
echo "exit=$?"
echo "stdout_has_warnings=$(jq -e '(.warnings | length) > 0' out.txt >/dev/null 2>&1 && echo yes || echo no)"
rm -f out.txt err.txt
```

- **exit**: `1`
- **stdout_has_warnings**: `yes`
- The result envelope (with warnings) goes to stdout.

---

# 10. Envelope conformance

## TC-ENVELOPE-001 — success envelope shape P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e 'has("schema_version") and has("ok") and has("concepts") and has("languages") and has("warnings")'
echo "jq_exit=$?"
```

- **jq_exit**: `0` — all required keys present.

## TC-ENVELOPE-002 — schema_version is 1 P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '.schema_version == 1'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-ENVELOPE-003 — languages is always an array, never null P1

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '.languages | type == "array"'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-ENVELOPE-004 — warnings is always an array, never null P1

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate 2>/dev/null \
  | jq -e '.warnings | type == "array"'
echo "jq_exit=$?"
```

- **jq_exit**: `0`

## TC-ENVELOPE-005 — error envelope shape P0

```sh
$TT validate 2>&1 >/dev/null \
  | jq -e 'has("schema_version") and has("ok") and has("error") and (.error | has("code") and has("message"))'
echo "jq_exit=$?"
```

- **jq_exit**: `0` — error envelope has required keys.

## TC-ENVELOPE-006 — error envelope hint field present when applicable P1

```sh
$TT validate 2>&1 >/dev/null \
  | jq -e '.error.hint != null and .error.hint != ""'
echo "jq_exit=$?"
```

- **jq_exit**: `0` — `no_tbx_path` has a hint.

---

# 11. Text format rendering

## TC-TEXT-001 — error in text format P1

```sh
$TT --format text validate 2>&1 >/dev/null | head -2
echo "exit=$?"
```

- **exit**: `2`
- **stderr**: starts with `✗ `, contains `no TBX file path provided`,
  followed by `  hint: set --tbx or TERMINOLOGY_TBX`.

## TC-TEXT-002 — validation_error in text format P2

```sh
WRONG_TBX=$(mktemp /tmp/wrong-XXXXXX.tbx)
echo '<?xml version="1.0"?><tbx type="TBX-Basic"><tbxHeader><fileDesc><sourceDesc><p>t</p></sourceDesc></fileDesc></tbxHeader><text><body/></text></tbx>' > "${WRONG_TBX}"
$TT --format text --tbx "${WRONG_TBX}" validate 2>&1 >/dev/null | head -2
echo "exit=$?"
rm -f "${WRONG_TBX}"
```

- **exit**: `65`
- **stderr**: starts with `✗ `, contains `TBX validation failed`.

---

# 12. All fixtures matrix

Iterate all canonical fixtures through `validate` to ensure none fail
to load. Any failure is a P0 finding.

```sh
for f in \
  "${FIXTURES}/canonical/minimal-dct.tbx" \
  "${FIXTURES}/canonical/minimal-dca.tbx" \
  "${FIXTURES}/canonical/rich-dct.tbx" \
  "${FIXTURES}/canonical/full-features.tbx" \
  "${FIXTURES}/canonical/with-transactions.tbx" \
  "${FIXTURES}/canonical/all-categories-dct.tbx" \
  "${FIXTURES}/canonical/all-categories-dca.tbx" \
  "${FIXTURES}/normalized/legacy-forms.tbx"; do
  $TT --tbx "$f" validate 2>/dev/null | jq -e '.ok == true' >/dev/null 2>&1
  rc=$?
  name=$(basename "$f")
  if [ "$rc" -eq 0 ]; then echo "ok: $name"; else echo "FAIL: $name"; fi
done
```

Every line should print `ok: ...`. Any `FAIL` is a P0 regression.

---

# 13. Sign-off checklist

Tick before declaring E2 manual QA pass.

- [ ] Section 1 (pre-flight): all cases pass.
- [ ] Section 2 (DCT loading): all P0 + P1 cases pass.
- [ ] Section 3 (DCA loading): all P0 + P1 cases pass; DCT≡DCA check
      passes.
- [ ] Section 4 (legacy normalization): loads without errors or
      warnings.
- [ ] Section 5 (error sentinels): all P0 + P1 cases pass; every error
      returns proper envelope with correct exit code.
- [ ] Section 6 (env var): `TERMINOLOGY_TBX` feeds `--tbx`; flag
      overrides env.
- [ ] Section 7 (reader warnings): warnings surface in envelope;
      warnings → exit 1.
- [ ] Section 8 (`--strict`): promotes warnings to errors (exit 65);
      clean files pass.
- [ ] Section 9 (stream routing): success → stdout, errors → stderr.
- [ ] Section 10 (envelope conformance): all shapes correct.
- [ ] Section 11 (text format): renders `✗` + `hint:` correctly.
- [ ] Section 12 (all fixtures matrix): every fixture loads
      successfully.
- [ ] No undocumented behaviour observed.

## On failure

Any failing case must be classified:

- **I/O-layer bug** — reader, writer, dialect detection, or error
  sentinel is wrong. File a new ticket tagged `e2,bug` linked to
  `ter-uqyn` and the offending task ticket.
- **Integration bug** — the validate command mis-routes output or
  wraps an error incorrectly. File a ticket tagged `e2,bug,integration`.
- **Spec ambiguity** — the test plan and the spec disagree. Update
  the spec **first** (PR), then re-derive the test case.

Do **not** silence a failing case by updating the test plan to match
buggy behaviour. The plan is derived from
[`docs/specs/002-domain-and-io.md`](../docs/specs/002-domain-and-io.md);
the spec wins.
