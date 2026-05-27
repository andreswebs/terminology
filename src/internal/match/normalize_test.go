package match

import (
	"testing"
)

func TestNormalize_NFC(t *testing.T) {
	// e + combining acute (U+0301) → NFC é
	decomposed := []byte("cafe\xcc\x81")

	p := Policy{Normalize: NFC}
	c := Normalize(decomposed, p)

	want := "café"
	if string(c.Bytes) != want {
		t.Errorf("got %q, want %q", string(c.Bytes), want)
	}
	if len(c.Map) != len(c.Bytes) {
		t.Errorf("Map length %d != Bytes length %d", len(c.Map), len(c.Bytes))
	}
}

func TestNormalize_OffsetMap_ASCII(t *testing.T) {
	p := Policy{CaseFold: true, Normalize: NFC}
	c := Normalize([]byte("ABC"), p)

	if string(c.Bytes) != "abc" {
		t.Errorf("got %q, want %q", string(c.Bytes), "abc")
	}
	wantMap := []int{0, 1, 2}
	for i, v := range c.Map {
		if v != wantMap[i] {
			t.Errorf("Map[%d] = %d, want %d", i, v, wantMap[i])
		}
	}
}

func TestNormalize_OffsetMap_NFC_Composition(t *testing.T) {
	// "e" (1 byte) + combining acute (2 bytes) → NFC "é" (2 bytes)
	// Source offsets: e=0, combining=1
	// NFC output "é" = 2 bytes, both map to segment start = 0
	src := []byte("cafe\xcc\x81")
	p := Policy{Normalize: NFC}
	c := Normalize(src, p)

	// "café" in NFC: c=0x63, a=0x61, f=0x66, é=0xC3 0xA9
	// Source: c=0, a=1, f=2, e=3, combining=4
	// Map: c→0, a→1, f→2, é[0]→3, é[1]→3
	wantMap := []int{0, 1, 2, 3, 3}
	if len(c.Map) != len(wantMap) {
		t.Fatalf("Map length %d, want %d", len(c.Map), len(wantMap))
	}
	for i, v := range c.Map {
		if v != wantMap[i] {
			t.Errorf("Map[%d] = %d, want %d", i, v, wantMap[i])
		}
	}
}

func TestNormalize_WhitespaceCollapse(t *testing.T) {
	src := []byte("hello  \n  world")
	p := Policy{Normalize: NFC}
	c := Normalize(src, p)

	if string(c.Bytes) != "hello world" {
		t.Errorf("got %q, want %q", string(c.Bytes), "hello world")
	}

	// The collapsed space should map to byte 5 (first space in "  \n  ")
	spaceIdx := 5 // position of space in "hello world"
	if c.Map[spaceIdx] != 5 {
		t.Errorf("collapsed space Map[%d] = %d, want 5", spaceIdx, c.Map[spaceIdx])
	}
}

func TestNormalize_NiqqudStrip(t *testing.T) {
	// שָׁלוֹם = shin + qamats + shin-dot + lamed + vav + holam + mem-final
	// With niqqud: שׁ ָ ל וֹ ם
	// Without: שלום
	withNiqqud := []byte("שָׁלוֹם")
	without := []byte("שלום")

	p := Policy{StripNiqqud: true, Normalize: NFC}
	c := Normalize(withNiqqud, p)

	if string(c.Bytes) != string(without) {
		t.Errorf("got %q, want %q", string(c.Bytes), string(without))
	}
}

func TestNormalize_NiqqudKept_WhenFlagFalse(t *testing.T) {
	withNiqqud := []byte("שָׁלוֹם")
	p := Policy{StripNiqqud: false, Normalize: NFC}
	c := Normalize(withNiqqud, p)

	if string(c.Bytes) != string(withNiqqud) {
		t.Errorf("niqqud should be preserved when StripNiqqud=false")
	}
}

func TestNormalize_CombinedPipeline(t *testing.T) {
	// Hebrew text with niqqud + mixed case Latin + whitespace
	// "Hello  שָׁלוֹם  World" → "hello שלום world"
	src := []byte("Hello  שָׁלוֹם  World")
	p := Policy{CaseFold: true, StripNiqqud: true, Normalize: NFC}
	c := Normalize(src, p)

	want := "hello שלום world"
	if string(c.Bytes) != want {
		t.Errorf("got %q, want %q", string(c.Bytes), want)
	}

	// Verify map invariants
	if len(c.Map) != len(c.Bytes) {
		t.Errorf("Map length %d != Bytes length %d", len(c.Map), len(c.Bytes))
	}
	for i := 1; i < len(c.Map); i++ {
		if c.Map[i] < c.Map[i-1] {
			t.Errorf("Map not monotonically non-decreasing at %d: %d < %d", i, c.Map[i], c.Map[i-1])
		}
	}
}

func TestNormalize_Empty(t *testing.T) {
	p := Policy{CaseFold: true, StripNiqqud: true, Normalize: NFC}
	c := Normalize(nil, p)

	if len(c.Bytes) != 0 {
		t.Errorf("Bytes should be empty, got %q", string(c.Bytes))
	}
	if len(c.Map) != 0 {
		t.Errorf("Map should be empty, got %v", c.Map)
	}

	c2 := Normalize([]byte{}, p)
	if len(c2.Bytes) != 0 {
		t.Errorf("Bytes should be empty for empty slice, got %q", string(c2.Bytes))
	}
}

func TestNormalize_CaseFold(t *testing.T) {
	p := Policy{CaseFold: true, Normalize: NFC}
	c := Normalize([]byte("Tzimtzum"), p)

	if string(c.Bytes) != "tzimtzum" {
		t.Errorf("got %q, want %q", string(c.Bytes), "tzimtzum")
	}
	if len(c.Map) != len(c.Bytes) {
		t.Errorf("Map length %d != Bytes length %d", len(c.Map), len(c.Bytes))
	}
}

func TestNormalize_CaseFold_Eszett(t *testing.T) {
	// ß case-folds to "ss" (one rune → two runes)
	p := Policy{CaseFold: true, Normalize: NFC}
	c := Normalize([]byte("Straße"), p)

	if string(c.Bytes) != "strasse" {
		t.Errorf("got %q, want %q", string(c.Bytes), "strasse")
	}
	// Both 's' bytes from ß should map to the source offset of ß
	ssStart := 4 // "Stra" = 4 bytes, then ß at offset 4
	sIdx := 4    // "stra" = positions 0-3, then "ss" at 4-5
	if c.Map[sIdx] != ssStart || c.Map[sIdx+1] != ssStart {
		t.Errorf("ß expansion: Map[%d]=%d Map[%d]=%d, want both %d",
			sIdx, c.Map[sIdx], sIdx+1, c.Map[sIdx+1], ssStart)
	}
}

func TestNormalize_MonotonicallyNonDecreasing(t *testing.T) {
	inputs := []string{
		"Hello World",
		"café  latte",
		"שָׁלוֹם  עוֹלָם",
		"NFKD test: ﬁ",
		"mixed  \t\n  spaces",
	}
	for _, in := range inputs {
		p := Policy{CaseFold: true, StripNiqqud: true, Normalize: NFC}
		c := Normalize([]byte(in), p)
		for i := 1; i < len(c.Map); i++ {
			if c.Map[i] < c.Map[i-1] {
				t.Errorf("input %q: Map[%d]=%d < Map[%d]=%d",
					in, i, c.Map[i], i-1, c.Map[i-1])
			}
		}
	}
}

func TestNormalize_MapLengthEqualsBytes(t *testing.T) {
	inputs := []string{
		"",
		"a",
		"Hello World",
		"שָׁלוֹם",
		"café  \n  latte",
	}
	for _, in := range inputs {
		p := Policy{CaseFold: true, StripNiqqud: true, Normalize: NFC}
		c := Normalize([]byte(in), p)
		if len(c.Map) != len(c.Bytes) {
			t.Errorf("input %q: Map length %d != Bytes length %d",
				in, len(c.Map), len(c.Bytes))
		}
	}
}

func TestNormalize_NFKD(t *testing.T) {
	// ﬁ (U+FB01 LATIN SMALL LIGATURE FI) → "fi" under NFKD
	p := Policy{Normalize: NFKD}
	c := Normalize([]byte("ﬁnd"), p)

	if string(c.Bytes) != "find" {
		t.Errorf("got %q, want %q", string(c.Bytes), "find")
	}
}

func TestNormalize_FoldDiacritics(t *testing.T) {
	p := Policy{FoldDiacritics: true, Normalize: NFC}
	c := Normalize([]byte("razón"), p)

	if string(c.Bytes) != "razon" {
		t.Errorf("got %q, want %q", string(c.Bytes), "razon")
	}
}

func TestNormalize_WhitespaceOnly(t *testing.T) {
	p := Policy{Normalize: NFC}
	c := Normalize([]byte("   \t\n  "), p)

	if string(c.Bytes) != " " {
		t.Errorf("got %q, want single space", string(c.Bytes))
	}
	if c.Map[0] != 0 {
		t.Errorf("Map[0] = %d, want 0", c.Map[0])
	}
}
