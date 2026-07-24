package linguist

import "testing"

// The Status string<->enum mapping is owned and tested by package tbx
// (TestParseStatus, TestStatusString); the decoder calls tbx.ParseStatus
// directly. Only the linguist-specific legacy register normalization is
// tested here.
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
