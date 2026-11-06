// Package repo manages a node's folder on disk: finding it, creating it, and reading it.
package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	HomeEnv        = "GOCRYPT_HOME" // environment variable that sets the node's folder
	DefaultDirName = ".gocrypt"     // folder name inside the user's home directory
)

// ResolveHome returns the absolute, clean path of the node's folder.
// flagValue (from --home) wins over GOCRYPT_HOME, which wins over ~/.gocrypt.
// A leading ~ is expanded, because the shell doesn't always do it.
func ResolveHome(flagValue string) (string, error) {
	dir := "~/" + DefaultDirName

	if flagValue != "" {
		dir = flagValue
	} else if h := os.Getenv(HomeEnv); h != "" {
		dir = h
	}

	// Without this we create a ~ folder in the current directory.
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home: %w", err)
		}
		dir = filepath.Join(userHome, dir[1:])
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}

	return abs, nil
}
