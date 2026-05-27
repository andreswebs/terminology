package extract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadStopwords_BasicFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "stopwords.txt")
	if err := os.WriteFile(path, []byte("the\na\nan\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sw, err := LoadStopwords(path)
	if err != nil {
		t.Fatalf("LoadStopwords: %v", err)
	}

	if len(sw) != 3 {
		t.Fatalf("got %d entries, want 3", len(sw))
	}
	for _, word := range []string{"the", "a", "an"} {
		if !sw[word] {
			t.Errorf("missing stopword %q", word)
		}
	}
}

func TestLoadStopwords_CommentsAndEmptyLines(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "stopwords.txt")
	content := "# comment\nthe\n\n# another comment\na\n\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	sw, err := LoadStopwords(path)
	if err != nil {
		t.Fatalf("LoadStopwords: %v", err)
	}

	if len(sw) != 2 {
		t.Fatalf("got %d entries, want 2", len(sw))
	}
	if !sw["the"] || !sw["a"] {
		t.Errorf("expected 'the' and 'a', got %v", sw)
	}
}

func TestLoadStopwords_CaseFolded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "stopwords.txt")
	if err := os.WriteFile(path, []byte("The\nAN\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sw, err := LoadStopwords(path)
	if err != nil {
		t.Fatalf("LoadStopwords: %v", err)
	}

	if !sw["the"] {
		t.Errorf("expected case-folded 'the'")
	}
	if !sw["an"] {
		t.Errorf("expected case-folded 'an'")
	}
}

func TestLoadStopwords_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := LoadStopwords("/nonexistent/stopwords.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
