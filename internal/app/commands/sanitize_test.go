package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andreswebs/terminology/internal/terr"
)

func assertSanitizeError(t *testing.T, err error, wantCode string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	coded, ok := err.(terr.Coded)
	if !ok {
		t.Fatalf("expected terr.Coded error, got %T: %v", err, err)
	}
	if coded.Code() != wantCode {
		t.Errorf("code = %q, want %q", coded.Code(), wantCode)
	}
	if coded.ExitCode() != 65 {
		t.Errorf("exit code = %d, want 65", coded.ExitCode())
	}
}

func TestSanitizeTerm(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "plain_ascii", input: "tzimtzum", wantErr: false},
		{name: "hebrew", input: "צמצום", wantErr: false},
		{name: "accented_latin", input: "Razón Histórica", wantErr: false},
		{name: "with_spaces", input: "hello world", wantErr: false},
		{name: "with_punctuation", input: "it's fine!", wantErr: false},
		{name: "with_hyphens", input: "self-referential", wantErr: false},
		{name: "with_newline", input: "hello\nworld", wantErr: true},
		{name: "with_tab", input: "hello\tworld", wantErr: true},
		{name: "with_null", input: "hello\x00world", wantErr: true},
		{name: "with_carriage_return", input: "hello\rworld", wantErr: true},
		{name: "with_escape", input: "hello\x1bworld", wantErr: true},
		{name: "empty", input: "", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizeTerm(tt.input)
			if tt.wantErr {
				assertSanitizeError(t, err, "invalid_term")
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSanitizeConceptID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "simple", input: "tzimtzum", wantErr: false},
		{name: "with_hyphen", input: "razon-historica", wantErr: false},
		{name: "with_digits", input: "concept-001", wantErr: false},
		{name: "single_char", input: "a", wantErr: false},

		{name: "control_null", input: "foo\x00bar", wantErr: true},
		{name: "control_tab", input: "foo\tbar", wantErr: true},
		{name: "control_newline", input: "foo\nbar", wantErr: true},

		{name: "path_traversal_dotdot", input: "../etc/passwd", wantErr: true},
		{name: "path_traversal_middle", input: "foo/../bar", wantErr: true},
		{name: "path_traversal_end", input: "foo/..", wantErr: true},

		{name: "percent_encoded_slash", input: "foo%2fbar", wantErr: true},
		{name: "percent_encoded_dot", input: "%2e%2e/etc", wantErr: true},
		{name: "percent_encoded_upper", input: "foo%2Fbar", wantErr: true},
		{name: "bare_percent", input: "foo%bar", wantErr: true},

		{name: "question_mark", input: "foo?bar=1", wantErr: true},
		{name: "hash", input: "foo#section", wantErr: true},

		{name: "single_dot_ok", input: "foo.bar", wantErr: false},
		{name: "empty", input: "", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizeConceptID(tt.input)
			if tt.wantErr {
				assertSanitizeError(t, err, "invalid_id")
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSanitizeLangTag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "en", input: "en", wantErr: false},
		{name: "es", input: "es", wantErr: false},
		{name: "he", input: "he", wantErr: false},
		{name: "pt_BR", input: "pt-BR", wantErr: false},
		{name: "zh_TW", input: "zh-TW", wantErr: false},
		{name: "en_US", input: "en-US", wantErr: false},
		{name: "three_letter", input: "yue", wantErr: false},

		{name: "empty", input: "", wantErr: true},
		{name: "single_char", input: "a", wantErr: true},
		{name: "too_long_primary", input: "abcdefghi", wantErr: true},
		{name: "digits_primary", input: "12", wantErr: true},
		{name: "empty_subtag", input: "en-", wantErr: true},
		{name: "subtag_too_long", input: "en-abcdefghi", wantErr: true},

		{name: "control_null", input: "en\x00", wantErr: true},
		{name: "control_newline", input: "en\n", wantErr: true},

		{name: "percent_encoded", input: "en%2DUS", wantErr: true},
		{name: "percent_in_tag", input: "%65n", wantErr: true},

		{name: "slash", input: "en/US", wantErr: true},
		{name: "space", input: "en US", wantErr: true},
		{name: "underscore", input: "en_US", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizeLangTag(tt.input)
			if tt.wantErr {
				assertSanitizeError(t, err, "invalid_lang_tag")
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	baseDir := t.TempDir()

	subDir := filepath.Join(baseDir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existingFile := filepath.Join(subDir, "file.md")
	if err := os.WriteFile(existingFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "relative_in_base", input: "sub/file.md", wantErr: false},
		{name: "just_filename", input: "file.md", wantErr: false},
		{name: "nested", input: "a/b/c.md", wantErr: false},

		{name: "control_null", input: "file\x00.md", wantErr: true},
		{name: "control_tab", input: "file\t.md", wantErr: true},

		{name: "percent_encoded_slash", input: "sub%2ffile.md", wantErr: true},
		{name: "percent_encoded_dot", input: "%2e%2e/etc", wantErr: true},

		{name: "question_mark", input: "file.md?v=1", wantErr: true},
		{name: "hash", input: "file.md#section", wantErr: true},

		{name: "dotdot_escape", input: "../outside.md", wantErr: true},
		{name: "dotdot_middle", input: "sub/../../outside.md", wantErr: true},

		{name: "absolute_outside", input: "/etc/passwd", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizePath(tt.input, baseDir)
			if tt.wantErr {
				assertSanitizeError(t, err, "invalid_path")
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == "" {
					t.Error("expected non-empty path result")
				}
			}
		})
	}
}

func TestSanitizeTBXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "absolute_path_ok", input: "/opt/glossaries/terms.tbx", wantErr: false},
		{name: "relative_path_ok", input: "glossary/terms.tbx", wantErr: false},
		{name: "home_dir_ok", input: "/home/user/project/terms.tbx", wantErr: false},

		{name: "dotdot_rejected", input: "../etc/terms.tbx", wantErr: true},
		{name: "dotdot_middle", input: "/opt/glossary/../secret.tbx", wantErr: true},
		{name: "percent_encoded", input: "/opt/glossary%2fterms.tbx", wantErr: true},
		{name: "query_param", input: "terms.tbx?v=1", wantErr: true},
		{name: "hash", input: "terms.tbx#section", wantErr: true},
		{name: "control_char", input: "terms\x00.tbx", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizeTBXPath(tt.input)
			if tt.wantErr {
				assertSanitizeError(t, err, "invalid_path")
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSanitizePath_SymlinkEscape(t *testing.T) {
	baseDir := t.TempDir()
	outsideDir := t.TempDir()

	outsideFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(baseDir, "escape-link")
	if err := os.Symlink(outsideDir, link); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	_, err := sanitizePath("escape-link/secret.txt", baseDir)
	assertSanitizeError(t, err, "invalid_path")
}

func TestSanitizePath_ReturnsAbsolute(t *testing.T) {
	baseDir := t.TempDir()
	subDir := filepath.Join(baseDir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := sanitizePath("sub", baseDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got %q", result)
	}
	// Resolve baseDir through EvalSymlinks to match what sanitizePath returns
	// (e.g. /var → /private/var on macOS).
	resolvedBase, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(resolvedBase, "sub")
	if result != expected {
		t.Errorf("result = %q, want %q", result, expected)
	}
}
