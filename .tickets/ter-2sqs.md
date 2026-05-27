---
id: ter-2sqs
status: closed
deps: [ter-uqyn, ter-bf0v, ter-cfeu]
links: []
created: 2026-05-22T19:19:19Z
type: epic
priority: 2
assignee: Andre Silva
tags: [epic, matcher]
---
# E5 — Matcher

Spec: docs/specs/005-matcher.md

Algorithmic core for scan/check. Multi-pattern matching via Aho-Corasick (cloudflare/ahocorasick) over canonical pre-normalized corpus. Supersedes the cli-design.md regexp-per-term sketch.

Pipeline: markdown → plain-text spans (E4) → canonical bytes + offset map (NFC, casefold, niqqud-strip, \s+→single space) → AC scan → boundary check on ORIGINAL text → longest-match-at-same-start filter → []Match.

v1 policy:
- Diacritics STRICT (razón ≠ razon)
- Niqqud stripped (Hebrew)
- No lemmatization, no inflection — glossary declares variants explicitly
- Per-language Policy table (he: StripNiqqud)
- Code regions skipped via internal/markdown (symmetric)
- Compiles patterns for every term variant (preferred/admitted/deprecated/superseded/unspecified); each Match tagged with status

CJK/Thai out of v1 (whitespace boundary contract). New dep: github.com/cloudflare/ahocorasick.

