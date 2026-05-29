package linguist

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
