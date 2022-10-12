package main

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	success string
	files   []string
}

const invalidParamExitCode = 64

func ParseFlags() Config {
	config := Config{}

	var isDispute bool
	isDisputeUsage := fmt.Sprintf("If there is a dispute set this flag and pass the two files.\nExit codes:\n\tSuccess: %d,\n\tFailure: %d,\n\tJudgment not possible: %d", judgementSuccess, judgementFailure, judgementNotPossible)

	flag.StringVar(&config.success, "checkSuccess", "", "This will validate that the exchange stored in the passed file was ended successfully ")
	flag.BoolVar(&isDispute, "isDispute", false, isDisputeUsage)
	flag.Parse()

	config.files = flag.Args()

	if config.success == "" {
		if len(config.files) != 2 {
			fmt.Printf("Wrong amount of parameters were passed. Expected 2, got %d\n", len(config.files))
			flag.PrintDefaults()
			os.Exit(invalidParamExitCode)
		}

		if !isDispute {
			fmt.Println("No flags were set. Usage:")
			flag.PrintDefaults()
			os.Exit(invalidParamExitCode)
		}
	} else if len(config.files) > 0 {
		if !isDispute {
			fmt.Println("Too many parameters were passed. Usage:")
			flag.PrintDefaults()
			os.Exit(invalidParamExitCode)
		}

		fmt.Println("Can not use both flags at the same time")
		os.Exit(invalidParamExitCode)
	}

	return config
}
