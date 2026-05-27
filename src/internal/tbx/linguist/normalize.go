package linguist

import "github.com/andreswebs/terminology/internal/tbx"

func normalizeStatus(s string) tbx.Status {
	switch s {
	case "preferredTerm-admn-sts", "preferredTerm":
		return tbx.StatusPreferred
	case "admittedTerm-admn-sts", "admittedTerm":
		return tbx.StatusAdmitted
	case "deprecatedTerm-admn-sts", "deprecatedTerm":
		return tbx.StatusDeprecated
	case "supersededTerm-admn-sts", "supersededTerm":
		return tbx.StatusSuperseded
	default:
		return tbx.StatusUnspecified
	}
}

func isLegacyStatus(s string) bool {
	switch s {
	case "preferredTerm", "admittedTerm", "deprecatedTerm", "supersededTerm":
		return true
	}
	return false
}

func normalizeRegister(s string) string {
	if s == "usageRegister" {
		return "register"
	}
	return s
}

func isLegacyRegister(s string) bool {
	return s == "usageRegister"
}
