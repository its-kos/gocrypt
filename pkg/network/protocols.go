package network

import (
	"fmt"
	"io"
	"log"

	"github.com/its-kos/gocrypt/pkg/encryption"
	"github.com/its-kos/gocrypt/pkg/filechunk"
	"github.com/libp2p/go-libp2p/core/network"
)

func HandleChunkStream(stream network.Stream) {
	buf := make([]byte, 1024)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from stream:", err)
			}
			break
		}
		fmt.Printf("Received %d bytes: %v\n", n, buf[:n])
	}

	decrypted := make([][]byte, 0)
	for _, encChunk := range buf {
		decryptedChunk, err := encryption.DecryptChunk(encChunk, key)
		if err != nil {
			log.Fatalf("Error decrypting chunk: %v", err)
		}
		decrypted = append(decrypted, decryptedChunk)
	}

	err := filechunk.StitchFile(decrypted, "./files/testfile_reconstructed.txt")
	if err != nil {
		log.Fatalf("Error splitting file: %v", err)
	}

	stream.Close()
}
