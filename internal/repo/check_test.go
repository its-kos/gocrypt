package repo

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// problemsMention fails the test unless exactly one problem mentions want.
func problemsMention(t *testing.T, problems []error, want string) {
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

func TestCheckHealthyNode(t *testing.T) {
	base := t.TempDir()
	tests := map[string]InitOptions{
		"plain key":        {Name: "n"},
		"encrypted key":    {Name: "n", Passphrase: passphrase("pw")}, // not checked, but not a problem
		"external storage": {Name: "n", StoragePath: filepath.Join(base, "disk")},
	}
	for name, opts := range tests {
		t.Run(name, func(t *testing.T) {
			info := mustInit(t, filepath.Join(base, name), opts)
			if problems := info.Check(); len(problems) != 0 {
				t.Fatalf("Check() of a fresh node = %v, want no problems", problems)
			}
		})
	}
}

func TestCheckPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permissions only")
	}

	t.Run("folder readable by others", func(t *testing.T) {
		info := mustInit(t, filepath.Join(t.TempDir(), "node"), InitOptions{Name: "n"})
		if err := os.Chmod(info.Home, 0o755); err != nil {
			t.Fatal(err)
		}
		problemsMention(t, info.Check(), "chmod 700")
	})

	t.Run("key readable by others", func(t *testing.T) {
		info := mustInit(t, filepath.Join(t.TempDir(), "node"), InitOptions{Name: "n"})
		if err := os.Chmod(filepath.Join(info.Home, keyFileName), 0o644); err != nil {
			t.Fatal(err)
		}
		problemsMention(t, info.Check(), "chmod 600")
	})
}

func TestCheckKeyMismatch(t *testing.T) {
	base := t.TempDir()
	info := mustInit(t, filepath.Join(base, "a"), InitOptions{Name: "a"})
	other := mustInit(t, filepath.Join(base, "b"), InitOptions{Name: "b"})
	writeFile(t, filepath.Join(info.Home, keyFileName), readFile(t, filepath.Join(other.Home, keyFileName)))

	problems := info.Check()
	found := false
	for _, p := range problems {
		if errors.Is(p, ErrKeyMismatch) {
			found = true
		}
	}
	if !found {
		t.Fatalf("Check() = %v, want a problem wrapping ErrKeyMismatch", problems)
	}
}

func TestCheckStorage(t *testing.T) {
	t.Run("missing (e.g. a disk that isn't mounted)", func(t *testing.T) {
		info := mustInit(t, filepath.Join(t.TempDir(), "node"), InitOptions{Name: "n"})
		if err := os.Remove(info.StoragePath); err != nil {
			t.Fatal(err)
		}
		problemsMention(t, info.Check(), "does not exist")
	})

	t.Run("a file instead of a folder", func(t *testing.T) {
		info := mustInit(t, filepath.Join(t.TempDir(), "node"), InitOptions{Name: "n"})
		if err := os.Remove(info.StoragePath); err != nil {
			t.Fatal(err)
		}
		writeFile(t, info.StoragePath, "oops")
		problemsMention(t, info.Check(), "not a folder")
	})
}

func TestCheckReportsAllProblems(t *testing.T) {
	// Several things wrong at once: all of them are reported, not just the
	// first, so the user can fix everything in one go.
	if runtime.GOOS == "windows" {
		t.Skip("Unix permissions only")
	}
	info := mustInit(t, filepath.Join(t.TempDir(), "node"), InitOptions{Name: "n"})
	if err := os.Chmod(filepath.Join(info.Home, keyFileName), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(info.StoragePath); err != nil {
		t.Fatal(err)
	}

	if problems := info.Check(); len(problems) != 2 {
		t.Fatalf("Check() = %v, want 2 problems", problems)
	}
}
