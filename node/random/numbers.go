package random

import (
	"crypto/rand"
	"math"
	"math/big"
	mRand "math/rand"
	"time"

	"node/logging"
)

// PositiveIntFromRange returns an int in the range of [min, max).
// Uses rand.Int and falls back to mRand if rand.Int should throw an error.
func PositiveIntFromRange(min int, max int) int {
	if max > math.MaxInt {
		log.Info.Printf("random/PositiveIntFromRange - Maximum is more than math.MaxInt => Set tot math.MaxInt -1")
		max = math.MaxInt - 1
	}

	number, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		log.Info.Printf("random/PositiveIntFromRange - Could not use crypto rand (%v) => Falling back to mRand", err)
		mRand.Seed(time.Now().UnixNano())
		return mRand.Intn(max-min) + min //nolint:gosec
	}

	return min + int(number.Int64())
}
