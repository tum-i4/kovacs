package nonRepudiationRequirement

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/blake2s"
	"node/constants"
	eR "node/encryption"
	"node/logging"
	pR "node/password"
)

// DecryptMessage gets a filled data object and returns the decoded string or error on failure.
func DecryptMessage(decryptionData *Data, encryptedMessage string) (string, error) {
	plainPassword := decryptionData.GetPlainPassword()
	salt := decryptionData.GetSalt()
	nonce := decryptionData.GetNonce()

	if bytes.Equal(plainPassword, []byte{}) {
		return "", errors.New("nonRepudiation/DecryptMessage - Received an empty plain password")
	}

	if bytes.Equal(salt, []byte{}) {
		return "", errors.New("nonRepudiation/DecryptMessage - Received an empty salt")
	}

	if bytes.Equal(nonce, []byte{}) {
		return "", errors.New("nonRepudiation/DecryptMessage - Received an empty nonce")
	}

	hashedPassword, err := pR.GeneratePasswordFromSalt(plainPassword, salt)
	if err != nil {
		return "", err
	}

	blake := blake2s.Sum256(hashedPassword)

	decrypted, err := eR.DecryptAESGCM(blake[:], nonce, encryptedMessage)
	if err != nil {
		return "", err
	}

	return decrypted, nil
}

// GenerateNonRepudiationRequirement returns a filled NonRepudiationRequirement struct.
func GenerateNonRepudiationRequirement() (NonRepudiationRequirement, error) {
	encryptionRequirement, err := eR.GenerateEncryptionRequirement()
	if err != nil {
		return NonRepudiationRequirement{}, fmt.Errorf("nonRepudiation/GenerateNonRepudiationRequirement - %w", err)
	}

	repetitionsBig, err := rand.Int(rand.Reader, big.NewInt(constants.MaxRepetitions-constants.MinimumRepetitions))
	if err != nil {
		return NonRepudiationRequirement{}, fmt.Errorf("nonRepudiation/GenerateNonRepudiationRequirement - Could not generate repetitions: %w", err)
	}
	// Cast the big int to int64 to int
	repetitions := constants.MinimumRepetitions + int(repetitionsBig.Int64())

	falseData := make([]Data, 0, repetitions)

	for i := 0; i < repetitions; i++ {
		data, errFake := generateFakeData()
		if errFake != nil {
			log.Error.Fatalf("nonRepudiation/GenerateNonRepudiationRequirement - %v", errFake)
		}

		falseData = append(falseData, data)
	}

	if len(falseData) != repetitions {
		log.Error.Fatalf("nonRepudiation/GenerateNonRepudiationRequirement - Invalid false data length: %d, expected %d", len(falseData), repetitions)
	}

	privateKey, err := GenerateRSAPrivateKey()
	if err != nil {
		log.Error.Fatalf("nonRepudiation/GenerateNonRepudiationRequirement - Could not generate private RSA key: %v", err)
	}

	return NonRepudiationRequirement{
		privateKey:            privateKey,
		encryptionRequirement: encryptionRequirement,
		repetitions:           repetitions,
		fakeData:              falseData,
	}, nil
}

// GenerateRSAPrivateKey returns a rsa private key. It is more efficient to calculate the public key where needed.
func GenerateRSAPrivateKey() (rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, constants.RSAKeySize)
	if err != nil {
		return rsa.PrivateKey{}, err
	}

	return *privateKey, nil
}

// generateFakeData creates an instance of Data which is filled with random values.
func generateFakeData() (Data, error) {
	password, err := pR.GeneratePlainPassword()
	if err != nil {
		return Data{}, fmt.Errorf("nonRepudiation/generateFakeData - Could not generate password: %w", err)
	}

	nonce, err := eR.GenerateNonce()
	if err != nil {
		return Data{}, fmt.Errorf("nonRepudiation/generateFakeData - Could not generate nonce: %w", err)
	}

	salt, err := pR.GenerateSalt()
	if err != nil {
		return Data{}, fmt.Errorf("nonRepudiation/generateFakeData - Could not generate salt: %w", err)
	}

	return Data{
		PlainPassword: password,
		Salt:          salt,
		Nonce:         nonce,
	}, nil
}
