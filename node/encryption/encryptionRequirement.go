package encryptionRequirement

import (
	"fmt"

	pR "node/password"
)

// GenerateEncryptionRequirement returns a filled EncryptionRequirement struct.
func GenerateEncryptionRequirement() (EncryptionRequirement, error) {
	passwordReq, err := pR.GeneratePasswordRequirement()
	if err != nil {
		return EncryptionRequirement{}, fmt.Errorf("encryptionRequirement/GenerateEncryptionRequirement - Could not generate password requirement: %w", err)
	}

	nonce, err := GenerateNonce()
	if err != nil {
		return EncryptionRequirement{}, fmt.Errorf("encryptionRequirement/GenerateEncryptionRequirement - Could not generate nonce: %w", err)
	}

	return EncryptionRequirement{
		passwordRequirement: passwordReq,
		nonce:               nonce,
	}, nil
}
