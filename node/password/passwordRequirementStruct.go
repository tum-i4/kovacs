package passwordRequirement

type PasswordRequirement struct {
	passwordPlain  []byte
	passwordHashed []byte
	salt           []byte
}

// Used pointer for improved performance

// GetPasswordPlain returns the plain password of the passed passwordRequirement.
func (requirement *PasswordRequirement) GetPasswordPlain() []byte {
	return requirement.passwordPlain
}

// GetPasswordHashed returns the hashed password of the passed passwordRequirement.
func (requirement *PasswordRequirement) GetPasswordHashed() []byte {
	return requirement.passwordHashed
}

// GetSalt returns the salt of the passed passwordRequirement.
func (requirement *PasswordRequirement) GetSalt() []byte {
	return requirement.salt
}
