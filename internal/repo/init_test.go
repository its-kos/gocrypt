package repo

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/its-kos/gocrypt/internal/identity"
)

func mustInit(t *testing.T, home string, opts InitOptions) *Info {
	t.Helper()
	info, created, err := Init(home, opts)
	if err != nil {
		t.Fatalf("Init(%s) error: %v", home, err)
	}
	if !created {
		t.Fatalf("Init(%s) created = false, want true", home)
	}
	return info
}

// checkPerm fails the test if path doesn't have exactly perm.
func checkPerm(t *testing.T, path string, perm os.FileMode) {
	t.Helper()
	if got := permOf(t, path); got != perm {
		t.Errorf("%s permissions = %04o, want %04o", filepath.Base(path), got, perm)
	}
}

func TestInitCreatesNode(t *testing.T) {
	parent := t.TempDir()
	home := filepath.Join(parent, "node")

	info := mustInit(t, home, InitOptions{Name: "hyperion"})

	// Files and permissions (D7, D8).
	checkPerm(t, home, 0o700)
	checkPerm(t, filepath.Join(home, keyFileName), 0o600)
	checkPerm(t, filepath.Join(home, configFileName), 0o600)
	checkPerm(t, filepath.Join(home, defaultStorageDir), 0o700)
	if got := dirEntries(t, home); !slices.Equal(got, []string{configFileName, keyFileName, defaultStorageDir}) {
		t.Errorf("node folder contains %v", got)
	}
	// No temp folder left next to the node.
	if got := dirEntries(t, parent); !slices.Equal(got, []string{"node"}) {
		t.Errorf("parent folder contains %v, want only [node]", got)
	}

	// Returned info.
	if info.Home != home || info.Config.Node.Name != "hyperion" {
		t.Errorf("info = %+v", info)
	}
	if want := filepath.Join(home, defaultStorageDir); info.StoragePath != want {
		t.Errorf("StoragePath = %q, want %q", info.StoragePath, want)
	}
	// The default storage path is written relative, so the folder can move (D11).
	if info.Config.Storage.Path != defaultStorageDir {
		t.Errorf("config storage path = %q, want %q", info.Config.Storage.Path, defaultStorageDir)
	}

	// The key on disk belongs to the ID in the config (D9).
	id, err := identity.Unmarshal([]byte(readFile(t, filepath.Join(home, keyFileName))), nil)
	if err != nil {
		t.Fatalf("identity.Unmarshal() error: %v", err)
	}
	if id.NodeID() != info.Config.Node.ID {
		t.Errorf("key's node ID %s != config's %s", id.NodeID(), info.Config.Node.ID)
	}
}

func TestInitDefaultName(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
		t.Skip("no hostname on this machine")
	}
	info := mustInit(t, filepath.Join(t.TempDir(), "node"), InitOptions{})
	if info.Config.Node.Name != hostname {
		t.Fatalf("name = %q, want the hostname %q", info.Config.Node.Name, hostname)
	}
}

func TestInitCreatesMissingParents(t *testing.T) {
	home := filepath.Join(t.TempDir(), "a", "b", "node")
	mustInit(t, home, InitOptions{Name: "n"})
}

func TestInitAlreadyInitialized(t *testing.T) {
	home := filepath.Join(t.TempDir(), "node")
	first := mustInit(t, home, InitOptions{Name: "first"})
	keyBefore := readFile(t, filepath.Join(home, keyFileName))
	cfgBefore := readFile(t, filepath.Join(home, configFileName))

	// Different options must not change anything: the node already exists (D6).
	info, created, err := Init(home, InitOptions{
		Name: "second",
		Passphrase: func() ([]byte, error) {
			t.Fatal("asked for a passphrase although the node already exists")
			return nil, nil
		},
	})
	if err != nil {
		t.Fatalf("second Init() error: %v", err)
	}
	if created {
		t.Fatal("second Init() created = true, want false")
	}
	if info.Config.Node.ID != first.Config.Node.ID || info.Config.Node.Name != "first" {
		t.Fatalf("second Init() returned %+v, want the existing node", info.Config.Node)
	}
	if readFile(t, filepath.Join(home, keyFileName)) != keyBefore ||
		readFile(t, filepath.Join(home, configFileName)) != cfgBefore {
		t.Fatal("second Init() changed the node's files")
	}
}

func TestInitEncryptedKey(t *testing.T) {
	home := filepath.Join(t.TempDir(), "node")
	info := mustInit(t, home, InitOptions{Name: "n", Passphrase: func() ([]byte, error) { return []byte("correct horse"), nil }})

	key := []byte(readFile(t, filepath.Join(home, keyFileName)))
	if !identity.IsEncrypted(key) {
		t.Fatal("identity.key is not encrypted")
	}
	id, err := identity.Unmarshal(key, func() ([]byte, error) { return []byte("correct horse"), nil })
	if err != nil {
		t.Fatalf("identity.Unmarshal() error: %v", err)
	}
	if id.NodeID() != info.Config.Node.ID {
		t.Fatal("encrypted key doesn't match the config's node ID")
	}
}

func TestInitEmptyExistingFolder(t *testing.T) {
	// Like a Docker volume: the folder exists (maybe with loose permissions)
	// and must be filled in place.
	home := filepath.Join(t.TempDir(), "volume")
	if err := os.Mkdir(home, 0o755); err != nil {
		t.Fatal(err)
	}

	mustInit(t, home, InitOptions{Name: "n"})
	checkPerm(t, home, 0o700)
	if _, err := Load(home); err != nil {
		t.Fatalf("Load() after in-place init: %v", err)
	}
}

func TestInitRefuses(t *testing.T) {
	// Each case must fail without changing anything that was there.
	tests := []struct {
		name  string
		setup func(t *testing.T, home string)
	}{
		{"folder with other files", func(t *testing.T, home string) {
			if err := os.Mkdir(home, 0o700); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(home, "notes.txt"), "mine")
		}},
		{"broken node", func(t *testing.T, home string) {
			writeTestNode(t, home, defaultStorageDir)
			writeFile(t, filepath.Join(home, configFileName), "not toml")
		}},
		{"key without config (interrupted init or lost config)", func(t *testing.T, home string) {
			if err := os.Mkdir(home, 0o700); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(home, keyFileName), "a key")
		}},
		{"home is a file", func(t *testing.T, home string) {
			writeFile(t, home, "a file")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := t.TempDir()
			home := filepath.Join(parent, "node")
			tt.setup(t, home)
			before := snapshot(t, parent)

			if _, _, err := Init(home, InitOptions{Name: "n"}); err == nil {
				t.Fatal("Init() succeeded, want error")
			}
			if after := snapshot(t, parent); !maps.Equal(before, after) {
				t.Fatalf("Init() changed files:\nbefore %v\nafter  %v", before, after)
			}
		})
	}
}

func TestInitPassphraseProblemsTouchNothing(t *testing.T) {
	tests := map[string]func() ([]byte, error){
		"empty passphrase": func() ([]byte, error) { return nil, nil },
		"callback fails":   func() ([]byte, error) { return nil, errors.New("user cancelled") },
	}
	for name, passphrase := range tests {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			if _, _, err := Init(filepath.Join(parent, "node"), InitOptions{Name: "n", Passphrase: passphrase}); err == nil {
				t.Fatal("Init() succeeded, want error")
			}
			if got := dirEntries(t, parent); len(got) != 0 {
				t.Fatalf("parent contains %v after a failed init, want nothing", got)
			}
		})
	}
}

func TestInitInvalidOptionsTouchNothing(t *testing.T) {
	parent := t.TempDir()
	home := filepath.Join(parent, "node")

	_, _, err := Init(home, InitOptions{Name: "bad\nname"})
	if err == nil || !strings.Contains(err.Error(), "node.name") {
		t.Fatalf("Init() error = %v, want a node.name error", err)
	}
	if got := dirEntries(t, parent); len(got) != 0 {
		t.Fatalf("parent contains %v after a failed init, want nothing", got)
	}
}

func TestInitRelativeHome(t *testing.T) {
	if _, _, err := Init("relative/node", InitOptions{Name: "n"}); err == nil {
		t.Fatal("Init() with a relative home succeeded, want error")
	}
}

func TestInitCustomStorage(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "node")

	t.Run("new folder elsewhere", func(t *testing.T) {
		disk := filepath.Join(base, "disk1", "gocrypt")
		info := mustInit(t, filepath.Join(base, "n1"), InitOptions{Name: "n", StoragePath: disk})
		if info.StoragePath != disk || info.Config.Storage.Path != disk {
			t.Fatalf("storage = %q (config %q), want %q", info.StoragePath, info.Config.Storage.Path, disk)
		}
		checkPerm(t, disk, 0o700)
	})

	t.Run("existing empty folder", func(t *testing.T) {
		disk := filepath.Join(base, "disk2")
		if err := os.Mkdir(disk, 0o755); err != nil {
			t.Fatal(err)
		}
		mustInit(t, filepath.Join(base, "n2"), InitOptions{Name: "n", StoragePath: disk})
	})

	t.Run("inside the node folder is stored relative", func(t *testing.T) {
		info := mustInit(t, home, InitOptions{Name: "n", StoragePath: filepath.Join(home, "data", "chunks")})
		if want := filepath.Join("data", "chunks"); info.Config.Storage.Path != want {
			t.Fatalf("config storage path = %q, want %q", info.Config.Storage.Path, want)
		}
	})
}

func TestInitCustomStorageRefused(t *testing.T) {
	tests := []struct {
		name    string
		storage func(t *testing.T, base string) string
	}{
		{"not empty", func(t *testing.T, base string) string {
			disk := filepath.Join(base, "disk")
			if err := os.Mkdir(disk, 0o700); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(disk, "other-data"), "x")
			return disk
		}},
		{"is the node folder", func(t *testing.T, base string) string { return filepath.Join(base, "node") }},
		{"contains the node folder", func(t *testing.T, base string) string { return base }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := t.TempDir()
			storage := tt.storage(t, base)
			before := snapshot(t, base)

			if _, _, err := Init(filepath.Join(base, "node"), InitOptions{Name: "n", StoragePath: storage}); err == nil {
				t.Fatal("Init() succeeded, want error")
			}
			if after := snapshot(t, base); !maps.Equal(before, after) {
				t.Fatalf("Init() changed files:\nbefore %v\nafter  %v", before, after)
			}
		})
	}
}

func TestInitConcurrent(t *testing.T) {
	// However many inits run at once, exactly one node is created and
	// everyone reports that same node.
	for _, mode := range []string{"new folder", "existing empty folder"} {
		t.Run(mode, func(t *testing.T) {
			parent := t.TempDir()
			home := filepath.Join(parent, "node")
			if mode == "existing empty folder" {
				if err := os.Mkdir(home, 0o700); err != nil {
					t.Fatal(err)
				}
			}

			const n = 8
			var (
				wg      sync.WaitGroup
				mu      sync.Mutex
				ids     []string
				creates int
				errs    []error
			)
			for range n {
				wg.Add(1)
				go func() {
					defer wg.Done()
					info, created, err := Init(home, InitOptions{Name: "n"})
					mu.Lock()
					defer mu.Unlock()
					if err != nil {
						errs = append(errs, err)
						return
					}
					ids = append(ids, info.Config.Node.ID)
					if created {
						creates++
					}
				}()
			}
			wg.Wait()

			if len(errs) > 0 {
				t.Fatalf("Init() errors: %v", errors.Join(errs...))
			}
			if creates != 1 {
				t.Fatalf("%d inits reported created = true, want exactly 1", creates)
			}
			for _, id := range ids {
				if id != ids[0] {
					t.Fatalf("inits reported different nodes: %v", ids)
				}
			}
			if got := dirEntries(t, parent); !slices.Equal(got, []string{"node"}) {
				t.Fatalf("parent contains %v, want only [node] (temp folders left behind)", got)
			}
			// The node on disk is consistent: its key matches its config.
			key := readFile(t, filepath.Join(home, keyFileName))
			id, err := identity.Unmarshal([]byte(key), nil)
			if err != nil || id.NodeID() != ids[0] {
				t.Fatalf("key on disk doesn't match the reported node (err %v)", err)
			}
		})
	}
}

// snapshot records every path under dir with its content (or "dir"), so a
// test can check that nothing changed.
func snapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		if d.IsDir() {
			snap[rel] = "dir"
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		snap[rel] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot(%s): %v", dir, err)
	}
	return snap
}
