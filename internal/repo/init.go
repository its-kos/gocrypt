package repo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/its-kos/gocrypt/internal/config"
	"github.com/its-kos/gocrypt/internal/identity"
)

// InitOptions configures a new node. The zero value is valid: the hostname as
// the name, storage inside the node's folder, and an unencrypted key.
type InitOptions struct {
	Name        string // node name, defaults to the hostname
	StoragePath string // where stored data goes, defaults to <home>/storage
	// Passphrase, if set, is called to get a passphrase for encrypting
	// identity.key. It is only called when a new node is actually created,
	// so re-running init on an existing node never prompts.
	Passphrase func() ([]byte, error)
}

// errLostRace means another init created the node while we were working.
var errLostRace = errors.New("another init created the node first")

// Init creates a node in home, which must be an absolute path.
//
// If home already holds a node, Init changes nothing and returns it with
// created == false (D6). Otherwise the node is created all or nothing: after
// a crash there is either a complete node or none (D7). A broken node, a
// non-empty folder that isn't a node, or an unusable storage path is an error.
func Init(home string, opts InitOptions) (info *Info, created bool, err error) {
	if !filepath.IsAbs(home) {
		return nil, false, fmt.Errorf("init node: home must be an absolute path, got %q", home)
	}

	// Already a node? Then there is nothing to do. Any error other than
	// "not initialized" means something is there that we must not replace.
	info, err = Load(home)
	if err == nil {
		return info, false, nil
	}
	if !errors.Is(err, ErrNotInitialized) {
		return nil, false, fmt.Errorf("init node: %w", err)
	}

	// Build everything in memory first. Invalid options fail here, before
	// anything is written to disk.
	files, err := prepare(home, opts)
	if err != nil {
		return nil, false, fmt.Errorf("init node: %w", err)
	}

	homeExists, err := isExistingDir(home)
	if err != nil {
		return nil, false, fmt.Errorf("init node: %w", err)
	}

	createdStorage, err := prepareExternalStorage(home, files.storage)
	if err != nil {
		return nil, false, fmt.Errorf("init node: %w", err)
	}

	if homeExists {
		// An existing folder, e.g. a Docker volume. It can't be replaced by
		// a rename, so fill it in place. Whether it's empty is checked under
		// a lock, because another init may be writing to it right now.
		err = createInPlace(home, files)
	} else {
		err = createByRename(home, files)
	}

	if errors.Is(err, errLostRace) {
		// Another init finished first. Its node is the node; report it.
		info, err = Load(home)
		if err != nil {
			return nil, false, fmt.Errorf("init node: %w", err)
		}
		return info, false, nil
	}
	if err != nil {
		if createdStorage {
			_ = os.Remove(files.storage) // only removes it if still empty
		}
		return nil, false, fmt.Errorf("init node: %w", err)
	}

	// Read back what we wrote, so the caller gets exactly what Load sees.
	info, err = Load(home)
	if err != nil {
		return nil, false, fmt.Errorf("init node: %w", err)
	}
	return info, true, nil
}

// nodeFiles is a complete node, prepared in memory.
type nodeFiles struct {
	key     []byte // contents of identity.key
	config  []byte // contents of config.toml
	storage string // storage path as written in the config (relative or absolute)
}

// prepare generates the identity and builds both files in memory. It has no
// side effects, so all validation happens before anything touches the disk.
func prepare(home string, opts InitOptions) (*nodeFiles, error) {
	name := opts.Name
	if name == "" {
		h, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("get hostname for the node name: %w", err)
		}
		name = h
	}

	storage, err := storagePathForConfig(home, opts.StoragePath)
	if err != nil {
		return nil, err
	}

	id, err := identity.Generate()
	if err != nil {
		return nil, err
	}
	var passphrase []byte
	if opts.Passphrase != nil {
		if passphrase, err = opts.Passphrase(); err != nil {
			return nil, fmt.Errorf("get passphrase: %w", err)
		}
		if len(passphrase) == 0 {
			return nil, errors.New("the passphrase must not be empty")
		}
	}
	key, err := identity.Marshal(id, passphrase)
	if err != nil {
		return nil, err
	}

	// Marshal validates the config, so a bad name or storage path fails here.
	cfg, err := config.Marshal(&config.Config{
		FormatVersion: config.CurrentFormatVersion,
		Node:          config.Node{Name: name, ID: id.NodeID()},
		Storage:       config.Storage{Path: storage},
	})
	if err != nil {
		return nil, err
	}

	return &nodeFiles{key: key, config: cfg, storage: storage}, nil
}

// storagePathForConfig decides what to write as the storage path. The default
// is "storage" inside the node's folder. A custom path is made absolute
// (relative to the current directory, like --home), unless it is inside the
// node's folder, in which case it is stored relative so the folder stays
// self-contained (D11).
func storagePathForConfig(home, flagValue string) (string, error) {
	if flagValue == "" {
		return defaultStorageDir, nil
	}
	abs, err := filepath.Abs(flagValue)
	if err != nil {
		return "", fmt.Errorf("storage path: %w", err)
	}
	if rel, err := filepath.Rel(home, abs); err == nil && filepath.IsLocal(rel) {
		return rel, nil // "." (the node folder itself) is rejected by config validation
	}
	if err := checkStorageOutsideHome(home, abs); err != nil {
		return "", err
	}
	return abs, nil
}

// isExistingDir reports whether home exists. It must either not exist or be
// a folder (a symlink to a folder is fine).
func isExistingDir(home string) (bool, error) {
	fi, err := os.Stat(home)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !fi.IsDir() {
		return false, fmt.Errorf("%s exists and is not a folder", home)
	}
	return true, nil
}

// checkEmpty fails unless dir is empty. A non-empty folder that isn't a node
// is someone else's data, and init must not touch it.
func checkEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("%s is not empty and is not a GoCrypt node; use a new or empty folder", dir)
	}
	return nil
}

// prepareExternalStorage creates an absolute storage folder if needed. It
// must be missing or an empty folder, so two nodes never share one. It
// reports whether it created the folder, so a failed init can remove it.
// Relative storage is created together with the node instead.
func prepareExternalStorage(home, storage string) (created bool, err error) {
	if !filepath.IsAbs(storage) {
		return false, nil
	}
	entries, err := os.ReadDir(storage)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if err := os.MkdirAll(storage, 0o700); err != nil {
			return false, fmt.Errorf("create storage folder: %w", err)
		}
		return true, nil
	case err != nil:
		return false, fmt.Errorf("storage folder: %w", err)
	case len(entries) > 0:
		return false, fmt.Errorf("storage folder %s is not empty; use a new or empty folder", storage)
	}
	return false, nil
}

// createByRename builds the node in a temporary folder next to home and
// renames it into place: the rename is the moment the node appears, whole.
func createByRename(home string, files *nodeFiles) (err error) {
	parent := filepath.Dir(home)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}

	// Same parent folder, so the rename stays on one filesystem. MkdirTemp
	// creates it 0700. A crash can leave it behind; its hidden ".init-" name
	// makes it recognisable, and it holds only a key that was never used.
	tmp, err := os.MkdirTemp(parent, "."+filepath.Base(home)+".init-*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tmp)
		}
	}()

	if err = writeNodeFiles(tmp, files); err != nil {
		return err
	}

	// os.Rename never replaces an existing folder. If home appeared while we
	// worked, another init won the race and its node stays untouched.
	if err = os.Rename(tmp, home); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return errLostRace
		}
		return err
	}
	return syncDir(parent)
}

// createInPlace fills an existing empty folder. A rename can't be used, so a
// lock makes sure only one init writes here at a time, and config.toml is
// written last: it is what makes the folder a node.
func createInPlace(home string, files *nodeFiles) (err error) {
	unlock, err := lockDir(home)
	if err != nil {
		return err
	}
	defer unlock()

	// Decide only now, under the lock: while we waited, another init may
	// have finished, or the folder may have been non-empty all along.
	if _, err := Load(home); err == nil {
		return errLostRace
	}
	if err := checkEmpty(home); err != nil {
		return err
	}

	defer func() {
		if err != nil {
			// The folder was empty when we took the lock, so everything in it
			// is ours. Leave the (empty) folder itself, it isn't ours.
			_ = os.Remove(filepath.Join(home, configFileName))
			_ = os.Remove(filepath.Join(home, keyFileName))
			if !filepath.IsAbs(files.storage) {
				_ = os.RemoveAll(filepath.Join(home, files.storage))
			}
		}
	}()

	if err = os.Chmod(home, 0o700); err != nil {
		return err
	}
	return writeNodeFiles(home, files)
}

// writeNodeFiles writes a node's files into dir: the key, the storage folder
// (if it lives inside the node folder), and config.toml last.
func writeNodeFiles(dir string, files *nodeFiles) error {
	if err := writeFileAtomic(filepath.Join(dir, keyFileName), files.key, 0o600); err != nil {
		return err
	}
	if !filepath.IsAbs(files.storage) {
		if err := os.MkdirAll(filepath.Join(dir, files.storage), 0o700); err != nil {
			return fmt.Errorf("create storage folder: %w", err)
		}
	}
	if err := writeFileAtomic(filepath.Join(dir, configFileName), files.config, 0o600); err != nil {
		return err
	}
	return syncDir(dir)
}
