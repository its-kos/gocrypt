package main

import (
	"log"

	"github.com/its-kos/gocrypt/pkg/filechunk"
)

func main() {

	filePath := "testfile.jpg"
	//outputPath := "outputfile.txt"

	chunkSize := 1024 // 1 KB chunks

	//encryptionKey := []byte("myverystrongpasswordo32bitlength")  // AES-256 requires 32 bytes

	// Step 1: Split the file into chunks
	_, err := filechunk.Chunk(filePath, chunkSize)
	if err != nil {
		log.Fatalf("Error splitting file: %v", err)
	}




	// Step 2: Encrypt each chunk
	// var encryptedChunks [][]byte
	// for _, chunk := range chunks {
	// 	encryptedChunk, err := encryption.EncryptChunk(chunk, encryptionKey)
	// 	if err != nil {
	// 		log.Fatalf("Error encrypting chunk: %v", err)
	// 	}
	// 	encryptedChunks = append(encryptedChunks, encryptedChunk)
	// }
	// fmt.Println("Chunks encrypted successfully.")

	// // Step 3: Decrypt the chunks (just for testing)
	// var decryptedChunks [][]byte
	// for _, encryptedChunk := range encryptedChunks {
	// 	decryptedChunk, err := encryption.DecryptChunk(encryptedChunk, encryptionKey)
	// 	if err != nil {
	// 		log.Fatalf("Error decrypting chunk: %v", err)
	// 	}
	// 	decryptedChunks = append(decryptedChunks, decryptedChunk)
	// }
	// fmt.Println("Chunks decrypted successfully.")

	// // Step 4: Merge the decrypted chunks back into a file
	// err = filechunk.MergeChunks(decryptedChunks, outputPath)
	// if err != nil {
	// 	log.Fatalf("Error merging chunks: %v", err)
	// }
	// fmt.Printf("File reassembled into: %s\n", outputPath)
}
