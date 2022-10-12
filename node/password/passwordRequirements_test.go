package passwordRequirement

import (
	"bytes"
	"testing"

	"node/constants"
)

const prRuns = 10

func TestGeneratePasswordRequirement(t *testing.T) {
	for i := 0; i < prRuns; i++ {
		err := addPasswordRequirement(nil)
		if err != nil {
			t.Errorf("TestGeneratePasswordRequirement - Could not add password requirement: %s\n", err)
		}

		lenBeforePop := len(passwordRequirements)

		if lenBeforePop == 0 {
			t.Errorf("TestGeneratePasswordRequirement - Did not add an element to the password requirement list. List length: %d\n", len(passwordRequirements))
		}

		val, err := GeneratePasswordRequirement()
		if err != nil {
			t.Errorf("TestGeneratePasswordRequirement - Coud not pop password requirement: %s\n", err)
		}

		if bytes.Equal(val.passwordPlain, []byte{}) {
			t.Errorf("TestGeneratePasswordRequirement - plained password is empty\n")
		} else if len(val.passwordPlain) != constants.PasswordPlainLength {
			t.Errorf("TestGeneratePasswordRequirement - plain password has invalid size: %d, expected: %d\n", len(val.passwordPlain), constants.PasswordPlainLength)
		}

		if bytes.Equal(val.passwordHashed, []byte{}) {
			t.Errorf("TestGeneratePasswordRequirement - hashed password is empty\n")
		} else if len(val.passwordHashed) != constants.PasswordHashedLength {
			t.Errorf("TestGeneratePasswordRequirement - hashed password has invalid size: %d, expected: %d\n", len(val.passwordPlain), constants.PasswordHashedLength)
		}

		if bytes.Equal(val.salt, []byte{}) {
			t.Errorf("TestGeneratePasswordRequirement - salt is empty\n")
		} else if len(val.salt) != constants.SaltLength {
			t.Errorf("TestGeneratePasswordRequirement - salt has invalid size: %d, expected: %d\n", len(val.salt), constants.SaltLength)
		}

		if len(passwordRequirements) >= lenBeforePop {
			t.Errorf("TestGeneratePasswordRequirement - requirement list length did not decrease\n")
		}
	}
}

func TestAddPasswordRequirement(t *testing.T) {
	for i := 0; i < prRuns; i++ {
		passwordRequirements = []PasswordRequirement{}

		err := addPasswordRequirement(nil)
		if err != nil {
			t.Errorf("TestAddPasswordRequirement - Could not add password requirement: %s\n", err)
		}

		if len(passwordRequirements) == 0 {
			t.Errorf("TestAddPasswordRequirement - Did not add an element to the password requirement list")
		}
	}
}

func TestAddRequirementFailure(t *testing.T) {
	if len(passwordRequirements) != constants.RequirementListLength {
		err := fillPasswordRequirementList()
		if err != nil {
			t.Errorf("TestAddRequirementFailure - Could not fill list")
		}
	}
}

func TestFillPasswordRequirementList(t *testing.T) {
	for i := 0; i < prRuns; i++ {
		passwordRequirements = make([]PasswordRequirement, 0)

		err := fillPasswordRequirementList()
		if err != nil {
			t.Errorf("TestFillPasswordRequirementList - Ran into error: %s\n", err)
		}

		if len(passwordRequirements) != constants.RequirementListLength {
			t.Errorf("TestFillPasswordRequirementList - Did not fill list. Got: %d, expected: %d\n", len(passwordRequirements), constants.RequirementListLength)
		}
	}
}

func TestGeneratePlainPassword(t *testing.T) {
	for i := 0; i < prRuns; i++ {
		pass, err := GeneratePlainPassword()
		if err != nil {
			t.Errorf("TestGeneratePlainPassword - Could not generate password: %s\n", err)
		}

		if len(pass) != 32 {
			t.Errorf("TestGeneratePlainPassword - Invalid password length of %d, expected %d\n", len(pass), 32)
		}
	}
}
