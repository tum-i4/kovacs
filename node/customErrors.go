package node

import (
	"fmt"
	"time"
)

type TimeOutError struct {
	MaxWaitTime time.Duration
}

func (err TimeOutError) Error() string {
	return fmt.Sprintf("Took too long to receive request. Maximum wait time is %v", err.MaxWaitTime)
}
