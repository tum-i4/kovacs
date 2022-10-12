package main

import (
	"crypto/rsa"
	"fmt"
	"log"

	ownLog "node/logging"
	"node/password"
	"node/revolori"
)

var revoloriPublicKey rsa.PublicKey
var globalPrivateKey rsa.PrivateKey
var cpuProf bool
var memProf bool

func main() {
	var err error
	port, printName := parseFlags()

	ownLog.Info.Println("\n\t===== Starting node =====")
	revoloriPublicKey, err = revolori.GetPublicKey()
	if err != nil {
		log.Fatalf("listener/main - Could not get Revolori's public key: %v\n", err)
	}

	// Check if identity card exists
	globalPrivateKey, err = revolori.Setup(true)
	if err != nil {
		ownLog.Error.Fatalln(err)
	}

	if printName {
		name, err := revolori.LoadOwnIdentityCard(&globalPrivateKey, &revoloriPublicKey)
		if err != nil {
			ownLog.Info.Fatalf("Could not get own name: %s", err)
		}

		fmt.Printf("I am: %s\n", name)
		ownLog.Info.Printf("I am: %s\n", name)
	}

	// Create password requirements
	passwordRequirement.Init()
	createNode(port)
}
