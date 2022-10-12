package main

import (
	"fmt"
	"node/revolori"
	"os"
	"testing"
	"time"

	"node/constants"
	log "node/logging"
)

func BenchmarkIDCardCreationTime(b *testing.B) {
	out := "Run, Registration time (ms)\n"

	for i := 0; i < b.N; i++ {
		err := _deleteKeyFile()
		if err != nil {
			log.Error.Printf("Could not delete keyfile: %v", err)
			b.Fatalf("Could not delete keyfile: %v", err)
		}

		start := time.Now()
		_, err = revolori.Setup(true)
		duration := time.Since(start)

		if err != nil {
			log.Error.Printf("Could not complete setup: %v", err)
			b.Fatalf("Could not complete setup: %v", err)
		}

		out += fmt.Sprintf("%d, %v\n", i, duration.Milliseconds())
	}

	// Write test data to file
	file, err := os.Create("registrationBenchmark.csv")
	if err != nil {
		log.Error.Printf("Could not open file to write csv: %v", err)
		log.Info.Printf(out)
		return
	}
	defer file.Close()

	_, err = file.WriteString(out)
	if err != nil {
		log.Error.Printf("Could not write output to file: %v", err)
		log.Info.Printf(out)
		return
	}
}

func _deleteKeyFile() error {
	_helper := func(path string) error {
		_, err := os.Stat(path)
		if err == nil {
			err = os.Remove(path)
			if err != nil {
				return fmt.Errorf("could not delete file: %v", err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("an error occurred when calling stat: %v", err)
		}

		return nil
	}

	err := _helper(constants.KeyFilePath)
	if err != nil {
		return err
	}

	return _helper(constants.IdentityFilePath)
}
