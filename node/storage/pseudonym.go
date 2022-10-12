package storage

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blake2s"
)

// GeneratePseudonym returns the hex representation of the generated public key.
func GeneratePseudonym(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("node/GeneratePseudonym - Could not marshal public key: %w", err)
	}

	blakeSum := blake2s.Sum256(publicKeyBytes)

	return hex.EncodeToString(blakeSum[:]), nil
}
