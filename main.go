package main

import (
	"fmt"
	"log"

	"github.com/its-kos/gocrypt/pkg/filechunk"
)

func main() {

	filePath := "testfile.jpg"

	chunkSize := 1024 // 1 KB chunks

	//encryptionKey := []byte("")  // AES-256 requires 32 bytes

	chunks, err := filechunk.Chunk(filePath, chunkSize)
	if err != nil {
		log.Fatalf("Error splitting file: %v", err)
	}
	fmt.Print("Chunks: ", chunks)
}
