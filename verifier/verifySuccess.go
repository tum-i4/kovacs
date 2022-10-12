package main

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"node/constants"
	nP "node/nonRepudiation"
	"node/p2p"
	"node/storage"
)

func verifySuccess(file string, revoloriPublicKey *rsa.PublicKey) {
	signedMessages, conversationPrivateKey, identityKey, err := storage.LoadExchange(file)
	if err != nil {
		log.Fatalf("verifySuccess - %v\n", err)
	}

	firstMessage, identityCard, err := p2p.ExtractAndVerifyMessages(signedMessages[:2], revoloriPublicKey)
	if err != nil {
		log.Fatalf("verifySuccess - Could not extract messages: %v\n", err)
	}

	/**	Since the *receiving* party stores the first message the types are switched **/
	if firstMessage.Type == constants.MessageTypeListener {
		fmt.Printf("The exchange was recoreded by the requester\n")
		fmt.Printf("The listener's SSOID is '%s'\n", identityCard.SSOID)

		_, err = verifyRequesterSuccess(signedMessages[2:], &firstMessage.PublicKey, firstMessage.Datum)
		if err != nil {
			log.Fatalf("verifySuccess - %v\n", err)
		}
	} else {
		fmt.Printf("The exchange was recoreded by the listener\n")
		fmt.Printf("The requester's SSOID is '%s'\n", identityCard.SSOID)

		_, err = verifyListenerSuccess(signedMessages[2:], &firstMessage.PublicKey, &conversationPrivateKey.PublicKey, &identityKey)
		if err != nil {
			log.Fatalf("verifySuccess - %v\n", err)
		}
	}

	fmt.Printf("Managed to decrpyt the ciphertext => Transaction ended successfully\n")
}

func verifyListenerSuccess(signedMessages []p2p.SignedMessage, signingKey *rsa.PublicKey, conversationSigningKey *rsa.PublicKey, identityKey *rsa.PublicKey) (string, error) {
	if len(signedMessages) != 2 {
		log.Fatalf("verifyListenerSuccess - Expected 2 signed message, got %d instead\n", len(signedMessages))
	}

	// Verify signatures
	for i, message := range signedMessages {
		err := message.VerifySignature(signingKey)
		if err != nil {
			return "", fmt.Errorf("verifyListenerSuccess - Could not verifySuccess the signature for message %d: %w", i, err)
		}
	}

	var firstMessage p2p.FirstMessage
	var data nP.Data

	// Get encrypted values
	firstMsgBytes, err := getAcknowledgementContent(signedMessages[0], identityKey)
	if err != nil {
		return "", fmt.Errorf("verifyListenerSuccess - Error on message 0: %w", err)
	}

	err = json.Unmarshal(firstMsgBytes, &firstMessage)
	if err != nil {
		return "", fmt.Errorf("verifyListenerSuccess - Could not unmarshal first message: %w", err)
	}
	encryptedData := firstMessage.Datum

	// Get decryption values
	dataBytes, err := getAcknowledgementContent(signedMessages[1], conversationSigningKey)
	if err != nil {
		return "", fmt.Errorf("verifyListenerSuccess - Error on message 1: %w", err)
	}

	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return "", fmt.Errorf("verifyListenerSuccess - Could not unmarshal data: %w", err)
	}

	// Check if the data can be decrypted
	decrypted, err := nP.DecryptMessage(&data, encryptedData)
	if err != nil {
		return "", fmt.Errorf("veriyListener - Could not decrypt data: %w", err)
	}

	return decrypted, nil
}

func verifyRequesterSuccess(signedMessages []p2p.SignedMessage, signingKey *rsa.PublicKey, encryptedDatum string) (string, error) {
	if len(signedMessages) != 1 {
		return "", fmt.Errorf("verifyRequesterSuccess - Expected 1 signed message, got %d", len(signedMessages))
	}
	signedMessage := signedMessages[0]

	err := signedMessage.VerifySignature(signingKey)
	if err != nil {
		return "", fmt.Errorf("verifyRequesterSuccess - Could not verifySuccess signature: %w", err)
	}

	var decryptionData nP.Data
	err = json.Unmarshal(signedMessage.Content, &decryptionData)
	if err != nil {
		return "", fmt.Errorf("verifyRequesterSuccess - Could not unmarshal decryption data: %w", err)
	}

	decrypted, err := nP.DecryptMessage(&decryptionData, encryptedDatum)
	if err != nil {
		return "", fmt.Errorf("verifyRequesterSuccess - Could not decrypt message: %w", err)
	}

	return decrypted, nil
}

func getAcknowledgementContent(message p2p.SignedMessage, signingKey *rsa.PublicKey) ([]byte, error) {
	var ack p2p.Acknowledgement
	var signedMsg p2p.SignedMessage

	err := json.Unmarshal(message.Content, &ack)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal acknowledgement: %w", err)
	}

	err = json.Unmarshal(ack.Content, &signedMsg)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal signed message: %w", err)
	}

	err = signedMsg.VerifySignature(signingKey)
	if err != nil {
		return nil, errors.New("could not verifySuccess signature")
	}

	return signedMsg.Content, nil
}
