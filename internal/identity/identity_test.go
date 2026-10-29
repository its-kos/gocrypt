package identity

import (
	"bytes"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
)

// knownSeed is a fixed seed (bytes 0..31) so tests can check exact outputs.
// knownNodeID is the node ID that seed must produce.
const knownNodeID = "aoqqpp7tzyil4hlq3umoos6atft6jvrqtosq2xy53sdgiesvgg4a"

func knownIdentity() *Identity {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i)
	}
	return &Identity{priv: ed25519.NewKeyFromSeed(seed)}
}

// mustGenerate creates an identity or stops the test.
func mustGenerate(t *testing.T) *Identity {
	t.Helper()
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	return id
}

func TestGenerateUnique(t *testing.T) {
	a := mustGenerate(t)
	b := mustGenerate(t)
	if a.PublicKey().Equal(b.PublicKey()) {
		t.Fatal("two generated identities have the same public key")
	}
}

func TestPublicKeyLength(t *testing.T) {
	id := mustGenerate(t)
	if got := len(id.PublicKey()); got != ed25519.PublicKeySize {
		t.Fatalf("public key length = %d, want %d", got, ed25519.PublicKeySize)
	}
}

func TestKeyPairIsLinked(t *testing.T) {
	id := mustGenerate(t)
	msg := []byte("prove it: sign this nonce")
	sig := ed25519.Sign(id.priv, msg)

	if !ed25519.Verify(id.PublicKey(), msg, sig) {
		t.Fatal("signature made with the private key did not verify with the public key")
	}

	// Copy msg before modifying it, so the original stays intact.
	tampered := append([]byte(nil), msg...)
	tampered[0] ^= 0xff
	if ed25519.Verify(id.PublicKey(), tampered, sig) {
		t.Fatal("signature verified for a modified message")
	}

	other := mustGenerate(t)
	if ed25519.Verify(other.PublicKey(), msg, sig) {
		t.Fatal("signature verified with an unrelated public key")
	}
}

func TestPublicKeyIsCopy(t *testing.T) {
	id := mustGenerate(t)
	orig := id.PublicKey()

	pk := id.PublicKey()
	pk[0] ^= 0xff

	if !id.PublicKey().Equal(orig) {
		t.Fatal("modifying the returned public key changed the identity")
	}
}

// --- Step 2: NodeID ---

func TestNodeIDKnownValue(t *testing.T) {
	if got := knownIdentity().NodeID(); got != knownNodeID {
		t.Fatalf("NodeID() = %q, want %q", got, knownNodeID)
	}
}

func TestNodeIDFormat(t *testing.T) {
	id := mustGenerate(t).NodeID()
	if len(id) != 52 {
		t.Fatalf("NodeID length = %d, want 52", len(id))
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z') && !(c >= '2' && c <= '7') {
			t.Fatalf("NodeID %q contains %q, want only a-z and 2-7", id, c)
		}
	}
}

func TestNodeIDStable(t *testing.T) {
	id := mustGenerate(t)
	if id.NodeID() != id.NodeID() {
		t.Fatal("NodeID() returned different values for the same identity")
	}
}

// --- Step 3: ParseNodeID ---

func TestParseNodeIDRoundTrip(t *testing.T) {
	id := mustGenerate(t)
	inputs := map[string]string{
		"as produced":            id.NodeID(),
		"upper case":             strings.ToUpper(id.NodeID()),
		"surrounding whitespace": "  " + id.NodeID() + "\n",
	}
	for name, in := range inputs {
		t.Run(name, func(t *testing.T) {
			pub, err := ParseNodeID(in)
			if err != nil {
				t.Fatalf("ParseNodeID(%q) error: %v", in, err)
			}
			if !pub.Equal(id.PublicKey()) {
				t.Fatal("ParseNodeID did not return the original public key")
			}
		})
	}
}

func TestParseNodeIDInvalid(t *testing.T) {
	// The decoder alone would accept the length and non-canonical cases
	// without an error. ParseNodeID must catch them itself.
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"too short (51 chars)", knownNodeID[:51]},
		{"too long (53 chars)", knownNodeID + "a"},
		{"invalid char 0", knownNodeID[:10] + "0" + knownNodeID[11:]},
		{"invalid char 1", knownNodeID[:10] + "1" + knownNodeID[11:]},
		{"invalid char !", knownNodeID[:10] + "!" + knownNodeID[11:]},
		{"space inside", knownNodeID[:10] + " " + knownNodeID[11:]},
		{"non-canonical last char", knownNodeID[:51] + "b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseNodeID(tt.in); err == nil {
				t.Fatalf("ParseNodeID(%q) succeeded, want error", tt.in)
			}
		})
	}
}

// --- Step 4: Marshal (plain) ---

// knownPlainPEM is what Marshal must produce for knownIdentity() without a passphrase.
// Body = PKCS#8 for Ed25519: a fixed 16-byte header followed by the 32-byte seed (00 01 02 ... 1f).
const knownPlainPEM = "-----BEGIN PRIVATE KEY-----\n" +
	"MC4CAQAwBQYDK2VwBCIEIAABAgMEBQYHCAkKCwwNDg8QERITFBUWFxgZGhscHR4f\n" +
	"-----END PRIVATE KEY-----\n"

func TestMarshalPlainKnownValue(t *testing.T) {
	got, err := Marshal(knownIdentity(), nil)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	if string(got) != knownPlainPEM {
		t.Fatalf("Marshal() =\n%s\nwant\n%s", got, knownPlainPEM)
	}
}

func TestMarshalPlainIsStandardPKCS8(t *testing.T) {
	id := mustGenerate(t)
	data, err := Marshal(id, nil)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	block, rest := pem.Decode(data)
	if block == nil {
		t.Fatal("output is not PEM")
	}
	if len(rest) != 0 {
		t.Fatalf("unexpected data after the PEM block: %q", rest)
	}
	if block.Type != "PRIVATE KEY" {
		t.Fatalf("PEM type = %q, want %q", block.Type, "PRIVATE KEY")
	}
	if len(block.Headers) != 0 {
		t.Fatalf("plain key has PEM headers: %v", block.Headers)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("standard library cannot parse the key: %v", err)
	}
	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		t.Fatalf("parsed key is %T, want ed25519.PrivateKey", key)
	}
	if !edKey.Equal(id.priv) {
		t.Fatal("parsed key differs from the original")
	}
}

func TestMarshalPlainDeterministic(t *testing.T) {
	id := mustGenerate(t)
	a, err := Marshal(id, nil)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	b, err := Marshal(id, []byte{}) // an empty passphrase also means "plain"
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("plain Marshal output differs between calls")
	}
}
