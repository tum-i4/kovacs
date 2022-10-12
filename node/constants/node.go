package constants

import "time"

const (
	// MaxWaitTime is the default maximum wait time for a peer to send data.
	MaxWaitTime = 2 * time.Second
	// P2PProtocolName is the name of the peer-to-peer protocol.
	P2PProtocolName = "/P3/1.0.0"
)
