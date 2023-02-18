package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// measureBlockchainSize returns the size of the geth directory in bytes
func measureBlockchainSizeInBytes() int64 {
	var dirSizeInBytes int64 = 0

	err := filepath.WalkDir("/build/geth", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Fatalf("Error at path %q: %s\n", path, err)
		}

		if !info.IsDir() {
			file, err := info.Info()
			if err != nil {
				log.Fatalf("Could not get info for %q: %s\n", path, err)
			}
			dirSizeInBytes += file.Size()
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking directory: %s\n", err)
	}

	return dirSizeInBytes
}

// measureSQLiteSize returns the size of the SQLite database in bytes
func measureSQLiteSizeInBytes() int64 {
	db, err := os.Stat("/build/measureStorage/database.db")
	if err != nil {
		log.Fatalf("Could not open database file: %s\n", err)
	}

	if db.IsDir() {
		log.Fatalf("Database is a directory?\n")
	}

	// FileInfo.Size returns the size in bytes, we want kiB
	return db.Size()
}
