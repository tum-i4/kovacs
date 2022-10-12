package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const invalidParamExitCode = 64

type flagConfig struct {
	pseudonym           string
	updateJustification string
	updateDatum         string

	delete       bool
	searchAll    bool
	searchSingle bool
	update       bool
	getLogInfo   bool
}

func (config *flagConfig) getBoolFlagCount() int {
	list := []bool{config.delete, config.searchAll, config.searchSingle, config.update, config.getLogInfo}
	count := 0

	for _, element := range list {
		if element {
			count++
		}
	}

	return count
}

func parseFlags() (flagConfig, []string) {
	config := flagConfig{}

	flag.BoolVar(&config.searchSingle, "single", false, "If you want to search for a pseudonym")
	flag.BoolVar(&config.searchAll, "all", false, "If you want to search for all logs")
	flag.BoolVar(&config.delete, "delete", false, "If you want to delete a log entry")
	flag.BoolVar(&config.update, "update", false, "If you want to update a log")
	flag.BoolVar(&config.getLogInfo, "info", false, "If you want additional info about the logs")

	flag.StringVar(&config.pseudonym, "pseudonym", "", "The pseudonym to be deleted, searched or updated")
	flag.StringVar(&config.updateJustification, "updateJustification", "", "The updated justification")
	flag.StringVar(&config.updateDatum, "updateDatum", "", "The updated datum")
	flag.Parse()
	directories := flag.Args()

	flagCount := config.getBoolFlagCount()

	if flagCount == 0 {
		fmt.Println("No parameters given. Usage:")
		flag.PrintDefaults()
		os.Exit(invalidParamExitCode)
	}

	if flagCount > 1 {
		fmt.Printf("Invalid input: Got 2 or more flags:\n")
		fmt.Printf("\tdelete: %t\n", config.delete)
		fmt.Printf("\tsearchAll: %t\n", config.searchAll)
		fmt.Printf("\tsearchSingle: %t\n", config.searchSingle)
		fmt.Printf("\tupdate: %t\n", config.update)
		fmt.Printf("\tgetLogIngo: %t\n", config.getLogInfo)
		os.Exit(invalidParamExitCode)
	}

	if (config.delete || config.searchSingle || config.update) && config.pseudonym == "" {
		fmt.Println("No pseudonym given. Usage:")
		flag.PrintDefaults()
		os.Exit(invalidParamExitCode)
	}

	if config.update {
		config.updateJustification = strings.TrimSpace(config.updateJustification)
		config.updateDatum = strings.TrimSpace(config.updateDatum)

		if len(config.updateJustification) == 0 {
			fmt.Println("No updated justification given. Usage:")
			flag.PrintDefaults()
			os.Exit(invalidParamExitCode)
		}

		if len(config.updateDatum) == 0 {
			// Note: The datum shouldn't be updateable; Instead look up existing block and copy the datum from there
			// However, allow for this function in the PoC
			fmt.Println("No updated datum given. Usage:")
			flag.PrintDefaults()
			os.Exit(invalidParamExitCode)
		}
	}

	if config.pseudonym != "" {
		if len(config.pseudonym) != 64 {
			fmt.Println("Passed pseudonym is invalid: wrong size")
			os.Exit(invalidParamExitCode)
		}
	}

	if len(directories) == 0 {
		fmt.Println("Must give at least one directory")
		os.Exit(invalidParamExitCode)
	}

	// Remove duplicate directories
	directories = removeDuplicateDirectories(directories)

	// Check if directories are valid
	for _, location := range directories {
		info, err := os.Stat(location)
		if err != nil {
			fmt.Printf("Could not read file '%s': %v\n", location, err)
			os.Exit(1)
		}

		if !info.IsDir() {
			fmt.Printf("File '%s' is not a directory\n", location)
			os.Exit(invalidParamExitCode)
		}
	}

	return config, directories
}

func removeDuplicateDirectories(list []string) []string {
	alreadyExists := map[string]bool{}
	uniqueElements := make([]string, 0)

	for _, element := range list {
		if !alreadyExists[element] {
			alreadyExists[element] = true
			uniqueElements = append(uniqueElements, element)
		}
	}

	return uniqueElements
}
