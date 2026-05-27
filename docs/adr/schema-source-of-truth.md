# Cross-cutting — Schema as source of truth

> **Status**: APPROVED. Go code is the canonical contract. The
> `terminology schema` subcommand exposes it reflectively for agents.

## The invariant

**Go code is the source of truth.** Three concrete locations together
form the contract every other surface mirrors:

- `internal/app/commands/*.go` — `urfave/cli/v3` command and flag
  declarations. Names, types, defaults, picklist values, short
  aliases, environment sources.
- `internal/output/types.go` — Go struct types for every envelope
  shape. JSON tags + field types fully describe what the binary emits.
- `internal/<pkg>/errors.go` — `terr.New(...)` sentinels. One sentinel
  per error code. The full set of sentinels IS the error registry
  (see [error-handling](error-handling.md)).

There is **no separately maintained JSON Schema file** to keep in sync.
The CLI's behavior is its specification.

## Agent-facing introspection

`terminology schema` is the canonical agent discovery surface. Its
implementation is **reflective**: at runtime it walks the live urfave
command tree, reflects over the output struct types, and enumerates the
sentinel registry, then emits a JSON document describing the binary's
contract.

```
$ terminology schema
{
  "schema_version": 1,
  "commands": [ ... ],     // from urfave introspection
  "envelopes": { ... },    // from internal/output reflection
  "error_codes": [ ... ]   // from terr sentinel registry
}

$ terminology schema --command validate
{
  "name": "validate",
  "flags": [ ... ],
  "envelope": { ... },
  "exit_codes": [0, 1, 65, 3]
}
```

Because the JSON is computed from the live binary, drift is structurally
impossible: there is nothing to drift *from*.

## Error registry & generated reference

Error codes are registered implicitly: every `var ErrXxx =
terr.New(code, ...)` is collected by `internal/terr` at package init
time (each producing package's `errors.go` imports `terr`, so the
collection happens whenever the binary's command tree is built).

For human-readable browsing, `make docs` (wired via `go generate`)
walks the sentinel registry once and produces
`docs/reference/errors.md`:

```markdown
# Error codes

## `unsupported_dialect` — exit 65

unsupported TBX dialect

**Hint:** supported: TBX-Linguist
**Produced by:** `internal/tbx`
```

CI asserts the generated doc matches the current sentinels (regenerate
& `git diff --exit-code`). The doc is checked in so reviewers see code
additions in PRs without running the binary.

## Envelope `schema_version`

Every envelope carries a top-level `schema_version` integer:

```json
{
  "schema_version": 1,
  "ok": true,
  ...
}
```

It tracks the envelope contract emitted by the Go code, not a stored
schema file. Bump rules:

- **Adding an envelope field**: no bump.
- **Adding a new command / flag / error code**: no bump.
- **Tightening a type**: bump.
- **Removing or renaming a field / flag / command / code**: bump.
- **Changing exit-code mapping for an error code**: bump.

Single integer (per
[E10 Q2](specs/010-release.md#q2--schema-version-integer-or-semver)).
Bumping is a release-gate concern; the constant lives in
`internal/output/version.go`.

## Bootstrap artifact

[`docs/specs/target.schema.json`](specs/target.schema.json) is a
**scaffolding artifact**. It served as the initial declarative source
for stubbing out commands, flags, envelope types, and error sentinels.

Once the Go scaffolding is in place, the file is **deleted**. It is not
maintained in parallel with the Go code; references to it in other
docs are pruned. The file's presence in `docs/specs/` during the
scaffolding phase does not make it canonical — Go code already is.

This document, after scaffolding, retains its filename for backwards
linkage but describes the post-scaffolding world: Go-as-truth, no
stored schema, reflective introspection.

## CI sync checks

With Go as the truth, most of the original sync surface evaporates.
What remains in `make validate`:

### 1. Scenarios parse cleanly

Every fenced `terminology …` invocation in
[`docs/examples/scenarios.md`](examples/scenarios.md) is fed to the
urfave parser (no action invocation). Assert: no parse errors. Catches
docs drift when commands/flags are renamed.

Test lives in `internal/app/scenarios_sync_test.go`.

### 2. Generated error reference is current

Re-run the docs generator; assert `git diff --exit-code` on
`docs/reference/errors.md`. Forces contributors to commit the
regenerated doc alongside any sentinel change.

Test lives in `internal/terr/docs_sync_test.go` (invokes the generator
in-process and compares output to the checked-in file).

### 3. Schema-version monotonicity

When `schema_version` changes, assert it only ever increases. Test
diffs `version.go` against the previous tagged release.

Test lives in `internal/output/version_test.go`.

The historical "schema ↔ urfave commands" and "schema ↔ output struct
types" checks no longer exist: urfave **is** the command truth and
the struct types **are** the envelope truth. There is nothing external
to compare against.

## Runtime envelope validation

The binary does **not** validate its own output against a schema at
runtime. There is no embedded schema to validate against, and the
output structs serialize deterministically by construction. Golden-file
tests cover envelope shapes per command (see
[testing](testing.md)).

## Deprecation handling

No flags or error codes are deprecated at v1. When the first
deprecation arrives, the convention will be added (likely a struct tag
on the urfave flag declaration + a `Deprecated` field on `terr.E`).
Designing it ahead of a real case is premature.

## Hand-curation discipline

The discipline that mattered under the curated-schema model
(remembering to update the registry when adding a code) is replaced by
the **sentinel-as-registry** rule: you cannot add an error code
without writing the sentinel — there is no other place to declare it.
The bidirectional check is intrinsic:

- Every code returned by the binary came from a `terr.E`, which came
  from a sentinel.
- Every sentinel appears in the generated reference doc; CI fails if
  the doc is stale.

The failure mode "added a code but forgot to register it" cannot occur.
