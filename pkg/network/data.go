package network

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

type discoveryNotifee struct {
	host host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("I: %v, discovered peer: %s\n", n.host.ID().ShortString(), pi.ID.ShortString())

	// Add peer's addresses to the peerstore
	n.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)

	// Optionally, connect to the peer
	err := n.host.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("Failed to connect to peer %s: %v\n", pi.ID.ShortString(), err)
	} else {
		fmt.Printf("Connected to peer %s\n", pi.ID.ShortString())
	}
}
