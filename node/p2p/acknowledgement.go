package p2p

import (
	"bytes"
	"fmt"
	"time"
)

type Acknowledgement struct {
	Content   []byte `json:"content"`
	ID        int    `json:"id"`
	TimeStamp int64  `json:"time_stamp"`
}

// CheckErr checks if the Acknowledgement is valid by verifying that the Acknowledgement field is true, comparing the
// provided ID with the expected one and checking that the timestamp is from the past.
func (ack *Acknowledgement) CheckErr(expectedID int, lastTimeStamp int64, expectedContent []byte) error {
	now := time.Now().Unix() //nolint: ifshort

	if ack.ID != expectedID {
		return fmt.Errorf("invalid ID. Got :%d, Expected: %d", ack.ID, expectedID)
	}

	// now > ack.TimeStamp <= time.Now().Unix()
	if ack.TimeStamp < lastTimeStamp {
		return fmt.Errorf("invalid timestamp: Got a time stamp from before the last one. Current %d, last: %d", ack.TimeStamp, lastTimeStamp)
	}

	if ack.TimeStamp > now {
		return fmt.Errorf("invalid timestamp: Got a time stamp from the future. Now: %d, received: %d", now, ack.TimeStamp)
	}

	if !bytes.Equal(ack.Content, expectedContent) {
		return fmt.Errorf("invalid content. Got: '%s', expected: '%s'", ack.Content, expectedContent)
	}

	return nil
}
