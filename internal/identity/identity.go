// Package identity manages a node's long-term Ed25519 identity key.
package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
)

// Identity is a node's long-term signing keypair.
type Identity struct {
	priv ed25519.PrivateKey // Unexported: only constructed via Generate/Unmarshal
}

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

	return ed25519.PublicKey(decodedBytes), nil
}

func Marshal(id *Identity, passphrase []byte) ([]byte, error) {

}

func Unmarshal(data []byte, passphrase func() ([]byte, error)) (*Identity, error) {

}

func IsEncrypted(data []byte) bool
