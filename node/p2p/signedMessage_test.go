package p2p

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"node/random"
)

const runs = 10

func TestSignAndVerify(t *testing.T) {
	for i := 0; i < runs; i++ {
		message := random.String(random.PositiveIntFromRange(16, 256))
		privateKey, err := rsa.GenerateKey(rand.Reader, 3072)
		if err != nil {
			t.Fatalf("TestSignAndVerify - Could not generate rsa key: %s\n", err)
		}

		signedMessage, err := CreateSignedMessage(struct {
			text string
		}{
			text: message,
		}, privateKey)
		if err != nil {
			t.Fatalf("TestSignAndVerify - Could not sign message: %s\n", err)
		}

		err = signedMessage.VerifySignature(&privateKey.PublicKey)
		if err != nil {
			t.Fatalf("TestSignAndVerify - Could not verify message: %s\n", err)
		}
	}
}
