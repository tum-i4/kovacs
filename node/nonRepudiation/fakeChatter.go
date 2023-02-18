package nonRepudiationRequirement

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"node/constants"
	eR "node/encryption"
	"node/logging"
)

// FakeChatterNonRepudiationRequirement returns a NonRepudiationRequirement struct with a hard coded encryption requirement.
// This function should only be used for fake chatter, otherwise the data consumer will be able to decrypt the data
// owner's private data without completing the non-repudiation protocol.
// For fake chatter, the exchanged datum is a randomly generated string, thus there is no information to be gained by
// a cheating data consumer. The connection between the data consumer and the data owner is encrypted, thus an
// eavesdropping attacker would not be able to notice that the encryption requirement is hard coded.
func FakeChatterNonRepudiationRequirement() (NonRepudiationRequirement, error) {
	encryptionRequirement := eR.FakeChatterEncryptionRequirement()

	repetitionsBig, err := rand.Int(rand.Reader, big.NewInt(constants.MaxRepetitions-constants.MinimumRepetitions))
	if err != nil {
		return NonRepudiationRequirement{}, fmt.Errorf("nonRepudiation/FakeChatterNonRepudiationRequirement - Could not generate repetitions: %w", err)
	}
	// Cast the big int to int64 to int
	repetitions := constants.MinimumRepetitions + int(repetitionsBig.Int64())

	falseData := make([]Data, 0, repetitions)

	for i := 0; i < repetitions; i++ {
		data, errFake := generateFakeData()
		if errFake != nil {
			log.Error.Fatalf("nonRepudiation/FakeChatterNonRepudiationRequirement - %v", errFake)
		}

		falseData = append(falseData, data)
	}

	if len(falseData) != repetitions {
		log.Error.Fatalf("nonRepudiation/FakeChatterNonRepudiationRequirement - Invalid false data length: %d, expected %d", len(falseData), repetitions)
	}

	privateKey, err := GenerateRSAPrivateKey()
	if err != nil {
		log.Error.Fatalf("nonRepudiation/FakeChatterNonRepudiationRequirement - Could not generate private RSA key: %v", err)
	}

	return NonRepudiationRequirement{
		privateKey:            privateKey,
		encryptionRequirement: encryptionRequirement,
		repetitions:           repetitions,
		fakeData:              falseData,
	}, nil
}
