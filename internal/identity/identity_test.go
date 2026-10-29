package identity

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
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

// --- Step 5: Marshal (encrypted) ---

const testPassphrase = "correct horse battery staple"

// passphraseFunc returns a passphrase callback that always gives p.
func passphraseFunc(p string) func() ([]byte, error) {
	return func() ([]byte, error) { return []byte(p), nil }
}

// mustNotBeCalled returns a passphrase callback that fails the test if used.
func mustNotBeCalled(t *testing.T) func() ([]byte, error) {
	t.Helper()
	return func() ([]byte, error) {
		t.Fatal("passphrase callback was called, but it should not be")
		return nil, nil
	}
}

// mustMarshal marshals id or stops the test.
func mustMarshal(t *testing.T, id *Identity, passphrase string) []byte {
	t.Helper()
	data, err := Marshal(id, []byte(passphrase))
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	return data
}

// mustDecodePEM decodes a single PEM block or stops the test.
func mustDecodePEM(t *testing.T, data []byte) *pem.Block {
	t.Helper()
	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatal("output is not PEM")
	}
	return block
}

func TestMarshalEncryptedFormat(t *testing.T) {
	id := knownIdentity()
	block := mustDecodePEM(t, mustMarshal(t, id, testPassphrase))

	if block.Type != pemEncryptedType {
		t.Fatalf("PEM type = %q, want %q", block.Type, pemEncryptedType)
	}
	if block.Headers[saltHeader] == "" {
		t.Fatal("encrypted key has no Salt header")
	}

	// The private key must not appear in readable form.
	der, err := x509.MarshalPKCS8PrivateKey(id.priv)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey() error: %v", err)
	}
	if bytes.Contains(block.Bytes, id.priv.Seed()) || bytes.Contains(block.Bytes, der) {
		t.Fatal("encrypted output contains the private key in readable form")
	}
}

func TestMarshalEncryptedIsRandomized(t *testing.T) {
	id := mustGenerate(t)
	a := mustDecodePEM(t, mustMarshal(t, id, testPassphrase))
	b := mustDecodePEM(t, mustMarshal(t, id, testPassphrase))

	if a.Headers[saltHeader] == b.Headers[saltHeader] {
		t.Fatal("two encrypted outputs use the same salt")
	}
	if bytes.Equal(a.Bytes, b.Bytes) {
		t.Fatal("two encrypted outputs have the same body")
	}
}

// --- Step 6: Unmarshal ---

func TestUnmarshalPlainKnownValue(t *testing.T) {
	id, err := Unmarshal([]byte(knownPlainPEM), mustNotBeCalled(t))
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if got := id.NodeID(); got != knownNodeID {
		t.Fatalf("NodeID() = %q, want %q", got, knownNodeID)
	}
}

func TestUnmarshalRoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		passphrase string
	}{
		{"plain", ""},
		{"encrypted", testPassphrase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := mustGenerate(t)
			got, err := Unmarshal(mustMarshal(t, id, tt.passphrase), passphraseFunc(tt.passphrase))
			if err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if !got.priv.Equal(id.priv) {
				t.Fatal("Unmarshal did not return the original key")
			}
		})
	}
}

func TestUnmarshalPlainSkipsPassphrase(t *testing.T) {
	data := mustMarshal(t, mustGenerate(t), "")
	if _, err := Unmarshal(data, mustNotBeCalled(t)); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if _, err := Unmarshal(data, nil); err != nil {
		t.Fatalf("Unmarshal() with nil callback error: %v", err)
	}
}

func TestUnmarshalWrongPassphrase(t *testing.T) {
	data := mustMarshal(t, mustGenerate(t), testPassphrase)
	_, err := Unmarshal(data, passphraseFunc("wrong"))
	if !errors.Is(err, ErrWrongPassphrase) {
		t.Fatalf("Unmarshal() error = %v, want ErrWrongPassphrase", err)
	}
}

func TestUnmarshalEncryptedNeedsPassphrase(t *testing.T) {
	data := mustMarshal(t, mustGenerate(t), testPassphrase)
	if _, err := Unmarshal(data, nil); err == nil {
		t.Fatal("Unmarshal() of a encrypted key with nil callback succeeded, want error")
	}
}

func TestUnmarshalPassphraseError(t *testing.T) {
	data := mustMarshal(t, mustGenerate(t), testPassphrase)
	errCancelled := errors.New("user cancelled")
	_, err := Unmarshal(data, func() ([]byte, error) { return nil, errCancelled })
	if !errors.Is(err, errCancelled) {
		t.Fatalf("Unmarshal() error = %v, want it to wrap %v", err, errCancelled)
	}
}

func TestUnmarshalTampered(t *testing.T) {
	block := mustDecodePEM(t, mustMarshal(t, mustGenerate(t), testPassphrase))
	block.Bytes[len(block.Bytes)-1] ^= 0xff

	if _, err := Unmarshal(pem.EncodeToMemory(block), passphraseFunc(testPassphrase)); err == nil {
		t.Fatal("Unmarshal() of a tampered key succeeded, want error")
	}
}

func TestUnmarshalInvalid(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey() error: %v", err)
	}
	ecDER, err := x509.MarshalPKCS8PrivateKey(ecKey)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey() error: %v", err)
	}

	encrypted := mustDecodePEM(t, mustMarshal(t, mustGenerate(t), testPassphrase))
	withSalt := func(salt string) []byte {
		return pem.EncodeToMemory(&pem.Block{
			Type:    pemEncryptedType,
			Headers: map[string]string{saltHeader: salt},
			Bytes:   encrypted.Bytes,
		})
	}

	// None of these should get far enough to ask for a passphrase.
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"not PEM", []byte("hello")},
		{"unknown PEM type", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte{1, 2, 3}})},
		{"plain with garbage body", pem.EncodeToMemory(&pem.Block{Type: pemType, Bytes: []byte{1, 2, 3}})},
		{"plain non-ed25519 key", pem.EncodeToMemory(&pem.Block{Type: pemType, Bytes: ecDER})},
		{"encrypted without salt", pem.EncodeToMemory(&pem.Block{Type: pemEncryptedType, Bytes: encrypted.Bytes})},
		{"encrypted with invalid base64 salt", withSalt("!!!not base64!!!")},
		{"encrypted with short salt", withSalt("AAAA")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Unmarshal(tt.data, mustNotBeCalled(t)); err == nil {
				t.Fatal("Unmarshal() succeeded, want error")
			}
		})
	}
}

// --- Step 7: IsEncrypted ---

func TestIsEncrypted(t *testing.T) {
	id := mustGenerate(t)
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"encrypted", mustMarshal(t, id, testPassphrase), true},
		{"plain", mustMarshal(t, id, ""), false},
		{"not PEM", []byte("hello"), false},
		{"empty", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEncrypted(tt.data); got != tt.want {
				t.Fatalf("IsEncrypted() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Argon2id settings ---

// TestDeriveKeyKnownValue pins the Argon2id settings. Changing any of them
// makes every existing encrypted identity.key unreadable, so the change must
// fail a test instead of passing silently.
func TestDeriveKeyKnownValue(t *testing.T) {
	salt := make([]byte, saltLength)
	for i := range salt {
		salt[i] = byte(i)
	}
	const want = "8eb345c7823aa06ab0bab02a1b7138f6f30610b6c50f74d79c4fd9c777ed4b33"
	if got := hex.EncodeToString(deriveKey([]byte("test"), salt)); got != want {
		t.Fatalf("deriveKey() = %s, want %s\n"+
			"The Argon2id settings changed. Existing encrypted keys would no longer open.", got, want)
	}
}
