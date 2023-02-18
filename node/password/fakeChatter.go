package passwordRequirement

// FakeChatterPasswordRequirement generates a nonsensical password requirement, which should only be used for fake
// chatter!
func FakeChatterPasswordRequirement() PasswordRequirement {
	passwordPlain := []byte("HardCodedPasswordForFake_Chatter")
	salt := base64Encode([]byte("_HardCoded_Salt_"))
	passwordHashed := base64Encode([]byte("Not_A_Hash_But_Thats_Ok."))

	return PasswordRequirement{
		passwordPlain:  passwordPlain,
		passwordHashed: passwordHashed,
		salt:           salt,
	}
}
