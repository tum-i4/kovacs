package storage

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	// Needed for sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
	"node/logging"
)

func ExportToSQLite(justification string, datum string, ownerPublicKey *rsa.PublicKey, consumerPublicKey *rsa.PublicKey) error {
	block, err := createBlockchainPayload(justification, datum, ownerPublicKey, consumerPublicKey)
	if err != nil {
		log.Info.Printf("listener/ExportToSQLite - %v", err)
		return err
	}

	db, err := openOrInitDB()
	if err != nil {
		log.Info.Printf("listener/ExportToSQLite - %v", err)
		return err
	}

	return commitToDB(db, &block)
}

func commitToDB(db *sql.DB, data *BlockchainPayload) error {
	start := time.Now()
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("node/commitToDB - could not begin transaction: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO exportTable (PseudonymConsumer, PseudonymOwner, EncryptedConsumer, EncryptedOwner) VALUES (?,?,?,?)")
	if err != nil {
		return fmt.Errorf("node/commitToDB - could not prepare statement: %w", err)
	}
	defer stmt.Close()

	encryptedConsumer, err := json.Marshal(data.EncryptedConsumer)
	if err != nil {
		return fmt.Errorf("node/commitToDB - could not marshal encrypted consumer: %w", err)
	}

	encryptedOwner, err := json.Marshal(data.EncryptedOwner)
	if err != nil {
		return fmt.Errorf("node/commitToDB - could not marshal encrypted owner: %w", err)
	}

	_, err = stmt.Exec(data.PseudonymConsumer, data.PseudonymOwner, string(encryptedConsumer), string(encryptedOwner))
	if err != nil {
		return fmt.Errorf("node/commitToDB - could not execute statement: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("node/commitToDB - could not commit transaction: %w", err)
	}
	duration := time.Since(start)

	log.Info.Printf(""+
		"SQLite export duration: %s",
		duration,
	)

	return nil
}

func openOrInitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return nil, fmt.Errorf("node/openOrInitDB - could not open db: %w", err)
	}

	// Create table if it does not exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS exportTable (PseudonymConsumer text, PseudonymOwner text, EncryptedConsumer text, EncryptedOwner text)")
	if err != nil {
		return nil, fmt.Errorf("node/openOrInitDB - could not create table: %w", err)
	}

	return db, nil
}
