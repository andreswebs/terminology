---
id: ter-9ixx
status: closed
deps: [ter-39xj]
links: []
created: 2026-05-27T19:35:43Z
type: task
priority: 2
assignee: Andre Silva
tags: [e10, codegen, docs]
---
# E10.T4 — Error reference generator (go:generate + docs/reference/errors.md)

Create a go:generate directive that produces docs/reference/errors.md from the terr sentinel registry.

## Spec reference

docs/specs/010-release.md §Docs / CI sync:

    Errors reference is fresh. docs/reference/errors.md is generated
    from the terr sentinel registry. CI re-runs the generator and
    fails on diff. Ensures documented codes match compiled-in codes.

## Design notes

- terr.All() returns []*terr.E with Code(), ExitCode(), Hint(), Error() methods
- terr.New() registers sentinels; terr.Newf() does not (runtime errors)
- The generator should be a small Go program (e.g. under tools/ or internal/gen/)
- Use //go:generate directive to invoke it
- Output: a Markdown file listing all error codes, their exit codes, hints, and descriptions
- The generator must import all packages that declare sentinels so the registry is populated
- The generated file should include a header warning that it is auto-generated

## Acceptance criteria

- go generate ./... produces docs/reference/errors.md
- The file lists every registered sentinel (code, exit code, hint, description)
- Running the generator twice produces identical output (deterministic)
- The generated file matches the current set of sentinels

