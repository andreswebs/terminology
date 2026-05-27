---
id: ter-uqyn
status: closed
deps: [ter-qxrg, ter-v203]
links: []
created: 2026-05-22T19:19:19Z
type: epic
priority: 1
assignee: Andre Silva
tags: [epic, foundation, tbx]
---
# E2 — Domain model & TBX I/O

Spec: docs/specs/002-domain-and-io.md

Dialect-agnostic internal domain types (Glossary, Concept, LangSection, Term, Transaction, NoteText) plus the first concrete reader/writer pair for TBX-Linguist.

- internal/tbx — model, Reader/Writer interfaces dispatched on <tbx>@type, atomic write helper, advisory lock wrapper
- internal/tbx/linguist — DCT+DCA reader, canonical-DCT hand-rolled deterministic emitter, legacy-form normalization
- internal/tbx/lineindex — streaming newline-index for line/col tracking
- Atomic write: temp file + fsync + os.Rename; sibling ${TBX}.lock via rogpeppe/go-internal/lockedfile (non-blocking, returns ErrTBXLocked)
- Hand-rolled XML emitter (~200 LOC) — encoding/xml unsuitable for canonical output

New sentinels: ErrTBXLocked, ErrUnsupportedDialect. Read commands do NOT acquire the lock.

Acceptance: round-trip property (read→write→read produces equivalent model); canonical output is byte-stable across runs; concurrent writes fail fast with tbx_locked.


## Notes

**2026-05-22T21:49:06Z**

Implemented E2 — Domain model & TBX I/O:

- internal/tbx/model.go: Glossary, Concept, LangSection, Term, Status, CrossRef, Transaction, NoteText types
- internal/tbx/style.go: Style enum (DCT/DCA), Dialect type with DialectLinguist constant
- internal/tbx/warning.go: Warning type for reader diagnostics
- internal/tbx/reader.go, writer.go: Reader/Writer interfaces
- internal/tbx/errors.go: ErrUnsupportedDialect (exit 65), ErrTBXLocked (exit 3), ErrNoTBXPath (exit 2)
- internal/tbx/io.go: Load/Save with dialect dispatch via registry, atomic write (temp+fsync+rename), advisory lock via lockedfile
- internal/tbx/linguist/reader.go: DCT+DCA decoder with full data category support
- internal/tbx/linguist/writer.go: Canonical DCT hand-rolled emitter (~250 LOC), deterministic output
- internal/tbx/linguist/normalize.go: Legacy form normalization (bare status forms, usageRegister)
- internal/tbx/linguist/register.go: Init-time dialect registration to break import cycle

Tests: round-trip property (4 canonical fixtures), DCA→DCT normalization, legacy forms, Load/Save/atomic write, normalize unit tests.

Acceptance criteria met: round-trip read→write→read produces identical bytes; canonical output is byte-stable; concurrent writes use advisory lock.
