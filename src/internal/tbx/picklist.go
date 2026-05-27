package tbx

func Format() []string { return []string{"json", "text"} }

func AdminStatus() []string {
	return []string{
		"preferredTerm-admn-sts", "admittedTerm-admn-sts",
		"deprecatedTerm-admn-sts", "supersededTerm-admn-sts",
		"preferredTerm", "admittedTerm", "deprecatedTerm", "supersededTerm",
	}
}

func PartOfSpeech() []string {
	return []string{"noun", "verb", "adjective", "adverb", "other"}
}

func GrammaticalGender() []string {
	return []string{"masculine", "feminine", "neuter", "other"}
}

func Register() []string {
	return []string{
		"colloquialRegister", "neutralRegister", "technicalRegister",
		"in-houseRegister", "bench-levelRegister", "slangRegister",
		"vulgarRegister",
		"usageRegister",
	}
}

func GrammaticalNumber() []string {
	return []string{"singular", "plural", "dual", "mass", "otherNumber"}
}

func TermType() []string {
	return []string{"fullForm", "acronym", "abbreviation", "shortForm", "variant", "phrase"}
}

func TransactionType() []string {
	return []string{"origination", "modification"}
}

func Script() []string {
	return []string{"latin", "hebrew", "cyrillic", "arabic", "any"}
}
