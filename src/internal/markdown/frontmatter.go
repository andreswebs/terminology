package markdown

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

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
