package p2p

import (
	"bufio"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"node/constants"
)

type IdentityCard struct {
	SSOID     string        `json:"ssoid"`
	PublicKey rsa.PublicKey `json:"public_key"`
}

// LoadSignedIdentityCard loads an identity card from storage and signs it with the passed private key. The resulting
// signed message is then returned.
func LoadSignedIdentityCard(privateKey *rsa.PrivateKey) (SignedMessage, error) {
	fileContent, err := ioutil.ReadFile(constants.IdentityFilePath)
	if err != nil {
		return SignedMessage{}, fmt.Errorf("node/LoadSignedIdentityCard - Could not read identity file: %w", err)
	}

	var tmp SignedMessage
	err = json.Unmarshal(fileContent, &tmp)
	if err != nil {
		return SignedMessage{}, fmt.Errorf("node/LoadSignedIdentityCard - Could not unmarshal signed identity card: %w", err)
	}

	signedCard := ExtendedSignedMessage{
		Content:   tmp.Content,
		Signature: tmp.Signature,
		Type:      constants.MessageTypeRealExchange,
	}

	msg, err := CreateSignedMessage(signedCard, privateKey)
	if err != nil {
		return SignedMessage{}, fmt.Errorf("node/LoadSignedIdentityCard - Could not send signed identity card: %w", err)
	}

	return msg, nil
}

// SendSignedIdentityCard receives a pre-signed identity card and writes it to the passed ReadWriter.
func SendSignedIdentityCard(signedCard SignedMessage, rw *bufio.ReadWriter) error {
	err := SendSignedMessage(signedCard, rw)
	if err != nil {
		return fmt.Errorf("node/SendSignedIdentityCard - Could not send signed identity card: %w", err)
	}

	return nil
}

// SendEmptyIdentityCard creates and sends an empty SignedMessage to inform the other party that the current exchange is
// not a real exchange but fake chatter.
func SendEmptyIdentityCard(privateKey *rsa.PrivateKey, rw *bufio.ReadWriter) error {
	signedCard := ExtendedSignedMessage{
		Content:   nil,
		Signature: nil,
		Type:      constants.MessageTypeFakeChatter,
	}

	err := CreateAndSendSignedMessage(signedCard, privateKey, rw)
	if err != nil {
		return fmt.Errorf("node/SendEmptyIdentityCard - Could not send signed identity card: %w", err)
	}

	return nil
}

// ReceiveAndVerifySignedIdentityCard reads the IdentityCard from the ReadWriter and verifies the public keys. For it to
// work the IdentityCard needs to be signed with the private key provided by Revolori.
// Returns the signed message, unmarshaled identity card and if this is a fake exchange.
func ReceiveAndVerifySignedIdentityCard(rw *bufio.ReadWriter, revoloriPublicKey *rsa.PublicKey) (SignedMessage, IdentityCard, bool, error) {
	// The expected message is a signed message (from peer) of a signed message (from Revolori) of the identity card
	var peerSignedMessage SignedMessage

	// Read the struct
	received, err := ReadLine(rw, '}', 10*time.Second)
	if err != nil {
		return SignedMessage{}, IdentityCard{}, false, fmt.Errorf("node/ReceiveAndVerifySignedIdentityCard - Could not read the json: %w", err)
	}

	// Parse the SignedMessage from the peer
	err = json.Unmarshal([]byte(received), &peerSignedMessage)
	if err != nil {
		return SignedMessage{}, IdentityCard{}, false, fmt.Errorf("node/ReceiveAndVerifySignedIdentityCard - Could not unmarshal the peer signed message: %w", err)
	}

	return VerifySignedIdentityCard(peerSignedMessage, revoloriPublicKey)
}

func VerifySignedIdentityCard(peerSignedMessage SignedMessage, revoloriPublicKey *rsa.PublicKey) (SignedMessage, IdentityCard, bool, error) {
	var extendedSignedMessage ExtendedSignedMessage
	var peerIdentityCard IdentityCard

	// Parse the SignedMessage from Revolori
	err := json.Unmarshal(peerSignedMessage.Content, &extendedSignedMessage)
	if err != nil {
		return SignedMessage{}, IdentityCard{}, false, fmt.Errorf("node/ReceiveAndVerifySignedIdentityCard - Could not unmarshal the extended signed message: %w", err)
	}

	if extendedSignedMessage.Type == constants.MessageTypeFakeChatter {
		return SignedMessage{}, IdentityCard{}, true, err
	}

	revoloriSignedMessage := SignedMessage{
		Content:   extendedSignedMessage.Content,
		Signature: extendedSignedMessage.Signature,
	}

	// Verify Revolori's signature
	err = revoloriSignedMessage.VerifySignature(revoloriPublicKey)
	if err != nil {
		return SignedMessage{}, IdentityCard{}, false, fmt.Errorf("node/ReceiveAndVerifySignedIdentityCard - Could not verify revolori's signature: %w", err)
	}

	// Parse the IdentityCard
	err = json.Unmarshal(revoloriSignedMessage.Content, &peerIdentityCard)
	if err != nil {
		return SignedMessage{}, IdentityCard{}, false, fmt.Errorf("node/ReceiveAndVerifySignedIdentityCard - Could not unmarshal the identiy card: %w", err)
	}

	// Check if the SSOID in the identity card is not empty
	if len(peerIdentityCard.SSOID) == 0 {
		return SignedMessage{}, IdentityCard{}, false, fmt.Errorf("node/ReceiveAndVerifySignedIdentityCard - Got an empty SSOID")
	}

	// Verify the peer's signature
	err = peerSignedMessage.VerifySignature(&peerIdentityCard.PublicKey)
	if err != nil {
		return SignedMessage{}, IdentityCard{}, false, fmt.Errorf("node/ReceiveAndVerifySignedIdentityCard - Could not verify the peer's signature: %w", err)
	}

	return peerSignedMessage, peerIdentityCard, false, nil
}
