package compressor

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func mustCompress(t *testing.T, data []byte) []byte {
	t.Helper()
	out, err := Compress(data)
	if err != nil {
		t.Fatalf("Compress() error: %v", err)
	}
	return out
}

func TestRoundTrip(t *testing.T) {
	random := make([]byte, 64*1024)
	rand.Read(random)

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"text", []byte("the quick brown fox jumps over the lazy dog")},
		{"random (incompressible)", random},
		{"repetitive", bytes.Repeat([]byte("abc"), 100_000)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decompress(mustCompress(t, tt.data))
			if err != nil {
				t.Fatalf("Decompress() error: %v", err)
			}
			if !bytes.Equal(got, tt.data) {
				t.Fatal("decompressed data differs from the original")
			}
		})
	}
}

func TestCompressShrinksRepetitiveData(t *testing.T) {
	data := bytes.Repeat([]byte("a"), 1<<20)
	if out := mustCompress(t, data); len(out) > len(data)/100 {
		t.Fatalf("1 MiB of 'a' compressed to %d bytes, want under 1%% of the input", len(out))
	}
}

func TestDecompressInvalid(t *testing.T) {
	valid := mustCompress(t, bytes.Repeat([]byte("hello "), 1000))

	corrupted := bytes.Clone(valid)
	corrupted[len(corrupted)-5] ^= 0xff // inside the CRC-32 checksum at the end

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"not gzip", []byte("plain text, not gzip")},
		{"truncated", valid[:len(valid)/2]},
		{"bad checksum", corrupted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Decompress(tt.data); err == nil {
				t.Fatal("Decompress() succeeded, want error")
			}
		})
	}
}
