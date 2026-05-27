package app_test

import (
	"testing"

	_ "github.com/andreswebs/terminology/internal/app"
	"github.com/andreswebs/terminology/internal/terr"
)

func TestSentinelRegistry_ContainsKnownCodes(t *testing.T) {
	all := terr.All()
	codes := make(map[string]bool, len(all))
	for _, e := range all {
		codes[e.Code()] = true
	}

	want := []string{
		"unsupported_dialect",
		"tbx_locked",
		"no_tbx_path",
		"validation_error",
		"conflicting_verbosity",
		"no_subcommand",
		"language_required",
	}

	for _, code := range want {
		if !codes[code] {
			t.Errorf("terr.All() missing sentinel with code %q", code)
		}
	}
}
