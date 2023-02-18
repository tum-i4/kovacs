package main

import (
	"flag"
	"log"
)

type configuration struct {
	stepSize     int
	target       int
	isBigNetwork bool
}

func parseFlags() configuration {
	config := configuration{}

	flag.IntVar(&config.stepSize, "stepSize", 25, "Dictates after how many logs the database and blockchain sizes are measured. Defaults to 25")
	flag.IntVar(&config.target, "target", 2000, "The number of logs to create. Must be a multiple of stepSize. Defaults to 2000")
	flag.BoolVar(&config.isBigNetwork, "bigNetwork", false, "Enable for bigger networks as otherwise the network can crash due to too many blockchain updates")
	flag.Parse()

	if config.stepSize < 1 {
		log.Fatalf("measureStorage/parseFlags - Invalid step size of %d\n", config.stepSize)
	}

	if config.target < 1 {
		log.Fatalf("measureStorage/parseFlags - Invalid target of %d\n", config.target)
	}

	if config.target%config.stepSize != 0 {
		log.Fatalf("measureStorage/parseFlags - The target of %d cannot be met using the step size of %d. The target must be a multiple of step size\n", config.target, config.stepSize)
	}

	return config
}
