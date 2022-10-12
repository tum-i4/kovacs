package p2p

import (
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"node/logging"
)

// Copied from https://github.com/libp2p/go-libp2p/tree/v0.16.0/examples/chat-with-mdns

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

// HandlePeerFound to be called when a new peer is found.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

// InitMDNS initializes the MDNS service.
func InitMDNS(peerhost host.Host) chan peer.AddrInfo {
	// register with service so that we get notified about peer discovery
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo, 512)

	// An hour might be a long, long period in practical applications. But this is fine for us
	ser := mdns.NewMdnsService(peerhost, "serviceName", n)
	if err := ser.Start(); err != nil {
		log.Error.Fatalf("node/mdns - %v\n", err)
	}
	return n.PeerChan
}
