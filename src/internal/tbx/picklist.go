package tbx

// Format returns the allowed output format values.
func Format() []string { return []string{"json", "text"} }

// AdminStatus returns the allowed administrative status values.
func AdminStatus() []string {
	return []string{
		"preferredTerm-admn-sts", "admittedTerm-admn-sts",
		"deprecatedTerm-admn-sts", "supersededTerm-admn-sts",
		"preferredTerm", "admittedTerm", "deprecatedTerm", "supersededTerm",
	}
}

// PartOfSpeech returns the allowed part-of-speech values.
func PartOfSpeech() []string {
	return []string{"noun", "verb", "adjective", "adverb", "other"}
}

// GrammaticalGender returns the allowed grammatical gender values.
func GrammaticalGender() []string {
	return []string{"masculine", "feminine", "neuter", "other"}
}

// Register returns the allowed register values.
func Register() []string {
	return []string{
		"colloquialRegister", "neutralRegister", "technicalRegister",
		"in-houseRegister", "bench-levelRegister", "slangRegister",
		"vulgarRegister",
		"usageRegister",
	}
}

// GrammaticalNumber returns the allowed grammatical number values.
func GrammaticalNumber() []string {
	return []string{"singular", "plural", "dual", "mass", "otherNumber"}
}

// TermType returns the allowed term type values.
func TermType() []string {
	return []string{"fullForm", "acronym", "abbreviation", "shortForm", "variant", "phrase"}
}

// TransactionType returns the allowed transaction type values.
func TransactionType() []string {
	return []string{"origination", "modification"}
}

// Script returns the allowed script values.
func Script() []string {
	return []string{"latin", "hebrew", "cyrillic", "arabic", "any"}
}
