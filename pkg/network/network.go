package network

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/its-kos/gocrypt/pkg/utils"
)

func StartNode(listenAddr string, conf Config) (host.Host, error) {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings(listenAddr),
		libp2p.
	)
	if err != nil {
		panic(err)
	}

	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}

	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return nil, err
	}
	fmt.Println("Host node address:", addrs[0])
	fmt.Println("Host node ID:", node.ID())

	return node, nil
}

func Connect(ctx context.Context, host host.Host, destAddr string) error {
	destMultiAddr, err := multiaddr.NewMultiaddr(destAddr)
	if err != nil {
		return err
	}
	destPeer, err := peerstore.AddrInfoFromP2pAddr(destMultiAddr)
	if err != nil {
		return err
	}

	if err := host.Connect(ctx, *destPeer); err != nil {
		return err
	}

	fmt.Printf("Successfully connected to peer: %s\n", destPeer.ID.ShortString())
	return nil
}

// func Ping(node host.Host, addr string) {

// 	fmt.Println("sending 5 ping messages to", addr)
// 	ch := pingService.Ping(context.Background(), peer.ID)
// 	for i := 0; i < 5; i++ {
// 		res := <-ch
// 		fmt.Println("pinged", addr, "in", res.RTT)
// 	}
// }
