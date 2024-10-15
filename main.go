package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Name:  "GoCrypt",
		Usage: "A P2P file encryption tool",
		Action: func(*cli.Context) error {
			fmt.Println("Ping! I say!")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "setup",
				Aliases: []string{"s"},
				Usage:   "Setup host for local node",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Setting up host in here...", cCtx.Args().First())
					return nil
				},
			},
			{
				Name:    "upload",
				Aliases: []string{"u"},
				Usage:   "Upload file to Gocrypt network",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Uploading file here...", cCtx.Args().First())
					return nil
				},
			},
			{
				Name:    "retrieve",
				Aliases: []string{"r"},
				Usage:   "Retrieve file from Gocrypt network",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Retrieving file in here...", cCtx.Args().First())
					return nil
				},
			}, {
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List all running nodes of the Gocrypt network",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Listing nodes in here...", cCtx.Args().First())
					return nil
				},
			},
			{
				Name:    "help",
				Aliases: []string{"h"},
				Usage:   "Print help message",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Printing help message here...", cCtx.Args().First())
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	// filePath := "./files/testfile.txt"
	// chunkSize := 1024 // 1 KB chunks

	// var wg sync.WaitGroup
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

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
