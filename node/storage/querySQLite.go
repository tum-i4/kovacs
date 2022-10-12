package storage

import (
	"encoding/json"
	"fmt"
)

func QuerySingleLogFromSQLite(pseudonym string) (BlockchainPayload, error) {
	db, err := openOrInitDB()
	if err != nil {
		return BlockchainPayload{}, err
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("select * from exportTable where PseudonymConsumer='%s' or PseudonymOwner='%s'", pseudonym, pseudonym))
	if err != nil {
		return BlockchainPayload{}, err
	}

	if rows.Err() != nil {
		return BlockchainPayload{}, rows.Err()
	}
	defer rows.Close()

	var payload BlockchainPayload
	var usageLogConsumer UsageLogContent
	var usageLogOwner UsageLogContent

	for rows.Next() {
		var encryptedConsumerStr string
		var encryptedOwnerStr string

		err = rows.Scan(&payload.PseudonymConsumer, &payload.PseudonymOwner, &encryptedConsumerStr, &encryptedOwnerStr)
		if err != nil {
			return BlockchainPayload{}, err
		}

		err = json.Unmarshal([]byte(encryptedConsumerStr), &usageLogConsumer)
		if err != nil {
			return BlockchainPayload{}, err
		}

		err = json.Unmarshal([]byte(encryptedOwnerStr), &usageLogOwner)
		if err != nil {
			return BlockchainPayload{}, err
		}

		payload.EncryptedConsumer = usageLogConsumer
		payload.EncryptedOwner = usageLogOwner

		return payload, nil //nolint: staticcheck
	}

	return BlockchainPayload{}, fmt.Errorf("node/QuerySingleLogsFromSQLite - could not find the requested log")
}

func QueryAllLogsFromSQLite() ([]BlockchainPayload, error) {
	db, err := openOrInitDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select * from exportTable")
	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()

	var payload BlockchainPayload
	var usageLogConsumer UsageLogContent
	var usageLogOwner UsageLogContent

	list := make([]BlockchainPayload, 0)
	for rows.Next() {
		var encryptedConsumerStr string
		var encryptedOwnerStr string

		err = rows.Scan(&payload.PseudonymConsumer, &payload.PseudonymOwner, &encryptedConsumerStr, &encryptedOwnerStr)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(encryptedConsumerStr), &usageLogConsumer)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(encryptedOwnerStr), &usageLogOwner)
		if err != nil {
			return nil, err
		}

		payload.EncryptedConsumer = usageLogConsumer
		payload.EncryptedOwner = usageLogOwner

		list = append(list, payload)
	}

	return list, nil
}
