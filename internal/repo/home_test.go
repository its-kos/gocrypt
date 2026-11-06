package repo

import (
	"os"
	"path/filepath"
	"testing"
)

// setupDirs gives each test its own fake home directory and working directory,
// so results don't depend on the machine running the tests.
func setupDirs(t *testing.T) (home, work string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)

	t.Chdir(t.TempDir())
	work, err := os.Getwd() // may differ from TempDir() if it contains symlinks (macOS)
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	return home, work
}

func TestResolveHome(t *testing.T) {
	home, work := setupDirs(t)

	tests := []struct {
		name string
		flag string
		env  string // "" means GOCRYPT_HOME is set but empty, which counts as not set
		want string
	}{
		{"default", "", "", filepath.Join(home, DefaultDirName)},
		{"env", "", "/srv/gocrypt", "/srv/gocrypt"},
		{"flag", "/data/node", "", "/data/node"},
		{"flag beats env", "/data/node", "/srv/gocrypt", "/data/node"},

		// Relative paths are resolved against the current directory, once.
		{"relative flag", "node1", "", filepath.Join(work, "node1")},
		{"relative env", "", "node2", filepath.Join(work, "node2")},

		// The shell doesn't expand ~ in --home=~/x or in Docker env settings.
		{"tilde in flag", "~/nodes/a", "", filepath.Join(home, "nodes", "a")},
		{"bare tilde", "~", "", home},
		{"tilde in env", "", "~/nodes/b", filepath.Join(home, "nodes", "b")},
		{"only the leading tilde", "~/backups/node~old", "", filepath.Join(home, "backups", "node~old")},
		{"tilde not at the start", "/srv/~/node", "", "/srv/~/node"},
		// ~bob means "bob's home" in a shell, which we don't support. It must not
		// become your home with "bob" glued on.
		{"tilde-user is not expanded", "~bob/node", "", filepath.Join(work, "~bob", "node")},

		// The result is always clean.
		{"cleaned", "/data/./x/../node/", "", "/data/node"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(HomeEnv, tt.env)
			got, err := ResolveHome(tt.flag)
			if err != nil {
				t.Fatalf("ResolveHome(%q) error: %v", tt.flag, err)
			}
			if got != tt.want {
				t.Fatalf("ResolveHome(%q) with %s=%q = %q, want %q", tt.flag, HomeEnv, tt.env, got, tt.want)
			}
		})
	}
}

func TestResolveHomeWithoutUserHome(t *testing.T) {
	// Minimal containers often have no $HOME.
	t.Setenv("HOME", "")
	t.Setenv(HomeEnv, "")

	if _, err := ResolveHome(""); err == nil {
		t.Error(`ResolveHome("") without $HOME succeeded, want error (no default location)`)
	}
	if _, err := ResolveHome("~/x"); err == nil {
		t.Error(`ResolveHome("~/x") without $HOME succeeded, want error (can't expand ~)`)
	}

	// An explicit absolute location doesn't need $HOME at all.
	if got, err := ResolveHome("/data/node"); err != nil || got != "/data/node" {
		t.Errorf(`ResolveHome("/data/node") = %q, %v; want "/data/node", nil`, got, err)
	}
	t.Setenv(HomeEnv, "/srv/gocrypt")
	if got, err := ResolveHome(""); err != nil || got != "/srv/gocrypt" {
		t.Errorf(`ResolveHome("") with %s set = %q, %v; want "/srv/gocrypt", nil`, HomeEnv, got, err)
	}
}
