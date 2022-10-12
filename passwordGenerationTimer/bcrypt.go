// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strconv"

	"golang.org/x/crypto/blowfish"
)

const (
	MinCost       int = 4  // the minimum allowable cost as passed in to GenerateFromPassword
	MaxCost       int = 31 // the maximum allowable cost as passed in to GenerateFromPassword
	DefaultCost   int = 10 // the cost that will actually be set if a cost below MinCost is passed into GenerateFromPassword; With default cost of 21 the hashing will take about 1 minute
	PassLengthMin int = 6
)

// InvalidHashPrefixError The error returned from CompareHashAndPassword when a hash starts with something other than '$'.
type InvalidHashPrefixError byte

func (ih InvalidHashPrefixError) Error() string {
	return fmt.Sprintf("adaptedBcrypt: bcrypt hashes must start with '$', but hashedSecret started with '%c'", byte(ih))
}

type InvalidCostError int

func (ic InvalidCostError) Error() string {
	return fmt.Sprintf("adaptedBcrypt: cost %d is outside allowed range (%d,%d)", int(ic), int(MinCost), int(MaxCost))
}

const (
	majorVersion       = '2'
	minorVersion       = 'a'
	maxSaltSize        = 16
	maxCryptedHashSize = 23
	encodedSaltSize    = 22
	encodedHashSize    = 31
)

// magicCipherData is an IV for the 64 Blowfish encryption calls in
// bcrypt(). It's the string "OrpheanBeholderScryDoubt" in big-endian bytes.
var magicCipherData = []byte{
	0x4f, 0x72, 0x70, 0x68,
	0x65, 0x61, 0x6e, 0x42,
	0x65, 0x68, 0x6f, 0x6c,
	0x64, 0x65, 0x72, 0x53,
	0x63, 0x72, 0x79, 0x44,
	0x6f, 0x75, 0x62, 0x74,
}

type hashed struct {
	hash  []byte
	salt  []byte
	cost  int // allowed range is MinCost to MaxCost
	major byte
	minor byte
}

func GenerateSalt() ([]byte, error) {
	unencodedSalt := make([]byte, maxSaltSize)
	_, err := io.ReadFull(rand.Reader, unencodedSalt)
	if err != nil {
		return nil, err
	}

	encodedSalt := base64Encode(unencodedSalt)

	return encodedSalt, nil
}

// GeneratePasswordReturnSalt returns the bcrypt hash and used salt of the password at the given
// cost. If the cost given is less than MinCost, the cost will be set to DefaultCost, instead.
func GeneratePasswordReturnSalt(password []byte, hashDifficulty int) ([]byte, []byte, error) {
	encodedSalt, err := GenerateSalt()
	if err != nil {
		return nil, nil, err
	}

	p, err := newFromPassword(password, hashDifficulty, encodedSalt)
	if err != nil {
		return nil, nil, err
	}
	return p.Hash(), p.salt, nil
}

func newFromPassword(password []byte, cost int, encodedSalt []byte) (*hashed, error) {
	if len(password) < PassLengthMin {
		return nil, errors.New("adaptedBcrypt: Password is too short. Min: " + strconv.Itoa(PassLengthMin) + ", Got: " + strconv.Itoa(len(password)))
	}

	if cost < MinCost {
		cost = DefaultCost
	}
	p := new(hashed)
	p.major = majorVersion
	p.minor = minorVersion

	err := checkCost(cost)
	if err != nil {
		return nil, err
	}
	p.cost = cost

	p.salt = encodedSalt
	hash, err := bcrypt(password, p.cost, p.salt)
	if err != nil {
		return nil, err
	}
	p.hash = hash
	return p, err
}

func bcrypt(password []byte, cost int, salt []byte) ([]byte, error) {
	cipherData := make([]byte, len(magicCipherData))
	copy(cipherData, magicCipherData)

	c, err := expensiveBlowfishSetup(password, uint32(cost), salt)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 24; i += 8 {
		for j := 0; j < 64; j++ {
			c.Encrypt(cipherData[i:i+8], cipherData[i:i+8])
		}
	}

	// Bug compatibility with C bcrypt implementations. We only encode 23 of
	// the 24 bytes encrypted.
	hsh := base64Encode(cipherData[:maxCryptedHashSize])
	return hsh, nil
}

func expensiveBlowfishSetup(key []byte, cost uint32, salt []byte) (*blowfish.Cipher, error) {
	csalt, err := base64Decode(salt)
	if err != nil {
		if salt != nil {
			return nil, err
		}

		csalt = nil
	}

	// Bug compatibility with C bcrypt implementations. They use the trailing
	// NULL in the key string during expansion.
	// We copy the key to prevent changing the underlying array.
	ckey := append(key[:len(key):len(key)], 0)

	c, err := blowfish.NewSaltedCipher(ckey, csalt)
	if err != nil {
		return nil, err
	}

	var i, rounds uint64
	rounds = 1 << cost
	for i = 0; i < rounds; i++ {
		blowfish.ExpandKey(ckey, c)

		if csalt != nil {
			blowfish.ExpandKey(csalt, c)
		}
	}

	return c, nil
}

func (p *hashed) Hash() []byte {
	arr := make([]byte, 60)
	arr[0] = '$'
	arr[1] = p.major
	n := 2
	if p.minor != 0 {
		arr[2] = p.minor
		n = 3
	}
	arr[n] = '$'
	n++
	copy(arr[n:], []byte(fmt.Sprintf("%02d", p.cost)))
	n += 2
	arr[n] = '$'
	n++
	copy(arr[n:], p.salt)
	n += encodedSaltSize
	copy(arr[n:], p.hash)
	n += encodedHashSize
	return arr[:n]
}

func (p *hashed) String() string {
	return fmt.Sprintf("&{hash: %#v, salt: %#v, cost: %d, major: %c, minor: %c}", string(p.hash), p.salt, p.cost, p.major, p.minor)
}

func checkCost(cost int) error {
	if cost < MinCost || cost > MaxCost {
		return InvalidCostError(cost)
	}
	return nil
}
