package repo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/its-kos/gocrypt/internal/config"
)

// File and folder names inside a node's folder.
const (
	configFileName    = "config.toml"
	keyFileName       = "identity.key"
	defaultStorageDir = "storage" // default storage path, relative to the node's folder
)

// ErrNotInitialized is returned by Load when the folder holds no GoCrypt node.
var ErrNotInitialized = errors.New("not a GoCrypt node")

// Info describes an existing node.
type Info struct {
	Home        string         // the node's folder (absolute)
	Config      *config.Config // the validated config.toml
	StoragePath string         // absolute path of the storage folder
}

// Load reads the node in home. It returns ErrNotInitialized only if there is
// no config.toml there. A node that exists but is broken or unreadable gives
// a different error, so callers never mistake it for an empty folder.
//
// Load does not read identity.key (that may need a passphrase), only checks
// that it exists.
func Load(home string) (*Info, error) {
	// A file where the folder should be would otherwise give a confusing
	// "config.toml/config.toml: not a directory".
	if fi, err := os.Stat(home); err == nil && !fi.IsDir() {
		return nil, fmt.Errorf("load node: %s is not a folder", home)
	}

	cfgPath := filepath.Join(home, configFileName)
	data, err := os.ReadFile(cfgPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("load node: %s: %w", home, ErrNotInitialized)
	}
	if err != nil {
		return nil, fmt.Errorf("load node: %w", err)
	}

	// config doesn't know where the bytes came from, so add the path here.
	cfg, err := config.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("load node: %s: %w", cfgPath, err)
	}

	keyPath := filepath.Join(home, keyFileName)
	fi, err := os.Stat(keyPath)
	if err != nil {
		return nil, fmt.Errorf("load node: %w", err)
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("load node: %s is not a regular file", keyPath)
	}

	storage := resolveStorage(home, cfg.Storage.Path)
	if err := checkStorageOutsideHome(home, storage); err != nil {
		return nil, fmt.Errorf("load node: %w", err)
	}

	return &Info{Home: home, Config: cfg, StoragePath: storage}, nil
}

// resolveStorage turns the config's storage path into an absolute path.
// Relative paths are relative to the node's folder (D11).
func resolveStorage(home, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(home, path)
}

// checkStorageOutsideHome rejects a storage folder that is the node's folder
// or contains it. Stored data must never share a folder with identity.key and
// config.toml, or cleaning up storage could delete them.
func checkStorageOutsideHome(home, storage string) error {
	// Rel answers "how do I get from storage to home?". If that path stays
	// inside storage ("." or "node"), home is storage or is inside it.
	rel, err := filepath.Rel(storage, home)
	if err == nil && filepath.IsLocal(rel) {
		return fmt.Errorf("storage folder %s must not be or contain the node folder %s", storage, home)
	}
	return nil
}
