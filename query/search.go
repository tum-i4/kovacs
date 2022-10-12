package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"node/storage"
)

type entry struct {
	Pseudonym  string
	PrivateKey rsa.PrivateKey
}

func SearchAllLogs(directories []string) {
	allEntries := make([]entry, 0)
	for _, directory := range directories {
		fmt.Printf("Searching in directory: %s\n", directory)

		entries, err := getAllLogs(directory)
		if err != nil {
			log.Fatalf("An error occurred: %v\n", err)
		}

		if len(entries) == 1 {
			fmt.Printf("\tFound 1 log\n")
		} else {
			fmt.Printf("\tFound %d logs\n", len(entries))
		}
		allEntries = append(allEntries, entries...)
	}

	if len(allEntries) == 0 {
		fmt.Printf("No logs were found\n")
		os.Exit(0)
	}

	allEntries = removeDuplicateEntries(allEntries)

	fmt.Printf("\n")
	if len(allEntries) == 1 {
		fmt.Printf("I have 1 unique log\n")
	} else {
		fmt.Printf("I have %d unique logs\n", len(allEntries))
	}

	allLogs, err := storage.QueryAllLogs()
	if err != nil {
		log.Fatalf("Could not query logs: %s", err)
	}

	myLogs := make(map[string]storage.UsageLogContent)
	for _, entry := range allEntries {
		for _, singleLog := range allLogs {
			if singleLog.PseudonymConsumer == entry.Pseudonym || singleLog.PseudonymOwner == entry.Pseudonym {
				// Attempt to decrypt the log
				decrypted, consumerErr, ownerErr := decryptLog(singleLog, &entry.PrivateKey)
				if consumerErr == nil && ownerErr == nil {
					myLogs[entry.Pseudonym] = decrypted
					break
				}

				// Could not decrypt log => Print error
				fmt.Printf("! Could not decrypt log with pseudonym %s\n\tConsumer error: %s\n\tOwner error: %s\n\tLog: %v\n", entry.Pseudonym, consumerErr, ownerErr, singleLog)

				// There is a maximum of one log with any given pseudonym
				// Thus we can stop our search here
				break
			}
		}

		_, ok := myLogs[entry.Pseudonym]
		if !ok {
			fmt.Printf("! Could not find a uage log for pseudonym '%s' in the blockchain!\n", entry.Pseudonym)
		}
	}

	printUsageLogs(myLogs)
}

func SearchSingleLog(directories []string, pseudonym string) {
	foundEntry := entry{
		Pseudonym: "",
	}

	for _, directory := range directories {
		fmt.Printf("Searching in directory: %s\n", directory)

		entries, err := getAllLogs(directory)
		if err != nil {
			log.Fatalf("An error occurred: %s\n", err)
		}

		// Check if a log with the matched pseudonym has been found
		for _, entry := range entries {
			if entry.Pseudonym == pseudonym {
				foundEntry = entry
				break
			}
		}

		if foundEntry.Pseudonym != "" {
			break
		}
	}

	if foundEntry.Pseudonym == "" {
		log.Fatalf("SearchSingleLog - No log was found with the passed pseudonym\n")
		return
	}

	fmt.Printf("Found the P3 log entry\n")
	allLogs, err := storage.QueryAllLogs()
	if err != nil {
		log.Fatalf("Could not query logs: %s", err)
	}

	for _, singleLog := range allLogs {
		if singleLog.PseudonymConsumer == pseudonym || singleLog.PseudonymOwner == pseudonym {
			// Attempt to decrypt the log
			decrypted, consumerErr, ownerErr := decryptLog(singleLog, &foundEntry.PrivateKey)
			if consumerErr == nil && ownerErr == nil {
				printUsageLogs(map[string]storage.UsageLogContent{
					foundEntry.Pseudonym: decrypted,
				})
				break
			}

			// Could not decrypt log => Print error
			log.Fatalf("! Could not decrypt log with pseudonym %s\n\tConsumer error: %s\n\tOwner error: %s\n\tLog: %v\n", foundEntry.Pseudonym, consumerErr, ownerErr, singleLog)
		}
	}
}

func printUsageLogs(logs map[string]storage.UsageLogContent) {
	maxTimestampLength := len("Unix Timestamp")
	maxPseudoLength := len("Pseudonym")
	maxJustificationLength := len("Justification")
	maxDatumLength := len("Usage Log Content")
	for pseudonym, payload := range logs {
		if maxPseudoLength < len(pseudonym) {
			maxPseudoLength = len(pseudonym)
		}

		if maxJustificationLength < len(payload.Justification) {
			maxJustificationLength = len(payload.Justification)
		}

		if maxDatumLength < len(payload.DatumRequest) {
			maxDatumLength = len(payload.DatumRequest)
		}

		timestampStr := strconv.FormatInt(payload.Timestamp, 10)
		if maxTimestampLength < len(timestampStr) {
			maxTimestampLength = len(timestampStr)
		}
	}

	padding := func(str string, length int) string {
		if len(str) > length {
			return str
		}

		return str + strings.Repeat(" ", length-len(str))
	}

	center := func(str string, length *int) string {
		if len(str) > *length {
			return str
		}

		newLength := *length - len(str)
		if newLength%2 != 0 {
			*length++
			newLength++
		}

		spaces := strings.Repeat(" ", newLength/2)
		return spaces + str + spaces
	}

	// Print header
	fmt.Printf("\n%s | %s | %s | %s\n",
		center("Unix Timestamp", &maxTimestampLength),
		center("Pseudonym", &maxPseudoLength),
		center("Usage Log Content", &maxDatumLength),
		center("Justification", &maxJustificationLength),
	)

	// Print separator
	fmt.Printf("%s|%s|%s|%s\n",
		strings.Repeat("=", maxTimestampLength+1),
		strings.Repeat("=", maxPseudoLength+1),
		strings.Repeat("=", maxDatumLength+1),
		strings.Repeat("=", maxJustificationLength+2),
	)

	// Print body
	for pseudonym, payload := range logs {
		fmt.Printf("%s | %s | %s | %s\n",
			padding(strconv.FormatInt(payload.Timestamp, 10), maxTimestampLength),
			padding(pseudonym, maxPseudoLength),
			padding(payload.DatumRequest, maxDatumLength),
			padding(payload.Justification, maxJustificationLength),
		)
	}
}

// decryptLog return the decrypted log or the decryption errors in the following order:
// dataConsumerError, dataOwnerError.
func decryptLog(singleLog storage.BlockchainPayload, privateKey *rsa.PrivateKey) (storage.UsageLogContent, error, error) {
	// Attempt to decrypt the log as data consumer
	decrypted, consumerErr := decryptUsageLogContent(&singleLog.EncryptedConsumer, privateKey)
	if consumerErr == nil {
		return decrypted, nil, nil
	}

	// Could not decrypt as consumer => Decrypt the log as data owner
	decrypted, ownerErr := decryptUsageLogContent(&singleLog.EncryptedOwner, privateKey)
	if ownerErr == nil {
		return decrypted, nil, nil
	}

	// Could not decrypt log
	return storage.UsageLogContent{}, consumerErr, ownerErr
}

func decryptUsageLogContent(usageLog *storage.UsageLogContent, privateKey *rsa.PrivateKey) (storage.UsageLogContent, error) {
	decryptedJustification, err := storage.PublicKeyDecryption(usageLog.Justification, privateKey)
	if err != nil {
		return storage.UsageLogContent{}, err
	}

	decryptedDatum, err := storage.PublicKeyDecryption(usageLog.DatumRequest, privateKey)
	if err != nil {
		return storage.UsageLogContent{}, err
	}

	return storage.UsageLogContent{
		Justification: string(decryptedJustification),
		DatumRequest:  string(decryptedDatum),
		Timestamp:     usageLog.Timestamp,
	}, nil
}

func getAllLogs(path string) ([]entry, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	entries := make([]entry, 0)
	for _, file := range files {
		if file.IsDir() {
			fmt.Printf("\tgetAllLogs - Found a directory at '%s' => Skipped it\n", file.Name())
			continue
		}

		if !strings.HasSuffix(file.Name(), ".json") {
			fmt.Printf("\tgetAllLogs - File (%s) is not json => Skipped it\n", file.Name())
			continue
		}

		_, conversationPrivateKey, _, err := storage.LoadExchange(path + file.Name())
		if err != nil {
			log.Printf("getAllLogs - Could not load exchange '%s' because '%v' => Skipped it\n", file.Name(), err)
			continue
		}

		pseudonym, err := storage.GeneratePseudonym(&conversationPrivateKey.PublicKey)
		if err != nil {
			log.Printf("getAllLogs - Could not generate pseudonym for '%s' because '%v' => Skipped it\n", file.Name(), err)
			continue
		}

		if !strings.HasSuffix(file.Name(), "-"+pseudonym+".json") {
			log.Printf("getAllLogs - File '%s' does not have the expected suffix of '%s' => Skipped it\n", file.Name(), pseudonym)
			continue
		}

		entries = append(entries, entry{
			Pseudonym:  pseudonym,
			PrivateKey: conversationPrivateKey,
		})
	}

	return entries, nil
}

func removeDuplicateEntries(entries []entry) []entry {
	if len(entries) == 1 {
		return entries
	}

	alreadyExists := map[string]bool{}
	uniqueElements := make([]entry, 0)

	for _, element := range entries {
		if !alreadyExists[element.Pseudonym] {
			alreadyExists[element.Pseudonym] = true
			uniqueElements = append(uniqueElements, element)
		}
	}

	return uniqueElements
}
