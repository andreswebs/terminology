package tbx_test

import (
	"slices"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
)

func TestPicklists_NonEmpty_UniqueElements(t *testing.T) {
	lists := map[string]func() []string{
		"Format":            tbx.Format,
		"AdminStatus":       tbx.AdminStatus,
		"PartOfSpeech":      tbx.PartOfSpeech,
		"GrammaticalGender": tbx.GrammaticalGender,
		"Register":          tbx.Register,
		"GrammaticalNumber": tbx.GrammaticalNumber,
		"TermType":          tbx.TermType,
		"TransactionType":   tbx.TransactionType,
		"Script":            tbx.Script,
	}

	for name, fn := range lists {
		t.Run(name, func(t *testing.T) {
			vals := fn()
			if len(vals) == 0 {
				t.Errorf("%s() returned empty slice", name)
			}
			seen := make(map[string]bool)
			for _, v := range vals {
				if seen[v] {
					t.Errorf("%s() contains duplicate %q", name, v)
				}
				seen[v] = true
			}
		})
	}
}

func TestPicklists_ContainCanonicalValues(t *testing.T) {
	cases := []struct {
		name string
		fn   func() []string
		want []string
	}{
		{"Format", tbx.Format, []string{"json", "text"}},
		{"AdminStatus", tbx.AdminStatus, []string{"preferredTerm-admn-sts", "admittedTerm-admn-sts", "deprecatedTerm-admn-sts", "supersededTerm-admn-sts"}},
		{"PartOfSpeech", tbx.PartOfSpeech, []string{"noun", "verb", "adjective"}},
		{"GrammaticalGender", tbx.GrammaticalGender, []string{"masculine", "feminine", "neuter"}},
		{"Register", tbx.Register, []string{"colloquialRegister", "neutralRegister", "technicalRegister"}},
		{"GrammaticalNumber", tbx.GrammaticalNumber, []string{"singular", "plural", "dual"}},
		{"TermType", tbx.TermType, []string{"fullForm", "acronym", "abbreviation"}},
		{"TransactionType", tbx.TransactionType, []string{"origination", "modification"}},
		{"Script", tbx.Script, []string{"latin", "hebrew", "any"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vals := tc.fn()
			for _, w := range tc.want {
				if !slices.Contains(vals, w) {
					t.Errorf("%s() missing canonical value %q", tc.name, w)
				}
			}
		})
	}
}

func TestAdminStatus_IncludesLegacyBareForms(t *testing.T) {
	vals := tbx.AdminStatus()
	legacy := []string{"preferredTerm", "admittedTerm", "deprecatedTerm", "supersededTerm"}
	for _, l := range legacy {
		if !slices.Contains(vals, l) {
			t.Errorf("AdminStatus() missing legacy bare form %q", l)
		}
	}
}

func TestPicklists_ReturnFreshSlices(t *testing.T) {
	a := tbx.AdminStatus()
	b := tbx.AdminStatus()
	if &a[0] == &b[0] {
		t.Error("AdminStatus() returns the same backing array on repeated calls")
	}
}
