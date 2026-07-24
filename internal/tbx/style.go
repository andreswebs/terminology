package tbx

// Style is the TBX data category representation style used by a glossary.
type Style int

// StyleDCT and StyleDCA are the supported data category representation styles.
const (
	StyleDCT Style = iota
	StyleDCA
)

func (s Style) String() string {
	switch s {
	case StyleDCT:
		return "dct"
	case StyleDCA:
		return "dca"
	default:
		return "unknown"
	}
}

// Dialect identifies a TBX dialect.
type Dialect string

// DialectLinguist is the TBX-Linguist dialect.
const DialectLinguist Dialect = "TBX-Linguist"
