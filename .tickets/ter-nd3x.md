---
id: ter-nd3x
status: closed
deps: [ter-uqyn, ter-8gyy, ter-uj1y]
links: []
created: 2026-05-22T19:19:20Z
type: epic
priority: 3
assignee: Andre Silva
tags: [epic, hardening, security]
---
# E9 — Hardening: input validation, fuzz, perf, security

Spec: docs/specs/009-hardening.md

Cross-cutting safety + regression infrastructure. Becomes its own epic once the command surface exists. Some rules referenced inline by E2/E3/E5/E7; E9 is the omnibus.

internal/app/commands/sanitize.go — unexported boundary validators called at the urfave action handler (single point of trust per request):
- sanitizeConceptID, sanitizeLangTag, sanitizePath(s, baseDir) → cleaned absolute path within baseDir, sanitizeTerm
- Sentinel codes: invalid_id, invalid_lang_tag, invalid_path, invalid_term

XML hardening:
- DOCTYPE policy: accept bare <!DOCTYPE tbx>; reject internal subsets and SYSTEM/PUBLIC external IDs (streaming pre-scan)
- xml.Decoder.Strict=true + 256-level nesting cap; above cap → nesting_too_deep

Bounded reads (io.LimitedReader): TBX 50 MB, markdown/stdin/extract-per-file 10 MB. ErrInputTooLarge sentinel.

Path sandbox: Clean → Abs → assert under CWD → resolve symlinks → reapply prefix check. --tbx EXEMPT from CWD sandbox (still sanitized).

Fuzz: linguist.Reader.Decode, Matcher.Scan, write.DeriveID; -fuzztime=30s nightly, Linux-only, corpus under testdata/fuzz/.

Perf budget: hardcoded constants in internal/<pkg>/perf_test.go behind 'perf' build tag; CI fails on exceedance.

CI assertions: scenarios.md parses via urfave argv parser; docs/reference/errors.md regenerated from terr registry (diff fails CI); help-text health (non-empty Usage/Description) in make lint.

