package network

import (
	"github.com/its-kos/gocrypt/pkg/utils"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

func StartNode(listenAddr string, conf utils.Config) (host.Host, error) {
	_, pk, err := utils.ReadKeys(conf)
	if err != nil {
		return nil, err
	}

	node, err := libp2p.New(
		libp2p.ListenAddrStrings(listenAddr),
		libp2p.Identity(pk),
		//libp2p.EnableNATService(),
	)
	if err != nil {
		return nil, err
	}

	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}

	_, err = peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return nil, err
	}
	// fmt.Println("Host node address:", addrs[0])
	// fmt.Println("Host node ID:", node.ID())

	return node, nil
}

func SetupMDNS(node host.Host) error {
	notifee := &discoveryNotifee{host: node}

	mdnsService := mdns.NewMdnsService(node, "gocrypt", notifee)
	mdnsService.Start()
	return nil
}
