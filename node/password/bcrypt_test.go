package passwordRequirement

import (
	"bytes"
	"encoding/hex"
	"testing"

	"node/random"
)

const (
	bCryptRuns = 10
)

func TestGenerateFromPasswordReturnSaltEquality(t *testing.T) {
	for i := 0; i < bCryptRuns; i++ {
		password := []byte(random.String(random.PositiveIntFromRange(6, 16)))
		hash1, salt, err := GeneratePasswordReturnSalt(password)
		if err != nil {
			t.Errorf("TestGenerateFromPasswordReturnSaltEquality - Could not generate password hash: %s\n", err)
		}

		if bytes.Equal(hash1, []byte{}) {
			t.Errorf("TestGenerateFromPasswordReturnSaltEquality - Empty hash\n")
		}

		if bytes.Equal(salt, []byte{}) {
			t.Errorf("TestGenerateFromPasswordReturnSaltEquality - Empty salt\n")
		}

		hash2, err := GeneratePasswordFromSalt(password, salt)
		if err != nil {
			t.Errorf("TestGenerateFromPasswordReturnSaltEquality - Could not generate the hash with the generated salt: %s\n", string(salt))
		}

		if !bytes.Equal(hash1, hash2) {
			t.Errorf("TestGenerateFromPasswordReturnSaltEquality - Hashes do not match\nHash1: %s\nHash2: %s\n", string(hash1), string(hash2))
		}
	}
}

func TestGenerateFromPasswordReturnSalt(t *testing.T) {
	for i := 0; i < bCryptRuns; i++ {
		password := []byte(random.String(random.PositiveIntFromRange(6, 16)))
		hash, salt, err := GeneratePasswordReturnSalt(password)
		if err != nil {
			t.Errorf("TestGenerateFromPasswordReturnSalt - Could not generate password hash: %s\n", err)
		}

		if bytes.Equal(hash, []byte{}) {
			t.Errorf("TestGenerateFromPasswordReturnSalt - Returned empty hash\n")
		}

		if bytes.Equal(salt, []byte{}) {
			t.Errorf("TestGenerateFromPasswordReturnSalt - Returned empty salt\n")
		}
	}
}

func TestGenerateFromPasswordReturnSaltCheckSalt(t *testing.T) {
	for i := 0; i < bCryptRuns; i++ {
		password := []byte(random.String(random.PositiveIntFromRange(6, 16)))
		hash1, salt1, err := GeneratePasswordReturnSalt(password)
		if err != nil {
			t.Errorf("TestGenerateFromPasswordReturnSaltCheckSalt - Could not generate password hash: %s\n", err)
		}

		hash2, salt2, err := GeneratePasswordReturnSalt(password)
		if err != nil {
			t.Errorf("TestGenerateFromPasswordReturnSaltCheckSalt - Could not generate password hash: %s\n", err)
		}

		if bytes.Equal(hash1, hash2) {
			t.Errorf("TestGenerateFromPasswordReturnSaltCheckSalt - Generated the same hash\nHash1: %s\nHash2: %s\n", hex.EncodeToString(hash1), hex.EncodeToString(hash2))
		}

		if bytes.Equal(salt1, salt2) {
			t.Errorf("TestGenerateFromPasswordReturnSaltCheckSalt - Generated the same salt\nSalt1: %s\nSalt2: %s\n", hex.EncodeToString(salt1), hex.EncodeToString(salt2))
		}
	}
}

func TestGenerateFromPasswordAndSaltEmptySalt(t *testing.T) {
	for i := 0; i < bCryptRuns; i++ {
		password := []byte(random.String(random.PositiveIntFromRange(6, 16)))
		_, err := GeneratePasswordFromSalt(password, []byte{})
		if err == nil {
			t.Errorf("TestGenerateFromPasswordAndSaltEmptySalt - Could generate hash with an empty salt\n")
		}
	}
}

func TestMinSaltLength(t *testing.T) {
	for i := 1; i < maxSaltSize; i++ {

		for k := 0; k < bCryptRuns; k++ {
			password := []byte(random.String(random.PositiveIntFromRange(6, 16)))
			salt := base64Encode([]byte(random.String(i)))
			if len(salt) > maxSaltSize {
				salt = salt[:maxSaltSize-1]
			}

			_, err := GeneratePasswordFromSalt(password, salt)
			if err != nil {
				if len(salt) != 8 {
					break
				}
			}
		}
	}
}

func TestGenerateFromPasswordAndSaltEquality(t *testing.T) {
	for i := 0; i < bCryptRuns; i++ {
		password := []byte(random.String(random.PositiveIntFromRange(6, 16)))
		var salt []byte

		for len(salt) == 0 || len(salt) == 8 || len(salt) == 16 {
			salt = base64Encode([]byte(random.String(16)))
			if len(salt) > maxSaltSize {
				salt = salt[:maxSaltSize-1]
			}
		}

		hash1, err := GeneratePasswordFromSalt(password, salt)
		if err != nil {
			t.Errorf("TestGenerateFromPasswordAndSaltEquality - Could not generate password hash: %s\n", err)
		}

		if bytes.Equal(hash1, []byte{}) {
			t.Errorf("TestGenerateFromPasswordAndSaltEquality - Returned empty hash\n")
		}

		hash1Str := hex.EncodeToString(hash1)

		hash2, err := GeneratePasswordFromSalt(password, salt)
		if err != nil {
			t.Errorf("TestGenerateFromPasswordAndSaltEquality - Could not generate password hash: %s\n", err)
		}

		if bytes.Equal(hash2, []byte{}) {
			t.Errorf("TestGenerateFromPasswordAndSaltEquality - Returned empty hash\n")
		}

		hash2Str := hex.EncodeToString(hash2)

		if hash1Str != hash2Str {
			t.Errorf("TestGenerateFromPasswordAndSaltEquality - Generated hashes do not match\nHash1: %s\nHash2: %s\n", hash1Str, hash2Str)
		}
	}
}
