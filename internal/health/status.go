// Package health checks a node for problems.
package health

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"

	"github.com/its-kos/gocrypt/internal/repo"
)

// Check returns every problem with the node, or nil if there are none. It
// never asks for a passphrase, so an encrypted key is not checked.
func Check(info *repo.Info) []error {
	var problems []error

	// Windows has no Unix permission bits.
	if runtime.GOOS != "windows" {
		if err := checkPrivate(info.Home, "node folder", "700"); err != nil {
			problems = append(problems, err)
		}
		if err := checkPrivate(info.KeyPath(), "key file", "600"); err != nil {
			problems = append(problems, err)
		}
	}

	encrypted, err := info.KeyEncrypted()
	if err != nil {
		problems = append(problems, err)
	} else if !encrypted {
		if _, err := info.LoadIdentity(nil); err != nil {
			problems = append(problems, err)
		}
	}

	if err := checkStorage(info.StoragePath); err != nil {
		problems = append(problems, err)
	}

	return problems
}

// checkPrivate fails if anyone other than the owner can access path.
func checkPrivate(path, what, mode string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%s %s can be accessed by other users; fix with: chmod %s %s", what, path, mode, path)
	}
	return nil
}

// checkStorage fails if the storage folder is missing or not a folder.
func checkStorage(path string) error {
	fi, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("storage folder %s does not exist (if it's on another disk, is it mounted?)", path)
	}
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("storage path %s is not a folder", path)
	}
	return nil
}
