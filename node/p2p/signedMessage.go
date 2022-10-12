package p2p

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"node/constants"
)

type SignedMessage struct {
	// Content has to be []byte otherwise the unmarshal fails
	Content   []byte `json:"content"`
	Signature []byte `json:"signature"`
}

// Only used for the identiy card exchange.
type ExtendedSignedMessage struct {
	Content   []byte                `json:"content"`
	Signature []byte                `json:"signature"`
	Type      constants.MessageType `json:"type"`
}

func (message *SignedMessage) VerifySignature(publicKey *rsa.PublicKey) error {
	hashed := sha256.Sum256(message.Content)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], message.Signature)
}

// ReceiveAndVerifySignedMessage receives the read bytes, tries to convert them into a SignedMessage. if that step was completed
// successfully then the signature is verified. If that step was completed successfully then the SignedMessage.Content
// is returned.
func ReceiveAndVerifySignedMessage(rw *bufio.ReadWriter, publicKey *rsa.PublicKey, returnStruct interface{}, waitTime ...time.Duration) (SignedMessage, error) {
	var signedMessage SignedMessage

	var received string
	var err error
	// Read the struct
	if len(waitTime) > 0 {
		// Wait time parameter was given
		received, err = ReadLine(rw, '}', waitTime[0])
	} else {
		// Use default time out
		received, err = ReadLine(rw, '}')
	}

	if err != nil {
		return SignedMessage{}, err
	}

	// Turn []bytes to SignedMessage struct
	err = json.Unmarshal([]byte(received), &signedMessage)
	if err != nil {
		return SignedMessage{}, fmt.Errorf("could not unmarshal signed response: %w", err)
	}

	// Verify the signature
	err = signedMessage.VerifySignature(publicKey)
	if err != nil {
		return SignedMessage{}, fmt.Errorf("could not verify signature: %w", err)
	}

	// Try to unmarshal Signature.Content
	err = json.Unmarshal(signedMessage.Content, returnStruct)
	if err != nil {
		return SignedMessage{}, fmt.Errorf("could not convert signed message to the given struct: %w", err)
	}

	return signedMessage, nil
}

// ReceiveAndVerifyFirstMessage is basically ReceiveAndVerifySignedMessage with optional signature verification for fake chatter.
func ReceiveAndVerifyFirstMessage(rw *bufio.ReadWriter, publicKey *rsa.PublicKey, returnStruct interface{}, isFakeChatter bool, waitTime ...time.Duration) (SignedMessage, error) {
	var signedMessage SignedMessage

	var received string
	var err error
	// Read the struct
	if len(waitTime) > 0 {
		// Wait time parameter was given
		received, err = ReadLine(rw, '}', waitTime[0])
	} else {
		// Use default time out
		received, err = ReadLine(rw, '}')
	}

	if err != nil {
		return SignedMessage{}, err
	}

	// Turn []bytes to SignedMessage struct
	err = json.Unmarshal([]byte(received), &signedMessage)
	if err != nil {
		return SignedMessage{}, fmt.Errorf("could not unmarshal signed response: %w", err)
	}

	// Signature verification is only necessary if the exchange is real
	if !isFakeChatter {
		err = signedMessage.VerifySignature(publicKey)
		if err != nil {
			return SignedMessage{}, fmt.Errorf("could not verify signature: %w", err)
		}
	}

	// Try to unmarshal Signature.Content
	err = json.Unmarshal(signedMessage.Content, returnStruct)
	if err != nil {
		return SignedMessage{}, fmt.Errorf("could not convert signed message to the given struct: %w", err)
	}

	return signedMessage, nil
}

func CreateAndSendSignedMessage(messageStruct interface{}, privateKey *rsa.PrivateKey, rw *bufio.ReadWriter) error {
	// Created a SignedMessage
	signedMessage, err := CreateSignedMessage(messageStruct, privateKey)
	if err != nil {
		return fmt.Errorf("node/CreateAndSendSignedMessage - could not create SignedMessage: %w", err)
	}

	return SendSignedMessage(signedMessage, rw)
}

func SendSignedMessage(signedMessage SignedMessage, rw *bufio.ReadWriter) error {
	// Create a SignedMessage JSON
	js, err := json.Marshal(signedMessage)
	if err != nil {
		return fmt.Errorf("node/CreateAndSendSignedMessage - could not marshal signedMessage: %w", err)
	}

	// Send the json
	err = Write(rw, string(js))
	if err != nil {
		return fmt.Errorf("node/CreateAndSendSignedMessage - could not write json: %w", err)
	}

	return nil
}

func CreateSignedMessage(messageStruct interface{}, privateKey *rsa.PrivateKey) (SignedMessage, error) {
	message, err := json.Marshal(messageStruct)
	if err != nil {
		return SignedMessage{}, err
	}

	signature, err := calculateSignature(message, privateKey)
	if err != nil {
		return SignedMessage{}, err
	}

	return SignedMessage{
		Content:   message,
		Signature: signature,
	}, nil
}

func CreateSendAndReturnSignedMessage(messageStruct interface{}, privateKey *rsa.PrivateKey, rw *bufio.ReadWriter) ([]byte, error) {
	ret, err := CreateSignedMessage(messageStruct, privateKey)
	if err != nil {
		return nil, err
	}

	retBytes, err := json.Marshal(ret)
	if err != nil {
		return nil, err
	}

	err = CreateAndSendSignedMessage(messageStruct, privateKey, rw)
	if err != nil {
		return nil, err
	}

	return retBytes, nil
}

func ExtractAndVerifyMessages(signedMessages []SignedMessage, revoloriPublicKey *rsa.PublicKey) (FirstMessage, IdentityCard, error) {
	if len(signedMessages) < 2 {
		return FirstMessage{}, IdentityCard{}, fmt.Errorf("there are too few messages (%d)", len(signedMessages))
	}

	// IdentityCard at signedMessages[0]
	signedIdentityCard := signedMessages[0]
	_, identityCard, isFakeChatter, err := VerifySignedIdentityCard(signedIdentityCard, revoloriPublicKey)
	if err != nil {
		return FirstMessage{}, IdentityCard{}, err
	}

	if isFakeChatter {
		return FirstMessage{}, IdentityCard{}, fmt.Errorf("identity card is marked as fake chatter")
	}

	// FirstMessage at signedMessages[1]
	signedFirstMessage := signedMessages[1]
	err = signedFirstMessage.VerifySignature(&identityCard.PublicKey)
	if err != nil {
		return FirstMessage{}, IdentityCard{}, err
	}

	var firstMessage FirstMessage
	err = json.Unmarshal(signedFirstMessage.Content, &firstMessage)
	if err != nil {
		return FirstMessage{}, IdentityCard{}, fmt.Errorf("could not unmarshal first message: %w", err)
	}

	err = firstMessage.CheckForContent()
	if err != nil {
		return FirstMessage{}, IdentityCard{}, fmt.Errorf("first message contains invalid content: %w", err)
	}

	return firstMessage, identityCard, nil
}

func calculateSignature(message []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	hashed := sha256.Sum256(message)
	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
}
