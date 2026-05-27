package tbx

import (
	"os"

	"github.com/rogpeppe/go-internal/lockedfile"
)

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
