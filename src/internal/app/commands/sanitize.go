package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

// ErrInvalidSanitizeID reports that a concept ID contains rejected characters.
var ErrInvalidSanitizeID = terr.New(
	"invalid_id", 65,
	"concept IDs must not contain control characters, path traversals, percent-encoded segments, or query parameters",
	"concept ID contains rejected characters",
)

// ErrInvalidLangTag reports that a language tag is not well-formed BCP 47.
var ErrInvalidLangTag = terr.New(
	"invalid_lang_tag", 65,
	"language tags must be well-formed BCP 47 (e.g. en, es, he, pt-BR)",
	"language tag is invalid",
)

// ErrInvalidPath reports that a file path contains rejected characters or
// escapes the sandbox.
var ErrInvalidPath = terr.New(
	"invalid_path", 65,
	"paths must not contain .., percent-encoded segments, or query parameters",
	"file path contains rejected characters or escapes the sandbox",
)

// ErrInvalidTerm reports that a term contains rejected characters.
var ErrInvalidTerm = terr.New(
	"invalid_term", 65,
	"terms must not contain control characters",
	"term contains rejected characters",
)

func hasControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func hasPercentEncoded(s string) bool {
	return strings.Contains(s, "%")
}

func hasQueryParams(s string) bool {
	return strings.ContainsAny(s, "?#")
}

func hasPathTraversal(s string) bool {
	return strings.Contains(s, "..")
}

func wrapLoadError(err error) error {
	if _, ok := err.(interface{ Code() string }); ok {
		return err
	}
	return tbx.ErrValidationError.Wrap(err)
}

func tbxPathFromRoot(cmd *urfcli.Command) (string, error) {
	p := cmd.Root().String("tbx")
	if p == "" {
		return "", tbx.ErrNoTBXPath
	}
	if err := sanitizeTBXPath(p); err != nil {
		return "", err
	}
	return p, nil
}

func sanitizeTBXPath(s string) error {
	if hasControlChars(s) {
		return ErrInvalidPath.Wrap(fmt.Errorf("path %q contains control characters", s))
	}
	if hasPercentEncoded(s) {
		return ErrInvalidPath.Wrap(fmt.Errorf("path %q contains percent-encoded segment", s))
	}
	if hasQueryParams(s) {
		return ErrInvalidPath.Wrap(fmt.Errorf("path %q contains query parameter characters", s))
	}
	if hasPathTraversal(s) {
		return ErrInvalidPath.Wrap(fmt.Errorf("path %q contains path traversal", s))
	}
	return nil
}

func sanitizeTerm(s string) error {
	if hasControlChars(s) {
		return ErrInvalidTerm.Wrap(fmt.Errorf("term %q contains control characters", s))
	}
	return nil
}

func sanitizeConceptID(s string) error {
	if hasControlChars(s) {
		return ErrInvalidSanitizeID.Wrap(fmt.Errorf("id %q contains control characters", s))
	}
	if hasPathTraversal(s) {
		return ErrInvalidSanitizeID.Wrap(fmt.Errorf("id %q contains path traversal", s))
	}
	if hasPercentEncoded(s) {
		return ErrInvalidSanitizeID.Wrap(fmt.Errorf("id %q contains percent-encoded segment", s))
	}
	if hasQueryParams(s) {
		return ErrInvalidSanitizeID.Wrap(fmt.Errorf("id %q contains query parameter characters", s))
	}
	return nil
}

func sanitizeLangTag(s string) error {
	if hasControlChars(s) {
		return ErrInvalidLangTag.Wrap(fmt.Errorf("lang tag %q contains control characters", s))
	}
	if hasPercentEncoded(s) {
		return ErrInvalidLangTag.Wrap(fmt.Errorf("lang tag %q contains percent-encoded segment", s))
	}
	if !isWellFormedBCP47(s) {
		return ErrInvalidLangTag.Wrap(fmt.Errorf("lang tag %q is not well-formed BCP 47", s))
	}
	return nil
}

func isWellFormedBCP47(s string) bool {
	if s == "" {
		return false
	}
	parts := strings.Split(s, "-")
	primary := parts[0]
	if len(primary) < 2 || len(primary) > 8 {
		return false
	}
	for _, r := range primary {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	for _, sub := range parts[1:] {
		if len(sub) == 0 || len(sub) > 8 {
			return false
		}
		for _, r := range sub {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
				return false
			}
		}
	}
	return true
}

func sanitizePath(s string, baseDir string) (string, error) {
	if hasControlChars(s) {
		return "", ErrInvalidPath.Wrap(fmt.Errorf("path %q contains control characters", s))
	}
	if hasPercentEncoded(s) {
		return "", ErrInvalidPath.Wrap(fmt.Errorf("path %q contains percent-encoded segment", s))
	}
	if hasQueryParams(s) {
		return "", ErrInvalidPath.Wrap(fmt.Errorf("path %q contains query parameter characters", s))
	}
	if hasPathTraversal(s) {
		return "", ErrInvalidPath.Wrap(fmt.Errorf("path %q contains path traversal", s))
	}

	cleaned, err := resolveAndSandbox(s, baseDir)
	if err != nil {
		return "", ErrInvalidPath.Wrap(err)
	}
	return cleaned, nil
}

func resolveAndSandbox(s, baseDir string) (string, error) {
	// Resolve baseDir itself so symlinks like /var → /private/var
	// don't cause false sandbox escapes.
	resolvedBase, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		resolvedBase = baseDir
	}

	cleaned := filepath.Clean(s)

	var abs string
	if filepath.IsAbs(cleaned) {
		abs = cleaned
	} else {
		abs = filepath.Join(resolvedBase, cleaned)
	}

	if !strings.HasPrefix(abs, resolvedBase+string(os.PathSeparator)) && abs != resolvedBase {
		return "", fmt.Errorf("path %q resolves outside base directory %q", s, baseDir)
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return abs, nil
		}
		return "", fmt.Errorf("resolving symlinks for %q: %w", abs, err)
	}

	if !strings.HasPrefix(resolved, resolvedBase+string(os.PathSeparator)) && resolved != resolvedBase {
		return "", fmt.Errorf("path %q resolves via symlink to %q, outside base directory %q", s, resolved, baseDir)
	}

	return resolved, nil
}
