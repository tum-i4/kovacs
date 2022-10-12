package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/profile"
	ownLog "node/logging"
	"node/revolori"
)

type returnValue struct {
	value   string
	success bool
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

	revoloriPublicKey, err = revolori.GetPublicKey()
	if err != nil {
		log.Printf("requester/main - Could not get Revolori's public key: %s\n", err)
		return 1
	}

	// Check if identity card exists
	globalPrivateKey, err = revolori.Setup(true)
	if err != nil {
		ownLog.Error.Println(err)
		return 1
	}

	if config.enableFakeChatter {
		go createFakeNodes(config.port)
	}

	go createRealNode(config)

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

	ownLog.Info.Printf("I completed a total of %d fake connections\n", fakeConnectionsAmount)
	if !ret.success {
		return 1
	}

	fmt.Println(ret.value)
	return 0
}

func main() {
	os.Exit(run())
}
