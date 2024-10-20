package network

import (
	"fmt"
	"io"
	"log"

	"github.com/its-kos/gocrypt/pkg/encryption"
	"github.com/its-kos/gocrypt/pkg/filechunk"
	"github.com/its-kos/gocrypt/pkg/utils"
	"github.com/libp2p/go-libp2p/core/network"
)

func HandleChunkStream(stream network.Stream) {
	conf, err := utils.SetupConfig()
	if err != nil {
		log.Fatalf("Error decrypting chunk: %v", err)
	}

	_, _, cKey, err := utils.ReadKeys(conf)
	if err != nil {
		log.Fatalf("Error decrypting chunk: %v", err)
	}

	decrypted := make([][]byte, 0)
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

	decryptedChunk, err := encryption.DecryptChunk(buf, cKey)
	if err != nil {
		log.Fatalf("Error decrypting chunk: %v", err)
	}
	decrypted = append(decrypted, decryptedChunk)

	err = filechunk.StitchFile(decrypted, "./files/testfile_reconstructed.txt")
	if err != nil {
		log.Fatalf("Error splitting file: %v", err)
	}

	stream.Close()
}
