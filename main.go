package main

import (
	"crypto/rand"
	"log"

	"github.com/its-kos/gocrypt/pkg/encryption"
	"github.com/its-kos/gocrypt/pkg/filechunk"
)

func main() {

	filePath := "./files/testfile.txt"

	chunkSize := 1024 // 1 KB chunks

	key := make([]byte, 32)
	_, err := rand.Reader.Read(key)
	if err != nil {
		log.Fatalf("Error generating random key: %v", err)
	}

	chunks, err := filechunk.ChunkFile(filePath, chunkSize)
	if err != nil {
		log.Fatalf("Error splitting file: %v", err)
	}

	encrypted := make([][]byte, 0)
	for _, chunk := range chunks {
		encryptedChunk, err := encryption.EncryptChunk(chunk, key)
		if err != nil {
			log.Fatalf("Error encrypting chunk: %v", err)
		}
		encrypted = append(encrypted, encryptedChunk)
	}

	err = filechunk.StitchFile(encrypted, "testfile_encrypted.txt")
	if err != nil {
		log.Fatalf("Error splitting file: %v", err)
	}

	decrypted := make([][]byte, 0)
	for _, encChunk := range encrypted {
		decryptedChunk, err := encryption.DecryptChunk(encChunk, key)
		if err != nil {
			log.Fatalf("Error decrypting chunk: %v", err)
		}
		decrypted = append(decrypted, decryptedChunk)
	}

	err = filechunk.StitchFile(decrypted, "testfile_reconstructed.txt")
	if err != nil {
		log.Fatalf("Error splitting file: %v", err)
	}
}
