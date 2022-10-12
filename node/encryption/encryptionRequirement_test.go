package encryptionRequirement

import (
	"bytes"
	"testing"

	"node/constants"
)

const (
	erRuns = 50
)

func TestGenerateEncryptionRequirement(t *testing.T) {
	for i := 0; i < erRuns; i++ {
		encryptionRequirement, err := GenerateEncryptionRequirement()
		if err != nil {
			t.Fatalf("TestGenerateEncryptionRequirement - Could not generate encryption requirement: %s\n", err)
		}

		// The assumption of passwordRequirement being valid can be made due to its testing in the passwordRequirement package
		if bytes.Equal(encryptionRequirement.nonce, []byte{}) {
			t.Errorf("TestGenerateEncryptionRequirement - nonce is empty\n")
		} else if len(encryptionRequirement.nonce) != constants.NonceLength {
			t.Errorf("TestGenerateEncryptionRequirement - nonce has an invalid size of %d, expected: %d\n", len(encryptionRequirement.nonce), constants.NonceLength)
		}
	}
}
