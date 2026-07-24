package tbx

import "testing"

func TestStyleString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		style Style
		want  string
	}{
		{StyleDCT, "dct"},
		{StyleDCA, "dca"},
		{Style(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.style.String(); got != tt.want {
			t.Errorf("Style(%d).String() = %q, want %q", tt.style, got, tt.want)
		}
	}
}

func TestStyleIotaValues(t *testing.T) {
	t.Parallel()

	if StyleDCT != 0 {
		t.Errorf("StyleDCT = %d, want 0", StyleDCT)
	}
	if StyleDCA != 1 {
		t.Errorf("StyleDCA = %d, want 1", StyleDCA)
	}
}

func TestDialectLinguist(t *testing.T) {
	t.Parallel()

	if DialectLinguist != Dialect("TBX-Linguist") {
		t.Errorf("DialectLinguist = %q, want %q", DialectLinguist, "TBX-Linguist")
	}
}

func TestWarningFields(t *testing.T) {
	t.Parallel()

	w := Warning{
		Code:      "legacy_form_normalized",
		Message:   "bare preferredTerm normalized to preferredTerm-admn-sts",
		ConceptID: "c001",
		Line:      10,
		Col:       5,
	}

	if w.Code != "legacy_form_normalized" {
		t.Errorf("Code = %q, want %q", w.Code, "legacy_form_normalized")
	}
	if w.Message != "bare preferredTerm normalized to preferredTerm-admn-sts" {
		t.Errorf("Message = %q, want %q", w.Message, "bare preferredTerm normalized to preferredTerm-admn-sts")
	}
	if w.ConceptID != "c001" {
		t.Errorf("ConceptID = %q, want %q", w.ConceptID, "c001")
	}
	if w.Line != 10 {
		t.Errorf("Line = %d, want 10", w.Line)
	}
	if w.Col != 5 {
		t.Errorf("Col = %d, want 5", w.Col)
	}
}

func TestWarningZeroValues(t *testing.T) {
	t.Parallel()

	var w Warning
	if w.Line != 0 || w.Col != 0 {
		t.Errorf("zero Warning: Line=%d, Col=%d, want 0, 0", w.Line, w.Col)
	}
	if w.Code != "" {
		t.Errorf("zero Warning: Code=%q, want empty", w.Code)
	}
}
