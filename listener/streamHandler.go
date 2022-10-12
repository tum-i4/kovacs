package main

import (
	"bufio"
	"fmt"
	"time"

	"github.com/pkg/profile"
	"node/constants"
	"node/logging"
	nP "node/nonRepudiation"
	"node/p2p"
	"node/random"
	"node/storage"
)

func streamHandler(rw *bufio.ReadWriter, connectionID int64) {
	if cpuProf {
		profilePath := fmt.Sprintf("cpu-%d", connectionID)
		connectionID++

		defer profile.Start(profile.CPUProfile, profile.Quiet, profile.ProfilePath(profilePath)).Stop()
	} else if memProf {
		profilePath := fmt.Sprintf("mem-%d", connectionID)
		connectionID++

		defer profile.Start(profile.MemProfile, profile.Quiet, profile.ProfilePath(profilePath)).Stop()
	}

	start := time.Now()
	lastTimeStamp := start.Unix()
	// Used to store all signed messages
	signedMessages := make([]p2p.SignedMessage, 0)

	// Get a non repudiation requirement needed for the repetition amount, RSA keys and encryption
	requirement, err := nP.GenerateNonRepudiationRequirement()
	if err != nil {
		log.Error.Printf("(%d) listener/streamHandler - Could not generate requirement: %v\n", connectionID, err)
		return
	}

	privateKey := requirement.GetPrivateKey()

	// Send own identity card
	err = p2p.SendSignedIdentityCard(ownSignedIdentityCard, rw)
	if err != nil {
		log.Error.Printf("(%d) listener/streamHandler - Could not send identity card: %v\n", connectionCount, err)
		return
	}

	// Parse consumer's identity card
	signedIdentityCard, identityCard, isFakeChatter, err := p2p.ReceiveAndVerifySignedIdentityCard(rw, &revoloriPublicKey)
	if err != nil {
		log.Error.Printf("(%d) listener/streamHandler - Could not parse identity card: %v\n", connectionID, err)
		return
	}
	signedMessages = append(signedMessages, signedIdentityCard)
	// Identity verification is complete

	// Receive first message with included datum request
	var firstMessageRequest p2p.FirstMessage
	signedMessage, err := p2p.ReceiveAndVerifyFirstMessage(rw, &identityCard.PublicKey, &firstMessageRequest, isFakeChatter)
	if err != nil {
		log.Error.Printf("(%d) listener/streamHandler - Could not parse signed first request: %v\n", connectionID, err)
		return
	}
	signedMessages = append(signedMessages, signedMessage)

	// Check if underlying request is valid
	err = firstMessageRequest.CheckForContentAndJustification()
	if err != nil {
		log.Error.Printf("(%d) listener/streamHandler - Invalid first message: %v\n", connectionID, err)
		return
	}

	if isFakeChatter && firstMessageRequest.Type != constants.MessageTypeFakeChatter {
		log.Error.Printf("(%d) listener/streamHandler - Identity card is marked as fake chatter (%d), but first message is not (%d)\n", connectionID, constants.MessageTypeFakeChatter, firstMessageRequest.Type)
		return
	}

	// Extract owner's public key that will be used to sign the following messages
	consumerPublicKey := firstMessageRequest.PublicKey

	var requestedDatum string
	if isFakeChatter {
		requestedDatum = random.String(random.PositiveIntFromRange(64, 512))
	} else {
		requestedDatum = "Requested datum: " + firstMessageRequest.Datum
	}

	messageCipher, err := requirement.EncryptMessage([]byte(requestedDatum))
	if err != nil {
		log.Error.Printf("(%d) listener/streamHandler - Could not encrypt message: %v\n", connectionID, err)
		return
	}

	response := p2p.FirstMessage{
		Datum:     messageCipher,
		PublicKey: privateKey.PublicKey,
		Type:      constants.MessageTypeListener,
	}

	// Create and send signed response
	responseBytes, err := p2p.CreateSendAndReturnSignedMessage(response, &globalPrivateKey, rw)
	if err != nil {
		log.Error.Printf("(%d) listener/streamHandler - Could not send signed first message: %v\n", connectionID, err)
		return
	}

	var ack p2p.Acknowledgement

	signedMessage, err = p2p.ReceiveAndVerifySignedMessage(rw, &consumerPublicKey, &ack)
	if err != nil {
		log.Error.Printf("(%d) listener/streamHandler - An error occurred when handling the received signed message: %v\n", connectionID, err)
		return
	}

	err = ack.CheckErr(0, lastTimeStamp, responseBytes)
	if err != nil {
		log.Error.Printf("(%d) listener/streamHandler - Invalid acknowledgement: %v\n", connectionID, err)
		return
	}

	signedMessages = append(signedMessages, signedMessage)
	lastTimeStamp = ack.TimeStamp

	var data nP.Data
	var msg []byte
	currentID := 1
	storeAck := false

	for i := 0; i < requirement.GetRepetitions()+1; i++ {
		if i < requirement.GetRepetitions() {
			// Get a fake datum
			data, err = requirement.PopFakeData()
			if err != nil {
				log.Error.Printf("(%d) listener/streamHandler - Could not pop fake data: %v\n", connectionID, err)
				return
			}
		} else {
			// Get the real decryption values
			data = requirement.GetDecryptionValues()
			storeAck = true
		}

		msg, err = p2p.CreateSendAndReturnSignedMessage(data, &privateKey, rw)
		if err != nil {
			log.Error.Printf("(%d) listener/streamHandler - An error occurred when sending the signed message: %v\n", connectionID, err)
			return
		}

		signedMessage, err = p2p.ReceiveAndVerifySignedMessage(rw, &consumerPublicKey, &ack)
		if err != nil {
			log.Error.Printf("(%d) listener/streamHandler - An error occurred when handling the received signed message: %v\n", connectionID, err)
			return
		}

		// Check acknowledgment validity
		err = ack.CheckErr(currentID, lastTimeStamp, msg)
		if err != nil {
			log.Error.Printf("(%d) listener/streamHandler - Invalid acknowledgement: %v\n", connectionID, err)
			return
		}

		lastTimeStamp = ack.TimeStamp

		if storeAck {
			signedMessages = append(signedMessages, signedMessage)
			storeAck = false
		}

		currentID++
	}

	if !isFakeChatter {
		log.Info.Printf("(%d) Exchange ended successfully\n", connectionID)

		err = storage.StoreExchange(signedMessages, &privateKey, &globalPrivateKey.PublicKey)
		if err != nil {
			log.Error.Printf("(%d) listener/streamHandler - Could not store data: %v\n", connectionID, err)
			return
		}

		log.Info.Printf("(%d) Storing exchange in SQLite\n", connectionID)
		exportStart := time.Now()
		err = storage.ExportToSQLite(firstMessageRequest.Justification, firstMessageRequest.Datum, &privateKey.PublicKey, &consumerPublicKey)
		if err != nil {
			log.Error.Printf("(%d) listener/streamHandler - Could not export data to sqlite: %v\n", connectionID, err)
			return
		}
		SQLiteExportDuration := time.Since(exportStart)

		log.Info.Printf("(%d) Storing exchange in blockchain\n", connectionID)
		exportStart = time.Now()
		err = storage.ExportToBlockchain(firstMessageRequest.Justification, firstMessageRequest.Datum, &privateKey.PublicKey, &consumerPublicKey)
		if err != nil {
			log.Error.Printf("(%d) listener/streamHandler - Could not export data to blockchain: %v\n", connectionID, err)
			return
		}
		blockchainExportDuration := time.Since(exportStart)

		log.Info.Printf("(%d) Exchange export successful.\n\t"+
			"Duration of entire exchange: %v\n\t"+
			"Duration of blockchain export: %v\n\t"+
			"Duration of sqlite export: %v\n",
			connectionID, time.Since(start), blockchainExportDuration, SQLiteExportDuration,
		)
	}
}
