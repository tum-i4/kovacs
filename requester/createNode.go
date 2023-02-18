package main

import (
	"bufio"
	"context"
	"time"

	libPeer "github.com/libp2p/go-libp2p-core/peer"
	"node/constants"
	log "node/logging"
	"node/p2p"
)

// Adapted from https://github.com/libp2p/go-libp2p/tree/v0.16.0/examples/chat-with-mdns
func createNode(config *configuration, peerStart *time.Time) {
	start := time.Now()
	signedIdentityCard, err := p2p.LoadSignedIdentityCard(&globalPrivateKey)
	if err != nil {
		log.Error.Fatalln(err)
	}
	loadIdDuration := time.Since(start)

	start = time.Now()
	ctx := context.Background()

	h, err := p2p.MakeHost(config.port)
	if err != nil {
		log.Error.Fatalln(err)
	}
	peerCreationDuration := time.Since(start)

	log.Info.Println("Starting search")
	peerSearchStart := time.Now()

	peerChan := p2p.InitMDNS(h)
	endTime := time.Now().Add(maxSearchTime)
	resetCount := 0

	var peer libPeer.AddrInfo
	for {
		select {
		case tmp := <-peerChan:
			peer = tmp
		case <-exchangeFailed:
			// The correct peet was found but the exchange failed for some reason. Thus, we restart the search.
			// Since go has no fallthrough for select stmts, the code from the time-out case was copied here
			if resetCount >= maxRetries {
				log.Error.Printf("Did not find peer even after retrying for %d times\n", resetCount)
				realDone <- returnValue{
					success: false,
					value:   "",
				}

				break
			}

			peerChan = p2p.InitMDNS(h)
			endTime = time.Now().Add(maxSearchTime)
			resetCount++
			log.Info.Printf("Exchange failed. Restarting the search process for the %d. time\n", resetCount)

			if config.enableFakeChatter {
				// Sleep for 2 seconds to avoid DoSing the network
				time.Sleep(2 * time.Second)
			}
		case <-time.After(maxSearchTime):
			// When using fake chatter, the node occasionally does not find the correct peer before timing out.
			// Thus, we restart the search and the timeout duration for a total of three times.
			if resetCount >= maxRetries {
				log.Error.Printf("Did not find peer even after retrying for %d times\n", resetCount)
				realDone <- returnValue{
					success: false,
					value:   "",
				}

				break
			}

			peerChan = p2p.InitMDNS(h)
			endTime = time.Now().Add(maxSearchTime)
			resetCount++
			log.Info.Printf("Did not find peer. Restarting the search process for the %d. time\n", resetCount)
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
			ret := returnValue{
				loadIDCardDuration:   loadIdDuration,
				hostCreationDuration: peerCreationDuration,
				searchRestarts:       resetCount,
			}

			go streamHandler(rw, config, &signedIdentityCard, peerStart, &peerSearchStart, &ret)
		}
	}
}
