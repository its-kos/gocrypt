//go:build unix

package repo

import (
	"fmt"
	"os"
	"syscall"
)

// lockDir takes an exclusive lock on dir, waiting if another process (or
// goroutine) holds it. The lock is released by calling unlock, or
// automatically by the OS if the process dies, so a crash never leaves a
// stale lock behind.
func lockDir(dir string) (unlock func(), err error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("lock %s: %w", dir, err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("lock %s: %w", dir, err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}
