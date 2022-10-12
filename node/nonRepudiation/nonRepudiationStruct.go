package nonRepudiationRequirement

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"

	"node/constants"
	eR "node/encryption"
)

type NonRepudiationRequirement struct {
	privateKey            rsa.PrivateKey
	encryptionRequirement eR.EncryptionRequirement
	fakeData              []Data
	repetitions           int
}

type Data struct {
	PlainPassword []byte `json:"plain_password"`
	Salt          []byte `json:"salt"`
	Nonce         []byte `json:"nonce"`
}

// EncryptMessage takes a byte array and returns the encrypted hash in a hex representation or error on failure.
func (requirement *NonRepudiationRequirement) EncryptMessage(message []byte) (string, error) {
	hashedPassword, nonce := requirement.GetEncryptionValues()
	if bytes.Equal(hashedPassword, []byte{}) {
		return "", errors.New("nonRepudiation/EncryptMessage - Received an empty hashed password")
	}

	if bytes.Equal(nonce, []byte{}) {
		return "", errors.New("nonRepudiation/EncryptMessage - Received an empty nonce")
	}

	encrypted, err := eR.EncryptAESGCM(hashedPassword, nonce, message)
	if err != nil {
		return "", err
	}

	return encrypted, nil
}

// GetRSAKeyPair returns the RSA keypair belonging to the requirement.
func (requirement *NonRepudiationRequirement) GetRSAKeyPair() (rsa.PrivateKey, rsa.PublicKey) {
	return requirement.GetPrivateKey(), requirement.GetPublicKey()
}

// GetPublicKey returns the public key stored in the requirement.
func (requirement *NonRepudiationRequirement) GetPublicKey() rsa.PublicKey {
	return requirement.privateKey.PublicKey
}

// GetPrivateKey returns the private key stored in the requirement.
func (requirement *NonRepudiationRequirement) GetPrivateKey() rsa.PrivateKey {
	return requirement.privateKey
}

// GetEncryptionValues returns the hashed password and nonce by calling GetEncryptionValues on the encryptionRequirement.
func (requirement *NonRepudiationRequirement) GetEncryptionValues() ([]byte, []byte) {
	return requirement.encryptionRequirement.GetEncryptionValues()
}

// GetDecryptionValues returns the plain password, salt and nonce by calling GetDecryptionValues on the encryptionRequirement.
func (requirement *NonRepudiationRequirement) GetDecryptionValues() Data {
	password, salt, nonce := requirement.encryptionRequirement.GetDecryptionValues()
	return Data{
		PlainPassword: password,
		Salt:          salt,
		Nonce:         nonce,
	}
}

// GetRepetitions returns the repetitions.
func (requirement *NonRepudiationRequirement) GetRepetitions() int {
	return requirement.repetitions
}

// GetFakeData returns all the fake data array.
func (requirement *NonRepudiationRequirement) GetFakeData() []Data {
	return requirement.fakeData
}

// PopFakeData pops the first data object from requirement.fakeData.
func (requirement *NonRepudiationRequirement) PopFakeData() (Data, error) {
	if len(requirement.fakeData) == 0 {
		return Data{}, errors.New("NonRepudiationRequirement.PopFakeData: Fake data queue is empty")
	}

	data := requirement.fakeData[0]
	requirement.fakeData = requirement.fakeData[1:]
	return data, nil
}

// CheckErr checks if the field lengths are equal to the constants defined in constants/password.go.
func (data *Data) CheckErr() error {
	if len(data.GetPlainPassword()) != constants.PasswordPlainLength {
		return fmt.Errorf("Data.CheckErr - Invalid plain password length: %d. Expected: %d", len(data.GetPlainPassword()), constants.PasswordPlainLength)
	}

	if len(data.GetSalt()) != constants.SaltLength {
		return fmt.Errorf("Data.CheckErr - Invalid salt length: %d. Expected: %d", len(data.GetSalt()), constants.SaltLength)
	}

	if len(data.GetNonce()) != constants.NonceLength {
		return fmt.Errorf("Data.CheckErr - Invalid nonce length: %d. Expected: %d", len(data.GetNonce()), constants.NonceLength)
	}

	return nil
}

// String returns a JSON object.
func (data *Data) String() (string, error) {
	jsonByte, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonByte), nil
}

// GetAll returns all data inside the struct in the order of [plainPassword, salt, nonce].
func (data *Data) GetAll() ([]byte, []byte, []byte) {
	return data.GetPlainPassword(), data.GetSalt(), data.GetPlainPassword()
}

// GetPlainPassword returns the plain password from the fake data struct.
func (data *Data) GetPlainPassword() []byte {
	return data.PlainPassword
}

// GetSalt returns the Salt from the fake data struct.
func (data *Data) GetSalt() []byte {
	return data.Salt
}

// GetNonce returns the Nonce from the fake data struct.
func (data *Data) GetNonce() []byte {
	return data.Nonce
}
