package tbx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andreswebs/terminology/internal/terr"
)

func TestAcquireLock_CreatesAndCleansUp(t *testing.T) {
	tmp := t.TempDir()
	lockPath := filepath.Join(tmp, "test.lock")

	unlock, err := acquireLock(lockPath)
	if err != nil {
		t.Fatalf("acquireLock: %v", err)
	}

	if _, err := os.Stat(lockPath); err != nil {
		t.Errorf("lock file should exist while held: %v", err)
	}

	unlock()

	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Errorf("lock file should be removed after cleanup, got err: %v", err)
	}
}

func TestAcquireLock_NonExistentDir(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "no-such-dir", "test.lock")

	_, err := acquireLock(lockPath)
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
	var coded terr.Coded
	if !assertCoded(t, err, &coded) {
		return
	}
	if got := coded.Code(); got != "tbx_locked" {
		t.Errorf("Code() = %q, want %q", got, "tbx_locked")
	}
}

func assertCoded(t *testing.T, err error, target *terr.Coded) bool {
	t.Helper()
	if err == nil {
		t.Error("expected non-nil error")
		return false
	}
	coded, ok := err.(terr.Coded)
	if !ok {
		t.Errorf("error %T does not implement terr.Coded", err)
		return false
	}
	*target = coded
	return true
}
