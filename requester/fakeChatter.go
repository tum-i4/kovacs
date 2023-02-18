package main

import (
	"bufio"
	"node"
	"sync/atomic"

	"node/constants"
	ownLog "node/logging"
	nP "node/nonRepudiation"
	"node/p2p"
	"node/random"
)

func fakeChatter(rw *bufio.ReadWriter, listenerIdentityCard *p2p.IdentityCard) {
	debugFakeChatter := false

	// Random RSA key pair that will be used to sign all messages
	privateKey, err := nP.GenerateRSAPrivateKey()
	if err != nil {
		if debugFakeChatter {
			ownLog.Error.Printf("requester/fakeChatter - Could not generate RSA key: %s\n", err)
		}

		return
	}

	// Send an empty identity card
	err = p2p.SendEmptyIdentityCard(&privateKey, rw)
	if err != nil {
		if debugFakeChatter {
			ownLog.Error.Printf("requester/fakeChatter - Could not send empty ID card: %s\n", err)
		}

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
		if debugFakeChatter {
			ownLog.Error.Printf("requester/fakeChatter - Could not send datum request: %s\n", err)
		}

		return
	}

	// Receive the response with the encrypted message
	// Increased wait longer in case that the encryption or file I/O take some time
	var firstMessageResponse p2p.FirstMessage
	signedMessage, err := p2p.ReceiveAndVerifySignedMessage(rw, &listenerIdentityCard.PublicKey, &firstMessageResponse, constants.MaxWaitTime*3)
	if err != nil {
		if debugFakeChatter {
			ownLog.Error.Printf("requester/fakeChatter - Could not handle received first message: %s\n", err)
		}

		return
	}

	err = firstMessageResponse.CheckForContent()
	if err != nil {
		if debugFakeChatter {
			ownLog.Error.Printf("requester/fakeChatter - First message has invalid content: %s\n", err)
		}

		return
	}

	// Extract owner's public key that will be used to verify the following messages
	ownerPublicKey := firstMessageResponse.PublicKey

	// Send acknowledgment for the encrypted data
	ack, err := createAck(signedMessage, 0)
	if err != nil {
		if debugFakeChatter {
			ownLog.Error.Printf("requester/fakeChatter - Could not create ack for fist message: %s\n", err)
		}

		return
	}

	err = p2p.CreateAndSendSignedMessage(ack, &privateKey, rw)
	if err != nil {
		if debugFakeChatter {
			ownLog.Error.Printf("requester/fakeChatter - Could not send ack for first message: %s\n", err)
		}

		return
	}

	// Store all data
	var data nP.Data
	for currentID := 1; ; currentID++ {
		// Read data
		signedMessage, err = p2p.ReceiveAndVerifySignedMessage(rw, &ownerPublicKey, &data)
		if err != nil {
			_, isTimeOutError := err.(*node.TimeOutError) //nolint:errorlint,ifshort
			if !isTimeOutError && debugFakeChatter {
				ownLog.Error.Printf("requester/fakeChatter - Could not handle fake decryption data: %s\n", err)
			}

			break
		}

		// Send an acknowledgment
		ack, err = createAck(signedMessage, currentID)
		if err != nil {
			if debugFakeChatter {
				ownLog.Error.Printf("requester/fakeChatter - Could not create ack for fake decryption data: %s\n", err)
			}

			return
		}

		err = p2p.CreateAndSendSignedMessage(ack, &privateKey, rw)
		if err != nil {
			if debugFakeChatter {
				ownLog.Error.Printf("requester/fakeChatter - Could not create ack for fake decryption data: %s\n", err)
			}

			return
		}
	}

	if atomic.AddInt32(&fakeConnectionsAmount, 1) == minFakeConnectionCount {
		if len(fakeDone) == 0 {
			fakeDone <- true
		}
	}
}
