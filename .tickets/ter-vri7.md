---
id: ter-vri7
status: closed
deps: []
links: [ter-7c2c, ter-ph7i]
created: 2026-05-26T03:19:44Z
type: bug
priority: 2
assignee: Andre Silva
parent: ter-bf0v
tags: [e4, bug, extract, heuristic]
---
# E4.BUG — Foreign-script heuristic ignores frontmatter lang for base script

## Summary

When a markdown file has `lang: he` in YAML frontmatter, the extract
command's foreign-script heuristic still treats Hebrew as the foreign
script (and Latin as the base), rather than switching the base to Hebrew.

## Expected

With `lang: he`, Latin-script tokens ("concept", "The", "Kabbalah",
"teaches", etc.) should be flagged as `foreign_script` candidates, since
they are foreign relative to the document's declared Hebrew base.

## Actual

Hebrew tokens (הוא, הצמצום, מרכזי) are flagged as `foreign_script`.
Latin tokens are not. The heuristic behaves as if the document language
is English regardless of the frontmatter.

## Steps to reproduce

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"
QA_TMP=$(mktemp -d)
cat > "${QA_TMP}/hebrew-doc.md" <<'FRONTMATTER'
---
lang: he
---

# מבוא

הצמצום הוא concept מרכזי. The Kabbalah teaches about divine light.
FRONTMATTER

${TT} extract "${QA_TMP}/hebrew-doc.md" | jq '[.candidates[] | select(.heuristic == "foreign_script") | .term]'
# Actual: ["הוא","הצמצום","מרכזי"]
# Expected: Latin-script tokens like "concept", "Kabbalah", etc.

rm -rf "${QA_TMP}"
```

## Refs

- QA report: [qa/E4-manual-qa.report.md](qa/E4-manual-qa.report.md) Observation 1
- QA plan: [qa/E4-manual-qa.md](qa/E4-manual-qa.md) TC-LANG-DET-001
- E4 spec: [docs/specs/004-read-commands.md](docs/specs/004-read-commands.md)
  §"extract — language detection precedence"
- Foreign-script heuristic: `src/internal/extract/foreign.go`
- Language detection: `src/internal/app/commands/extract.go`

## Fix

The detected language (from frontmatter, `--lang`, or default) must be
mapped to a Unicode script and passed to the foreign-script heuristic as
the base script. Currently the base script appears to be hardcoded or
derived from the first paragraph's dominant script without considering
the frontmatter.


## Notes

**2026-05-26T03:51:05Z**

Fixed by adding BaseLang field to extract.Options. When set, ForeignScriptTokens uses langToScript(BaseLang) as the document-wide base script instead of computing dominantScript per-span. The extract command action now sets fileOpts.BaseLang = fs.lang (from frontmatter/flag/default) before calling ForeignScriptTokens. langToScript maps BCP 47 tags to Unicode script tables (he→Hebrew, ar→Arabic, ru→Cyrillic, el→Greek, etc., default→Latin). Unit tests in foreign_test.go cover BaseLang=he, BaseLang=es, and empty-BaseLang fallback. Integration test TestExtract_FrontmatterLangAffectsForeignScript validates the full pipeline with a Hebrew-frontmatter fixture.
