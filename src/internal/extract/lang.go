package extract

import "github.com/andreswebs/terminology/internal/markdown"

func DetectLang(src []byte, flagLang string) string {
	if lang := markdown.FrontmatterLang(src); lang != "" {
		return lang
	}
	if flagLang != "" {
		return flagLang
	}
	return "en"
}
