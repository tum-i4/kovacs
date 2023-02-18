package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"time"

	ownLog "node/logging"
	nP "node/nonRepudiation"
	"node/p2p"
	"node/storage"
)

func writeLog(idCardPrivateKey *rsa.PrivateKey, i int) {
	var ret bool
	channelDone := make(chan bool, 1)
	privateKey, err := nP.GenerateRSAPrivateKey()
	if err != nil {
		log.Fatalf("could not generate RSA key: %s\n", err)
	}

	datum := fmt.Sprintf("Datum %d", i)
	justification := fmt.Sprintf("Justification for %s", datum)

	// Blockchain export (run as go routine to catch if it takes too long)
	go blockchainExport(justification, datum, &idCardPrivateKey.PublicKey, &privateKey.PublicKey, channelDone)
	select {
	case ret = <-channelDone:
		if !ret {
			log.Fatalf("Blockchain export returned false\n")
		}
	case <-time.After(15 * time.Second):
		log.Fatalf("Blockchain timed out!\n")
	}

	// SQLite export
	err = storage.ExportToSQLite(justification, datum, &idCardPrivateKey.PublicKey, &privateKey.PublicKey)
	if err != nil {
		log.Fatalf("could not export data to sqlite: %s\n", err)
	}

	// Write "decryption" data
	messages := []p2p.SignedMessage{
		p2p.SignedMessage{},
		p2p.SignedMessage{},
		p2p.SignedMessage{},
	}
	storage.StoreExchange(messages, &privateKey, &idCardPrivateKey.PublicKey)
}

func blockchainExport(justification string, datum string, idCardPublicKey *rsa.PublicKey, publicKey *rsa.PublicKey, done chan bool) {
	durations, err := storage.ExportToBlockchain(justification, datum, idCardPublicKey, publicKey)
	if err != nil {
		done <- false
		log.Fatalf("could not export data to blockchain: %s\n", err)
	}

	ownLog.Info.Printf("Blockchain export summary:\n" +
		fmt.Sprintf("\tTotal duration: %v", durations.TotalDuration) +
		fmt.Sprintf("\tDuration of account creation: %v\n", durations.AccountCreation) +
		fmt.Sprintf("\tDuration of account unlock: %v\n", durations.AccountUnlock) +
		fmt.Sprintf("\tDuration of mining: %v\n", durations.Mining) +
		fmt.Sprintf("\tDuration of transaction: %v\n", durations.TransactionDuration),
	)

	done <- true
}
