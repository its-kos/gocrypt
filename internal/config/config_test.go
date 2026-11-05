package config

import (
	"strings"
	"testing"
)

// A valid node ID (from the identity package's known test key).
const testNodeID = "aoqqpp7tzyil4hlq3umoos6atft6jvrqtosq2xy53sdgiesvgg4a"

func validConfig() *Config {
	return &Config{
		FormatVersion: CurrentFormatVersion,
		Node:          Node{Name: "hyperion", ID: testNodeID},
		Storage:       Storage{Path: "storage"}, // relative to the node's home folder
	}
}

// knownTOML is what Marshal must produce for validConfig().
const knownTOML = `format_version = 1

[node]
  name = "hyperion"
  id = "aoqqpp7tzyil4hlq3umoos6atft6jvrqtosq2xy53sdgiesvgg4a"

[storage]
  path = "storage"
`

func TestMarshalKnownValue(t *testing.T) {
	got, err := Marshal(validConfig())
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	if string(got) != knownTOML {
		t.Fatalf("Marshal() =\n%s\nwant\n%s", got, knownTOML)
	}
}

func TestRoundTrip(t *testing.T) {
	want := validConfig()
	data, err := Marshal(want)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if *got != *want {
		t.Fatalf("Unmarshal(Marshal(c)) = %+v, want %+v", *got, *want)
	}
}

func TestUnmarshalHandEdited(t *testing.T) {
	// Users may edit the file: comments, different order, other spacing.
	data := `# my homelab node
format_version = 1

[storage]
path = "/mnt/big-disk/gocrypt"   # moved to the big disk

[node]
id   = "` + testNodeID + `"
name = "homelab"
`
	got, err := Unmarshal([]byte(data))
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if got.Node.Name != "homelab" || got.Storage.Path != "/mnt/big-disk/gocrypt" {
		t.Fatalf("Unmarshal() = %+v, values not read correctly", *got)
	}
}

func TestUnmarshalInvalid(t *testing.T) {
	// Each case breaks one rule. Most of them are things the TOML library
	// itself accepts without an error.
	valid := knownTOML
	replace := func(old, new string) string {
		if !strings.Contains(valid, old) {
			t.Fatalf("test setup: %q not found in knownTOML", old)
		}
		return strings.Replace(valid, old, new, 1)
	}

	tests := []struct {
		name string
		data string
	}{
		// Syntax and types (the library catches these).
		{"not TOML", "this is not toml"},
		{"version is a string", replace("format_version = 1", `format_version = "1"`)},

		// Version.
		{"empty file", ""},
		{"missing version", replace("format_version = 1\n", "")},
		{"newer version", replace("format_version = 1", "format_version = 2")},

		// Unknown keys (typos would otherwise be silently ignored).
		{"misspelled key", replace(`name = "hyperion"`, `nmae = "hyperion"`)},
		{"unknown key", replace(`[storage]`, "[storage]\n  colour = \"blue\"")},
		{"unknown table", valid + "\n[extra]\n  x = 1\n"},

		// Node name.
		{"empty name", replace(`name = "hyperion"`, `name = ""`)},
		{"name too long", replace(`name = "hyperion"`, `name = "`+strings.Repeat("a", MaxNameLength+1)+`"`)},
		{"name with surrounding space", replace(`name = "hyperion"`, `name = " hyperion"`)},
		{"name with newline", replace(`name = "hyperion"`, `name = "hyper\nion"`)},
		{"name with terminal escape", replace(`name = "hyperion"`, `name = "\u001b[31mred"`)},

		// Node ID.
		{"missing id", replace(`id = "`+testNodeID+`"`, "")},
		{"invalid id", replace(testNodeID, "not-a-node-id")},
		{"id too short", replace(testNodeID, testNodeID[:51])},
		{"upper-case id", replace(testNodeID, strings.ToUpper(testNodeID))},

		// Storage path (detailed cases are in TestValidateStoragePath).
		{"missing path", replace(`path = "storage"`, "")},
		{"path escapes home", replace(`path = "storage"`, `path = "../storage"`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Unmarshal([]byte(tt.data)); err == nil {
				t.Fatalf("Unmarshal() succeeded, want error. Input:\n%s", tt.data)
			}
		})
	}
}

func TestUnmarshalNewerVersionMessage(t *testing.T) {
	// A newer file may have keys this version doesn't know. The error must
	// be about the version, not about the unknown keys, because that's what
	// the user can act on (upgrade GoCrypt).
	data := strings.Replace(knownTOML, "format_version = 1", "format_version = 2\nnew_setting = true", 1)
	_, err := Unmarshal([]byte(data))
	if err == nil {
		t.Fatal("Unmarshal() succeeded, want error")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Fatalf("error = %q, want it to mention the version", err)
	}
}

func TestValidateNameLength(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"exactly the limit", strings.Repeat("a", MaxNameLength), false},
		{"one over the limit", strings.Repeat("a", MaxNameLength+1), true},
		// The limit counts characters, not bytes: é is 2 bytes in UTF-8.
		{"multi-byte characters at the limit", strings.Repeat("é", MaxNameLength), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig()
			c.Node.Name = tt.value
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStoragePath(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
	}{
		// Relative paths are inside the node's home folder.
		{"storage", false},
		{"data/chunks", false},
		// Absolute paths are a custom location, e.g. another disk.
		{"/mnt/big-disk/gocrypt", false},
		{"/home/test/.gocrypt/storage", false},

		{"", true},                    // missing
		{".", true},                   // the home folder itself: would mix data with the key and config
		{"../storage", true},          // escapes the home folder
		{"a/../../storage", true},     // escapes, hidden behind a ..
		{"~/storage", true},           // ~ isn't expanded in a file; this would be a folder named "~"
		{"~", true},                   // same
		{"./storage", true},           // not clean: should be "storage"
		{"storage/", true},            // not clean: trailing slash
		{"a/../storage", true},        // not clean: should be "storage"
		{"/mnt/big-disk/", true},      // not clean: trailing slash
		{"/mnt/../etc/gocrypt", true}, // not clean: hides where data really goes
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			c := validConfig()
			c.Storage.Path = tt.path
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() with path %q: error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestMarshalRejectsInvalid(t *testing.T) {
	// Never write a config file that Unmarshal would refuse to read back.
	c := validConfig()
	c.Storage.Path = "../outside"
	if _, err := Marshal(c); err == nil {
		t.Fatal("Marshal() of an invalid config succeeded, want error")
	}
}
