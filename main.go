package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/its-kos/gocrypt/pkg/encryption"
	"github.com/its-kos/gocrypt/pkg/filechunk"
	"github.com/its-kos/gocrypt/pkg/network"
	"github.com/its-kos/gocrypt/pkg/utils"

	libnet "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func handleStream(stream libnet.Stream) {
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

type config struct {
	listenHost string
	listenPort int
}

func main() {

	filePath := "./files/testfile.txt"
	chunkSize := 1024 // 1 KB chunks

	c := &config{}

	flag.StringVar(&c.listenHost, "host", "", "whatever")
	flag.IntVar(&c.listenPort, "port", 0, "node listen port (0 pick a random unused port)")
	flag.Parse()

	ctx := context.Background()
	//r := rand.Reader

	conf, err := utils.SetupConfig()
	if err != nil {
		log.Fatalf("Error creating config files: %v\n", err)
	}

	host, err := network.StartNode(fmt.Sprintf("/ip4/%s/tcp/%d", c.listenHost, c.listenPort), *conf)
	if err != nil {
		log.Fatalf("Error creating Host node: %v\n", err)
	}
	fmt.Printf("Successfully initialized host: %v", host.ID().ShortString())

	host.SetStreamHandler(protocol.ID("/gocrypt/chunk-transfer/1.0.0"), handleStream)

	key := make([]byte, 32)
	_, err = rand.Reader.Read(key)
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

	peerChan := network.InitMDNS(host, "gocrypt")

	for {
		peer := <-peerChan // block until we discover a peer
		if peer.ID > host.ID() {
			// if other end peer id greater than us, don't connect to it, just wait for it to connect us
			fmt.Println("Found peer:", peer, " id is greater than us, wait for it to connect to us")
			continue
		}
		fmt.Println("Found peer:", peer, ", connecting")

		if err := host.Connect(ctx, peer); err != nil {
			fmt.Println("Connection failed:", err)
			continue
		}

		// open a stream, this stream will be handled by handleStream other end
		stream, err := host.NewStream(ctx, peer.ID, protocol.ID("/gocrypt/chunk-transfer/1.0.0"))
		if err != nil {
			fmt.Println("Stream open failed", err)
		} else {
			fmt.Println("Connected to:", peer)

			stream.Write(encrypted[0])
			stream.Close()
		}
	}

	// wg.Wait()
	// The functionality is stupid but i'm doing it for
	// ease of development. Right now each node stores their
	// PK and uses it to regenerate the same Node ID upon
	// node start.

	// In the future this will be changed to each node
	// communicating the locally saved chunks to the entire
	// network when it gets online so the DHT will be updated upon
	// every note reconnect. More overhead but a better solution.
}
