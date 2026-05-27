package linguist

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
)

func TestDCAtoCanonicalDCT(t *testing.T) {
	dcaPairs := []struct {
		dca string
		dct string
	}{
		{"testdata/canonical/minimal-dca.tbx", "testdata/canonical/minimal-dct.tbx"},
		{"testdata/canonical/all-categories-dca.tbx", "testdata/canonical/all-categories-dct.tbx"},
	}

	for _, pair := range dcaPairs {
		name := filepath.Base(pair.dca)
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(pair.dca)
			if err != nil {
				t.Fatalf("open fixture: %v", err)
			}
			defer func() { _ = f.Close() }()

			r := NewReader()
			g, _, err := r.Decode(f)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}

			expected, err := os.ReadFile(pair.dct)
			if err != nil {
				t.Fatalf("read expected: %v", err)
			}

			var buf bytes.Buffer
			w := NewWriter()
			if err := w.Encode(&buf, g); err != nil {
				t.Fatalf("encode: %v", err)
			}

			if !bytes.Equal(expected, buf.Bytes()) {
				t.Errorf("DCA→DCT mismatch:\n--- expected ---\n%s\n--- got ---\n%s",
					string(expected), buf.String())
			}
		})
	}
}

func TestDCAModelEquivalence(t *testing.T) {
	pairs := []struct {
		dca string
		dct string
	}{
		{"testdata/canonical/minimal-dca.tbx", "testdata/canonical/minimal-dct.tbx"},
		{"testdata/canonical/all-categories-dca.tbx", "testdata/canonical/all-categories-dct.tbx"},
	}

	for _, pair := range pairs {
		name := filepath.Base(pair.dca)
		t.Run(name, func(t *testing.T) {
			dcaFile, err := os.Open(pair.dca)
			if err != nil {
				t.Fatalf("open DCA fixture: %v", err)
			}
			defer func() { _ = dcaFile.Close() }()

			dctFile, err := os.Open(pair.dct)
			if err != nil {
				t.Fatalf("open DCT fixture: %v", err)
			}
			defer func() { _ = dctFile.Close() }()

			r := NewReader()
			dcaModel, _, err := r.Decode(dcaFile)
			if err != nil {
				t.Fatalf("decode DCA: %v", err)
			}

			r2 := NewReader()
			dctModel, _, err := r2.Decode(dctFile)
			if err != nil {
				t.Fatalf("decode DCT: %v", err)
			}

			// Normalize Style so the comparison focuses on content
			dcaModel.Style = tbx.StyleDCT

			if !reflect.DeepEqual(dcaModel, dctModel) {
				t.Errorf("DCA and DCT models differ for %s", name)
			}
		})
	}
}

func TestLegacyNormalization_WriteCanonical(t *testing.T) {
	f, err := os.Open("testdata/normalized/legacy-forms.tbx")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	r := NewReader()
	g, _, err := r.Decode(f)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	var buf bytes.Buffer
	w := NewWriter()
	if err := w.Encode(&buf, g); err != nil {
		t.Fatalf("encode: %v", err)
	}

	output := buf.String()

	normalizedStatuses := []string{
		"preferredTerm-admn-sts",
		"admittedTerm-admn-sts",
		"deprecatedTerm-admn-sts",
		"supersededTerm-admn-sts",
	}
	for _, status := range normalizedStatuses {
		if !strings.Contains(output, status) {
			t.Errorf("output missing normalized status %q", status)
		}
	}

	bareStatuses := []string{
		">preferredTerm<",
		">admittedTerm<",
		">deprecatedTerm<",
		">supersededTerm<",
	}
	for _, bare := range bareStatuses {
		if strings.Contains(output, bare) {
			t.Errorf("output contains bare (non-normalized) status: %s", bare)
		}
	}

	if strings.Contains(output, "usageRegister") {
		t.Error("output contains legacy usageRegister instead of normalized register")
	}
}

func TestRoundTrip_Stability(t *testing.T) {
	fixtures, err := filepath.Glob("testdata/canonical/*.tbx")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	for _, path := range fixtures {
		name := filepath.Base(path)
		if strings.HasSuffix(name, "-dca.tbx") {
			continue
		}
		t.Run(name, func(t *testing.T) {
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}

			r1 := NewReader()
			g1, _, err := r1.Decode(bytes.NewReader(original))
			if err != nil {
				t.Fatalf("decode pass 1: %v", err)
			}

			var buf1 bytes.Buffer
			w1 := NewWriter()
			if err := w1.Encode(&buf1, g1); err != nil {
				t.Fatalf("encode pass 1: %v", err)
			}

			r2 := NewReader()
			g2, _, err := r2.Decode(bytes.NewReader(buf1.Bytes()))
			if err != nil {
				t.Fatalf("decode pass 2: %v", err)
			}

			var buf2 bytes.Buffer
			w2 := NewWriter()
			if err := w2.Encode(&buf2, g2); err != nil {
				t.Fatalf("encode pass 2: %v", err)
			}

			if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
				t.Errorf("output not stable across two round-trips for %s", name)
			}
		})
	}
}

func TestRoundTrip_Canonical(t *testing.T) {
	fixtures, err := filepath.Glob("testdata/canonical/*.tbx")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no canonical fixtures found")
	}

	for _, path := range fixtures {
		name := filepath.Base(path)
		// Only round-trip DCT files (DCA→write produces DCT, not DCA)
		if strings.HasSuffix(name, "-dca.tbx") {
			continue
		}
		t.Run(name, func(t *testing.T) {
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}

			r := NewReader()
			g, _, err := r.Decode(bytes.NewReader(original))
			if err != nil {
				t.Fatalf("decode: %v", err)
			}

			var buf bytes.Buffer
			w := NewWriter()
			if err := w.Encode(&buf, g); err != nil {
				t.Fatalf("encode: %v", err)
			}

			if !bytes.Equal(original, buf.Bytes()) {
				t.Errorf("round-trip mismatch:\n--- original ---\n%s\n--- written ---\n%s",
					string(original), buf.String())
			}
		})
	}
}
