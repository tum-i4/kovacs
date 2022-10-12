package main

import (
	"fmt"
	"log"

	"node/revolori"
)

func main() {
	config := ParseFlags() //nolint:ifshort

	fmt.Printf("Loading Revolori's public key...\n\n")
	revoloriPublicKey, err := revolori.GetPublicKey()
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	if config.success != "" {
		verifySuccess(config.success, &revoloriPublicKey)
	} else {
		solveDispute(config.files, &revoloriPublicKey)
	}
}
