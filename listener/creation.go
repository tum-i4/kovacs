package main

import (
	"bufio"
	"sync/atomic"

	"github.com/libp2p/go-libp2p-core/network"
	"node/constants"
	"node/logging"
	"node/p2p"
)

var connectionCount int64 = 0 //nolint:revive
var ownSignedIdentityCard p2p.SignedMessage

// Adapted from https://github.com/libp2p/go-libp2p/tree/v0.16.0/examples/chat-with-mdns
func createNode(port int) {
	var err error

	ownSignedIdentityCard, err = p2p.LoadSignedIdentityCard(&globalPrivateKey)
	if err != nil {
		log.Error.Fatalln(err)
	}

	h, err := p2p.MakeHost(port)
	if err != nil {
		log.Error.Fatalln(err)
	}

	h.SetStreamHandler(constants.P2PProtocolName, handleListenerStream)

	_ = p2p.InitMDNS(h)

	select {} // wait here
}

func handleListenerStream(s network.Stream) {
	current := atomic.AddInt64(&connectionCount, 1)

	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go streamHandler(rw, current)
}
