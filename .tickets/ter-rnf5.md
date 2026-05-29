---
id: ter-rnf5
status: closed
deps: [ter-wer0]
links: []
created: 2026-05-29T01:13:40Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-8gyy
tags: [beta-feedback, docs]
---
# BETA — document admitted variant key and minimal TBX skeleton

Beta feedback: the 'admitted' variant key is undocumented — the JSON-stdin example in docs/skills/terminology/references/write-details.md:31-51 shows only 'preferred'. Update that example to show all four term-group keys from output/types.go WriteTermGroup: preferred (object), admitted (array), deprecated (array), superseded (array). Mirror the admitted array into docs/skills/terminology/SKILL.md. Also document the minimal valid TBX skeleton (now produced by 'terminology init') in SKILL.md so the shape is discoverable without probing.

## Acceptance Criteria

write-details.md JSON example shows preferred + admitted + deprecated + superseded with correct shapes
SKILL.md includes an admitted example and documents the minimal TBX skeleton / references terminology init
Docs consistent with code (WriteTermGroup) and the init command output
No stale claims about header handling


## Notes

**2026-05-29T01:32:46Z**

Updated docs/skills/terminology/references/write-details.md JSON example to show all four WriteTermGroup keys (preferred/admitted/deprecated/superseded) with a key-shape reference table; mirrored an admitted example into the apply payload format. Updated SKILL.md to add a new 'init' command section documenting the minimal valid TBX skeleton (xmlns prefixes, tbxHeader, empty text/body) and noting that <titleStmt> is omitted when --title is not supplied; expanded the concept add JSON-stdin example to demonstrate all four variant arrays. Verified end-to-end against the binary: init produces the documented skeleton and concept add accepts a payload with all four keys, emitting four <termSec> blocks under the same <langSec>.
