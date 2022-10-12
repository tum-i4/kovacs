package storage

import (
	"crypto/rsa"
	"fmt"
	"time"
)

// BlockchainPayload contains all fields that will be stored in the blockchain.
type BlockchainPayload struct {
	PseudonymConsumer string          `json:"pseudonym_consumer"`
	PseudonymOwner    string          `json:"pseudonym_owner"`
	EncryptedConsumer UsageLogContent `json:"encrypted_consumer"`
	EncryptedOwner    UsageLogContent `json:"encrypted_owner"`
}

type UsageLogContent struct {
	Justification string `json:"explanation"`
	DatumRequest  string `json:"datum"`
	Timestamp     int64  `json:"timestamp"`
}

func createBlockchainPayload(justification string, datum string, ownerPublicKey *rsa.PublicKey, consumerPublicKey *rsa.PublicKey) (BlockchainPayload, error) {
	// Pseudonym creation
	pseudonymConsumer, err := GeneratePseudonym(consumerPublicKey)
	if err != nil {
		return BlockchainPayload{}, fmt.Errorf("listener/createBlockchainPayload - Could not calculate the consumer's pseudonym")
	}

	pseudonymOwner, err := GeneratePseudonym(ownerPublicKey)
	if err != nil {
		return BlockchainPayload{}, fmt.Errorf("listener/createBlockchainPayload - Could not calculate the owner's pseudonym")
	}

	logConsumer, err := createAndEncryptLogContent(justification, datum, consumerPublicKey)
	if err != nil {
		return BlockchainPayload{}, fmt.Errorf("listener/createBlockchainPayload - %w", err)
	}

	logOwner, err := createAndEncryptLogContent(justification, datum, ownerPublicKey)
	if err != nil {
		return BlockchainPayload{}, fmt.Errorf("listener/createBlockchainPayload - %w", err)
	}

	return BlockchainPayload{
		PseudonymConsumer: pseudonymConsumer,
		PseudonymOwner:    pseudonymOwner,
		EncryptedConsumer: logConsumer,
		EncryptedOwner:    logOwner,
	}, nil
}

func createAndEncryptLogContent(justification string, datum string, publicKey *rsa.PublicKey) (UsageLogContent, error) {
	encryptedJustification, err := PublicKeyEncryption(justification, publicKey)
	if err != nil {
		return UsageLogContent{}, fmt.Errorf("could not encrypt justification ('%s') because: %w", justification, err)
	}

	encryptedDatum, err := PublicKeyEncryption(datum, publicKey)
	if err != nil {
		return UsageLogContent{}, fmt.Errorf("could not encrypt datum ('%s') because: %w", datum, err)
	}

	return UsageLogContent{
		Justification: encryptedJustification,
		DatumRequest:  encryptedDatum,
		Timestamp:     time.Now().Unix(),
	}, nil
}
