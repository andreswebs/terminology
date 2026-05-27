package linguist

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
)

func FuzzLinguistDecode(f *testing.F) {
	fixtures, err := filepath.Glob("testdata/canonical/*.tbx")
	if err == nil {
		for _, fix := range fixtures {
			data, err := os.ReadFile(fix)
			if err == nil {
				f.Add(data)
			}
		}
	}

	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<tbx type="TBX-Linguist" style="dct" xml:lang="en" xmlns="urn:iso:std:iso:30042:ed-2" xmlns:min="http://www.tbxinfo.net/ns/min" xmlns:basic="http://www.tbxinfo.net/ns/basic" xmlns:ling="http://www.tbxinfo.net/ns/linguist">
  <tbxHeader><fileDesc><sourceDesc><p>test</p></sourceDesc></fileDesc></tbxHeader>
  <text><body>
    <conceptEntry id="c1"><langSec xml:lang="en"><termSec><term>test</term></termSec></langSec></conceptEntry>
  </body></text>
</tbx>`))

	f.Add([]byte(""))
	f.Add([]byte("not xml at all"))
	f.Add([]byte("<tbx></tbx>"))
	f.Add([]byte{0x00, 0xFF, 0xFE})

	f.Fuzz(func(t *testing.T, data []byte) {
		r := NewReader()
		g, _, err := r.Decode(strings.NewReader(string(data)))

		if err != nil {
			return
		}

		if g == nil {
			t.Fatal("Decode returned nil glossary with nil error")
		}

		if g.Style != tbx.StyleDCT && g.Style != tbx.StyleDCA {
			t.Fatalf("unexpected style: %d", g.Style)
		}

		for _, c := range g.Concepts {
			if c.ID == "" {
				t.Fatal("concept with empty ID")
			}
			for lang, ls := range c.Languages {
				if lang == "" {
					t.Fatal("empty language key")
				}
				if ls.Lang != lang {
					t.Fatalf("lang mismatch: key=%q, field=%q", lang, ls.Lang)
				}
			}
		}
	})
}
