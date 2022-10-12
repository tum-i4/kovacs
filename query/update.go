package main

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"node/constants"
	"node/p2p"
	"node/revolori"
	"node/storage"
)

func UpdateLog(directories []string, pseudonym string, updatedJustification string, updatedDatum string) {
	var signedMessages []p2p.SignedMessage
	var conversationPrivateKey rsa.PrivateKey
	var _ rsa.PublicKey
	var err error

	for _, directory := range directories {
		signedMessages, conversationPrivateKey, _, err = query(directory, pseudonym)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Fatalf("UpdateLog - Could not find log: %v\n", err)
	}

	revoloriPublicKey, err := revolori.GetPublicKey()
	if err != nil {
		log.Fatalf("UpdateLog - Could not get Revolori's key: %v\n", err)
	}

	firstMessage, _, err := p2p.ExtractAndVerifyMessages(signedMessages[:2], &revoloriPublicKey)
	if err != nil {
		log.Fatalf("UpdateLog - Could not extract the first message: %v\n", err)
	}

	var ownerKey rsa.PublicKey
	var consumerKey rsa.PublicKey

	if firstMessage.Type == constants.MessageTypeListener {
		// I am the requester
		fmt.Printf("I am the requester\n")
		ownerKey = firstMessage.PublicKey
		consumerKey = conversationPrivateKey.PublicKey
	} else {
		// I am the listener
		fmt.Printf("I am the listener\n")
		ownerKey = conversationPrivateKey.PublicKey
		consumerKey = firstMessage.PublicKey
	}

	err = storage.ExportToBlockchain(updatedJustification, updatedDatum, &ownerKey, &consumerKey)
	if err != nil {
		log.Fatalf("UpdateLog - Could not create updated block: %v\n", err)
	}

	err = DeleteLog(directories, pseudonym)
	if err != nil {
		log.Fatalf("UpdateLog - Could not delete old block: %v\n", err)
	}
}

func query(path string, pseudonym string) ([]p2p.SignedMessage, rsa.PrivateKey, rsa.PublicKey, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, rsa.PrivateKey{}, rsa.PublicKey{}, err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "-"+pseudonym+".json") {
			return storage.LoadExchange(path + file.Name())
		}
	}

	return nil, rsa.PrivateKey{}, rsa.PublicKey{}, errors.New("no match was found")
}
