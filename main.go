package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/its-kos/gocrypt/pkg/network"
	"github.com/its-kos/gocrypt/pkg/utils"

	libnet "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

func handleStream(stream libnet.Stream) {
	fmt.Println("Got a new stream!")

	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

type config struct {
	RendezvousString string
	ProtocolID       string
	listenHost       string
	listenPort       int
}

func main() {

	// filePath := "./files/testfile.txt"
	// chunkSize := 1024 // 1 KB chunks
	// var wg sync.WaitGroup
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	c := &config{}

	// flag.StringVar(&c.RendezvousString, "rendezvous", "meetme", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	// flag.StringVar(&c.listenHost, "host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.IntVar(&c.listenPort, "port", 0, "node listen port (0 pick a random unused port)")
	flag.Parse()

	ctx := context.Background()
	//r := rand.Reader

	conf, err := utils.SetupConfig()
	if err != nil {
		log.Fatalf("Error creating config files: %v\n", err)
	}

	host, err := network.StartNode(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", c.listenPort), *conf)
	if err != nil {
		log.Fatalf("Error creating Host node: %v\n", err)
	}
	fmt.Printf("Successfully initialized host: %v", host.ID().ShortString())

	// Set a function as stream handler.
	// This function is called when a peer initiates a connection and starts a stream with this peer.
	host.SetStreamHandler(protocol.ID("/chat/1.1.0"), handleStream)

	peerChan := network.InitMDNS(host, "gocrypt")

	for { // allows multiple peers to join
		peer := <-peerChan // will block until we discover a peer
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
		stream, err := host.NewStream(ctx, peer.ID, protocol.ID("/chat/1.1.0"))

		if err != nil {
			fmt.Println("Stream open failed", err)
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			go writeData(rw)
			go readData(rw)
			fmt.Println("Connected to:", peer)
		}
	}
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
