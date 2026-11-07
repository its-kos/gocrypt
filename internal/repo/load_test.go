package repo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/its-kos/gocrypt/internal/config"
	"github.com/its-kos/gocrypt/internal/identity"
)

// writeTestNode creates a valid node in home by hand, without Init, so Load
// can be tested on its own. It returns the config it wrote.
func writeTestNode(t *testing.T, home, storagePath string) *config.Config {
	t.Helper()
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatal(err)
	}

	id, err := identity.Generate()
	if err != nil {
		t.Fatal(err)
	}
	key, err := identity.Marshal(id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, keyFileName), key, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		FormatVersion: config.CurrentFormatVersion,
		Node:          config.Node{Name: "test-node", ID: id.NodeID()},
		Storage:       config.Storage{Path: storagePath},
	}
	data, err := config.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, configFileName), data, 0o600); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func mustLoad(t *testing.T, home string) *Info {
	t.Helper()
	info, err := Load(home)
	if err != nil {
		t.Fatalf("Load(%s) error: %v", home, err)
	}
	return info
}

func TestLoad(t *testing.T) {
	home := filepath.Join(t.TempDir(), "node")
	cfg := writeTestNode(t, home, defaultStorageDir)

	info := mustLoad(t, home)
	if info.Home != home {
		t.Errorf("Home = %q, want %q", info.Home, home)
	}
	if *info.Config != *cfg {
		t.Errorf("Config = %+v, want %+v", *info.Config, *cfg)
	}
	// A relative storage path is resolved against the node's folder (D11).
	if want := filepath.Join(home, defaultStorageDir); info.StoragePath != want {
		t.Errorf("StoragePath = %q, want %q", info.StoragePath, want)
	}
}

func TestLoadAbsoluteStorage(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "node")
	disk := filepath.Join(base, "big-disk", "gocrypt") // e.g. another disk
	writeTestNode(t, home, disk)

	if got := mustLoad(t, home).StoragePath; got != disk {
		t.Fatalf("StoragePath = %q, want %q", got, disk)
	}
}

func TestLoadNotInitialized(t *testing.T) {
	base := t.TempDir()

	empty := filepath.Join(base, "empty")
	if err := os.Mkdir(empty, 0o700); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(base, "other")
	if err := os.Mkdir(other, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "notes.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := map[string]string{
		"folder does not exist":        filepath.Join(base, "missing"),
		"empty folder":                 empty,
		"folder without a config file": other,
	}
	for name, home := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Load(home)
			if !errors.Is(err, ErrNotInitialized) {
				t.Fatalf("Load() error = %v, want ErrNotInitialized", err)
			}
		})
	}
}

func TestLoadBroken(t *testing.T) {
	// A node that exists but can't be used must NOT look "not initialized",
	// or init would think it may create a new node in its place.
	tests := []struct {
		name   string
		damage func(t *testing.T, home string)
		want   string // part of the error message
	}{
		{
			name: "invalid config",
			damage: func(t *testing.T, home string) {
				writeFile(t, filepath.Join(home, configFileName), "this is not toml")
			},
			want: configFileName,
		},
		{
			name: "config fails validation",
			damage: func(t *testing.T, home string) {
				data := strings.Replace(readFile(t, filepath.Join(home, configFileName)), "test-node", "", 1)
				writeFile(t, filepath.Join(home, configFileName), data)
			},
			want: "node.name",
		},
		{
			name: "config is a directory",
			damage: func(t *testing.T, home string) {
				path := filepath.Join(home, configFileName)
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
			},
			want: configFileName,
		},
		{
			name: "identity key missing",
			damage: func(t *testing.T, home string) {
				if err := os.Remove(filepath.Join(home, keyFileName)); err != nil {
					t.Fatal(err)
				}
			},
			want: keyFileName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := filepath.Join(t.TempDir(), "node")
			writeTestNode(t, home, defaultStorageDir)
			tt.damage(t, home)

			_, err := Load(home)
			if err == nil {
				t.Fatal("Load() succeeded, want error")
			}
			if errors.Is(err, ErrNotInitialized) {
				t.Fatalf("Load() error = %v; a broken node must not be reported as not initialized", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load() error = %q, want it to mention %q", err, tt.want)
			}
		})
	}
}

func TestLoadUnreadableConfig(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read any file, so permissions can't be tested")
	}
	home := filepath.Join(t.TempDir(), "node")
	writeTestNode(t, home, defaultStorageDir)
	if err := os.Chmod(filepath.Join(home, configFileName), 0o000); err != nil {
		t.Fatal(err)
	}

	_, err := Load(home)
	if err == nil || errors.Is(err, ErrNotInitialized) {
		t.Fatalf("Load() error = %v, want a read error that is not ErrNotInitialized", err)
	}
}

func TestLoadStorageMustNotContainHome(t *testing.T) {
	// The storage folder must not be the node folder or contain it, or stored
	// data would mix with identity.key and config.toml. config can't check
	// this, because it doesn't know where the node folder is.
	tests := []struct {
		name    string
		storage func(home string) string
		wantErr bool
	}{
		{"sibling folder is fine", func(home string) string { return filepath.Join(filepath.Dir(home), "storage") }, false},
		{"storage is the node folder", func(home string) string { return home }, true},
		{"storage is the parent folder", func(home string) string { return filepath.Dir(home) }, true},
		{"storage is the filesystem root", func(string) string { return "/" }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := filepath.Join(t.TempDir(), "node")
			storage := tt.storage(home)
			writeTestNode(t, home, storage)

			_, err := Load(home)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() with storage %q: error = %v, wantErr %v", storage, err, tt.wantErr)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// passphrase returns a callback that always gives pw.
func passphrase(pw string) func() ([]byte, error) {
	return func() ([]byte, error) { return []byte(pw), nil }
}

func TestLoadIdentityPlain(t *testing.T) {
	info := mustInit(t, filepath.Join(t.TempDir(), "node"), InitOptions{Name: "n"})

	id, err := info.LoadIdentity(nil) // no passphrase needed for a plain key
	if err != nil {
		t.Fatalf("LoadIdentity() error: %v", err)
	}
	if id.NodeID() != info.Config.Node.ID {
		t.Fatalf("loaded key has ID %s, config says %s", id.NodeID(), info.Config.Node.ID)
	}
}

func TestLoadIdentityEncrypted(t *testing.T) {
	info := mustInit(t, filepath.Join(t.TempDir(), "node"), InitOptions{Name: "n", Passphrase: passphrase("pw")})

	id, err := info.LoadIdentity(passphrase("pw"))
	if err != nil {
		t.Fatalf("LoadIdentity() error: %v", err)
	}
	if id.NodeID() != info.Config.Node.ID {
		t.Fatal("loaded key doesn't match the config")
	}

	if _, err := info.LoadIdentity(nil); err == nil {
		t.Error("LoadIdentity(nil) of an encrypted key succeeded, want error")
	}
	if _, err := info.LoadIdentity(passphrase("wrong")); !errors.Is(err, identity.ErrWrongPassphrase) {
		t.Errorf("LoadIdentity(wrong) error = %v, want identity.ErrWrongPassphrase", err)
	}
}

func TestLoadIdentityMismatch(t *testing.T) {
	// Copy another node's key over this node's key: the config and key no
	// longer belong together (D9).
	base := t.TempDir()
	info := mustInit(t, filepath.Join(base, "a"), InitOptions{Name: "a"})
	other := mustInit(t, filepath.Join(base, "b"), InitOptions{Name: "b"})
	otherKey := readFile(t, filepath.Join(other.Home, keyFileName))
	writeFile(t, filepath.Join(info.Home, keyFileName), otherKey)

	_, err := info.LoadIdentity(nil)
	if !errors.Is(err, ErrKeyMismatch) {
		t.Fatalf("LoadIdentity() error = %v, want ErrKeyMismatch", err)
	}
}

func TestLoadIdentityBrokenKey(t *testing.T) {
	info := mustInit(t, filepath.Join(t.TempDir(), "node"), InitOptions{Name: "n"})
	writeFile(t, filepath.Join(info.Home, keyFileName), "not a key")

	_, err := info.LoadIdentity(nil)
	if err == nil {
		t.Fatal("LoadIdentity() of a broken key succeeded, want error")
	}
	if errors.Is(err, ErrKeyMismatch) {
		t.Fatal("a broken key must not be reported as a mismatch")
	}
}

func TestLoadIdentityMissingKey(t *testing.T) {
	info := mustInit(t, filepath.Join(t.TempDir(), "node"), InitOptions{Name: "n"})
	if err := os.Remove(filepath.Join(info.Home, keyFileName)); err != nil {
		t.Fatal(err)
	}
	if _, err := info.LoadIdentity(nil); err == nil {
		t.Fatal("LoadIdentity() with no key file succeeded, want error")
	}
}

func TestKeyEncrypted(t *testing.T) {
	base := t.TempDir()
	plain := mustInit(t, filepath.Join(base, "plain"), InitOptions{Name: "n"})
	enc := mustInit(t, filepath.Join(base, "enc"), InitOptions{Name: "n", Passphrase: passphrase("pw")})

	if got, err := plain.KeyEncrypted(); err != nil || got {
		t.Errorf("plain.KeyEncrypted() = %v, %v; want false, nil", got, err)
	}
	if got, err := enc.KeyEncrypted(); err != nil || !got {
		t.Errorf("enc.KeyEncrypted() = %v, %v; want true, nil", got, err)
	}
}
