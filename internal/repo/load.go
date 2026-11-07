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

// Files inside a node's folder.
const (
	configFileName    = "config.toml"
	keyFileName       = "identity.key"
	defaultStorageDir = "storage"
)

var (
	// ErrNotInitialized means the folder holds no GoCrypt node.
	ErrNotInitialized = errors.New("not a GoCrypt node")

	// ErrKeyMismatch means identity.key and config.toml belong to different nodes.
	ErrKeyMismatch = errors.New("identity.key does not belong to this node's config")
)

// Info describes an existing node.
type Info struct {
	Home        string         // the node's folder
	Config      *config.Config // contents of config.toml
	StoragePath string         // absolute path of the storage folder
}

// Load opens the node in home and reads its config. It returns
// ErrNotInitialized only if there is no config.toml; a broken node gives a
// different error. The key is not read here (it may need a passphrase).
func Load(home string) (*Info, error) {
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

	cfg, err := config.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("load node: %s: %w", cfgPath, err)
	}

	keyPath := filepath.Join(home, keyFileName)
	if fi, err := os.Stat(keyPath); err != nil {
		return nil, fmt.Errorf("load node: %w", err)
	} else if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("load node: %s is not a regular file", keyPath)
	}

	storage := cfg.Storage.Path
	if !filepath.IsAbs(storage) {
		storage = filepath.Join(home, storage) // relative to the node's folder (D11)
	}
	if err := checkStorageOutsideHome(home, storage); err != nil {
		return nil, fmt.Errorf("load node: %w", err)
	}

	return &Info{Home: home, Config: cfg, StoragePath: storage}, nil
}

// LoadIdentity reads the node's key and checks that it belongs to this node's
// config. passphrase is only called if the key is encrypted; it may be nil.
func (info *Info) LoadIdentity(passphrase func() ([]byte, error)) (*identity.Identity, error) {
	data, err := os.ReadFile(filepath.Join(info.Home, keyFileName))
	if err != nil {
		return nil, fmt.Errorf("load identity: %w", err)
	}

	key, err := identity.Unmarshal(data, passphrase)
	if err != nil {
		return nil, fmt.Errorf("load identity: %w", err)
	}

	if key.NodeID() != info.Config.Node.ID {
		return nil, fmt.Errorf("load identity: key is %s, config says %s: %w",
			key.NodeID(), info.Config.Node.ID, ErrKeyMismatch)
	}

	return key, nil
}

// KeyEncrypted reports whether the node's key is protected by a passphrase.
func (info *Info) KeyEncrypted() (bool, error) {
	data, err := os.ReadFile(filepath.Join(info.Home, keyFileName))
	if err != nil {
		return false, err
	}
	return identity.IsEncrypted(data), nil
}

// checkStorageOutsideHome rejects a storage folder that is the node's folder
// or contains it, so stored data never mixes with the key and config.
func checkStorageOutsideHome(home, storage string) error {
	rel, err := filepath.Rel(storage, home)
	if err == nil && filepath.IsLocal(rel) {
		return fmt.Errorf("storage folder %s must not be or contain the node folder %s", storage, home)
	}
	return nil
}
