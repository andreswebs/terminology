package extract

import "testing"

func TestDetectLang_FrontmatterPresent(t *testing.T) {
	src := []byte("---\nlang: es\n---\nHello world")
	got := DetectLang(src, "")
	if got != "es" {
		t.Errorf("DetectLang = %q, want %q", got, "es")
	}
}

func TestDetectLang_FlagFallback(t *testing.T) {
	src := []byte("Hello world without frontmatter")
	got := DetectLang(src, "he")
	if got != "he" {
		t.Errorf("DetectLang = %q, want %q", got, "he")
	}
}

func TestDetectLang_DefaultToEn(t *testing.T) {
	src := []byte("Hello world without frontmatter")
	got := DetectLang(src, "")
	if got != "en" {
		t.Errorf("DetectLang = %q, want %q", got, "en")
	}
}

func TestDetectLang_FrontmatterOverridesFlag(t *testing.T) {
	src := []byte("---\nlang: he\n---\nContent")
	got := DetectLang(src, "es")
	if got != "he" {
		t.Errorf("DetectLang = %q, want %q", got, "he")
	}
}

func TestDetectLang_FrontmatterWithOtherKeys(t *testing.T) {
	src := []byte("---\ntitle: Chapter 1\nlang: fr\nauthor: Test\n---\nContent")
	got := DetectLang(src, "")
	if got != "fr" {
		t.Errorf("DetectLang = %q, want %q", got, "fr")
	}
}

func TestDetectLang_UnclosedFrontmatter(t *testing.T) {
	src := []byte("---\nlang: es\nContent without closing")
	got := DetectLang(src, "de")
	if got != "de" {
		t.Errorf("DetectLang = %q, want %q (unclosed frontmatter should be ignored)", got, "de")
	}
}

func TestDetectLang_EmptyLangValue(t *testing.T) {
	src := []byte("---\nlang:\n---\nContent")
	got := DetectLang(src, "he")
	if got != "he" {
		t.Errorf("DetectLang = %q, want %q (empty lang should fall through)", got, "he")
	}
}

func TestDetectLang_NoFrontmatterDelimiter(t *testing.T) {
	src := []byte("lang: es\nJust a plain file")
	got := DetectLang(src, "")
	if got != "en" {
		t.Errorf("DetectLang = %q, want %q (no frontmatter delimiter)", got, "en")
	}
}

func TestDetectLang_EmptyInput(t *testing.T) {
	got := DetectLang([]byte{}, "")
	if got != "en" {
		t.Errorf("DetectLang = %q, want %q", got, "en")
	}
}

func TestDetectLang_WindowsLineEndings(t *testing.T) {
	src := []byte("---\r\nlang: pt\r\n---\r\nContent")
	got := DetectLang(src, "")
	if got != "pt" {
		t.Errorf("DetectLang = %q, want %q", got, "pt")
	}
}
