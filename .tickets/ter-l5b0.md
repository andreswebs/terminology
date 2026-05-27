---
id: ter-l5b0
status: closed
deps: [ter-yfoe, ter-1zvj, ter-6z5g]
links: []
created: 2026-05-24T00:29:54Z
type: task
priority: 2
assignee: Andre Silva
parent: ter-uqyn
tags: [e2, task, testing, fixtures]
---
# E2.T14 — Test fixtures + round-trip property tests

## Goal

Create the comprehensive test fixture corpus and the round-trip property tests that assert the E2 acceptance criteria: \`read → write → read\` produces an equivalent model, and canonical output is byte-stable across runs.

## Refs

- E2 spec: [docs/specs/002-domain-and-io.md](docs/specs/002-domain-and-io.md) §"Round-trip fidelity for unknown elements", §"Scope" acceptance criteria
- Testing ADR: [docs/adr/testing.md](docs/adr/testing.md) §"Property tests" (round-trip invariant), §"Test fixtures layout"
- Determinism ADR: [docs/adr/determinism.md](docs/adr/determinism.md) §"Read/write canonicalization model" — sort on write only

## Files to create

### Canonical fixtures (\`src/internal/tbx/linguist/testdata/canonical/\`)
- \`minimal-dct.tbx\` — one concept, two languages, one term each (created in T6 if not already present)
- \`minimal-dca.tbx\` — same content in DCA style (created in T7 if not already present)
- \`rich-dct.tbx\` — two concepts, three languages, multiple terms with varying statuses, definitions, cross-references
- \`full-features.tbx\` — exercises all data categories: subjectField, definitions, crossReferences, externalRefs, graphics, sources, customerSubset, projectSubset, notes, transactions, adminGrp (reading/readingNote), contexts, transferComment, all term-level elements
- \`with-transactions.tbx\` — transaction groups at concept and term level

### Normalized fixtures (\`src/internal/tbx/linguist/testdata/normalized/\`)
- \`legacy-forms.tbx\` — bare status forms + usageRegister (created in T9 if not already present)

### Test files
- \`src/internal/tbx/linguist/roundtrip_test.go\`

## Fixture design principles

- **Canonical fixtures must be in canonical form**: the writer produces them byte-identically on round-trip. This means concepts sorted by ID, languages sorted by tag, terms sorted by status, 2-space indent, LF endings, trailing newline, no BOM, fixed namespace prefixes.
- **Each fixture tests one thing**: minimal = smallest valid file, rich = multiple concepts/terms, full-features = all data categories, with-transactions = transaction groups. No monolithic catch-all.
- **Naming**: one fixture per scenario, named by what it exercises.

## Round-trip test

```go
func TestRoundTrip_Canonical(t *testing.T) {
    fixtures, err := filepath.Glob("testdata/canonical/*.tbx")
    // ...
    for _, path := range fixtures {
        name := filepath.Base(path)
        if name == "minimal-dca.tbx" {
            continue // DCA→write produces DCT, not DCA
        }
        t.Run(name, func(t *testing.T) {
            original, err := os.ReadFile(path)
            // ...
            r := NewReader()
            g, _, err := r.Decode(bytes.NewReader(original))
            // ...
            var buf bytes.Buffer
            w := NewWriter()
            if err := w.Encode(&buf, g); err != nil {
                t.Fatalf("Encode: %v", err)
            }
            if !bytes.Equal(original, buf.Bytes()) {
                t.Errorf("round-trip mismatch for %s:\n--- original ---\n%s\n--- written ---\n%s",
                    name, string(original), buf.String())
            }
        })
    }
}
```

Design notes:
- **DCA fixtures are excluded from round-trip** because the writer always produces DCT. DCA→DCT conversion is tested separately (DCA fixture → write → compare against DCT fixture).
- **Byte-for-byte comparison** (\`bytes.Equal\`) is intentional, not structural JSON comparison. The determinism contract guarantees byte-stable output; a structural comparator would mask the exact bugs (key order, whitespace, encoding quirks) that the determinism layer prevents.
- **No -update flag** for round-trip tests — fixtures are hand-written canonical forms, not auto-generated. If the round-trip fails, the fixture or the writer needs fixing, not auto-updating.

## DCA→DCT conversion test

```go
func TestDCAtoCanonicalDCT(t *testing.T) {
    // Decode minimal-dca.tbx (DCA input)
    // Encode to buffer (produces DCT)
    // Compare against minimal-dct.tbx (expected DCT output)
}
```

## TDD cycles

### Cycle 1 — Minimal round-trip
RED: Read minimal-dct.tbx → write → compare bytes. Assert equal.
GREEN: Ensure fixture is in canonical form and writer matches.

### Cycle 2 — Rich fixture round-trip
RED: Create rich-dct.tbx with two concepts, three languages, multiple status values. Read → write → compare.
GREEN: Fix any ordering or formatting issues in the writer.

### Cycle 3 — Full features round-trip
RED: Create full-features.tbx with all data categories. Read → write → compare.
GREEN: Ensure all data categories round-trip correctly.

### Cycle 4 — Transaction round-trip
RED: Create with-transactions.tbx. Read → write → compare.
GREEN: Ensure transaction groups at both levels round-trip.

### Cycle 5 — DCA→DCT conversion
RED: Decode minimal-dca.tbx, encode, compare against minimal-dct.tbx.
GREEN: Verify the writer normalizes DCA-sourced data to canonical DCT.

### Cycle 6 — All canonical fixtures
RED: Glob all canonical DCT fixtures and assert round-trip for each.
GREEN: The glob-based test runner catches any new fixture that breaks round-trip.

### Cycle 7 — Legacy normalization integration
RED: Decode legacy-forms.tbx, assert statuses are normalized enum values, assert usageRegister→register normalization.
GREEN: Normalization from T9 is integrated into reader from T6/T7.

## Out of scope

- Golden CLI tests (E3+ — those test urfave commands, not the reader/writer directly)
- Fuzz tests (E9)
- Performance tests (E9)
- Validation tests (E3)

## Acceptance

- \`make build\` passes
- All canonical DCT fixtures round-trip byte-identically
- DCA fixtures decode to equivalent models as their DCT counterparts
- DCA→write produces canonical DCT matching the expected fixture
- Legacy-forms fixture demonstrates normalization working correctly
- Fixture corpus covers: minimal, multi-concept, all data categories, transactions


## Notes

**2026-05-25T13:50:26Z**

All fixtures and round-trip tests were already in place from prior tickets (T6–T10). Added three new tests to roundtrip_test.go to close remaining acceptance criteria gaps: (1) TestDCAModelEquivalence — decodes DCA and DCT fixtures, normalizes Style, asserts reflect.DeepEqual to verify structural model equivalence. (2) TestLegacyNormalization_WriteCanonical — decodes legacy-forms.tbx and verifies the writer outputs normalized suffixed statuses and register (not bare forms or usageRegister). (3) TestRoundTrip_Stability — runs read→write→read→write and asserts the two outputs are byte-identical, verifying the determinism contract across consecutive passes. All 52 linguist package tests pass. make build clean.
