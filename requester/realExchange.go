package main

import (
	"bufio"
	"node"
	"node/constants"
	log "node/logging"
	nP "node/nonRepudiation"
	"node/p2p"
	"node/storage"
	"time"
)

func realExchange(rw *bufio.ReadWriter, config *configuration, ownSignedIDCard *p2p.SignedMessage, listenerIdentityCard *p2p.IdentityCard, signedMessages []p2p.SignedMessage, idVerificationStart *time.Time, peerStart *time.Time, ret *returnValue) {
	log.Info.Printf("Found the correct SSOID (%s)!\n", config.ssoid)
	log.Info.Println("Starting message exchange")

	// Send my (consumer's) identity card
	err := p2p.SendSignedIdentityCard(*ownSignedIDCard, rw)
	if err != nil {
		cleanUpAfterFailure()
		log.Error.Printf("requester/streamHandler - Could not send identity card: %s\n", err)
		return
	}

	// Identity verification is complete
	idVerificationDuration := time.Since(*idVerificationStart)
	log.Info.Println("Identity verification ended successfully")

	// Random RSA key pair that will be used to sign messages after the identity verification
	newUsageStart := time.Now()
	privateKey, err := nP.GenerateRSAPrivateKey()
	if err != nil {
		cleanUpAfterFailure()
		log.Error.Printf("requester/streamHandler - Could not generate private rsa key\n")
		return
	}

	// Send datum request
	request := p2p.FirstMessage{
		Datum:         config.requestedDatum,
		Justification: config.justification,
		PublicKey:     privateKey.PublicKey,
		Type:          constants.MessageTypeRequester,
	}

	err = p2p.CreateAndSendSignedMessage(request, &globalPrivateKey, rw)
	if err != nil {
		cleanUpAfterFailure()
		log.Error.Printf("requester/streamHandler - Could not send signed first message: %s\n", err)
		return
	}

	// Receive the response with the encrypted message
	// Increased wait longer in case that the encryption or file I/O take some time
	var firstMessageResponse p2p.FirstMessage
	signedMessage, err := p2p.ReceiveAndVerifySignedMessage(rw, &listenerIdentityCard.PublicKey, &firstMessageResponse, constants.MaxWaitTime*3)
	if err != nil {
		cleanUpAfterFailure()
		log.Error.Printf("requester/streamHandler - Error receiving the first message: %s\n", err)
		return
	}

	err = firstMessageResponse.CheckForContent()
	if err != nil {
		cleanUpAfterFailure()
		log.Error.Printf("requester/streamHandler - Invalid first message: %s\n", err)
		return
	}
	signedMessages = append(signedMessages, signedMessage)

	// Extract owner's public key that will be used to verify the following messages
	ownerPublicKey := firstMessageResponse.PublicKey

	// Send acknowledgment for the encrypted data
	ack, err := createAck(signedMessage, 0)
	if err != nil {
		cleanUpAfterFailure()
		log.Error.Printf("requester/streamHandler - Could not create first acknowledgement: %s\n", err)
		return
	}

	err = p2p.CreateAndSendSignedMessage(ack, &privateKey, rw)
	if err != nil {
		cleanUpAfterFailure()
		log.Error.Printf("requester/streamHandler - Could not send first acknowledgement: %s\n", err)
		return
	}

	// Store all data
	var data nP.Data
	var latestSignedMessage p2p.SignedMessage

	for currentID := 1; ; currentID++ {
		// Read data
		signedMessage, err = p2p.ReceiveAndVerifySignedMessage(rw, &ownerPublicKey, &data)
		if err != nil {
			break
		}

		// Check data validity
		err = data.CheckErr()
		if err != nil {
			cleanUpAfterFailure()
			log.Error.Printf("requester/streamHandler - Received invalid data: %s\n", err)
			return
		}

		// Send an acknowledgment
		ack, err = createAck(signedMessage, currentID)
		if err != nil {
			cleanUpAfterFailure()
			log.Error.Printf("requester/streamHandler - Failed to create an acknowledgement: %s\n", err)
			return
		}

		err = p2p.CreateAndSendSignedMessage(ack, &privateKey, rw)
		if err != nil {
			cleanUpAfterFailure()
			log.Error.Printf("requester/streamHandler - Failed to send acknowledgment: %s\n", err)
			return
		}

		latestSignedMessage = signedMessage
	}

	newUsageMsgDuration := time.Since(newUsageStart)

	// Check if last error was a timeout
	_, ok := err.(*node.TimeOutError) //nolint:errorlint,ifshort
	if !ok {
		// Some other error happened
		log.Info.Printf("requester/streamHandler - An error occurred handling the received signed message: %s\n", err)
		log.Info.Printf("requester/streamHandler - Attempting to decypher anyway\n")
	} else {
		log.Info.Printf("requester/streamHandler - Experienced a time out. Trying to decrypt the message")
	}

	// Attempt to decrypt the message using the last data struct
	decryptionStart := time.Now()
	plaintext, err := nP.DecryptMessage(&data, firstMessageResponse.Datum)
	if err != nil {
		log.Error.Fatalf("requester/streamHandler - Could not decrypt encrypted message: %s; Protocol failed!\n", err)
	}
	decryptionDuration := time.Since(decryptionStart)

	log.Info.Printf("Successfully completed; Message is: '%s'\n", plaintext)

	signedMessages = append(signedMessages, latestSignedMessage)

	proofStart := time.Now()
	err = storage.StoreExchange(signedMessages, &privateKey, &globalPrivateKey.PublicKey)
	if err != nil {
		log.Error.Fatalf("requester/streamHandler - Could not store data: %s\n", err)
	}
	proofDuration := time.Since(proofStart)

	ret.value = plaintext
	ret.success = true
	ret.exchangeDuration = time.Since(*peerStart)
	ret.idVerificationDuration = idVerificationDuration
	ret.newUsageMsgDuration = newUsageMsgDuration
	ret.decryptionDuration = decryptionDuration
	ret.proofDuration = proofDuration

	realDone <- *ret
}

func cleanUpAfterFailure() {
	foundCorrectPeer = false
	if len(exchangeFailed) == 0 {
		exchangeFailed <- true
	}
}
