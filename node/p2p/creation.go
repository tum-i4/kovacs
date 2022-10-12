package p2p

import (
	"crypto/rand"
	"fmt"
	"net"
	"strconv"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/multiformats/go-multiaddr"
	"node/constants"
	"node/logging"
)

// Adapted from https://github.com/libp2p/go-libp2p/tree/v0.16.0/examples/chat-with-mdns

// MakeHost returns a p2p host.
func MakeHost(port int) (host.Host, error) {
	randomness := rand.Reader
	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, constants.RSAKeySize, randomness)
	if err != nil {
		log.Info.Println(err)
		return nil, err
	}

	if !portIsFree(port) {
		return nil, fmt.Errorf("node/MakeHost - Port %d is already in use", port)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	return libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
}

func portIsFree(port int) bool {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}

	_ = listener.Close()
	return true
}
