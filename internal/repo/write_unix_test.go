//go:build unix

package repo

import (
	"path/filepath"
	"syscall"
	"testing"
)

func TestWriteFileAtomicIgnoresUmask(t *testing.T) {
	// A strict umask would turn a requested 0644 into 0600 with os.WriteFile.
	// writeFileAtomic must set exactly the permissions it was asked for.
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)

	path := filepath.Join(t.TempDir(), "public")
	if err := writeFileAtomic(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("writeFileAtomic() error: %v", err)
	}
	if got := permOf(t, path); got != 0o644 {
		t.Fatalf("permissions = %04o, want 0644", got)
	}
}
