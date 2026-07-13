package extract

import "github.com/andreswebs/terminology/internal/markdown"

// DetectLang resolves the language of src, preferring the markdown frontmatter
// lang field, then flagLang, and defaulting to "en".
func DetectLang(src []byte, flagLang string) string {
	if lang := markdown.FrontmatterLang(src); lang != "" {
		return lang
	}
	if flagLang != "" {
		return flagLang
	}
	return "en"
}
