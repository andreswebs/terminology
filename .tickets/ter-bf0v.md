---
id: ter-bf0v
status: closed
deps: [ter-uqyn, ter-m9l4]
links: []
created: 2026-05-22T19:19:19Z
type: epic
priority: 2
assignee: Andre Silva
tags: [epic, read]
---
# E4 — Read commands: lookup, schema, extract

Spec: docs/specs/004-read-commands.md

Three read commands plus the shared envelope/markdown machinery consumed across all reads and writes.

- lookup TERM — case-fold + NFC exact surface form against domain model. --lang restricts. Not-found → results:[], exit 1.
- schema — REFLECTIVE introspection over live urfave tree + internal/output types + terr sentinel registry. No embedded JSON file. --command NAME filters.
- extract — heuristic engine: capitalized phrases + foreign-script tokens + frequency. --exclude, --script, --lang, --stopwords, --min-freq. No bundled stoplists in v1.

Shared infra introduced:
- internal/output — JSON/text envelope helpers, --fields projection with reflective path validation against json tags
- internal/markdown — goldmark-backed plain-text Spans iterator with line/col preservation; reused by E5/E6

New dependency: github.com/yuin/goldmark. New sentinel: ErrInvalidField (exit 2 usage error). Language detection precedence: frontmatter lang: → --lang → default en (extract only).

