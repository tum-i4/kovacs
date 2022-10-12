package main

import (
	"flag"
	"log"
)

func parseFlags() (int, bool) {
	var port int
	var printName bool

	flag.BoolVar(&printName, "whoAmI", true, "Print username associated with this listener")
	flag.IntVar(&port, "port", 40000, "Port to listen to. Defaults to 40000")
	flag.BoolVar(&cpuProf, "cpuProf", false, "Enable CPU profiling")
	flag.BoolVar(&memProf, "memProf", false, "Enable memory profiling")
	flag.Parse()

	if port < 1024 {
		log.Fatalf("listener/main - Port provided is too small (<1024)\n")
	}

	return port, printName
}
