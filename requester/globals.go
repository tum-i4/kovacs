package main

import (
	"crypto/rsa"
	"time"
)

const (
	// Connection search time.
	maxSearchTime          = 30 * time.Second // Avg. search time when using fake chatter is 15 seconds. Adding 15 seconds as overhead.
	minFakeConnectionCount = 5
)

var (
	// Public keys.
	revoloriPublicKey rsa.PublicKey
	globalPrivateKey  rsa.PrivateKey

	// Channels.
	realDone       = make(chan returnValue, 1)
	fakeDone       = make(chan bool, 1)
	exchangeFailed = make(chan bool, 1)

	// peer search.
	maxRetries       = 20
	foundCorrectPeer = false

	// fake chatter.
	fakeConnectionsAmount int32 = 0 //nolint: revive
)
