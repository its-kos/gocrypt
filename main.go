package main

import (
	"fmt"
	"log"

	"github.com/its-kos/gocrypt/pkg/network"
	"github.com/its-kos/gocrypt/pkg/utils"
)

func main() {

	//filePath := "./files/testfile.txt"
	//chunkSize := 1024 // 1 KB chunks
	//var wg sync.WaitGroup
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	conf, err := utils.SetupConfig()
	if err != nil {
		log.Fatalf("Error creating config files: %v\n", err)
	}

	host, err := network.StartNode("/ip4/127.0.0.1/tcp/0", *conf)
	if err != nil {
		log.Fatalf("Error creating Host node: %v\n", err)
	}
	fmt.Printf("Successfully initialized host from config: %v", host.ID().ShortString())

	// sigs := make(chan os.Signal, 1) // Buffered cause we don't wanna block, only 1 SIG is enough.
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	<-sigs
	// 	fmt.Printf("Received shutdown signal. Closing host %v...\n", host.ID().ShortString())
	// 	host.Close()
	// 	cancel()
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		default:
	// 			time.Sleep(2 * time.Second)
	// 			fmt.Printf("Host %v is running...\n", host.ID().ShortString())
	// 		}
	// 	}
	// }()

	// wg.Wait()
	// The functionality is stupid but i'm doing it for
	// ease of development. Right now each node stores their
	// PK and uses it to regenerate the same Node ID upon
	// node start.

	// In the future this will be changed to each node
	// communicating the locally saved chunks to the entire
	// network when it gets online so the DHT will be updated upon
	// every note reconnect. More overhead but a better solution.

	// key := make([]byte, 32)
	// _, err := rand.Reader.Read(key)
	// if err != nil {
	// 	log.Fatalf("Error generating random key: %v", err)
	// }

	// chunks, err := filechunk.ChunkFile(filePath, chunkSize)
	// if err != nil {
	// 	log.Fatalf("Error splitting file: %v", err)
	// }

	// encrypted := make([][]byte, 0)
	// for _, chunk := range chunks {
	// 	encryptedChunk, err := encryption.EncryptChunk(chunk, key)
	// 	if err != nil {
	// 		log.Fatalf("Error encrypting chunk: %v", err)
	// 	}
	// 	encrypted = append(encrypted, encryptedChunk)
	// }

	// err = filechunk.StitchFile(encrypted, "./files/testfile_encrypted.txt")
	// if err != nil {
	// 	log.Fatalf("Error splitting file: %v", err)
	// }

	// decrypted := make([][]byte, 0)
	// for _, encChunk := range encrypted {
	// 	decryptedChunk, err := encryption.DecryptChunk(encChunk, key)
	// 	if err != nil {
	// 		log.Fatalf("Error decrypting chunk: %v", err)
	// 	}
	// 	decrypted = append(decrypted, decryptedChunk)
	// }

	// err = filechunk.StitchFile(decrypted, "./files/testfile_reconstructed.txt")
	// if err != nil {
	// 	log.Fatalf("Error splitting file: %v", err)
	// }

	// host, err := network.StartNode("/ip4/127.0.0.1/tcp/0")
	// if err != nil {
	// 	log.Fatalf("Error creating Host node: %v\n", err)
	// }

	// sigs := make(chan os.Signal, 1) // Buffered cause we don't wanna block, only 1 SIG is enough.
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	<-sigs
	// 	fmt.Printf("Received shutdown signal. Closing host %v...\n", host.ID().ShortString())
	// 	host.Close()
	// 	cancel()
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		default:
	// 			time.Sleep(2 * time.Second)
	// 			fmt.Printf("Host %v is running...\n", host.ID().ShortString())
	// 		}
	// 	}
	// }()

	// wg.Wait()
	// fmt.Println("Exiting program.")

	//network.Connect(ctx, host, "test")
}
