package main

import (
	"fmt"
	"log"
	"node/constants"
	"os"
	"time"

	"github.com/pkg/profile"
	ownLog "node/logging"
	"node/revolori"
)

type returnValue struct {
	value                  string
	success                bool
	searchRestarts         int
	exchangeDuration       time.Duration
	loadIDCardDuration     time.Duration
	hostCreationDuration   time.Duration
	peerSearchDuration     time.Duration
	idVerificationDuration time.Duration
	newUsageMsgDuration    time.Duration
	decryptionDuration     time.Duration
	proofDuration          time.Duration
}

// run is needed since exiting the program with os.Exit or log.Fatal* results in defer not triggering. Thus, this
// function only returns the exit code.
func run() int {
	var err error
	config := parseFlags()

	if config.cpuProf {
		profilePath := fmt.Sprintf("cpu-%d", time.Now().Unix())

		defer profile.Start(profile.CPUProfile, profile.Quiet, profile.ProfilePath(profilePath)).Stop()
	} else if config.memProf {
		profilePath := fmt.Sprintf("mem-%d", time.Now().Unix())

		defer profile.Start(profile.MemProfile, profile.Quiet, profile.ProfilePath(profilePath)).Stop()
	}

	ownLog.Info.Println("\n\t===== Starting node =====")
	if !config.enableFakeChatter {
		fmt.Println("[!] Fake chatter has been disabled")
		ownLog.Info.Println("[!] Fake chatter has been disabled")
	}

	peerStart := time.Now()
	revoloriPublicKey, err = revolori.GetPublicKey()
	if err != nil {
		log.Printf("requester/run - Could not get Revolori's public key: %s\n", err)
		return 1
	}

	// Check if identity card exists
	globalPrivateKey, err = revolori.Setup(true)
	if err != nil {
		ownLog.Error.Println(err)
		return 1
	}
	startUpDuration := time.Since(peerStart)

	go createNode(&config, &peerStart)

	ret := <-realDone

	if config.enableFakeChatter {
		timeOut := 15 * time.Second

		select {
		case <-fakeDone:
			// There were at least 5 fake exchanges
		case <-time.After(timeOut):
			// There were less than 5 fake exchanges
			ownLog.Info.Println("There were less than 5 fake exchanges. Terminated after time-out.")
		}
	}

	ownLog.Info.Printf("Exchange summary\n" +
		fmt.Sprintf("\tCompleted fake exchanges: %d\n", fakeConnectionsAmount) +
		fmt.Sprintf("\tSearch restarts: %d\n", ret.searchRestarts) +
		fmt.Sprintf("\tDuration of entire exchange: %dms\n", ret.exchangeDuration.Milliseconds()) +
		fmt.Sprintf("\tGet Revolori's public key and create/load private key: %dms\n", startUpDuration.Milliseconds()) +
		fmt.Sprintf("\tLoad ID duration: %dms\n", ret.loadIDCardDuration.Milliseconds()) +
		fmt.Sprintf("\tCreate node: %dms\n", ret.hostCreationDuration.Milliseconds()) +
		fmt.Sprintf("\tPeer search duration: %dms\n", ret.peerSearchDuration.Milliseconds()) +
		fmt.Sprintf("\tDuration of id verification: %dms\n", ret.idVerificationDuration.Milliseconds()) +
		fmt.Sprintf("\tDuration of the new-usage protocol: %dms\n", (ret.newUsageMsgDuration+ret.decryptionDuration).Milliseconds()) +
		fmt.Sprintf("\t\tDuration of msg exchange + timeout: %dms\n", ret.newUsageMsgDuration.Milliseconds()) +
		fmt.Sprintf("\t\tTimeout duration: %dms\n", constants.MaxWaitTime.Milliseconds()) +
		fmt.Sprintf("\t\tDuration of decryption: %dms\n", ret.decryptionDuration.Milliseconds()) +
		fmt.Sprintf("\t\tDuration of writing proof of non-repudiation: %dms", ret.proofDuration.Milliseconds()),
	)

	if !ret.success {
		ownLog.Error.Printf("Exchange failed!")
		return 1
	}

	fmt.Println(ret.value)
	return 0
}

func main() {
	os.Exit(run())
}
