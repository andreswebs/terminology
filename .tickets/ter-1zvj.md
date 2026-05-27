---
id: ter-1zvj
status: closed
deps: [ter-qlw3, ter-anta, ter-x40w, ter-2398, ter-6z5g]
links: []
created: 2026-05-24T00:29:52Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, io, lock, foundation]
---
# E2.T13 — Save + atomic write + advisory lock

## Goal

Implement \`tbx.Save(path, glossary)\` with the atomic write contract and cross-process advisory locking. Save acquires a lock, writes to a temp file, fsyncs, renames over the target, and releases the lock. If another process holds the lock, Save returns ErrTBXLocked immediately (non-blocking).

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Atomic write contract", §"Cross-process write safety"
- Error-handling ADR: [docs/adr/error-handling.md](docs/adr/error-handling.md) — ErrTBXLocked is a terr sentinel

## Files to create/modify

- \`src/internal/tbx/io.go\` — Save function
- \`src/internal/tbx/lock.go\` — advisory lock acquisition
- \`src/internal/tbx/io_test.go\` — additional tests

## Dependencies

- \`github.com/rogpeppe/go-internal/lockedfile\` — add to go.mod

## API

```go
// In io.go
func Save(path string, g *Glossary) error

// In lock.go
func acquireLock(lockPath string) (func(), error)
```

Design notes:

### Atomic write contract
1. Acquire advisory write lock on \`\${TARGET}.lock\`
2. Write new content to \`\${TARGET}.tmp-XXXX\` in same directory
3. \`fsync\` the temp file
4. \`os.Rename\` over target (atomic on POSIX, atomic-enough on Windows)
5. Release lock
6. On any error before rename, remove temp file

### Advisory lock
- **Lock target**: sibling \`.lock\` file (\`\${TBX}.lock\`), not the TBX file itself. Separate file so the lock fd's lifecycle is independent of the file replacement done by rename.
- **Non-blocking**: If another process holds the lock, return \`ErrTBXLocked\` immediately with hint "another terminology process is writing; retry". No hanging, no retry loop. Agents see deterministic failure.
- **Implementation**: \`rogpeppe/go-internal/lockedfile\` — used by Go toolchain itself. Handles POSIX flock and Windows LockFileEx.
- **Lock file cleanup**: removed after successful Save. The file's presence between runs is benign (it's a marker, not state).
- **Scope**: lock held across read → modify → write → rename. Released via defer.

### lock.go

```go
package tbx

import (
    "os"
    "github.com/rogpeppe/go-internal/lockedfile"
)

func acquireLock(lockPath string) (func(), error) {
    f, err := lockedfile.Create(lockPath)
    if err != nil {
        return nil, ErrTBXLocked.Wrap(err)
    }
    return func() {
        _ = f.Close()
        _ = os.Remove(lockPath)
    }, nil
}
```

Design notes for lock.go:
- **ErrTBXLocked.Wrap(err)** preserves the underlying OS error while maintaining the typed code. Callers can check \`errors.Is(err, ErrTBXLocked)\` and also unwrap to see the OS-level detail.
- **acquireLock is unexported** — only Save calls it.
- **Separate file from io.go** per the spec's architecture diagram. This keeps I/O concerns (file open/close/rename) separate from locking concerns.

## Deviation note

The current implementation puts acquireLock inside io.go. The spec's architecture diagram shows \`lock.go\` as a separate file. This ticket creates the separate \`lock.go\` file as the spec intends. Additionally, the current implementation uses \`lockedfile.Create\` which creates the file with exclusive access — this may not be truly non-blocking on all platforms. The spec says "non-blocking; if another process holds the lock, return ErrTBXLocked". Verify that \`lockedfile.Create\` behaves non-blockingly or document the platform limitation.

## TDD cycles

### Cycle 1 — Save produces valid file
RED: Create a Glossary, Save to a temp dir. Read the output file. Assert it contains valid TBX XML.
GREEN: Implement Save with temp file → fsync → rename.

### Cycle 2 — Temp file cleanup on success
RED: After successful Save, glob for \`.terminology-*.tmp\` files in the output dir. Assert none found.
GREEN: Ensure rename removes the temp file reference; defer handles cleanup path.

### Cycle 3 — Atomic: original untouched on write error
RED: Save with a Writer that returns an error mid-write. Assert the target file either doesn't exist or retains its original content.
GREEN: Temp file is removed in the defer; rename never happens.

### Cycle 4 — Lock file cleanup
RED: After successful Save, assert \`\${path}.lock\` does not exist.
GREEN: acquireLock's cleanup function removes the lock file.

### Cycle 5 — Round-trip: Load → Save → Load
RED: Load a fixture, Save to a new path, Load the new path. Assert glossaries are equivalent.
GREEN: Integration of reader + writer + Save.

### Cycle 6 — acquireLock returns cleanup function
RED: Call acquireLock, assert the lock file exists, call cleanup, assert it's removed.
GREEN: Implement acquireLock.

### Cycle 7 — acquireLock error wrapping
RED: Simulate lock failure (e.g. directory doesn't exist). Assert errors.Is(err, ErrTBXLocked).
GREEN: Wrap with ErrTBXLocked.Wrap(err).

## Out of scope

- Concurrent lock contention testing (E9 hardening)
- Reader implementation (T6-T8)
- Writer implementation (T10)

## Acceptance

- \`make build\` passes
- Save produces valid TBX file via atomic write (temp → fsync → rename)
- No temp files left behind on success or failure
- Lock file created during write, cleaned up after
- ErrTBXLocked returned when lock cannot be acquired
- \`go.mod\` includes rogpeppe/go-internal dependency


## Notes

**2026-05-25T13:46:14Z**

Extracted acquireLock from io.go into lock.go as spec intended. Added missing test coverage: (1) lock_test.go with internal tests for acquireLock create/cleanup and error wrapping, (2) io_test.go with TestSave_OriginalPreservedOnWriteError (cycle 3 — original file untouched when writer fails) and TestSave_ErrTBXLocked_NonExistentDir (cycle 7 — terr.Coded check). Key learning: terr.E.Wrap creates a copy so errors.Is against the original sentinel fails; use terr.Coded type assertion + Code() check instead.
