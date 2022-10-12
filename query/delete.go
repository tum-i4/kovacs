package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func DeleteLog(directories []string, pseudonym string) error {
	foundFile := false
	for _, directory := range directories {
		fmt.Printf("Searching in directory: %s\n", directory)

		files, err := os.ReadDir(directory)
		if err != nil {
			return fmt.Errorf("DeleteLog - Could not read directory: %w", err)
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), "-"+pseudonym+".json") {
				foundFile = true

				err = os.Remove(directory + file.Name())
				if err != nil {
					return fmt.Errorf("DeleteLog - Could not delete the log: %w", err)
				}
			}
		}
	}

	if !foundFile {
		return errors.New("DeleteLog - no log matched the passed pseudonym")
	}

	fmt.Printf("Deleted the requested log\n")
	return nil
}
