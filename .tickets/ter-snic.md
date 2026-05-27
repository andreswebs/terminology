---
id: ter-snic
status: closed
deps: [ter-rho9, ter-39xj]
links: []
created: 2026-05-27T19:35:17Z
type: task
priority: 2
assignee: Andre Silva
tags: [e10, makefile, release]
---
# E10.T2 — make release target (tag + dist + gh release create)

Add a make release VERSION=v0.x.y target that:

1. Tags the current commit with the given version
2. Runs make dist (which runs build-all)
3. Uploads to GitHub Releases via gh release create

## Spec reference

docs/specs/010-release.md §Build & release flow:

    make release VERSION=v0.x.y — tags the commit, regenerates dist,
    uploads to GitHub Releases via gh release create. Injects the
    version string via -ldflags.

## Design notes

- VERSION is already a Makefile variable (defaulting to git describe)
- The target should fail early if VERSION is not explicitly set (no default for release)
- The target should fail if the working tree is dirty
- Tag format: v0.x.y (must match semver)
- gh release create should attach all dist/ archives + SHA256SUMS.txt
- Release notes: capture Go version and toolchain for reproducibility, per spec
- The release tag should be an annotated tag (git tag -a)

## Acceptance criteria

- make release VERSION=v0.1.0 creates annotated tag v0.1.0
- make release VERSION=v0.1.0 produces dist/ archives with version in filenames
- make release VERSION=v0.1.0 calls gh release create with attached artifacts
- Target fails if VERSION not explicitly provided
- Target fails if working tree is dirty
- --version output matches the injected version string

