package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"time"

	log "node/logging"
	"node/p2p"
)

func streamHandler(rw *bufio.ReadWriter, config *configuration, ownSignedIDCard *p2p.SignedMessage, peerStart *time.Time, peerSearchStart *time.Time, ret *returnValue) {
	if foundCorrectPeer && !config.enableFakeChatter {
		return
	}

	// This slice will be used to store all signed messages
	signedMessages := make([]p2p.SignedMessage, 0)

	// Parse owner's identity card
	idVerificationStart := time.Now()
	signedIdentityCard, listenerIdentityCard, isFakeChatter, err := p2p.ReceiveAndVerifySignedIdentityCard(rw, &revoloriPublicKey)
	if err != nil {
		log.Error.Printf("requester/streamHandler - Could not parse identity card: %s\n", err)
		return
	} else if isFakeChatter {
		log.Error.Printf("requester/streamHandler - Owner send ID Card marked as fake chatter?\n")
		return
	}
	signedMessages = append(signedMessages, signedIdentityCard)

	if listenerIdentityCard.SSOID != config.ssoid {
		if config.enableFakeChatter {
			fakeChatter(rw, &listenerIdentityCard)
		}

		return
	} else if !foundCorrectPeer {
		// If this check did not exist, a single data request could lead to multiple usage logs.
		foundCorrectPeer = true
		ret.peerSearchDuration = time.Since(*peerSearchStart)
		realExchange(rw, config, ownSignedIDCard, &listenerIdentityCard, signedMessages, &idVerificationStart, peerStart, ret)
	}
}

func createAck(signedMessage p2p.SignedMessage, currentID int) (p2p.Acknowledgement, error) {
	ackContent, err := json.Marshal(signedMessage)
	if err != nil {
		return p2p.Acknowledgement{}, errors.New("requester/createAck - Could not marshal received signed message")
	}

	return p2p.Acknowledgement{
		ID:        currentID,
		TimeStamp: time.Now().Unix(),
		Content:   ackContent,
	}, nil
}
