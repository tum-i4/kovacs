package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/crypto/blake2s"
)

func main() {
	fmt.Println("Amount of repetitions: 10")

	for i := 4; i < 32; i++ { // Try all difficulties
		fmt.Printf("Cost: %d\n", i)

		duration := 0 * time.Nanosecond
		for k := 0; k < 10; k++ { // Repeat each difficulty 20 times
			currentTime := timePasswordRequirementGeneration(i)
			fmt.Printf("Run %d: %s\n", k, currentTime)
			duration += currentTime
		}

		fmt.Printf("Average duration: %s\n======\n", duration/10)
	}
}

func timePasswordRequirementGeneration(difficulty int) time.Duration {
	start := time.Now()
	passwordPlain, err := GeneratePlainPassword()
	if err != nil {
		log.Fatalf("passwordRequirement/addPasswordRequirement - Could not generate plain password: %s", err)
	}

	passwordHashedTooLarge, _, err := GeneratePasswordReturnSalt(passwordPlain, difficulty)
	if err != nil {
		log.Fatalf("passwordRequirement/addPasswordRequirement - Could not calculate hash: %s", err)
	}

	// Calculate a blake2s hash of the password to guarantee that it is individual and exactly 32 bytes
	_ = blake2s.Sum256(passwordHashedTooLarge)

	return time.Since(start)
}

func GeneratePlainPassword() ([]byte, error) {
	passwordPlain := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, passwordPlain); err != nil {
		return nil, err
	}

	return passwordPlain, nil
}
