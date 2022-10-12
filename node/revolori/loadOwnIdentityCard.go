package revolori

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"node/constants"
	"node/p2p"
)

func LoadOwnIdentityCard(ownPrivateKey *rsa.PrivateKey, revoloriPublicKey *rsa.PublicKey) (string, error) {
	// Load identity card from disk
	fileContent, err := ioutil.ReadFile(constants.IdentityFilePath)
	if err != nil {
		return "", fmt.Errorf("node/LoadOwnIdentityCard - Could not read identity file: %w", err)
	}

	// Turn the []byte into a struct
	var signedCard p2p.SignedMessage
	err = json.Unmarshal(fileContent, &signedCard)
	if err != nil {
		return "", fmt.Errorf("node/LoadOwnIdentityCard - Could not unmarshal signed identity card: %w", err)
	}

	// Needed because VerifySignedIdentityCard requires a signedMessage
	signedMessage, err := p2p.CreateSignedMessage(signedCard, ownPrivateKey)
	if err != nil {
		return "", fmt.Errorf("node/LoadOwnIdentityCard - could not create SignedMessage: %w", err)
	}

	_, ownIdentityCard, isFakeChatter, err := p2p.VerifySignedIdentityCard(signedMessage, revoloriPublicKey)
	if err != nil {
		return "", fmt.Errorf("node/LoadOwnIdentityCard - could not verify: %w", err)
	}

	if isFakeChatter {
		return "", fmt.Errorf("node/LoadOwnIdentityCard - identity card is marked as fake chatter")
	}

	return ownIdentityCard.SSOID, nil
}
