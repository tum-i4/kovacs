package main

import (
	"fmt"
	"node/random"
	"testing"
	"time"

	"node/logging"
	"node/storage"
)

func BenchmarkQueryEntireBlockchain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		start := time.Now()
		allLogs, err := storage.QueryAllLogs()
		duration := time.Since(start)

		if err != nil {
			log.Error.Printf("Could not query logs from blockchain: %s", err)
			b.Errorf("Could not query logs from blockchain: %s", err)
		}

		log.Info.Printf("Reading from blockchain\n\tEntries: %d\n\tDuration: %v\n", len(allLogs), duration)
	}
}

func BenchmarkQueryEntireSQLiteDB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		start := time.Now()
		allLogs, err := storage.QueryAllLogsFromSQLite()
		duration := time.Since(start)

		if err != nil {
			log.Error.Printf("Could not query logs from SQLite DB: %s", err)
			b.Errorf("Could not query logs from SQLite DB: %s", err)
		}

		log.Info.Printf("Reading from SQLite DB\n\tEntries: %d\n\tDuration: %v\n", len(allLogs), duration)
	}
}

func BenchmarkSearchSingleLogBlockchain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pseudonym, err := _getRandomPseudonym("../listener/storage/")
		if err != nil {
			log.Error.Printf("Could not get random pseudonym: %s", err)
			b.Errorf("Could not get random pseudonym: %s", err)
		}

		start := time.Now()
		_, err = storage.QuerySingleLogFromBlockchain(pseudonym)
		duration := time.Since(start)

		if err != nil {
			log.Error.Printf("Could not query single log from blockchain: %s", err)
			b.Errorf("Could not query single log from blockchain: %s", err)
		}

		log.Info.Printf("Reading single from blockchain: %v\n", duration)
	}
}

func BenchmarkSearchSingleLogSQLite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pseudonym, err := _getRandomPseudonym("../listener/storage/")
		if err != nil {
			log.Error.Printf("Could not get random pseudonym: %s", err)
			b.Errorf("Could not get random pseudonym: %s", err)
		}

		start := time.Now()
		_, err = storage.QuerySingleLogFromSQLite(pseudonym)
		duration := time.Since(start)

		if err != nil {
			log.Error.Printf("Could not query single log from SQLite DB: %s", err)
			b.Errorf("Could not query single log from SQLite DB: %s", err)
		}

		log.Info.Printf("Reading single from SQLite DB: %v\n", duration)
	}
}

// helper function
func _getRandomPseudonym(directory string) (string, error) {
	entries, err := getAllLogs(directory)
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("no logs found")
	}

	return entries[random.PositiveIntFromRange(0, len(entries))].Pseudonym, nil
}
