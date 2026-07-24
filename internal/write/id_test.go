package write

import (
	"errors"
	"strings"
	"testing"
)

func TestDeriveID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "basic latin",
			input: "Tzimtzum",
			want:  "tzimtzum",
		},
		{
			name:  "accented characters",
			input: "Razón Histórica",
			want:  "razon-historica",
		},
		{
			name:  "non-alphanumeric runs become single hyphen",
			input: "hello   ---  world",
			want:  "hello-world",
		},
		{
			name:  "already slugified",
			input: "tzimtzum",
			want:  "tzimtzum",
		},
		{
			name:    "hebrew only returns invalid_id",
			input:   "צמצום",
			wantErr: ErrInvalidID,
		},
		{
			name:    "empty string returns invalid_id",
			input:   "",
			wantErr: ErrInvalidID,
		},
		{
			name:  "mixed script keeps latin portion",
			input: "Ein Sof אין סוף",
			want:  "ein-sof",
		},
		{
			name:    "CJK only returns invalid_id",
			input:   "太極",
			wantErr: ErrInvalidID,
		},
		{
			name:  "numbers preserved",
			input: "Chapter 42",
			want:  "chapter-42",
		},
		{
			name:  "leading and trailing special chars trimmed",
			input: "---hello---",
			want:  "hello",
		},
		{
			name:  "german eszett folds",
			input: "Straße",
			want:  "strasse",
		},
		{
			name:  "truncation cuts at last hyphen within 64",
			input: "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-overflow",
			want:  "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij",
		},
		{
			name:  "truncation at clean hyphen boundary",
			input: "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefg-overflow",
			want:  "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefg",
		},
		{
			name:  "truncation no hyphen in first 64 keeps all 64",
			input: "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdextra",
			want:  "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcd",
		},
		{
			name:  "exactly 64 codepoints unchanged",
			input: "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefg",
			want:  "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeriveID(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrInvalidID(t *testing.T) {
	var coded interface {
		Code() string
		ExitCode() int
		Hint() string
	} = ErrInvalidID

	if coded.Code() != "invalid_id" {
		t.Errorf("code: got %q, want %q", coded.Code(), "invalid_id")
	}
	if coded.ExitCode() != 65 {
		t.Errorf("exit code: got %d, want %d", coded.ExitCode(), 65)
	}
	if coded.Hint() == "" {
		t.Error("hint should not be empty")
	}
}

func TestDeriveID_Idempotence(t *testing.T) {
	inputs := []string{
		"Tzimtzum",
		"Razón Histórica",
		"Ein Sof אין סוף",
		"hello   ---  world",
		"Straße",
		"Chapter 42",
		strings.Repeat("abcdefghij-", 7) + "overflow",
		"already-slugified-id",
	}

	for _, input := range inputs {
		first, err := DeriveID(input)
		if err != nil {
			continue
		}
		second, err := DeriveID(first)
		if err != nil {
			t.Errorf("DeriveID(%q) succeeded but DeriveID(%q) failed: %v", input, first, err)
			continue
		}
		if first != second {
			t.Errorf("not idempotent for %q: DeriveID→%q, DeriveID→DeriveID→%q", input, first, second)
		}
	}
}
