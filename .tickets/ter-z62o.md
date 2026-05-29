---
id: ter-z62o
status: closed
deps: []
links: []
created: 2026-05-29T01:13:40Z
type: feature
priority: 1
assignee: Andre Silva
parent: ter-8gyy
tags: [beta-feedback, tbx, header]
---
# BETA — preserve TBX header on round-trip (titleStmt + root xml:lang)

Beta feedback: the canonical writer regenerates the header from scratch every write (src/internal/tbx/linguist/writer.go:18-48), dropping <titleStmt> and hard-coding root xml:lang="en" (line 27) — surprising for an es/he glossary with no English. The model (src/internal/tbx/model.go:3-8) only retains SourceDesc. Decision (Andre): capture the FULL tbxHeader/fileDesc set into the model and preserve root xml:lang from read. Re-emit canonically (ordered) to satisfy docs/adr/determinism.md — preserve content, not byte layout.

## Design

Extend Glossary with a Header struct capturing tbxHeader/fileDesc children: titleStmt, publicationStmt, sourceDesc, encodingDesc, revisionDesc; plus SourceLang (root <tbx> xml:lang). Reader (src/internal/tbx/linguist/reader.go) populates these, replacing the lone sourceDesc capture. Writer emits xml:lang from SourceLang (remove hard-coded "en" at writer.go:27) and serializes captured header fields in a fixed canonical order. Keep emitting a sensible default header when fields are empty (e.g. fresh glossaries).

## Acceptance Criteria

make build passes
Glossary model captures full tbxHeader/fileDesc set + SourceLang
Reader populates header fields and root xml:lang
Writer emits root xml:lang from SourceLang (no hard-coded en) and re-serializes header fields canonically
Round-trip test: es/he fixture read->write preserves titleStmt and root xml:lang
Output remains deterministic (determinism ADR tests pass)
Existing TBX golden/round-trip tests updated and passing


## Notes

**2026-05-29T01:24:03Z**

Added Glossary.SourceLang and Glossary.Header (Title, PublicationStmts, SourceDescs, EncodingDescs, RevisionDescs). Reader walks tbxHeader/fileDesc and populates these; root xml:lang captured into SourceLang. Writer emits xml:lang from SourceLang (default "en") and serializes header sections in canonical order (titleStmt, publicationStmt, sourceDesc, encodingDesc, revisionDesc) - sections omitted when empty. Glossary.SourceDesc kept for back-compat; populated from first Header.SourceDescs paragraph on read. New canonical fixture testdata/canonical/header-es-he.tbx exercises titleStmt + xml:lang=es; round-trips byte-exact via TestRoundTrip_Canonical.
