package main

import (
	"bufio"
	"context"

	"node/constants"
	"node/p2p"
	"node/random"
)

func getFreePort(portsInUse []int) int {
	const minPort = 5000
	const maxPort = 60000

	port := random.PositiveIntFromRange(minPort, maxPort)

	for _, currentPort := range portsInUse {
		if currentPort == port {
			// Port is already in use
			return getFreePort(portsInUse)
		}
	}

	// Port is free
	return port
}

func createFakeNodes(usedPort int) {
	portsInUse := []int{usedPort}

	for i := 0; i < 8; i++ {
		port := getFreePort(portsInUse)
		portsInUse = append(portsInUse, port)
		go createFakeNode(port)
	}
}

// Adapted from https://github.com/libp2p/go-libp2p/tree/v0.16.0/examples/chat-with-mdns
func createFakeNode(port int) {
	ctx := context.Background()

	h, err := p2p.MakeHost(port)
	if err != nil {
		return
	}

	peerChan := p2p.InitMDNS(h)

	for {
		if foundCorrectPeer {
			terminationIncreaseInPercent = 10

			// Roll dice to determine if we should end
			dice := random.PositiveIntFromRange(0, 101)
			if int32(dice) < terminationChanceInPercent {
				if len(fakeDone) == 0 {
					fakeDone <- true
				}

				return
			}
		}

		peer := <-peerChan

		if err := h.Connect(ctx, peer); err != nil {
			continue
		}

		// open a stream, this stream will be handled by handleStream other end
		stream, err := h.NewStream(ctx, peer.ID, constants.P2PProtocolName)
		if err == nil {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
			go fakeChatter(rw)
		}
	}
}
