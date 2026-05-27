package tbx

type Style int

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

type Dialect string

const DialectLinguist Dialect = "TBX-Linguist"
