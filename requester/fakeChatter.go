package main

import (
	"bufio"
	"sync/atomic"

	"node/constants"
	nP "node/nonRepudiation"
	"node/p2p"
	"node/random"
)

func fakeChatter(rw *bufio.ReadWriter) {
	// Random RSA key pair that will be used to sign all messages
	privateKey, err := nP.GenerateRSAPrivateKey()
	if err != nil {
		return
	}

	// Parse owner's identity card
	_, identityCard, _, err := p2p.ReceiveAndVerifySignedIdentityCard(rw, &revoloriPublicKey)
	if err != nil {
		return
	}

	// Send an empty identity card
	err = p2p.SendEmptyIdentityCard(&privateKey, rw)
	if err != nil {
		return
	}

	// Send datum request
	request := p2p.FirstMessage{
		Datum:         random.String(random.PositiveIntFromRange(16, 64)),
		Justification: "FakeChatter",
		PublicKey:     privateKey.PublicKey,
		Type:          constants.MessageTypeFakeChatter,
	}

	err = p2p.CreateAndSendSignedMessage(request, &privateKey, rw)
	if err != nil {
		return
	}

	// Receive the response with the encrypted message
	// Increased wait longer in case that the encryption or file I/O take some time
	var firstMessageResponse p2p.FirstMessage
	signedMessage, err := p2p.ReceiveAndVerifySignedMessage(rw, &identityCard.PublicKey, &firstMessageResponse, constants.MaxWaitTime*3)
	if err != nil {
		return
	}

	err = firstMessageResponse.CheckForContent()
	if err != nil {
		return
	}

	// Extract owner's public key that will be used to verify the following messages
	ownerPublicKey := firstMessageResponse.PublicKey

	// Send acknowledgment for the encrypted data
	ack, err := createAck(signedMessage, 0)
	if err != nil {
		return
	}

	err = p2p.CreateAndSendSignedMessage(ack, &privateKey, rw)
	if err != nil {
		return
	}

	// Store all data
	var data nP.Data
	for currentID := 1; ; currentID++ {
		// Read data
		signedMessage, err = p2p.ReceiveAndVerifySignedMessage(rw, &ownerPublicKey, &data)
		if err != nil {
			break
		}

		// Send an acknowledgment
		ack, err = createAck(signedMessage, currentID)
		if err != nil {
			return
		}

		err = p2p.CreateAndSendSignedMessage(ack, &privateKey, rw)
		if err != nil {
			return
		}
	}

	if atomic.AddInt32(&fakeConnectionsAmount, 1) == minFakeConnectionCount {
		atomic.AddInt32(&terminationChanceInPercent, terminationIncreaseInPercent)
		if len(fakeDone) == 0 {
			fakeDone <- true
		}
	}
}
