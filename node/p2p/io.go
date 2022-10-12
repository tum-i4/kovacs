package p2p

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"time"

	"node"
	"node/constants"
)

// Write writes to the ReadWriter and checks for errors.
func Write(rw *bufio.ReadWriter, data string) error {
	_, err := rw.WriteString(data)
	if err != nil {
		return fmt.Errorf("node/Write - Could not write string to ReadWriter - %w", err)
	}

	err = rw.Flush()
	if err != nil {
		return fmt.Errorf("node/Write - Could not flush ReadWriter - %w", err)
	}

	return nil
}

// ReadLine attempts to read from the ReadWriter. If no waitTime is passed then node.MaxWaitTime is used. Returns the
// read string on success or an error on failure.
func ReadLine(rw *bufio.ReadWriter, delim byte, waitTime ...time.Duration) (string, error) {
	// Get rid of line feed
	peak, err := rw.Peek(rw.Reader.Buffered())
	if err != nil {
		return "", fmt.Errorf("node/ReadLine - Could not peak: %w", err)
	}

	if bytes.Equal(peak, []byte{10}) {
		// Got a line feed -> move reader by one byte
		_, err = rw.ReadByte()
		if err != nil {
			return "", fmt.Errorf("node/ReadLine - Failed to move reader after a line feed: %w", err)
		}
	}

	// Init vars to be able to assign to them
	var str string
	var timeOut time.Duration

	if len(waitTime) > 0 {
		timeOut = waitTime[0]
	} else {
		timeOut = constants.MaxWaitTime
	}

	channelOutput := make(chan string, 1)
	channelError := make(chan error, 1)

	go read(rw, delim, channelOutput, channelError)

	select {
	case str = <-channelOutput:
		return strings.TrimSpace(str), nil
	case err = <-channelError:
		return "", fmt.Errorf("an error occurred when attempting to read from the ReadWriter: %w", err)
	case <-time.After(timeOut):
		return "", &node.TimeOutError{
			MaxWaitTime: timeOut,
		}
	}
}

func read(rw *bufio.ReadWriter, delim byte, outChan chan string, errorChan chan error) {
	read, err := rw.ReadString(delim)
	if err != nil {
		errorChan <- err
		return
	}

	outChan <- read
}
