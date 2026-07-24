package tbx

import (
	"os"

	"github.com/rogpeppe/go-internal/lockedfile"
)

// AcquireLock creates a lock file at lockPath and returns a release function
// that closes and removes it. It fails with ErrTBXLocked if the lock is held.
func AcquireLock(lockPath string) (func(), error) {
	return acquireLock(lockPath)
}

func acquireLock(lockPath string) (func(), error) {
	f, err := lockedfile.Create(lockPath)
	if err != nil {
		return nil, ErrTBXLocked.Wrap(err)
	}
	return func() {
		_ = f.Close()
		_ = os.Remove(lockPath)
	}, nil
}
