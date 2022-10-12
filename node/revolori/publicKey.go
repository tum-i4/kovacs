package revolori

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// GetPublicKey gets Revolori's public key from env and parses it to a PublicKey struct.
func GetPublicKey() (rsa.PublicKey, error) {
	revoloriBase, ok := os.LookupEnv("REVOLORI_ADDRESS")
	if !ok || len(strings.TrimSpace(revoloriBase)) == 0 {
		return rsa.PublicKey{}, fmt.Errorf("node/GetRevoloriPublicKey - REVOLORI_ADDRESS is not set")
	}

	resp, err := http.Get(revoloriBase + "/key/show")
	if err != nil {
		return rsa.PublicKey{}, fmt.Errorf("node/GetRevoloriPublicKey - Could not contact revolori: %w", err)
	}
	defer resp.Body.Close()

	// Read the body
	revoloriPublicKeyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rsa.PublicKey{}, fmt.Errorf("node/GetRevoloriPublicKey - Could not read the response body: %w", err)
	}

	var publicKey rsa.PublicKey
	err = json.Unmarshal(revoloriPublicKeyBytes, &publicKey)
	if err != nil {
		return rsa.PublicKey{}, fmt.Errorf("node/GetRevoloriPublicKey - Could not unmarshal revolori's key: %w", err)
	}

	return publicKey, nil
}
