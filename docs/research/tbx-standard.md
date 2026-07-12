---
type: Explanation
title: "TermBase eXchange (TBX) Standard"
description: "History, conceptual model, modular architecture, governance, criticisms, and continued relevance of the TBX terminology exchange standard."
tags: [diataxis:explanation]
---

# TermBase eXchange (TBX) Standard

The TermBase eXchange (TBX) standard is an international standard for the representation and interchange of terminological data. This document covers its history, its conceptual commitments, its modular architecture, its governance, its limitations, and what its persistence (or fragility) says about the field of terminology management.

## What TBX is

TermBase eXchange, or TBX, is an XML-based open standard for representing concept-oriented terminological data — termbases — in a way that can be moved losslessly between tools and organizations. It is published as ISO 30042 by ISO Technical Committee 37, the body that since the 1950s has been responsible for international standardization of terminology and other language and content resources. The current edition is ISO 30042:2019, often called TBX Version 3 or TBX 3.0, and a successor revision is in early drafting (ISO/AWI 30042) as of 2026.

TBX is not a translation memory format, not a bilingual working file format, and not a presentation format. Its single concern is the representation of _concepts_ and the _terms_ that denote them across one or more languages, together with the metadata terminologists and translators need to manage that vocabulary professionally. This narrowness is deliberate, and it sits inside a larger family of localization standards — TMX for translation memories, XLIFF for bilingual working files, ITS for source-document metadata — each addressing a different slice of the localization problem.

## Why TBX exists at all

The need TBX addresses is older than the standard itself, and older than computing in any meaningful sense. Technical and institutional translation has always had a _terminology problem_: a body of text that is internally consistent in its source vocabulary will produce an internally inconsistent translation unless the translators share a discipline about how each technical term, name, or concept is rendered. Pre-digital workflows handled this with paper glossaries maintained by terminologists. The work of organizations like ISO TC 37, founded in 1947 as the heir to earlier International Federation of National Standardizing Associations work on terminology principles, was for decades primarily about how such glossaries should be structured — what fields they should contain, how concepts should be related to terms, what counted as a valid definition.

When computing entered terminology, the field already had a mature theoretical apparatus. ISO 704 codified the principles of terminology work; ISO 1087 standardized the vocabulary of terminology theory itself. These were not industry deliverables — they were the substrate. The challenge from the 1990s onward was to translate this conceptual machinery into a machine-readable form that could be exchanged between increasingly heterogeneous tools.

The justification for a _standard_ (rather than, say, each tool exporting its own glossary format) is the same justification that produces TMX for translation memories: terminology databases are _assets_. A pharmaceutical company that has spent two decades curating the approved translations of its drug names, indications, and warnings cannot afford to lose access to that work when it changes vendors. A multilingual standards body cannot afford to have its vocabulary rendered illegible by a software upgrade. The format must outlive the tool.

## Where TBX came from

The lineage of TBX is unusually traceable, because each step was itself an ISO or quasi-ISO publication.

**MATER (1986).** ISO 6156, "Magnetic tape exchange format for terminological/lexicographical records," is the earliest direct ancestor. It defined a tape-based exchange format and was, in retrospect, the moment the field accepted that terminology was something machines would need to read.

**MicroMATER (1991).** Alan Melby's adaptation of MATER for personal computers, in the era when PC terminology tools were becoming viable.

**TEI P3, Chapter 13 (1994).** The Text Encoding Initiative's third release devoted a chapter to representing terminological data in SGML. This was the first attempt to express terminology data using a general-purpose markup language rather than a custom record format, and it brought the field into contact with the broader humanities-computing community.

**MARTIF — ISO 12200 (1999).** The Machine-Readable Terminology Interchange Format took TEI P3 Chapter 13 and refined it into a freestanding ISO standard. MARTIF was SGML-based and remained the working format for a small number of specialized tools through the early 2000s. Traces of MARTIF are still visible in TBX: TBX 2.0 documents had `<martif>` as their root element name, and only with TBX 3.0 in 2019 was this finally renamed to `<tbx>`.

**TBX 1.0 (2002).** LISA's OSCAR special interest group reformulated MARTIF in XML and published it as TBX. This is the moment TBX entered the localization industry as a working format. It was released as a freely accessible open standard, a fact that mattered politically because previous ISO terminology standards were available only by purchase.

**TBX 2.0 — ISO 30042:2008.** LISA and ISO TC 37/SC 3 co-published TBX as a formal international standard. TBX-Basic, a simplified profile aimed at practical translation work rather than full-scale terminological standardization, appeared in the same year, also published by LISA's Terminology Special Interest Group.

**TBX-Min (2013).** LTAC Global, a successor to parts of the dissolved LISA, published the simplest of the dialects — TBX-Min — targeted explicitly at the use case of sending a small bilingual glossary to a translator alongside a job.

**TBX 3.0 — ISO 30042:2019.** The current edition. The most consequential structural revision since 2002. It rebuilt TBX as a modular Core plus optional modules, formalized the dialect mechanism, replaced the monolithic DTD with schema-language-neutral specifications, introduced the DCT (Data Category as Tag) style alongside the original DCA (Data Category as Attribute) style, and harmonized the inline data model with XLIFF 2.

## The conceptual model

Underneath the XML, TBX rests on a particular theory of what terminology _is_. This theory comes from ISO TC 37 and is older than computing, but its commitments shape every detail of the format.

### Concept-orientation

The defining choice is that a TBX entry represents a _concept_, not a word. A concept — say, the Lurianic notion of צמצום — has terms attached to it: the Hebrew term `צמצום`, the Spanish `tzimtzum`, the English `tzimtzum`, perhaps the German `Zimzum`. The entry is keyed by the concept; the terms are properties of the entry. This is the opposite of how a bilingual dictionary works, where each entry is keyed by a headword in a source language and lists translations.

The consequence is that TBX naturally accommodates a multilingual termbase. There is no privileged "source" language at the data model level; all language sections within a `<conceptEntry>` are formally equal. A four-language termbase covering Spanish, Hebrew, English, and German is structurally identical to a two-language termbase with one extra `<langSec>` removed.

This is also the most criticized commitment of TBX, and the criticism is worth taking seriously. Concept-orientation works cleanly when the concept exists across the languages in question — drug names, mathematical operators, parts in a parts catalog, ISO-defined units. It strains when concepts do not align across languages, which is most of the time outside narrowly technical domains. The notion of _saudade_ in Portuguese has no Spanish, English, or German equivalent that maps onto it as the same concept; representing the four "translations" as a single concept entry asserts a unity that does not exist. Gerhard Thurmair's 2006 EuroTermBank analysis argued exactly this: TBX's concept-orientation works for technical terminology and breaks for terminology shaped by cultural or societal forces.

For a translator working on a technical document, this is usually fine. For an academic translation of a culturally-embedded source — religious texts, philosophy, literature — the concept-oriented model has to be applied with care, treating "concept" sometimes as a shorthand for "the cluster of partially-overlapping terms we have decided to handle together," rather than as a metaphysical claim about meaning.

### Concept, language section, term section

The structural hierarchy follows the conceptual one:

- A `<conceptEntry>` represents one concept.
- Inside it, one `<langSec>` per language collects everything about that concept in that language.
- Inside each `<langSec>`, one or more `<termSec>` collects everything about a single term — because a concept may have multiple terms in the same language (a preferred form, an admitted variant, a deprecated form, an acronym).

This three-level nesting (concept → language → term) is the recurring rhythm of every TBX document. Metadata can attach at any level: a definition might attach to the concept (definition true in all languages) or to a language section (definition specific to that language's usage) or to a term (a contextual note about that particular surface form).

### The Terminological Markup Framework

Above TBX itself sits ISO 16642, the Terminological Markup Framework (TMF), published in 2003. TMF is not a format but an abstract data model — a meta-specification that says what a terminology markup language ought to be able to express. TBX is the most prominent concrete realization of TMF, but other markup languages can be TMF-compliant without being TBX. This layering is part of why TBX feels academic compared to TMX or XLIFF: it inherits a more disciplined conceptual stack from the ISO TC 37 tradition.

## The modular architecture

The 2019 revision restructured TBX as a Core plus modules, organized into dialects on a _telescoping_ principle. Understanding this structure is what makes the rest of the standard navigable.

### The Core

The Core is the non-negotiable skeleton. Every TBX document, regardless of dialect, must conform to the Core structure: an `<tbx>` root, a `<tbxHeader>`, a `<text>` containing a `<body>`, one or more `<conceptEntry>` elements, the language-section-and-term-section hierarchy within each. The Core defines only three data categories: `term`, `date`, and `note`. Everything else — definitions, parts of speech, geographic restrictions, administrative status, source attributions — comes from modules layered on top.

A document that uses only the Core is a valid TBX-Core document. It is also rather useless in practice; the Core exists to be a foundation for richer dialects, not to be used directly.

### The telescoping dialects

The public dialects extend Core in a nested fashion:

| Dialect      | Composition                           | Practical purpose                                                                        |
| ------------ | ------------------------------------- | ---------------------------------------------------------------------------------------- |
| TBX-Core     | Core only                             | The non-negotiable foundation                                                            |
| TBX-Min      | Core + Min module                     | Simple bilingual glossaries, translator handoff                                          |
| TBX-Basic    | Core + Min + Basic modules            | Practical translation-industry work with definitions, contexts, and grammatical metadata |
| TBX-Linguist | Core + Min + Basic + Linguist modules | Richer linguistic annotation for terminologists                                          |

The telescoping is important. Because TBX-Basic contains everything in TBX-Min, a tool that knows how to parse TBX-Basic can read a TBX-Min file without modification. A tool that supports only TBX-Min can read at least the Min-portion of a TBX-Basic file and degrade gracefully. This forward-compatibility is what allows the standard to support both lightweight and heavyweight uses without fragmenting into incompatible profiles.

Beyond the public dialects, the standard permits private dialects: an organization can define its own module to capture specialized metadata (regulatory authority for pharmaceutical terminology, character-name continuity for entertainment localization), and as long as the document declares its dialect on the root element, it remains a conformant TBX document.

### Public vs private dialects

A public dialect is one whose definition has been approved by the TBX Council under LTAC/TerminOrgs governance, is documented openly, and has tooling support intended on TBXinfo.net. A private dialect can be anything an organization needs. The distinction matters for interoperability: receiving a public-dialect TBX file means you can find the schema and documentation on the open web; receiving a private-dialect file means you need the producer's documentation.

### DCA vs DCT style

A subtler structural choice in TBX 3.0 is the introduction of two parallel XML styles for expressing data categories:

- **DCA (Data Category as Attribute)** — the historical TBX style. A generic element carries a `type` attribute naming the data category: `<descrip type="definition">…</descrip>`, `<termNote type="partOfSpeech">noun</termNote>`. The element names are few and generic; the categories live in attribute values.
- **DCT (Data Category as Tag)** — the new style introduced in 2019. The data category becomes the element name in a module namespace: `<basic:definition>…</basic:definition>`, `<basic:partOfSpeech>noun</basic:partOfSpeech>`. Each module contributes its own XML namespace.

Both styles express the same data model and a tool must be able to read either, but the choice affects readability, validation tooling, and how naturally the format integrates with XML-aware editors. DCT is generally easier for humans to read and for ad-hoc tooling (XPath, XSLT, simple scripts) to process. DCA is more compact and was what existing TBX 2.0 implementations already produced. The TBX-Core dialect is the only one where the two styles look identical, because Core contains no data categories beyond the structural elements themselves.

## Data categories and the DCR

A persistent thread through the TBX standards family is the relationship to a _Data Category Repository_, the registry that defines what each data category like `definition` or `partOfSpeech` formally means.

The original repository was ISOcat, established under ISO 12620:2009. ISOcat aimed to be the master registry for linguistic and terminological data categories across all language resources — not just TBX, but anything that needed to refer to standardized field types. ISOcat had governance problems: contributions came from many communities with diverging needs, the harmonization process was slow, and persistent identifiers (PIDs) were issued for categories that were later deprecated or revised, creating link-rot inside the very registry meant to provide stability.

In 2019, with the publication of ISO 12620:2019 ("Management of terminology resources — Data category specifications"), the registry was restructured. ISOcat was retired. The functional successor is **DatCatInfo**, hosted at datcatinfo.net, governed by the DatCat Council under LTAC Global / TerminOrgs, with database services provided by Interverbum Tech's TermWeb product. ISO 12620 is now a series — ISO 12620-1:2022 covering the principles of data category specification — rather than a single registry document. Critically, ISO 12620:2019 narrowed scope: it is now restricted to _terminology resources_ specifically, not the broader cross-domain remit ISOcat aspired to.

For working TBX use, the practical effect is that each data category in a TBX document conceptually references an entry in DatCatInfo, and the dialect specification (e.g., TBX-Basic) specifies which DatCatInfo categories are admissible and what their element or attribute representation is. Producers don't normally cite DatCatInfo URLs in TBX files; the dialect provides the mapping.

## Governance

The governance picture is layered enough to be worth naming explicitly, because it determines who can actually change the standard.

**ISO TC 37 / SC 3** is the international standardization body. ISO 30042:2019 was published by Subcommittee 3 (Systems to manage terminology, knowledge and content) of Technical Committee 37 (Language and terminology). Revisions go through the standard ISO process: working drafts, committee drafts, draft international standards, final drafts, publication, five-year reviews.

**LISA**, the Localization Industry Standards Association, was the trade body that originally hosted TBX through its OSCAR special interest group. LISA dissolved in 2011, releasing its standards under Creative Commons. The orphaning forced the question of who would carry on the work.

**LTAC Global** (the Language Terminology Coordination), with TerminOrgs (Terminology for Large Organizations) as a sister entity, picked up the operational stewardship — maintaining DatCatInfo, hosting tbxinfo.net, coordinating dialect definitions outside the formal ISO process, and providing the liaison to ISO TC 37/SC 3 that allows industry input into the standard's evolution.

**GALA** (Globalization and Localization Association) is the closest thing to LISA's industry-trade-body successor in the broader localization standards space. GALA hosts TMX (which never made it to ISO) and is part of the constellation around XLIFF and TBX, though it is not the formal publisher of TBX.

**FIT** (Fédération Internationale des Traducteurs), the international federation of translator associations, provides Alan Melby — the same Alan Melby who proposed the translator's workstation in 1981 — as a liaison for review participation. Melby's continued involvement is one of the few personal threads connecting the original conceptual work to the current standard.

This fragmented stewardship is part of why TBX has evolved slowly. There is no single owner; consensus must form across ISO, LTAC/TerminOrgs, and the user community before substantive changes happen.

## What TBX is not

It is worth being explicit about the boundaries.

TBX is not a translation memory format. TMX handles that. A TBX file does not record "this sentence was translated as that sentence"; it records "this term is the approved translation of that term, in this language pair, with this status."

TBX is not a bilingual working file. XLIFF handles that. A TBX file is not edited as part of the translation of a document; it is consulted alongside the editing and updated separately.

TBX is not a presentation format. ISO 30042:2019 itself states that TBX is limited in its ability to represent presentational markup. A TBX entry may carry inline highlighting or foreign-script tagging in its term forms, but it does not concern itself with how a termbase will be displayed in a tool's UI or printed in a glossary publication.

TBX is not an ontology language. Where a knowledge representation system needs typed relationships between concepts (is-a, part-of, has-property), TBX provides only weak cross-references via `<crossReference>` or `xref` elements. For richer concept modeling, the natural neighbors are SKOS (Simple Knowledge Organization System), which the W3C produced for thesaurus and taxonomy representation, and OntoLex-Lemon, the W3C Community Group's effort to bridge lexical resources and OWL ontologies. SKOS in particular has grown into a more widely adopted standard for concept hierarchies than TBX ever became, partly because of TBX's slower uptake and partly because SKOS rides on the more general RDF/Linked Data infrastructure.

## Criticisms and tensions

Several substantive critiques of TBX have surfaced over the years, and they are worth weighing rather than dismissing.

**Concept-orientation as overcommitment.** The Thurmair line of criticism, repeated by others, is that TBX projects a structural unity on terminology that does not always hold. The deepest version of this criticism is not that TBX is wrong about technical terminology — it is that by treating concept-orientation as universal, TBX exports a particular philosophical view of language to all use cases.

**Complexity.** Even the 2019 modularization, designed to simplify, produced a standard whose conformance landscape is intricate: Core vs Min vs Basic vs Linguist; DCA vs DCT style; public vs private dialects; namespace declarations; the DatCatInfo reference machinery. A 2014 paper by Romary on TEI-TBX integration described TBX 2.0 as having "too many options" and noted that this complexity had visibly slowed industry adoption, opening the field to lighter-weight alternatives like SKOS.

**Slow tool uptake.** Compared to TMX — which by 2026 every commercial CAT tool supports as a baseline — TBX support is uneven. Tools commonly support a subset of TBX-Basic; full TBX 3.0 conformance, including private dialect handling, is rare even among major commercial CAT vendors. Tbxinfo.net maintains a list of tools claiming TBX support, but the list itself acknowledges that validation of those claims is incomplete. In practice, working with TBX across tools often requires testing what subset actually round-trips correctly.

**Dialect fragmentation.** The telescoping principle is elegant but it places the burden of compatibility on tool developers, who must decide which dialects to support and to what depth. A producer can emit valid TBX-Basic that a consumer's TBX-Min-only tool will fail on, even though the standard says the consumer should degrade gracefully.

**The XCS legacy.** TBX 2.0 used XCS (XML Constraint Specification) files for dialect customization. These were proprietary in feel — a TBX-specific schema language for declaring what a dialect did and did not include. The 2019 revision removed XCS in favor of schema-language-neutral declarations using mainstream technologies like RelaxNG, but the legacy of XCS-era complexity remains in the form of older files and older tools that have not been updated.

## Where TBX still earns its keep

For all the criticisms, TBX continues to be the right tool for several use cases that newer technologies have not improved on.

**Regulated industries.** Pharmaceutical, medical-device, aerospace, and financial-services localization often have regulatory requirements about traceability of approved terminology — who approved what, when, and with what authority. TBX-Basic's `<transac>` and `<admin>` machinery is built for exactly this. Replacing it with an LLM-generated glossary is not a defensible response to an auditor.

**Large multilingual organizations.** The EU's IATE database, the World Health Organization's terminology, the UN system's vocabularies — these are concept-oriented, multilingual, and need to outlive any individual software vendor. TBX is what they export and ingest.

**Academic translation.** Where a single conceptual vocabulary must be rendered consistently across a long body of work — Greek philosophical terms in a multi-volume edition of Plato, kabbalistic concepts across the works of Scholem in three or four languages — TBX-Basic captures the disciplinary needs better than any glossary format invented since. The concept-orientation that strains for cultural translation is exactly right for the disciplined rendering of technical vocabularies in scholarly text.

**As a stable export target.** Even when actual translation work happens in newer tools — neural MT post-editing platforms, LLM-assisted glossaries — exporting the resulting terminology to TBX preserves the asset in a format that the next generation of tools will also be able to read. This is the same logic that keeps TMX alive: format stability is itself a kind of value, distinct from format expressiveness.

## TBX in the AI era

The arrival of large language models has not made TBX obsolete in any clean way. LLMs can perform many of the tasks TBX-supported workflows used to require — extracting candidate terms from a corpus, drafting bilingual glossaries, checking translation consistency — but they do these things in a way that is non-traceable, non-deterministic, and non-auditable. Where the older terminology workflow produced an artifact (the TBX file) that captured every decision with a timestamp and a responsible agent, the LLM workflow produces a conversation log at best.

The likely trajectory is that TBX persists as the _durable artifact_ layer beneath increasingly LLM-driven _production_ layers. The LLM extracts candidates, drafts entries, checks consistency; the human reviews; the resulting decisions land in a TBX file that becomes the institutional asset. This is consistent with how the standard's own pre-AI conception of the human role — inspect, accept, adapt, override — extends into a world where the proposing machine is more capable, but the standard of record remains the human-curated termbase.

Whether the successor ISO 30042 currently in drafting will adapt to this trajectory — perhaps by adding fields for LLM provenance, confidence scores, or generation timestamps — is one of the live questions in the standard's evolution. The 2019 edition already added directionality support and aligned the inline data model with XLIFF 2; the next edition could plausibly add AI-provenance affordances. The slow consensus-based ISO process makes it unlikely this happens quickly.

## Official standards and where to download them

The TBX standard sits on a stack of ISO documents. The headline standard (ISO 30042) is paywalled at the ISO catalog, but TBXinfo.net hosts freely available companion materials that ISO 30042:2019 deliberately split off into the public domain, and the LTAC-Global GitHub organization publishes the dialect schemas in full.

### Primary standard

- **ISO 30042:2019 — Management of terminology resources — TermBase eXchange (TBX)** — the current normative spec (TBX 3.0). Paid (CHF 198). [iso.org/standard/62510.html](https://www.iso.org/standard/62510.html)
- **ISO/AWI 30042** — the in-progress successor revision; tracking page only. [iso.org/standard/90295.html](https://www.iso.org/standard/90295.html)
- **ISO 30042:2008** — the superseded second edition (TBX 2.0), still useful when working with legacy files. [iso.org/standard/45797.html](https://www.iso.org/standard/45797.html)

### Underlying framework standards

- **ISO 16642:2017 — Terminological Markup Framework (TMF)** — the abstract data model TBX implements. [iso.org/standard/56063.html](https://www.iso.org/standard/56063.html)
- **ISO 12620-1:2022 — Data category specifications, Part 1: Specifications** — governs the data category model behind DatCatInfo. [iso.org/standard/79078.html](https://www.iso.org/standard/79078.html)
- **ISO 12620:2019** — superseded single-document edition, scoped to terminology resources. [iso.org/standard/69550.html](https://www.iso.org/standard/69550.html)
- **ISO 1087:2019 — Terminology work and terminology science — Vocabulary** — defines the terms used inside the TBX spec itself. [iso.org/standard/62330.html](https://www.iso.org/standard/62330.html)
- **ISO 704:2022 — Terminology work — Principles and methods** — the conceptual substrate. [iso.org/standard/79077.html](https://www.iso.org/standard/79077.html)

### Free companion materials (no paywall)

- **TBXinfo.net** — the official companion site maintained by LTAC/TerminOrgs. [tbxinfo.net](https://www.tbxinfo.net/)
- **TBX-Basic dialect — schemas (RNG, XSD) for DCA and DCT styles** — LTAC-Global GitHub. [github.com/LTAC-Global/TBX-Basic_dialect](https://github.com/LTAC-Global/TBX-Basic_dialect)
- **TBX-Min dialect — schemas and documentation** — LTAC-Global GitHub. [github.com/LTAC-Global/TBX-Min_dialect](https://github.com/LTAC-Global/TBX-Min_dialect)
- **TBX-Core schemas and the core structure RNG/XSD** — browse the TBX repositories under the LTAC-Global organization. [github.com/LTAC-Global](https://github.com/LTAC-Global)
- **DatCatInfo** — the live Data Category Repository referenced by ISO 30042 dialects. [datcatinfo.net](https://datcatinfo.net/)
- **Historical TBX 2.0 specification** — preserved at ttt.org, useful for understanding the lineage. [ttt.org/oscarStandards/tbx/](https://www.ttt.org/oscarStandards/tbx/)
- **TerminOrgs TBX-Basic landing page** — links to the PDF specification of TBX-Basic. [terminorgs.net/TBX-Basic.html](https://www.terminorgs.net/TBX-Basic.html)

## Further reading

- TBX Dialects page on TBXinfo, including the telescoping principle and the public/private dialect distinction. [tbxinfo.net/tbx-dialects](https://www.tbxinfo.net/tbx-dialects/)
- TBX Data Category Modules. [tbxinfo.net/tbx-modules](https://www.tbxinfo.net/tbx-modules/)
- "TBX Version 3 published at ISO," MultiLingual, July/August 2019 — the standards-community announcement that explains the 2019 changes. [multilingual.com](https://multilingual.com/issues/july-aug-2019/tbx-version-3-published-at-iso/)
- Thurmair, G. (2006). EuroTermBank-related critique of TBX's concept-based, non-directional model.
- Romary, L. (2014). "TBX goes TEI — Implementing a TBX basic extension for the Text Encoding Initiative guidelines." [arxiv.org/pdf/1403.0052](https://arxiv.org/pdf/1403.0052)
- Wikipedia, "TermBase eXchange." [en.wikipedia.org/wiki/TermBase_eXchange](https://en.wikipedia.org/wiki/TermBase_eXchange)
