package storage

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	log "node/logging"
)

func QuerySingleLogFromBlockchain(pseudonym string) (BlockchainPayload, error) {
	// Get current block number
	blockNumberHex, err := makeGethRequestString("eth_blockNumber", []string{})
	if err != nil {
		return BlockchainPayload{}, fmt.Errorf("node/QuerySingleBlockchainLog - could not get current block number: %w", err)
	}

	blockNumber, err := strconv.ParseInt(blockNumberHex, 0, 64)
	if err != nil {
		return BlockchainPayload{}, fmt.Errorf("node/QuerySingleBlockchainLog - could not parse block number ('%s') because: %w", blockNumberHex, err)
	}

	var i int64
	for i = 0; i < blockNumber+1; i++ {
		blockResult, err := queryBlockByNumber(i)
		if err != nil {
			return BlockchainPayload{}, fmt.Errorf("node/QuerySingleBlockchainLog - %w", err)
		}

		for _, payload := range blockResult {
			if payload.PseudonymOwner == pseudonym || payload.PseudonymConsumer == pseudonym {
				return payload, nil
			}
		}
	}

	return BlockchainPayload{}, fmt.Errorf("node/QuerySingleBlockchainLog - could not find requested log")
}

func QueryAllLogs() ([]BlockchainPayload, error) {
	// Get current block number
	blockNumberHex, err := makeGethRequestString("eth_blockNumber", []string{})
	if err != nil {
		return nil, fmt.Errorf("node/QueryAll - Could not get current block number: %w", err)
	}

	blockNumber, err := strconv.ParseInt(blockNumberHex, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("node/QueryAll - Could not parse block number ('%s') because: %w", blockNumberHex, err)
	}

	var i int64
	var allLogs = make([]BlockchainPayload, 0)
	for i = 0; i < blockNumber+1; i++ {
		blockResult, err := queryBlockByNumber(i)
		if err != nil {
			return nil, fmt.Errorf("node/QueryAll - %w", err)
		}

		allLogs = append(allLogs, blockResult...)
	}

	return allLogs, nil
}

func queryBlockByNumber(i int64) ([]BlockchainPayload, error) { //nolint:funlen
	var response gethResponse
	var expectedResultType map[string]interface{}
	var expectedTransactionArrayType []interface{}
	var expectedTransactionType map[string]interface{}

	hexNumber := "0x" + strconv.FormatInt(i, 16)
	response, err := makeGethRequestInterface("eth_getBlockByNumber", []interface{}{hexNumber, true})
	if err != nil {
		return nil, fmt.Errorf("node/queryBlockByNumber - Could not get block with id '%s' because: %w", hexNumber, err)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("node/queryBlockByNumber - Got invalid response. Result type: '%s'; expected %s", reflect.TypeOf(response.Result), reflect.TypeOf(expectedResultType))
	}

	transactionsInterface, ok := result["transactions"]
	if !ok {
		return nil, fmt.Errorf("node/queryBlockByNumber - response.Result is missing 'transactions'-key: %v", result)
	}

	transactions, ok := transactionsInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("node/queryBlockByNumber - Got invalid []transaction type: '%s'; expected %s", reflect.TypeOf(transactionsInterface), reflect.TypeOf(expectedTransactionArrayType))
	}

	if len(transactions) == 0 {
		return nil, nil
	}

	logs := make([]BlockchainPayload, 0)
	var payload BlockchainPayload
	for _, singleTransaction := range transactions {
		transactionMap, ok := singleTransaction.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("node/queryBlockByNumber - Got invalid transaction type: '%s'; expected %s", reflect.TypeOf(singleTransaction), reflect.TypeOf(expectedTransactionType))
		}

		input, ok := transactionMap["input"]
		if !ok {
			return nil, fmt.Errorf("node/queryBlockByNumber - transaction is missing 'input'-key: %v", result)
		}

		if reflect.TypeOf(input) != reflect.TypeOf("") {
			return nil, fmt.Errorf("node/queryBlockByNumber - Got invalid transaction.input type: '%s'; expected %s", reflect.TypeOf(input), reflect.TypeOf(""))
		}

		out, err := hex.DecodeString(strings.Replace(input.(string), "0x", "", 1))
		if err != nil {
			return nil, fmt.Errorf("node/queryBlockByNumber - could not decode input: %w", err)
		}

		err = json.Unmarshal(out, &payload)
		if err != nil {
			log.Error.Printf("node/queryBlockchain - Found a malformed transaction input: %s\n", err)
			continue
		}

		logs = append(logs, payload)
	}

	return logs, nil
}
