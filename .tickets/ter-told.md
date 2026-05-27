---
id: ter-told
status: closed
deps: [ter-uqyn, ter-19rb]
links: []
created: 2026-05-22T19:19:19Z
type: epic
priority: 2
assignee: Andre Silva
tags: [epic, validate, read]
---
# E3 — terminology validate

Spec: docs/specs/003-validate-command.md

First real command end-to-end. Three validation tiers:

1. Well-formedness — XML parses, required structure present.
2. Schema/dialect — every element in TBX-Linguist supported set; attributes well-typed; picklist values in the accepted set.
3. Semantic — BCP 47 well-formedness (x/text/language.Parse — syntactic only), unique concept IDs, crossReference IDREFs resolve, dialect consistency rules.

--strict opt-in: promotes unknown-element silent→warning and unresolved-crossref warning→error. Tier 1 short-circuits; tiers 2+3 run together aggregating warnings.

Populates internal/tbx/picklist.go (single source shared with E7 urfave enums). Reports as-found concept counts (collisions reported as warnings, not masked). languages array sorted ASCII byte order. New sentinel: ErrValidationError (exit 65 = EX_DATAERR).

Acceptance: warnings carry concept_id + line/col; --strict promotes correctly; tier 1 short-circuits.


## Notes

**2026-05-24T00:15:30Z**

E2 implementation included validate.go + validate_test.go (~300 LOC) in internal/tbx. This covers: Glossary.Validate(strict bool) ValidateResult method, ValidateResult struct (Concepts int, Languages []string, Warnings/Errors []Warning), duplicate concept ID detection, invalid BCP 47 lang tag detection (via x/text/language.Parse), missing-term-in-langSec warning, unresolved cross-reference detection (warning in lenient, error in strict), and ASCII-sorted language list. This overlaps with E3 spec tiers 2-3. E3 tickets should account for this existing code: some validation logic may already be in place and needs to be reviewed for spec alignment, extended with tier-1 well-formedness checks, and wired into the validate command's envelope output.
