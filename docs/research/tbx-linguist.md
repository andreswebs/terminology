# TBX-Linguist

## Overview

The TBX-Linguist dialect (ISO 30042:2019 family) is a **small delta** over TBX-Basic. The Linguist module adds only five data categories on top of Basic+Min+Core. Everything one might expect to be Linguist-specific — `definition`, `context`, `grammaticalGender`, `partOfSpeech`, `subjectField`, `xGraphic`, transactions, cross-references, inline markup — is already inherited from Core, Min, or Basic. The dialect's value is the combination, not novel structure.

Practical implication: an implementation that already supports TBX-Basic correctly only needs to (a) accept the new dialect identifier, (b) declare the Linguist namespace, and (c) implement the five extra data categories listed below.

## Dialect identity

| Item                | Value                                                 |
| ------------------- | ----------------------------------------------------- |
| `<tbx>` `type=`     | `TBX-Linguist`                                        |
| `<tbx>` `style=`    | `dca` or `dct`                                        |
| Default xmlns       | `urn:iso:std:iso:30042:ed-2`                          |
| Linguist namespace  | `http://www.tbxinfo.net/ns/linguist` (prefix `ling:`) |
| Schema availability | RNG + Schematron + NVDL (no XSD published)            |

A DCT-style document must also declare the transitively-included module namespaces, because Linguist sits on top of Basic and Min:

```xml
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
```

The dialect-level Schematron enforces that `@type` exactly equals `TBX-Linguist` and that `@style` matches the file (DCA vs DCT).

### Namespace URI caveat

Prose on tbxinfo.net mentions `urn:iso:std:iso:30042:ed:3.0`, but the actual dialect schemas and LTAC-Global example files use `urn:iso:std:iso:30042:ed-2`. **The dialect files are the source of truth for parsers.** Accept `ed-2` on read; emit `ed-2` on write.

## Module governance

| Item                                | Value                                                                 |
| ----------------------------------- | --------------------------------------------------------------------- |
| Dialect repo                        | `github.com/LTAC-Global/TBX-Linguist_dialect` (last update 2019-04)   |
| Module repo                         | `github.com/LTAC-Global/TBX_linguist_module` (last update 2018-11)    |
| Module version                      | 1.0 / 1.1                                                             |
| Listed on tbxinfo.net/tbx-dialects/ | No — only Core, Min, Basic appear                                     |
| Authoritative spec                  | `Linguist.tbxmd` + `Linguist.rng` + `Linguist.sch` in the module repo |

Linguist is not annexed into the published ISO 30042:2019 PDF; the standard defines the dialect methodology, and the dialect itself lives in the LTAC-Global GitHub repository.

## Data categories added by the Linguist module

These five are the entire Linguist contribution. All are constrained to `termSec` level.

| Name                | Spec type     | DatCat ID | Value type | DCA carrier                                            | DCT element                |
| ------------------- | ------------- | --------- | ---------- | ------------------------------------------------------ | -------------------------- |
| `grammaticalNumber` | termNoteSpec  | DC-251    | picklist   | `<termNote type="grammaticalNumber">`                  | `<ling:grammaticalNumber>` |
| `register`          | termNoteSpec  | DC-423    | picklist   | `<termNote type="register">`                           | `<ling:register>`          |
| `transferComment`   | termNoteSpec  | DC-520    | string     | `<termNote type="transferComment">`                    | `<ling:transferComment>`   |
| `reading`           | adminSpec     | UNKNOWN   | string     | `<admin type="reading">` (inside `<adminGrp>`)         | `<ling:reading>`           |
| `readingNote`       | adminNoteSpec | UNKNOWN   | noteText   | `<adminNote type="readingNote">` (inside `<adminGrp>`) | `<ling:readingNote>`       |

### Picklist values

`grammaticalNumber`:

- `singular`
- `plural`
- `dual`
- `mass`
- `otherNumber`

`register`:

- `colloquialRegister`
- `neutralRegister`
- `technicalRegister`
- `in-houseRegister`
- `bench-levelRegister`
- `slangRegister`
- `vulgarRegister`

### Notes on individual categories

**`register`** was previously called `usageRegister`. The Linguist module description explicitly says implementations should accept legacy `usageRegister` on read and normalize it to `register`.

**`reading` / `readingNote`** support orthographic readings (e.g., Japanese kana readings of kanji terms). They are the genuinely "linguistic" element of the module. Both have no DatCatInfo identifier — they exist only inside the Linguist module.

**Schematron placement rules** (from `Linguist.sch`):

- `ling:reading` — parent must be `tbx:termSec`, or `tbx:adminGrp` inside `tbx:termSec`.
- `ling:readingNote` — parent must be `tbx:adminGrp` inside `tbx:termSec`. (Bare `adminNote` is not allowed.)
- `ling:grammaticalNumber`, `ling:register`, `ling:transferComment` — must be at `termSec` or inside `termNoteGrp/termSec`.

## What Linguist inherits (and where from)

Categories commonly assumed to be "Linguist features" but actually defined in lower modules:

### From the Basic module (`http://www.tbxinfo.net/ns/basic`)

| Category            | DCA carrier   | Levels                               | Value type                                                                                                                                                                                                                                                                                 |
| ------------------- | ------------- | ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `context`           | `descrip`     | `termSec`                            | noteText                                                                                                                                                                                                                                                                                   |
| `crossReference`    | `ref`         | `conceptEntry`, `termSec`            | string                                                                                                                                                                                                                                                                                     |
| `definition`        | `descrip`     | `conceptEntry`, `langSec`            | noteText                                                                                                                                                                                                                                                                                   |
| `geographicalUsage` | `termNote`    | `termSec`                            | free-form string                                                                                                                                                                                                                                                                           |
| `grammaticalGender` | `termNote`    | `termSec`                            | picklist: `masculine`, `feminine`, `neuter`, `other`                                                                                                                                                                                                                                       |
| `projectSubset`     | `admin`       | `conceptEntry`, `termSec`            | string                                                                                                                                                                                                                                                                                     |
| `responsibility`    | `transacNote` | `conceptEntry`, `langSec`, `termSec` | string                                                                                                                                                                                                                                                                                     |
| `source`            | `admin`       | `conceptEntry`, `langSec`, `termSec` | string                                                                                                                                                                                                                                                                                     |
| `termLocation`      | `termNote`    | `termSec`                            | picklist (18 UI-element values: `checkBox`, `comboBox`, `comboBoxElement`, `dialogBox`, `groupBox`, `informativeMessage`, `interactiveMessage`, `menuItem`, `progressBar`, `pushButton`, `radioButton`, `slider`, `spinBox`, `tab`, `tableText`, `textBox`, `toolTip`, `user-definedType`) |
| `termType`          | `termNote`    | `termSec`                            | picklist: `fullForm`, `acronym`, `abbreviation`, `shortForm`, `variant`, `phrase`                                                                                                                                                                                                          |
| `transactionType`   | `transac`     | `conceptEntry`, `langSec`, `termSec` | picklist: `origination`, `modification`                                                                                                                                                                                                                                                    |
| `xGraphic`          | `xref`        | `conceptEntry`, `langSec`            | URL                                                                                                                                                                                                                                                                                        |

### From the Min module (`http://www.tbxinfo.net/ns/min`)

| Category                 | DCA carrier | Levels                               | Value type                                                                                                        |
| ------------------------ | ----------- | ------------------------------------ | ----------------------------------------------------------------------------------------------------------------- |
| `administrativeStatus`   | `termNote`  | `termSec`                            | picklist: `admittedTerm-admn-sts`, `deprecatedTerm-admn-sts`, `supersededTerm-admn-sts`, `preferredTerm-admn-sts` |
| `customerSubset`         | `admin`     | `conceptEntry`, `termSec`            | string                                                                                                            |
| `externalCrossReference` | `xref`      | `conceptEntry`, `langSec`, `termSec` | URL                                                                                                               |
| `partOfSpeech`           | `termNote`  | `termSec`                            | picklist: `adjective`, `noun`, `other`, `verb`, `adverb`                                                          |
| `subjectField`           | `descrip`   | `conceptEntry`                       | string                                                                                                            |

### From the Core (`urn:iso:std:iso:30042:ed-2`)

The only data categories defined directly in Core are `term`, `date`, and `note`. Everything else structural — `conceptEntry`, `langSec`, `termSec`, `transacGrp`, `descripGrp`, `adminGrp`, `termNoteGrp`, `ref`, `xref` — is also defined in Core, with no Linguist-specific additions.

### Important caveats

- **`administrativeStatus` picklist values include the `-admn-sts` suffix.** TBX-Basic documents in the wild often use bare values like `preferredTerm` or `deprecatedTerm`; per the spec these should be `preferredTerm-admn-sts`, `deprecatedTerm-admn-sts`, etc. An implementation that already accepts the short forms (as the current project does) should keep doing so but normalize on emit to the suffixed form.
- **`geographicalUsage` is free-form**, not an ISO 3166 picklist. Implementations sometimes assume the contrary.

## Structural elements

The Linguist module adds **no new structural elements**. All structural elements (`conceptEntry`, `langSec`, `termSec`, `transacGrp`, `transac`, `transacNote`, `descrip`, `descripGrp`, `descripNote`, `admin`, `adminGrp`, `adminNote`, `termNote`, `termNoteGrp`, `ref`, `xref`, `note`) are inherited from Core.

Notable consequences:

- **Transactions** (`<transacGrp>`, `<transac>`, `<transacNote>`) are Core elements but their semantic content (`transactionType`, `responsibility`) is Basic.
- **Cross-references** (`<ref>`, `<xref>`) are Core elements; the categories `crossReference`, `externalCrossReference`, and `xGraphic` live in Basic / Min.
- **Bibliographic references** have no dedicated element. Use `source` (admin) for citation, and `<note>` / `<ref>` for inline citation context.
- **Reliability codes** do not exist in any of Min, Basic, or Linguist. The reliability data categories from the legacy 2008 TBX-Default dialect did not survive into the 2019 modular dialects.
- **Figure references** are handled by `xGraphic` (Basic) — an `xref`-style URL pointing to an image resource.

## Inline markup

Linguist adds no inline markup. The TBX Core ships the XLIFF-2-aligned inline data model:

| Element     | Purpose                    | Notable attributes                                                                                           |
| ----------- | -------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `<hi>`      | Highlighting / inline term | `target`, `type` (picklist: `entailedTerm`, `hotkey`, `italics`, `bold`, `superscript`, `subscript`, `math`) |
| `<sc>`      | XLIFF 2 start code         | `id`, `isolated` (`yes`/`no`)                                                                                |
| `<ec>`      | XLIFF 2 end code           | `startRef`, `isolated`, `disp`, `equiv`, `id`, `type`, `subtype`, `target`                                   |
| `<ph>`      | XLIFF 2 placeholder        | `type`, `id`                                                                                                 |
| `<foreign>` | Foreign-language run       | `id`, `lang`                                                                                                 |

The dialect Schematron `XLIFF.inlineConstraints` pattern enforces:

- Paired `<sc>` / `<ec>` must have `isolated='no'`.
- An orphan `<ec>` must have an `@id` and `isolated='yes'`.

These constraints apply identically across TBX-Basic and TBX-Linguist.

## ID / IDREF model

Linguist introduces no new ID/IDREF relationships beyond Basic. The general rules:

- Every `@target` value must either match an `@id` declared elsewhere in the document (IDREF semantics) or be an HTTP(S) URL.
- DCA-style `@type` classification attributes are forbidden inside the body of a DCT document.
- `responsibility/@target` typically references a `<person>` entry's ID in the `<tbxHeader>`.
- `xGraphic/@target` and `externalCrossReference/@target` are URLs.

## Validation toolchain

DCT and DCA use different validation strategies — the same TBX-wide pattern, applied to Linguist:

### DCT validation

Driven by NVDL. The file `DCT/TBX-Linguist.nvdl` dispatches per namespace:

| Namespace                               | Schema                 | Schematron             |
| --------------------------------------- | ---------------------- | ---------------------- |
| `urn:iso:std:iso:30042:ed-2` (TBX core) | `TBXcoreStructV03.rng` | `TBX-Linguist_DCT.sch` |
| `http://www.tbxinfo.net/ns/min`         | `Min.rng`              | `Min.sch`              |
| `http://www.tbxinfo.net/ns/basic`       | `Basic.rng`            | `Basic.sch`            |
| `http://www.tbxinfo.net/ns/linguist`    | `Linguist.rng`         | `Linguist.sch`         |

NVDL modes: `core` (entry), `RNG`, `SCH`. Foreign namespaces are rejected via `<anyNamespace><reject/></anyNamespace>`.

### DCA validation

A single integrated RNG plus the dialect Schematron:

- `DCA/TBXcoreStructV03_TBX-Linguist_integrated.rng` — modified copy of the Core RNG with the `@type` picklists restricted to the union of Min + Basic + Linguist categories.
- `DCA/TBX-Linguist_DCA.sch` — dialect-level rules (type/style enforcement, IDREF/URL targets, XLIFF inline constraints).

### Schema availability

No XSD is published for Linguist (or for any TBX-3.0 dialect other than Core). Implementations needing XSD must generate via Trang from RNG or hand-write.

### Practical validation choices for Go

- Embed the RNG / Schematron / NVDL files from LTAC-Global in the binary, or
- Reimplement the Schematron rules natively in Go.

The latter is preferable for a single-binary tool — the rule set is small and the constraints are easily expressed as Go predicates after XML parsing.

## Practical example (DCT)

```xml
<?xml version="1.0" encoding="utf-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en"
     xmlns="urn:iso:std:iso:30042:ed-2"
     xmlns:min="http://www.tbxinfo.net/ns/min"
     xmlns:basic="http://www.tbxinfo.net/ns/basic"
     xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader>
    <fileDesc>
      <sourceDesc><p>Sample TBX-Linguist DCT file</p></sourceDesc>
    </fileDesc>
  </tbxHeader>
  <text>
    <body>
      <conceptEntry id="c42">
        <min:subjectField>astronomy</min:subjectField>
        <transacGrp>
          <basic:transactionType>origination</basic:transactionType>
          <date>2024-01-15</date>
          <basic:responsibility target="p001">Yuki</basic:responsibility>
        </transacGrp>
        <langSec xml:lang="ja">
          <termSec>
            <term>銀河団</term>
            <min:partOfSpeech>noun</min:partOfSpeech>
            <ling:grammaticalNumber>singular</ling:grammaticalNumber>
            <ling:register>technicalRegister</ling:register>
            <adminGrp>
              <ling:reading>ぎんがだん</ling:reading>
              <ling:readingNote>Hiragana reading of 銀河団.</ling:readingNote>
            </adminGrp>
            <ling:transferComment>Prefer over loanword クラスター in scientific contexts.</ling:transferComment>
          </termSec>
        </langSec>
        <langSec xml:lang="en">
          <termSec>
            <term>galaxy cluster</term>
            <min:partOfSpeech>noun</min:partOfSpeech>
            <ling:register>technicalRegister</ling:register>
            <basic:geographicalUsage>en-US</basic:geographicalUsage>
          </termSec>
        </langSec>
      </conceptEntry>
    </body>
  </text>
</tbx>
```

DCA equivalents of the Linguist-specific fragments:

```xml
<termNote type="grammaticalNumber">singular</termNote>
<termNote type="register">technicalRegister</termNote>
<termNote type="transferComment">Prefer over loanword …</termNote>
<adminGrp>
  <admin type="reading">ぎんがだん</admin>
  <adminNote type="readingNote">Hiragana reading of 銀河団.</adminNote>
</adminGrp>
```

## Open questions and ambiguities

- **DatCatInfo coverage of `reading` and `readingNote`** — neither has a DC-ID in the published Linguist module. Round-tripping to other DCR-based tools may lose these.
- **Repo staleness** — the Linguist dialect and module repos have not been updated since 2018/2019. The upcoming ISO/AWI 30042 revision may reshape the module; track that work before committing to long-lived assumptions.
- **Absence from tbxinfo.net's dialect listing** — Linguist is referenced in the MultiLingual 2019 announcement of TBX v3 but is not currently listed alongside Core/Min/Basic on tbxinfo.net. Governance is implicitly via LTAC-Global.

## Implementation checklist for this project

Building on the existing TBX-Basic support:

1.  Accept `<tbx type="TBX-Linguist">` in addition to `TBX-Basic`.
2.  Register the `http://www.tbxinfo.net/ns/linguist` namespace (prefix `ling:`).
3.  Parse the five Linguist data categories at `termSec` level:
    - `grammaticalNumber` (picklist)
    - `register` (picklist, accept legacy `usageRegister` on read)
    - `transferComment` (string)
    - `reading` (string, inside `adminGrp`)
    - `readingNote` (string, inside `adminGrp`)
4.  Recognize both DCA and DCT encodings for each of the five.
5.  Enforce the Schematron placement rules (parent must be `termSec` or `adminGrp` as specified).
6.  Normalize `administrativeStatus` values: accept both short (`preferredTerm`) and suffixed (`preferredTerm-admn-sts`) forms on read.
7.  Expand `internal/tbx` model types (`Term`, `LangSection`, etc.) to carry the new fields, or use an open metadata map to keep the surface stable.

## Sources

- <https://github.com/LTAC-Global/TBX-Linguist_dialect>
- <https://github.com/LTAC-Global/TBX_linguist_module>
- `Linguist.tbxmd`, `Linguist.rng`, `Linguist.sch` from the module repo
- `DCT/TBX-Linguist.nvdl`, `DCT/TBX-Linguist_DCT.sch`, `DCT/Example_Astronomy_DCT_VALID.tbx` from the dialect repo
- `DCA/Example_Astronomy_DCA_VALID.tbx` from the dialect repo
- <https://www.tbxinfo.net/tbx-dialects/>
- <https://www.tbxinfo.net/tbx-modules/>
- <https://www.tbxinfo.net/developer-resources/tbx-elements-reference/>
- <https://multilingual.com/issues/july-aug-2019/tbx-version-3-published-at-iso/>
- Basic and Min `.tbxmd` files (for delta comparison)
