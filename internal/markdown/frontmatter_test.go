package markdown

import "testing"

func TestFrontmatterLang_BasicExtraction(t *testing.T) {
	src := []byte("---\nlang: es\n---\nHello world")
	got := FrontmatterLang(src)
	if got != "es" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "es")
	}
}

func TestFrontmatterLang_NoFrontmatter(t *testing.T) {
	src := []byte("Hello world")
	got := FrontmatterLang(src)
	if got != "" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "")
	}
}

func TestFrontmatterLang_UnclosedFrontmatter(t *testing.T) {
	src := []byte("---\nlang: de\nHello")
	got := FrontmatterLang(src)
	if got != "" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "")
	}
}

func TestFrontmatterLang_QuotedValue(t *testing.T) {
	src := []byte("---\nlang: \"pt-BR\"\n---\n")
	got := FrontmatterLang(src)
	if got != "pt-BR" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "pt-BR")
	}
}

func TestFrontmatterLang_SingleQuotedValue(t *testing.T) {
	src := []byte("---\nlang: 'zh-TW'\n---\n")
	got := FrontmatterLang(src)
	if got != "zh-TW" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "zh-TW")
	}
}

func TestFrontmatterLang_WithOtherKeys(t *testing.T) {
	src := []byte("---\ntitle: Chapter 1\nlang: fr\nauthor: Test\n---\nContent")
	got := FrontmatterLang(src)
	if got != "fr" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "fr")
	}
}

func TestFrontmatterLang_EmptyLangValue(t *testing.T) {
	src := []byte("---\nlang:\n---\nContent")
	got := FrontmatterLang(src)
	if got != "" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "")
	}
}

func TestFrontmatterLang_NoFrontmatterDelimiter(t *testing.T) {
	src := []byte("lang: es\nJust a plain file")
	got := FrontmatterLang(src)
	if got != "" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "")
	}
}

func TestFrontmatterLang_EmptyInput(t *testing.T) {
	got := FrontmatterLang([]byte{})
	if got != "" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "")
	}
}

func TestFrontmatterLang_WindowsLineEndings(t *testing.T) {
	src := []byte("---\r\nlang: pt\r\n---\r\nContent")
	got := FrontmatterLang(src)
	if got != "pt" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "pt")
	}
}

func TestFrontmatterLang_InlineComment(t *testing.T) {
	src := []byte("---\nlang: he # Hebrew\n---\nContent")
	got := FrontmatterLang(src)
	if got != "he" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "he")
	}
}

func TestFrontmatterLang_NilInput(t *testing.T) {
	got := FrontmatterLang(nil)
	if got != "" {
		t.Errorf("FrontmatterLang = %q, want %q", got, "")
	}
}
