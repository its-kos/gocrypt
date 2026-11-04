package encryptor

import (
	"bytes"
	"crypto/rand"
	"testing"
)

// AES-GCM adds a 12-byte nonce in front and a 16-byte authentication tag at the end.
const (
	nonceSize = 12
	tagSize   = 16
)

func newKey() []byte {
	key := make([]byte, 32)
	rand.Read(key)
	return key
}

func mustEncrypt(t *testing.T, plaintext, key []byte) []byte {
	t.Helper()
	out, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}
	return out
}

func TestRoundTrip(t *testing.T) {
	big := make([]byte, 1<<20)
	rand.Read(big)

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello")},
		{"1 MiB random", big},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := newKey()
			got, err := Decrypt(mustEncrypt(t, tt.plaintext, key), key)
			if err != nil {
				t.Fatalf("Decrypt() error: %v", err)
			}
			if !bytes.Equal(got, tt.plaintext) {
				t.Fatal("decrypted data differs from the original")
			}
		})
	}
}

func TestOutputLayout(t *testing.T) {
	plaintext := []byte("some plaintext")
	out := mustEncrypt(t, plaintext, newKey())
	if want := nonceSize + len(plaintext) + tagSize; len(out) != want {
		t.Fatalf("output length = %d, want %d (nonce + plaintext + tag)", len(out), want)
	}
	if bytes.Contains(out, plaintext) {
		t.Fatal("output contains the plaintext")
	}
}

func TestEncryptIsRandomized(t *testing.T) {
	key := newKey()
	plaintext := []byte("same input twice")
	a := mustEncrypt(t, plaintext, key)
	b := mustEncrypt(t, plaintext, key)

	if bytes.Equal(a[:nonceSize], b[:nonceSize]) {
		t.Fatal("two encryptions used the same nonce")
	}
	if bytes.Equal(a, b) {
		t.Fatal("two encryptions of the same data produced the same output")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	out := mustEncrypt(t, []byte("secret"), newKey())
	if _, err := Decrypt(out, newKey()); err == nil {
		t.Fatal("Decrypt() with the wrong key succeeded, want error")
	}
}

func TestDecryptTampered(t *testing.T) {
	key := newKey()
	out := mustEncrypt(t, []byte("do not modify me"), key)

	// Flip one byte in each part: nonce, ciphertext, tag.
	positions := map[string]int{
		"nonce":      0,
		"ciphertext": nonceSize,
		"tag":        len(out) - 1,
	}
	for name, pos := range positions {
		t.Run(name, func(t *testing.T) {
			tampered := bytes.Clone(out)
			tampered[pos] ^= 0xff
			if _, err := Decrypt(tampered, key); err == nil {
				t.Fatalf("Decrypt() of data with a modified %s succeeded, want error", name)
			}
		})
	}
}

func TestDecryptTooShort(t *testing.T) {
	key := newKey()
	for _, n := range []int{0, 1, nonceSize, nonceSize + tagSize - 1} {
		if _, err := Decrypt(make([]byte, n), key); err == nil {
			t.Fatalf("Decrypt() of %d bytes succeeded, want error", n)
		}
	}
}

func TestInvalidKeySize(t *testing.T) {
	// AES only accepts 16, 24 or 32-byte keys.
	for _, n := range []int{0, 15, 31, 33, 64} {
		key := make([]byte, n)
		if _, err := Encrypt([]byte("x"), key); err == nil {
			t.Errorf("Encrypt() with a %d-byte key succeeded, want error", n)
		}
		if _, err := Decrypt(make([]byte, 64), key); err == nil {
			t.Errorf("Decrypt() with a %d-byte key succeeded, want error", n)
		}
	}
}
