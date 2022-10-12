package main

import (
	"fmt"
	"log"
	"node/p2p"
	"os"
	"strings"

	"node/constants"
	"node/revolori"
	"node/storage"
)

func LogInfo(directories []string) {
	messageTypes := make([]string, 0)
	ssoids := make([]string, 0)
	myPseudonyms := make([]string, 0)
	otherPseudonyms := make([]string, 0)
	dates := make([]string, 0)

	revoloriPublicKey, err := revolori.GetPublicKey()
	if err != nil {
		log.Fatalf("logInfo - Could not get Revolori's key: %v\n", err)
	}

	for _, directory := range directories {
		fmt.Printf("Searching in directory '%s':\n", directory)

		files, err := os.ReadDir(directory)
		if err != nil {
			log.Printf("logInfo: Could not read the directory '%s': %v => Skipped it\n", directory, err)
			continue
		}

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				fmt.Printf("\tFound a file that is not json: '%s' => Skipped it\n", file.Name())
				continue
			}

			signedMessages, conversationPrivateKey, _, err := storage.LoadExchange(directory + file.Name())
			if err != nil {
				log.Printf("logInfo: Could not load exchange for '%s': %v\n", directory+file.Name(), err)
			}

			myPseudonym, err := storage.GeneratePseudonym(&conversationPrivateKey.PublicKey)
			if err != nil {
				log.Printf("\tCould not generate my pseudonym for '%s': %v => Skipping it\n", file.Name(), err)
				continue
			}

			if !strings.HasSuffix(file.Name(), "-"+myPseudonym+".json") {
				log.Printf("logInfo - File '%s' does not have the expected suffix of '%s' => Skipped it\n", file.Name(), myPseudonym)
				continue
			}

			firstMessage, identityCard, err := p2p.ExtractAndVerifyMessages(signedMessages[:2], &revoloriPublicKey)
			if err != nil {
				log.Fatalf("logInfo - Could not extract the first message: %v\n", err)
			}

			otherPseudonym, err := storage.GeneratePseudonym(&firstMessage.PublicKey)
			if err != nil {
				log.Printf("\tCould not generate pseudonym of the other person for '%s': %v => Skipping it\n", file.Name(), err)
				continue
			}

			myType := ""
			switch firstMessage.Type {
			case constants.MessageTypeListener:
				myType = "requester"
			case constants.MessageTypeRequester:
				myType = "listener"
			case constants.MessageTypeFailure, constants.MessageTypeFakeChatter, constants.MessageTypeRealExchange:
				log.Fatalf("Unexpected message type: %d\n", firstMessage.Type)
			default:
				log.Fatalf("Invalid message type: %d\n", firstMessage.Type)
			}

			messageTypes = append(messageTypes, myType)
			ssoids = append(ssoids, identityCard.SSOID)
			myPseudonyms = append(myPseudonyms, myPseudonym)
			otherPseudonyms = append(otherPseudonyms, otherPseudonym)
			dates = append(dates, file.Name()[:19])
		}

		logPrettyPrint(messageTypes, ssoids, myPseudonyms, otherPseudonyms, dates)
	}
}

func logPrettyPrint(messageTypes []string, ssoids []string, myPseudonyms []string, otherPseudonyms []string, dates []string) {
	length := len(messageTypes)
	if length != len(ssoids) || length != len(myPseudonyms) || length != len(otherPseudonyms) || length != len(dates) {
		log.Fatalln("logPrettyPrint - Unequal lengths!")
	}

	headline := ""
	separatorLength := 0

	getMaxWidth := func(list []string) int {
		maxWidth := 0
		for _, element := range list {
			if len(element) > maxWidth {
				maxWidth = len(element)
			}
		}

		return maxWidth
	}

	getHeadline := func(headerLine string, content []string) {
		var maxLineWidth int
		var widthToScale int
		headerLine = strings.TrimSpace(headerLine)
		lineWidth := getMaxWidth(content) - len(headerLine)

		if lineWidth < 0 {
			headline += fmt.Sprintf(" %s |", headerLine)
			maxLineWidth = len(headerLine) + 3
			widthToScale = maxLineWidth - 3
		} else {
			lineWidth += lineWidth % 2
			spaces := strings.Repeat(" ", lineWidth/2)
			headline += fmt.Sprintf(" %s%s%s |", spaces, headerLine, spaces)
			maxLineWidth = lineWidth + 3 + len(headerLine)
			widthToScale = maxLineWidth - 3
		}

		for i := 0; i < len(content); i++ {
			if widthToScale > len(content[i]) {
				content[i] += strings.Repeat(" ", widthToScale-len(content[i]))
			}
		}

		separatorLength += maxLineWidth
	}

	getHeadline("Date and time", dates)
	getHeadline("My pseudonym", myPseudonyms)
	getHeadline("My Type", messageTypes)
	getHeadline("Counterpart's SSOID", ssoids)
	getHeadline("Counterpart's pseudonym", otherPseudonyms)

	fmt.Printf("\n\n")
	fmt.Printf("%s\n", headline)
	fmt.Printf("%s\n", strings.Repeat("=", separatorLength))

	for i, messageType := range messageTypes {
		fmt.Printf(" %s | %s | %s | %s | %s |\n", dates[i], myPseudonyms[i], messageType, ssoids[i], otherPseudonyms[i])
	}

	fmt.Printf("\n\n")
}
