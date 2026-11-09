package repo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// Check looks for problems that would stop the node from working safely and
// returns all of them, so they can be fixed in one go. It returns nil for a
// healthy node.
//
// Check never asks for a passphrase: an encrypted key is not compared with
// the config (that happens when the node starts).
func (info *Info) Check() []error {
	var problems []error

	// Like SSH refusing a private key others can read: the folder and the key
	// must be private to the owner. Windows has no such permission bits.
	if runtime.GOOS != "windows" {
		if err := checkPrivate(info.Home, "node folder", "700"); err != nil {
			problems = append(problems, err)
		}
		if err := checkPrivate(filepath.Join(info.Home, keyFileName), "key file", "600"); err != nil {
			problems = append(problems, err)
		}
	}

	encrypted, err := info.KeyEncrypted()
	switch {
	case err != nil:
		problems = append(problems, err)
	case !encrypted:
		if _, err := info.LoadIdentity(nil); err != nil {
			problems = append(problems, err)
		}
	}

	if err := checkStorage(info.StoragePath); err != nil {
		problems = append(problems, err)
	}

	return problems
}

// checkPrivate fails if anyone but the owner can access path (any group or
// other permission bit set). The message includes the command that fixes it.
func checkPrivate(path, what, wantMode string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if perm := fi.Mode().Perm(); perm&0o077 != 0 {
		return fmt.Errorf("%s %s can be accessed by other users (mode %04o); fix with: chmod %s %s",
			what, path, perm, wantMode, path)
	}
	return nil
}

// checkStorage fails if the storage folder is missing or not a folder. A
// missing folder on an external disk usually means th
// and writing there would fill the wrong disk.
func checkStorage(path string) error {
	fi, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("storage folder %s does not exiss the disk mounted?)", path)
	}
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("storage path %s is not a folder", path)
	}
	return nil
}
