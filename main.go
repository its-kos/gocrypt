package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/its-kos/gocrypt/pkg/encryption"
	"github.com/its-kos/gocrypt/pkg/filechunk"
	"github.com/its-kos/gocrypt/pkg/network"
	"github.com/its-kos/gocrypt/pkg/utils"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/urfave/cli/v2"
)

type config struct {
	listenHost string
	listenPort int
}

func main() {

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

	filePath := "./files/testfile.jpg"
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

	host.SetStreamHandler(protocol.ID("/gocrypt/chunk-transfer/1.0.0"), network.HandleChunkStream)

	chunks, err := filechunk.ChunkFile(filePath, chunkSize)
	if err != nil {
		log.Fatalf("Error splitting file: %v", err)
	}

	_, _, key, err := utils.ReadKeys(conf)
	if err != nil {
		log.Fatalf("Error generating cipher key: %v", err)
	}

	encrypted := make([][]byte, 0)
	for _, chunk := range chunks {
		encryptedChunk, err := encryption.EncryptChunk(chunk, key)
		if err != nil {
			return fmt.Errorf("error encrypting chunk: %v", err)
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

		stream, err := host.NewStream(ctx, peer.ID, protocol.ID("/gocrypt/chunk-transfer/1.0.0"))
		if err != nil {
			fmt.Println("Stream open failed", err)
		} else {
			fmt.Println("Connected to:", peer)

			for _, b := range encrypted {
				stream.Write(b)
			}
			stream.Close()
		}
	}

	// The functionality is stupid but i'm doing it for
	// ease of development. Right now each node stores their
	// PK and uses it to regenerate the same Node ID upon
	// node start.

	// In the future this will be changed to each node
	// communicating the locally saved chunks to the entire
	// network when it gets online so the DHT will be updated upon
	// every note reconnect. More overhead but a better solution.
}
