package nonRepudiationRequirement

import (
	"bytes"
	"testing"

	"node/constants"
	"node/random"
)

const nrRuns = 50

func TestGenerateRSAPrivateKey(t *testing.T) {
	for i := 0; i < nrRuns; i++ {
		_, err := GenerateRSAPrivateKey()
		if err != nil {
			t.Fatalf("TestGenerateRSAPrivateKey - Could not generate private rsa key: %s\n", err)
		}
	}
}

func TestEnDecryption(t *testing.T) {
	for i := 0; i < nrRuns; i++ {
		nRR, err := GenerateNonRepudiationRequirement()
		if err != nil {
			t.Fatalf("TestEnDecryption - Could not generate non repudiation requirement: %s\n", err)
		}

		message := random.String(random.PositiveIntFromRange(1, 1028))

		encrypted, err := nRR.EncryptMessage([]byte(message))
		if err != nil {
			t.Errorf("TestEnDecryption - Could not encrypt message: %s\n", err)
		}

		decryptDatum := nRR.GetDecryptionValues()

		decryptedMessage, err := DecryptMessage(&decryptDatum, encrypted)
		if err != nil {
			t.Errorf("TestEnDecryption - Could not decrypt message: %s\n", err)
		}

		if decryptedMessage != message {
			t.Errorf("TestEnDecryption - Messages do not match\nOriginal: %s\nDecrypted: %s\n", message, decryptedMessage)
		}
	}
}

func TestGenerateNonRepudiationRequirement(t *testing.T) {
	for i := 0; i < nrRuns*10; i++ {
		nRR, err := GenerateNonRepudiationRequirement()
		if err != nil {
			t.Fatalf("TestGenerateNonRepudiationRequirement - Could not generate non repudiation requirement: %s\n", err)
		}

		// Assumption that encryptionRequirement is valid due to its testing

		if nRR.repetitions < constants.MinimumRepetitions || nRR.repetitions > constants.MaxRepetitions {
			t.Errorf("TestGenerateNonRepudiationRequirement - Invalid repetition count: %d, expected a value in [%d, %d]\n", nRR.repetitions, constants.MinimumRepetitions, constants.MaxRepetitions)
		}

		if len(nRR.fakeData) != nRR.repetitions {
			t.Errorf("TestGenerateNonRepudiationRequirement - Invalid amount of fake data: %d, expected %d\n", len(nRR.fakeData), nRR.repetitions)
		}

		for _, data := range nRR.fakeData {
			if len(data.GetPlainPassword()) != constants.PasswordPlainLength {
				t.Errorf("TestGenerateNonRepudiationRequirement - Invalid plain password length: %d, expected: %d\n", len(data.GetPlainPassword()), constants.PasswordPlainLength)
			}

			if len(data.GetSalt()) != constants.SaltLength {
				t.Errorf("TestGenerateNonRepudiationRequirement - Invalid salt length: %d, expected: %d\n", len(data.GetSalt()), constants.SaltLength)
			}

			if len(data.GetNonce()) != constants.NonceLength {
				t.Errorf("TestGenerateNonRepudiationRequirement - Invalid nonce length: %d, expected: %d\n", len(data.GetNonce()), constants.NonceLength)
			}
		}
	}
}

func TestGenerateFakeData(t *testing.T) {
	for i := 0; i < nrRuns; i++ {
		falseData, err := generateFakeData()
		if err != nil {
			t.Fatalf("TestGenerateFakeData - Could not generate fake data: %s\n", err)
		}

		if bytes.Equal(falseData.GetPlainPassword(), []byte{}) {
			t.Errorf("TestGenerateFakeData - Got empty plain password\n")
		} else if len(falseData.GetPlainPassword()) != constants.PasswordPlainLength {
			t.Errorf("TestGenerateFakeData - Invalid plain password length: %d, expected %d\n", len(falseData.GetPlainPassword()), constants.PasswordPlainLength)
		}

		if bytes.Equal(falseData.GetSalt(), []byte{}) {
			t.Errorf("TestGenerateFakeData - Got empty Salt\n")
		} else if len(falseData.GetSalt()) != constants.SaltLength {
			t.Errorf("TestGenerateFakeData - Invalid salt length: %d, expected %d\n", len(falseData.GetSalt()), constants.SaltLength)
		}

		if bytes.Equal(falseData.GetNonce(), []byte{}) {
			t.Errorf("TestGenerateFakeData - Got empty Nonce\n")
		} else if len(falseData.GetNonce()) != constants.NonceLength {
			t.Errorf("TestGenerateFakeData - Invalid nonce length: %d, expected %d\n", len(falseData.GetNonce()), constants.NonceLength)
		}
	}
}
