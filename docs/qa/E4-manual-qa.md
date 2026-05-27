# Test Plan: E4 — Read commands manual QA

> **Status**: ready to execute once E4 is completed
> Run end-to-end in a single sitting.

## Purpose

E4's deliverable is **three read commands** (`lookup`, `schema`, `extract`)
plus shared infrastructure (`--fields` projection, `internal/markdown` spans,
`ErrInvalidField` sentinel, terr sentinel registry).

This is **read-command QA**. E3 QA already covered validation-layer concerns
(tiers, `--strict`, picklist checks, line/col). E4 tests the lookup matching
logic, schema reflective introspection, extract heuristic engine, `--fields`
projection, and the envelope shapes for all three commands.

## Scope

### In scope

- `lookup TERM` — case-fold + NFC exact matching, `--lang` filter,
  not-found exit 1, `results: []` envelope.
- `schema` — reflective introspection over live command tree, output
  envelope types, and terr sentinel registry. `--command NAME` filter.
- `extract FILE...` — three heuristics (capitalized phrases, foreign-script
  tokens, high-frequency tokens), `--exclude`, `--script`, `--lang`,
  `--stopwords`, `--min-freq`, markdown awareness (code blocks excluded).
- `--fields` projection — path validation against json tags, wildcard `*`,
  `invalid_field` error on typos (exit 2).
- `ErrInvalidField` sentinel — code `invalid_field`, exit 2.
- Envelope shapes for all three commands.
- Exit codes: 0 (results found / success), 1 (no results / no candidates),
  2 (usage error).
- Language detection precedence for extract: frontmatter `lang:` →
  `--lang` flag → default `en`.

### Out of scope

- Validation logic (`validate` command) — covered by E3 QA.
- Domain model / I/O layer — covered by E2 QA.
- Matcher-based commands (`scan`, `check`) — E6.
- Write commands (`concept`, `term`, `apply`) — E7/E8.
- Bundled stoplists — explicitly out of v1 scope.
- NDJSON streaming — v2 candidate.

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

Create test markdown files for extract tests:

````sh
export QA_TMP=$(mktemp -d /tmp/e4-qa-XXXXXX)

cat > "${QA_TMP}/corpus.md" <<'EOF'
# Introduction

The Holy Temple was an important structure. The Dead Sea Scrolls were
discovered near the Dead Sea. The concept of צמצום is central to Kabbalah.

The Holy Temple appears again in later chapters. The Dead Sea Scrolls
have been studied extensively. The term צמצום represents divine
contraction.

Some common words appear often. Temple is mentioned in many contexts.
Temple architecture varied across periods. Temple rituals were complex.
EOF

cat > "${QA_TMP}/code-blocks.md" <<'EOF'
# Technical Documentation

The Holy Temple was significant.

```go
func getUserById(id string) *User {
    return db.FindUser(id)
}
````

The `getUserById` function retrieves users.
EOF

cat > "${QA_TMP}/empty.md" <<'EOF'

# Nothing here

Just some ordinary text with no interesting terms.
EOF

cat > "${QA_TMP}/hebrew-frontmatter.md" <<'EOF'
---
lang: he
---

# מבוא

הצמצום הוא concept מרכזי. The Kabbalah teaches about divine light.
EOF

cat > "${QA_TMP}/stopwords.txt" <<'EOF'
the
a
an
is
was
were
have
been
EOF

````

If any setup step fails, **stop**: build or fixture tree is broken.

## Entry criteria

- All E4 tickets closed.
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

| Risk                                              | Mitigation                                       |
| ------------------------------------------------- | ------------------------------------------------ |
| Case-fold or NFC mismatch on non-Latin scripts    | TC-LKP-002/003 test Hebrew + accented terms      |
| `--fields` silently drops data on bad path syntax | TC-FLD-001/002 verify error on typo              |
| Schema drift from live binary                     | TC-SCH-001/002 verify commands array is populated |
| Extract code-block leakage                        | TC-EXT-002 verifies no code identifiers           |
| Foreign-script detection false positives           | TC-FST-001/002 test Hebrew-in-Latin explicitly    |
| Stopwords applied to wrong heuristic              | TC-FREQ-002 verifies only frequency heuristic     |
| Language detection precedence wrong               | TC-LANG-001/002/003 test all three levels         |

## Conventions

Same as E2/E3 QA: see [E2-manual-qa.md](E2-manual-qa.md) §Conventions for
envelope shapes, exit code map, and test case format.

### Exit code map (E4 surface)

| Code | Meaning     | Source                                              |
| ---- | ----------- | --------------------------------------------------- |
| 0    | success     | Results/candidates found, or schema emitted          |
| 1    | no results  | lookup: not found; extract: no candidates            |
| 2    | usage error | `no_tbx_path`, `invalid_field`, unknown `--command`  |

---

# 1. Lookup — basic matching

## TC-LKP-001 — exact match found P0

```sh
TERM="tzimtzum"  # or a known term in your fixture
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM}" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.results | length > 0' out.json
jq -e '.schema_version == 1' out.json
````

- **exit**: `0`
- `results` array is non-empty.
- `ok` is `true`.
- `schema_version` is `1`.

## TC-LKP-002 — case-insensitive match P0

```sh
TERM_UPPER="TZIMTZUM"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM_UPPER}" >out.json 2>err.json
echo "exit=$?"
jq -e '.results | length > 0' out.json
```

- **exit**: `0`
- Same results as TC-LKP-001 — case-fold matches.

## TC-LKP-003 — NFC normalization match P1

```sh
# If your fixture has a term with combining characters (e.g. café),
# query with the precomposed form.
TERM_NFC="café"
$TT --tbx "${FIXTURES}/canonical/all-categories-dct.tbx" lookup "${TERM_NFC}" >out.json 2>err.json
echo "exit=$?"
# Expected: match if the term exists in any normalization form.
```

- **exit**: `0` if term exists, `1` if not.
- NFC normalization applied to both query and glossary.

## TC-LKP-004 — not found returns empty results and exit 1 P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "xyznonexistent" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.results == []' out.json
```

- **exit**: `1`
- `ok` is `true` (operation succeeded, just nothing found).
- `results` is `[]` (empty array, not null).

## TC-LKP-005 — no --tbx flag P0

```sh
$TT lookup "anything" >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "no_tbx_path"' err.json
```

- **exit**: `2`
- Error envelope on stderr with `code: "no_tbx_path"`.

---

# 2. Lookup — --lang filter

## TC-LANG-LKP-001 — --lang restricts search to that language's terms P1

`--lang` restricts which language sections are **searched**, not just
which sections appear in the output. To match within Hebrew, query with
the Hebrew-language term.

```sh
TERM_HE="צמצום"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM_HE}" --lang he >out.json 2>err.json
echo "exit=$?"
jq -e '.results | length > 0' out.json
jq -e '.results[0] | has("concept_id", "languages")' out.json

# Verify the English term is NOT found when scoped to Hebrew
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "tzimtzum" --lang he >out2.json 2>err2.json
echo "cross-lang exit=$?"
jq -e '.results == []' out2.json
```

- **exit**: `0` — Hebrew term matched within Hebrew language section.
- Full concept returned (all language sections included in output).
- Cross-language check: querying with the English term (`tzimtzum`)
  and `--lang he` returns empty (exit 1), because `--lang` scopes the
  term search to that language only.

## TC-LANG-LKP-002 — --lang with no match in that language P1

```sh
TERM="tzimtzum"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM}" --lang fr >out.json 2>err.json
echo "exit=$?"
jq -e '.results == []' out.json
```

- **exit**: `1`
- No results when term doesn't exist in requested language.

---

# 3. Lookup — envelope shape

## TC-LKP-ENV-001 — envelope has correct top-level keys P0

```sh
TERM="tzimtzum"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM}" >out.json 2>err.json
echo "exit=$?"
jq -e 'keys == ["ok","results","schema_version"]' out.json
```

- **exit**: `0`
- Top-level keys are exactly `ok`, `results`, `schema_version`.

## TC-LKP-ENV-002 — result shape matches spec P1

```sh
TERM="tzimtzum"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM}" >out.json 2>err.json
echo "exit=$?"
jq -e '.results[0] | has("concept_id", "languages")' out.json
jq -e '.results[0].languages | to_entries[0].value | has("preferred")' out.json
```

- **exit**: `0`
- Each result has `concept_id` and `languages`.
- Each language entry has `preferred` (and optionally `admitted`).

---

# 4. --fields projection

## TC-FLD-001 — valid field path projects output P1

Field paths are **envelope-relative** (e.g. `results.concept_id`, not
bare `concept_id`). Use `terminology schema --command CMD` to discover
valid paths.

```sh
TERM="tzimtzum"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM}" --fields results.concept_id >out.json 2>err.json
echo "exit=$?"
jq -e '.results[0] | has("concept_id")' out.json
jq -e '.results[0] | has("languages") | not' out.json
```

- **exit**: `0`
- Output includes only `concept_id` in results (plus envelope boilerplate).
- `languages` and other unprojected fields absent.

## TC-FLD-002 — invalid field path returns error P0

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "anything" --fields concpet_id >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code == "invalid_field"' err.json
jq -e '.error.hint | test("schema")' err.json
```

- **exit**: `2`
- Error code `invalid_field`.
- Hint mentions `terminology schema`.

## TC-FLD-003 — wildcard path works P2

```sh
TERM="tzimtzum"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM}" --fields "results.languages.*.preferred.term" >out.json 2>err.json
echo "exit=$?"
jq -e '.results[0].languages | to_entries[0].value.preferred.term' out.json
```

- **exit**: `0`
- Wildcard `*` traverses map keys correctly.

---

# 5. Schema — full output

## TC-SCH-001 — full schema output has required keys P0

```sh
$TT schema >out.json 2>err.json
echo "exit=$?"
jq -e 'has("schema_version", "commands", "envelopes", "error_codes")' out.json
jq -e '.schema_version == 1' out.json
```

- **exit**: `0`
- Top-level keys: `schema_version`, `commands`, `envelopes`, `error_codes`.

## TC-SCH-002 — commands array contains all commands P0

```sh
$TT schema >out.json 2>err.json
CMDS=$(jq -r '[.commands[].name] | sort | join(",")' out.json)
echo "commands=${CMDS}"
# Verify known commands are present
echo "${CMDS}" | grep -q "validate" && echo "validate: ok"
echo "${CMDS}" | grep -q "lookup" && echo "lookup: ok"
echo "${CMDS}" | grep -q "schema" && echo "schema: ok"
echo "${CMDS}" | grep -q "extract" && echo "extract: ok"
```

- All four commands present in the `commands` array.

## TC-SCH-003 — error codes enumerated P1

```sh
$TT schema >out.json 2>err.json
jq -e '.error_codes | length > 0' out.json
jq -e '[.error_codes[].code] | contains(["validation_error", "no_tbx_path", "invalid_field"])' out.json
```

- `error_codes` array is non-empty.
- Contains at least `validation_error`, `no_tbx_path`, `invalid_field`.

## TC-SCH-004 — envelopes map populated P1

```sh
$TT schema >out.json 2>err.json
jq -e '.envelopes | has("validate")' out.json
jq -e '.envelopes | has("lookup")' out.json
```

- `envelopes` map has entries for `validate` and `lookup`.

## TC-SCH-005 — no --tbx required P0

```sh
# schema is reflective, not data-dependent — should work without --tbx
$TT schema >out.json 2>err.json
echo "exit=$?"
jq -e '.ok // .schema_version == 1' out.json
```

- **exit**: `0`
- Schema emitted without any TBX file.

---

# 6. Schema — --command filter

## TC-SCH-CMD-001 — filter to single command P1

```sh
$TT schema --command validate >out.json 2>err.json
echo "exit=$?"
jq -e '.name == "validate"' out.json
jq -e 'has("flags")' out.json
jq -e 'has("envelope")' out.json
```

- **exit**: `0`
- Output has `name`, `flags`, `envelope` for the filtered command.
- Does NOT have the `commands` array (single-command view).

## TC-SCH-CMD-002 — flags detail for validate P2

```sh
$TT schema --command validate >out.json 2>err.json
jq -e '[.flags[].name] | contains(["strict"])' out.json
```

- `--strict` flag appears in the validate command's flags list.

## TC-SCH-CMD-003 — unknown command returns error P0

```sh
$TT schema --command nonexistent >out.json 2>err.json
echo "exit=$?"
jq -e '.error.code' err.json
```

- **exit**: `2`
- Error envelope with usage-error code.

---

# 7. Extract — capitalized phrases

## TC-CAP-001 — detects capitalized phrases mid-sentence P0

```sh
$TT extract "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.candidates | length > 0' out.json
jq -e '[.candidates[].term] | contains(["Holy Temple"])' out.json
jq -e '[.candidates[].term] | contains(["Dead Sea Scrolls"])' out.json
```

- **exit**: `0`
- `Holy Temple` and `Dead Sea Scrolls` appear as candidates.

## TC-CAP-002 — sentence-start words not flagged P2

```sh
$TT extract "${QA_TMP}/corpus.md" >out.json 2>err.json
# "The" at sentence start should not be a candidate
jq -e '[.candidates[].term] | map(select(. == "The")) | length == 0' out.json
```

- Bare "The" (sentence-start) is not a candidate.

## TC-CAP-003 — frequency tracked P1

```sh
$TT extract "${QA_TMP}/corpus.md" >out.json 2>err.json
jq -e '.candidates[] | select(.term == "Holy Temple") | .frequency >= 2' out.json
```

- `Holy Temple` appears at least twice → frequency >= 2.

---

# 8. Extract — foreign-script tokens

## TC-FST-001 — Hebrew in Latin text detected P0

```sh
$TT extract "${QA_TMP}/corpus.md" >out.json 2>err.json
jq -e '[.candidates[].term] | any(test("צמצום"))' out.json
```

- Hebrew token `צמצום` detected as foreign-script candidate.

## TC-FST-002 — --script filter P1

```sh
$TT extract --script hebrew "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
jq -e '.candidates | length > 0' out.json
# All candidates should be Hebrew-script
jq -e '[.candidates[].term] | all(test("[\\p{Hebrew}]"))' out.json
```

- **exit**: `0`
- Only Hebrew-script candidates returned.

## TC-FST-003 — --script latin excludes Hebrew P2

```sh
$TT extract --script latin "${QA_TMP}/corpus.md" >out.json 2>err.json
jq -e '[.candidates[].term] | all(test("[\\p{Hebrew}]")) | not' out.json
```

- No Hebrew-script candidates when filtering to Latin.

---

# 9. Extract — high-frequency tokens

## TC-FREQ-001 — --min-freq threshold gates output P1

```sh
$TT extract --min-freq 10 "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
# With min-freq 10, most frequency-based candidates should be filtered out
jq '.candidates | length' out.json
```

- Only candidates appearing 10+ times survive (likely very few or zero from
  the test corpus).

## TC-FREQ-002 — --stopwords filters common words P1

```sh
$TT extract --stopwords "${QA_TMP}/stopwords.txt" --min-freq 1 "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
# Words in stopwords.txt should not appear as frequency candidates
jq -e '[.candidates[] | select(.heuristic == "high_frequency") | .term] | map(select(. == "the" or . == "was" or . == "were")) | length == 0' out.json
```

- Stopwords excluded from high-frequency candidates.
- Stopwords do NOT affect capitalized-phrase or foreign-script heuristics.

## TC-FREQ-003 — default --min-freq is 3 P2

```sh
$TT extract "${QA_TMP}/corpus.md" >out.json 2>err.json
# Frequency candidates should only include terms appearing 3+ times
jq -e '[.candidates[] | select(.heuristic == "high_frequency") | .frequency] | all(. >= 3)' out.json
```

- All high-frequency candidates have frequency >= 3 (default threshold).

---

# 10. Extract — markdown awareness

## TC-EXT-MD-001 — code blocks excluded P0

```sh
$TT extract "${QA_TMP}/code-blocks.md" >out.json 2>err.json
echo "exit=$?"
# getUserById should not appear as a candidate
jq -e '[.candidates[].term] | map(select(test("getUserById"))) | length == 0' out.json
# But "Holy Temple" should still be found
jq -e '[.candidates[].term] | any(. == "Holy Temple")' out.json
```

- **exit**: `0`
- `getUserById` (from fenced code block) not in candidates.
- `getUserById` (from inline code) not in candidates.
- `Holy Temple` (from prose) still detected.

## TC-EXT-MD-002 — inline code excluded P1

```sh
$TT extract "${QA_TMP}/code-blocks.md" >out.json 2>err.json
jq -e '[.candidates[].term] | map(select(test("getUserById"))) | length == 0' out.json
```

- Inline code tokens excluded.

---

# 11. Extract — --exclude glossary terms

## TC-EXCL-001 — exclude terms from TBX P1

```sh
TERM_IN_TBX="tzimtzum"  # a term known to be in the fixture TBX
$TT extract --exclude "${FIXTURES}/canonical/minimal-dct.tbx" "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
# The excluded term should not appear in candidates
jq -e "[.candidates[].term] | map(select(test(\"${TERM_IN_TBX}\"))) | length == 0" out.json
```

- Terms already in the glossary are excluded from candidates.

## TC-EXCL-002 — non-excluded terms still appear P2

```sh
$TT extract --exclude "${FIXTURES}/canonical/minimal-dct.tbx" "${QA_TMP}/corpus.md" >out.json 2>err.json
jq -e '.candidates | length > 0' out.json
jq -e '[.candidates[].term] | any(. == "Holy Temple")' out.json
```

- Non-glossary terms still appear as candidates.

---

# 12. Extract — language detection

## TC-LANG-DET-001 — frontmatter lang detected P1

```sh
$TT extract "${QA_TMP}/hebrew-frontmatter.md" >out.json 2>err.json
echo "exit=$?"
# Should detect lang: he from frontmatter and adjust heuristics accordingly
jq -e '.ok == true' out.json
```

- **exit**: `0`
- File with `lang: he` frontmatter processed successfully.
- Foreign-script detection uses `he` as the base script.

## TC-LANG-DET-002 — --lang flag fallback P2

```sh
$TT extract --lang es "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- `--lang es` applied to file without frontmatter.

## TC-LANG-DET-003 — default to en P2

```sh
$TT extract "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
```

- **exit**: `0`
- No frontmatter, no `--lang` → defaults to `en`.

---

# 13. Extract — envelope shape

## TC-EXT-ENV-001 — envelope has correct top-level keys P0

```sh
$TT extract "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
jq -e 'has("schema_version", "ok", "candidates")' out.json
jq -e '.schema_version == 1' out.json
jq -e '.ok == true' out.json
```

- Top-level keys: `schema_version`, `ok`, `candidates`.

## TC-EXT-ENV-002 — candidate shape matches spec P1

```sh
$TT extract "${QA_TMP}/corpus.md" >out.json 2>err.json
jq -e '.candidates[0] | has("term", "frequency", "heuristic")' out.json
```

- Each candidate has `term`, `frequency`, `heuristic`.

## TC-EXT-ENV-003 — no candidates returns empty array and exit 1 P0

```sh
$TT extract "${QA_TMP}/empty.md" --min-freq 999 >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.candidates == []' out.json
```

- **exit**: `1`
- `ok` is `true`.
- `candidates` is `[]` (empty array, not null).

---

# 14. Extract — exit codes

## TC-EXT-EXIT-001 — missing files P0

```sh
$TT extract /nonexistent/file.md >out.json 2>err.json
echo "exit=$?"
```

- **exit**: non-zero (likely `2` for usage error or `1`).
- Error reported.

## TC-EXT-EXIT-002 — no file arguments P0

```sh
$TT extract >out.json 2>err.json
echo "exit=$?"
```

- **exit**: `2`
- Error envelope: missing argument.

---

# 15. --fields on extract P2

## TC-FLD-EXT-001 — field projection on extract output

```sh
$TT extract --fields candidates.term,candidates.frequency "${QA_TMP}/corpus.md" >out.json 2>err.json
echo "exit=$?"
jq -e '.candidates[0] | has("term", "frequency")' out.json
jq -e '.candidates[0] | has("heuristic") | not' out.json
```

- Only `term` and `frequency` in projected candidates.
- `heuristic` and `locations` absent.

---

# 16. Schema — error code detail

## TC-SCH-ERR-001 — error codes have required fields P1

```sh
$TT schema >out.json 2>err.json
jq -e '.error_codes[0] | has("code", "exit_code", "message")' out.json
```

- Each error code entry has `code`, `exit_code`, `message`.

## TC-SCH-ERR-002 — exit codes in per-command view P2

```sh
$TT schema --command validate >out.json 2>err.json
jq -e 'has("exit_codes")' out.json
jq -e '.exit_codes | sort == [0,1,2,3,65] or .exit_codes | length > 0' out.json
```

- Per-command view includes `exit_codes` array.

---

# 17. Stream routing

## TC-STREAM-001 — success envelope on stdout, nothing on stderr P0

```sh
TERM="tzimtzum"
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" lookup "${TERM}" >out.json 2>err.json
echo "exit=$?"
test -s out.json && echo "stdout: non-empty"
test ! -s err.json && echo "stderr: empty"
```

- stdout has the JSON envelope.
- stderr is empty on success.

## TC-STREAM-002 — error envelope on stderr, nothing on stdout P0

```sh
$TT lookup "anything" >out.json 2>err.json
echo "exit=$?"
test ! -s out.json && echo "stdout: empty"
test -s err.json && echo "stderr: non-empty"
```

- stderr has the error envelope.
- stdout is empty on error.

---

# 18. Regression — validate still works P0

## TC-REG-001 — validate command unaffected by E4 changes

```sh
$TT --tbx "${FIXTURES}/canonical/minimal-dct.tbx" validate >out.json 2>err.json
echo "exit=$?"
jq -e '.ok == true' out.json
jq -e '.schema_version == 1' out.json
```

- **exit**: `0`
- Validate command still works after E4 changes.
- Envelope shape unchanged.

---

# Cleanup

```sh
rm -rf "${QA_TMP}"
rm -f out.json err.json
```

---

# Test case summary

| Section                      | Cases  | Priority |
| ---------------------------- | ------ | -------- |
| Lookup — basic matching      | 5      | P0–P1    |
| Lookup — --lang filter       | 2      | P1       |
| Lookup — envelope shape      | 2      | P0–P1    |
| --fields projection          | 3      | P0–P2    |
| Schema — full output         | 5      | P0–P1    |
| Schema — --command filter    | 3      | P0–P2    |
| Extract — capitalized        | 3      | P0–P2    |
| Extract — foreign-script     | 3      | P0–P2    |
| Extract — high-frequency     | 3      | P1–P2    |
| Extract — markdown           | 2      | P0–P1    |
| Extract — --exclude          | 2      | P1–P2    |
| Extract — language detection | 3      | P1–P2    |
| Extract — envelope shape     | 3      | P0–P1    |
| Extract — exit codes         | 2      | P0       |
| --fields on extract          | 1      | P2       |
| Schema — error code detail   | 2      | P1–P2    |
| Stream routing               | 2      | P0       |
| Regression — validate        | 1      | P0       |
| **Total**                    | **46** |          |
