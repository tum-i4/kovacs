package p2p

import (
	"crypto/rsa"
	"errors"
	"strings"

	"node/constants"
)

type FirstMessage struct {
	Datum         string                `json:"datum"`
	Justification string                `json:"justification"`
	PublicKey     rsa.PublicKey         `json:"public_key"`
	Type          constants.MessageType `json:"type"`
}

// CheckForContent verifies that the struct's fields are not empty.
func (message *FirstMessage) CheckForContent() error {
	if len(message.Datum) == 0 {
		return errors.New("FirstMessage.CheckForContent - The datum field is empty")
	}

	if message.Type != constants.MessageTypeRequester && message.Type != constants.MessageTypeListener && message.Type != constants.MessageTypeFakeChatter {
		return errors.New("FirstMessage.CheckForContent - Invalid type")
	}

	emptyKey := rsa.PublicKey{} //nolint: ifshort
	if message.PublicKey == emptyKey {
		return errors.New("FirstMessage.CheckForContent - The public key field is empty")
	}

	return nil
}

func (message *FirstMessage) CheckForContentAndJustification() error {
	err := message.CheckForContent()
	if err != nil {
		return err
	}

	if len(strings.TrimSpace(message.Justification)) == 0 {
		return errors.New("FirstMessage.CheckForContentAndJustification - Missing justification: " + message.Justification)
	}

	return nil
}
