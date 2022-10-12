package main

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strings"

	"node/constants"
	"node/p2p"
	"node/storage"
)

const (
	judgementSuccess     = 0
	judgementFailure     = 1
	judgementNotPossible = 2
)

func solveDispute(files []string, revoloriPublicKey *rsa.PublicKey) {
	err := verifyThatFilesBelongTogether(files, revoloriPublicKey)
	if err != nil {
		printJudgment(err.Error(), judgementNotPossible)
	}

	firstRecorder, firstSSOID, firstDecrypted, firstErr := verifyExchange(files[0], revoloriPublicKey)
	printFileInfo(firstRecorder, firstSSOID, firstErr, 1)

	secondRecorder, secondSSOID, secondDecrypted, secondErr := verifyExchange(files[1], revoloriPublicKey)
	printFileInfo(secondRecorder, secondSSOID, secondErr, 2)

	if firstRecorder == constants.MessageTypeFailure || secondRecorder == constants.MessageTypeFailure {
		out := "" +
			"At least one file failed to parse\n" +
			"It is therefore not possible to determine whether both files belong to the same exchange\n" +
			"=> Unable to make a decision\n"
		printJudgment(out, judgementNotPossible)
	}

	if firstRecorder == secondRecorder {
		out := "" +
			"The files have the same type which means that they do not belong together\n" +
			"=> Wrong input files provided\n"

		printJudgment(out, judgementNotPossible)
	}

	if firstDecrypted != secondDecrypted {
		out := "" +
			"The decrypted content is not equal\n" +
			"=> Unable to make a decision"

		printJudgment(out, judgementNotPossible)
	}

	if firstErr == nil && secondErr == nil {
		out := "" +
			"Both files state that the exchange ended successfully\n" +
			"=> Protocol ended successfully\n"

		printJudgment(out, judgementSuccess)
	}

	if firstErr != nil {
		out := "" +
			"While the first file indicates that the exchange failed, the second file proves that it ended successfully\n" +
			"=> Protocol ended successfully\n"

		printJudgment(out, judgementSuccess)
	}

	if secondErr != nil {
		out := "" +
			"While the second file indicates that the exchange failed, the first file proves that it ended successfully\n" +
			"=> Protocol ended successfully\n"

		printJudgment(out, judgementFailure)
	}

	fmt.Println("Reached a state that should be impossible to reach")
	fmt.Println("Debugging")
	fmt.Printf("\tfirstRecorder: %d\n\tfirstSSOID: %s\n\tfirstErr: %s\n", firstRecorder, firstSSOID, firstErr)
	fmt.Println("=====")
	fmt.Printf("\tsecondRecorder: %d\n\tsecondSSOID: %s\n\tsecondErr: %s\n", secondRecorder, secondSSOID, secondErr)
	os.Exit(judgementNotPossible)
}

func verifyExchange(file string, revoloriPublicKey *rsa.PublicKey) (constants.MessageType, string, string, error) {
	signedMessages, conversationPrivateKey, identityKey, err := storage.LoadExchange(file)
	if err != nil {
		return constants.MessageTypeFailure, "", "", err
	}

	firstMessage, identityCard, err := p2p.ExtractAndVerifyMessages(signedMessages[:2], revoloriPublicKey)
	if err != nil {
		return constants.MessageTypeFailure, "", "", err
	}

	/**	Since the *receiving* party stores the first message the types are switched **/
	if firstMessage.Type == constants.MessageTypeListener {
		decrypted, errVerify := verifyRequesterSuccess(signedMessages[2:], &firstMessage.PublicKey, firstMessage.Datum)
		return constants.MessageTypeRequester, identityCard.SSOID, decrypted, errVerify
	}

	decrypted, err := verifyListenerSuccess(signedMessages[2:], &firstMessage.PublicKey, &conversationPrivateKey.PublicKey, &identityKey)
	return constants.MessageTypeListener, identityCard.SSOID, decrypted, err
}

func verifyThatFilesBelongTogether(files []string, revoloriPublicKey *rsa.PublicKey) error {
	if len(files) != 2 {
		return fmt.Errorf("invalid amount of fields passed: %d", len(files))
	}

	signedMessages1, conversationPrivateKey1, _, err := storage.LoadExchange(files[0])
	if err != nil {
		return fmt.Errorf("could not load the first file: %w", err)
	}

	signedMessages2, conversationPrivateKey2, _, err := storage.LoadExchange(files[1])
	if err != nil {
		return fmt.Errorf("could not load the second file: %w", err)
	}

	firstMessage1, identityCard1, err := p2p.ExtractAndVerifyMessages(signedMessages1[:2], revoloriPublicKey)
	if err != nil {
		return fmt.Errorf("could not parse the first file: %w", err)
	}

	firstMessage2, identityCard2, err := p2p.ExtractAndVerifyMessages(signedMessages2[:2], revoloriPublicKey)
	if err != nil {
		return fmt.Errorf("could not parse the second file: %w", err)
	}

	if firstMessage1.Type == firstMessage2.Type {
		return errors.New("the files have the same type => they do not belong together")
	}

	if identityCard1.SSOID == identityCard2.SSOID {
		return errors.New("the files have the same SSOID => they cannot belong to the exchange")
	}

	pseudonym1, err := storage.GeneratePseudonym(&firstMessage1.PublicKey)
	if err != nil {
		return fmt.Errorf("could not generate the pseudonym for the first file: %w", err)
	}

	pseudonym2, err := storage.GeneratePseudonym(&firstMessage2.PublicKey)
	if err != nil {
		return fmt.Errorf("could not generate the pseudonym for the second file: %w", err)
	}

	pseudonymStored1, err := storage.GeneratePseudonym(&conversationPrivateKey1.PublicKey)
	if err != nil {
		return fmt.Errorf("could not generate the pseudonym for the first file's private key: %w", err)
	}

	pseudonymStored2, err := storage.GeneratePseudonym(&conversationPrivateKey2.PublicKey)
	if err != nil {
		return fmt.Errorf("could not generate the pseudonym for the second file's private key: %w", err)
	}

	if pseudonym1 != pseudonymStored2 || pseudonym2 != pseudonymStored1 {
		return errors.New("pseudonyms do not match => these files do not belong to the same exchange")
	}

	return nil
}

func printFileInfo(recorder constants.MessageType, ssoid string, err error, fileNumber int) {
	fmt.Printf("============== Analysis of file %d ==============\n", fileNumber)

	if recorder == constants.MessageTypeFailure {
		fmt.Printf("* Failed to parse the file: %s\n", err)
		return
	}

	var strType string
	if recorder == constants.MessageTypeListener {
		strType = "Listener                             |"
	} else {
		strType = "Requester                            |"
	}

	spaces := 36 - len(ssoid)
	if spaces < 0 {
		spaces = 0
	}

	fmt.Printf("| * SSOID: %s%s|\n", ssoid, strings.Repeat(" ", spaces))
	fmt.Printf("| * Type: %s\n", strType)
	fmt.Printf("| * Successfully completed the protocol: %t  |\n", err == nil)

	fmt.Printf("================================================\n\n")
}

func printJudgment(content string, exitCode int) {
	arr := strings.Split(strings.TrimSpace(content), "\n")
	maxLength := len(arr[0])

	for _, elem := range arr {
		if len(elem) > maxLength {
			maxLength = len(elem)
		}
	}

	maxLength -= 6

	if maxLength%2 != 0 {
		maxLength++
	}

	tmp := strings.Repeat("=", maxLength/2)
	fmt.Printf("%s Judgment %s\n", tmp, tmp)

	for _, elem := range arr {
		spaces := strings.Repeat(" ", maxLength-len(elem)+7)
		fmt.Printf("| %s%s|\n", elem, spaces)
	}

	fmt.Printf("%s==========%s\n", tmp, tmp)

	os.Exit(exitCode)
}
