package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	nP "node/nonRepudiation"
)

const (
	sleepDurationAfterExport = 3
)

func main() {
	config := parseFlags()

	// Private key that will be used as if it belongs to the requester's identity
	idCardPrivateKey, err := nP.GenerateRSAPrivateKey()
	if err != nil {
		log.Fatalf("could not create identity card: %s\n", err)
	}

	fmt.Printf("Step size: %d\n", config.stepSize)
	fmt.Printf("Target: %d\n\n", config.target)

	out := "Log amount, Blockchain size (B), SQLite size (B)\n"

	durations := make([]float64, 0)
	for i := 0; i < config.target+1; i++ {
		if i%config.stepSize == 0 {
			// measure size
			var sqliteSize int64 = 0
			fmt.Printf("Measuring")
			if i != 0 {
				sqliteSize = measureSQLiteSizeInBytes()
				// Sleep to give blockchain time to publish blocks
				time.Sleep(5 * time.Second)
			}

			blockchainSize := measureBlockchainSizeInBytes()

			sizes := fmt.Sprintf("%d, %d, %d\n", i, blockchainSize, sqliteSize)
			out += sizes
			fmt.Printf(": %s\n", sizes)

			if i == config.target {
				break
			}
		}

		// write log
		start := time.Now()
		writeLog(&idCardPrivateKey, i)
		if config.isBigNetwork {
			time.Sleep(sleepDurationAfterExport * time.Second)
		}

		durations = append(durations, time.Since(start).Seconds())
		percentage := (float64(i+1) / float64(config.target)) * 100

		fmt.Printf("Created log %d/%d (%.2f%s) - about %s left\n", i+1, config.target, percentage, "%", predictDuration(durations, config))
	}

	err = os.WriteFile("size.csv", []byte(out), 0644)
	if err != nil {
		log.Printf(out)
		log.Fatalf("\n\nCould not create size.csv: %s\n", err)
	}
}

func predictDuration(durations []float64, config configuration) string {
	sort.Float64s(durations) // sort changes the underlying array, however we don't care about that
	medianDuration := 0.0
	if len(durations) > 0 {
		if len(durations)%2 != 0 {
			medianDuration = durations[len(durations)/2]
		} else {
			medianDuration = (durations[(len(durations)/2)-1] + durations[len(durations)/2]) / 2
		}
	}

	// Estimate how long the remaining logs will take
	remaining := float64(config.target-len(durations)) * medianDuration
	// Add the 5 second wait time before measuring
	remaining += float64(5 * (config.target/config.stepSize - (len(durations)-1)/config.stepSize))
	remainingDuration := time.Duration(remaining) * time.Second

	return remainingDuration.String()
}
