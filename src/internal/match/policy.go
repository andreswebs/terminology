package match

// Form selects the Unicode normalization form applied during canonical normalization.
type Form int

// Normalization forms selectable by a Policy.
const (
	NFC Form = iota
	NFKD
)

// Policy encodes per-axis decisions for the matcher's normalization pipeline.
type Policy struct {
	CaseFold       bool
	FoldDiacritics bool
	StripNiqqud    bool
	Normalize      Form
}

// Baseline is the default match policy: case-fold on, diacritics strict, niqqud kept, NFC.
var Baseline = Policy{CaseFold: true, FoldDiacritics: false, StripNiqqud: false, Normalize: NFC}

var byLanguage = map[string]Policy{
	"he": {CaseFold: true, FoldDiacritics: false, StripNiqqud: true, Normalize: NFC},
}

// PolicyFor returns the language-specific policy if one exists, otherwise the baseline.
func PolicyFor(lang string) Policy {
	if p, ok := byLanguage[lang]; ok {
		return p
	}
	return Baseline
}

func mergePolicy(a, b Policy) Policy {
	m := a
	if b.CaseFold {
		m.CaseFold = true
	}
	if b.FoldDiacritics {
		m.FoldDiacritics = true
	}
	if b.StripNiqqud {
		m.StripNiqqud = true
	}
	return m
}
