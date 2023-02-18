package encryptionRequirement

import (
	pR "node/password"
)

// FakeChatterEncryptionRequirement returns a hard coded EncryptionRequirement struct.
// Should only be used for fake chatter!
func FakeChatterEncryptionRequirement() EncryptionRequirement {
	return EncryptionRequirement{
		passwordRequirement: pR.FakeChatterPasswordRequirement(),
		nonce:               []byte("Fake_Chatter"),
	}
}
