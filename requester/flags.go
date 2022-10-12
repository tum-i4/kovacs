package main

import (
	"flag"
	"log"
	"strings"

	ownLog "node/logging"
)

type configuration struct {
	ssoid             string
	justification     string
	requestedDatum    string
	port              int
	enableFakeChatter bool
	cpuProf           bool
	memProf           bool
}

func parseFlags() configuration {
	config := configuration{}

	flag.StringVar(&config.ssoid, "ssoid", "", "SSOID of the peer you wish to connect to")
	flag.StringVar(&config.justification, "justification", "Requesting data", "Justification for data access, defaults to 'Requesting data'")
	flag.StringVar(&config.requestedDatum, "datum", "No datum given", "Which data you wish to request. Since this is a PoC, the listener doesn't care what data is requested")
	flag.IntVar(&config.port, "port", 41000, "Port to listen to, defaults to 41000")
	flag.BoolVar(&config.enableFakeChatter, "fakeChatter", false, "Set to true to enable fake chatter")
	flag.BoolVar(&config.cpuProf, "cpuProf", false, "Enable CPU profiling")
	flag.BoolVar(&config.memProf, "memProf", false, "Enable memory profiling")
	flag.Parse()

	config.ssoid = strings.TrimSpace(config.ssoid)
	config.justification = strings.TrimSpace(config.justification)
	config.requestedDatum = strings.TrimSpace(config.requestedDatum)

	if len(config.ssoid) == 0 {
		ownLog.Error.Printf("requester/parseFlags - No SSOID provided\n")
		log.Fatalf("requester/parseFlags - No SSOID provided\n")
	}

	if config.port < 1024 {
		ownLog.Error.Printf("requester/parseFlags - Port provided is too small (<1024)\n")
		log.Fatalf("requester/parseFlags - Port provided is too small (<1024)\n")
	}

	if config.cpuProf && config.memProf {
		ownLog.Error.Printf("requester/parseFlags - Both profilings have been enabled\n")
		log.Fatalf("requester/parseFlags - Both profilings have been enabled\n")
	}

	return config
}
