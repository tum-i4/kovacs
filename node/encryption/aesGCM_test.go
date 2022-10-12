package encryptionRequirement

import (
	"strings"
	"testing"

	"node/random"
)

var runs = 20

// Checks the determinism
func TestEncryptAESGCMEquality(t *testing.T) {
	for i := 0; i < runs; i++ {
		password := []byte(random.String(16))
		plaintext := []byte(random.String(random.PositiveIntFromRange(1, 256)))
		nonce, err := GenerateNonce()
		if err != nil {
			t.Errorf("TestEncryptAESGCMEquality - Could not generate nonce: %s\n", err)
		}

		hash1, err := EncryptAESGCM(password, nonce, plaintext)
		if err != nil {
			t.Errorf("TestEncryptAESGCMEquality - Could not encrypt: %s\n", err)
		}

		hash2, err := EncryptAESGCM(password, nonce, plaintext)
		if err != nil {
			t.Errorf("TestEncryptAESGCMEquality - Could not encrypt: %s\n", err)
		}

		if hash1 != hash2 {
			t.Errorf("TestEncryptAESGCMEquality - Encrypted hashes do not match\nHash1: %s\nHash2: %s\n", hash1, hash2)
		}
	}
}

// Check if the encrypted text can be decrypted
func TestEncryptionDecryption(t *testing.T) {
	for i := 0; i < runs; i++ {
		password := []byte(random.String(16))
		plaintext := []byte(random.String(random.PositiveIntFromRange(1, 256)))
		nonce, err := GenerateNonce()
		if err != nil {
			t.Errorf("TestEncryptDecrypt - Could not generate nonce: %s\n", err)
		}

		ciphertext, err := EncryptAESGCM(password, nonce, plaintext)
		if err != nil {
			t.Errorf("TestEncryptDecrypt - Could not encrypt: %s\n", err)
		}

		decrypted, err := DecryptAESGCM(password, nonce, ciphertext)
		if err != nil {
			t.Errorf("TestEncryptDecrypt - Could not decrypt: %s\n", err)
		}

		if string(plaintext) != decrypted {
			t.Errorf("TestEncryptDecrypt - Decrypted text does not match plaintext\nPlaintext: %s\nDecrypted: %s\n", plaintext, decrypted)
		}
	}
}

// Check if an encrypted text can be decrypted with invalid password or nonce
func TestEncryptionDecryptionFailure(t *testing.T) {
	for i := 0; i < runs; i++ {
		password := []byte(random.String(16))
		plaintext := []byte(random.String(random.PositiveIntFromRange(1, 256)))

		nonceCorrect, err := GenerateNonce()
		if err != nil {
			t.Errorf("TestEncryptionDecryptionFailure - Could not generate nonce: %s\n", err)
		}
		nonceFalse, err := GenerateNonce()
		if err != nil {
			t.Errorf("TestEncryptionDecryptionFailure - Could not generate nonce: %s\n", err)
		}

		cipher, err := EncryptAESGCM(password, nonceCorrect, plaintext)
		if err != nil {
			t.Errorf("TestEncryptAESGCMEquality - Could not encrypt: %s\n", err)
		}

		// Use wrong password
		decrypted, err := DecryptAESGCM([]byte(random.String(0)), nonceCorrect, cipher)
		if err == nil {
			t.Errorf("TestEncryptionDecryptionFailure - Managed to decrypt with wrong password: %s\n", decrypted)
		}

		// Use wrong nonce
		decrypted, err = DecryptAESGCM(password, nonceFalse, cipher)
		if err == nil {
			t.Errorf("TestEncryptionDecryptionFailure - Managed to decrypt with wrong nonce: %s\n", decrypted)
		}

		// Use wrong nonce and password
		decrypted, err = DecryptAESGCM([]byte(random.String(16)), nonceFalse, cipher)
		if err == nil {
			t.Errorf("TestEncryptionDecryptionFailure - Managed to decrypt with wrong nonce: %s\n", decrypted)
		}

		// Change cipher
		cipher = strings.Replace(cipher, string(cipher[0]), "aa", 2)
		decrypted, err = DecryptAESGCM(password, nonceCorrect, cipher)
		if err == nil {
			t.Errorf("TestEncryptionDecryptionFailure - Managed to decrypt with wrong nonce: %s\n", decrypted)
		}
	}
}

func TestEncryptInvalidPasswordLength(t *testing.T) {
	for i := 0; i < runs; i++ {
		// Password is too short
		password := []byte(random.String(random.PositiveIntFromRange(0, 15)))
		plaintext := []byte(random.String(random.PositiveIntFromRange(1, 256)))
		nonce, err := GenerateNonce()
		if err != nil {
			t.Errorf("TestEncryptInvalidPassword - Could not generate nonce: %s\n", err)
		}

		_, err = EncryptAESGCM(password, nonce, plaintext)
		if err == nil {
			t.Errorf("TestEncryptAESGCMEquality - Encrypted with invalid password: %s\n", string(password))
		}

		// Password is too long
		password = []byte(random.String(random.PositiveIntFromRange(17, 64)))
		_, err = EncryptAESGCM(password, nonce, plaintext)
		if err == nil {
			t.Errorf("TestEncryptAESGCMEquality - Encrypted with invalid password: %s\n", string(password))
		}
	}
}

func TestDecryptInvalidPasswordLength(t *testing.T) {
	for i := 0; i < runs; i++ {
		// Password is too short
		password := []byte(random.String(random.PositiveIntFromRange(0, 15)))
		plaintext := random.String(random.PositiveIntFromRange(1, 256))
		nonce, err := GenerateNonce()
		if err != nil {
			t.Errorf("TestEncryptInvalidPassword - Could not generate nonce: %s\n", err)
		}

		_, err = DecryptAESGCM(password, nonce, plaintext)
		if err == nil {
			t.Errorf("TestEncryptAESGCMEquality - Encrypted with invalid password: %s\n", string(password))
		}

		// Password is too long
		password = []byte(random.String(random.PositiveIntFromRange(17, 64)))
		_, err = DecryptAESGCM(password, nonce, plaintext)
		if err == nil {
			t.Errorf("TestEncryptAESGCMEquality - Encrypted with invalid password: %s\n", string(password))
		}
	}
}
