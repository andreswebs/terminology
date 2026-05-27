# Usage scenarios

Concrete end-to-end transcripts exercising the surface defined in
[`docs/design.md`](../design.md). Each scenario shows the agent's or human's
intent, the commands run, representative payloads, and what comes back
(including exit codes).

Transcripts assume `TERMINOLOGY_TBX=glossary/terms.tbx` is set in the
environment unless otherwise noted, and that `--format=json` is the default
(elided from command lines for brevity).

---

## 1. First-time scan of an inherited corpus

**Setting.** A translator joins a project mid-stream. The repository has
Spanish source under `source/`, a partial English translation under
`target/`, and a sparse `glossary/terms.tbx` started by a previous
collaborator. Goal: understand what's covered, what's missing, and where the
existing translation already drifts.

### Step 1 — see what the glossary covers

```sh
terminology validate
```

```json
{
  "ok": true,
  "concepts": 12,
  "languages": ["es", "he", "en"],
  "warnings": []
}
```

Exit `0`. Twelve concepts, three languages — small glossary.

### Step 2 — find candidate terms across the whole source corpus

```sh
terminology extract source/*.md \
    --exclude "$TERMINOLOGY_TBX" \
    --min-freq 5 \
    > candidates.json
```

```json
{
  "ok": true,
  "candidates": [
    {
      "term": "צמצום",
      "script": "hebrew",
      "frequency": 47,
      "sample_contexts": [
        "...la noción de צמצום aparece...",
        "...el concepto de צמצום..."
      ]
    },
    {
      "term": "Razón Histórica",
      "script": "latin",
      "frequency": 23,
      "sample_contexts": ["...la Razón Histórica orteguiana..."]
    },
    {
      "term": "Sefirot",
      "script": "latin",
      "frequency": 18,
      "sample_contexts": ["..."]
    }
  ]
}
```

Exit `0`. `--exclude` filters out terms already in the TBX, so the candidates
are genuinely new territory.

### Step 3 — verify the existing translation against the existing glossary

```sh
for src in source/ch*.md; do
    tgt="target/$(basename "$src")"
    [ -f "$tgt" ] && terminology check "$src" "$tgt"
  done
```

Output for `ch01.md`:

```json
{
  "ok": false,
  "source": "source/ch01.md",
  "target": "target/ch01.md",
  "violations": [
    {
      "type": "forbidden_variant",
      "concept_id": "tzimtzum",
      "variant": "contraction",
      "line": 17,
      "column": 4,
      "context": "...the divine contraction described by Luria..."
    }
  ],
  "warnings": [],
  "summary": {
    "violations": 1,
    "warnings": 0,
    "concepts_checked": 8
  }
}
```

Exit `1` for `ch01.md` (one violation), exit `0` for chapters with no
findings. The translator now has a triage list: 1 inherited violation +
N new candidate terms to curate.

---

## 2. Agent-driven chapter translation, fix-loop

**Setting.** An agent is translating `source/ch3.md` to `target/ch3.md`. The
glossary covers the key terms. The agent uses `scan` once up front, drafts
the translation, then iterates with `check` until exit 0.

### Step 1 — learn which glossary terms appear in this chapter

```sh
terminology scan source/ch3.md --lang es
```

```json
{
  "ok": true,
  "file": "source/ch3.md",
  "matches": [
    {
      "concept_id": "tzimtzum",
      "term": "tzimtzum",
      "lang": "es",
      "line": 14,
      "column": 23,
      "context": "...El concepto de tzimtzum es central..."
    },
    {
      "concept_id": "razon-historica",
      "term": "Razón Histórica",
      "lang": "es",
      "line": 41,
      "column": 12,
      "context": "...la Razón Histórica orteguiana..."
    }
  ],
  "summary": {
    "total_matches": 7,
    "unique_concepts": 2
  }
}
```

The agent now knows the two concepts in play. Concept IDs are stable, so it
can refer to them across iterations.

### Step 2 — first draft, then check

After producing `target/ch3.md`, the agent runs:

```sh
terminology check source/ch3.md target/ch3.md \
    --source-lang es --target-lang en
```

```json
{
  "ok": false,
  "source": "source/ch3.md",
  "target": "target/ch3.md",
  "violations": [
    {
      "type": "missing",
      "concept_id": "tzimtzum",
      "source_term": "tzimtzum",
      "source_occurrences": 4,
      "expected_target": "tzimtzum",
      "target_occurrences": 1
    },
    {
      "type": "forbidden_variant",
      "concept_id": "tzimtzum",
      "variant": "contraction",
      "line": 22,
      "column": 8,
      "context": "...the primordial contraction makes room..."
    },
    {
      "type": "forbidden_variant",
      "concept_id": "tzimtzum",
      "variant": "constriction",
      "line": 58,
      "column": 14,
      "context": "...this constriction is paradoxical..."
    }
  ],
  "warnings": [],
  "summary": {
    "violations": 3,
    "warnings": 0,
    "concepts_checked": 2
  }
}
```

Exit `1`. Every violation embeds the **expected_target** and a **context
window**, so the agent doesn't need a follow-up `lookup`. It replaces both
"contraction" and "constriction" with "tzimtzum" using the line/column
coordinates.

### Step 3 — re-run until green

```sh
terminology check source/ch3.md target/ch3.md \
    --source-lang es --target-lang en
```

```json
{
  "ok": true,
  "source": "source/ch3.md",
  "target": "target/ch3.md",
  "violations": [],
  "warnings": [],
  "summary": {
    "violations": 0,
    "warnings": 0,
    "concepts_checked": 2
  }
}
```

Exit `0`. Chapter done.

---

## 3. Curating new concepts from extraction output

**Setting.** Following on from scenario 1, the curator wants to promote three
new concepts into the TBX: a simple single-language stub, a Spanish-English
pair, and a richer Spanish-Hebrew-English concept with definitions and
notes. Author identity comes from the environment.

```sh
export TERMINOLOGY_AUTHOR="Shoushani"
```

### Quick single-language add via flags

```sh
terminology concept add \
    --lang es --term "Sefirot" \
    --subject-field kabbalah \
    --dry-run
```

Dry-run preview (final state of the touched concept):

```json
{
  "ok": true,
  "preview": {
    "concept_id": "sefirot",
    "subject_field": "kabbalah",
    "languages": {
      "es": {
        "preferred": { "term": "Sefirot" }
      }
    }
  }
}
```

Exit `0`. The slug `sefirot` was derived from the canonical-language term
(`en` not present → falls back to first lang in doc order = `es`). Commit:

```sh
terminology concept add \
    --lang es --term "Sefirot" \
    --subject-field kabbalah \
    --transaction
```

```json
{
  "ok": true,
  "concept": {
    "concept_id": "sefirot",
    "subject_field": "kabbalah",
    "languages": {
      "es": { "preferred": { "term": "Sefirot" } }
    }
  }
}
```

### Multi-language add via stdin JSON

```sh
cat <<'EOF' | terminology concept add --transaction
{
  "concept_id": "razon-historica",
  "subject_field": "philosophy",
  "definitions": [
    "Ortega y Gasset's notion that reason is fundamentally historical."
  ],
  "notes": ["Retain Spanish surface form in English target."],
  "languages": {
    "es": {
      "preferred": {
        "term": "Razón Histórica",
        "part_of_speech": "noun",
        "grammatical_gender": "feminine"
      }
    },
    "en": {
      "preferred": {
        "term": "Razón Histórica",
        "part_of_speech": "noun"
      }
    }
  }
}
EOF
```

Exit `0`. Returns the created concept.

### Rich concept with deprecated variants

```sh
cat <<'EOF' | terminology concept add --transaction
{
  "concept_id": "tzimtzum",
  "subject_field": "kabbalah",
  "definitions": [
    "Kabbalistic concept of divine self-contraction preceding creation."
  ],
  "notes": ["Retain transliteration in English target."],
  "languages": {
    "he": {
      "preferred": {
        "term": "צמצום",
        "part_of_speech": "noun",
        "grammatical_gender": "masculine"
      }
    },
    "es": {
      "preferred": {
        "term": "tzimtzum",
        "part_of_speech": "noun",
        "register": "technicalRegister"
      }
    },
    "en": {
      "preferred": {
        "term": "tzimtzum",
        "part_of_speech": "noun"
      },
      "deprecated": [
        { "term": "contraction" },
        { "term": "constriction" }
      ]
    }
  }
}
EOF
```

### Collision case

If the caller tries to re-add `tzimtzum` later, the tool refuses:

```sh
terminology concept add --id tzimtzum --lang en --term tzimtzum
```

stderr:

```json
{
  "ok": false,
  "error": {
    "code": "duplicate_id",
    "message": "concept 'tzimtzum' already exists",
    "hint": "use `terminology concept update tzimtzum` to modify, or pick a different --id"
  }
}
```

Exit `2`.

---

## 4. Bulk import from an external glossary

**Setting.** A collaborator hands over a hand-edited TBX fragment containing
thirty new concepts authored in OmegaT (DCA style). The curator wants to
merge it without disturbing the existing glossary, but with full visibility
into what will change.

### Step 1 — dry-run apply

```sh
terminology apply -f incoming.tbx --format tbx --dry-run
```

```json
{
  "ok": true,
  "applied": {
    "added": [
      "binah",
      "chesed",
      "gevurah",
      "tiferet",
      "netzach",
      "hod",
      "yesod",
      "malkhut",
      "ein-sof",
      "or-ein-sof"
    ],
    "updated": ["tzimtzum"],
    "removed": [],
    "unchanged": ["razon-historica", "sefirot"]
  },
  "warnings": [
    {
      "code": "dangling_crossref_target",
      "concept_id": "or-ein-sof",
      "target": "ein-sof",
      "message": "crossReference resolves within this batch but not in the existing file"
    }
  ]
}
```

Exit `0`. Twenty-five concepts in the incoming file matched existing entries
unchanged, ten will be added, `tzimtzum` will be updated (the new file
included an additional Hebrew note). The warning is informational — the
reference resolves once the batch is applied.

### Step 2 — commit

```sh
terminology apply -f incoming.tbx --format tbx --transaction
```

Same envelope, this time with `warnings: []` (the references now resolve
on-disk). Exit `0`.

The file on disk is canonical DCT regardless of the DCA input.

### Step 3 — confirm integrity

```sh
terminology validate
```

```json
{
  "ok": true,
  "concepts": 25,
  "languages": ["es", "he", "en"],
  "warnings": []
}
```

Exit `0`.

### What a hard failure looks like

If `incoming.tbx` had a truly dangling reference (target absent from both
sides), `apply` would fail without touching the file:

stderr:

```json
{
  "ok": false,
  "error": {
    "code": "dangling_crossref",
    "message": "concept 'or-ein-sof' references unresolved target 'ein-sof'",
    "hint": "include 'ein-sof' in the payload, or remove the crossReference"
  }
}
```

Exit `3`. The TBX is left untouched.

---

## 5. Deprecating a variant mid-project

**Setting.** Editorial decision: in the English translation, "contraction"
must no longer be an accepted variant of `tzimtzum`. Mark it deprecated, then
sweep every translated chapter for newly-surfaced violations.

### Step 1 — deprecate the variant

```sh
terminology term deprecate tzimtzum \
    --lang en --term contraction \
    --transaction --author "editor"
```

```json
{
  "ok": true,
  "concept": {
    "concept_id": "tzimtzum",
    "languages": {
      "en": {
        "preferred": { "term": "tzimtzum" },
        "deprecated": [{ "term": "contraction" }, { "term": "constriction" }]
      }
    }
  }
}
```

Exit `0`. `contraction` was previously absent from the `deprecated` list; it
is now flagged.

### Step 2 — sweep all translated chapters

```sh
for src in source/ch*.md; do
    tgt="target/$(basename "$src")"
    [ -f "$tgt" ] || continue
    terminology check "$src" "$tgt" \
        --source-lang es --target-lang en \
        --fields violations.concept_id,violations.variant,violations.line,violations.context \
        > "checks/$(basename "$src" .md).json"
  done
```

The `--fields` projection trims each chapter's report to just what the agent
needs to fix violations — concept ID, variant, location, context — keeping
the aggregate payload small enough to feed back to the agent in one shot.

Sample chapter result:

```json
{
  "ok": false,
  "violations": [
    {
      "concept_id": "tzimtzum",
      "variant": "contraction",
      "line": 17,
      "context": "...the divine contraction described by Luria..."
    },
    {
      "concept_id": "tzimtzum",
      "variant": "contraction",
      "line": 84,
      "context": "...this contraction precedes any emanation..."
    }
  ]
}
```

Exit `1`. The agent (or human) sweeps each occurrence, replaces it with
`tzimtzum`, and re-runs `check` per chapter until everything is `ok: true`.

### Step 3 — confirm the glossary itself is still consistent

```sh
terminology validate --strict
```

```json
{
  "ok": true,
  "concepts": 25,
  "languages": ["es", "he", "en"],
  "warnings": []
}
```

Exit `0`. The editorial decision is reflected in the TBX, propagated through
the corpus, and durable across future translation runs.
