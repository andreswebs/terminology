---
id: ter-wer0
status: closed
deps: [ter-z62o]
links: []
created: 2026-05-29T01:13:40Z
type: feature
priority: 1
assignee: Andre Silva
parent: ter-8gyy
tags: [beta-feedback, write, init]
---
# BETA — add terminology init command (create minimal valid TBX)

Beta feedback: there is no way to create a TBX; users hand-author the skeleton. Add a 'terminology init' command that writes a minimal valid TBX-Linguist skeleton (root + tbxHeader/fileDesc + empty <text><body>). Register in src/internal/app/root.go; new src/internal/app/commands/init.go. Decisions (Andre): --source-lang REQUIRED (missing -> exit 2 usage error) so we never silently emit the wrong xml:lang; --title optional (seeds <titleStmt>); refuse to overwrite an existing target (clean io_error) with NO --force; support standard --dry-run preview. Emit via LinguistWriter, seeding SourceLang from --source-lang and titleStmt from --title.

## Design

Depends on the header-preservation work so --source-lang lands in root xml:lang and --title in titleStmt. Add schema introspection support (terminology schema --command init) and a JSON result envelope consistent with other write commands. Reuse io_error for the already-exists refusal.

## Acceptance Criteria

make build passes
terminology init --tbx PATH --source-lang LANG creates a valid TBX-Linguist skeleton that passes terminology validate
--source-lang required; omission is a usage error (exit 2)
--title seeds <titleStmt>; root xml:lang reflects --source-lang
Refuses to overwrite an existing target (io_error); no --force flag
--dry-run previews the skeleton without writing
terminology schema --command init describes flags and envelope
Golden test for init output and integration test for the validate round-trip


## Notes

**2026-05-29T01:30:31Z**

Added 'terminology init' command (src/internal/app/commands/init.go) registered in src/internal/app/root.go. Flags: --source-lang (required, BCP47-validated), --title (optional), --dry-run. Refuses to overwrite existing target (io_error / exit 3); no --force. Emits new output.InitEnvelope (schema_version, ok, source_lang, title, dry_run). Reuses tbx.Save / LinguistWriter — header preservation work (ter-z62o) means --source-lang lands in root xml:lang and --title in <titleStmt>. Tests in src/internal/app/init_test.go cover happy path, validate round-trip, missing --source-lang (exit 2), existing file (exit 3), dry-run (no write), invalid lang tag, schema --command init introspection, plus a golden in testdata/init/happy. Path intentionally omitted from envelope so the JSON output is stable (test temp dirs vary).
