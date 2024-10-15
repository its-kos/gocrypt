package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/its-kos/gocrypt/pkg/encryption"
	"github.com/its-kos/gocrypt/pkg/filechunk"
	"github.com/its-kos/gocrypt/pkg/network"
	"github.com/urfave/cli/v2"
)

func main() {

	var filePath, chunkSize string

	// ! This is highly stupid but offers temporary ease of use
	// Key should be stored somewhere safe and retrieved for each file
	var key = make([]byte, 32)
	//

	app := &cli.App{
		Name:  "GoCrypt",
		Usage: "A P2P file encryption tool",
		Commands: []*cli.Command{
			{
				Name:    "setup",
				Aliases: []string{"s"},
				Usage:   "Setup host for local node",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "addr",
						Aliases: []string{"a"},
						Value:   "/ip4/127.0.0.1/tcp/0",
						Usage:   "Address for host to listen to",
					},
				},
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Setting up host in here...", cCtx.Args().First())
					_, err := network.StartNode(cCtx.String("addr"))
					if err != nil {
						log.Fatalf("Error creating host: %v\n", err)
					}
					return nil
				},
			},
			{
				Name:    "upload",
				Aliases: []string{"u"},
				Usage:   "Upload file to Gocrypt network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "path",
						Aliases:     []string{"p"},
						Usage:       "Path of the file to upload",
						Destination: &filePath,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "chunk",
						Aliases:     []string{"c"},
						Value:       "1024", // We default to 1 KB
						Usage:       "Chunksize for file splitting",
						Destination: &chunkSize,
						Required:    false,
					},
				},
				Action: func(cCtx *cli.Context) error {
					chunkSize, err := strconv.Atoi(chunkSize)
					if err != nil {
						return fmt.Errorf("invalid chunk size: %v", err)
					}

					fmt.Printf("Uploading file %v, split into %v byte chunks here...\n", filePath, chunkSize)

					_, err = rand.Reader.Read(key)
					if err != nil {
						return fmt.Errorf("error generating random key: %v", err)
					}

					chunks, err := filechunk.ChunkFile(filePath, chunkSize)
					if err != nil {
						return fmt.Errorf("error splitting file: %v", err)
					}

					encrypted := make([][]byte, 0)
					for _, chunk := range chunks {
						encryptedChunk, err := encryption.EncryptChunk(chunk, key)
						if err != nil {
							return fmt.Errorf("error encrypting chunk: %v", err)
						}
						encrypted = append(encrypted, encryptedChunk)
					}

					// This block is only used for testing and should be removed at some point
					err = filechunk.StitchFile(encrypted, "./files/testfile_encrypted.txt")
					if err != nil {
						log.Fatalf("Error splitting file: %v", err)
					}
					//

					return nil
				},
			},
			{
				Name:    "retrieve",
				Aliases: []string{"r"},
				Usage:   "Retrieve file from Gocrypt network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "path",
						Aliases:     []string{"p"},
						Usage:       "Path of the save location of the file",
						Destination: &filePath,
						Required:    true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Connecting to other nodes to get file chunks here...")
					decrypted := make([][]byte, 0)
					for _, encChunk := range encrypted {
						decryptedChunk, err := encryption.DecryptChunk(encChunk, key)
						if err != nil {
							return fmt.Errorf("error decrypting chunk: %v", err)
						}
						decrypted = append(decrypted, decryptedChunk)
					}

					err := filechunk.StitchFile(decrypted, filePath)
					if err != nil {
						return fmt.Errorf("error splitting file: %v", err)
					}
					fmt.Println("Retrieving file in here...")
					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List all running nodes of the Gocrypt network",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("Listing nodes in here...")
					return nil
				},
			},
		},
		CommandNotFound: func(c *cli.Context, command string) {
			fmt.Printf("Unrecognized command: '%s'\n\n", command)
			cli.ShowAppHelp(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	// var wg sync.WaitGroup
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

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
