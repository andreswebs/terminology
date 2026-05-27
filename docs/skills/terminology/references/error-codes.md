# Error codes

Complete error code catalog for all terminology commands.

## Error code table

| Code                      | Exit | Description                                                                | Commands                                                 |
| ------------------------- | ---- | -------------------------------------------------------------------------- | -------------------------------------------------------- |
| `no_tbx_path`             | 2    | Missing `--tbx` / `TERMINOLOGY_TBX`                                        | All commands needing a glossary                          |
| `no_subcommand`           | 2    | Bare `terminology` invocation                                              | Root command                                             |
| `unknown_subcommand`      | 2    | Unknown subcommand name                                                    | Root command                                             |
| `missing_argument`        | 2    | Missing positional argument                                                | Commands with positional args                            |
| `excess_arguments`        | 2    | Too many positional arguments                                              | All commands                                             |
| `missing_required_flag`   | 2    | Required flag not provided                                                 | e.g. `--file` for apply                                  |
| `invalid_value`           | 2    | Invalid enum value                                                         | All commands with enum flags                             |
| `invalid_field`           | 2    | Bad `--fields` path. Hint: use `terminology schema`                        | Read commands                                            |
| `conflicting_verbosity`   | 2    | Multiple verbosity flags (`--verbose` + `--debug`, etc.)                   | All commands                                             |
| `merge_replace_mutex`     | 2    | Both `--merge` and `--replace` provided                                    | concept update                                           |
| `merge_replace_required`  | 2    | Neither `--merge` nor `--replace` provided                                 | concept update                                           |
| `language_required`       | 2    | Missing source/target language (frontmatter + flags both absent)           | check                                                    |
| `unknown_command`         | 2    | `schema --command` with unknown name                                       | schema                                                   |
| `validation_error`        | 65   | TBX validation failure (bad dialect, malformed XML, strict errors)         | validate                                                 |
| `duplicate_id`            | 65   | concept add with an already-existing concept ID                            | concept add                                              |
| `not_found`               | 65   | update/remove/deprecate for nonexistent concept or term                    | concept update, concept remove, term add, term deprecate |
| `dangling_crossref`       | 65   | remove or prune would break an inbound cross-reference (without `--force`) | concept remove, apply --prune                            |
| `invalid_id`              | 65   | Concept ID derivation produces empty result (e.g. Hebrew-only input)       | concept add                                              |
| `invalid_input`           | 65   | Malformed JSON, unknown fields, or full TBX document passed as stdin input | concept add, concept update, term add, apply             |
| `apply_validation_failed` | 1    | Per-concept validation failures during apply (with `failures[]` detail)    | apply                                                    |
| `violations`              | 1    | Violations summary (emitted on stderr)                                     | check                                                    |
| `io_error`                | 3    | File I/O failure (file not found, permission denied)                       | scan, check, extract, apply                              |
| `tbx_locked`              | 3    | Advisory file lock held by another process                                 | All write commands                                       |
| `under_construction`      | 75   | Unimplemented command (legacy stub)                                        | N/A                                                      |

## Exit code summary by command

### validate

| Exit | Condition                                                              |
| ---- | ---------------------------------------------------------------------- |
| 0    | Valid TBX, no warnings                                                 |
| 1    | Valid TBX, warnings present (`ok:true`, `warnings[]` populated)        |
| 2    | Usage error (`no_tbx_path`, conflicting flags)                         |
| 65   | Validation failure (unsupported dialect, malformed XML, strict errors) |

### lookup

| Exit | Condition                                    |
| ---- | -------------------------------------------- |
| 0    | Term found (`results[]` populated)           |
| 1    | Term not found (`ok:true`, `results:[]`)     |
| 2    | Usage error (`no_tbx_path`, `invalid_field`) |

### schema

| Exit | Condition         |
| ---- | ----------------- |
| 0    | Schema emitted    |
| 2    | `unknown_command` |

### extract

| Exit | Condition                                  |
| ---- | ------------------------------------------ |
| 0    | Candidates found                           |
| 1    | No candidates (`ok:true`, `candidates:[]`) |
| 2    | Usage error                                |
| 3    | I/O error (file not found)                 |

### scan

| Exit | Condition                                                      |
| ---- | -------------------------------------------------------------- |
| 0    | Scan complete (always, even with zero matches — informational) |
| 2    | Usage error                                                    |
| 3    | I/O error (file not found)                                     |

### check

| Exit | Condition                                                         |
| ---- | ----------------------------------------------------------------- |
| 0    | Clean (no violations)                                             |
| 1    | Violations found (`ok:false`)                                     |
| 2    | Usage error (`no_tbx_path`, `language_required`, `invalid_field`) |
| 3    | I/O error                                                         |

### Write commands (concept add/update/remove, term add/deprecate)

| Exit | Condition                                                                                    |
| ---- | -------------------------------------------------------------------------------------------- |
| 0    | Write succeeded (or dry-run success)                                                         |
| 2    | Usage error (`no_tbx_path`, `merge_replace_*`, `invalid_value`)                              |
| 3    | I/O error (`tbx_locked`)                                                                     |
| 65   | Data error (`duplicate_id`, `not_found`, `dangling_crossref`, `invalid_id`, `invalid_input`) |

### apply

| Exit | Condition                                                                   |
| ---- | --------------------------------------------------------------------------- |
| 0    | Reconciliation applied (or dry-run)                                         |
| 1    | Per-concept validation failed (`apply_validation_failed` with `failures[]`) |
| 2    | Usage error (`no_tbx_path`, missing `--file`)                               |
| 3    | I/O error (file not found, `tbx_locked`)                                    |
| 65   | Data error (`invalid_input`, `dangling_crossref`)                           |

## Warning codes (validate)

Seven warning codes, produced by validation tiers 2 and 3:

| Code                     | Tier | Description                                                            |
| ------------------------ | ---- | ---------------------------------------------------------------------- |
| `duplicate_id`           | 3    | Duplicate concept entry ID                                             |
| `invalid_lang_tag`       | 3    | Invalid BCP 47 language tag                                            |
| `invalid_picklist`       | 2    | Invalid closed-enum value                                              |
| `legacy_form_normalized` | 2    | Bare form normalized silently (suppressed unless `--strict`)           |
| `missing_term`           | 3    | langSec with no term entries                                           |
| `unknown_element`        | 2    | Unrecognized element (warning in default, promoted in strict)          |
| `unresolved_crossref`    | 3    | Cross-reference target not found (warning in default, error in strict) |

Warning shape:

```json
{
  "code": "...",
  "message": "...",
  "concept_id": "optional",
  "line": 0,
  "column": 0
}
```

`line` and `column` are populated (> 0) for reader-level warnings.

## apply_validation_failed detail

When apply encounters per-concept validation failures:

```json
{
  "schema_version": 1,
  "ok": false,
  "error": {
    "code": "apply_validation_failed",
    "message": "...",
    "details": {
      "failures": [{ "concept_id": "...", "code": "...", "message": "..." }]
    }
  }
}
```

This is exit 1 (not 65) — recoverable with per-concept detail. The glossary file is untouched when this error occurs.
