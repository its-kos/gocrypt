package health

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/its-kos/gocrypt/internal/repo"
)

// newNode creates a fresh, healthy node to break in a test.
func newNode(t *testing.T, opts repo.InitOptions) *repo.Info {
	t.Helper()
	if opts.Name == "" {
		opts.Name = "test-node"
	}
	info, _, err := repo.Init(filepath.Join(t.TempDir(), "node"), opts)
	if err != nil {
		t.Fatalf("repo.Init() error: %v", err)
	}
	return info
}

func passphrase(pw string) func() ([]byte, error) {
	return func() ([]byte, error) { return []byte(pw), nil }
}

// copyKey puts another node's key into info's node.
func copyKey(t *testing.T, from, to *repo.Info) {
	t.Helper()
	data, err := os.ReadFile(from.KeyPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(to.KeyPath(), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

// mentions fails unless exactly one problem contains want.
func mentions(t *testing.T, problems []error, want string) {
	t.Helper()
	n := 0
	for _, p := range problems {
		if strings.Contains(p.Error(), want) {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("want exactly one problem mentioning %q, got %v", want, problems)
	}
}

func skipOnWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permissions only")
	}
}

func TestHealthyNode(t *testing.T) {
	tests := map[string]repo.InitOptions{
		"plain key":        {},
		"encrypted key":    {Passphrase: passphrase("pw")},
		"external storage": {StoragePath: filepath.Join(t.TempDir(), "disk")},
	}
	for name, opts := range tests {
		t.Run(name, func(t *testing.T) {
			if problems := Check(newNode(t, opts)); len(problems) != 0 {
				t.Fatalf("Check() of a fresh node = %v, want no problems", problems)
			}
		})
	}
}

func TestFolderNotPrivate(t *testing.T) {
	skipOnWindows(t)
	info := newNode(t, repo.InitOptions{})
	if err := os.Chmod(info.Home, 0o755); err != nil {
		t.Fatal(err)
	}
	mentions(t, Check(info), "chmod 700")
}

func TestKeyNotPrivate(t *testing.T) {
	skipOnWindows(t)
	info := newNode(t, repo.InitOptions{})
	if err := os.Chmod(info.KeyPath(), 0o644); err != nil {
		t.Fatal(err)
	}
	mentions(t, Check(info), "chmod 600")
}

func TestKeyFromOtherNode(t *testing.T) {
	info := newNode(t, repo.InitOptions{})
	copyKey(t, newNode(t, repo.InitOptions{}), info)

	for _, p := range Check(info) {
		if errors.Is(p, repo.ErrKeyMismatch) {
			return
		}
	}
	t.Fatal("Check() didn't report repo.ErrKeyMismatch for another node's key")
}

func TestEncryptedKeyIsNotChecked(t *testing.T) {
	// Checking an encrypted key needs the passphrase, and Check never asks
	// for one. So even a wrong encrypted key is not reported here.
	info := newNode(t, repo.InitOptions{Passphrase: passphrase("pw")})
	copyKey(t, newNode(t, repo.InitOptions{Passphrase: passphrase("pw")}), info)

	if problems := Check(info); len(problems) != 0 {
		t.Fatalf("Check() = %v, want no problems (an encrypted key can't be checked)", problems)
	}
}

func TestStorageMissing(t *testing.T) {
	// For storage on another disk, this is what an unmounted disk looks like.
	info := newNode(t, repo.InitOptions{})
	if err := os.Remove(info.StoragePath); err != nil {
		t.Fatal(err)
	}
	mentions(t, Check(info), "does not exist")
}

func TestStorageNotAFolder(t *testing.T) {
	info := newNode(t, repo.InitOptions{})
	if err := os.Remove(info.StoragePath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(info.StoragePath, []byte("oops"), 0o600); err != nil {
		t.Fatal(err)
	}
	mentions(t, Check(info), "not a folder")
}

func TestReportsEveryProblem(t *testing.T) {
	// Two things wrong at once: both are reported, not just the first.
	skipOnWindows(t)
	info := newNode(t, repo.InitOptions{})
	if err := os.Chmod(info.KeyPath(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(info.StoragePath); err != nil {
		t.Fatal(err)
	}
	if problems := Check(info); len(problems) != 2 {
		t.Fatalf("Check() = %v, want 2 problems", problems)
	}
}
