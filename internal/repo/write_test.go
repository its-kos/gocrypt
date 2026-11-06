package repo

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error: %v", path, err)
	}
	return string(data)
}

func permOf(t *testing.T, path string) os.FileMode {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%s) error: %v", path, err)
	}
	return fi.Mode().Perm()
}

// dirEntries lists the names in dir, so tests can check nothing was left behind.
func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s) error: %v", dir, err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestWriteFileAtomicCreates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if err := writeFileAtomic(path, []byte("hello"), 0o600); err != nil {
		t.Fatalf("writeFileAtomic() error: %v", err)
	}
	if got := readFile(t, path); got != "hello" {
		t.Fatalf("content = %q, want %q", got, "hello")
	}
	if got := permOf(t, path); got != 0o600 {
		t.Fatalf("permissions = %04o, want 0600", got)
	}
}

func TestWriteFileAtomicEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty")
	if err := writeFileAtomic(path, nil, 0o600); err != nil {
		t.Fatalf("writeFileAtomic() error: %v", err)
	}
	if got := readFile(t, path); got != "" {
		t.Fatalf("content = %q, want empty", got)
	}
}

func TestWriteFileAtomicReplaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := writeFileAtomic(path, []byte("new"), 0o600); err != nil {
		t.Fatalf("writeFileAtomic() error: %v", err)
	}
	if got := readFile(t, path); got != "new" {
		t.Fatalf("content = %q, want %q", got, "new")
	}
	// The new file's permissions apply, not the old file's.
	if got := permOf(t, path); got != 0o600 {
		t.Fatalf("permissions = %04o, want 0600", got)
	}
}

func TestWriteFileAtomicLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	if err := writeFileAtomic(filepath.Join(dir, "a"), []byte("x"), 0o600); err != nil {
		t.Fatalf("writeFileAtomic() error: %v", err)
	}
	if got := dirEntries(t, dir); !slices.Equal(got, []string{"a"}) {
		t.Fatalf("directory contains %v after success, want only [a]", got)
	}
}

func TestWriteFileAtomicCleansUpOnFailure(t *testing.T) {
	dir := t.TempDir()
	// A directory where the file should go makes the final rename fail,
	// after the temp file has already been written.
	if err := os.Mkdir(filepath.Join(dir, "target"), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := writeFileAtomic(filepath.Join(dir, "target"), []byte("x"), 0o600); err == nil {
		t.Fatal("writeFileAtomic() onto a directory succeeded, want error")
	}
	if got := dirEntries(t, dir); !slices.Equal(got, []string{"target"}) {
		t.Fatalf("directory contains %v after failure, want only [target] (temp file not cleaned up)", got)
	}
}

func TestWriteFileAtomicMissingDir(t *testing.T) {
	dir := t.TempDir()
	if err := writeFileAtomic(filepath.Join(dir, "missing", "f"), []byte("x"), 0o600); err == nil {
		t.Fatal("writeFileAtomic() into a missing directory succeeded, want error")
	}
	if got := dirEntries(t, dir); len(got) != 0 {
		t.Fatalf("directory contains %v, want nothing", got)
	}
}
