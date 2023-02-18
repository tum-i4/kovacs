package storage

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"node/constants"
	"node/p2p"
)

type storedExchange struct {
	PrivateKey        rsa.PrivateKey      `json:"private_key"`
	PublicIdentityKey rsa.PublicKey       `json:"public_identity_key"`
	Messages          []p2p.SignedMessage `json:"messages"`
}

func StoreExchange(messages []p2p.SignedMessage, privateKey *rsa.PrivateKey, publicIdentityKey *rsa.PublicKey) error {
	if len(messages) == 0 {
		return errors.New("node.Store - Message is either null or empty")
	} else if privateKey.Equal(rsa.PrivateKey{}) {
		return errors.New("node.Store - Empty private key")
	}

	// Check if output directory exists and create it if necessary
	err := createOutputDirectory(constants.StorageOutputPath)
	if err != nil {
		return fmt.Errorf("node.Store - Could not create output direcory: %w", err)
	}

	// Generate unique file name
	pseudonym, err := GeneratePseudonym(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("node.Store - Could not generate pseudonym: %w", err)
	}
	// fileName length is 84+5 characters
	fileName := fmt.Sprintf("%s-%s.json", time.Now().Format("2006-01-02T15-04-05"), pseudonym)

	// Marshal data into JSON
	toWrite := storedExchange{
		Messages:          messages,
		PrivateKey:        *privateKey,
		PublicIdentityKey: *publicIdentityKey,
	}

	out, err := json.Marshal(toWrite)
	if err != nil {
		return fmt.Errorf("node.Store - Could not marshal json: %w", err)
	}

	err = os.WriteFile(constants.StorageOutputPath+fileName, out, 0o644) //nolint: gosec
	if err != nil {
		return fmt.Errorf("node.Store - Could not write to file: %w", err)
	}

	return nil
}

func LoadExchange(path string) ([]p2p.SignedMessage, rsa.PrivateKey, rsa.PublicKey, error) {
	readBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, rsa.PrivateKey{}, rsa.PublicKey{}, fmt.Errorf("node.Load - Could not load file: %w", err)
	}

	var exchange storedExchange
	dec := json.NewDecoder(bytes.NewReader(readBytes))
	dec.DisallowUnknownFields()

	err = dec.Decode(&exchange)
	if err != nil {
		return nil, rsa.PrivateKey{}, rsa.PublicKey{}, fmt.Errorf("node.Load - Could not marshal file: %w", err)
	}

	if !filenameMatchesPseudonym(path, &exchange.PrivateKey.PublicKey) {
		return nil, rsa.PrivateKey{}, rsa.PublicKey{}, fmt.Errorf("node.Load - Pseudonym of file and key do not match")
	}

	if len(exchange.Messages) < 3 {
		return nil, rsa.PrivateKey{}, rsa.PublicKey{}, fmt.Errorf("node.Load - Too little messages were found")
	}

	return exchange.Messages, exchange.PrivateKey, exchange.PublicIdentityKey, nil
}

func filenameMatchesPseudonym(fileName string, publicKey *rsa.PublicKey) bool {
	pseudonymFile := fileName[strings.LastIndex(fileName, "-")+1:]
	pseudonymFile = strings.TrimSuffix(pseudonymFile, ".json")

	pseudonymKey, err := GeneratePseudonym(publicKey)
	if err != nil {
		return false
	}

	return pseudonymFile == pseudonymKey
}

func createOutputDirectory(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		// Directory exists
		return nil
	} else if os.IsNotExist(err) {
		// Directory does not exist => Create it
		return os.Mkdir(path, 0o777)
	}

	return err
}
