package match

import "testing"

func TestPolicyFor_EmptyLang_ReturnsBaseline(t *testing.T) {
	p := PolicyFor("")
	if p != Baseline {
		t.Errorf("PolicyFor(\"\") = %+v, want %+v", p, Baseline)
	}
}

func TestBaseline_Defaults(t *testing.T) {
	if !Baseline.CaseFold {
		t.Error("Baseline.CaseFold should be true")
	}
	if Baseline.FoldDiacritics {
		t.Error("Baseline.FoldDiacritics should be false")
	}
	if Baseline.StripNiqqud {
		t.Error("Baseline.StripNiqqud should be false")
	}
	if Baseline.Normalize != NFC {
		t.Errorf("Baseline.Normalize = %v, want NFC", Baseline.Normalize)
	}
}

func TestPolicyFor_Hebrew_StripsNiqqud(t *testing.T) {
	p := PolicyFor("he")
	if !p.StripNiqqud {
		t.Error("PolicyFor(\"he\").StripNiqqud should be true")
	}
	if !p.CaseFold {
		t.Error("PolicyFor(\"he\").CaseFold should be true")
	}
	if p.FoldDiacritics {
		t.Error("PolicyFor(\"he\").FoldDiacritics should be false")
	}
	if p.Normalize != NFC {
		t.Errorf("PolicyFor(\"he\").Normalize = %v, want NFC", p.Normalize)
	}
}

func TestPolicyFor_UnknownLang_ReturnsBaseline(t *testing.T) {
	for _, lang := range []string{"fr", "de", "es", "ja", "unknown"} {
		p := PolicyFor(lang)
		if p != Baseline {
			t.Errorf("PolicyFor(%q) = %+v, want baseline %+v", lang, p, Baseline)
		}
	}
}

func TestForm_DistinctValues(t *testing.T) {
	if NFC == NFKD {
		t.Error("NFC and NFKD should be distinct values")
	}
}
