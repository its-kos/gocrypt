package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/its-kos/gocrypt/internal/identity"
	"github.com/its-kos/gocrypt/internal/repo"
)

// run executes gocrypt with args, like the real command line, and returns
// what it printed to stdout.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

// isolate keeps tests away from the real ~/.gocrypt and environment.
func isolate(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv(repo.HomeEnv, "")
	t.Setenv(passphraseFileEnv, "")
	return dir
}

func TestInitCommand(t *testing.T) {
	home := filepath.Join(isolate(t), "node")

	out, err := run(t, "init", "--home", home, "--name", "test-node")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}
	info, err := repo.Load(home)
	if err != nil {
		t.Fatalf("no node after init: %v", err)
	}
	for _, want := range []string{"Initialized GoCrypt node in " + home, "test-node", info.Config.Node.ID, info.StoragePath} {
		if !strings.Contains(out, want) {
			t.Errorf("output doesn't contain %q:\n%s", want, out)
		}
	}
}

func TestInitCommandTwice(t *testing.T) {
	home := filepath.Join(isolate(t), "node")
	if _, err := run(t, "init", "--home", home, "--name", "first"); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "init", "--home", home, "--name", "second")
	if err != nil {
		t.Fatalf("second init error: %v (re-running init must succeed)", err)
	}
	if !strings.Contains(out, "already initialized") || !strings.Contains(out, "first") {
		t.Fatalf("second init output:\n%s\nwant it to report the existing node", out)
	}
}

func TestInitCommandHomeFromEnv(t *testing.T) {
	home := filepath.Join(isolate(t), "from-env")
	t.Setenv(repo.HomeEnv, home)

	if _, err := run(t, "init", "--name", "n"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Load(home); err != nil {
		t.Fatalf("no node in $%s: %v", repo.HomeEnv, err)
	}
}

func TestInitCommandDefaultHome(t *testing.T) {
	userHome := isolate(t)
	if _, err := run(t, "init", "--name", "n"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Load(filepath.Join(userHome, repo.DefaultDirName)); err != nil {
		t.Fatalf("no node in ~/%s: %v", repo.DefaultDirName, err)
	}
}

func TestInitCommandRelativeStoragePath(t *testing.T) {
	dir := isolate(t)
	t.Chdir(dir)
	work, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := run(t, "init", "--home", filepath.Join(dir, "node"), "--name", "n", "--storage-path", "disk"); err != nil {
		t.Fatal(err)
	}
	info, err := repo.Load(filepath.Join(dir, "node"))
	if err != nil {
		t.Fatal(err)
	}
	// Relative to where the command ran, like --home.
	if want := filepath.Join(work, "disk"); info.StoragePath != want {
		t.Fatalf("StoragePath = %q, want %q", info.StoragePath, want)
	}
}

// decryptsWith checks that the node's key is encrypted and opens with passphrase.
func decryptsWith(t *testing.T, home, passphrase string) {
	t.Helper()
	key, err := os.ReadFile(filepath.Join(home, "identity.key"))
	if err != nil {
		t.Fatal(err)
	}
	if !identity.IsEncrypted(key) {
		t.Fatal("identity.key is not encrypted")
	}
	if _, err := identity.Unmarshal(key, func() ([]byte, error) { return []byte(passphrase), nil }); err != nil {
		t.Fatalf("key doesn't open with %q: %v", passphrase, err)
	}
}

func TestInitCommandPassphraseFile(t *testing.T) {
	dir := isolate(t)
	pwFile := filepath.Join(dir, "pw")
	// A trailing newline (as written by echo) is not part of the passphrase.
	if err := os.WriteFile(pwFile, []byte("correct horse\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Run("flag", func(t *testing.T) {
		home := filepath.Join(dir, "node1")
		if _, err := run(t, "init", "--home", home, "--name", "n", "--encrypt-key", "--passphrase-file", pwFile); err != nil {
			t.Fatal(err)
		}
		decryptsWith(t, home, "correct horse")
	})

	t.Run("environment variable", func(t *testing.T) {
		t.Setenv(passphraseFileEnv, pwFile)
		home := filepath.Join(dir, "node2")
		if _, err := run(t, "init", "--home", home, "--name", "n", "--encrypt-key"); err != nil {
			t.Fatal(err)
		}
		decryptsWith(t, home, "correct horse")
	})
}

func TestInitCommandErrors(t *testing.T) {
	dir := isolate(t)
	empty := filepath.Join(dir, "empty-pw")
	if err := os.WriteFile(empty, []byte("\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		args []string
		want string // part of the error message
	}{
		{"unexpected argument", []string{"extra"}, "unknown command"},
		{"passphrase file without --encrypt-key", []string{"--passphrase-file", empty}, "--encrypt-key"},
		{"empty passphrase file", []string{"--encrypt-key", "--passphrase-file", empty}, "empty"},
		{"missing passphrase file", []string{"--encrypt-key", "--passphrase-file", filepath.Join(dir, "nope")}, "no such file"},
		// Tests don't run in a terminal, so with no file there is no way to
		// get a passphrase. It must fail instead of waiting forever.
		{"no terminal and no passphrase file", []string{"--encrypt-key"}, "--passphrase-file"},
		{"invalid name", []string{"--name", " spaces "}, "node.name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := filepath.Join(t.TempDir(), "node")
			args := append([]string{"init", "--home", home}, tt.args...)

			_, err := run(t, args...)
			if err == nil {
				t.Fatal("init succeeded, want error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want it to mention %q", err, tt.want)
			}
			if _, statErr := os.Stat(home); statErr == nil {
				t.Fatal("a failed init left a node folder behind")
			}
		})
	}
}
