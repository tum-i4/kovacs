package main

import "log"

func main() {
	config, directories := parseFlags()

	if config.delete {
		err := DeleteLog(directories, config.pseudonym)
		if err != nil {
			log.Fatalln(err)
		}

		return
	}

	if config.searchAll {
		SearchAllLogs(directories)
		return
	}

	if config.searchSingle {
		SearchSingleLog(directories, config.pseudonym)
		return
	}

	if config.update {
		UpdateLog(directories, config.pseudonym, config.updateJustification, config.updateDatum)
		return
	}

	if config.getLogInfo {
		LogInfo(directories)
		return
	}
}
