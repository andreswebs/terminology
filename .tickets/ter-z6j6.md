---
id: ter-z6j6
status: closed
deps: [ter-fswn]
links: []
created: 2026-05-27T15:17:15Z
type: task
priority: 1
assignee: Andre Silva
parent: ter-nd3x
tags: [e9, hardening, security]
---
# E9.T5 — Path sandbox (CWD enforcement, --tbx exemption)

Enforce CWD subtree sandbox for output paths and --file. The --tbx flag is exempt from CWD sandbox but still sanitized (reject .., percent-encoding). Symlink resolution with reapplied prefix check.

## Acceptance Criteria

- sanitizePath (from E9.T1) implements: filepath.Clean → filepath.Abs → assert under baseDir → filepath.EvalSymlinks → reapply prefix check
- --file and output paths sandboxed to CWD subtree
- --tbx exempt from CWD sandbox: path is sanitized (reject .. segments, percent-encoded segments) but the resulting absolute path is NOT pinned to CWD
- Symlinks that escape the sandbox are rejected (resolve first, then re-check prefix)
- invalid_path error (exit 65) on sandbox violation
- Unit tests: path within CWD (ok), path outside CWD (rejected), symlink escape (rejected), --tbx absolute path (ok), --tbx with .. (rejected)
- Golden tests for path sandbox violations
- make build passes


## Notes

**2026-05-27T16:58:51Z**

E9.T5 completed. The core sanitization implementation (sanitizePath with CWD sandbox, sanitizeTBXPath without CWD sandbox, resolveAndSandbox with symlink resolution) was already in place from E9.T1/T2. This ticket added the missing test coverage per the acceptance criteria:

1. Added TestSanitizeTBXPath unit test (9 cases): absolute path ok, relative path ok, home dir ok, dotdot rejected, dotdot middle rejected, percent encoded rejected, query param rejected, hash rejected, control char rejected.

2. Added 3 golden CLI tests for absolute-path sandbox violations (distinct from existing path_traversal tests which use .. and are caught by string-level checks): scan_absolute_outside, check_absolute_outside, extract_absolute_outside — all exit 65 with invalid_path code. These exercise the resolveAndSandbox prefix check path specifically.

All existing unit tests (TestSanitizePath, TestSanitizePath_SymlinkEscape, TestSanitizePath_ReturnsAbsolute) already covered the remaining acceptance criteria items. make build passes clean.
