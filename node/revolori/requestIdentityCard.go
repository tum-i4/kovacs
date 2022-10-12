package revolori

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/term"
	"node/constants"
	"node/logging"
	nP "node/nonRepudiation"
	"node/p2p"
)

// Setup tries to load a private key. If the key exists then it looks for the related identity card. If the key doesn't
// exist then a new one is created, and it is sent to Revolori to be signed.
func Setup(deleteOnFailure bool) (rsa.PrivateKey, error) {
	_, err := os.Stat(constants.KeyFilePath)
	if err == nil { //nolint: gocritic
		// Key file exists => return it
		return loadPrivateKey(deleteOnFailure)
	} else if os.IsNotExist(err) {
		// Key file does not exist => create it
		return requestPrivateKey(deleteOnFailure)
	} else {
		// Some other error
		if deleteOnFailure {
			revoloriCleanUpOnFailure()
		}

		return rsa.PrivateKey{}, err
	}
}

func loadPrivateKey(deleteOnFailure bool) (rsa.PrivateKey, error) {
	log.Info.Printf("Loading an existing key file: ")
	key, errLoadKey := loadPrivateRSAKey(constants.KeyFilePath)
	if errLoadKey != nil {
		log.Info.Println("Failure")
		if deleteOnFailure {
			revoloriCleanUpOnFailure()
		}

		return rsa.PrivateKey{}, errLoadKey
	}
	log.Info.Println("Success")

	log.Info.Printf("Looking for identity.json: ")
	_, err := os.Stat(constants.IdentityFilePath)
	if err == nil {
		log.Info.Printf("Found\n")
		return key, nil
	} else if os.IsNotExist(err) {
		log.Info.Printf("Not found\n")
		if deleteOnFailure {
			revoloriCleanUpOnFailure()
		}

		return rsa.PrivateKey{}, err
	}

	log.Info.Printf("An error occurred\n")
	if deleteOnFailure {
		revoloriCleanUpOnFailure()
	}

	return rsa.PrivateKey{}, err
}

func requestPrivateKey(deleteOnFailure bool) (rsa.PrivateKey, error) {
	log.Info.Print("Creating a new key:")
	key, errKey := createPrivateRSAKey(constants.KeyFilePath)
	if errKey != nil {
		log.Info.Println("\tFailure")
		if deleteOnFailure {
			revoloriCleanUpOnFailure()
		}

		return rsa.PrivateKey{}, errKey
	}
	log.Info.Println("\tSuccess")

	// Get Signature method
	var username, password, token []byte
	var err error
	var ok bool

	// Try to load credentials from env
	username, password, token, ok = readCredentialsFromEnv()
	if !ok {
		log.Info.Println("Could not load credentials from env")
		username, password, token, err = readCredentialsFromTerm()
		if err != nil {
			if deleteOnFailure {
				revoloriCleanUpOnFailure()
			}

			return rsa.PrivateKey{}, err
		}
	}

	// Get Revolori's address from env
	revoloriBase, ok := os.LookupEnv("REVOLORI_ADDRESS")
	if !ok || len(strings.TrimSpace(revoloriBase)) == 0 {
		return rsa.PrivateKey{}, fmt.Errorf("node/requestPrivateKey - REVOLORI_ADDRESS is not set")
	}

	revoloriURL := revoloriBase + "/key/sign"

	// Request Revolori to sign it
	if token != nil {
		log.Info.Print("Signing the private key with a token:")
		err = requestRevoloriSignatureWithToken(revoloriURL, constants.IdentityFilePath, &key.PublicKey, token)
	} else {
		log.Info.Print("Signing the private key with username and password")
		err = requestRevoloriSignatureWithCredentials(revoloriURL, constants.IdentityFilePath, &key.PublicKey, username, password)
	}

	if err != nil {
		log.Info.Println("\tFailure")
		if deleteOnFailure {
			revoloriCleanUpOnFailure()
		}

		return rsa.PrivateKey{}, err
	}

	log.Info.Println("\tSuccess")
	return key, nil
}

// requestRevoloriSignatureWithCredentials sends the public key with the username and password to Revolori so that they can be signed.
func requestRevoloriSignatureWithCredentials(revoloriURL string, identityFilePath string, publicKey *rsa.PublicKey, username []byte, password []byte) error {
	data := struct {
		Email     string        `json:"email"`
		Password  string        `json:"password"`
		PublicKey rsa.PublicKey `json:"publicKey"`
	}{
		Email:     string(username),
		Password:  string(password),
		PublicKey: *publicKey,
	}

	marshaledData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("node/requestRevoloriSignatureWithCredentials - Could not marshal json: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, revoloriURL, strings.NewReader(string(marshaledData)))
	if err != nil {
		return fmt.Errorf("node/requestRevoloriSignatureWithCredentials - Could not create new request: %w", err)
	}

	return makeRequest(request, identityFilePath)
}

// requestRevoloriSignatureWithToken sends the public key along with the token cookie to be signed by Revolori.
func requestRevoloriSignatureWithToken(revoloriURL string, identityFilePath string, publicKey *rsa.PublicKey, token []byte) error {
	data := struct {
		PublicKey rsa.PublicKey `json:"publicKey"`
	}{
		PublicKey: *publicKey,
	}

	marshaledData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("node/requestRevoloriSignatureWithToken - Could not marshal data: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, revoloriURL, strings.NewReader(string(marshaledData)))
	if err != nil {
		return fmt.Errorf("node/requestRevoloriSignatureWithToken - Could not create new request: %w", err)
	}

	request.AddCookie(&http.Cookie{Name: "token", Value: strings.Replace(string(token), "\"", "", 2)})

	return makeRequest(request, identityFilePath)
}

func makeRequest(request *http.Request, identityFilePath string) error {
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("node/makeRequest - Could not make request: %w", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("node/makeRequest - Could not read body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("node/makeRequest - Status code of '%d' indicates failure. Body: %s", response.StatusCode, string(body))
	}

	var signedMessage p2p.SignedMessage

	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&signedMessage)
	if err != nil {
		return fmt.Errorf("node/makeRequest - Could not unmarshal json (%s): %w", string(body), err)
	}

	err = ioutil.WriteFile(identityFilePath, body, 0o644) //nolint: gosec
	if err != nil {
		return fmt.Errorf("node/makeRequest - Could not write identity file: %w", err)
	}

	return nil
}

// readCredentialsFromEnv tries to load (username, password) or token from environment.
func readCredentialsFromEnv() ([]byte, []byte, []byte, bool) {
	// Check if a token exists
	token, ok := os.LookupEnv("REVOLORI_TOKEN")
	if ok && len(token) > 0 {
		return nil, nil, []byte(token), true
	}

	// Check if username exists
	username, ok := os.LookupEnv("REVOLORI_USERNAME")
	if !ok || len(strings.TrimSpace(username)) == 0 {
		return nil, nil, nil, false
	}

	// Check if password exists
	password, ok := os.LookupEnv("REVOLORI_PASSWORD")
	if !ok || len(strings.TrimSpace(password)) == 0 {
		return nil, nil, nil, false
	}

	return []byte(username), []byte(password), nil, true
}

// readCredentialsFromTerm returns (username, password) or token.
func readCredentialsFromTerm() ([]byte, []byte, []byte, error) {
	fmt.Print("Do you want to login via [u]sername or [t]oken: ")
	reader := bufio.NewReader(os.Stdin)
	stdin := int(os.Stdin.Fd())

	inputType, err := reader.ReadString('\n')
	if err != nil {
		return nil, nil, nil, fmt.Errorf("node/readCredentialsFromTerm - Could not read input type: %w", err)
	}
	inputType = strings.TrimSpace(inputType)

	if inputType == "t" { //nolint:nestif
		fmt.Print("Please enter the token: ")
		token, err := term.ReadPassword(stdin)
		fmt.Println()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("node/readCredentialsFromTerm - Could not read the token: %w", err)
		}
		token = []byte(strings.TrimSpace(string(token)))

		if len(token) == 0 {
			return nil, nil, nil, errors.New("node/readCredentialsFromTerm - Empty token passed")
		}

		return nil, nil, token, nil
	} else if inputType == "u" {
		fmt.Print("Please enter the email address: ")
		username, err := term.ReadPassword(stdin)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("node/readCredentialsFromTerm - Could not read the email: %w", err)
		}
		fmt.Println()
		username = []byte(strings.TrimSpace(string(username)))

		if len(username) == 0 {
			return nil, nil, nil, errors.New("node/readCredentialsFromTerm - Empty email passed")
		}

		fmt.Print("Please enter the password: ")
		password, err := term.ReadPassword(stdin)
		fmt.Println()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("node/readCredentialsFromTerm - Could not read the password: %w", err)
		}
		password = []byte(strings.TrimSpace(string(password)))

		if len(password) == 0 {
			return nil, nil, nil, errors.New("node/readCredentialsFromTerm - Empty password passed")
		}

		return username, password, nil, nil
	}

	return nil, nil, nil, fmt.Errorf("node/readCredentialsFromTerm - Invalid input type: " + strings.TrimSpace(inputType))
}

func revoloriCleanUpOnFailure() {
	log.Info.Print("Detected a failure in the signup process. Cleaning up")

	_helper := func(path string) {
		_, err := os.Stat(path)
		if err == nil {
			err = os.Remove(path)
			if err != nil {
				log.Info.Printf("\tCould not delete file: %v", err)
				return
			}
		} else if !os.IsNotExist(err) {
			log.Info.Printf("\tAn error occurred when calling stat: %v", err)
			return
		}
	}

	_helper(constants.KeyFilePath)
	_helper(constants.IdentityFilePath)
}

// loadPrivateRSAKey reads, parses and returns the private key.
func loadPrivateRSAKey(path string) (rsa.PrivateKey, error) {
	pemContent, err := ioutil.ReadFile(path)
	if err != nil {
		return rsa.PrivateKey{}, err
	}

	block, _ := pem.Decode(pemContent)
	if block == nil || block.Type != constants.PemKeyName {
		return rsa.PrivateKey{}, err
	}

	pub, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return rsa.PrivateKey{}, err
	}

	return *pub, nil
}

// createPrivateRSAKey creates a private key and writes it as pem to the passed path.
func createPrivateRSAKey(path string) (rsa.PrivateKey, error) {
	key, err := nP.GenerateRSAPrivateKey()
	if err != nil {
		return rsa.PrivateKey{}, err
	}

	block := &pem.Block{
		Type:  constants.PemKeyName,
		Bytes: x509.MarshalPKCS1PrivateKey(&key),
	}

	content := pem.EncodeToMemory(block)

	return key, ioutil.WriteFile(path, content, 0o644) //nolint: gosec
}
