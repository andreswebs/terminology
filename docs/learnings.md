# Learnings

## E1.T2 — internal/output

- `AssertEnvelopeShape` conformance helper uses `t.Errorf` (not `t.Fatalf`) so that negative tests can pass a bare `*testing.T{}` without triggering `runtime.Goexit` panics.
- `EmitError` takes `format string` as a parameter rather than reading from `*cli.Context`, keeping the output package free of urfave dependencies.
- The `_, _ =` pattern satisfies `errcheck` for stderr writes at the process exit boundary where there's no meaningful recovery.

## E1.T3 — internal/version

- `go tool nm` cannot inspect cross-compiled, stripped binaries (`-s -w` ldflags); use `strings` to verify symbol injection instead.
- Under `go test`, `debug.ReadBuildInfo().Main.Version` is always `"(devel)"`, so the BuildInfo-with-real-version branch can only be integration-tested via `go install ...@vX.Y.Z` (deferred to E10).

## E1.T4 — internal/logctx

- `crypto/rand.Read` with `_, _ =` is the correct ignore pattern here — the stdlib documents that it always returns `len(b), nil` on supported platforms, and there's no meaningful recovery at run-ID generation time. This is consistent with the `_, _ =` convention established in E1.T2 for stderr writes.
- The wrong-type fallback test (`context.WithValue(ctx, ctxKey{}, 42)`) works because the test file is in the same package, giving access to the unexported `ctxKey` — no test helper or exported key needed.

## E1.T5 — Tracer bullet (Root + stub + main + harness)

- `underConstruction` lives in `internal/app/commands` (package-private) rather than exported from `internal/app`, to avoid the import cycle: `app` → `commands` → `app`.
- urfave v3 `cmd.FullName()` includes the root command name (e.g. `"terminology validate"`), so `terr.UnderConstruction(cmd.FullName())` produces a doubled prefix (`"terminology terminology validate"`). The `stubPath()` helper strips the root name prefix.
- urfave v3's default `ExitErrHandler` prints errors to stderr. Setting it to a no-op on the root command keeps `main.go` as the single envelope emission point per the error-handling ADR.
- urfave v3 `ExitErrHandlerFunc` signature is `func(context.Context, *cli.Command, error)` (3 params, not 2).
- `rootAction` returns a `terr.New` sentinel (not `urfcli.Exit`) so that `output.ExitCodeFor` (which checks for `terr.Coded`) correctly extracts exit code 2. urfave's `ExitCoder` interface is not recognized by the `terr.Coded` path.
- Golden test harness: `WriteFile` errors must be checked to satisfy `errcheck` lint — use a `writeGolden` helper that calls `t.Fatalf` on failure.

## E1.T9 — Command stub: extract

- `enumValidator` in `root.go` (package `app`) is not accessible from `commands` package. Per-command enum validators (e.g. `scriptValidator`) live in the command file itself. T17 (shared flag groups refactor) is the planned consolidation point.
- urfave v3 `StringArgs` with `Min: 1, Max: -1` enforces at least one variadic positional; omitting all args produces an error from urfave before the action fires.
- urfave `Validator` on `StringFlag` fires for the default `Value` as well, so the default `"any"` must be in the picklist.

## E1.T12 — Command stub: concept parent + concept add

- Parent commands (help-only, no business logic) must use `terr.New` for the "no subcommand specified" error — not `urfcli.Exit` — because `output.ExitCodeFor` only recognizes `terr.Coded`, not urfave's `ExitCoder`. This is the same pattern as `rootAction` (learned in T5).
- The `conceptEnumValidator` follows the per-command local validator pattern established in T9 (`scriptValidator` in `extract.go`), not the `enumValidator` in `root.go` (package `app`), which is inaccessible from `commands`.

## E1.T13 — Command stub: concept update

- The `Before` hook on a urfave `Command` is the right place for flag-mutex checks (`requireMergeXorReplace`). Returning a `terr.New` sentinel from `Before` follows the same error-handling path as `Action` — `output.ExitCodeFor` extracts the exit code and `EmitError` produces the structured envelope.
- Error sentinels for the mutex (`errMergeReplaceMutex`, `errMergeReplaceRequired`) live in the command file itself (package `commands`), not in `app/errors.go`, because they are command-specific. The `app/errors.go` pattern is for cross-cutting concerns like `ErrConflictingVerbosity`.
- Golden tests support subdirectories: `runGolden(t, "concept_update/merge", ...)` writes to `testdata/concept_update/merge/`.

## E1.T6 — Command stub: lookup

- urfave v3 `StringArg` (singular, `ArgumentBase`) has no `Min`/`Max` fields — it always represents exactly one required positional. `StringArgs` (plural, `ArgumentsBase`) has `Min`/`Max` for variadic positionals. The ticket's surface spec referenced `Min: 1, Max: 1` which only applies to the plural form.

## E1.T17 — Refactor: extract shared flag groups

- `urfcli.ValueSourceChain` is a struct (not a pointer), so `sf.Sources == nil` doesn't compile. Use `len(sf.Sources.Chain) == 0` to check for absence of env sources.
- The `--script` flag on `extract` is the only enum flag with a `Value` default (`"any"`), so it was kept as a local `scriptPickFlag()` helper rather than generalizing the shared `pickFlag` constructor with a rarely-used default parameter.
- When extracting shared flag constructors, passing usage strings as parameters (e.g. `dryRunFlag(usage)`, `langFlag(required, usage)`) preserves per-command wording without introducing a surface change.

## E2.T7 — Linguist DCA reader

- DCA decode functions (`decodeConceptDCA`, `decodeLangSecDCA`, `decodeTermDCA`) were implemented alongside DCT in T6, sharing the same dispatch pattern via `style` parameter in `decodeConceptChild`, `decodeLangSec`, and `decodeTermSec`.
- DCA self-closing elements (`<xref type="externalCrossReference" target="..."/>`) work identically to DCT self-closing elements — `readCharData` returns empty string and the `target` attribute supplies the value.
- Comprehensive DCA test coverage reuses shared assertion helpers (`assertAllTermCategories`, `assertAllConceptCategories`, `assertLangSecCategories`) with both DCT and DCA fixtures, avoiding test logic duplication.
- `TestRoundTrip_Canonical` must skip all `*-dca.tbx` files (not just `minimal-dca.tbx`) since DCA→write always produces DCT.

## E2.T6 — Linguist DCT reader (core decoder)

- Self-closing elements (`<min:externalCrossReference target="..."/>`) are handled correctly by `readCharData`: the XML decoder emits StartElement then EndElement immediately, so `readCharData` returns empty string and the `target` attribute supplies the value.
- `readNoteText` tracks inline markup depth to distinguish the outer element's EndElement from nested inline EndElements. The `depth` counter ensures `<hi>bold</hi>` inside a `<basic:definition>` is captured in `Raw` but stripped from `Plain`.
- The `all-categories-dct.tbx` fixture must follow the exact element order the writer produces (e.g. subjectField → definition → crossReference → externalCrossReference → xGraphic → source → customerSubset → projectSubset → note → langSec) to pass round-trip tests.

## E2.T8 — Linguist reader: adminGrp + transacGrp decoders

- `decodeAdminGrp` and `decodeTransacGrp` were already implemented in `reader.go` during T6/T7. T8's contribution was adding explicit unit tests for all paths (DCT/DCA, concept/term level, unknown children, langSec-level skip).
- DCA transaction tests use inline XML fixtures rather than file fixtures, since the DCA→DCT conversion is already covered by `TestDCAtoCanonicalDCT` and there's no need for a separate canonical DCA transactions file.
- The `with-transactions.tbx` canonical fixture was updated to include term-level transactions. The round-trip test (`TestRoundTrip_Canonical`) automatically covers this fixture, providing byte-for-byte validation.

## E1.BUG — Positional argument bounds (Min/Max) not enforced

- urfave v3 `StringArg` (singular) `Parse()` silently returns when `len(s) == 0` — it does not enforce that an argument was provided. Only `StringArgs` (plural) checks `Min`, and neither form rejects excess arguments (they go to `parsedArgs`).
- Fix: `argBounds(min, max)` returns a `urfcli.BeforeFunc` that checks `cmd.Args().Len()` against declared bounds, returning `terr.New` sentinels with `missing_argument` or `excess_arguments` codes (exit 2).
- `chainBefore(fns ...urfcli.BeforeFunc)` composes multiple `Before` hooks sequentially. Used on `concept update` where `argBounds(1, 1)` must run before `requireMergeXorReplace`.
- `cmd.FullName()` in the `Before` hook includes the root name (e.g. `"terminology lookup"`), which is useful for error messages but matches the doubled-prefix behavior noted in T5. The `Before` context uses it intentionally for user-facing messages.

## E2 — Domain model & TBX I/O

- `tbx/` → `tbx/linguist/` → `tbx/` import cycle is broken via a registry pattern: `tbx.RegisterDialect()` accepts constructor functions, and `linguist/register.go` calls it at `init()` time. `app/root.go` imports `_ "internal/tbx/linguist"` to trigger registration.
- `encoding/xml` is used for parsing (adequate) but not for writing (nondeterministic namespace prefixes, whitespace issues). The canonical DCT writer is hand-rolled (~250 LOC) with fixed prefix/indent/ordering rules.
- Round-trip tests compare byte-for-byte against canonical fixtures. Non-canonical inputs (DCA, legacy forms) are tested via separate fixtures with known expected outputs.
- `lockedfile.Create` from `rogpeppe/go-internal` handles cross-process advisory locking. Lock files are cleaned up after write.
- `SourceDesc` from `<tbxHeader>` must be preserved in the model for round-trip fidelity — the writer cannot hardcode it.
- Term sort order in writer: preferred → admitted → deprecated → superseded → unspecified, with `sort.SliceStable` to preserve declaration order as tiebreak.

## E2.T10 — Linguist canonical DCT writer tests

- The writer implementation (`writer.go`) was already in place when the test ticket was reached; `writer_test.go` was the missing deliverable.
- `failWriter` test double (writes succeed up to N bytes, then error) is sufficient for testing the xmlBuilder sticky-error pattern — no need for more sophisticated mocks.
- `sortedConcepts`, `sortedTerms`, and `sortedLangs` all return copies, so tests verify the original slice is not mutated.

## E2.T14 — Test fixtures + round-trip property tests

- All canonical fixtures and the core `TestRoundTrip_Canonical` / `TestDCAtoCanonicalDCT` tests were already created incrementally during T6–T10. T14's contribution was filling acceptance criteria gaps: DCA model equivalence (`reflect.DeepEqual` after normalizing `Style`), legacy normalization write verification, and multi-pass stability.
- `reflect.DeepEqual` works well for DCA/DCT model equivalence because the only expected difference is `Glossary.Style` — setting it to a common value before comparison is simpler than writing field-by-field assertions.
- The stability test (read→write→read→write, compare two outputs) is a stronger property than single round-trip: it catches cases where read-after-write introduces drift that a single pass wouldn't reveal.

## E2.T13 — Save + atomic write + advisory lock

- `terr.E.Wrap(cause)` creates a copy of the sentinel (`cp := *e`), so `errors.Is(wrappedErr, ErrTBXLocked)` fails — the wrapped error is a different pointer with no `Is` method. Use `terr.Coded` type assertion and check `.Code()` instead: `coded, ok := err.(terr.Coded); ok && coded.Code() == "tbx_locked"`.
- `acquireLock` was extracted from `io.go` into `lock.go` per the spec's architecture diagram. The function remains unexported — only `Save` calls it.
- Testing `ErrTBXLocked` via `Save` to a non-existent directory triggers the lock acquisition failure path without needing to simulate file-level lock contention.

## E1.BUG — Unknown subcommand not identified by name in error message

- urfave v3 does not raise an error for unknown subcommands — it falls through to the parent command's `Action`. The unknown token ends up in `cmd.Args()`.
- Fix: `rootAction` checks `cmd.Args().Len() > 0` and returns a distinct `terr.New` with code `unknown_subcommand` and the name embedded in the message. The bare-invocation case (no args) still returns the original `no_subcommand` sentinel.
- The error is constructed inline (not a package-level `var`) because the message includes the dynamic subcommand name via `terr.New`'s `fmt.Sprintf` formatting.

## E1.BUG — urfave-origin errors exit 1 instead of exit 2

- Urfave wraps validator errors (`urfcli.Exit(msg, 2)`) with `fmt.Errorf("invalid value %q for flag -%s: %v", ...)` using `%v` not `%w`, so the `ExitCoder` interface is lost before the error reaches `main.go`.
- Go's `flag` package errors ("flag provided but not defined") are plain `error`s that never had `ExitCoder`.
- Fix: `classifyUsageError()` in `output/errors.go` pattern-matches known urfave error prefixes to assign exit code 2 and structured error codes. This avoids importing urfave in the output package.
- Known urfave error patterns: `"flag provided but not defined"` → `unknown_flag`, `"invalid value"/"could not parse"` → `invalid_value`, `"Required flag(s)"` → `missing_required_flag`, `"cant duplicate this flag"` → `duplicate_flag`, `"option ... cannot be set along with"` → `mutually_exclusive_flags`, `"sufficient count of arg"` → `missing_argument`.

## E1.BUG — extract no-args urfave error not classified as usage error

- urfave v3 `StringArgs` with `Min: 1` emits `"sufficient count of arg <name> not provided, given 0 expected 1"` when no positional arguments are supplied. This is a distinct error format from the flag-related patterns — it uses `"sufficient count of arg"` as the prefix.
- Fix: added this prefix to `classifyUsageError()` → `missing_argument` code, exit 2. Same pattern-matching approach as the flag classifier from ter-smt4.

## E3.T9 — Tier-2: unknown_element detection

- Setting `Warning.ConceptID` at the `decodeConceptEntry` level (after all child warnings have bubbled up) avoids threading a `conceptID string` parameter through every nested decode function signature. The pattern: `for i := range warnings { if warnings[i].ConceptID == "" { warnings[i].ConceptID = c.ID } }`.
- DCT and DCA unknown elements need different message formats: DCT uses `se.Name.Local` (since the namespace prefix isn't available from `encoding/xml`), DCA includes the `type` attribute value (e.g. `<descrip type="custom">`) because the carrier element name alone (`descrip`) is a known element — the unknown part is the type.
- `decodeTransacGrp` doesn't return `[]tbx.Warning` (only `error`), so unknown elements inside transaction groups remain silent. This is acceptable for v1; transaction groups have a small known element set. (Note: T8 changed this — `decodeTransacGrp` now returns `(Transaction, []Warning, error)` to support `invalid_picklist` warnings on `transactionType`.)

## E3.T8 — Tier-2: invalid_picklist validation on read

- `checkPicklist` validates the raw text value BEFORE normalization (e.g. before `normalizeStatus()` or `normalizeRegister()`). The picklist slices from `tbx.AdminStatus()` and `tbx.Register()` include both legacy and normalized forms, so legacy forms like `preferredTerm` and `usageRegister` pass validation without producing warnings.
- `decodeTransacGrp` signature was changed from `(Transaction, error)` to `(Transaction, []Warning, error)` to propagate `invalid_picklist` warnings for `transactionType`. All three callers (concept-level in `decodeConceptChild`, term-level in `decodeTermSec`) were updated.
- The `ConceptID` field on `invalid_picklist` warnings is populated by the existing `decodeConceptEntry` backfill loop (`warnings[i].ConceptID = c.ID`), so `checkPicklist` doesn't need to know the concept ID.

## E3.T11 — LineIndex wiring into reader

- The `lineindex.Index` API uses `New(r io.Reader)` which reads all data upfront — there is no streaming `Wrap` method despite the ticket spec referencing one. The approach: `io.ReadAll(r)` once, then `lineindex.New(bytes.NewReader(data))` and `xml.NewDecoder(bytes.NewReader(data))` from the same buffer.
- A `decodeCtx` struct bundles `*xml.Decoder` + `*lineindex.Index` with a `pos()` helper. All decode functions changed from `*xml.Decoder` to `*decodeCtx`; `readCharData` and `readNoteText` kept with `*xml.Decoder` since they don't emit warnings.
- `xml.Decoder.InputOffset()` returns the byte offset right after the most recently consumed token. For `unknown_element` warnings, `dc.pos()` is called after the `StartElement` token is consumed — the position points to the opening tag's line, which is accurate.
- For `invalid_picklist` warnings, position must be captured BEFORE `readCharData()`, because `readCharData` advances the decoder past the closing tag. Capturing after would report the wrong line for multi-line elements.

## E3.T10 — legacy_form_normalized + --strict promotions

- `isLegacyStatus()` and `isLegacyRegister()` in `normalize.go` detect whether the raw input was a legacy form, separate from `normalizeStatus`/`normalizeRegister` which return the canonical value. Checking the before/after values of `normalizeStatus` is awkward because it returns a `tbx.Status` enum, not a string — a dedicated predicate is cleaner.
- Legacy warning emission happens at all 4 normalization call sites (DCT and DCA paths for both admin status and register). Each site collects `invalid_picklist` and `legacy_form_normalized` warnings into a shared slice, since a legacy form can validly appear in the picklist (it's accepted on read) but still trigger the normalization warning.
- Strict-only filtering belongs in the validate command (`isStrictOnly`), not in the reader or `Validate()`. The reader always collects all warnings regardless of mode, keeping the separation clean: reader = collect, command = filter/route.
- The `with-legacy-and-unknown.tbx` test fixture combines legacy forms AND unknown elements in one file, enabling a single integration test to verify both filtering paths.

## E3.T4 — Tier-1 well-formedness validation

- Tier-1 well-formedness is entirely implicit: `encoding/xml.Decoder` rejects malformed XML, and `detectDialect` in `tbx/io.go` returns `ErrUnsupportedDialect` for missing/wrong `type` attributes. No structural code changes were needed.
- The validate command's `validateAction` already wraps `tbx.Load` errors with `ErrValidationError.Wrap(err)`, producing exit 65 for any tier-1 failure.
- Empty `<body>` (no `<conceptEntry>` elements) is not a tier-1 failure — the reader returns a valid Glossary with `len(Concepts) == 0`. The distinction: tier-1 catches XML that can't be parsed, while tier-3 (T5/T6/T7) catches semantic issues like `missing_term`.
- Malformed XML fixtures must be carefully crafted — `encoding/xml` is lenient about some things (e.g. mismatched tags at certain nesting levels may not error). An unclosed `<term>` tag inside `<termSec>` reliably triggers a parse error because `readCharData` hits unexpected tokens.

## E3.T13 — Validate golden CLI tests

- The `with-legacy-and-unknown.tbx` fixture covers both `unknown_element` and `legacy_form_normalized` warnings in a single file, so `strict_with_legacy` and `unknown_element_strict` golden tests produce identical output. Separate tests are kept for conceptual coverage clarity.
- `invalid_picklist` warnings are not strict-only — they appear in both lenient and strict modes. Only `unknown_element` and `legacy_form_normalized` are filtered by `isStrictOnly()`.
- Golden files contain `<` / `>` for `<` / `>` in JSON because Go's `encoding/json` HTML-escapes angle brackets by default. This is correct and deterministic.

## E3.BUG — Tier-1 accepts TBX missing `<text><body>` structure

- Empty `<body>` (present but no `<conceptEntry>` children) is valid and returns an empty glossary. Missing `<body>` entirely (no `<text>` element at all) is a tier-1 structural failure. The `inBody` flag in `Decode()` distinguishes the two cases: after the parsing loop, `!inBody` means the element was never encountered.
- The fix is a single check after the parsing loop — no need to thread state through nested decode functions since `<body>` is a top-level element handled directly in the loop.

## E3.BUG — Line/column not populated on Glossary.Validate() warnings

- Tier-3 semantic warnings (`duplicate_id`, `unresolved_crossref`, `invalid_lang_tag`, `missing_term`) operate on the post-parse `Glossary` model and had no access to XML source positions. The fix stores `StartLine`/`StartCol` directly on `Concept` and `LangSection` during decode, avoiding the need to pass `LineIndex` to `Validate()`.
- Position capture for `<conceptEntry>` is done in the main decode loop (before `decodeConceptEntry` is called), not inside `decodeConceptEntry`, because `dc.pos()` returns the offset after the most recently consumed token — calling it right after the StartElement case match gives the correct line for the opening tag.
- Position capture for `<langSec>` is done at the `decodeConceptChild` call site (before `decodeLangSec`), passed as parameters. This avoids the issue of `dc.pos()` inside `decodeLangSec` pointing past the start element.
- `invalid_lang_tag` and `missing_term` use `LangSection` position (more precise), while `duplicate_id` and `unresolved_crossref` use `Concept` position.
- The position fields don't affect the writer or round-trip tests — they are zero-valued in programmatically constructed glossaries and ignored by the canonical DCT writer.

## E4.T4 — Terr sentinel registry

- `terr.New()` auto-registers sentinels into a package-level slice. This is safe because `New()` is called during `var` initialization (single-threaded). `terr.All()` returns a copy.
- `terr.Newf()` was added as a non-registering constructor for runtime error creation (e.g. `argcheck.go` bounds errors, `root.go` unknown-subcommand). These are dynamic errors with interpolated messages, not sentinel declarations.
- `UnderConstruction()` was changed to construct `*E` directly instead of calling `New()`, so ephemeral stub errors don't pollute the registry.
- Integration tests for the registry belong in `app_test` (which transitively imports all sentinel-declaring packages), not in `terr_test` (which only sees the `terr` package itself).

## E4.T2 — --fields projection engine

- `collectJSONPaths` must dereference pointer types at two levels: map values (`map[string]*Struct`) and slice elements (`[]*Struct`). The initial implementation handled `map[string]Struct` but not `map[string]*Struct`, causing wildcard paths like `languages.*.preferred.term` to fail validation when the map value type was a pointer.
- `fmt.Errorf("%w: ...", ErrInvalidField)` preserves the `errors.Is` chain while adding context (the bad path name and valid paths hint). No need for `ErrInvalidField.Wrap()` — stdlib `%w` wrapping is sufficient and more idiomatic.
- `ProjectFields` uses a trie built from the dotted field paths to efficiently project nested JSON structures. The wildcard `*` in the trie matches all keys in a JSON object (map), which handles both explicit map types and any dynamic key structure in the serialized JSON.

## E4.T5 — Lookup envelope type

- Go's `encoding/json` marshals a nil slice as `null`, not `[]`. For envelope types where the spec requires `"results":[]` (never `null`), a custom `MarshalJSON` using a type alias (`type Alias LookupEnvelope`) coerces nil to empty slice before delegation. The alias avoids infinite recursion.

## E4.T3 — Markdown spans package (goldmark)

- Goldmark's `ast.Text` nodes carry a `text.Segment` with `Start`/`Stop` byte offsets into the original source, making offset preservation trivial — no separate position mapping needed.
- `ast.KindCodeBlock` (indented code blocks) is a separate kind from `ast.KindFencedCodeBlock` — both must be explicitly skipped in the AST walk.
- The walk function returns `bool` for early termination (iterator protocol): returning `true` from the skip cases means "continue walking siblings", while `false` from `yield` means "caller broke out of iteration".
- Goldmark treats soft line breaks as separate `ast.Text` nodes within a paragraph, so `"hello\nworld"` produces two text spans — this is the desired behavior for line/col accuracy in downstream `scan`/`check` consumers.

## E4.T6 — Lookup: case-fold + NFC matching

- `cases.Fold()` from `golang.org/x/text/cases` does not require a `language.Tag` argument — it performs Unicode default case-folding which is language-independent, unlike `cases.Lower()`.
- `Lookup` deduplicates by concept: if the same term surface appears in multiple language sections (e.g. "tzimtzum" in both `en` and `es`), only the first match per concept is returned. The `matchConcept` helper returns on the first matching term, which naturally produces one `LookupMatch` per concept.
- `golang.org/x/text/language` was not needed — `cases.Fold()` takes no arguments. The ticket's spec suggested it would be needed but `Fold()` is the Unicode default casefold, not a locale-specific operation.

## E4.T7 — Lookup command action + exit codes

- urfave v3 `StringArg` values must be accessed via `cmd.StringArg("name")`, not `cmd.Args().First()`. After `StringArg` consumes the positional, `cmd.Args()` returns only unmatched/remaining args (empty for a single-arg command). This is distinct from the `argBounds` pattern which checks `cmd.Args().Len()` — that works because `Before` hooks run before argument parsing populates the `StringArg` slot.
- The `lookupNotFoundError` type follows the same pattern as `warningsError` in `validate.go`: a private struct implementing `terr.Coded` with `ExitCode() int { return 1 }`. The envelope is emitted to stdout with `ok: true` before the error is returned, so exit 1 signals "no results" without producing an error envelope on stderr.
- `buildLookupResult` maps `StatusPreferred` and `StatusUnspecified` to the `Preferred` slot, `StatusAdmitted` to the `Admitted` slice, and omits `StatusDeprecated`/`StatusSuperseded` from the lean output (they would appear only with `--fields` opt-in).
- `TestUnderConstruction_ReturnsCoded` was using `lookup` as its example stub command. After implementing lookup, updated the test to use `scan` (still a stub) to avoid a false failure.

## E4.T11 — Extract: capitalized phrases heuristic

- `extract.Span` is a separate type from `markdown.Span` to avoid coupling the extract package to goldmark. The fields are identical (`Text`, `Line`, `Col`, `Offset`); the command action will convert between them.
- Each span is treated as a paragraph/sentence start (`afterSentenceStart=true` resets per span), which correctly handles goldmark producing separate text nodes for different paragraphs.
- Sentence boundary detection checks trailing `.`, `!`, `?` on word tokens (not just separators), because punctuation attaches to the preceding word in the tokenizer (e.g. "sat." is a single token). The `endsSentence` check must fire both in the `afterSentenceStart` branch (to chain consecutive sentence-ending words like "Mr.") and after the normal capitalized/lowercase branch.
- `stripTrailingPunct` is needed because word tokens include trailing punctuation (e.g. "Temple." → "Temple"). Uses `unicode.IsPunct` for correctness across scripts.
- The initial `utf8.DecodeRuneInString` in the tokenizer loop must discard the size with `_` to avoid an `ineffassign` lint error — the size is only needed in the inner loop where it advances `i`.

## E4.T8 — Schema: output type registry

- Registry tests that reset the package-level `envelopes` map must save and restore it (via `t.Cleanup`) to avoid wiping `init()`-registered entries visible to other tests in the same package. A `saveAndResetEnvelopes(t)` helper handles this.
- `RegisterEnvelope` is called in `init()` in `types.go` — same single-threaded init-time safety as `terr.New()` auto-registration.

## E4.T14 — Extract envelope type + language detection

- `ExtractEnvelope.MarshalJSON` follows the same nil→`[]` coercion pattern as `LookupEnvelope`: type alias to avoid recursion, nil check before delegation.
- `DetectLang` frontmatter scanner checks for both `---\n` and `---\r\n` prefixes, and uses `bytes.TrimRight(line, "\r")` inside the block to handle Windows line endings without an external YAML parser.
- Unclosed frontmatter (no closing `---`) is silently ignored — the function falls through to the flag and then to the `"en"` default, which is the correct behavior since malformed frontmatter shouldn't be treated as authoritative.

## E4.T9 — Schema: reflective walkers

- The walkers live in `internal/schema/` (not `internal/app/`) to avoid a circular dependency: `commands/` needs to import the walkers for the T10 schema command action, but `commands/` cannot import `app/` (since `app` imports `commands`).
- urfave's `Flag` interface does not expose `Usage` directly. A `flagUsage()` helper uses `reflect` to read the `Usage` field from the concrete flag struct types.
- `StringArg` (singular) always represents exactly one required positional (min=1, max=1). `StringArgs` (plural) has explicit `Min`/`Max` fields. The `describeArgument` function maps both to a uniform `ArgDesc{Name, Min, Max}`.
- `EnumerateErrors()` only returns sentinels visible via packages imported transitively. Unit tests in `schema_test` register their own test sentinels; integration tests in `app_test` verify real sentinels since `app_test` transitively imports all sentinel-declaring packages.

## E4.T12 — Extract: foreign-script tokens heuristic

- `dominantScript()` uses `unicode.In(r, unicode.Common, unicode.Inherited)` to skip Common/Inherited runes before counting per-script runes. Tokens consisting entirely of Common characters (digits, punctuation) return `nil` and are naturally excluded from foreign-script candidates.
- `scriptByName()` maps the CLI `--script` picklist values (`latin`, `hebrew`, `cyrillic`, `arabic`) to `*unicode.RangeTable` pointers. Pointer comparison (`ws != scriptFilter`) works because `unicode.Latin` etc. are package-level singletons — no need for string-based script name tracking on candidates.
- The `Options` type is defined in `foreign.go` with a `Script` field. It will be extended by T13 (high-frequency heuristic) with `MinFreq` and stopwords fields.

## E4.T10 — Schema command action + --command filter

- The schema command does not require `--tbx` since it reflects on the live binary's command tree, output types, and error sentinels — no TBX file is involved.
- `findCommand` needs to be recursive to support `--command` filtering for nested subcommands (e.g. `concept add`), not just top-level commands.
- `schemaFiltered` includes all `errorCodes` (not just command-specific ones) because error sentinels are registered globally and there's no per-command mapping in v1.
- Two separate envelope types (`schemaFullEnvelope` and `schemaFilteredEnvelope`) are cleaner than one overloaded type, since the full output has `commands`/`envelopes` (plural, arrays/maps) while the filtered output has `name`/`flags`/`envelope` (singular, for one command).
- `cmd.Root()` inside `schemaAction` gives the live command tree for `schema.WalkCommands`, so the schema always reflects the actual binary state — no drift possible.

## E4.T13 — Extract: high-frequency tokens + stoplist

- `Options` struct in `foreign.go` was extended with `MinFreq int` and `Stopwords map[string]bool` fields as planned in T12. `HighFrequencyTokens` accepts the shared `Options` type, not a separate `FrequencyOptions` type.
- `cases.Fold()` from `golang.org/x/text/cases` is used for case-folding (same as T6 lookup), combined with `norm.NFC` for normalization. This ensures "Temple", "temple", and "TEMPLE" aggregate to the same key.
- `LoadStopwords` NFC-normalizes and case-folds each line on load, so the stopwords map uses the same key format as `HighFrequencyTokens`'s aggregation — no normalization needed at lookup time.
- `MinFreq` defaults to 3 when zero or negative, matching the CLI's `--min-freq` default. The check is `opts.MinFreq <= 0` rather than `== 0` to be defensive against negative values.
- `defer func() { _ = f.Close() }()` is the project convention for read-only file handles (established in E1.T2 and used throughout `tbx/io.go` and test files).

## E4.T15 — Extract command action

- urfave v3 `StringArgs` (plural) values are accessed via `cmd.StringArgs("name")` which returns `[]string`. This is distinct from `cmd.StringArg("name")` (singular) which returns a single `string`.
- The `--script` flag filters ALL candidate results by script, not just foreign-script heuristic results. `termMatchesScript` checks if any rune in the candidate term belongs to the target `*unicode.RangeTable`, so Hebrew terms pass `--script hebrew` regardless of which heuristic produced them.
- Exclusion set for `--exclude` uses case-folded + NFC-normalized keys (same normalization as `HighFrequencyTokens`) so that glossary terms like "Tzimtzum" match candidates "tzimtzum" regardless of case.
- `terr.Newf` (non-registering) is used for dynamic I/O errors with interpolated file paths, not `terr.New` (which auto-registers into the sentinel registry and shouldn't have format args without values).
- Old stub tests (`TestExtract_Stub_ExitCode75`, `TestExtract_ValidScript_Stub`, `TestExtract_Stub_Golden`) must be updated when replacing a stub with a real action — they assert exit code 75 which no longer applies.

## E4.T16 — Golden CLI tests for lookup, schema, extract

- Golden test subdirectories (e.g. `extract/basic`, `schema/full`) coexist with root-level golden files (e.g. `extract/clean.*`) in the same parent directory. `runGolden` maps the `name` parameter directly to a subdirectory path under `testdata/`, so `extract/basic` and `extract` are independent — no conflict.
- Schema full golden output is large (~100+ KB) because it serializes the entire CLI surface reflectively (all commands, flags, envelopes, error codes). This is expected and deterministic; the golden file catches any unintended surface drift.
- Stale stub golden files (e.g. `testdata/schema/clean.*` with exit code 75) must be removed when replacing a stub command with a real implementation and adding new golden tests under subdirectories, otherwise the orphan files accumulate with no test referencing them.
- The `extractNoCandidatesError` type follows the same pattern as `lookupNotFoundError`: a private struct implementing `terr.Coded` with `ExitCode() int { return 1 }`. The envelope is emitted to stdout before the error is returned.

## E4.BUG — exit_codes missing from per-command schema view

- Per-command exit codes use a registry pattern parallel to the envelope registry: `RegisterExitCodes` / `ExitCodesFor` / `AllExitCodes` in `output/registry.go`. Registration happens in `output/types.go` `init()`, same single-threaded init-time safety.
- `ExitCodesFor` returns a defensive copy (same pattern as `AllEnvelopes`), preventing callers from mutating the registry.
- Exit codes are declared statically per command rather than derived from error sentinels, because not all exit codes come from `terr` sentinels (e.g. exit 1 from `warningsError`, `lookupNotFoundError`, `extractNoCandidatesError` are private types not in the registry), and exit 0 (success) has no sentinel at all.

## E4.BUG — Foreign-script heuristic ignores frontmatter lang for base script

- `ForeignScriptTokens` was computing `dominantScript` per-span, which meant a Latin-heavy span in a Hebrew document treated Latin as the base and Hebrew tokens as foreign. The fix adds `BaseLang` to `Options`; when set, `langToScript(BaseLang)` provides a document-wide base script that overrides per-span detection.
- `langToScript` maps BCP 47 tags to `*unicode.RangeTable` (he→Hebrew, ar→Arabic, ru→Cyrillic, el→Greek, zh/ja→Han, default→Latin). Pointer comparison with `dominantScript` results works because `unicode.Hebrew` etc. are package-level singletons.
- The extract command action creates a per-file copy of `Options` (`fileOpts := opts; fileOpts.BaseLang = fs.lang`) so different files in a multi-file invocation can have different base scripts.
- Integration test for this bug checks that `foreign_script`-heuristic Latin tokens like "concept" (lowercase, not caught by `capitalized_phrase`) are flagged, while Hebrew tokens are not. Capitalized Latin tokens like "Kabbalah" may be merged under `capitalized_phrase` heuristic due to `mergeCandidate` keeping the first heuristic label.

## E5.T3 — Aho-Corasick dependency + automaton build

- `github.com/cloudflare/ahocorasick` `Matcher.Match()` returns only pattern IDs (which patterns matched), not match positions. The `match()` internal function iterates byte-by-byte tracking trie state but the position information is not exposed. `Search` uses AC for O(N) pattern detection, then `bytes.Index` loops for position finding on matched patterns only.
- `ahocorasick.NewMatcher` panics on nil/empty input. `buildAutomaton` guards with an early return for empty patterns, and `Search` checks `len(a.patterns) == 0`.
- AC `Match` uses an internal counter for deduplication across a single call, incrementing `m.counter` each time. Multiple `Search` calls on the same automaton work correctly because each `Match` call increments the counter.
- Results are sorted by start position (ascending), with longest match first at same start, to provide deterministic output for downstream consumers (T5 longest-match filter).

## E5.T2 — Canonical normalization + offset map

- `norm.Iter` does not expose how many source bytes each `Next()` call consumed. The workaround uses `norm.Form.Properties(src[pos:]).BoundaryBefore()` to manually find normalization segment boundaries (starter + following non-starters), then normalizes each segment independently. This is correct because NFC/NFD segment boundaries are context-free.
- For the offset map, all bytes of a normalized segment map to the segment's start position in the source. For single-rune segments (the common case), this gives per-rune precision. For multi-rune composed segments (e.g., `e` + combining acute → `é`), both output bytes map to the base character's offset.
- `FoldDiacritics` requires NFD decomposition (not NFC) as the initial normalization step, so that composed characters like `é` (single codepoint) are decomposed into base + combining mark before the combining mark can be stripped. The `Normalize` field in `Policy` is overridden to NFD when `FoldDiacritics` is true.
- `cases.Fold().Bytes()` is applied per-rune (not to the whole string) to maintain correct offset tracking. Case folding can expand one rune to multiple (e.g., `ß` → `ss`); all output bytes map to the source offset of the original rune.

## E5.T4 — Word-boundary check

- `utf8.DecodeLastRune(orig[:start])` is the correct way to check the rune immediately before the match start — not `orig[start-1]`, which would only work for single-byte runes. Similarly, `utf8.DecodeRune(orig[end:])` handles multi-byte runes at the end boundary.
- `unicode.Is(unicode.Letter, r)` and `unicode.Is(unicode.Number, r)` correspond to `\p{L}` and `\p{N}` respectively, providing correct cross-script word-boundary detection (Latin, Hebrew, Arabic, etc.).

## E5.T5 — Longest-match-at-same-start filter

- `longestMatchPerStart` exploits the pre-sorted order from `automaton.Search()` (start ascending, longest first at same start). The filter reduces to a simple deduplication by `Start` position — keeping the first match seen per start group — rather than requiring explicit grouping and length comparison.

## E5.T6 — Matcher API (New + Scan)

- `New(g, lang, policy)` follows the spec's 3-parameter signature (explicit `Policy`), not the ticket's 2-parameter version. The spec is authoritative per CLAUDE.md.
- Pattern compilation and text normalization must use the **same** `Policy` for AC matching to work correctly. Using `PolicyFor(lang)` per-language for patterns but a different policy for text would produce mismatches. The caller-supplied `policy` is used for both.
- `origEnd` calculation from canonical match positions requires `utf8.DecodeRune(spBytes[lastSrcOff:])` to advance past the full rune at the mapped source offset, not just `Map[rm.End-1] + 1` which only works for single-byte runes. Multi-byte characters (Hebrew, accented Latin) would produce incorrect end offsets otherwise.
- `offsetToLineCol` iterates bytes (not runes) to track newlines, since byte offset is what the canonical map provides. Column counting is byte-based within the span, offset from the span's starting `Col`.

## E5.T7 — Scan envelope type

- Adding a new envelope registration in `output/types.go` `init()` changes the reflective schema output, which breaks `TestSchema_Full_Golden`. The golden file must be regenerated with `go test ./internal/app/ -run TestSchema_Full_Golden -update`.
- `ScanEnvelope.MarshalJSON` follows the same type-alias nil→`[]` coercion pattern as `LookupEnvelope` and `ExtractEnvelope`. The `Status` field on `ScanMatch` is a string (not an enum) because it maps to the human-readable status label in the scan output, not the raw TBX picklist value.

## E5.T8 — Scan command action

- `match.PolicyFor(lang)` is used for both pattern compilation and text normalization, ensuring the same canonical form is used for AC matching. When `--lang` is empty, `PolicyFor("")` returns `Baseline`.
- `cmd.StringArg("file")` (singular `StringArg`) retrieves the positional argument value, consistent with the pattern established in T7 for lookup's `cmd.StringArg("term")`.
- `stub_test.go` `TestUnderConstruction_ReturnsCoded` was using `scan` as its example stub command. After implementing scan, updated to use `check` (still a stub) — same pattern as the T7 update from `lookup` to `scan`.
- Stale golden files (`testdata/scan/clean.*` with exit code 75) must be removed when replacing a stub with a real action, otherwise orphan files accumulate. Same lesson as E4.T16.
- Scan always exits 0 for successful scans (informational, per spec). I/O errors (file not found) exit 3 via `terr.Newf`. No TBX path exits 2 via `tbx.ErrNoTBXPath`.

## E5.T9 — Golden CLI tests for scan

- The `--context` flag was defined on the scan command but never read in `scanAction`; `extractContext` hardcoded `const window = 40`. Fixed by adding a `contextSize int` parameter to `Matcher.Scan()` (0→80 default) and threading `cmd.Int("context")` from `scanAction`. `extractContext` now computes `window = contextSize / 2`. The golden test for `context_window` was regenerated to match the corrected behavior.
- Matcher-specific golden tests (code_blocks_skipped, multi_word, status_tagging) use small focused fixtures isolating a single behavior, rather than sharing the main corpus with `found`. This avoids duplicate golden files and makes each test's assertion clear from the fixture alone.
- `invalid_field` on scan exits 2 (usage error), same as lookup's `invalid_field` — the `ValidateFields` error wraps `ErrInvalidField` which `ExitCodeFor` maps to exit 2.

## E5.BUG — Niqqud matching fails without --lang filter

- When `lang=""`, `PolicyFor("")` returns `Baseline` with `StripNiqqud: false`. Hebrew glossary terms with niqqud (e.g. `סְפִירָה`) normalize to different canonical bytes than plain Hebrew text (`ספירה`), so the AC automaton can't match them.
- Fix: `New()` normalizes each pattern using `PolicyFor(ltag)` (the language-specific policy) instead of the caller's single policy. A `mergePolicy()` function ORs boolean fields across all language policies to produce a merged policy stored on the Matcher for text normalization.
- `mergePolicy` is safe because niqqud stripping on Latin text is a no-op (no niqqud runes), and case folding on Hebrew text is a no-op (Hebrew has no case). The merged policy is strictly additive.

## E5.BUG — Multi-word terms not matched across line breaks

- Goldmark produces separate `ast.Text` nodes for each line within a paragraph (soft line breaks). The gap between consecutive sibling Text nodes is exactly the `\n` byte in the source. Using `src[first.Segment.Start:last.Segment.Stop]` to build the merged span preserves 1:1 byte offset mapping to the original source.
- `ast.Text.SoftLineBreak()` returns true only for soft line breaks, not hard line breaks (two trailing spaces + newline). Hard line breaks produce a separate `ast.Hardlinebreak` node, so the merging logic naturally excludes them.
- The merge only applies to consecutive sibling Text nodes within the same parent. Text nodes separated by inline markup (emphasis, links) remain in separate spans, preserving accurate position reporting for those cases.
- The extract package (capitalized, foreign-script, high-frequency heuristics) computed positions as `span.Col + token.offset` without accounting for newlines in the span text. A `Span.lineColAt(byteOffset)` helper was added — same pattern as `offsetToLineCol` in `match/match.go` — to walk through the text counting newlines for correct line/col.
- The matcher's existing `offsetToLineCol` already handled newlines in span text correctly, so no changes were needed in `match/match.go`.

## E6.T1 — Frontmatter language extraction (shared)

- `gopkg.in/yaml.v3` replaces the hand-rolled frontmatter parser. It handles quoted values (`"pt-BR"`, `'zh-TW'`), inline comments (`lang: he # Hebrew`), and other YAML edge cases that the line-scanning approach silently broke on.
- The `\n---` closing delimiter search works for both `\n` and `\r\n` line endings because `\n---` appears at the start of the closing line regardless — the `\r` from Windows endings sits before the `\n`.
- `extract.DetectLang` keeps its "default to `en`" fallback, which is appropriate for corpus analysis. The E6 spec requires `scan`/`check` to fail with `ErrLanguageRequired` instead — that distinction lives in the command action, not in the shared `FrontmatterLang` function.

## E6.T2 — ErrLanguageRequired sentinel

- Adding a new `terr.New` sentinel changes the `terr.All()` registry, which in turn changes the `terminology schema` output. Both the full schema golden file and the command-filter golden file must be regenerated (`-update` flag) after any sentinel addition.
- `terr.E.Wrap` returns `*terr.E` (a concrete pointer, not an interface), so type assertions like `wrapped.(terr.Coded)` don't compile — use the concrete method calls directly (`wrapped.Code()`, `wrapped.ExitCode()`) and cast to `error` for `errors.Unwrap`.

## E6.T3 — Check envelope + violation types

- `CheckEnvelope.MarshalJSON` follows the same type-alias nil→`[]` coercion pattern as `LookupEnvelope`, `ExtractEnvelope`, and `ScanEnvelope`, applied to both `Violations` and `Warnings` slices.
- `CheckViolation` uses `omitempty` on all type-specific fields (`source_term`, `expected_target`, `source_occurrences` for `missing`; `variant`, `line`, `column`, `context` for `forbidden_variant`/`admitted_variant`). This means a `missing` violation naturally omits positional fields, and a `forbidden_variant` naturally omits source fields — no custom marshaler needed.
- Adding a new envelope registration in `init()` changes the reflective schema output, requiring regeneration of the schema full golden file (`go test ./internal/app/ -run TestSchema_Full_Golden -update`). Same pattern as E5.T7 and E6.T2.

## E6.T5 — Check --strict + admitted_variant

- `CheckWarning` was extended with `Variant`, `Line`, `Column`, `Context` fields (all `omitempty`) to carry positional data for `admitted_variant` warnings. The existing `Message` field remains for non-positional warnings.
- Admitted target matches are collected in a separate `admittedHits` slice alongside `forbiddenHits`, using the same `posHit` struct (renamed from `forbiddenHit` to be reusable).
- Admitted variants do NOT satisfy the preferred-term presence check — this is independent. A concept with only admitted variants in TGT produces both a `missing` violation and an `admitted_variant` warning (or violation under strict).
- Adding fields to `CheckWarning` changes the reflective schema output. The schema full golden file must be regenerated with `go test ./internal/app/ -run TestSchema_Full_Golden -update`. Same pattern as E5.T7, E6.T2, E6.T3.

## E6.T4 — Check algorithm: missing + forbidden_variant

- Source matches filter out `deprecated` and `superseded` status terms — only `preferred`, `admitted`, and `unspecified` source occurrences trigger the check for a concept. This matches the spec: "scan SRC for source-language preferred + admitted terms."
- `preferredTarget` falls back to the first term in the target language section when no term has `StatusPreferred`, consistent with the TBX semantics documented in cli-design.md: "If none is marked, the first `<termSec>` is treated as preferred."
- Violation sorting: positional violations (`forbidden_variant`) sorted by `(line, column)` come first; `missing` violations (no position in TGT) sort to the end grouped by `concept_id` ASCII order. Uses `sort.SliceStable` to preserve insertion order as tiebreak.
- `ConceptsChecked` counts only concepts that have terms in both source and target language sections AND were found in the source text. Concepts without a target language section are silently skipped (no violation, no count).

## E6.T6 — Violation ordering

- The existing `sortViolations` was missing the `concept_id` tiebreak for same `(line, column)` positions. Without it, `sort.SliceStable` preserved insertion order, which is non-deterministic when violations come from map iteration or concurrent matching.
- Warnings use the same sort rules as violations but need a different "positional" predicate: violations use `Type != "missing"`, warnings use `Line != 0 || Column != 0` since warning types don't have a "missing" category — non-positional warnings are identified by zero line/column.
- `sortWarnings` is called alongside `sortViolations` in `Check` to ensure deterministic output for golden-file stability.

## E6.T7 — Check command action

- `ErrLanguageRequired` was moved from `app/errors.go` to `commands/check.go` because `commands` cannot import `app` (import cycle: `app` → `commands` → `app`). The sentinel is exported as `commands.ErrLanguageRequired`. The `app/errors_test.go` tests were updated to import from `commands` — this works because `app_test` (external test package) can import `commands` without creating a cycle.
- `StringArgs` (plural, `Min: 2, Max: 2`) returns a `[]string` via `cmd.StringArgs("files")`. SRC is at index 0, TGT at index 1. This is different from `StringArg` (singular) which returns a single string.
- The `violationsError` type follows the same pattern as `warningsError` in `validate.go`: a private struct implementing `terr.Coded` with `ExitCode() int { return 1 }`. The envelope is emitted to stdout with the correct `ok` value before the error is returned, so exit 1 signals "violations present" without producing an error envelope on stderr.
- Language resolution precedence: frontmatter (`markdown.FrontmatterLang`) → CLI flag (`--source-lang`/`--target-lang`) → `ErrLanguageRequired`. Frontmatter wins because it is file-specific and the strongest signal, preventing misclassification when the user passes a flag that doesn't match the file's actual language.
- Moving `ErrLanguageRequired` changed its hint text (dropping `--lang` since check uses `--source-lang`/`--target-lang`, not `--lang`). This required regenerating both schema golden files (`TestSchema_Full_Golden` and `TestSchema_CommandFilter_Golden`) via `-update`.
- `stub_test.go` `TestUnderConstruction_ReturnsCoded` was using `check` as its example stub command. After implementing check, updated to use `concept add` (still a stub) — same pattern as prior E5/E4 updates.

## E6.T8 — Scan frontmatter language resolution

- The scan command's frontmatter resolution is simpler than check's: scan doesn't error on missing language (it scans all languages as fallback), so no `ErrLanguageRequired` is needed. The change is a 3-line addition before the existing `--lang` flag read.
- Existing scan test fixtures had no frontmatter, so all existing tests passed unchanged. New tests use a dedicated `scan-frontmatter-he.md` fixture with `lang: he` frontmatter to verify both detection and flag-override behavior.

## E6.T9 — Golden CLI tests for check

- The `check/clean` and `check/clean_frontmatter` tests are identical in argv because `check-source.md` and `check-target-clean.md` both have frontmatter — so the "frontmatter auto-detection" path is the default happy path. There's no separate "no-frontmatter clean" golden because that would require a different fixture set.
- The `admitted_warning` test (non-strict) still exits 1, not 0, because the `missing` violation for `malkhut` is present (admitted variants do NOT satisfy the preferred-term presence check). This is consistent with E6.T5 learnings.
- The `check-target-forbidden.md` fixture has `lang: en` (same as source), making the check operate en→en. No `--source-lang`/`--target-lang` flags are needed since both files have frontmatter.
- For `violation_ordering`, the `check-target-multi-violations.md` fixture has `malchut` (deprecated) twice at different positions and is missing both `malkhut` and `tzimtzum` preferred terms, producing 4 violations: 2 positional (forbidden_variant) sorted by line, then 2 non-positional (missing) sorted by concept_id alphabetically.

## E6.T10 — Manual QA plan review

- Hebrew test fixtures must use glossary terms as standalone words — the definite article prefix `ה` (hey) causes the matcher's `\p{L}` word-boundary check to reject the match. E.g. `הצמצום` does NOT match glossary term `צמצום`. The golden test fixtures in `testdata/fixtures/` avoid this by using en→en checks or placing Hebrew terms after whitespace/punctuation boundaries.
- When testing admitted variants (non-strict), if only the admitted variant appears in the target (without the preferred term), both a `missing` violation AND an `admitted_variant` warning are produced (`ok: false`, exit 1). To test the pure "admitted as warning" path, the fixture must include BOTH the preferred and admitted terms.
- `TERMINOLOGY_TBX` env var resolution should be tested for `check` (not just `scan`), since both commands use the global `--tbx` resolution logic.

## E7.T4 — Concept-ID derivation (internal/write/id.go)

- `cases.Fold().Bytes()` handles eszett (`ß` → `ss`) and other case expansions correctly. Combined with NFKD + `unicode.Mn` mark-dropping, it produces clean ASCII slugs from accented Latin input.
- Truncation at 64 codepoints uses `strings.LastIndex(result, "-")` to cut at the last hyphen boundary. If the truncated string ends with a hyphen, `strings.TrimRight` removes trailing hyphens first. If no hyphen exists in the first 64 runes, the full 64-rune string is kept.
- The `internal/write` package is new and not yet imported transitively by `internal/app`, so the `ErrInvalidID` sentinel (registered via `terr.New()`) won't appear in `terminology schema` output until a command action imports the `write` package (expected in T6/T7). The schema golden files don't need updating until then.
- `regexp.MustCompile(`[^a-z0-9]+`)` is compiled at package level (var) for the hyphen-replacement step. This is safe because the regex is stateless and the function is pure.

## E7.T2 — Write error sentinels

- `invalid_picklist` is NOT a write sentinel — it's handled by urfave's `pickFlag` validator at the CLI layer (exit 2), not in `internal/write`. The spec lists it alongside write error codes but it's already covered and doesn't need a sentinel in `errors.go`.
- New sentinels (`ErrDuplicateID`, `ErrNotFound`, `ErrDanglingCrossref`, `ErrInvalidInput`) won't appear in `terminology schema` output until a command action in `internal/app/commands` imports `internal/write`, triggering `init()`-time registration. This is the same deferred-visibility pattern noted in E7.T4 for `ErrInvalidID`. The schema golden file does not need updating until then.

## E7.T1 — Clock injection (internal/clock)

- The `internal/clock` package mirrors the `internal/logctx` context-injection pattern exactly: `Clock` interface, `With`/`From` on `context.Context`, package-level `Real` variable as default.
- `realClock.Now()` returns `time.Now().UTC()` so callers cannot accidentally emit host-local time — same rationale as the determinism ADR §5.
- The `fakeClock` test double is trivial (struct with a `T time.Time` field) and lives in the test file. Downstream packages (e.g. `write/transaction.go`) will use the same pattern: `clock.With(ctx, fakeClock{T: testTime})`.
- No package-level `var now = time.Now` anti-pattern — rejected for the same reasons as in `logctx` (untestable global state).

## E7.T3 — Write envelope types (output/types.go)

- `WriteResult` is richer than `LookupResult` — it carries the full concept shape (definitions, cross-refs, sources, notes at concept level, plus all TBX data categories at term level) to support read→modify→write round-trip. The `LookupResult` remains the lean projection for read commands.
- `WriteEnvelope.MarshalJSON` coerces nil `Languages` map to `make(map[string]WriteTermGroup)` so it serializes as `{}` rather than `null`. This follows the same type-alias pattern as other envelopes but operates on a map instead of a slice.
- `WriteTerm` includes `Reading`, `ReadingNote`, `Contexts`, and `CrossRefs` fields that mirror `tbx.Term` — these are Linguist/Basic module data categories that the lean `LookupTerm` omits.
- `WriteCrossRef` is a separate type from `tbx.CrossRef` to keep the output package decoupled from the TBX domain model.
- Adding new envelope registrations in `init()` changes the reflective schema output, requiring regeneration of schema golden files (`go test ./internal/app/ -run TestSchema_Full_Golden -update` and `TestSchema_CommandFilter_Golden -update`). Same pattern as E5.T7, E6.T2, E6.T3, E6.T5.
- Exit codes for write commands were already registered in `types.go` `init()` from a prior ticket, so no exit code changes were needed.

## E7.T5 — Transaction record builder

- `NewTransaction(ctx, author)` follows the context-injection patterns from `clock` and `logctx`: `clock.Now(ctx)` for deterministic timestamps, `logctx.From(ctx)` for structured logging. No package-level state needed.
- `time.RFC3339` produces seconds-precision UTC timestamps (e.g. `2025-03-15T10:30:00Z`) when the input `time.Time` has zero nanoseconds, which `clock.Real` guarantees via `.UTC()`. No custom format string needed.
- Testing the WARN log for missing author uses a `bytes.Buffer`-backed `slog.TextHandler` attached to context via `logctx.With`. This avoids global logger mutation and keeps tests parallel-safe.

## E7.T6 — Write pipeline (internal/write/write.go)

- `Glossary.Validate(false)` treats `duplicate_id`, `unresolved_crossref`, `missing_term`, and `invalid_lang_tag` as warnings (not errors). The write pipeline must promote these to fatal errors via `fatalWarningCodes` map, since a write that introduces these is invalid even though the validator only warns.
- `ParseTBXFragment` reuses the existing `LinguistReader.Decode` by wrapping fragments in a synthetic TBX shell. This avoids duplicating any parsing logic. Required exporting `tbx.ReaderForDialect()` as a thin wrapper around the internal `readerFor()`.
- `extractListInner` for `<conceptEntryList>` uses `xml.Decoder.InputOffset()` before each token + `dec.Skip()` to capture raw byte ranges for each `<conceptEntry>` child. The raw bytes are concatenated and inserted into the TBX shell, preserving the original XML exactly.
- `json.Decoder.DisallowUnknownFields()` rejects unknown JSON keys at decode time, producing `invalid_input` errors without needing a separate schema validator. The `output.WriteResult` struct IS the schema.
- `terr.E.Wrap(cause)` returns a new `*terr.E` that carries the original code and exit code but wraps the cause — so `ErrInvalidInput.Wrap(err)` produces an error where `coded.Code() == "invalid_input"` and `errors.Unwrap()` gives the JSON/XML parse error. This pattern is used throughout both `ParseJSONInput` and `ParseTBXFragment`.

## E7.T7 — concept add command action

- Stdin detection uses `os.Stdin.Stat()` with `os.ModeCharDevice` check — if stdin is a pipe (not a TTY), the data is read; otherwise flags are used. This gives clean auto-detection without a `--format` flag for input type.
- XML vs JSON auto-detection on stdin: skip whitespace, then check first non-whitespace byte for `<` (XML) vs anything else (JSON). The `looksLikeXML` function handles BOM-stripped data.
- `writeResultToConcept` converts `output.WriteResult` → `tbx.Concept` for JSON stdin input. This is the inverse of `buildWriteResult`, enabling the read→modify→write round-trip described in the spec.
- `ParseStatus(string) Status` and `Status.String()` were added to `tbx/model.go` as public API. The `linguist` package had its own unexported `normalizeStatus`/`statusString` — the new public methods avoid coupling command code to the internal linguist reader/writer packages.
- `buildWriteResult` and `tbxTermToWriteTerm` live in `write_helpers.go` (shared across all write commands). `StatusUnspecified` terms with no preferred term in a `LangSection` are placed in the `Preferred` slot (first-term-is-preferred fallback), matching the TBX semantics documented in cli-design.md.
- Implementing the first real write command changes the `terr.All()` sentinel registry (write package sentinels become transitively imported), which changes `terminology schema` output. Both `TestSchema_Full_Golden` and `TestSchema_CommandFilter_Golden` need regeneration with `-update`.
- `stub_test.go` was using `concept add` as the example stub command. After implementing concept add, updated to `concept update` — same pattern as prior E5/E4/E6 updates.
- The `copyTBXFixture` test helper copies a TBX file to a temp directory before write tests, since the write pipeline mutates the file. This avoids polluting shared fixtures.

## E7.T8 — concept update command action

- `parseConceptUpdateInput` reuses `readStdinIfAvailable` and `parseConceptFromStdin` from concept_add.go. Only `parseUpdateFromFlags` is unique to update (no ID derivation needed since ID comes from the positional arg).
- `replaceConcept` saves the ID, does `*existing = *payload`, then restores the ID. Simple and correct for the ID stability contract.
- `mergeConcept` uses replace-if-present semantics for Definitions/CrossRefs/Sources/Notes (avoiding the "append forever" pitfall on repeated merges). Languages map overlays: payload keys merge, absent keys preserved.
- Term natural key matching uses `(Surface, AdministrativeStatus)` tuple. If a match is found, `mergeTermFields` overlays non-empty fields. If no match, the term is appended to the existing `LangSection.Terms` slice.
- Transaction records are appended inside the mutator (after merge/replace), not before — this ensures they attach to the already-mutated concept. For concept add, the transaction was appended before the mutator since the concept was being created.
- Adding picklist flags (`--status`, `--part-of-speech`, `--register`, `--grammatical-gender`) to the command definition changed the schema output. Both `TestSchema_Full_Golden` and `TestSchema_CommandFilter_Golden` needed regeneration with `-update`.
- `stub_test.go` was using `concept update` as its example stub command. After implementing concept update, updated to use `concept remove` (still a stub) — same pattern as prior updates.

## E7.T9 — concept remove command action

- `concept remove` cannot use `write.Execute` because `--force` needs to bypass the `unresolved_crossref` fatal warning in `validateForWrite`. A manual pipeline (load → check → remove → save) is used instead. This is safe because removing a concept cannot introduce `duplicate_id`, `missing_term`, or `invalid_lang_tag` issues — removal only reduces the set.
- `findCrossRefsTo` checks both concept-level `CrossRefs` and term-level `CrossRefs` across all other concepts. Both levels are specified in the TBX data model and must be checked for completeness.
- Transaction records are attached to the removed concept copy (for output envelope display only). Since the concept is deleted from the glossary, the transaction has no persistent home — this matches the spec's statement that "a removed entity has no persistent home."
- The `--force` flag deliberately skips the dangling crossref check AND skips `validateForWrite`. The integration test (`remove --force` → `validate`) verifies that `validate` surfaces the resulting `unresolved_crossref` warnings.
- `stub_test.go` was using `concept remove` as its example stub command. After implementing concept remove, updated to use `term add` (still a stub) — same pattern as prior updates.
- Stale golden files for `concept_remove/` must be removed when replacing a stub with a real implementation, otherwise orphan files accumulate with no test referencing them. Same lesson as E4.T16 and E5.T8.

## E7.T10 — term add command action

- `termAddAction` follows the same mutator pattern as `conceptUpdateAction`: find concept by ID, modify in-place, return pointer. Simpler than concept update because there is no merge/replace distinction — term add always appends.
- Transaction records are attached to the `Term` struct (termSec level), not the `Concept` (conceptEntry level). This matches the spec: "Transaction record placed at termSec level." The `t.Transactions = append(t.Transactions, txn)` is done inside the mutator before appending the term to the langSec.
- `LangSection` creation uses the same nil-map guard as `mergeConcept`: `if existing.Languages == nil { existing.Languages = make(map[string]tbx.LangSection) }`. The map value is fetched, modified, and written back (Go map values are not addressable).
- `stub_test.go` was using `term add` as its example stub command. After implementing term add, updated to use `term deprecate` (still a stub) — same pattern as prior updates.
- Stale golden files for `term_add/` must be removed when replacing a stub with a real implementation. Same lesson as E4.T16, E5.T8, and E7.T9.

## E7.T11 — term deprecate command action

- `termDeprecateAction` follows the same mutator + `write.Execute` pattern as `termAddAction`, but instead of appending a term, it finds an existing term by `(conceptID, lang, surface)` triple and sets `AdministrativeStatus = tbx.StatusDeprecated`.
- Three levels of `not_found` are needed: concept not found, langSec not found, term not found. Each returns `write.ErrNotFound.Wrap(fmt.Errorf(...))` with a descriptive message. This is more granular than `term add` which only checks for the concept.
- Unlike `term add`, `term deprecate` does NOT create a missing `LangSection` — the lang must already exist since we're modifying an existing term.
- Go map values are not addressable, so `ls.Terms[termIdx].AdministrativeStatus = ...` works because `ls` is a local copy fetched from the map; the modified `ls` is written back with `existing.Languages[lang] = ls`. Same pattern as `term add`.
- `stub_test.go` was using `term deprecate` as its example stub command. After implementing term deprecate, updated to use `apply` (the last remaining stub).
- Stale golden files for `term_deprecate/` must be removed when replacing a stub with a real implementation. Same lesson as E4.T16, E5.T8, E7.T9, and E7.T10.
- No schema golden file regeneration was needed because `term deprecate` envelope and exit codes were already registered in `output/types.go` `init()` from a prior ticket.

## E7.T12 — Golden CLI tests for write commands

- `runGoldenCtx(t, name, argv, ctx)` was added to `golden_test.go` as a variant of `runGolden` that accepts a custom `context.Context`. This enables fake clock injection for deterministic transaction timestamps via `clock.With(ctx, fakeClock{...})`. The original `runGolden` delegates to `runGoldenCtx` with `context.Background()`.
- Write command golden tests require `copyFixture(t, name)` to copy the TBX file to a temp directory before each test, since write operations mutate the file. This is the same `copyTBXFixture` pattern used in functional tests but renamed to avoid collision.
- `WriteResult` (the output envelope type) does not include transaction records — transactions are persisted to the TBX file but not reflected in the JSON stdout. The fake clock injection is still valuable for determinism if transactions ever surface in output, and it demonstrates the pattern for future use.
- `pipeStdin(t, data)` returns a cleanup function that restores `os.Stdin`. It uses `os.Pipe()` to inject data, matching the pattern established in `TestConceptUpdate_JSONStdin`. The deferred restore ensures test isolation.
- JSON stdin and TBX fragment golden tests for `concept add` verify the full stdin auto-detection path (`looksLikeXML` dispatch). The TBX fragment test uses raw `<conceptEntry>` XML with namespace-prefixed elements — no `<tbx>` shell needed since `ParseTBXFragment` wraps fragments automatically.
- ID derivation golden test confirms `Razón Histórica` → `razon-historica` (NFKD decompose, strip combining marks, case-fold, hyphen-slug). ID stability golden test confirms that `concept update --replace` with a different term preserves the original concept ID `tzimtzum`.
- `concept_remove/dangling_crossref` uses `crossref-dct.tbx` (which has sefirot→tzimtzum cross-reference). Removing tzimtzum without `--force` produces exit 65 with `dangling_crossref` error; with `--force` it succeeds (exit 0).

## E8.T1 — Apply envelope types and error sentinel

- `ApplyEnvelope.MarshalJSON` must coerce five nil slices: `Applied.Added`, `Applied.Updated`, `Applied.Removed`, `Applied.Unchanged`, and `Warnings`. The nested `ApplyResult` struct fields are accessed through the alias copy (`a.Applied.Added`), not through a separate receiver — same copy-and-mutate pattern as other envelopes.
- `ErrApplyValidationFailed` uses exit code 1 (recoverable validation failure), unlike other write sentinels which use exit 65. This matches the spec: apply validation failure is a recoverable condition (fix the payload and retry), not a data-format rejection.
- Adding `ErrApplyValidationFailed` to `write/errors.go` changes `terr.All()` (the sentinel is auto-registered via `terr.New()`), which changes the reflective `terminology schema` output. Both `TestSchema_Full_Golden` and `TestSchema_CommandFilter_Golden` needed regeneration with `-update`. Same pattern as E5.T7, E6.T2, E6.T3, E7.T7.
- `ApplyFailure` type (concept_id, code, message) is defined in `output/types.go` alongside the envelope but is not yet wired into the error emission path — that integration belongs to T4 (reconciliation algorithm). The type exists now so downstream tickets can reference it.
- Apply exit codes were already partially registered as `{0, 2, 3, 65}` from a prior ticket; T1 updated to `{0, 1, 2, 3, 65}` to include exit 1 for `ErrApplyValidationFailed`.

## E8.T2 — Concept equality (transacGrp-stripping canonical comparison)

- `WriterForDialect` was added to `tbx/registry.go` (parallel to `ReaderForDialect`) so the `write` package can obtain a writer without importing the `linguist` package directly (avoiding coupling and keeping the registry pattern).
- The regex `(?s) *<transacGrp>.*?</transacGrp>\n` safely strips transacGrp from canonical writer output because: (1) the canonical writer produces deterministic indented XML with one element per line, (2) transacGrp elements never nest other transacGrp elements, and (3) the `.*?` non-greedy match with `(?s)` dotall correctly handles multi-line transacGrp blocks.
- Wrapping each concept in a full Glossary+Encode is correct because both concepts get the identical TBX shell (fixed `SourceDesc`, same namespace declarations), so the shell bytes cancel out in comparison. No need to extract just the `<conceptEntry>` portion.
- The `_ "github.com/andreswebs/terminology/internal/tbx/linguist"` blank import in tests is required to trigger the linguist dialect registration via `init()`. Without it, `WriterForDialect(DialectLinguist)` returns `ErrUnsupportedDialect`.

## E8.T3 — Apply payload parsing (JSON + TBX + format detection)

- `WriteResultToConcept` and `WriteTermToTBXTerm` were moved from `commands/concept_add.go` to `write/apply.go` as exported functions. The `commands` package now calls `write.WriteResultToConcept(wr)` instead of an unexported local copy. This avoids duplication since both `concept add` (single-concept JSON stdin) and `apply` (multi-concept `{"concepts": [...]}` wrapper) need the same conversion logic.
- `ApplyPayload` wraps `[]output.WriteResult` in a `{"concepts": [...]}` JSON envelope. `json.Decoder.DisallowUnknownFields()` rejects unknown keys at both the wrapper and nested `WriteResult` level — no separate schema validator needed.
- Format detection uses extension first (`.json`→JSON, `.tbx`/`.xml`→TBX), then content sniffing (first non-whitespace byte: `<`→TBX, `{`/`[`→JSON). For stdin (`--file -`), content sniffing is the only path. The ticket spec diverges from the 008-apply.md spec (which says stdin with no `--format` should error) — the ticket's content-sniffing approach was implemented as specified in the ticket.
- `sniffFormat` in `write/apply.go` mirrors the `looksLikeXML` function in `commands/concept_add.go` but returns a typed `PayloadFormat` and structured errors instead of a bool. Both coexist: `looksLikeXML` handles concept add's stdin path, `sniffFormat` handles apply's file/stdin path.

## E8.T4 — Reconciliation algorithm

- `Reconcile` modifies the glossary in-place and returns a `ReconcileResult` with sorted ID lists. `ReconcileWithTxn` wraps it with transaction record injection for added/updated concepts.
- `findCrossRefsToInSlice` is a standalone helper parallel to `concept_remove.go`'s `findCrossRefsTo`. It checks the `remaining` slice (concepts that will survive pruning) for references to each removal candidate, catching both concept-level and term-level `CrossRefs`.
- `fakeClock` is already declared in `transaction_test.go` (same package `write`). Test files in the same package share type declarations, so `reconcile_test.go` must not redeclare it. The existing `fakeClock` uses a value receiver (`func (fc fakeClock) Now()`), so callers pass `fakeClock{...}` not `&fakeClock{...}`.
- Wholesale replace copies the payload concept over the existing one, then restores the original ID: `g.Concepts[idx] = pc; g.Concepts[idx].ID = existingID`. This preserves ID stability per the spec.
- The prune step builds a `remaining` slice before checking crossrefs, so that removal candidates are not included in the reference check as potential referencing sources — only surviving concepts can cause dangling refs.

## E8.T5 — Apply command action

- `acquireLock` was exported as `AcquireLock` (public wrapper) and `SaveLocked` was added to `tbx/io.go` to support the lock-spanning read-modify-write pattern. `SaveLocked` delegates to the shared `writeFile` helper, skipping lock acquisition since the caller already holds it. `Save` still acquires+releases its own lock for all other callers.
- Apply uses a manual pipeline (like `concept remove`) rather than `write.Execute`, because it handles multiple concepts, reconciliation results, and prune logic — none of which fit the single-concept `Mutator` pattern.
- `stub.go` and `stub_test.go` were removed since apply was the last command using `underConstruction`. The `terr.UnderConstruction` function remains available in the `terr` package for future use but is no longer referenced by any command action.
- Dangling crossref tests for `--prune` must include the crossReference in the payload concept, not omit it. If the payload concept omits the crossref, reconciliation replaces the concept with a version that lacks the ref, making the prune succeed — the ref is gone before the prune check runs.
- Reconcile only categorizes concepts that appear in the payload (added/updated/unchanged). Glossary concepts absent from the payload are silently left alone unless `--prune` is set — they do not appear in any result category.

## E8.T6 — Golden CLI tests for apply

- Golden tests for write commands live in `write_golden_test.go` (and now `apply_golden_test.go`), separate from `commands_test.go`. Write commands need `copyFixture` to copy TBX to a temp dir since writes mutate the file; read command golden tests reference fixtures directly.
- Apply golden tests use a `writePayloadFile` helper (writes JSON/TBX payload to temp dir and returns the path) since `apply` takes `--file PATH` rather than stdin. This differs from `concept add` which uses `pipeStdin` for stdin-based input.
- `ErrApplyValidationFailed` (exit 1) is defined in `write/errors.go` but not used by any code path in the current reconciliation implementation. The `validateForWrite` path produces `tbx.ErrValidationError` (exit 65) instead. The golden test for validation failure uses a dangling crossref in the payload to trigger `validation_error` (exit 65).
- Duplicate concept IDs in a single apply payload are processed sequentially (first as add, then as update), not rejected as a validation error. This is because `reconcile` updates its `currentIndex` map after each add, so the second occurrence finds the first and treats it as an update.
- The idempotent golden test requires a two-pass approach: first apply converges the file, second apply (the golden test) should show all concepts as unchanged. Without the first pass, the file's canonical form may differ from the payload's representation.

## E8.T8 — Remove under-construction scaffolding

- Output tests (`errors_test.go`, `conformance_test.go`) used `terr.UnderConstruction()` as a convenient fixture for testing `EmitError` with a `terr.Coded` error. Replaced with `terr.Newf()` (non-registering) to avoid polluting the sentinel registry while still producing a `Coded` error with known code/hint/message fields.
- No schema golden file regeneration was needed because `UnderConstruction()` was non-registering (it constructed `*terr.E` directly, bypassing `terr.New()`), so removing it did not change `terr.All()` output.

## E8.BUG — Crossref validator single-pass ordering breaks idempotency after sorted write

- `Glossary.Validate()` used a single-pass approach: `idSeen` was populated left-to-right as concepts were iterated, and crossref targets were checked against only the IDs seen so far. When the canonical writer sorts concepts alphabetically, any crossref whose source sorts before its target produced a false-positive `unresolved_crossref` warning.
- Fix: split into two passes. Pass 1 collects all concept IDs (and checks `duplicate_id`). Pass 2 validates lang tags, missing terms, and crossrefs against the complete ID set. This makes crossref validation order-independent.
- The warning order in golden files changed: `duplicate_id` warnings now appear before `unresolved_crossref` warnings (collected in pass 1 vs pass 2). The `validate/warnings` golden file was regenerated with `-update`.

## E8.BUG — ErrApplyValidationFailed sentinel declared but never returned

- `validateForWrite` fails on the first fatal warning, returning `tbx.ErrValidationError` (exit 65). For apply, the spec requires collecting ALL per-concept failures into a `details.failures[]` array with `apply_validation_failed` code (exit 1). A separate `validateForApply` function was added that collects all fatal warnings before returning.
- `ApplyValidationError` is a concrete type (not a `terr.E` wrapper) because it needs to implement both `terr.Coded` (for exit code/code/hint) and the new `output.Detailed` interface (for `ErrorDetails() any`). `terr.E.Wrap()` returns `*terr.E` which can't carry arbitrary detail payloads.
- The `output.Detailed` interface (`ErrorDetails() any`) is checked via `errors.As` in `EmitError`, and the returned value is set on `errorDetail.Details` (a new `any` field with `omitempty`). Non-detailed errors produce identical output to before (no `details` key in JSON).
- `validateForApply` still returns `tbx.ErrValidationError` (exit 65) for hard `res.Errors` (unparseable structure), since those are not per-concept recoverable failures. Only fatal warnings (the `fatalWarningCodes` map) produce `ApplyValidationError`.
- `validateForWrite` is preserved unchanged for single-concept write commands (`concept add/update/remove`, `term add/deprecate`) which don't need per-concept failure aggregation.

## E8.BUG — Nonexistent payload file returns exit 1 instead of exit 3

- `readPayloadFile` wraps `os.ReadFile` errors with `fmt.Errorf("reading payload file: %w", err)`, which preserves the `fs.ErrNotExist` cause in the error chain. `applyAction` checks `errors.Is(err, fs.ErrNotExist)` on the `LoadApplyFile` error and returns `terr.Newf("io_error", 3, ...)` for exit 3.
- This is consistent with the I/O error handling pattern in `scan`, `check`, and `extract` command actions, which all use `terr.Newf("io_error", 3, ...)` for file-not-found errors at the command layer.

## E9.T1 — Input sanitization functions (sanitize.go)

- `ErrInvalidSanitizeID` is a separate sentinel from `write.ErrInvalidID` because they serve different purposes: the sanitizer rejects user-supplied IDs with dangerous characters (control chars, path traversal, percent-encoding, query params), while the write package's sentinel handles empty derived IDs from non-Latin terms. Both use code `invalid_id` and exit 65 — two sentinels with the same code are allowed in the registry.
- BCP 47 validation uses a lightweight structural check (2–8 alpha primary tag, 1–8 alphanumeric subtags separated by `-`) rather than `golang.org/x/text/language.Parse`, keeping the `commands` package free of that dependency. The `tbx.Validate()` function already uses `language.Parse` for full semantic validation; the sanitizer only needs to reject injections and obviously malformed tags.
- `sanitizePath` uses `filepath.EvalSymlinks` to catch symlink escapes. If the target doesn't exist yet, `EvalSymlinks` returns `os.ErrNotExist` — in that case, the pre-symlink absolute path is returned (write targets may not exist yet). The prefix check runs both before and after symlink resolution.
- staticcheck `QF1001` (De Morgan's law) flags `!((a || b))` as improvable. The fix `(not-a && not-b)` reads identically but satisfies the linter. This applies to BCP 47 character-class checks.
- Adding new `terr.New()` sentinels in `sanitize.go` changes `terr.All()` output, which changes the reflective `terminology schema` output. Both `TestSchema_Full_Golden` and `TestSchema_CommandFilter_Golden` must be regenerated with `-update`. Same pattern as E5.T7, E6.T2, E7.T7, E8.T1.

## E9.T2 — Wire sanitizers into command action handlers

- `tbxPathFromRoot(cmd)` helper combines `cmd.Root().String("tbx")` retrieval, empty check (returns `tbx.ErrNoTBXPath`), and `sanitizeTBXPath()` call into one function. All commands that use `--tbx` call this instead of the 3-line inline pattern.
- `sanitizeTBXPath` performs the same string-level checks as `sanitizePath` (control chars, `..`, `%`, `?#`) but omits `resolveAndSandbox()` — the `--tbx` flag is exempt from CWD sandboxing per spec, since agents legitimately pass absolute paths to shared glossaries.
- `sanitizePath` converts relative paths to absolute via `resolveAndSandbox`. Commands that display paths in output envelopes (scan's `file`, check's `source`/`target`, extract's location `file`) must keep the original user-supplied path for display and use the sanitized absolute path only for I/O. This avoids changing output format while still validating paths.
- Apply's `--file` is sandboxed to CWD (per spec "Output paths and `--file` are sandboxed"). Test helpers (`writePayloadFile`, `writePayloadInCWD`) create temp dirs under `testdata/` (within CWD) using `os.MkdirTemp("testdata", "payload-*")` with `t.Cleanup` for removal. Using `t.TempDir()` would create under `/tmp/` which fails the sandbox check.
- Apply post-parse payload validation iterates `payload` concepts to sanitize concept IDs and lang tags after `write.LoadApplyFile()`. This catches injection in JSON/TBX payloads, not just CLI flags.
- `os.ReadFile(absolutePath)` includes the absolute path in its error message, so golden files for file-not-found scenarios show the resolved absolute path in the inner error even when the display path is relative. This is correct behavior — the user-facing message prefix uses the display path, and the wrapped OS error naturally includes the actual path attempted.
- `_ = os.RemoveAll(dir)` pattern is required in test cleanup functions to satisfy `errcheck` lint. Same `_, _ =` convention as established in E1.T2.

## E9.T3 — XML parser hardening (DOCTYPE policy, nesting depth)

- `CheckDoctype` scans raw bytes (before XML decoding) using `bytes.ToUpper` + linear scan for `<!DOCTYPE` token. After the DOCTYPE keyword, the scanner looks for `>` (bare — accepted), `[` (internal subset — rejected), or `SYSTEM`/`PUBLIC` keywords (external IDs — rejected). This avoids regex and is O(N) in input size.
- `CheckDoctype` is wired into both `tbx.Load()` (pre-`detectDialect`) and `linguist.Reader.Decode()`. The double-check on the `Load→Decode` path is harmless since `CheckDoctype` is idempotent and cheap.
- `tbx.Load()` was refactored to `io.ReadAll` the file once and pass `bytes.NewReader(data)` to both `detectDialect` and `Decode`, eliminating the `f.Seek(0, io.SeekStart)` pattern. This was required because `CheckDoctype` needs the raw bytes before any decoder sees the data.
- Nesting depth tracking uses a `depth int` field on `decodeCtx` and a `dc.token()` wrapper method around `xml.Decoder.Token()`. All token reads throughout the reader — including `readCharData` and `readNoteText` — go through `dc.token()`. `readCharData` and `readNoteText` were changed from `*xml.Decoder` to `*decodeCtx` parameter to participate in depth tracking.
- `dc.skip()` replaces `dc.dec.Skip()` to maintain correct depth tracking. `xml.Decoder.Skip()` consumes tokens without going through the depth wrapper, causing the counter to drift. The custom `skip()` manually iterates tokens tracking a local depth counter until the matching end element.
- Adding `ErrDangerousDoctype` and `ErrNestingTooDeep` sentinels via `terr.New()` changes `terr.All()` output, which changes the reflective `terminology schema` output. Both `TestSchema_Full_Golden` and `TestSchema_CommandFilter_Golden` must be regenerated with `-update`. Same pattern as E5.T7, E6.T2, E7.T7, E8.T1, E9.T1.
- `xml.Decoder.Strict = true` was set on all 4 decode paths: `linguist.Reader.Decode`, `detectDialect`, `firstElementName`, `extractListInner`. Go's `encoding/xml` is already lenient by default; Strict mode rejects malformed entities and requires well-formed XML.

## E9.T4 — Bounded reads (io.LimitedReader)

- `ReadBounded(r io.Reader, limit int64)` reads `limit+1` bytes via `io.LimitReader` then checks `len(data) > limit`. Reading one extra byte distinguishes "exactly at limit" (ok) from "over limit" (error) without needing the input size upfront.
- `ReadFileBounded` should NOT wrap `os.Open` errors with an "opening file:" prefix — callers already wrap with their own context (e.g. `"reading %s: %s"`). Double-wrapping produces confusing messages and breaks golden file expectations.
- `wrapLoadError(err)` checks if an error already implements `Code() string` before wrapping with `ErrValidationError`. Without this, `ErrInputTooLarge` and `ErrDangerousDoctype` from `tbx.Load` get masked by `validation_error` code. Side effect: the `doctype_entity` golden file changed from `validation_error` to `invalid_input` — which is actually more correct.
- `write/apply.go`'s `readPayloadFile` wraps errors with `fmt.Errorf("reading payload file: %w", ...)`. This `%w` wrapping preserves `errors.Is` chains but loses the `Coded` interface for type assertions. The fix: check `Coded` before wrapping, pass coded errors through directly.
- Integration tests for bounded reads must create oversized files within CWD subtree (under `testdata/`), not in `t.TempDir()` (which uses `/tmp/`), because `sanitizePath` rejects paths outside the CWD sandbox. The `writeLargeFileInCWD` helper uses `os.MkdirTemp("testdata", "bounded-*")` with `t.Cleanup`.
- Adding `ErrInputTooLarge` sentinel changes `terr.All()` → schema golden files need regeneration with `-update`. Same pattern as E5.T7, E6.T2, E7.T7, E8.T1, E9.T1.

## E9.T5 — Path sandbox (CWD enforcement, --tbx exemption)

- The existing `sanitizePath` string-level checks (`hasPathTraversal`, `hasPercentEncoded`, etc.) catch most attack vectors before `resolveAndSandbox` is reached. Absolute paths like `/etc/passwd` pass all string checks and are only caught by the prefix check inside `resolveAndSandbox`. Golden tests should cover both vectors (string-level rejection via `..` AND sandbox-level rejection via absolute paths) to ensure both layers are exercised.
- `sanitizeTBXPath` deliberately omits the `resolveAndSandbox` call — agents legitimately pass absolute paths to shared glossaries outside CWD. The only defense is the string-level check (no `..`, no `%`, no `?#`, no control chars).
- `TestSanitizeTBXPath` verifies that absolute paths are accepted (not sandboxed) while `..` and injection patterns are still rejected. This is distinct from `TestSanitizePath` which rejects absolute paths outside the base directory.

## E9.T6 — Fuzz tests (TBX decoder, matcher, DeriveID)

- Go's `f.Add()` calls are the primary seed mechanism; files under `testdata/fuzz/<FuncName>/` are additional corpus entries. Go stores fuzzer-discovered interesting inputs in the system cache (`$GOCACHE/fuzz/`), not in the project directory — so `testdata/fuzz/` stays clean with only manually crafted seeds.
- `FuzzMatcherScan` builds the `Matcher` once in the fuzz function (before `f.Fuzz()`), not inside the fuzz loop. `match.New` is expensive (builds Aho-Corasick automaton) and the glossary is fixed — only the input text varies.
- `FuzzLinguistDecode` must use `strings.NewReader(string(data))` rather than `bytes.NewReader(data)` because the `Decode` method signature takes `io.Reader` — both work, but `strings.NewReader` avoids a lint nit about converting `[]byte` to `string` being redundant with `bytes.NewReader`.
- staticcheck `QF1001` (De Morgan's law) fires on `!((a || b || c))` patterns in fuzz assertions. Use `(not-a && not-b && not-c)` form instead. Same lesson as E9.T1.

## E9.T7 — Perf budget tests + make perf target

- Synthetic markdown fixtures must control term density (how often glossary terms appear in the generated text). Dense fixtures where every sentence contains a term cause quadratic blowup in match processing — a 5000-concept × 5MB scan took 66s with dense terms but only 450ms with sparse (1-in-50) term density. Real-world academic prose has sparse term occurrences.
- `//go:build perf` tag correctly excludes perf tests from `go test ./...`, `go vet ./...`, and `golangci-lint run ./...` — no special configuration needed in linter config or vet flags.
- `make perf` uses `-count=1` to disable test caching (perf tests should always re-run) and `-v` to show timing output from `t.Logf`.
- Perf test files use `_test` package suffix (e.g. `package tbx_test`) to test through the public API, consistent with the project's test convention. Fixture generators are local to each test file since they need package-specific types.
- The `check` budget (10s) is roughly 2× the `scan` budget (5s) because check runs two scans (source + target) plus violation analysis. The actual ratio is less than 2× due to shared overhead (glossary loading, goldmark parsing).
