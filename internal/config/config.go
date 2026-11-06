// Package config reads and writes a node's config.toml.
package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/BurntSushi/toml"

	"github.com/its-kos/gocrypt/internal/identity"
)

const (
	// CurrentFormatVersion is the config layout this version of GoCrypt reads and writes.
	// Any change to the file's layout or meaning should raise CurrentFormatVersion by one.
	CurrentFormatVersion = 1

	// MaxNameLength is the longest allowed node name, in characters (not bytes).
	MaxNameLength = 64
)

// Config is the contents of a node's config.toml.
type Config struct {
	FormatVersion int     `toml:"format_version"`
	Node          Node    `toml:"node"`
	Storage       Storage `toml:"storage"`
}

// Node holds the node's name and ID.
type Node struct {
	Name string `toml:"name"` // shown to humans, defaults to the hostname
	ID   string `toml:"id"`   // copy of the public node ID, checked against identity.key
}

// Storage says where the node keeps the data it stores.
type Storage struct {
	// Path is relative to the node's home folder, or absolute for a custom location.
	Path string `toml:"path"`
}

// Validate reports the first problem in c, naming the field.
func (c *Config) Validate() error {
	if err := checkVersion(c.FormatVersion); err != nil {
		return err
	}
	if err := validateName(c.Node.Name); err != nil {
		return fmt.Errorf("node.name: %w", err)
	}
	if err := validateID(c.Node.ID); err != nil {
		return fmt.Errorf("node.id: %w", err)
	}
	if err := validateStoragePath(c.Storage.Path); err != nil {
		return fmt.Errorf("storage.path: %w", err)
	}
	return nil
}

// Marshal validates c and encodes it as TOML. It never produces a file
// that Unmarshal would refuse to read back.
func Marshal(c *Config) ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	return data, nil
}

// Unmarshal decodes and validates a config file. Unknown keys are an error,
// so a misspelled setting fails loudly instead of being ignored.
func Unmarshal(data []byte) (*Config, error) {
	var c Config
	md, err := toml.Decode(string(data), &c)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Check the version before unknown keys: a file from a newer GoCrypt
	// probably has keys this version doesn't know, and "upgrade" is the
	// message the user can act on.
	if err := checkVersion(c.FormatVersion); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if keys := md.Undecoded(); len(keys) > 0 {
		return nil, fmt.Errorf("parse config: unknown keys %v", keys)
	}

	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &c, nil
}

func checkVersion(v int) error {
	switch {
	case v == CurrentFormatVersion:
		return nil
	case v == 0:
		return errors.New("format_version is missing")
	case v > CurrentFormatVersion:
		return fmt.Errorf("format_version is %d, but this version of GoCrypt supports %d; upgrade GoCrypt", v, CurrentFormatVersion)
	default:
		return fmt.Errorf("format_version %d is not supported", v)
	}
}

func validateName(name string) error {
	if name == "" {
		return errors.New("must not be empty")
	}
	if strings.TrimSpace(name) != name {
		return errors.New("must not start or end with whitespace")
	}
	if n := utf8.RuneCountInString(name); n > MaxNameLength {
		return fmt.Errorf("must be at most %d characters, got %d", MaxNameLength, n)
	}
	if strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return errors.New("must not contain control characters")
	}
	return nil
}

func validateID(id string) error {
	if id == "" {
		return errors.New("must not be empty")
	}
	if _, err := identity.ParseNodeID(id); err != nil {
		return err
	}
	// ParseNodeID accepts upper case and surrounding spaces for people typing
	// IDs. The file must hold the exact form, because it is compared with the
	// key's node ID as a string.
	if strings.ToLower(strings.TrimSpace(id)) != id {
		return errors.New("must be lower case with no surrounding whitespace")
	}
	return nil
}

func validateStoragePath(p string) error {
	if p == "" {
		return errors.New("must not be empty")
	}

	if !filepath.IsAbs(p) {
		switch {
		case strings.HasPrefix(p, "~"):
			return errors.New("~ is not expanded here; use an absolute path or one relative to the node's home folder")
		case !filepath.IsLocal(p):
			return errors.New("a relative path must stay inside the node's home folder")
		case filepath.Clean(p) == ".":
			return errors.New("must not be the node's home folder itself")
		}
	}

	if clean := filepath.Clean(p); clean != p {
		return fmt.Errorf("must be written as %q", clean)
	}
	return nil
}
