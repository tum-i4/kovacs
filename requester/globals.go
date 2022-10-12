package main

import (
	"crypto/rsa"
	"time"
)

// Public keys.
var revoloriPublicKey rsa.PublicKey
var globalPrivateKey rsa.PrivateKey

var realDone = make(chan returnValue, 1)
var fakeDone = make(chan bool, 1)

// peer search.
var foundCorrectPeer = false

// fake chatter.
var fakeConnectionsAmount int32 = 0 //nolint: revive

// Connection search time.
const maxSearchTime = 90 * time.Second // Max search time when using fake chatter is 63 seconds. Adding 27 seconds as overhead.
const minFakeConnectionCount = 5

var terminationChanceInPercent int32 = 5
var terminationIncreaseInPercent int32 = 5
