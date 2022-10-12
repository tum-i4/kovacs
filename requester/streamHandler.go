package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"time"

	"node"
	"node/constants"
	log "node/logging"
	nP "node/nonRepudiation"
	"node/p2p"
	"node/storage"
)

func streamHandler(rw *bufio.ReadWriter, ssoid string, justification string, requestedDatum string, ownSignedIdentityCard p2p.SignedMessage) {
	// This slice will be used to store all signed messages
	signedMessages := make([]p2p.SignedMessage, 0)

	// Random RSA key pair that will be used to sign messages after the identity verification
	privateKey, err := nP.GenerateRSAPrivateKey()
	if err != nil {
		log.Error.Printf("requester/streamHandler - Could not generate private rsa key\n")
		return
	}

	// Parse owner's identity card
	signedIdentityCard, identityCard, isFakeChatter, err := p2p.ReceiveAndVerifySignedIdentityCard(rw, &revoloriPublicKey)
	if err != nil {
		log.Error.Printf("requester/streamHandler - Could not parse identity card: %s\n", err)
		return
	} else if isFakeChatter {
		log.Error.Printf("requester/streamHandler - Owner send ID Card marked as fake chatter?\n")
		return
	}
	signedMessages = append(signedMessages, signedIdentityCard)

	if identityCard.SSOID != ssoid {
		return
	}
	foundCorrectPeer = true
	log.Info.Printf("Found the correct SSOID (%s)!\n", ssoid)
	log.Info.Println("Starting message exchange")

	// Send my (consumer's) identity card
	err = p2p.SendSignedIdentityCard(ownSignedIdentityCard, rw)
	if err != nil {
		log.Error.Fatalf("requester/streamHandler - Could not send identity card: %s\n", err)
	}

	// Identity verification is complete
	log.Info.Println("Identity verification ended successfully")

	// Send datum request
	request := p2p.FirstMessage{
		Datum:         requestedDatum,
		Justification: justification,
		PublicKey:     privateKey.PublicKey,
		Type:          constants.MessageTypeRequester,
	}

	err = p2p.CreateAndSendSignedMessage(request, &globalPrivateKey, rw)
	if err != nil {
		log.Error.Fatalf("requester/streamHandler - Could not send signed first message: %s\n", err)
	}

	// Receive the response with the encrypted message
	// Increased wait longer in case that the encryption or file I/O take some time
	var firstMessageResponse p2p.FirstMessage
	signedMessage, err := p2p.ReceiveAndVerifySignedMessage(rw, &identityCard.PublicKey, &firstMessageResponse, constants.MaxWaitTime*3)
	if err != nil {
		log.Error.Fatalf("requester/streamHandler - Signed message error: %s\n", err)
	}

	err = firstMessageResponse.CheckForContent()
	if err != nil {
		log.Error.Fatalf("requester/streamHandler - Invalid first message: %s\n", err)
	}
	signedMessages = append(signedMessages, signedMessage)

	// Extract owner's public key that will be used to verify the following messages
	ownerPublicKey := firstMessageResponse.PublicKey

	// Send acknowledgment for the encrypted data
	ack, err := createAck(signedMessage, 0)
	if err != nil {
		log.Error.Fatalf("requester/streamHandler - Could not create first acknowledgement: %s\n", err)
	}

	err = p2p.CreateAndSendSignedMessage(ack, &privateKey, rw)
	if err != nil {
		log.Error.Fatalf("requester/streamHandler - Could not send first acknowledgement: %s\n", err)
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
			log.Error.Fatalf("requester/streamHandler - Received invalid data: %s\n", err)
		}

		// Send an acknowledgment
		ack, err = createAck(signedMessage, currentID)
		if err != nil {
			log.Error.Fatalf("%s\n", err)
		}

		err = p2p.CreateAndSendSignedMessage(ack, &privateKey, rw)
		if err != nil {
			log.Error.Fatalf("requester/streamHandler - An error occurred creating or sending the signed message: %s\n", err)
		}

		latestSignedMessage = signedMessage
	}

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
	plaintext, err := nP.DecryptMessage(&data, firstMessageResponse.Datum)
	if err != nil {
		log.Error.Fatalf("requester/streamHandler - Could not decrypt encrypted message: %s; Protocol failed!\n", err)
	}

	log.Info.Printf("Successfully completed; Message is: '%s'\n", plaintext)

	signedMessages = append(signedMessages, latestSignedMessage)
	err = storage.StoreExchange(signedMessages, &privateKey, &globalPrivateKey.PublicKey)
	if err != nil {
		log.Error.Fatalf("requester/streamHandler - Could not store data: %s\n", err)
	}

	realDone <- returnValue{
		success: true,
		value:   plaintext,
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
