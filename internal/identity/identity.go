// Package identity manages a node's long-term Ed25519 identity key.
package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/its-kos/gocrypt/internal/encryptor"
	"golang.org/x/crypto/argon2"
)

// Argon2id settings for turning a passphrase into an encryption key.
// Changing any of these makes existing encrypted keys unreadable.
const (
	argonTime      = 3
	argonMemory    = 64 * 1024 // KiB, so 64 MiB
	argonThreads   = 4
	argonKeyLength = 32 // Bytes, for AES-256
	saltLength     = 16 // Bytes
)

// PEM block types written by Marshal.
const (
	pemType          = "PRIVATE KEY"                   // Standard PKCS#8, readable by other tools
	pemEncryptedType = "GOCRYPT ENCRYPTED PRIVATE KEY" // Custom format, only readable by this package
	saltHeader       = "Salt"                          // Const just cause its cleaner
)

// Identity is a node's long-term signing keypair.
type Identity struct {
	priv ed25519.PrivateKey // Unexported: only constructed via Generate/Unmarshal
}

// ErrWrongPassphrase is returned by Unmarshal when an encrypted key cannot be
// decrypted with the given passphrase.
var ErrWrongPassphrase = errors.New("wrong passphrase")

var nodeIDEncoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

// Generate creates a new random identity.
func Generate() (*Identity, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate identity: %w", err)
	}

	return &Identity{priv: priv}, nil
}

// PublicKey returns a copy of the public key derived from the identity.
func (id *Identity) PublicKey() ed25519.PublicKey {
	// crypto.PublicKey must be asserted to ed25519
	return id.priv.Public().(ed25519.PublicKey)
}

// NodeID returns the node's ID: its public key in lowercase, unpadded base32 (52 chars).
func (id *Identity) NodeID() string {
	return nodeIDEncoding.EncodeToString(id.PublicKey())
}

// ParseNodeID returns the node's public key from its ID.
func ParseNodeID(s string) (ed25519.PublicKey, error) {
	cleanStr := strings.ToLower(strings.TrimSpace(s))
	decodedBytes, err := nodeIDEncoding.DecodeString(cleanStr)
	if err != nil {
		return nil, fmt.Errorf("parse node id: %w", err)
	}

	if len(decodedBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("parse node id: expected %d bytes, got %d", ed25519.PublicKeySize, len(decodedBytes))
	}

	if nodeIDEncoding.EncodeToString(decodedBytes) != cleanStr {
		return nil, errors.New("parse node id: nodeID is not canonical")
	}

	return decodedBytes, nil
}

// Marshal encodes the identity as PEM for saving to disk.
//
// With an empty passphrase the output is a standard PKCS#8 "PRIVATE KEY" block
// that other tools can read. Otherwise, the key is encrypted with a key derived
// from the passphrase, and the output can only be read back by Unmarshal.
func Marshal(id *Identity, passphrase []byte) ([]byte, error) {
	pkcsKey, err := x509.MarshalPKCS8PrivateKey(id.priv)
	if err != nil {
		return nil, fmt.Errorf("marshal identity: %w", err)
	}

	if len(passphrase) == 0 {
		return pem.EncodeToMemory(&pem.Block{
			Type:  pemType,
			Bytes: pkcsKey,
		}), nil
	}

	salt := make([]byte, saltLength)
	_, _ = rand.Read(salt) // Never returns an error since Go 1.24

	encryptedKey, err := encryptor.Encrypt(pkcsKey, deriveKey(passphrase, salt))
	if err != nil {
		return nil, fmt.Errorf("marshal identity: %w", err)
	}

	// Unmarshal needs the same salt to derive the same key. The salt is not
	// secret, so it is stored in plain text.
	return pem.EncodeToMemory(&pem.Block{
		Type:    pemEncryptedType,
		Headers: map[string]string{saltHeader: base64.StdEncoding.EncodeToString(salt)},
		Bytes:   encryptedKey,
	}), nil
}

// Unmarshal decodes an identity produced by Marshal.
//
// The passphrase callback is only called if the key is encrypted, so callers
// can prompt the user lazily. It may be nil if the key is known to be plain.
// If the passphrase does not decrypt the key, Unmarshal returns an error
// wrapping ErrWrongPassphrase.
func Unmarshal(data []byte, passphrase func() ([]byte, error)) (*Identity, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("unmarshal identity: no PEM block found")
	}

	var pkcsKey []byte

	switch block.Type {
	case pemType:
		pkcsKey = block.Bytes
	case pemEncryptedType:
		// Validate the file before asking for a passphrase, so a broken file
		// fails right away instead of after the user types it.
		salt, err := base64.StdEncoding.DecodeString(block.Headers[saltHeader])
		if err != nil {
			return nil, fmt.Errorf("unmarshal identity: decode salt: %w", err)
		}

		if len(salt) != saltLength {
			return nil, fmt.Errorf("unmarshal identity: expected %d byte salt, got %d", saltLength, len(salt))
		}

		if passphrase == nil {
			return nil, errors.New("unmarshal identity: key is encrypted but no passphrase was provided")
		}

		pass, err := passphrase()
		if err != nil {
			return nil, fmt.Errorf("unmarshal identity: get passphrase: %w", err)
		}

		// AES-GCM can't tell a wrong key from corrupted data, but a wrong
		// passphrase is by far the likeliest cause.
		pkcsKey, err = encryptor.Decrypt(block.Bytes, deriveKey(pass, salt))
		if err != nil {
			return nil, fmt.Errorf("unmarshal identity: %w", ErrWrongPassphrase)
		}
	default:
		return nil, fmt.Errorf("unmarshal identity: unknown PEM block type %q", block.Type)
	}

	key, err := x509.ParsePKCS8PrivateKey(pkcsKey)
	if err != nil {
		return nil, fmt.Errorf("unmarshal identity: %w", err)
	}

	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("unmarshal identity: expected ed25519.PrivateKey, got %T", key)
	}

	return &Identity{priv: priv}, nil
}

// IsEncrypted reports whether data is an encrypted identity produced by Marshal.
func IsEncrypted(data []byte) bool {
	block, _ := pem.Decode(data)
	if block == nil {
		return false
	}

	return block.Type == pemEncryptedType
}

// deriveKey turns a passphrase and salt into an AES-256 key using Argon2id.
func deriveKey(passphrase, salt []byte) []byte {
	return argon2.IDKey(passphrase, salt, argonTime, argonMemory, argonThreads, argonKeyLength)
}
