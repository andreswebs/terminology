# Release notes draft: exit-code taxonomy migration

Breaking change (v0.x). The CLI now follows the family-wide exit-code
taxonomy in [ADR 0001](../adr/0001-exit-code-taxonomy.md). Exit codes are
part of the machine contract, so consumers that branch on them must be
updated.

## Changed exit codes

| Old | New | Class                                                    |
| --- | --- | -------------------------------------------------------- |
| 2   | 64  | Usage error (`EX_USAGE`)                                 |
| 2   | 65  | Oversize input (`input_too_large`; data rejected)        |
| 3   | 74  | I/O error (`EX_IOERR`, `io_error`)                       |
| 3   | 75  | Lock contention (`EX_TEMPFAIL`, `tbx_locked`; retryable) |
| 1   | 70  | Unclassified internal error fallback (`EX_SOFTWARE`)     |

Unchanged: `0` (success), `1` (recoverable result: violations, not-found,
partial, apply per-concept validation failure), and `65` (data/validation
errors already emitted by `validate`, `apply`, and write-command input
rejection).

## Migration guidance for consumers

- Treat any code `>= 64` as a hard failure; `0` and `1` are results.
- Scripts previously testing `== 2` for usage errors should test `== 64`.
- Scripts previously testing `== 3` for I/O should test `== 74`, and may
  additionally handle `75` (lock held by a concurrent writer) with a retry.
