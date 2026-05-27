package linguist

import (
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
)

func TestNormalizeStatus(t *testing.T) {
	cases := []struct {
		input string
		want  tbx.Status
	}{
		{"preferredTerm-admn-sts", tbx.StatusPreferred},
		{"preferredTerm", tbx.StatusPreferred},
		{"admittedTerm-admn-sts", tbx.StatusAdmitted},
		{"admittedTerm", tbx.StatusAdmitted},
		{"deprecatedTerm-admn-sts", tbx.StatusDeprecated},
		{"deprecatedTerm", tbx.StatusDeprecated},
		{"supersededTerm-admn-sts", tbx.StatusSuperseded},
		{"supersededTerm", tbx.StatusSuperseded},
		{"unknown", tbx.StatusUnspecified},
		{"", tbx.StatusUnspecified},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := normalizeStatus(tc.input)
			if got != tc.want {
				t.Errorf("normalizeStatus(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeRegister(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"usageRegister", "register"},
		{"colloquialRegister", "colloquialRegister"},
		{"neutralRegister", "neutralRegister"},
		{"technicalRegister", "technicalRegister"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := normalizeRegister(tc.input)
			if got != tc.want {
				t.Errorf("normalizeRegister(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
