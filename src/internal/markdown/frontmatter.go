// Package markdown parses markdown source into plain-text spans and reads YAML
// frontmatter metadata.
package markdown

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

// FrontmatterLang returns the lang field from src's YAML frontmatter, or the
// empty string when there is no frontmatter or no lang field.
func FrontmatterLang(src []byte) string {
	if !bytes.HasPrefix(src, []byte("---\n")) && !bytes.HasPrefix(src, []byte("---\r\n")) {
		return ""
	}

	rest := src[4:]
	before, _, ok := bytes.Cut(rest, []byte("\n---"))
	if !ok {
		return ""
	}

	block := before

	var fm struct {
		Lang string `yaml:"lang"`
	}
	if err := yaml.Unmarshal(block, &fm); err != nil {
		return ""
	}
	return fm.Lang
}
