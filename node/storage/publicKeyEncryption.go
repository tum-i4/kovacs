package storage

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"node/constants"
)

// PublicKeyEncryption encrypts the plaintext using OAEP RSA. Returns the ciphertext as string.
func PublicKeyEncryption(plainText string, publicKey *rsa.PublicKey) (string, error) {
	cipherText, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, []byte(plainText), []byte(constants.RSAEncryptionLabel))
	if err != nil {
		return "", fmt.Errorf("node/PublicKeyEncryption - Could not encrypt ciphertext: %w", err)
	}

	return hex.EncodeToString(cipherText), nil
}

// PublicKeyDecryption takes the hex cipher text and a private key and returns the decrypted message as []byte.
func PublicKeyDecryption(cipherTextHex string, privateKey *rsa.PrivateKey) ([]byte, error) {
	cipherBytes, err := hex.DecodeString(cipherTextHex)
	if err != nil {
		return nil, fmt.Errorf("node/PublicKeyDecryption - Could not decode hex ciphertext: %w", err)
	}

	plainText, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, cipherBytes, []byte(constants.RSAEncryptionLabel))
	if err != nil {
		return nil, fmt.Errorf("node/PublicKeyDecryption - Could not decrypt the ciphertex: %w", err)
	}

	return plainText, nil
}
