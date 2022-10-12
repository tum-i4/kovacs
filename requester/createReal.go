package main

import (
	"bufio"
	"context"
	"time"

	libPeer "github.com/libp2p/go-libp2p-core/peer"
	"node/constants"
	log "node/logging"
	"node/p2p"
	"node/random"
)

// Adapted from https://github.com/libp2p/go-libp2p/tree/v0.16.0/examples/chat-with-mdns
func createRealNode(config configuration) {
	// Should be random to prevent an attacker of knowing which connection is real
	time.Sleep(time.Duration(random.PositiveIntFromRange(2, 6)) * time.Second)
	signedIdentityCard, err := p2p.LoadSignedIdentityCard(&globalPrivateKey)
	if err != nil {
		log.Error.Fatalln(err)
	}

	ctx := context.Background()

	h, err := p2p.MakeHost(config.port)
	if err != nil {
		log.Error.Fatalln(err)
	}

	peerChan := p2p.InitMDNS(h)
	endTime := time.Now().Add(maxSearchTime)
	resetCount := 0

	var peer libPeer.AddrInfo
	for {
		if foundCorrectPeer {
			break
		}

		select {
		case tmp := <-peerChan:
			peer = tmp
		case <-time.After(maxSearchTime):
			// When using fake chatter, the node occasionally does not find the correct peer before timing out.
			// Thus, we restart the search and the timeout duration for a total of three times.
			if resetCount >= 3 {
				break
			}

			peerChan = p2p.InitMDNS(h)
			endTime = time.Now().Add(maxSearchTime)
			resetCount++
			log.Info.Printf("Did not find peer. Restarting the search process for the %d. time\n", resetCount)
		}

		if foundCorrectPeer {
			break
		}

		if time.Now().After(endTime) {
			log.Error.Fatalf("Could not find the requested user after restarting the search %d times\n", resetCount)
		}

		if err := h.Connect(ctx, peer); err != nil {
			if err.Error() == "dial to self attempted" {
				continue
			}

			log.Info.Printf("Connection failed: %s\n", err)
		}

		// open a stream, this stream will be handled by handleStream other end
		stream, err := h.NewStream(ctx, peer.ID, constants.P2PProtocolName)

		if err != nil {
			if err.Error() == "protocol not supported" {
				continue
			}

			log.Info.Printf("Stream open failed: %s", err)
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
			go streamHandler(rw, config.ssoid, config.justification, config.requestedDatum, signedIdentityCard)
		}
	}
}
