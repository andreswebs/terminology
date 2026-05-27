package write

import (
	"errors"
	"testing"
	"unicode/utf8"
)

func FuzzDeriveID(f *testing.F) {
	f.Add("tzimtzum")
	f.Add("Razón Histórica")
	f.Add("café")
	f.Add("über")
	f.Add("צמצום")
	f.Add("")
	f.Add("日本語テスト")
	f.Add("hello-world")
	f.Add("a")
	f.Add("---")
	f.Add("123")
	f.Add("a" + string(make([]byte, 200)))
	f.Add("\x00\x01\x02")
	f.Add("ß")

	f.Fuzz(func(t *testing.T, input string) {
		id, err := DeriveID(input)

		if err != nil {
			if !errors.Is(err, ErrInvalidID) {
				t.Fatalf("unexpected error type: %v", err)
			}
			return
		}

		if id == "" {
			t.Fatal("DeriveID returned empty string with nil error")
		}

		if !utf8.ValidString(id) {
			t.Fatalf("DeriveID returned invalid UTF-8: %q", id)
		}

		if len([]rune(id)) > 64 {
			t.Fatalf("DeriveID returned ID longer than 64 codepoints: %d", len([]rune(id)))
		}

		for _, r := range id {
			if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
				t.Fatalf("DeriveID returned ID with invalid character %q: %q", r, id)
			}
		}

		if id[0] == '-' || id[len(id)-1] == '-' {
			t.Fatalf("DeriveID returned ID with leading/trailing hyphen: %q", id)
		}
	})
}
