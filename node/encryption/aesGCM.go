package encryptionRequirement

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// GenerateNonce creates a random 12 byte long nonce.
func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return nonce, nil
}

// EncryptAESGCM receives the key, plaintext and nonce as bytes returns the encrypted text as hex string.
func EncryptAESGCM(password []byte, nonce []byte, plaintext []byte) (string, error) {
	if len(password) != 32 {
		return "", fmt.Errorf("encryptionRequirement/aesGCM: The password must be 32 byte long, got: %d", len(password))
	}

	block, err := aes.NewCipher(password)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)

	return hex.EncodeToString(ciphertext), nil
}

// DecryptAESGCM receives the key and nonce as byte array, the ciphertext as hex and returns the plaintext string.
func DecryptAESGCM(password []byte, nonce []byte, ciphertextHex string) (string, error) {
	if len(password) != 32 {
		return "", fmt.Errorf("encryptionRequirement/aesGCM: The password must be 32 byte long, got: %d", len(password))
	}

	ciphertext, _ := hex.DecodeString(ciphertextHex)

	block, err := aes.NewCipher(password)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
