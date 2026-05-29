---
name: terminology
compatibility: Requires the `terminology` CLI to be installed and available on PATH.
description: |
  Go CLI for agent-driven academic translation with terminology enforcement against a TBX-Linguist glossary. Use when translating academic text with a glossary, validating TBX files, scanning markdown for term matches, checking term consistency between source and target, adding/updating/removing glossary concepts, or bulk-reconciling a glossary. Outputs JSON by default with meaningful exit codes. Run `terminology schema --command CMD` to discover any command's flags, envelope shape, and error codes.
---

# terminology

CLI for agent-driven, terminology-focused academic translation. Reads markdown source, enforces consistent terminology against a TBX-Linguist glossary, and exposes deterministic operations as subcommands.

## Global flags

| Flag                                | Short | Env var           | Purpose                                          |
| ----------------------------------- | ----- | ----------------- | ------------------------------------------------ |
| `--tbx PATH`                        | `-T`  | `TERMINOLOGY_TBX` | Path to TBX glossary (required by most commands) |
| `--format json\|text`               |       |                   | Output format (default: `json`)                  |
| `--verbose` / `--debug` / `--quiet` |       |                   | Mutually exclusive verbosity levels              |

## Output contract

All commands produce JSON on stdout for success and a JSON error envelope on stderr:

```json
{"schema_version":1, "ok":true, ...}
```

```json
{
  "schema_version": 1,
  "ok": false,
  "error": { "code": "...", "message": "...", "hint": "..." }
}
```

- `hint` is omitted (not empty string) when absent
- `--format text` renders: `✗ <message>` + optional `  hint: <hint>` on next line

### Stream routing

| Scenario            | stdout                            | stderr                                |
| ------------------- | --------------------------------- | ------------------------------------- |
| Success             | result envelope                   | empty                                 |
| Warnings (validate) | result envelope with `warnings[]` | empty                                 |
| Violations (check)  | full results envelope             | short summary (`"code":"violations"`) |
| Error               | empty                             | error envelope                        |

Capture pattern: `terminology ... 2>err.json` — stdout has results, stderr has errors.

## Exit codes

| Code | Meaning                                                      | When                                                                                                               |
| ---- | ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------ |
| 0    | Success                                                      | All commands                                                                                                       |
| 1    | Warnings / no results / violations / apply validation failed | validate (warnings), lookup (not found), extract (no candidates), check (violations), apply (per-concept failures) |
| 2    | Usage error                                                  | All commands (bad flags, missing args, mutex violations)                                                           |
| 3    | I/O error                                                    | scan, check, extract, apply (file not found, locked)                                                               |
| 65   | Validation / data error                                      | validate (bad TBX, strict errors), write commands, apply (invalid_input, dangling_crossref)                        |

**Exit 1 with `ok:true`** means "no results" (not an error). **Exit 1 with `ok:false`** means violations or validation failures.

## Commands

### init

Create a minimal valid TBX-Linguist skeleton at the target path.

```sh
terminology init --tbx glossary.tbx --source-lang en --title "Project glossary"
terminology init --tbx glossary.tbx --source-lang en --dry-run   # render to stdout
```

- Required: `--source-lang` (BCP 47 tag, becomes the root `xml:lang`)
- Refuses to overwrite an existing file (exit 3, `io_error`)
- `--dry-run` renders the skeleton without writing

The skeleton is the canonical empty glossary — bootstrap with this before
calling any write command. Shape:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader>
    <fileDesc>
      <titleStmt><title>Project glossary</title></titleStmt>
      <sourceDesc><p>Terminology glossary</p></sourceDesc>
    </fileDesc>
  </tbxHeader>
  <text>
    <body>
    </body>
  </text>
</tbx>
```

`<titleStmt>` is omitted when `--title` is not supplied.

### validate

Validate a TBX-Linguist glossary.

```sh
terminology validate --tbx glossary.tbx
terminology validate --tbx glossary.tbx --strict   # warnings promoted to errors (exit 65)
```

- Three tiers: well-formedness → dialect/schema → semantic
- Supported dialects: TBX-Linguist DCT and DCA; legacy forms normalize silently
- Unsupported dialects (TBX-Basic) → exit 65

### lookup

Look up a term in the glossary.

```sh
terminology lookup "algorithm" --tbx glossary.tbx
terminology lookup "algorithm" --tbx glossary.tbx --lang en --fields results.concept_id
```

- Case-insensitive + NFC normalization matching
- `--lang` restricts which language sections are searched
- Not found → exit 1, `ok:true`, `results:[]`

### schema

Self-documenting command reference. No `--tbx` required.

```sh
terminology schema                              # list all commands
terminology schema --command lookup             # single-command detail
```

Returns flags, envelope structure, exit codes, and error codes. Use this to discover valid `--fields` paths for any command. Unknown `--command` → exit 2.

### extract

Extract terminology candidates from a markdown file.

```sh
terminology extract document.md --tbx glossary.tbx
terminology extract document.md --script hebrew --min-freq 2
```

- Three heuristics: `capitalized_phrase`, `foreign_script`, `high_frequency`
- `--exclude` removes terms already in glossary
- `--script` filters by script family (`latin`, `hebrew`, `cyrillic`, `arabic`, `any`)
- `--min-freq N` (default 3) gates frequency candidates
- `--lang` fallback: frontmatter `lang:` → `--lang` flag → default `en`
- Code blocks (fenced + inline) excluded from extraction
- No candidates → exit 1, `ok:true`, `candidates:[]`

### scan

Scan a document for glossary term matches.

```sh
terminology scan document.md --tbx glossary.tbx
terminology scan document.md --tbx glossary.tbx --lang en --context 120
```

- Matches sorted by `(line, column)` — deterministic
- `--context N` (default 80): N/2 chars per side of match
- `--lang` restricts compiled patterns; frontmatter `lang:` overrides flag
- Always exits 0 on success (informational, even with zero matches)
- Exit 3 for I/O errors

Matcher: case-insensitive, Hebrew niqqud stripping, diacritics strict (accented chars must match exactly — `razón` ≠ `razon`), multi-word terms match across whitespace.

### check

Check terminology consistency between source and target files.

```sh
terminology check source.md target.md --tbx glossary.tbx --source-lang en --target-lang he
```

- Language resolution: frontmatter `lang:` → CLI flags → `language_required` error (exit 2)
- Violation types: `missing` (preferred target absent), `forbidden_variant` (deprecated/superseded used), `admitted_variant` (admitted term used)
- `--strict` promotes `admitted_variant` from warning to violation
- Only checks concepts found in source file
- Exit 0 = clean, exit 1 = violations (`ok:false`)

### concept add / update / remove

Manage glossary concepts.

```sh
# Add via flags
terminology concept add --tbx glossary.tbx --lang en --term "algorithm" --subject-field "CS"

# Add via JSON stdin (preferred + admitted + deprecated + superseded variants)
echo '{
  "concept_id": "algorithm",
  "languages": {
    "en": {
      "preferred":  { "term": "algorithm",  "administrative_status": "preferredTerm-admn-sts" },
      "admitted":   [{ "term": "algo",      "administrative_status": "admittedTerm-admn-sts" }],
      "deprecated": [{ "term": "recipe",    "administrative_status": "deprecatedTerm-admn-sts" }],
      "superseded": [{ "term": "procedure", "administrative_status": "supersededTerm-admn-sts" }]
    }
  }
}' | terminology concept add --tbx glossary.tbx

# Update (exactly one of --merge or --replace required)
terminology concept update algorithm --merge --tbx glossary.tbx --lang he --term "אלגוריתם"

# Remove (refuses if cross-referenced, unless --force)
terminology concept remove algorithm --tbx glossary.tbx
```

- Concept ID derived from preferred term: NFKD → casefold → keep [a-z0-9] → hyphen-join → truncate 64
- `--id` overrides derivation; `--canonical-lang` selects derivation language
- `--merge`: adds new langSecs, preserves absent ones, overlays existing terms
- `--replace`: replaces entire concept except concept ID; omitted langSecs dropped
- `--force` on remove overrides dangling crossref protection
- Three input modes: flags, JSON stdin, TBX fragment stdin (auto-detected)
- Full TBX document as input → rejected (exit 65, `invalid_input`)
- Concept ID **never changes** after creation

### term add / deprecate

Add or deprecate a term within an existing concept.

```sh
terminology term add concept-id --tbx glossary.tbx --lang en --term "variant" \
  --status admittedTerm-admn-sts
terminology term deprecate concept-id --tbx glossary.tbx --lang en --term "old-term"
```

### apply

Declarative bulk reconciliation against a payload.

```sh
terminology apply --tbx glossary.tbx --file payload.json
terminology apply --tbx glossary.tbx --file payload.json --dry-run --prune
echo '{"concepts":[...]}' | terminology apply --tbx glossary.tbx --file -
```

- Computes minimal add/update/remove (all-or-nothing, idempotent)
- `--prune`: removes concepts absent from payload
- `--file -`: read stdin; format auto-detected (`.json` → JSON, `.tbx`/`.xml` → TBX)
- Per-concept validation failures → exit 1, `apply_validation_failed` with `failures[]`
- Concept equality: canonicalized XML byte-identical after stripping `<transacGrp>`
- `--transaction --author NAME`: adds transacGrp on added/updated concepts only

## Shared write flags

Apply to concept add/update/remove, term add/deprecate, and apply:

| Flag                   | Env var              | Purpose                                         |
| ---------------------- | -------------------- | ----------------------------------------------- |
| `--dry-run` / `-n`     |                      | Show result without modifying file              |
| `--transaction`        |                      | Add `<transacGrp>` records to modified concepts |
| `--author NAME` / `-a` | `TERMINOLOGY_AUTHOR` | Author for transaction responsibility           |

- `--transaction` without `--author`: omits responsibility, WARN on stderr
- `--author` without `--transaction`: silently ignored
- Pre-write validation always runs — file is never written if validation fails

## Agent patterns

### Discover command details

```sh
terminology schema --command lookup
```

Authoritative source for valid flags, envelope structure, and error codes.

### Safe writes with dry-run

```sh
terminology concept add --tbx glossary.tbx --dry-run --lang en --term "test"
```

Result envelope shows what would happen. File remains untouched (checksum unchanged).

### Capture errors separately

```sh
terminology check source.md target.md --tbx glossary.tbx \
  --source-lang en --target-lang he 2>err.json
```

stdout has the results envelope. stderr has errors or violation summary.

### Mutation testing

```sh
WORK=$(cp glossary.tbx "$(mktemp)" && echo "$(mktemp)")
# or: WORK=$(mktemp); cp glossary.tbx "$WORK"
terminology concept add --tbx "$WORK" --lang en --term "test"
```

Always work on copies when testing writes.

### Field projection

```sh
terminology lookup "term" --tbx glossary.tbx --fields results.concept_id,results.languages.en
```

Paths are envelope-relative. Invalid path → exit 2, `invalid_field`. Use `terminology schema --command CMD` to discover valid paths.

## Gotchas

- **Exit 1 is not always an error.** For lookup/extract, exit 1 + `ok:true` = "no results." For check, exit 1 + `ok:false` = violations. For apply, exit 1
  - `ok:false` = per-concept validation failures (recoverable, with detail).
- **Diacritics are strict.** `razón` ≠ `razon` in scan matching. Accented characters must match exactly.
- **Frontmatter overrides flags.** In scan/check/extract, markdown frontmatter `lang:` takes precedence over `--lang`.
- **check only checks concepts in source.** A concept absent from the source won't generate violations in the target, even if its terms are misused.
- **Concept ID never changes.** Updating the preferred term does not alter the ID.
- **Full TBX documents rejected on write.** Only `<conceptEntry>` fragments or `<conceptEntryList>` wrappers accepted as stdin.
- **check streams differ from other commands.** Violations produce full results on stdout AND a short summary on stderr.
- **apply update is wholesale replace.** Payload is authoritative — omitted fields and langSecs are dropped, not preserved.
- **Advisory file locking.** Writes acquire `${TBX}.lock` (fcntl). Non-blocking, fails fast with `tbx_locked` (exit 3).

## Reference files

Load `references/error-codes.md` when you need the complete error code catalog with exit codes, descriptions, and trigger conditions for every command.

Load `references/write-details.md` when performing write operations and you need detailed semantics: input layering (flags/JSON/TBX fragment), merge vs replace behavior, transaction record structure, concept ID derivation rules, apply patch model, or JSON/TBX payload formats.
