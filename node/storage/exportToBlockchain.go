package storage

import (
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	"node/logging"
	"node/random"
)

type transaction struct {
	From  string
	To    string
	Input string
	Value string
}

var transactionMutex sync.Mutex

func ExportToBlockchain(justification string, datum string, ownerPublicKey *rsa.PublicKey, consumerPublicKey *rsa.PublicKey) error {
	block, err := createBlockchainPayload(justification, datum, ownerPublicKey, consumerPublicKey)
	if err != nil {
		log.Info.Printf("listener/exportToBlockchain - %v", err)
		return err
	}

	blockBytes, err := json.Marshal(block)
	if err != nil {
		log.Info.Printf("listener/exportToBlockchain - Could not marshal the block: %v", err)
		return err
	}

	return commitToBlockchain(blockBytes)
}

func commitToBlockchain(input []byte) error { //nolint:funlen
	password := random.String(32)

	// This function is called from a go routine. Mutex is needed to prevent a go routine deleting accounts that could
	// still be in use.
	transactionMutex.Lock()
	defer transactionMutex.Unlock()

	accountCreationStart := time.Now()
	// Create first account
	firstAccountAddress, err := makeGethRequestString("personal_newAccount", []string{password})
	if err != nil {
		return fmt.Errorf("node/commitToBlockchain - Could not create first account: %w", err)
	}
	accountCreationDuration := time.Since(accountCreationStart)

	start := time.Now()
	// Unlock the first account which is needed for sending the transaction
	// This requires the "--allow-insecure-unlock" flag to be set for geth client
	_, err = makeGethRequestString("personal_unlockAccount", []string{firstAccountAddress, password})
	if err != nil {
		return fmt.Errorf("node/commitToBlockchain - Could not unlock account: %w", err)
	}
	accountUnlockDuration := time.Since(start)

	// Send all transactions to the same address, since the from and to are not useful for us
	secondAccountAddress := "0x0000000000000000000000000000000000000000"

	beforeFirstMine := time.Now()
	// Mine with first account in order to be able to make a transfer
	err = mine(firstAccountAddress)
	if err != nil {
		return fmt.Errorf("node/commitToBlockchain - %w", err)
	}
	firstMineDuration := time.Since(beforeFirstMine)

	beforeTransaction := time.Now()
	// Create transaction
	transactionArray := []transaction{{
		From:  firstAccountAddress,
		To:    secondAccountAddress,
		Input: "0x" + hex.EncodeToString(input),
		Value: "0x1",
	}}

	_, err = makeGethRequestString("eth_sendTransaction", transactionArray)
	if err != nil {
		return fmt.Errorf("node/commitToBlockchain - Could not send transaction: %w", err)
	}
	transactionDuration := time.Since(beforeTransaction)

	beforeSecondMine := time.Now()
	// Mine in order to move the transaction from pending to blockchain
	err = mine(firstAccountAddress)
	if err != nil {
		return fmt.Errorf("node/commitToBlockchain - %w", err)
	}
	secondMineDuration := time.Since(beforeSecondMine)

	log.Info.Printf(""+
		"Timing export duration"+
		"\n\tAccount creation: %s"+
		"\n\tAccount unlock: %s"+
		"\n\tMining two blocks: %s"+
		"\n\tTransaction: %s",
		accountCreationDuration,
		accountUnlockDuration,
		firstMineDuration+secondMineDuration,
		transactionDuration,
	)

	return deleteKeystore()
}

func mine(accountAddress string) error {
	// Set the first account as miner
	_, err := makeGethRequestString("miner_setEtherbase", []string{accountAddress})
	if err != nil {
		return fmt.Errorf("could not update etherbase: %w", err)
	}

	// Mine with 2 threads
	_, err = makeGethRequestString("miner_start", []int{})
	if err != nil {
		return fmt.Errorf("could not start miner: %w", err)
	}

	var balanceHex string
	var pendingTransactions bool
	for {
		// Check balance
		balanceHex, err = makeGethRequestString("eth_getBalance", []string{accountAddress, "latest"})
		if err != nil {
			return fmt.Errorf("could not get balance: %w", err)
		}

		if balanceHex != "0x0" {
			// Check for pending transactions
			pendingTransactions, err = checkForPendingTransactions()
			if err != nil {
				return err
			}

			if !pendingTransactions {
				break
			}
		}

		time.Sleep(250 * time.Millisecond)
	}

	_, err = makeGethRequestString("miner_stop", []string{})
	if err != nil {
		return fmt.Errorf("could not stop miner: %w", err)
	}

	return nil
}

func checkForPendingTransactions() (bool, error) {
	response, err := makeGethRequestInterface("eth_pendingTransactions", []string{})
	if err != nil {
		return false, err
	}

	result := response.Result
	var expectedType []interface{}
	if reflect.TypeOf(result) != reflect.TypeOf(expectedType) {
		return false, fmt.Errorf("got invalid pending request type: '%s'; expected %s", reflect.TypeOf(result), reflect.TypeOf(expectedType))
	}

	return len(result.([]interface{})) > 0, nil //nolint: forcetypeassert
}

func deleteKeystore() error {
	gethKeystorePath, ok := os.LookupEnv("GETH_KEYSTORE_PATH")
	if !ok || len(gethKeystorePath) == 0 {
		return fmt.Errorf("node/deleteKeystore - could not find 'GETH_KEYSTORE_PATH' in env")
	}

	err := os.RemoveAll(gethKeystorePath)
	if err != nil {
		return fmt.Errorf("node/deleteKeyStore - Could not delete keystore: %w", err)
	}

	err = os.Mkdir(gethKeystorePath, 0o700)
	if err != nil {
		return fmt.Errorf("node/deleteKeyStore - Could not create keystore: %w", err)
	}

	return nil
}
