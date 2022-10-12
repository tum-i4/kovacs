package passwordRequirement

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/blake2s"
	"node/constants"
	"node/logging"
)

var (
	passwordRequirements      = make([]PasswordRequirement, 0, constants.RequirementListLength)
	passwordRequirementsMutex = &sync.Mutex{}
)

// Init fills the password requirement list. The function blocks until the first element is added. The rest is
// added in a go routine to avoid blocking the main execution for longer than necessary.
// The first element is blocking to ensure that the listener can handle a connection as soon as it is online.
func Init() {
	err := addPasswordRequirement(nil)
	if err != nil {
		log.Info.Printf("passwordRequirement/init - Could not add first password requirement: %v", err)
	}

	go func() {
		err = fillPasswordRequirementList()
		if err != nil {
			log.Info.Printf("passwordRequirement/init - Failed to fill password requirement list: %v", err)
		}
	}()
}

// GeneratePasswordRequirement returns the first element of the password requirement lists and adds a new one to replace the
// old one.
func GeneratePasswordRequirement() (PasswordRequirement, error) {
	var err error

	if len(passwordRequirements) == 0 {
		log.Info.Println("passwordRequirement/GeneratePasswordRequirement - Ran out of password requirements. Generating new ones")

		err = addPasswordRequirement(nil)
		if err != nil {
			log.Info.Fatalf("passwordRequirement/GeneratePasswordRequirement - Could not create a new, last password requirement: %v", err)
		}
	}

	passwordRequirementsMutex.Lock()
	element := passwordRequirements[0]
	passwordRequirements = passwordRequirements[1:]
	passwordRequirementsMutex.Unlock()

	go func() {
		err = addPasswordRequirement(nil)
		if err != nil {
			log.Info.Printf("passwordRequirement/GeneratePasswordRequirement - Could not add password requirement: %v", err)
		}
	}()

	return element, err
}

// addPasswordRequirement appends an element to the requirement list. It's supposed to be called as a go routine.
func addPasswordRequirement(group *sync.WaitGroup) error {
	if group != nil {
		defer group.Done()
	}

	if len(passwordRequirements) >= constants.RequirementListLength {
		log.Info.Println("passwordRequirement/addPasswordRequirement - Wanted to add password to a full requirement list")
		return nil
	}

	passwordPlain, err := GeneratePlainPassword()
	if err != nil {
		return fmt.Errorf("passwordRequirement/addPasswordRequirement - Could not generate plain password: %w", err)
	}

	passwordHashedTooLarge, salt, err := GeneratePasswordReturnSalt(passwordPlain)
	if err != nil {
		return fmt.Errorf("passwordRequirement/addPasswordRequirement - Could not calculate hash: %w", err)
	}

	// Calculate a blake2s hash of the password to guarantee that it is individual and exactly 32 bytes
	blakeSum := blake2s.Sum256(passwordHashedTooLarge)

	requirement := PasswordRequirement{
		passwordPlain:  passwordPlain,
		passwordHashed: blakeSum[:],
		salt:           salt,
	}

	passwordRequirementsMutex.Lock()
	passwordRequirements = append(passwordRequirements, requirement)
	passwordRequirementsMutex.Unlock()

	return nil
}

// GeneratePlainPassword returns a random byte array with the fixed size of 32byte.
func GeneratePlainPassword() ([]byte, error) {
	passwordPlain := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, passwordPlain); err != nil {
		return nil, err
	}

	return passwordPlain, nil
}

// fillPasswordRequirementList should be called at start up to fill the password requirement list.
func fillPasswordRequirementList() error {
	var group sync.WaitGroup

	var errEnd error

	for i := 0; i < constants.RequirementListLength-len(passwordRequirements); i++ {
		group.Add(1)

		go func() {
			err := addPasswordRequirement(&group)
			if err != nil {
				errEnd = err
				log.Info.Printf("passwordRequirement/fillPasswordRequirementList - Could not add password requirement: %v", errEnd)
			}
		}()
	}

	group.Wait()

	return errEnd
}
