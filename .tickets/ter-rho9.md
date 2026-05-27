---
id: ter-rho9
status: closed
deps: [ter-39xj]
links: []
created: 2026-05-27T19:35:03Z
type: task
priority: 2
assignee: Andre Silva
tags: [e10, makefile, build]
---
# E10.T1 — Reproducible-build flags in Makefile LDFLAGS

Add -trimpath and -buildid= to the Makefile so every build produces reproducible output.

## Spec reference

docs/specs/010-release.md §Reproducible builds:

    go build -trimpath -ldflags "-s -w -buildid= -X main.version=${VERSION}"

## Current state

LDFLAGS currently: -s -w -X ...version.Override=$(VERSION)

Missing: -trimpath (go build flag, not an ldflag) and -buildid= (ldflag).

## Acceptance criteria

- make build uses -trimpath and -buildid= in addition to existing flags
- All build targets (build-local, build-target template, run) inherit the flags
- -trimpath is a go build flag (not an ldflag) — add it to the go build invocation, not LDFLAGS
- -buildid= is an ldflag — add it to LDFLAGS
- make build still passes

