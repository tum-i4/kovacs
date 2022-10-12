package encryptionRequirement

import pR "node/password"

type EncryptionRequirement struct {
	passwordRequirement pR.PasswordRequirement
	nonce               []byte
}

// GetNonce returns the nonce of the passed encryptionRequirement.
func (requirement *EncryptionRequirement) GetNonce() []byte {
	return requirement.nonce
}

// GetEncryptionValues returns the hashed password and the nonce which is all that is needed for encryption.
func (requirement *EncryptionRequirement) GetEncryptionValues() ([]byte, []byte) {
	return requirement.passwordRequirement.GetPasswordHashed(), requirement.GetNonce()
}

// GetDecryptionValues returns the plain password, salt and nonce which are all that is needed for decryption.
func (requirement *EncryptionRequirement) GetDecryptionValues() ([]byte, []byte, []byte) {
	return requirement.passwordRequirement.GetPasswordPlain(), requirement.passwordRequirement.GetSalt(), requirement.GetNonce()
}
