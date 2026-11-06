package repo

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeFileAtomic writes data to path so that after a crash the file holds
// either its old content or the complete new content, never a mix. The file
// gets exactly perm, regardless of the umask.
func writeFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	// The temp file must be in the same directory: rename is only atomic
	// within one filesystem. CreateTemp makes it 0600, so it is private from
	// the first byte. The hidden ".tmp-" name makes leftovers recognisable
	// if the process is killed before it can clean up.
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	tmpName := f.Name()

	// On any failure below, remove the temp file. err is the named result,
	// so this sees whatever error the function is about to return.
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tmpName)
		}
	}()

	if _, err = f.Write(data); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	// Chmod is not affected by the umask, so the file gets exactly perm.
	if err = f.Chmod(perm); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	// Make sure the contents are on disk before the rename makes them
	// visible. Otherwise a power cut can leave a correctly named, empty file.
	if err = f.Sync(); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	if err = f.Close(); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// The atomic swap: readers see the old file or the new one, nothing in between.
	if err = os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Make the rename itself survive a power cut.
	return syncDir(dir)
}

// syncDir flushes a directory's entries (e.g. a rename) to disk.
func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("sync dir: %w", err)
	}
	defer d.Close()

	if err := d.Sync(); err != nil {
		return fmt.Errorf("sync dir: %w", err)
	}
	return nil
}
